package main

import (
	"strings"
)

func RequestVoteRPC(term int, candidateId string, lastLogIndex int, lastLogTerm int) (int, bool) {
	currentTerm := 2
	voteGranted := true
	return currentTerm, voteGranted
}

func AppendEntriesRPC(term int, leaderId string, previousLogIndex int, previousLogTerm int, entries []string, leaderCommitIndex int) (int, bool) {
	currentTerm := 5
	success := true
	WriteToFile(leaderId, strings.Join(entries, ""))
	return currentTerm, success
}
