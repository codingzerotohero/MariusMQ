// main.go
package main

import (
	"fmt"
	"net"
)

const serverPassword = "pass"
const messageDelimiter = '\n'

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to the server: ", err)
		panic(err)
	}

	defer conn.Close()

	data := []byte("CONNECT" + ";" + serverPassword + string(messageDelimiter))
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending data: ", err)
		return
	}

	subData := []byte("SUBSCRIBE" + ";" + "my-message-queue" + string(messageDelimiter))
	_, err = conn.Write(subData)
	if err != nil {
		fmt.Println("Error sending data: ", err)
	}

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading message", err)
			return
		}
		fmt.Printf("Received: %s\n", buffer[:n])
	}
}
