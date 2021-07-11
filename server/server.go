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

var debug = true

type serverType int

type raftServer struct {
	rpcClient
	serverType
	state              map[string]Entry
	currentTerm        int
	previousEntryIndex int
	previousEntryTerm  int
	commitIndex        int
	id                 string
	votedFor           string
	sqlLiteDb
}

type Entry struct {
	Index      int
	Value      int
	Key        string
	TermNumber int
}

const (
	Follower serverType = iota + 1
	Leader
	Candidate
)

func (s raftServer) startServer(id string) {
	s.state = make(map[string]Entry)
	s.previousEntryIndex = -1
	s.previousEntryTerm = -1
	s.commitIndex = -1
	log.Print("Starting server...")
	s.sqlLiteDb = NewDb()

	identifyServer()
	s.votedFor = s.sqlLiteDb.GetVotedFor()
	s.currentTerm = s.sqlLiteDb.GetCurrentTerm()
	stateRebuilt := s.RebuildStateFromLog()
	s.sqlLiteDb.PersistValue("d", 23, 20)
	if !stateRebuilt {
		log.Panic("Couldn't rebuild state")
	}

	go func() {
		err := handleRPC()
		Check(err)
	}()

	go func() {
		s.rpcClient.SetupRpcClient()
	}()

	handleRequests()
	//setElectionTimer?

}

func (s raftServer) RebuildStateFromLog() bool {
	entries := s.sqlLiteDb.GetLog()
	for _, entry := range entries {
		s.state[entry.Key] = entry
	}
	return true
}

func handleRPC() error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	Check(err)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRaftRpcServer(grpcServer, &rpcServer{})
	log.Printf("RPC listening on port: %d", port)
	return grpcServer.Serve(lis)
}

func identifyServer() {
	server.id = os.Getenv("SERVER_ID")
	if debug {
		server.id = "Kim"
		others = append(others, laszloId, rickyId)
	}
	if server.id == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}

	switch server.id {
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
