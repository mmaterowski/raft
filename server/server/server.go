package server

import (
	"context"
	"database/sql"
	"math/rand"
	"sync"
	"time"

	"github.com/mmaterowski/raft/model/entry"
	model "github.com/mmaterowski/raft/model/server"
	"github.com/mmaterowski/raft/persistence"
	rpcClient "github.com/mmaterowski/raft/rpc/client"
	protoBuff "github.com/mmaterowski/raft/rpc/raft_rpc"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/guard"
	"github.com/mmaterowski/raft/utils/helpers"
	log "github.com/sirupsen/logrus"
)

var Raft serverInterface = &server{}

var (
	State            state
	Type             model.ServerType
	Id               string
	mu               sync.Mutex
	Election         election
	heartbeatTicker  *time.Ticker
	triggerHeartbeat chan struct{}
	Others           []string
)

type server struct {
}

type state struct {
	VotedFor           string
	Entries            *map[string]entry.Entry
	CurrentTerm        int
	PreviousEntryIndex int
	PreviousEntryTerm  int
	CommitIndex        int
}

type serverInterface interface {
	StartServer(id string, isLocalEnv bool, isIntegrationTesting bool)
	IdentifyServer(local bool)
	VoteFor(candidateId string) bool
	RebuildStateFromLog() bool
	CommitEntries(leaderCommitIndex int) error
	MakeSureLastEntryDataIsAvailable()
	SetupElection()
	StartHeartbeat()
	ApplyEntryToState(entry *entry.Entry)
	SetCommitIndex(index int)
	BecomeFollower()
}

type election struct {
	Ticker      *time.Ticker
	ResetTicker chan struct{}
}

var (
	startServerInfo     = "Starting server..."
	electionTimeoutInfo = "Election timeout set to: %d + %d"
	logRebuiltInfo      = "Log successfully rebuilt.\nEntries: %s"
	commitIndexInfo     = "Leader commitIndex: %d. Server commitIndex: %d"
)

func (s *server) StartServer(id string, isLocalEnv bool, isIntegrationTesting bool) {
	Id = id

	s.IdentifyServer(isLocalEnv)
	Type = model.Candidate
	state := make(map[string]entry.Entry)
	State.Entries = &state
	State.PreviousEntryIndex = consts.NoPreviousEntryValue
	State.PreviousEntryTerm = consts.TermInitialValue
	State.CommitIndex = consts.LeaderCommitInitialValue
	log.Info(startServerInfo)
	s.VoteFor("")
	persistence.Repository.SetCurrentTerm(context.Background(), consts.TermUninitializedValue)
	State.CurrentTerm, _ = persistence.Repository.GetCurrentTerm(context.Background())
	Type = model.Follower
	triggerHeartbeat = make(chan struct{})
	heartbeatTicker = time.NewTicker(consts.HeartbeatInterval)
	seed := rand.NewSource(time.Now().UnixNano())
	electionTimeout := rand.New(seed).Intn(100)*3 + 3000
	log.Infof(electionTimeoutInfo, electionTimeout, consts.HeartbeatInterval)

	Election.Ticker = time.NewTicker(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)
	Election.ResetTicker = make(chan struct{})
}

func (s *server) IdentifyServer(local bool) {
	logContext := log.WithFields(log.Fields{"method": helpers.GetFunctionName(s.IdentifyServer), "isLocalEnv": local, "serverId": Id})

	if guard.AgainstEmptyString(Id) {
		logContext.Info("ServerId is empty.")
	}
	if local {
		Others = append(Others, consts.LaszloId, consts.RickyId)
		return
	}

	switch Id {
	case consts.KimId:
		Others = append(Others, consts.LaszloId, consts.RickyId)
	case consts.RickyId:
		Others = append(Others, consts.LaszloId, consts.KimId)
	case consts.LaszloId:
		Others = append(Others, consts.RickyId, consts.KimId)
	default:
		log.Panic("Couldn't identify Server")
	}
}

func (s *server) VoteFor(candidateId string) bool {
	err := persistence.Repository.SetVotedFor(context.Background(), candidateId)
	if err != nil {
		log.Print(err)
		return false
	}
	State.VotedFor = candidateId
	return true
}

func (s *server) RebuildStateFromLog() bool {
	if Type != model.Leader {
		return false
	}

	State.CommitIndex = consts.LeaderCommitInitialValue
	State.Entries = &map[string]entry.Entry{}

	entries, _ := persistence.Repository.GetLog(context.Background())
	for _, entry := range *entries {
		(*State.Entries)[entry.Key] = entry
		State.CommitIndex = entry.Index
	}
	// raft_signalr.AppHub.SendServerTypeChanged(Id, Type)
	log.Infof(logRebuiltInfo, helpers.PrettyPrint(entries))
	return true
}

func (s *server) CommitEntries(leaderCommitIndex int) error {
	mu.Lock()
	defer mu.Unlock()
	for leaderCommitIndex != State.CommitIndex {
		//TODO Optimize:Get all entries at once
		log.Infof(commitIndexInfo, leaderCommitIndex, State.CommitIndex)
		nextEntryIndexToCommit := State.CommitIndex + 1
		entry, err := persistence.Repository.GetEntryAtIndex(context.Background(), nextEntryIndexToCommit)
		if err != nil {
			log.Println("Error while commiting entry to state.")
			if err == sql.ErrNoRows {
				log.Print("Follower does not have yet all entries in log")
				break
			}
			log.Println(err)
			if Type == model.Leader {
				Type = model.Follower
				log.Printf("Server state set Leader->Follower")
			}
			break
		} else {
			log.Print("Commiting new entry to state. Key: ", entry.Key, " Entry: ", entry)

			// raft_signalr.AppHub.SendStateUpdated(Id, entry.Key, entry.Value)
			(*State.Entries)[entry.Key] = *entry
			State.CommitIndex++

		}

	}
	return nil
}

func (server *server) MakeSureLastEntryDataIsAvailable() {
	entry, _ := persistence.Repository.GetLastEntry(context.Background())
	if !entry.IsEmpty() {
		State.PreviousEntryIndex = entry.Index
		State.PreviousEntryTerm = entry.TermNumber
		return
	}
	State.PreviousEntryIndex = consts.NoPreviousEntryValue
	State.PreviousEntryTerm = consts.TermInitialValue
}

func (server *server) SetupElection() {
	go func() {
		for {
			select {
			case <-Election.Ticker.C:
				log.Info("Ticker timeout: Start election...")
				State.CurrentTerm++
				success := server.VoteFor(Id)
				if !success {
					log.Info("Something bad happened, couldn't vote for itself")
				}

				log.WithFields(log.Fields{
					"electionIssuer":   Id,
					"incrementsTermTo": State.CurrentTerm,
				}).Info("Election started")

				mu.Lock()
				Type = model.Candidate
				heartbeatTicker.Stop()
				mu.Unlock()

				for _, otherServer := range Others {
					go func(serverId string) {
						request := protoBuff.RequestVoteRequest{Term: int32(State.CurrentTerm), CandidateID: Id, LastLogIndex: int32(State.PreviousEntryIndex), LastLogTerm: int32(State.PreviousEntryTerm)}
						reply, err := rpcClient.Client.GetFor(serverId).RequestVote(context.Background(), &request)
						rpcLogContext := log.WithFields(log.Fields{"request": &request, "reply": reply, "error": err, "currentTerm": State.CurrentTerm, "serverType": Type})
						if err != nil {
							rpcLogContext.Error("Request vote error")
						}

						if reply == nil {
							rpcLogContext.Warn("No errors but nil reply, something werid happened")
							return
						}

						rpcLogContext.Info("Reply received")

						if reply.Term > int32(State.CurrentTerm) {
							log.WithField("NewCurrentTermValue", reply.Term).Info("Response term higher than candidate's. Changing current term")
							State.CurrentTerm = int(reply.Term)
							persistence.Repository.SetCurrentTerm(context.Background(), int(reply.Term))
						}

						if reply.VoteGranted && reply.Term <= int32(State.CurrentTerm) && Type != model.Leader {
							log.WithField("Leader", Id).Info("Becoming a leader")
							Type = model.Leader
							server.RebuildStateFromLog()
							triggerHeartbeat <- struct{}{}
							Election.ResetTicker <- struct{}{}
						}
					}(otherServer)
				}
			case <-Election.ResetTicker:
				seed := rand.NewSource(time.Now().UnixNano())
				electionTimeout := rand.New(seed).Intn(100) * 20
				log.WithField("elecitonTimeout", electionTimeout+int(consts.HeartbeatInterval)).Info("Resetting election ticker")
				Election.Ticker.Reset(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)
			}
		}
	}()
}

func (server *server) StartHeartbeat() {
	heartbeatTicker.Stop()
	go func() {
		for {
			select {
			case <-heartbeatTicker.C:
				if Election.ResetTicker != nil {
					Election.ResetTicker <- struct{}{}
				}
				for _, otherServer := range Others {
					go func(commitIndex int, term int, id string, otherServer string) {
						request := protoBuff.AppendEntriesRequest{LeaderCommitIndex: int32(commitIndex), Term: int32(term), LeaderId: id}
						log.WithField("Request", &request).Info("Sending hearbeat")
						rpcClient.Client.GetFor(otherServer).AppendEntries(context.Background(), &request)
					}(State.CommitIndex, State.CurrentTerm, Id, otherServer)
				}
			case <-triggerHeartbeat:
				heartbeatTicker.Reset(consts.HeartbeatInterval)
				log.WithField("interval", consts.HeartbeatInterval).Info("Heartbeat was resetted, expect next heartbeat after set interval")
				//just to recalculate select values
			}
		}
	}()
}

func (server *server) BecomeFollower() {
	heartbeatTicker.Stop()
	Type = model.Follower
}

func (server *server) ApplyEntryToState(entry *entry.Entry) {
	if entry.IsEmpty() {
		log.Warn("Tried to apply empty entry to state.")
		return
	}

	(*State.Entries)[entry.Key] = *entry
}

func (server *server) SetCommitIndex(index int) {
	if guard.AgainstNegativeValue(index) {
		log.Warn("Tried to set commit index to negative value")
	}

	State.CommitIndex = index
}
