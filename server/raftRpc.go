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

type AppendEntriesRequest struct {
	Term              int
	LeaderId          string
	PreviousLogIndex  int
	Entries           map[string]Entry
	LeaderCommitIndex int
}

func AppendEntriesRPC(request AppendEntriesRequest, sendTo string) (int, bool) {
	currentTerm := 5
	success := true
	j, err := json.Marshal(request.Entries)
	fmt.Println(string(j), err)
	return currentTerm, success
}
