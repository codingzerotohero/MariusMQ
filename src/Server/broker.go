package main

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Broker struct {
	Queues        map[string]*Queue
	BrokerChannel chan BrokerNotification
	MessageChan   chan string
	mu            sync.Mutex
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
	broker.mu.Lock()
	defer broker.mu.Unlock()

	name := frame.QueueName

	queue := &Queue{Id: uuid.New(), Name: name, Subscriptions: map[string]*Subscription{}, Messages: map[string]*Message{}}
	broker.Queues[name] = queue

	log.Println("Queue created: " + name + " - Id: " + queue.Id.String())
}

func (broker *Broker) Publish(frame Frame) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	queueName := frame.QueueName
	message := frame.Message

	msg := &Message{
		Id:           uuid.New(),
		Message:      "{" + message + "}",
		Acknowledged: false,
		Delivered:    false,
	}

	queue := broker.Queues[queueName]

	if queue != nil {
		queue.Messages[msg.Id.String()] = msg

		broker.MessageChan <- queueName + ";" + msg.Id.String()
	}
}

func (broker *Broker) DeleteQueue(frame Frame) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	delete(broker.Queues, frame.QueueName)
}

func (broker *Broker) Subscribe(frame Frame) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	queue := broker.Queues[frame.QueueName]
	if queue != nil {
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

		queue.Subscriptions[frame.ClientIp] = subscription

		go broker.DistributeAll(queue)
	}
}

func (broker *Broker) QueueListen() {
	for queueNameAndMessageId := range broker.MessageChan {
		queueName := strings.Split(queueNameAndMessageId, ";")[0]
		msgId := strings.Split(queueNameAndMessageId, ";")[1]

		queue := broker.Queues[queueName]
		if queue != nil && len(queue.Subscriptions) > 0 {
			go broker.DistributeMessage(msgId, queue)
		}
	}
}

func (broker *Broker) DistributeMessage(strMsgId string, queue *Queue) {
	message := queue.Messages[strMsgId]
	if message.Delivered == false {
		return
	}

outer:
	for clientIP, subscription := range queue.Subscriptions {
		if subscription.Type != "CONSUMER" {
			continue
		}
		log.Println("Subscriber client IP: " + clientIP)

		for consumerTag, consumer := range subscription.Consumers {
			if consumer.Ready {
				go broker.SendMessageNotification(clientIP, consumerTag, *message, strconv.Itoa(subscription.ChannelID))
				consumer.Ready = false
				message.Delivered = true
				break outer
			}
		}
	}
}

func (broker *Broker) DistributeAll(queue *Queue) {
	for messageId := range queue.Messages {
		go broker.DistributeMessage(messageId, queue)
	}
}

func (broker *Broker) SendMessageNotification(clientIP string, consumerTag string, message Message, channelID string) {
	broker.BrokerChannel <- BrokerNotification{
		ClientIP:    clientIP,
		ChannelID:   channelID,
		Message:     message,
		ConsumerTag: consumerTag,
	}
}

func (broker *Broker) ConsumerAcknowledgeMessage(frame Frame) {
	queue := broker.Queues[frame.QueueName]
	subscription := queue.Subscriptions[frame.ClientIp]
	consumer := subscription.Consumers[frame.ConsumerTag]
	message := queue.Messages[frame.MessageId]

	message.Acknowledged = true
	consumer.Ready = true

	delete(queue.Messages, frame.MessageId)

}

type Queue struct {
	Id            uuid.UUID
	Name          string
	Subscriptions map[string]*Subscription
	Messages      map[string]*Message
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

type BrokerNotification struct {
	ClientIP    string
	ChannelID   string
	Message     Message
	ConsumerTag string
}
