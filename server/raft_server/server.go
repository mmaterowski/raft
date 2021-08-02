package raft_server

import (
	"context"
	"database/sql"
	"log"
	"sync"

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
	mu sync.Mutex
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
	//Kim is always a leader, change when election module implemented
	if id == consts.KimId {
		stateRebuilt := s.RebuildStateFromLog()
		if !stateRebuilt {
			log.Panic("Couldn't rebuild state")
		}
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

func (s *Server) CommitEntries(leaderCommitIndex int) error {
	if leaderCommitIndex == 0 {
		return nil
	}
	ctx := context.Background()
	s.mu.Lock()
	defer s.mu.Unlock()
	for leaderCommitIndex != s.CommitIndex {
		//TODO Optimize:Get all entries at once
		log.Print("Leader commit index: ", leaderCommitIndex)
		log.Print("Server commit index: ", s.CommitIndex)
		nextEntryIndexToCommit := s.CommitIndex + 1
		entry, err := s.AppRepository.GetEntryAtIndex(ctx, nextEntryIndexToCommit)
		if err != nil {
			log.Println("Error while commiting entry to  state.")
			if err == sql.ErrNoRows {
				log.Print("Follower does not have yet all entries in log")
				break
			}
			log.Println(err)
			if s.ServerType == structs.Leader {
				s.ServerType = structs.Follower
				log.Printf("Server state set Leader->Follower")
			}
			break
		} else {
			log.Print("Commiting new entry to state. Key: ", entry.Key, " Entry: ", entry)
			(*s.State)[entry.Key] = *entry
			s.CommitIndex++

		}

	}
	return nil
}
