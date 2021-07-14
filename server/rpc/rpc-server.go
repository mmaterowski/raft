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
	reply := &pb.AppendEntriesReply{}
	successReply := reply
	successReply.Success = true
	failReply := reply
	failReply.Success = false

	//current term or what?
	reply.Term = int32(s.Server.CurrentTerm)

	if in.Term < int32(s.Server.CurrentTerm) {
		return successReply, nil
	}

	if len(in.Entries) == 0 {
		return successReply, nil
	}

	if len(in.Entries) != 1 {
		log.Printf("Handling only single entry append for now!")
		return failReply, nil
	}

	entry, _ := s.Server.Context.GetEntryAtIndex(int(in.PreviousLogIndex))
	match := lastEntryInLogMatchesWithLeader(entry, int(in.PreviousLogIndex), int(in.PreviousLogTerm))
	if !match {
		return failReply, nil
	}

	if in.PreviousLogIndex != 0 {
		entry, err := s.Server.Context.GetEntryAtIndex(int(in.PreviousLogIndex))
		if err != nil {
			log.Println("Follower is missing last entry present on Leader. Request:")
			log.Println(helpers.PrettyPrint(in))

		}

		if entry.TermNumber != int(in.PreviousLogTerm) {
			success := s.Server.Context.DeleteAllEntriesStartingFrom(int(in.PreviousLogIndex))
			if !success {
				log.Panic("Found conflicting entry, but couldn't delete. That's bad, I panic")
			}
			log.Printf("Follower has different term number on the entry with index '%d'. Entry term: '%d', but expected '%d'", in.PreviousLogIndex, entry.TermNumber, in.PreviousLogTerm)
			log.Println(helpers.PrettyPrint(in))
			return failReply, nil
		}
	}

	entries := mapRaftEntriesToEntries(in.Entries)

	//Do I really have to check it every time?
	entryExists, _ := s.Server.Context.GetEntryAtIndex(entries[0].Index)
	if entryExists != (structs.Entry{}) {
		log.Printf("Tried to append entry that already exists.")
		return failReply, nil
	}

	success, lastAppended := s.Server.Context.PersistValues(entries)

	if !success {
		log.Printf("Unexpected error. Failed to persist entries")
		return failReply, nil
	}

	if in.LeaderCommitIndex > int32(s.Server.CommitIndex) {
		s.Server.CommitIndex = helpers.Min(int(in.LeaderCommitIndex), lastAppended.TermNumber)
	}

	s.Server.PreviousEntryIndex = lastAppended.Index
	s.Server.PreviousEntryTerm = lastAppended.TermNumber

	return successReply, nil
}

func lastEntryInLogMatchesWithLeader(entry structs.Entry, leaderPreviousLogIndex int, leaderPreviousLogTerm int) bool {
	//TODO Handle no previous logs case
	if leaderPreviousLogIndex != 0 {
		if entry == (structs.Entry{}) {
			log.Println("Follower is missing last entry present on Leader. Request:")
			return false
		}

		if entry.TermNumber != int(leaderPreviousLogTerm) {
			log.Printf("Follower has different term number on the entry with index '%d'. Entry term: '%d', but expected '%d'", leaderPreviousLogIndex, entry.TermNumber, leaderPreviousLogTerm)
			return false
		}
		return true
	}
	return true
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
