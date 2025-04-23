package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const messageDelimiter = '\n'

type ClientHandler struct {
	Id            uuid.UUID
	Clients       map[string]*Client
	AuthHandler   *AuthHandler
	Broker        *Broker
	AuthChannel   chan AuthNotification
	BrokerChannel chan BrokerNotification
}

func (c *ClientHandler) handleClient(conn net.Conn) {
	reader := bufio.NewReader(conn)

	log.Println("Client connected: ", conn.RemoteAddr().String())

	client := &Client{
		Id:         uuid.New(),
		IPAddress:  conn.RemoteAddr().String(),
		Channels:   map[int]*Channel{},
		Connection: conn,
	}

	c.Clients[client.IPAddress] = client

	for {
		message, err := reader.ReadString(messageDelimiter)
		if err != nil {
			log.Println("Error reading message", err)
			return
		}

		frame, err := ParseMessage(message, client.IPAddress)
		if err != nil {
			log.Println("ERROR PARSE: ", err)
			return
		}

		channel := client.Channels[frame.ChannelID]
		if channel == nil {
			client.Channels[frame.ChannelID] = &Channel{
				Id:   frame.ChannelID,
				Open: true,
			}
		}

		switch frame.ChannelID {
		case 0:
			go c.AuthHandler.Handle(*frame)
		default:
			if !client.Authenticated {
				log.Println("❌ Auth failed - client has not been authenticated.")
				client.WriteToClient("❌ Not authenticated! Rejected.")
				client.Connection.Close()
				return
			} else {
				go c.Broker.Handle(*frame)
			}
		}
	}
}

func (c *ClientHandler) AuthChanListen() {
	for notification := range c.AuthChannel {
		switch notification.AuthMessage {
		case ACCEPT:
			c.OnClientAccepted(notification.ClientIP)
		case DENY:
			c.OnClientDenied(notification.ClientIP, notification.Reason)
		}
	}
}

func (c *ClientHandler) OnClientAccepted(clientIP string) {
	client := c.Clients[clientIP]
	client.Authenticated = true

	client.WriteToClient("CONNECTED")
}

func (c *ClientHandler) OnClientDenied(clientIP string, reason DenyReason) {
	client := c.Clients[clientIP]
	client.Authenticated = false

	client.WriteToClient("DENIED - REASON: " + string(reason))

	client.Connection.Close()

	delete(c.Clients, clientIP)
}

func (c *ClientHandler) BrokerChanListen() {
	for notification := range c.BrokerChannel {
		client := c.Clients[notification.ClientIP]
		if client == nil {
			log.Println("Client could not be retrieved from Broker notification clientIP. Skipping.")
			continue
		}

		client.WriteToClient(notification.ChannelID + ";" + notification.ConsumerTag + ";" + notification.Message.Id.String() + ";" + notification.Message.Message)
	}
}

func (client *Client) WriteToClient(value string) {
	data := []byte(value + string('\n'))
	_, err := client.Connection.Write(data)
	if err != nil {
		log.Println("Error writing data to client:", err)
	}
}

func ParseMessage(data string, clientIp string) (*Frame, error) {
	data = strings.TrimSpace(data)

	dataArr := strings.Split(data, ";")

	channelID, err := strconv.Atoi(dataArr[0])
	if err != nil {
		return nil, err
	}

	command := dataArr[1]

	consumerTag := ""
	queueName := ""
	message := ""
	password := ""
	msgId := ""

	switch len(dataArr) {
	case 3:
		if command == "CONNECT" {
			password = dataArr[2]
		} else {
			queueName = dataArr[2]
		}
	case 4:
		if command == "PUBLISH" {
			queueName = dataArr[2]
			message = dataArr[3]
		}
		if command == "SUBSCRIBE" {
			consumerTag = dataArr[2]
			queueName = dataArr[3]
		}
	case 5:
		if command == "ACK" {
			consumerTag = dataArr[2]
			queueName = dataArr[3]
			msgId = dataArr[4]
		}
	}

	return &Frame{
		ChannelID:   channelID,
		ClientIp:    clientIp,
		Command:     command,
		Password:    password,
		Message:     message,
		QueueName:   queueName,
		ConsumerTag: consumerTag,
		MessageId:   msgId,
	}, nil
}

type Client struct {
	Id            uuid.UUID
	IPAddress     string
	Channels      map[int]*Channel
	Subscriptions map[int]*Subscription
	Connection    net.Conn
	Authenticated bool
}

type FrameHandler interface {
	Handle(frame Frame) bool
}

type Channel struct {
	Id        int
	Open      bool
	Consumer  Consumer
	Publisher Publisher
}

type Frame struct {
	ChannelID   int
	Command     string
	Password    string
	ClientIp    string
	QueueName   string
	ConsumerTag string
	Message     string
	MessageId   string
}
