// server.go
package main

import (
	"log"
	"net"

	"github.com/google/uuid"
)

var conns []net.Conn

func startServer(configuration Configuration) {

	listener, err := net.Listen("tcp", configuration.MARIUSMQ_SERVERADDRESS+":"+configuration.MARIUSMQ_SERVERPORT)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	defer listener.Close()

	log.Println("Server is listening on port: " + configuration.MARIUSMQ_SERVERPORT)

	authHandler := &AuthHandler{ServerPassword: configuration.MARIUSMQ_PASSWORD, AuthChannel: make(chan AuthNotification)}

	broker := &Broker{Queues: map[string]*Queue{}, BrokerChannel: make(chan BrokerNotification), MessageChan: make(chan string)}

	clientHandler := &ClientHandler{Id: uuid.New(), Clients: map[string]*Client{}, AuthHandler: authHandler, Broker: broker, AuthChannel: authHandler.AuthChannel, BrokerChannel: broker.BrokerChannel}

	go clientHandler.AuthChanListen()
	go clientHandler.ListenBrokerNotifications()
	go broker.QueueListen()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		go clientHandler.handleClient(conn)
	}
}
