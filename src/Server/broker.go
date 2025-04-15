package main

import (
	"log"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Broker struct {
	Queues        map[string]*Queue
	ClientHandler *ClientHandler
	EventBus      chan *Queue
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

func (broker *Broker) CreateQueue(frame Frame) bool {
	broker.mu.Lock()
	defer broker.mu.Unlock()

	name := frame.Payload

	queue := &Queue{Id: uuid.New(), Name: name, Subscriptions: map[string]*Subscription{}, Messages: []*Message{}}
	broker.Queues[name] = queue

	log.Println("Queue created: " + name + " - Id: " + queue.Id.String())

	broker.ClientHandler.Clients[frame.ClientIp].Channels[frame.ChannelID] = &Channel{
		Id:     frame.ChannelID,
		Inbox:  []string{},
		Outbox: []string{},
	}

	return true
}

func (broker *Broker) Publish(frame Frame) bool {
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

		broker.ClientHandler.Clients[frame.ClientIp].Channels[frame.ChannelID] = &Channel{
			Id:     frame.ChannelID,
			Inbox:  []string{},
			Outbox: []string{},
		}

		broker.EventBus <- queue
	} else {
		return false
	}

	return true
}

func (broker *Broker) Subscribe(frame Frame) bool {
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

		broker.ClientHandler.Clients[frame.ClientIp].Channels[frame.ChannelID] = &Channel{
			Id:     frame.ChannelID,
			Inbox:  []string{},
			Outbox: []string{},
		}
	}
	return true
}

func (broker *Broker) StartListener() {
	for queue := range broker.EventBus {
		if queue != nil && len(queue.Subscriptions) > 0 {
			for clientIP, subscription := range queue.Subscriptions {
				if subscription.Type != "CONSUMER" {
					continue
				}
				log.Println("Subscriber client IP: " + clientIP)
				client := broker.ClientHandler.Clients[clientIP]
				channel := client.Channels[subscription.ChannelID]
				if client == nil || channel == nil {
					log.Println("Client or Channel null - skipping this subscriber")
					if client == nil {
						log.Println("Client is null!")
					}
					if channel == nil {
						log.Println("Channel is nullers!!")
					}
					continue
				}
				messages := queue.Messages

				for _, message := range messages {
					broker.ClientHandler.SendMessageToConsumer(*client, *channel, message.Message)
					message.Delivered = true
				}
			}

			for i := range queue.Messages {
				queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
			}
		}
	}
}

/*
func (broker *Broker){

		for queue := range broker.EventBus{
			for clientIP, subscription := range queue.Subscriptions {
				if subscription.Type == "CONSUMER" {
					client := broker.ClientHandler.Clients[clientIP]
					channel := client.Channels[3]
					for _, message := range queue.Messages {
						channel.Outbox = append(channel.Outbox, message.Message)
					}
				}
			}
		}
	}
*/
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
