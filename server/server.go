package main

import (
	"log"
	"os"

	"github.com/segmentio/ksuid"
)

var serverLog []string
var currentTerm int
var votedFor ksuid.KSUID
var serverId string
var stateMachine map[string]int
var serverType ServerType

type ServerType int

const (
	Follower  ServerType = iota + 1 // EnumIndex = 1
	Leader                          // EnumIndex = 2
	Candidate                       // EnumIndex = 3
)

func startServer(id string) {
	log.Print("Starting server...")
	setServerIdFromEnv()
	connected := ConnectToRedis()
	if connected {
		log.Print("Connected to redis")
	}

	handleRequests()
	//getLogFromPersistence
	//rebuildStateServerState
	//setElectionTimer?

}

func setServerIdFromEnv() {
	serverId = os.Getenv("SERVER_ID")
	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}
}
