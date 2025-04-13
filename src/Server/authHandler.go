package main

import "log"

type AuthHandler struct {
	ServerPassword string
}

func (a *AuthHandler) Handle(f Frame) bool {
	if f.Command == "CONNECT" && f.Payload == a.ServerPassword {
		log.Println("✅ Auth success")
		return true
	} else {
		log.Println("❌ Auth failed")
		return false
	}
}
