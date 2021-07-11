package main

import (
	"context"
	pb "raft/raft_rpc"
)

type rpcServer struct {
	pb.UnimplementedRaftRpcServer
}

func (s *rpcServer) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	entries := mapRaftEntriesToEntries(in.Entries)
	success, lastAppended := server.sqlLiteDb.PersistValues(entries)

	server.previousEntryIndex = lastAppended.Index
	server.previousEntryTerm = lastAppended.TermNumber

	return &pb.AppendEntriesReply{Term: int32(server.previousEntryTerm), Success: success}, nil
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
