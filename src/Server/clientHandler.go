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
}

type ClientPool struct {
	Connections []net.Conn
}

func (c *ClientHandler) handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	log.Println("Client connected: ", conn.LocalAddr())

	for {
		n, err := reader.ReadString(messageDelimiter)
		if err != nil {
			log.Println("Error reading message", err)
			return
		}
		n = strings.TrimSpace(n)
		log.Printf("Client message: %s\n", n)

		var messageArr = strings.Split(n, ";")

		if len(messageArr) > 2 {
			log.Println("Invalid message format!")
			continue
		}

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
				data := []byte("CONNECT REQUEST DENIED: INCORRECT PASSWORD.")
				_, err = conn.Write(data)
				if err != nil {
					log.Println("Error sending data: ", err)
					return
				}
			}
		case CREATE:
			log.Println("Handle create queue request")
		default:
			log.Println("Something else" + cmd)
		}
	}
}
