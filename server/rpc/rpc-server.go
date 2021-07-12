package rpc

import (
	"context"

	pb "github.com/mmaterowski/raft/raft_rpc"
	s "github.com/mmaterowski/raft/raft_server"
)

type Server struct {
	pb.UnimplementedRaftRpcServer
	s.RaftServer
}

func (s *Server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	entries := mapRaftEntriesToEntries(in.Entries)
	success, lastAppended := s.RaftServer.SqlLiteDb.PersistValues(entries)

	s.RaftServer.PreviousEntryIndex = lastAppended.Index
	s.RaftServer.PreviousEntryTerm = lastAppended.TermNumber

	return &pb.AppendEntriesReply{Term: int32(s.RaftServer.PreviousEntryTerm), Success: success}, nil
}

func mapRaftEntriesToEntries(rpcEntries []*pb.Entry) []Entry {
	var entries []Entry
	for _, raftEntry := range rpcEntries {
		entries = append(entries, (getEntryFromRaftEntry(raftEntry)))
	}
	return entries
}
func getEntryFromRaftEntry(rpcEntry *pb.Entry) Entry {
	entry := Entry{Key: rpcEntry.Key, Index: int(rpcEntry.Index), Value: int(rpcEntry.Value), TermNumber: int(rpcEntry.TermNumber)}
	return entry
}
