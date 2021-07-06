package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "raft/raft_rpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHelloAgain(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	message := "Hello again " + in.GetName()
	return &pb.HelloReply{Message: &message}, nil

}
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
