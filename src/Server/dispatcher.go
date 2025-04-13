package main

import (
	"log"
)

type Dispatcher struct {
	Handlers map[int]FrameHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		Handlers: make(map[int]FrameHandler),
	}
}

func (d *Dispatcher) Register(channelID int, handler FrameHandler) {
	d.Handlers[channelID] = handler
}

func (d *Dispatcher) Dispatch(frame Frame, client *Client) bool {
	handler, ok := d.Handlers[frame.ChannelID]
	if !ok {
		log.Printf("⚠️ No handler registered for channel %d\n", frame.ChannelID)
		return false
	}

	result := handler.Handle(frame)

	if !result {
		switch handler.(type) {
		case *AuthHandler:
			client.Authenticated = false
		}
	}
	return result
}
