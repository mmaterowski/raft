package raftlog

import (
	"sync"

	. "github.com/mmaterowski/raft/persistence"
)

type RaftLogService struct {
	appRepo AppRepository
	mu      sync.Mutex
}

func NewRaftLogService(repo AppRepository) *RaftLogService {
	return &RaftLogService{appRepo: repo}
}
