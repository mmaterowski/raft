package structs

type ServerType int

const (
	Follower ServerType = iota + 1
	Leader
	Candidate
)

type Entry struct {
	Index      int
	Value      int
	Key        string
	TermNumber int
}
