package main

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	orderedmap "github.com/wk8/go-ordered-map"
)

type Broker struct {
	Queues                map[string]*Queue
	BrokerMessageChan     chan BrokerMessageNotification
	BrokerQueueActionChan chan BrokerQueueActionNotification
	InternalMessageChan   chan string
	mu                    sync.RWMutex
}

func (broker *Broker) Handle(frame Frame) {
	switch frame.ChannelID {
	case 1:
		go broker.CreateQueue(frame)
	case 2:
		go broker.DeleteQueue(frame)
	case 3:
		go broker.Publish(frame)
	case 4:
		go broker.Subscribe(frame)
	case 5:
		go broker.ConsumerAcknowledgeMessage(frame)
	}
}

func (broker *Broker) CreateQueue(frame Frame) {
	name := frame.QueueName

	broker.mu.RLock()
	queue := broker.Queues[name]
	broker.mu.RUnlock()

	if queue != nil {
		go broker.SendQueueActionNotification(CREATE_QUEUE_FAIL_ALREADY_EXISTS, frame.ClientIp, strconv.Itoa(frame.ChannelID), name)
		return
	}

	queue = &Queue{Id: uuid.New(),
		Name:          name,
		Subscriptions: map[string]*Subscription{},
		Messages:      orderedmap.New(),
	}

	broker.mu.Lock()
	broker.Queues[name] = queue
	broker.mu.Unlock()

	log.Println("Queue created: " + name + " - Id: " + queue.Id.String())

}

func (broker *Broker) Publish(frame Frame) {
	queueName := frame.QueueName
	message := frame.Message

	msg := &Message{
		Id:           uuid.New(),
		Message:      "{" + message + "}",
		Acknowledged: false,
		Delivered:    false,
	}

	broker.mu.RLock()
	queue := broker.Queues[queueName]
	broker.mu.RUnlock()

	if queue == nil {
		go broker.SendQueueActionNotification(PUBLISH_MESSAGE_FAIL_QUEUE_DOES_NOT_EXIST, frame.ClientIp, strconv.Itoa(frame.ChannelID), frame.QueueName)
		return
	}

	queue.mu.Lock()

	//OrderedMap implementation
	queue.Messages.Set(msg.Id.String(), msg)
	//queue.Messages[msg.Id.String()] = msg
	queue.mu.Unlock()

	broker.InternalMessageChan <- queueName + ";" + msg.Id.String()

}

func (broker *Broker) DeleteQueue(frame Frame) {
	broker.mu.Lock()
	queue := broker.Queues[frame.QueueName]
	if queue == nil {
		broker.mu.Unlock()
		go broker.SendQueueActionNotification(DELETE_QUEUE_FAIL_DOES_NOT_EXIST, frame.ClientIp, strconv.Itoa(frame.ChannelID), frame.QueueName)
		return
	}

	delete(broker.Queues, frame.QueueName)
	broker.mu.Unlock()

	queue.mu.Lock()

	//queue.Messages = nil
	queue.Messages = nil

	for clientIP, subscription := range queue.Subscriptions {
		go broker.SendQueueActionNotification(QUEUE_DELETED, clientIP, strconv.Itoa(subscription.ChannelID), frame.QueueName)
		subscription.Consumers = nil
		subscription.Publishers = nil
	}

	queue.Subscriptions = nil
	queue.mu.Unlock()
}

func (broker *Broker) Subscribe(frame Frame) {
	broker.mu.RLock()
	queue := broker.Queues[frame.QueueName]
	broker.mu.RUnlock()

	if queue == nil {
		go broker.SendQueueActionNotification(SUBSCRIBE_QUEUE_FAIL_DOES_NOT_EXIST, frame.ClientIp, strconv.Itoa(frame.ChannelID), frame.QueueName)
		return
	}

	subscription := &Subscription{
		ChannelID: frame.ChannelID,
		QueueName: frame.QueueName,
		Type:      "CONSUMER",
		Consumers: map[string]*Consumer{},
	}

	subscription.Consumers[frame.ConsumerTag] = &Consumer{
		Id:    uuid.New(),
		Name:  frame.ConsumerTag,
		Ready: true,
	}

	queue.mu.Lock()
	queue.Subscriptions[frame.ClientIp] = subscription
	queue.mu.Unlock()

	go broker.DistributeAll(queue)

}

func (broker *Broker) QueueListen() {
	for queueNameAndMessageId := range broker.InternalMessageChan {
		queueName := strings.Split(queueNameAndMessageId, ";")[0]
		msgId := strings.Split(queueNameAndMessageId, ";")[1]

		broker.mu.RLock()
		queue := broker.Queues[queueName]
		broker.mu.RUnlock()

		if queue != nil && len(queue.Subscriptions) > 0 {
			go broker.DistributeMessage(msgId, queue)
		}
	}
}

func (broker *Broker) DistributeMessage(strMsgId string, queue *Queue) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	log.Println("Sending message: " + strMsgId)

	value, exists := queue.Messages.Get(strMsgId)
	if !exists {
		log.Println("Could not find message with id: " + strMsgId + " in Queue: " + queue.Name)
		return
	}

	message, ok := value.(*Message)
	if !ok {
		log.Println("Type assertion failed")
		return
	}

	if message.Delivered == true {
		return
	}

outer:
	for clientIP, subscription := range queue.Subscriptions {
		if subscription.Type != "CONSUMER" {
			continue
		}

		for consumerTag, consumer := range subscription.Consumers {
			if consumer.Ready {
				go broker.SendMessageNotification(SEND_MESSAGE, clientIP, consumerTag, *message, strconv.Itoa(subscription.ChannelID))
				consumer.Ready = false
				message.Delivered = true
				break outer
			}
		}
	}
}

func (broker *Broker) DistributeAll(queue *Queue) {
	queue.mu.Lock()
	messageIDs := make([]string, 0, queue.Messages.Len())

	for pair := queue.Messages.Oldest(); pair != nil; pair = pair.Next() {
		message, _ := pair.Value.(*Message)
		if message.Delivered == true {
			continue
		}

		messageId, _ := pair.Key.(string)

		log.Println("Appending message: " + messageId)
		messageIDs = append(messageIDs, messageId)
	}

	queue.mu.Unlock()

	for _, messageId := range messageIDs {
		log.Println("Distributing message: " + messageId)
		broker.DistributeMessage(messageId, queue)
	}
}

func (broker *Broker) SendMessageNotification(notificationType NotificationType, clientIP string, consumerTag string, message Message, channelID string) {
	broker.BrokerMessageChan <- BrokerMessageNotification{
		NotificationType: notificationType,
		ClientIP:         clientIP,
		ChannelID:        channelID,
		Message:          message,
		ConsumerTag:      consumerTag,
	}
}

func (broker *Broker) SendQueueActionNotification(notificationType NotificationType, clientIP string, channelID string, queueName string) {
	broker.BrokerQueueActionChan <- BrokerQueueActionNotification{
		NotificationType: notificationType,
		ClientIP:         clientIP,
		ChannelID:        channelID,
		QueueName:        queueName,
	}
}

func (broker *Broker) ConsumerAcknowledgeMessage(frame Frame) {
	queue := broker.Queues[frame.QueueName]

	queue.mu.Lock()

	subscription := queue.Subscriptions[frame.ClientIp]
	consumer := subscription.Consumers[frame.ConsumerTag]
	//OrderedMap implementation
	value, exists := queue.Messages.Get(frame.MessageId)
	if !exists {
		log.Println("Could not find message with id: " + frame.MessageId + " in Queue: " + queue.Name)
		return
	}

	message, ok := value.(*Message)
	if !ok {
		log.Println("Type assertion failed")
		return
	}
	//message := queue.Messages[frame.MessageId]

	message.Acknowledged = true
	consumer.Ready = true

	//OrderedMap implementation
	queue.Messages.Delete(frame.MessageId)

	//delete(queue.Messages, frame.MessageId)
	queue.mu.Unlock()

	broker.DistributeAll(queue)
}

type Queue struct {
	Id            uuid.UUID
	Name          string
	Subscriptions map[string]*Subscription
	Messages      *orderedmap.OrderedMap
	mu            sync.Mutex
}

type Consumer struct {
	Id    uuid.UUID
	Name  string
	Ready bool
}

type Publisher struct {
	Id        uuid.UUID
	Name      string
	QueueName string
}

type Message struct {
	Id           uuid.UUID
	Message      string
	Acknowledged bool
	Delivered    bool
}

type Subscription struct {
	ChannelID  int
	QueueName  string
	Consumers  map[string]*Consumer
	Publishers map[string]*Publisher
	Type       string
}

type BrokerMessageNotification struct {
	NotificationType NotificationType
	ClientIP         string
	ChannelID        string
	Message          Message
	ConsumerTag      string
}

type BrokerQueueActionNotification struct {
	NotificationType NotificationType
	ClientIP         string
	ChannelID        string
	QueueName        string
}
