package server

type ServerType int

const (
	Follower ServerType = iota + 1
	Leader
	Candidate
)
