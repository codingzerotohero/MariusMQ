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
	Clients     []*Client
	AuthHandler AuthHandler
	Dispatcher  Dispatcher
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

		frame, err := ParseMessage(message)
		if err != nil {
			log.Println("ERROR PARSE: ", err)
			return
		}
		c.Dispatcher.Dispatch(*frame)
		//c.Route(n, client)
	}
}

func closeConnection(conn net.Conn) {
	conn.Close()
}

func (r *ClientHandler) Route(message string, Client *Client) {

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

func ParseMessage(message string) (*Frame, error) {
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
	}, nil
}

type Client struct {
	Id            uuid.UUID
	IPAddress     string
	Channels      map[int]*Channel
	Connection    net.Conn
	Authenticated bool
}

type FrameHandler interface {
	Handle(frame Frame) bool
}

type Channel struct {
	Id int
}

type Frame struct {
	ChannelID int
	Command   string
	Payload   string
}
