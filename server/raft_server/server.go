package raft_server

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/mmaterowski/raft/persistence"
	. "github.com/mmaterowski/raft/structs"
)

type Server struct {
	ServerType
	State              map[string]Entry
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
	s.ServerType = ServerType(Candidate)
	s.State = make(map[string]Entry)
	s.PreviousEntryIndex = -1
	s.PreviousEntryTerm = -1
	s.CommitIndex = -1
	log.Print("Starting server...")
	s.VotedFor = s.Context.GetVotedFor()
	s.CurrentTerm = s.Context.GetCurrentTerm()
	stateRebuilt := s.RebuildStateFromLog()
	s.Context.PersistValue("d", 23, 20)
	if !stateRebuilt {
		log.Panic("Couldn't rebuild state")
	}

	//setElectionTimer?

}

func (s Server) RebuildStateFromLog() bool {
	entries := s.Context.GetLog()
	for _, entry := range entries {
		s.State[entry.Key] = entry
	}
	return true
}
