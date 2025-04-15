package main

import (
	"log"
	"sync"
)

type AuthHandler struct {
	ServerPassword string
	ClientHandler  *ClientHandler
	mu             sync.Mutex
}

func (a *AuthHandler) Handle(f Frame, client *Client) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if f.Command == "CONNECT" && f.Payload == a.ServerPassword {
		log.Println("✅ Auth success")
		data := []byte("CONNECTED" + string('\n'))
		_, err := client.Connection.Write(data)
		if err != nil {
			log.Println("Error writing data to client:", err)
		}
		client.Channels[f.ChannelID] = &Channel{
			Id:     f.ChannelID,
			Inbox:  []string{},
			Outbox: []string{},
		}
		client.Authenticated = true
		a.ClientHandler.Clients[client.IPAddress] = client
	} else {
		log.Println("❌ Auth failed")
		data := []byte("Could not authenticate - invalid password!")
		_, err := client.Connection.Write(data)
		if err != nil {
			log.Println("Error writing data to client:", err)
		}
		client.Authenticated = false
		client.Connection.Close()
	}
}
