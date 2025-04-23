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

	if f.Command == "CONNECT" && f.Password == a.ServerPassword {
		log.Println("✅ Auth success")

		a.SendAcceptDenyNotification(true, "", f.ClientIp)
	} else {
		log.Println("❌ Auth failed")
		a.SendAcceptDenyNotification(false, INVALID_PASSWORD, f.ClientIp)
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
