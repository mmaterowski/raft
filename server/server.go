package main

import (
	"fmt"
	"log"
	"net"
	"os"

	pb "raft/raft_rpc"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

var serverLog []string
var currentTerm int
var votedFor string
var serverId string
var state map[string]Entry = make(map[string]Entry)
var serverType ServerType
var debug = false
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
		err := handleRPC()
		Check(err)
	}()

	go func() {
		SetupRpcClient()
	}()

	handleRequests()
	//setElectionTimer?

}

func handleRPC() error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	Check(err)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRaftRpcServer(grpcServer, &server{})
	log.Printf("RPC listening on port: %d", port)
	return grpcServer.Serve(lis)
}

func identifyServer() {
	serverId = os.Getenv("SERVER_ID")
	if debug {
		serverId = "Kim"
		others = append(others, laszloId, rickyId)
	}
	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}

	switch serverId {
	case "Kim":
		others = append(others, laszloId, rickyId)
	case "Ricky":
		others = append(others, laszloId, kimId)
	case "Laszlo":
		others = append(others, rickyId, kimId)
	default:
		log.Panic("Couldn't identify server")
	}

}
