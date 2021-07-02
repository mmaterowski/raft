package main

import (
	"sync"

	"github.com/segmentio/ksuid"
)

var done bool
var mu sync.Mutex

func main() {
	RemoveContents("persistence")
	id := ksuid.New()
	currentTerm := 3
	lastLogIndex := 0
	lastLogTerm := 2
	term, voteGranted := RequestVoteRPC(currentTerm, id, lastLogIndex, lastLogTerm)

	previousLogIndex := 3
	previousLogTerm := 8
	entries := []string{"a", "b"}
	leaderCommitIndex := 7
	appendTerm, success := AppendEntriesRPC(currentTerm, id, previousLogIndex, previousLogTerm, entries, leaderCommitIndex)

	print(appendTerm, success)
	print("\n")
	print(term, voteGranted)
}
