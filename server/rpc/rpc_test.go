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
	s.Context = persistence.Db{Context: &inMemContext}

	inMemContext.SetCurrentTerm(10)

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
	s.Context = persistence.Db{Context: &inMemContext}

	inMemContext.SetCurrentTerm(1)

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
	s.Context = persistence.Db{Context: &inMemContext}
	inMemContext.SetCurrentTerm(1)

	s.StartServer("TestServ")

	request := pb.AppendEntriesRequest{Term: 10, Entries: []*pb.Entry{}}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if !reply.Success {
		t.Errorf("Expected success reply: %s", reply)
	}
}

func TestAppendFailsIfMoreThanOneEntryInRequest(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.Context = persistence.Db{Context: &inMemContext}
	inMemContext.SetCurrentTerm(1)
	s.StartServer("TestServ")
	entries := make([]*pb.Entry, 2)
	request := pb.AppendEntriesRequest{Term: 10, Entries: entries}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if reply.Success {
		t.Errorf("Expected failed reply: %s", reply)
	}
}

func TestLastEntryDoesNotMatchWithLeader(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.Context = persistence.Db{Context: &inMemContext}
	inMemContext.SetCurrentTerm(1)
	inMemContext.PersistValue("A", 2, inMemContext.GetCurrentTerm())
	inMemContext.PersistValue("B", 3, inMemContext.GetCurrentTerm())

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

func TestLastEntryOnFollowerDoesNotExist(t *testing.T) {
	inMemContext := persistence.InMemoryContext{}
	s := Server{}
	s.Context = persistence.Db{Context: &inMemContext}
	inMemContext.SetCurrentTerm(1)
	inMemContext.PersistValue("A", 2, inMemContext.GetCurrentTerm())
	inMemContext.PersistValue("B", 3, inMemContext.GetCurrentTerm())

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
