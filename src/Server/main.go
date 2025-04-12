package main

import (
	"log"
	"os"
)

const defaultPassword = "pass"
const defaultServerPort = "8080"
const defaultServerAddress = "localhost"

func main() {
	log.Println("** WELCOME TO MariusMQ Server - the newest, fastest and coolest message broker on the planet! **")
	log.Println("Initializing server config and setting up services/dependencies")

	var configuration Configuration

	password, passEnvSet := os.LookupEnv("MARIUSMQ_PASSWORD")
	port, portEnvSet := os.LookupEnv("MARIUSMQ_SERVERPORT")
	address, addrEnvSet := os.LookupEnv("MARIUSMQ_SERVERADDRESS")

	if passEnvSet && password != "" {
		configuration.MARIUSMQ_PASSWORD = password
		log.Println("Server has been set to use an overriden password through environment variable")
	} else {
		log.Println("Server has been set to use the default password")
		configuration.MARIUSMQ_PASSWORD = defaultPassword
	}

	if portEnvSet && port != "" {
		configuration.MARIUSMQ_SERVERPORT = port
		log.Println("Server has been set to use an overriden port through environment variable")
	} else {
		log.Println("Server has been set to use the default port")
		configuration.MARIUSMQ_SERVERPORT = defaultServerPort
	}

	if addrEnvSet && address != "" {
		configuration.MARIUSMQ_SERVERADDRESS = address
		log.Println("Server has been set to use an overriden address through environment variable")
	} else {
		log.Println("Server has been set to use the default address")
		configuration.MARIUSMQ_SERVERADDRESS = defaultServerAddress
	}

	startServer(configuration)
}

type Configuration struct {
	MARIUSMQ_PASSWORD      string
	MARIUSMQ_SERVERPORT    string
	MARIUSMQ_SERVERADDRESS string
}
