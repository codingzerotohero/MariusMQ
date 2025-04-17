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
				Id:     frame.ChannelID,
				Inbox:  []string{},
				Outbox: []string{},
			}
		}

		switch frame.ChannelID {
		case 0:
			go c.AuthHandler.Handle(*frame)
		default:
			if !client.Authenticated {
				log.Println("❌ Auth failed - client has not been authenticated.")
				data := []byte("❌ Not authenticated! Rejected.")
				_, err = client.Connection.Write(data)
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

	data := []byte("CONNECTED" + string('\n'))
	_, err := client.Connection.Write(data)
	if err != nil {
		log.Println("Error writing data to client:", err)
	}
}

func (c *ClientHandler) OnClientDenied(clientIP string, reason DenyReason) {
	client := c.Clients[clientIP]
	client.Authenticated = false

	data := []byte("DENIED - REASON: " + string(reason) + string('\n'))
	_, err := client.Connection.Write(data)
	if err != nil {
		log.Println("Error writing data to client:", err)
	}

	client.Connection.Close()

	delete(c.Clients, clientIP)
}

func (c *ClientHandler) ListenBrokerNotifications() {
	for notification := range c.BrokerChannel {
		client := c.Clients[notification.ClientIP]
		if client == nil {
			log.Println("Client could not be retrieved from Broker notification clientIP. Skipping.")
			continue
		}

		data := []byte(notification.ChannelID + ";" + notification.Message)
		client.Connection.Write(data)
	}
}

func closeConnection(conn net.Conn) {
	conn.Close()
}

func ParseMessage(message string, clientIp string) (*Frame, error) {
	message = strings.TrimSpace(message)

	var messageArr = strings.Split(message, ";")

	channelID, err := strconv.Atoi(messageArr[0])
	if err != nil {
		return nil, err
	}

	payload := messageArr[2]

	return &Frame{
		ChannelID: channelID,
		Command:   messageArr[1],
		Payload:   payload,
		ClientIp:  clientIp,
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
	Id     int
	Inbox  []string
	Outbox []string
}

type Frame struct {
	ChannelID int
	Command   string
	Payload   string
	ClientIp  string
}
