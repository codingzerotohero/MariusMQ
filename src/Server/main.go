// main.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type MessageQueue struct {
	queueName string
	password  string
}

type ClientPool struct {
	Connections []net.Conn
}

type Command string

const (
	CONNECT   Command = "CONNECT"
	PUBLISH   Command = "PUBLISH"
	SUBSCRIBE Command = "SUBSCRIBE"
	CREATE    Command = "CREATE"
	DELETE    Command = "DELETE"
)

const addr = "localhost"
const port = "8080"
const serverPassword = "pass@word"
const messageDelimiter = '\n'

var conns []net.Conn

func main() {

	listener, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	defer listener.Close()

	fmt.Println("Server is listening on port: " + port)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	//buffer := make([]byte, 1024)
	reader := bufio.NewReader(conn)

	fmt.Println("Client connected: ", conn.LocalAddr())

	for {
		//n, err := conn.Read(buffer)
		n, err := reader.ReadString(messageDelimiter)
		if err != nil {
			fmt.Println("Error reading message", err)
			return
		}
		n = strings.TrimSpace(n)
		fmt.Printf("Client message: %s\n", n)

		var messageArr = strings.Split(n, ";")

		if len(messageArr) > 2 {
			fmt.Println("Invalid message format!")
			continue
		}

		cmd := Command(messageArr[0])
		value := messageArr[1]

		fmt.Println(cmd)

		switch cmd {
		case CONNECT:
			fmt.Println("Handle connection request")
			if verifyPassword(value) {
				fmt.Println("Client connection request accepted - " + conn.RemoteAddr().String())
				conns = append(conns, conn)
				data := []byte("CONNECT REQUEST ACCEPTED")
				_, err = conn.Write(data)
				if err != nil {
					fmt.Println("Error sending data: ", err)
					return
				}

			} else {
				data := []byte("CONNECT REQUEST DENIED: INCORRECT PASSWORD.")
				_, err = conn.Write(data)
				if err != nil {
					fmt.Println("Error sending data: ", err)
					return
				}
			}
		case CREATE:
			fmt.Println("Handle create queue request")
		default:
			fmt.Println("Something else" + cmd)
		}
	}
}

func verifyPassword(password string) bool {
	return password == serverPassword
}

func handleCreateQueue() {

}

func handleSubscribeQueue() {

}

func handlePublishMessage() {

}
