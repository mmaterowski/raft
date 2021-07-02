package main

import (
	"sync"

	"github.com/segmentio/ksuid"
)

var done bool
var mu sync.Mutex

func main() {
	id := ksuid.New()
	currentTerm := 3
	lastLogIndex := 0
	lastLogTerm := 2
	term, voteGranted := requestVoteRPC(currentTerm, id, lastLogIndex, lastLogTerm)
	previousLogIndex := 3
	previousLogTerm := 8
	entries := []string{"a", "b"}
	leaderCommitIndex := 7

	appendTerm, success := appendEntriesRPC(currentTerm, id, previousLogIndex, previousLogTerm, entries, leaderCommitIndex)
	print(appendTerm, success)
	print(term, voteGranted)
}

func requestVoteRPC(term int, candidateId ksuid.KSUID, lastLogIndex int, lastLogTerm int) (int, bool) {
	currentTerm := 2
	voteGranted := true
	return currentTerm, voteGranted
}

func appendEntriesRPC(term int, leaderId ksuid.KSUID, previousLogIndex int, previousLogTerm int, entries []string, leaderCommitIndex int) (int, bool) {
	currentTerm := 5
	success := true
	return currentTerm, success
}
