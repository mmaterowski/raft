package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gomodule/redigo/redis"
)

var done bool
var mu sync.Mutex

func main() {
	handleRequests()

	// RemoveContents("persistence")
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.
	defer conn.Close()

	// Send our command across the connection. The first parameter to
	// Do() is always the name of the Redis command (in this example
	// HMSET), optionally followed by any necessary arguments (in this
	// example the key, followed by the various hash fields and values).
	_, err = conn.Do("HMSET", "album:2", "title", "Electric Ladyland", "artist", "Jimi Hendrix", "price", 4.95, "likes", 8)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Electric Ladyland added!")
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
