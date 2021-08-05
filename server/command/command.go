package command

import (
	"context"
	"time"

	api "github.com/mmaterowski/raft/api"

	"github.com/mmaterowski/raft/consts"
	. "github.com/mmaterowski/raft/persistence"
	"github.com/mmaterowski/raft/raft_rpc"
	pb "github.com/mmaterowski/raft/raft_rpc"
	raftServer "github.com/mmaterowski/raft/raft_server"
	. "github.com/mmaterowski/raft/rpc_client"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AcceptLogEntry struct {
	key         string
	value       int
	currentTerm int
}

type AcceptLogEntryHandler struct {
	server             *raftServer.Server
	repo               AppRepository
	syncRequestService *api.SyncRequestService
	rpcClient          *Client
}

var retryIntervalValue = 1 * time.Second

func NewAcceptLogEntryHandler(repo AppRepository, server *raftServer.Server, onGoingSyncReq *api.SyncRequestService, client *Client) AcceptLogEntryHandler {
	if repo == nil {
		panic("nil repo")
	}
	if server == nil {
		panic("nil server")
	}

	return AcceptLogEntryHandler{server, repo, onGoingSyncReq, client}
}

func (h AcceptLogEntryHandler) Handle(ctx context.Context, cmd AcceptLogEntry) (err error) {
	// defer func() {
	// 	log.LogCommandExecution("ApproveTrainingReschedule", cmd, err)
	// }()
	cmdLogContext := log.WithField("command", cmd)
	h.server.MakeSureLastEntryDataIsAvailable()
	entry, persistErr := h.repo.PersistValue(ctx, cmd.key, cmd.value, cmd.currentTerm)
	if persistErr != nil {
		return persistErr
	}

	cmdLogContext.WithField("entry", entry).Info("Entry persisted")

	entries := []*raft_rpc.Entry{&pb.Entry{Index: int32(entry.Index), Value: int32(entry.Value), Key: entry.Key, TermNumber: int32(entry.TermNumber)}}

	h.syncRequestService.CancelOngoingRequests()

	c := make(chan struct{})

	for _, otherServer := range h.server.Others {
		go func(leaderId string, previousEntryIndex int, previousEntryTerm int, commitIndex int, otherServer string) {
			ctx, cancel := context.WithCancel(ctx)
			h.syncRequestService.AddHandler(otherServer, cancel)
			appendEntriesRequest := &pb.AppendEntriesRequest{Term: int32(cmd.currentTerm), LeaderId: h.server.Id, PreviousLogIndex: int32(previousEntryIndex), PreviousLogTerm: int32(previousEntryTerm), Entries: entries, LeaderCommitIndex: int32(commitIndex)}
			client := h.rpcClient.GetClientFor(otherServer)
			log.Print("Sending append entries request to: ", otherServer)
			reply, cancelled := retryUntilNoErrorReceived(client, h.syncRequestService, ctx, appendEntriesRequest, otherServer)
			if cancelled {
				log.Print("Append entries request cancelled")
				return
			}
			for !reply.Success {
				log.Printf("Follower %s did not accepted entry, syncing log", otherServer)
				if appendEntriesRequest.PreviousLogIndex == -1 {
					log.Panic("Follower should accept entry, because leader log is empty")
				}
				previousEntry, _ := h.repo.GetEntryAtIndex(ctx, int(appendEntriesRequest.PreviousLogIndex))
				if appendEntriesRequest.PreviousLogIndex > int32(consts.FirstEntryIndex) {
					appendEntriesRequest.PreviousLogIndex -= 1
				}
				appendEntriesRequest.PreviousLogTerm = int32(previousEntry.TermNumber)
				appendEntriesRequest.Entries = append([]*pb.Entry{{Index: int32(previousEntry.Index), Value: int32(previousEntry.Value), Key: previousEntry.Key, TermNumber: int32(previousEntry.TermNumber)}}, appendEntriesRequest.Entries...)
				reply, cancelled = retryUntilNoErrorReceived(client, h.syncRequestService, ctx, appendEntriesRequest, otherServer)
				if cancelled {
					log.Print("Append entries request cancelled")
					return
				}
			}

			if reply != nil {
				log.Printf("Follower %s responded to sync request: %s", otherServer, reply.String())
			}

			defer func() {
				c <- struct{}{}
			}()
		}(h.server.Id, h.server.PreviousEntryIndex, h.server.PreviousEntryTerm, h.server.CommitIndex, otherServer)
	}
	for i := 0; i < (len(h.server.Others) / 2); i++ {
		log.Print("Waiting for goroutine to finish work. i: ", i)
		<-c
	}

	h.server.ApplyEntryToState(entry)
	h.server.SetCommitIndex(entry.Index)
	return nil
}

func retryUntilNoErrorReceived(client pb.RaftRpcClient, cancelService *api.SyncRequestService, ctx context.Context, appendEntriesRequest *pb.AppendEntriesRequest, serverName string) (*pb.AppendEntriesReply, bool) {
	reply, rpcRequestError := client.AppendEntries(ctx, appendEntriesRequest, grpc.EmptyCallOption{})
	if rpcRequestError != nil {
		for rpcRequestError != nil {
			log.Print("Append entries request: ", appendEntriesRequest, "failed, because of rpc/network error:  ", rpcRequestError)
			select {
			case <-time.After(retryIntervalValue):
				log.Print("Retrying with delay...", retryIntervalValue)
				reply, rpcRequestError = client.AppendEntries(context.Background(), appendEntriesRequest, grpc.EmptyCallOption{})
			case <-ctx.Done():
				log.Print("New request arrived, cancelling sync request")
				cancelService.DeleteHandlerFor(serverName)
				return nil, true
			}

		}
	}
	cancelService.DeleteHandlerFor(serverName)
	log.Print("Append entries success: ", reply, " Error:", rpcRequestError)
	return reply, false
}
