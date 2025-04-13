// server.go
package main

import (
	"fmt"
	"net"

	"github.com/google/uuid"
)

var conns []net.Conn

func startServer(configuration Configuration) {

	listener, err := net.Listen("tcp", configuration.MARIUSMQ_SERVERADDRESS+":"+configuration.MARIUSMQ_SERVERPORT)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	defer listener.Close()

	fmt.Println("Server is listening on port: " + configuration.MARIUSMQ_SERVERPORT)

	clientHandler := ClientHandler{Id: uuid.New(), AuthHandler: AuthHandler{ServerPassword: configuration.MARIUSMQ_PASSWORD}, Dispatcher: *NewDispatcher() /*, QueueHandler: QueueHandler{MessageQueues: map[string]MessageQueue{}}*/}
	clientHandler.Dispatcher.Register(0, &clientHandler.AuthHandler)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}
		go clientHandler.handleClient(conn)
	}
}

func handleCreateQueue() {

}

func handleSubscribeQueue() {

}

func handlePublishMessage() {

}
