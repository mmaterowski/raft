package main

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var serverLog []string
var currentTerm int
var votedFor string
var serverId string
var state map[string]Entry = make(map[string]Entry)
var serverType ServerType
var debug = true
var previousEntryIndex int = -1
var previousEntryTerm int = -1
var commitIndex int = -1

type ServerType int

const (
	Follower ServerType = iota + 1
	Leader
	Candidate
)

func startServer(id string) {
	log.Print("Starting server...")
	success := SetupDB()
	if !success {
		log.Panic("Db not initialized properly")
	}

	identifyServer()
	votedFor = GetVotedFor()
	currentTerm = GetCurrentTerm()
	stateRebuilt := RebuildStateFromLog()
	PersistValue("d", 23, 20)
	if !stateRebuilt {
		log.Panic("Couldn't rebuild state")
	}

	go func() {
		handleRPC()
	}()
	go func() {
		handleRequests()
	}()

	//setElectionTimer?

}

func identifyServer() {
	serverId = os.Getenv("SERVER_ID")
	if debug {
		serverId = "Kim"
		others = append(others, laszloAddress, rickyAddress)
	}
	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}

	switch serverId {
	case "Kim":
		others = append(others, laszloAddress, rickyAddress)
	case "Ricky":
		others = append(others, laszloAddress, kimAddress)
	case "Laszlo":
		others = append(others, rickyAddress, kimAddress)
	default:
		log.Panic("Couldn't identify server")
	}

}
