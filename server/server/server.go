package server

import (
	"context"
	"database/sql"
	"math/rand"
	"sync"
	"time"

	"github.com/mmaterowski/raft/model/entry"
	server_model "github.com/mmaterowski/raft/model/server"
	. "github.com/mmaterowski/raft/persistence"
	rpcClient "github.com/mmaterowski/raft/rpc/client"
	protoBuff "github.com/mmaterowski/raft/rpc/raft_rpc"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/guard"
	"github.com/mmaterowski/raft/utils/helpers"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	server_model.ServerType
	State              *map[string]entry.Entry
	CurrentTerm        int
	PreviousEntryIndex int
	PreviousEntryTerm  int
	CommitIndex        int
	Id                 string
	VotedFor           string
	AppRepository
	mu               sync.Mutex
	Election         Election
	HeartbeatTicker  *time.Ticker
	TriggerHeartbeat chan struct{}
	RpcClient        rpcClient.Client
	Others           []string
}

type Election struct {
	Ticker      *time.Ticker
	ResetTicker chan struct{}
}

var (
	startServerInfo     = "Starting server..."
	electionTimeoutInfo = "Election timeout set to: %d + %d"
	logRebuiltInfo      = "Log successfully rebuilt.\nEntries: %s"
	commitIndexInfo     = "Leader commitIndex: %d. Server commitIndex: %d"
)

func (s *Server) StartServer(id string, isLocalEnv bool, isIntegrationTesting bool) {
	s.Id = id
	s.IdentifyServer(isLocalEnv)
	s.ServerType = server_model.ServerType(server_model.Candidate)
	state := make(map[string]entry.Entry)
	s.State = &state
	s.PreviousEntryIndex = consts.NoPreviousEntryValue
	s.PreviousEntryTerm = consts.TermInitialValue
	s.CommitIndex = consts.LeaderCommitInitialValue
	log.Info(startServerInfo)
	s.VoteFor("")
	s.AppRepository.SetCurrentTerm(context.Background(), consts.TermUninitializedValue)
	s.CurrentTerm, _ = s.AppRepository.GetCurrentTerm(context.Background())
	s.ServerType = server_model.Follower
	s.TriggerHeartbeat = make(chan struct{})
	s.HeartbeatTicker = time.NewTicker(consts.HeartbeatInterval)
	seed := rand.NewSource(time.Now().UnixNano())
	electionTimeout := rand.New(seed).Intn(100)*300 + 100
	log.Infof(electionTimeoutInfo, electionTimeout, consts.HeartbeatInterval)

	if isIntegrationTesting {
		return
	}
	s.Election.Ticker = time.NewTicker(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)
	s.Election.ResetTicker = make(chan struct{})

}

func (s *Server) IdentifyServer(local bool) {
	logContext := log.WithFields(log.Fields{"method": helpers.GetFunctionName(s.IdentifyServer), "isLocalEnv": local, "serverId": s.Id})

	if guard.AgainstEmptyString(s.Id) {
		logContext.Info("ServerId is empty.")
	}
	if local {
		s.Others = append(s.Others, consts.LaszloId, consts.RickyId)
		return
	}

	switch s.Id {
	case consts.KimId:
		s.Others = append(s.Others, consts.LaszloId, consts.RickyId)
	case consts.RickyId:
		s.Others = append(s.Others, consts.LaszloId, consts.KimId)
	case consts.LaszloId:
		s.Others = append(s.Others, consts.RickyId, consts.KimId)
	default:
		log.Panic("Couldn't identify Server")
	}
}

func (s *Server) VoteFor(candidateId string) bool {
	err := s.SetVotedFor(context.Background(), candidateId)
	if err != nil {
		log.Print(err)
		return false
	}
	s.VotedFor = candidateId
	return true
}

func (s *Server) RebuildStateFromLog() bool {
	entries, _ := s.AppRepository.GetLog(context.Background())
	for _, entry := range *entries {
		(*s.State)[entry.Key] = entry
		s.CommitIndex = entry.Index
	}
	log.Infof(logRebuiltInfo, helpers.PrettyPrint(entries))
	return true
}

func (s *Server) CommitEntries(leaderCommitIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for leaderCommitIndex != s.CommitIndex {
		//TODO Optimize:Get all entries at once
		log.Infof(commitIndexInfo, leaderCommitIndex, s.CommitIndex)
		nextEntryIndexToCommit := s.CommitIndex + 1
		entry, err := s.AppRepository.GetEntryAtIndex(context.Background(), nextEntryIndexToCommit)
		if err != nil {
			log.Println("Error while commiting entry to state.")
			if err == sql.ErrNoRows {
				log.Print("Follower does not have yet all entries in log")
				break
			}
			log.Println(err)
			if s.ServerType == server_model.Leader {
				s.ServerType = server_model.Follower
				log.Printf("Server state set Leader->Follower")
			}
			break
		} else {
			log.Print("Commiting new entry to state. Key: ", entry.Key, " Entry: ", entry)
			(*s.State)[entry.Key] = *entry
			s.CommitIndex++

		}

	}
	return nil
}

func (server *Server) MakeSureLastEntryDataIsAvailable() {
	entry, _ := server.AppRepository.GetLastEntry(context.Background())
	if !entry.IsEmpty() {
		server.PreviousEntryIndex = entry.Index
		server.PreviousEntryTerm = entry.TermNumber
		return
	}
	server.PreviousEntryIndex = consts.NoPreviousEntryValue
	server.PreviousEntryTerm = consts.TermInitialValue
}

func (server *Server) SetupElection() {
	go func() {
		for {
			select {
			case <-server.Election.Ticker.C:
				log.Info("Ticker timeout: Start election...")
				server.CurrentTerm++
				success := server.VoteFor(server.Id)
				if !success {
					log.Info("Something bad happened, couldn't vote for itself")
				}

				log.WithFields(log.Fields{
					"electionIssuer":   server.Id,
					"incrementsTermTo": server.CurrentTerm,
				}).Info("Election started")

				server.mu.Lock()
				server.ServerType = server_model.Candidate
				server.HeartbeatTicker.Stop()
				server.mu.Unlock()

				for _, otherServer := range server.Others {
					go func(serverId string) {
						request := protoBuff.RequestVoteRequest{Term: int32(server.CurrentTerm), CandidateID: server.Id, LastLogIndex: int32(server.PreviousEntryIndex), LastLogTerm: int32(server.PreviousEntryTerm)}
						reply, err := server.RpcClient.GetClientFor(serverId).RequestVote(context.Background(), &request)
						rpcLogContext := log.WithFields(log.Fields{"request": &request, "reply": reply, "error": err, "currentTerm": server.CurrentTerm, "serverType": server.ServerType})
						if err != nil {
							rpcLogContext.Error("Request vote error")
						}
						if reply == nil {
							rpcLogContext.Warn("No errors but nil reply, something werid happened")
							return
						}

						if reply.Term > int32(server.CurrentTerm) {
							log.WithField("NewCurrentTermValue", reply.Term).Info("Response term higher than candidate's. Changing current term")
							server.CurrentTerm = int(reply.Term)
							server.SetCurrentTerm(context.Background(), int(reply.Term))
						}

						rpcLogContext.Info("Reply received")
						if reply.VoteGranted && reply.Term <= int32(server.CurrentTerm) && server.ServerType != server_model.Leader {
							log.WithField("Leader", server.Id).Info("Becoming a leader")
							server.ServerType = server_model.Leader
							server.HeartbeatTicker = time.NewTicker(consts.HeartbeatInterval)
							server.TriggerHeartbeat <- struct{}{}
							server.Election.ResetTicker <- struct{}{}
						}
					}(otherServer)
				}
			case <-server.Election.ResetTicker:
				seed := rand.NewSource(time.Now().UnixNano())
				electionTimeout := rand.New(seed).Intn(100)*300 + 100
				log.WithField("elecitonTimeout", electionTimeout+int(consts.HeartbeatInterval)).Info("Resetting election ticker")
				server.Election.Ticker.Reset(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)
			}
		}
	}()
}

func (server *Server) StartHeartbeat() {
	if server.Id != consts.KimId {
		log.Info("Election not implemented fully, heartbeating only for Kim")
		return
	}
	server.HeartbeatTicker.Stop()
	//not heartbeating with integration tests for kim, channel conflicts!
	go func() {
		for {
			select {
			case <-server.HeartbeatTicker.C:
				if server.Election.ResetTicker != nil {
					server.Election.ResetTicker <- struct{}{}
				}
				for _, otherServer := range server.Others {
					go func(commitIndex int, term int, id string, otherServer string) {
						request := protoBuff.AppendEntriesRequest{LeaderCommitIndex: int32(commitIndex), Term: int32(term), LeaderId: id}
						log.WithField("Request", &request).Info("Sending hearbeat")
						server.RpcClient.GetClientFor(otherServer).AppendEntries(context.Background(), &request)
					}(server.CommitIndex, server.CurrentTerm, server.Id, otherServer)
				}
			case <-server.TriggerHeartbeat:
				log.WithField("interval", consts.HeartbeatInterval).Info("Heartbeat was resetted, expect next heargbeat after set interval")
				//just to recalculate select values
			}
		}
	}()
}

func (server *Server) ApplyEntryToState(entry *entry.Entry) {
	if entry.IsEmpty() {
		log.Warn("Tried to apply empty entry to state.")
		return
	}

	(*server.State)[entry.Key] = *entry
}

func (server *Server) SetCommitIndex(index int) {
	if guard.AgainstNegativeValue(index) {
		log.Warn("Tried to set commit index to negative value")
	}

	server.CommitIndex = index

}
