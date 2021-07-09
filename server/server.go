package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"raft/raft_rpc"
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

	if !debug {
		go func() {
			if serverId == "Kim" {
				handleRPC()
			}
		}()
		go func() {
			time.Sleep(5 * time.Second)
			if serverId == "Kim" {
				serverPort := "kim:6960"
				conn, err := grpc.Dial(serverPort, grpc.WithInsecure())
				Check(err)
				client := pb.NewRaftRpcClient(conn)
				entries := []*raft_rpc.Entry{}
				entries = append(entries, &raft_rpc.Entry{
					Index:      2,
					Value:      34,
					Key:        "jebanko",
					TermNumber: 3,
				})
				feature, err := client.AppendEntries(context.Background(), &pb.AppendEntriesRequest{Term: 3, Entries: entries}, grpc.EmptyCallOption{})
				Check(err)
				log.Println(feature)
				defer conn.Close()
			}

		}()
	}

	handleRequests()
	//setElectionTimer?

}

func handleRPC() {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	Check(err)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRaftRpcServer(grpcServer, &server{})
	log.Printf("RPC listening on port: %d", port)
	grpcServer.Serve(lis)
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
