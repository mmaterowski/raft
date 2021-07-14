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

	request := pb.AppendEntriesRequest{Term: 10}
	reply, _ := s.AppendEntries(context.Background(), &request)
	if reply.Success {
		t.Errorf("Expected false reply: %s", reply)
	}
}
