package main

import (
	"log"
	"os"
)

const defaultPassword = "pass"

func main() {
	log.Println("** WELCOME TO MariusMQ Server - the newest, fastest and coolest message broker on the planet! **")
	log.Println("Initializing server config and setting up services/dependencies")

	var configuration Configuration

	value, envExists := os.LookupEnv("MARIUSMQ_PASSWORD")

	if envExists && value != "" {
		configuration.MARIUSMQ_PASSWORD = value
		log.Println("Server has been set to use an overriden password through environment variable")
	} else {
		log.Println("Server has been set to use the default password")
		configuration.MARIUSMQ_PASSWORD = defaultPassword
	}

	startServer(configuration)
}

type Configuration struct {
	MARIUSMQ_PASSWORD      string
	MARIUSMQ_SERVERPORT    string
	MARIUSMQ_SERVERADDRESS string
}
