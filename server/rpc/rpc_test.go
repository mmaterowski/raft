package rpc

import (
	"context"
	"testing"

	"github.com/mmaterowski/raft/persistence"
	pb "github.com/mmaterowski/raft/raft_rpc"
)

func TestAppendFailsIfLeadersTermLowerThanCurrentTerm(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}

	inMemContext.SetCurrentTerm(context.Background(), 10)

	s.StartServer("TestServ")

	request := pb.AppendEntriesRequest{Term: 3}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if reply.Success {
		t.Errorf("Expected false reply: %s", reply)
	}
}

func TestAppendSuccessIfLeadersTermHigherThanCurrentTerm(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}

	inMemContext.SetCurrentTerm(context.Background(), 1)

	s.StartServer("TestServ")

	request := pb.AppendEntriesRequest{Term: 10}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if !reply.Success {
		t.Errorf("Expected success reply: %s", reply)
	}
}

func TestAppendSuccessIfNoEntriesToAppend(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}
	inMemContext.SetCurrentTerm(context.Background(), 1)

	s.StartServer("TestServ")

	request := pb.AppendEntriesRequest{Term: 10, Entries: []*pb.Entry{}}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if !reply.Success {
		t.Errorf("Expected success reply: %s", reply)
	}
}

func TestAppendDoNotFailIfMoreThanOneEntryInRequest(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}
	inMemContext.SetCurrentTerm(context.Background(), 1)
	s.StartServer("TestServ")
	entries := make([]*pb.Entry, 2)
	request := pb.AppendEntriesRequest{Term: 10, Entries: entries}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if !reply.Success {
		t.Errorf("Expected success reply: %s", reply)
	}
}

func TestLastEntryFoundButDoesNotMatchWithLeaderTerm(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}
	crrentTerm, _ := inMemContext.GetCurrentTerm(context.Background())
	inMemContext.SetCurrentTerm(context.Background(), 1)
	inMemContext.PersistValue(context.Background(), "A", 2, crrentTerm)
	inMemContext.PersistValue(context.Background(), "B", 3, crrentTerm)

	s.StartServer("TestServ")

	entries := make([]*pb.Entry, 1)
	prevLogIndex := 1
	prevLogTerm := 2
	leaderTerm := 32

	request := pb.AppendEntriesRequest{Term: int32(leaderTerm), Entries: entries, PreviousLogIndex: int32(prevLogIndex), PreviousLogTerm: int32(prevLogTerm)}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if reply.Success {
		t.Errorf("Expected failed reply: %s", reply)
	}
}

func TestLastEntryFoundAndMatchesWithLeaderTerm(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}
	inMemContext.SetCurrentTerm(context.Background(), 1)
	crrentTerm, _ := inMemContext.GetCurrentTerm(context.Background())

	inMemContext.PersistValue(context.Background(), "A", 2, crrentTerm)
	inMemContext.PersistValue(context.Background(), "B", 3, crrentTerm)

	s.StartServer("TestServ")

	entries := make([]*pb.Entry, 1)
	entries[0] = &pb.Entry{Index: 2, Value: 10, Key: "z", TermNumber: 10}
	prevLogIndex := 1
	prevLogTerm := 1
	leaderTerm := 32

	request := pb.AppendEntriesRequest{Term: int32(leaderTerm), Entries: entries, PreviousLogIndex: int32(prevLogIndex), PreviousLogTerm: int32(prevLogTerm)}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if !reply.Success {
		t.Errorf("Expected failed reply: %s", reply)
	}
}

func TestLastEntryOnFollowerDoesNotExist(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.AppRepository = persistence.Db{AppRepository: &inMemContext}
	inMemContext.SetCurrentTerm(context.Background(), 1)
	crrentTerm, _ := inMemContext.GetCurrentTerm(context.Background())

	inMemContext.PersistValue(context.Background(), "A", 2, crrentTerm)
	inMemContext.PersistValue(context.Background(), "B", 3, crrentTerm)

	s.StartServer("TestServ")

	entries := make([]*pb.Entry, 1)
	prevLogIndex := 3
	prevLogTerm := 5
	leaderTerm := 32

	request := pb.AppendEntriesRequest{Term: int32(leaderTerm), Entries: entries, PreviousLogIndex: int32(prevLogIndex), PreviousLogTerm: int32(prevLogTerm)}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if reply.Success {
		t.Errorf("Expected failed reply: %s", reply)
	}
}
