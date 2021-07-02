package main

import "github.com/segmentio/ksuid"

var serverLog []string
var currentTerm int
var votedFor ksuid.KSUID
var serverId string

type ServerType int

const (
	Follower  ServerType = iota + 1 // EnumIndex = 1
	Leader                          // EnumIndex = 2
	Candidate                       // EnumIndex = 3
)

func startServer() {

}
