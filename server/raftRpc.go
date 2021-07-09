package main

import (
	"context"
	pb "raft/raft_rpc"
)

type server struct {
	pb.UnimplementedRaftRpcServer
}

func (s *server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	entries := mapRaftEntriesToEntries(in.Entries)
	success, lastAppended := PersistValues(entries)

	previousEntryIndex = lastAppended.Index
	previousEntryTerm = lastAppended.TermNumber

	return &pb.AppendEntriesReply{Term: int32(previousEntryTerm), Success: success}, nil
}

func mapRaftEntriesToEntries(rpcEntries []*pb.Entry) []Entry {
	var entries []Entry
	for _, raftEntry := range rpcEntries {
		entries = append(entries, (getEntryFromRaftEntry(raftEntry)))
	}
	return entries
}
func getEntryFromRaftEntry(rpcEntry *pb.Entry) Entry {
	entry := Entry{Key: rpcEntry.Key, Index: int(rpcEntry.Index), Value: int(rpcEntry.Valuy), TermNumber: int(rpcEntry.TermNumber)}
	return entry
}
