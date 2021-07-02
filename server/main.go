package main

import (
	"sync"
)

var done bool
var mu sync.Mutex

func main() {
	startServer()
	server1 := "Ricky"
	server2 := "Laszlo"
	server3 := "Kim"

	currentTerm := 3
	lastLogIndex := 0
	lastLogTerm := 2
	term, voteGranted := RequestVoteRPC(currentTerm, server1, lastLogIndex, lastLogTerm)

	previousLogIndex := 3
	previousLogTerm := 8
	entries := make(map[string]int)
	entries["a"] = 3
	entries["b"] = 5
	leaderCommitIndex := 7
	appendTerm, success := AppendEntriesRPC(currentTerm, server1, previousLogIndex, previousLogTerm, entries, leaderCommitIndex)
	AppendEntriesRPC(currentTerm, server2, previousLogIndex, previousLogTerm, entries, leaderCommitIndex)
	AppendEntriesRPC(currentTerm, server3, previousLogIndex, previousLogTerm, entries, leaderCommitIndex)

	print(appendTerm, success)
	print("\n")
	print(term, voteGranted)
}