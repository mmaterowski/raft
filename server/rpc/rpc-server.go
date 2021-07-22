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
	log.Printf("Received request to commit entry")
	leaderCommitIndex := int(in.LeaderCommitIndex)
	if s.Server.CommitIndex > leaderCommitIndex {
		log.Println("Got request to commit state, but current commit index is greated than received")
	}
	log.Printf("Leader commit index: %d, server commit index: %d", leaderCommitIndex, s.Server.CommitIndex)
	for leaderCommitIndex != s.Server.CommitIndex {
		//TODO Optimize:Get all entries at once
		nextEntryIndexToCommit := s.Server.CommitIndex + 1
		last, _ := s.Server.AppRepository.GetLastEntry(ctx)
		log.Print("Last entry in log: ", last)
		entry, err := s.Server.AppRepository.GetEntryAtIndex(ctx, nextEntryIndexToCommit)
		if err != nil {
			log.Println("Error while commiting state:")
			log.Println(err)
			if s.ServerType == structs.Leader {
				s.ServerType = structs.Follower
				log.Printf("Server state set Leader->Follower")
			}
			return &pb.Empty{}, err
		}
		log.Print("Commiting new entry to state. Key: ", entry.Key, " Entry: ", entry)
		log.Print("Server state before:", (*s.Server.State))
		(*s.Server.State)[entry.Key] = *entry
		log.Print("Server state after:", (*s.Server.State))

		s.Server.CommitIndex++
	}

	return &pb.Empty{}, nil

}

func (s *Server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	//TODO not sure which term should i return
	log.Print("Appending entries request received")
	term := int32(666)
	successReply := &pb.AppendEntriesReply{Success: true, Term: term}
	failReply := &pb.AppendEntriesReply{Success: false, Term: term}

	log.Printf("Server current term: %d. Request term: %d", s.Server.CurrentTerm, in.Term)
	if in.Term < int32(s.Server.CurrentTerm) {
		return failReply, nil
	}

	if len(in.Entries) == 0 {
		return successReply, nil
	}
	log.Print("Leader PreviousLogIndex: ", in.PreviousLogIndex)
	if in.PreviousLogIndex != -1 {
		entry, err := s.Server.AppRepository.GetLastEntry(ctx)
		log.Print("Checking previous entry", entry)

		if entry.IsEmpty() {
			log.Print("Expected to find entry, but there's none", err)
			return failReply, nil
		}

		log.Printf("Previous entry term: %d. Previous entry term from leader: %d", entry.TermNumber, int(in.PreviousLogTerm))
		if entry.Index != int(in.PreviousLogIndex) || entry.TermNumber != int(in.PreviousLogTerm) {
			err := s.Server.AppRepository.DeleteAllEntriesStartingFrom(ctx, int(in.PreviousLogIndex))
			if err != nil {
				log.Panic("Found conflicting entry, but couldn't delete it. That's bad, I panic")
			}
			log.Printf("Follower has different term number on the entry with index '%d'. Entry term: '%d', but expected '%d'", in.PreviousLogIndex, entry.TermNumber, in.PreviousLogTerm)
			return failReply, nil
		}
	}
	entries := mapRaftEntriesToEntries(in.Entries)
	log.Print("Conditions satisfied. Persisting entries", entries)
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
	log.Println("Follower: Succesfully appended entries to log")
	log.Println(helpers.PrettyPrint(in))
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
