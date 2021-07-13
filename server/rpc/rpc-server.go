package rpc

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/helpers"
	pb "github.com/mmaterowski/raft/raft_rpc"
	s "github.com/mmaterowski/raft/raft_server"
	structs "github.com/mmaterowski/raft/structs"
)

type Server struct {
	pb.UnimplementedRaftRpcServer
	s.Server
}

func (s *Server) CommitAvailableEntries(ctx context.Context, in *pb.CommitAvailableEntriesRequest) (*pb.Empty, error) {
	leaderCommitIndex := int(in.LeaderCommitIndex)
	if s.Server.CommitIndex > leaderCommitIndex {
		log.Println("Got request to commit state, but current commit index is greated than received")
	}
	for leaderCommitIndex != s.Server.CommitIndex {
		//TODO Optimize:Get all entries at once
		nextEntryIndexToCommit := s.Server.CommitIndex + 1
		entry, err := s.Server.GetEntryAtIndex(nextEntryIndexToCommit)
		if err != nil {
			log.Println("Error while commiting state:")
			log.Println(err)
			if s.ServerType == structs.Leader {
				s.ServerType = structs.Follower
				log.Printf("Server state set Leader->Follower")
			}
			return &pb.Empty{}, err
		}

		s.Server.State[entry.Key] = entry
		s.Server.CommitIndex++
	}

	return &pb.Empty{}, nil

}

func (s *Server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	if len(in.Entries) == 0 {
		return &pb.AppendEntriesReply{Term: int32(s.Server.PreviousEntryTerm), Success: true}, nil
	}

	//TODO Handle no previous logs case
	if in.PreviousLogIndex != 0 {
		entry, err := s.Server.Context.GetEntryAtIndex(int(in.PreviousLogIndex))
		if err != nil {
			log.Println("Follower is missing last entry present on Leader. Request:")
			log.Println(helpers.PrettyPrint(in))
			return &pb.AppendEntriesReply{Term: int32(s.Server.PreviousEntryTerm), Success: false}, nil
		}

		if entry.TermNumber != int(in.PreviousLogTerm) {
			log.Printf("Follower has different term number on the entry with index '%d'. Entry term: '%d', but expected '%d'", in.PreviousLogIndex, entry.TermNumber, in.PreviousLogTerm)
			log.Println(helpers.PrettyPrint(in))
			return &pb.AppendEntriesReply{Term: int32(s.Server.PreviousEntryTerm), Success: false}, nil
		}
	}

	entries := mapRaftEntriesToEntries(in.Entries)
	success, lastAppended := s.Server.Context.PersistValues(entries)

	s.Server.PreviousEntryIndex = lastAppended.Index
	s.Server.PreviousEntryTerm = lastAppended.TermNumber

	return &pb.AppendEntriesReply{Term: int32(s.Server.PreviousEntryTerm), Success: success}, nil
}

func mapRaftEntriesToEntries(rpcEntries []*pb.Entry) []structs.Entry {
	var entries []structs.Entry
	for _, raftEntry := range rpcEntries {
		entries = append(entries, (getEntryFromRaftEntry(raftEntry)))
	}
	return entries
}
func getEntryFromRaftEntry(rpcEntry *pb.Entry) structs.Entry {
	entry := structs.Entry{Key: rpcEntry.Key, Index: int(rpcEntry.Index), Value: int(rpcEntry.Value), TermNumber: int(rpcEntry.TermNumber)}
	return entry
}
