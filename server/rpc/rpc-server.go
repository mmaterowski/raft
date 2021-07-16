package rpc

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/entry"
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
		entry, err := s.Server.GetEntryAtIndex(ctx, nextEntryIndexToCommit)
		if err != nil {
			log.Println("Error while commiting state:")
			log.Println(err)
			if s.ServerType == structs.Leader {
				s.ServerType = structs.Follower
				log.Printf("Server state set Leader->Follower")
			}
			return &pb.Empty{}, err
		}

		s.Server.State[entry.Key] = *entry
		s.Server.CommitIndex++
	}

	return &pb.Empty{}, nil

}

func (s *Server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	//TODO not sure which term should i return
	term := int32(666)
	successReply := &pb.AppendEntriesReply{Success: true, Term: term}
	failReply := &pb.AppendEntriesReply{Success: false, Term: term}

	if in.Term < int32(s.Server.CurrentTerm) {
		return failReply, nil
	}

	if len(in.Entries) == 0 {
		return successReply, nil
	}

	if len(in.Entries) != 1 {
		log.Printf("Handling only single entry append for now!")
		return failReply, nil
	}

	if in.PreviousLogIndex != 0 {
		entry, err := s.Server.AppRepository.GetEntryAtIndex(ctx, int(in.PreviousLogIndex))
		if entry.IsEmpty() {
			log.Print(err)
			return failReply, nil
		}

		if entry.TermNumber != int(in.PreviousLogTerm) {
			err := s.Server.AppRepository.DeleteAllEntriesStartingFrom(ctx, int(in.PreviousLogIndex))
			if err != nil {
				log.Panic("Found conflicting entry, but couldn't delete. That's bad, I panic")
			}
			log.Printf("Follower has different term number on the entry with index '%d'. Entry term: '%d', but expected '%d'", in.PreviousLogIndex, entry.TermNumber, in.PreviousLogTerm)
			log.Println(helpers.PrettyPrint(in))
			return failReply, nil
		}
	}

	entries := mapRaftEntriesToEntries(in.Entries)

	//Do I really have to check it every time, or consistency check does this for me?
	entry, _ := s.Server.AppRepository.GetEntryAtIndex(ctx, entries[0].Index)
	if !entry.IsEmpty() {
		log.Printf("Tried to append entry that already exists.")
		return failReply, nil
	}

	lastAppendedEntry, err := s.Server.AppRepository.PersistValues(ctx, entries)

	if err != nil {
		log.Printf("Unexpected error. Failed to persist entries")
		return failReply, nil
	}

	if in.LeaderCommitIndex > int32(s.Server.CommitIndex) {
		s.Server.CommitIndex = helpers.Min(int(in.LeaderCommitIndex), lastAppendedEntry.TermNumber)
	}

	s.Server.PreviousEntryIndex = lastAppendedEntry.Index
	s.Server.PreviousEntryTerm = lastAppendedEntry.TermNumber

	return successReply, nil
}

func mapRaftEntriesToEntries(rpcEntries []*pb.Entry) []entry.Entry {
	var entries []entry.Entry
	for _, raftEntry := range rpcEntries {
		entries = append(entries, (getEntryFromRaftEntry(raftEntry)))
	}
	return entries
}
func getEntryFromRaftEntry(rpcEntry *pb.Entry) entry.Entry {
	entry := entry.Entry{Key: rpcEntry.Key, Index: int(rpcEntry.Index), Value: int(rpcEntry.Value), TermNumber: int(rpcEntry.TermNumber)}
	return entry
}
