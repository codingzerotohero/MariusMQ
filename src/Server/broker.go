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
		go broker.Publish(frame)
	case 3:
		go broker.Subscribe(frame)
	}
}

func (broker *Broker) CreateQueue(frame Frame) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	name := frame.Payload

	queue := &Queue{Id: uuid.New(), Name: name, Subscriptions: map[string]*Subscription{}, Messages: []*Message{}}
	broker.Queues[name] = queue

	log.Println("Queue created: " + name + " - Id: " + queue.Id.String())
}

func (broker *Broker) Publish(frame Frame) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	payload := frame.Payload

	queueName := strings.Split(payload, "|")[0]
	message := strings.Split(payload, "|")[1]

	msg := &Message{
		Message:      "{" + message + "}",
		Acknowledged: false,
		Delivered:    false,
	}

	queue := broker.Queues[queueName]

	if queue != nil {
		queue.Messages = append(queue.Messages, msg)
		msg.Acknowledged = true

		broker.MessageChan <- queueName
	}
}

func (broker *Broker) Subscribe(frame Frame) {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	queueName := frame.Payload
	queue := broker.Queues[queueName]
	if queue != nil {
		subscription := &Subscription{
			ChannelID: frame.ChannelID,
			QueueName: queueName,
			Type:      "CONSUMER",
		}
		queue.Subscriptions[frame.ClientIp] = subscription
	}
}

func (broker *Broker) QueueListen() {
	for queueName := range broker.MessageChan {
		queue := broker.Queues[queueName]
		if queue != nil && len(queue.Subscriptions) > 0 {
			for clientIP, subscription := range queue.Subscriptions {
				if subscription.Type != "CONSUMER" {
					continue
				}
				log.Println("Subscriber client IP: " + clientIP)
				messages := queue.Messages

				for _, message := range messages {
					go broker.SendMessageNotification(clientIP, message, strconv.Itoa(subscription.ChannelID))
					message.Delivered = true
				}
			}

			for i := range queue.Messages {
				queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
			}
		}
	}
}

func (broker *Broker) SendMessageNotification(clientIP string, message *Message, channelID string) {
	broker.BrokerChannel <- BrokerNotification{
		Notification: Notification{
			Type: "SEND",
		},
		ClientIP:  clientIP,
		ChannelID: channelID,
		Message:   message.Message,
	}
}

type Queue struct {
	Id            uuid.UUID
	Name          string
	Subscriptions map[string]*Subscription
	Messages      []*Message
}

type Message struct {
	Message      string
	Acknowledged bool
	Delivered    bool
}

type Subscription struct {
	ChannelID int
	QueueName string
	Type      string
}

type BrokerNotification struct {
	Notification
	ClientIP  string
	ChannelID string
	Message   string
}
