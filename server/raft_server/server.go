package raft_server

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"sync"
	"time"

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
	mu                  sync.Mutex
	ElectionTicker      *time.Ticker
	ResetElectionTicker chan struct{}
	HeartbeatTicker     *time.Ticker
	TriggerHeartbeat    chan struct{}
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
	log.Print("Setting voted for and current term to initial values until election improved")
	s.AppRepository.SetVotedFor(context.Background(), "")
	s.AppRepository.SetCurrentTerm(context.Background(), consts.TermUninitializedValue)
	s.VotedFor, _ = s.AppRepository.GetVotedFor(context.Background())
	s.CurrentTerm, _ = s.AppRepository.GetCurrentTerm(context.Background())
	s.ServerType = structs.Follower

	s.ResetElectionTicker = make(chan struct{})
	s.TriggerHeartbeat = make(chan struct{})
	s.HeartbeatTicker = time.NewTicker(consts.HeartbeatInterval)
	seed := rand.NewSource(time.Now().UnixNano())
	electionTimeout := rand.New(seed).Intn(100)*300 + 100
	log.Printf("Election timeout set to: %d + %d", electionTimeout, consts.HeartbeatInterval)
	s.ElectionTicker = time.NewTicker(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)
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
