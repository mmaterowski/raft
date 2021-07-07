package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	pb "raft/raft_rpc"
	"sync"

	"google.golang.org/grpc"
)

func handleRPC() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 6960))
	Check(err)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRaftRpcServer(grpcServer, pb.UnimplementedRaftRpcServer{})
	grpcServer.Serve(lis)
}

type server struct {
	pb.UnimplementedRaftRpcServer
	mu sync.Mutex
}

func (s *server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	return &pb.AppendEntriesReply{Term: 3, Success: true}, nil
}

type RequestVoteRequest struct {
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
