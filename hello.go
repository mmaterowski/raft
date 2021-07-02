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
	print(term, voteGranted)
}

func requestVoteRPC(term int, candidateId ksuid.KSUID, lastLogIndex int, lastLogTerm int) (int, bool) {
	return 3, true
}
