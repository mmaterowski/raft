package server

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/mmaterowski/raft/model/entry"
	"github.com/mmaterowski/raft/model/server"
	"github.com/mmaterowski/raft/persistence"
	"github.com/mmaterowski/raft/rpc/raft_rpc"
	raftServer "github.com/mmaterowski/raft/server"
	"github.com/mmaterowski/raft/utils/helpers"
	raftWs "github.com/mmaterowski/raft/ws"
)

type Server struct {
	raft_rpc.UnimplementedRaftRpcServer
}

func (s *Server) RequestVote(ctx context.Context, in *raft_rpc.RequestVoteRequest) (*raft_rpc.RequestVoteReply, error) {
	log.Print("________RequestVoteRPC________")
	log.Print("Request: ", in)
	reply := &raft_rpc.RequestVoteReply{Term: int32(raftServer.State.CurrentTerm), VoteGranted: false}

	if in.Term < int32(raftServer.State.CurrentTerm) {
		log.Print("Issuing candidate term is lower than server")
		return reply, nil
	}

	if in.Term > int32(raftServer.State.CurrentTerm) {
		log.Print("Request term higher than server's. Resetting voted for.")
		raftServer.State.CurrentTerm = int(in.Term)
		persistence.Repository.SetCurrentTerm(ctx, int(in.Term))
		persistence.Repository.SetVotedFor(ctx, "")
	}

	log.Print("Voted for: ", raftServer.State.VotedFor, " Server previous entry index: ", raftServer.State.PreviousEntryIndex, " previous entry term: ", raftServer.State.PreviousEntryTerm)
	if raftServer.State.VotedFor == "" || (in.LastLogIndex == int32(raftServer.State.PreviousEntryIndex) && in.LastLogTerm == int32(raftServer.State.PreviousEntryTerm)) {
		log.Printf("Gonna grant a vote for: %s", in.CandidateID)
		reply.VoteGranted = true
		raftServer.Raft.VoteFor(in.CandidateID)
		return reply, nil
	}
	log.Printf("No condition satisfied, not granting a vote for: %s", in.CandidateID)
	return reply, nil
}

func (s *Server) AppendEntries(ctx context.Context, in *raft_rpc.AppendEntriesRequest) (*raft_rpc.AppendEntriesReply, error) {
	log.Print("________AppendEntriesRPC___________")
	if raftServer.Election.ResetTicker != nil {
		log.Print("Resetting election timer")
		raftServer.Election.ResetTicker <- struct{}{}
	}

	term := int32(raftServer.State.CurrentTerm)
	successReply := &raft_rpc.AppendEntriesReply{Success: true, Term: term}
	failReply := &raft_rpc.AppendEntriesReply{Success: false, Term: term}

	if in.Term >= int32(raftServer.State.CurrentTerm) && raftServer.Type == server.Leader {
		raftServer.Raft.BecomeFollower()
	}
	if in.Term < int32(raftServer.State.CurrentTerm) {
		return failReply, nil
	}

	var lastEntryInSync bool
	var entry *entry.Entry
	if len(in.Entries) == 0 {
		log.Info("Heartbeat received")
		raftWs.AppHub.SendHeartbeatReceivedMessage(raftServer.Id)
		entry, _ = persistence.Repository.GetLastEntry(ctx)
		lastEntryInSync = isLastEntryInSync(int(in.PreviousLogIndex), int(in.PreviousLogTerm), *entry)

		if lastEntryInSync && int(in.LeaderCommitIndex) > raftServer.State.CommitIndex {
			commitEntries(s, int(in.LeaderCommitIndex))
		}
		return successReply, nil
	}

	log.Print("Leader PreviousLogIndex: ", in.PreviousLogIndex)
	entry, _ = persistence.Repository.GetLastEntry(ctx)
	lastEntryInSync = isLastEntryInSync(int(in.PreviousLogIndex), int(in.PreviousLogTerm), *entry)

	if !lastEntryInSync {
		err := persistence.Repository.DeleteAllEntriesStartingFrom(ctx, int(in.PreviousLogIndex))
		if err != nil {
			log.Panic("Found conflicting entry, but couldn't delete it. That's bad, I panic")
		}
		logEntries, _ := persistence.Repository.GetLog(ctx)
		log.Print("Last entry not in sync. LastEntry: ", entry, "Follower after deleting conflicts: ", logEntries)
		return failReply, nil
	}

	entries := mapRaftEntriesToEntries(in.Entries)
	log.Print("Conditions satisfied. Persisting entries", entries)
	lastAppendedEntry, err := persistence.Repository.PersistValues(ctx, entries)

	if err != nil {
		log.Printf("Unexpected error. Failed to persist entries")
		return failReply, nil
	}

	newCommitIndex := helpers.Min(int(in.LeaderCommitIndex), entries[0].Index-1)
	if newCommitIndex > raftServer.State.CommitIndex {
		commitEntries(s, newCommitIndex)
	}

	raftServer.State.PreviousEntryIndex = lastAppendedEntry.Index
	raftServer.State.PreviousEntryTerm = lastAppendedEntry.TermNumber
	log.Println("Follower: Succesfully appended entries to log")
	log.Println(helpers.PrettyPrint(in))
	return successReply, nil
}

func commitEntries(s *Server, newCommitIndex int) {
	log.Print("Gonna commit entries. Leader commit index: ", newCommitIndex, "Server commit index: ", raftServer.State.CommitIndex)
	go func(leaderCommitIndex int) {
		raftServer.Raft.CommitEntries(leaderCommitIndex)
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

func mapRaftEntriesToEntries(rpcEntries []*raft_rpc.Entry) []entry.Entry {
	var entries []entry.Entry
	for _, raftEntry := range rpcEntries {
		entries = append(entries, (getEntryFromRaftEntry(raftEntry)))
	}
	return entries
}

func getEntryFromRaftEntry(rpcEntry *raft_rpc.Entry) entry.Entry {
	entry := entry.Entry{Key: rpcEntry.Key, Index: int(rpcEntry.Index), Value: int(rpcEntry.Value), TermNumber: int(rpcEntry.TermNumber)}
	return entry
}
