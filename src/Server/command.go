package main

type Command string

const (
	CONNECT   Command = "CONNECT"
	PUBLISH   Command = "PUBLISH"
	SUBSCRIBE Command = "SUBSCRIBE"
	CREATE    Command = "CREATE"
	DELETE    Command = "DELETE"
)
