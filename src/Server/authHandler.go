package main

import (
	"log"
	"sync"
)

type AuthHandler struct {
	ServerPassword string
	AuthChannel    chan AuthNotification
	mu             sync.Mutex
}

func (a *AuthHandler) Handle(f Frame) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if f.Command == "CONNECT" && f.Payload == a.ServerPassword {
		log.Println("✅ Auth success")

		a.SendAcceptDenyNotification(true, "", f.ClientIp)
		/*
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
		*/
	} else {
		log.Println("❌ Auth failed")
		a.SendAcceptDenyNotification(false, INVALID_PASSWORD, f.ClientIp)
		/*
			data := []byte("Could not authenticate - invalid password!")
			_, err := client.Connection.Write(data)
			if err != nil {
				log.Println("Error writing data to client:", err)
			}
			client.Authenticated = false
			client.Connection.Close()
		*/
	}
}

func (a *AuthHandler) SendAcceptDenyNotification(accept bool, reason DenyReason, clientIP string) {
	notification := AuthNotification{
		ClientIP: clientIP,
		Reason:   reason,
	}

	if accept {
		notification.AuthMessage = ACCEPT
	} else {
		notification.AuthMessage = DENY
	}

	a.AuthChannel <- notification

}

type AuthNotification struct {
	AuthMessage AuthVerdict
	ClientIP    string
	Reason      DenyReason
}

type AuthVerdict string

const (
	ACCEPT AuthVerdict = "ACCEPT"
	DENY   AuthVerdict = "DENY"
)

type DenyReason string

const (
	INVALID_PASSWORD DenyReason = "INVALID PASSWORD"
)
