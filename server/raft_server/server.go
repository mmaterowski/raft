package raft_server

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/consts"
	"github.com/mmaterowski/raft/entry"
	. "github.com/mmaterowski/raft/persistence"
	"github.com/mmaterowski/raft/structs"
)

type Server struct {
	structs.ServerType
	State              *map[string]entry.Entry
	CurrentTerm        int
	PreviousEntryIndex int
	PreviousEntryTerm  int
	CommitIndex        int
	Id                 string
	VotedFor           string
	AppRepository
}

func (s *Server) StartServer(id string) {
	s.Id = id
	s.ServerType = structs.ServerType(structs.Candidate)
	state := make(map[string]entry.Entry)
	s.State = &state
	s.PreviousEntryIndex = consts.NoPreviousEntryValue
	s.PreviousEntryTerm = consts.TermInitialValue
	s.CommitIndex = consts.LeaderCommitInitialValue
	log.Print("Starting server...")
	s.VotedFor, _ = s.AppRepository.GetVotedFor(context.Background())
	s.CurrentTerm, _ = s.AppRepository.GetCurrentTerm(context.Background())
	stateRebuilt := s.RebuildStateFromLog()
	if !stateRebuilt {
		log.Panic("Couldn't rebuild state")
	}

	//setElectionTimer?

}

func (s *Server) RebuildStateFromLog() bool {
	entries, _ := s.AppRepository.GetLog(context.Background())
	for _, entry := range *entries {
		(*s.State)[entry.Key] = entry
		s.CommitIndex = entry.Index
	}
	log.Print("Log successfully rebuilt. Entries: ", entries)
	return true
}
