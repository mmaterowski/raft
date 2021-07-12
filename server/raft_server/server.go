package raft_server

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	persistence "github.com/mmaterowski/raft/persistence"
)

type ServerType int

type RaftServer struct {
	ServerType
	State              map[string]persistence.Entry
	CurrentTerm        int
	PreviousEntryIndex int
	PreviousEntryTerm  int
	CommitIndex        int
	Id                 string
	VotedFor           string
	SqlLiteDb          persistence.SqlLiteDb
}

const (
	Follower ServerType = iota + 1
	Leader
	Candidate
)

func (s RaftServer) StartServer(id string, debug bool) {
	s.Id = os.Getenv("SERVER_ID")
	if s.Id == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}
	s.State = make(map[string]persistence.Entry)
	s.PreviousEntryIndex = -1
	s.PreviousEntryTerm = -1
	s.CommitIndex = -1
	log.Print("Starting server...")
	s.SqlLiteDb = persistence.NewDb(debug)
	s.VotedFor = s.SqlLiteDb.GetVotedFor()
	s.CurrentTerm = s.SqlLiteDb.GetCurrentTerm()
	stateRebuilt := s.RebuildStateFromLog()
	s.SqlLiteDb.PersistValue("d", 23, 20)
	if !stateRebuilt {
		log.Panic("Couldn't rebuild state")
	}

	//setElectionTimer?

}

func (s RaftServer) RebuildStateFromLog() bool {
	entries := s.SqlLiteDb.GetLog()
	for _, entry := range entries {
		s.State[entry.Key] = entry
	}
	return true
}
