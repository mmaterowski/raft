package server

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/model/entry"
	pb "github.com/mmaterowski/raft/rpc/raft_rpc"
	s "github.com/mmaterowski/raft/server"
	"github.com/mmaterowski/raft/utils/helpers"
)

type Server struct {
	pb.UnimplementedRaftRpcServer
	s.Server
}

func (s *Server) RequestVote(ctx context.Context, in *pb.RequestVoteRequest) (*pb.RequestVoteReply, error) {
	log.Print("________RequestVoteRPC________")
	log.Print("Request: ", in)
	reply := &pb.RequestVoteReply{Term: int32(s.CurrentTerm), VoteGranted: false}

	if in.Term < int32(s.CurrentTerm) {
		log.Print("Issuing candidate term is lower than server")
		return reply, nil
	}

	if in.Term > int32(s.CurrentTerm) {
		log.Print("Request term higher than server's. Resetting voted for.")
		s.CurrentTerm = int(in.Term)
		s.SetCurrentTerm(ctx, int(in.Term))
		s.SetVotedFor(ctx, "")
	}

	log.Print("Voted for: ", s.VotedFor, " Server previous entry index: ", s.PreviousEntryIndex, " previous entry term: ", s.PreviousEntryTerm)
	if s.VotedFor == "" && in.LastLogIndex == int32(s.PreviousEntryIndex) && in.LastLogTerm == int32(s.PreviousEntryTerm) {
		log.Printf("Gonna grant a vote for: %s", in.CandidateID)
		reply.VoteGranted = true
		s.VoteFor(in.CandidateID)
		return reply, nil
	}
	log.Printf("No condition satisfied, not granting a vote for: %s", in.CandidateID)
	return reply, nil
}

func (s *Server) AppendEntries(ctx context.Context, in *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, error) {
	log.Print("________AppendEntriesRPC___________")
	log.Print("Resetting election timer")
	if s.Election.ResetTicker != nil {
		s.Election.ResetTicker <- struct{}{}
	}

	term := int32(s.Server.CurrentTerm)
	successReply := &pb.AppendEntriesReply{Success: true, Term: term}
	failReply := &pb.AppendEntriesReply{Success: false, Term: term}

	if in.Term < int32(s.Server.CurrentTerm) {
		return failReply, nil
	}

	var lastEntryInSync bool
	var entry *entry.Entry
	if len(in.Entries) == 0 {
		log.Printf("Heartbeat received")
		entry, _ = s.Server.AppRepository.GetLastEntry(ctx)
		lastEntryInSync = isLastEntryInSync(int(in.PreviousLogIndex), int(in.PreviousLogTerm), *entry)

		if lastEntryInSync && int(in.LeaderCommitIndex) > s.Server.CommitIndex {
			commitEntries(s, int(in.LeaderCommitIndex))
		}
		return successReply, nil
	}

	log.Print("Leader PreviousLogIndex: ", in.PreviousLogIndex)
	entry, _ = s.Server.AppRepository.GetLastEntry(ctx)
	lastEntryInSync = isLastEntryInSync(int(in.PreviousLogIndex), int(in.PreviousLogTerm), *entry)

	if !lastEntryInSync {
		err := s.Server.AppRepository.DeleteAllEntriesStartingFrom(ctx, int(in.PreviousLogIndex))
		if err != nil {
			log.Panic("Found conflicting entry, but couldn't delete it. That's bad, I panic")
		}
		logEntries, _ := s.Server.AppRepository.GetLog(ctx)
		log.Print("Last entry not in sync. LastEntry: ", entry, "Follower after deleting conflicts: ", logEntries)
		return failReply, nil
	}

	entries := mapRaftEntriesToEntries(in.Entries)
	log.Print("Conditions satisfied. Persisting entries", entries)
	lastAppendedEntry, err := s.Server.AppRepository.PersistValues(ctx, entries)

	if err != nil {
		log.Printf("Unexpected error. Failed to persist entries")
		return failReply, nil
	}

	newCommitIndex := helpers.Min(int(in.LeaderCommitIndex), entries[0].Index-1)
	if newCommitIndex > s.Server.CommitIndex {
		commitEntries(s, newCommitIndex)
	}

	s.Server.PreviousEntryIndex = lastAppendedEntry.Index
	s.Server.PreviousEntryTerm = lastAppendedEntry.TermNumber
	log.Println("Follower: Succesfully appended entries to log")
	log.Println(helpers.PrettyPrint(in))
	return successReply, nil
}

func commitEntries(s *Server, newCommitIndex int) {
	log.Print("Gonna commit entries. Leader commit index: ", newCommitIndex, "Server commit index: ", s.CommitIndex)
	go func(leaderCommitIndex int) {
		s.Server.CommitEntries(leaderCommitIndex)
	}(newCommitIndex)
}

func isLastEntryInSync(leaderLastLogIndex int, leaderLastLogTerm int, lastEntry entry.Entry) bool {
	if leaderLastLogIndex != 0 {
		if lastEntry.IsEmpty() {
			return false
		}

		if lastEntry.Index != leaderLastLogIndex || lastEntry.TermNumber != leaderLastLogTerm {
			return false
		}
	}
	return true
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
