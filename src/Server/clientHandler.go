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
	Id          uuid.UUID
	Clients     map[string]*Client
	AuthHandler *AuthHandler
	Broker      *Broker
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

		switch frame.ChannelID {
		case 0:
			go c.AuthHandler.Handle(*frame, client)
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

		/*
			if !c.Dispatcher.Dispatch(*frame, client) {
				switch frame.ChannelID {
				case 0:
					if !client.Authenticated {
						data := []byte("Authentication failed - bad password")
						_, err = conn.Write(data)
						if err != nil {
							log.Println("Error sending data: ", err)
						}
						client.Connection.Close()
						return
					}
				}
			}*/
	}
}

func (c *ClientHandler) SendMessageToConsumer(client Client, channel Channel, message string) {
	data := []byte(strconv.Itoa(channel.Id) + ";" + message)
	client.Connection.Write(data)
}

func closeConnection(conn net.Conn) {
	conn.Close()
}

func (c *ClientHandler) Route(message string, Client *Client) {

	//cmd := Command(frame.Command)

	/*
		switch cmd {
		case CONNECT:
			log.Println("Routing to Auth")
		case CREATE:
			//log.Println("Handle create queue request")
			//go c.QueueHandler.CreateQueue(value)
			//data := []byte("CREATED")
			//_, err = conn.Write(data)
			//if err != nil {
			//	log.Println("Error sending data: ", err)
			//}
		case PUBLISH:
			log.Println("Handle publish request")
		case SUBSCRIBE:
			log.Println("Handle subscribe request")
		default:
			log.Println("Something else" + cmd)
		}*/
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

func (c *ClientHandler) SendMessageToClient(client *Client, message string) {

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
