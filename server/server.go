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
var state map[string]int
var serverType ServerType
var debug = true

type ServerType int

const (
	Follower  ServerType = iota + 1 // EnumIndex = 1
	Leader                          // EnumIndex = 2
	Candidate                       // EnumIndex = 3
)

func startServer(id string) {
	log.Print("Starting server...")
	success := SetupDB()
	if !success {
		log.Panic("Db not initialized properly")
	}

	// statement, _ = database.Prepare("INSERT INTO people (firstname, lastname) VALUES (?, ?)")
	// statement.Exec("Nic", "Raboy")
	// rows, _ := database.Query("SELECT id, firstname, lastname FROM people")
	// var _id int
	// var firstname string
	// var lastname string
	// for rows.Next() {
	// 	rows.Scan(&_id, &firstname, &lastname)
	// 	log.Println(strconv.Itoa(_id) + ": " + firstname + " " + lastname)
	// }

	setServerIdFromEnv()
	votedFor = GetVotedFor()
	currentTerm = GetCurrentTerm()
	handleRequests()
	// setCurrentTerm()
	//rebuildStateServerState
	//setElectionTimer?

}

func setServerIdFromEnv() {
	serverId = os.Getenv("SERVER_ID")
	if debug {
		serverId = "Kim"
	}
	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}
}
