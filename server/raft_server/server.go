package raft_server

import (
	"context"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmaterowski/raft/entry"
	. "github.com/mmaterowski/raft/persistence"
	"github.com/mmaterowski/raft/structs"
)

type Server struct {
	structs.ServerType
	State              map[string]entry.Entry
	CurrentTerm        int
	PreviousEntryIndex int
	PreviousEntryTerm  int
	CommitIndex        int
	Id                 string
	VotedFor           string
	Context
}

func (s *Server) StartServer(id string) {
	s.Id = id
	s.ServerType = structs.ServerType(structs.Candidate)
	s.State = make(map[string]entry.Entry)
	s.PreviousEntryIndex = -1
	s.PreviousEntryTerm = -1
	s.CommitIndex = -1
	log.Print("Starting server...")
	s.VotedFor, _ = s.Context.GetVotedFor(context.Background())
	s.CurrentTerm, _ = s.Context.GetCurrentTerm(context.Background())
	stateRebuilt := s.RebuildStateFromLog()
	if !stateRebuilt {
		log.Panic("Couldn't rebuild state")
	}

	//setElectionTimer?

}

func (s Server) RebuildStateFromLog() bool {
	entries, _ := s.Context.GetLog(context.Background())
	for _, entry := range *entries {
		s.State[entry.Key] = entry
	}
	return true
}
