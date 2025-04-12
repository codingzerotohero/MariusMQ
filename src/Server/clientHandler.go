package main

import (
	"bufio"
	"log"
	"net"
	"strings"

	"github.com/google/uuid"
)

const messageDelimiter = '\n'

type ClientHandler struct {
	Id             uuid.UUID
	ServerPassword string
	Connections    []net.Conn
	QueueHandler   QueueHandler
}

func (c *ClientHandler) handleClient(conn net.Conn) {
	reader := bufio.NewReader(conn)

	log.Println("Client connected: ", conn.RemoteAddr().String())

	for {
		n, err := reader.ReadString(messageDelimiter)
		if err != nil {
			log.Println("Error reading message", err)
			return
		}
		n = strings.TrimSpace(n)
		log.Printf("Client message: %s\n", n)

		var messageArr = strings.Split(n, ";")

		//if len(messageArr) > 2 {
		//	log.Println("Invalid message format!")
		//	continue
		//}

		cmd := Command(messageArr[0])
		value := messageArr[1]

		log.Println(cmd)

		switch cmd {
		case CONNECT:
			log.Println("Handle connection request")
			if value == c.ServerPassword {
				log.Println("Client connection request accepted - " + conn.RemoteAddr().String())
				data := []byte("CONNECT REQUEST ACCEPTED")
				_, err = conn.Write(data)
				if err != nil {
					log.Println("Error sending data: ", err)
					return
				}

			} else {
				log.Println("Refused connection request from client: " + conn.RemoteAddr().String() + " due to bad password")

				data := []byte("CONNECT REQUEST DENIED: INCORRECT PASSWORD.")
				_, err = conn.Write(data)
				if err != nil {
					log.Println("Error sending data: ", err)
				}
				conn.Close()
				return
			}
		case CREATE:
			log.Println("Handle create queue request")
			c.QueueHandler.CreateQueue(value)
		case PUBLISH:
			log.Println("Handle publish request")
		case SUBSCRIBE:
			log.Println("Handle subscribe request")
		default:
			log.Println("Something else" + cmd)
		}
	}
}

func closeConnection(conn net.Conn) {
	conn.Close()
}

func (c *ClientHandler) closeAllConnections() {
	for _, conn := range c.Connections {
		conn.Close()
	}
}
