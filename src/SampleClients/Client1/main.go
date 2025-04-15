// main.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
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

	reader := bufio.NewReader(conn)

	data := []byte("0" + ";" + "CONNECT" + ";" + serverPassword + string(messageDelimiter))
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending data: ", err)
		return
	}

	for {
		n, err := reader.ReadString(messageDelimiter)
		if err != nil {
			fmt.Println("Error reading message", err)
			return
		}
		n = strings.TrimSpace(n)
		if n == "CONNECTED" {
			break
		} else {
			log.Println(n)
			return
		}
	}

	createData := []byte("1" + ";" + "CREATE" + ";" + "my-message-queue" + string(messageDelimiter))
	log.Println(createData)
	_, err = conn.Write(createData)
	if err != nil {
		fmt.Println("Error sending data: ", err)
	}
	/*
		publishData := []byte("PUBLISH" + ";" + "my-message-queue" + ";" + "this is my message" + string(messageDelimiter))
		_, err = conn.Write(publishData)
		if err != nil {
			fmt.Println("Error sending data: ", err)
		}*/

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

func connect() {

}

func create() {

}
