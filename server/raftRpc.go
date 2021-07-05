package main

import (
	"encoding/json"
	"fmt"
)

func RequestVoteRPC(term int, candidateId string, lastLogIndex int, lastLogTerm int) (int, bool) {
	currentTerm := 2
	voteGranted := true
	return currentTerm, voteGranted
}

func AppendEntriesRPC(term int, leaderId string, previousLogIndex int, previousLogTerm int, entries map[string]int, leaderCommitIndex int) (int, bool) {
	currentTerm := 5
	success := true
	j, err := json.Marshal(entries)
	fmt.Println(string(j), err)
	return currentTerm, success
}
