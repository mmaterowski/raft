package command

import (
	"context"
	"time"

	"github.com/mmaterowski/raft/model/entry"
	"github.com/mmaterowski/raft/persistence"
	"github.com/mmaterowski/raft/server"

	"github.com/mmaterowski/raft/rpc/client"
	"github.com/mmaterowski/raft/rpc/raft_rpc"
	pb "github.com/mmaterowski/raft/rpc/raft_rpc"
	"github.com/mmaterowski/raft/services"
	"github.com/mmaterowski/raft/utils/consts"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AcceptLogEntry struct {
	Key   string
	Value int
}

type AcceptLogEntryHandler struct {
	syncLogChannel chan struct{}
}

var retryIntervalValue = 1 * time.Second

func NewAcceptLogEntryHandler() AcceptLogEntryHandler {

	syncLogChannel := make(chan struct{})
	return AcceptLogEntryHandler{syncLogChannel}
}

func (h AcceptLogEntryHandler) Handle(ctx context.Context, cmd AcceptLogEntry) (err error) {
	cmdLogContext := log.WithField("command", cmd)
	defer func() {
		log.Info("Executed command: AcceptLogEntry", cmd, err)
	}()

	server.Raft.MakeSureLastEntryDataIsAvailable()
	entry, persistErr := persistence.Repository.PersistValue(ctx, cmd.Key, cmd.Value, server.State.CurrentTerm)
	if persistErr != nil {
		return persistErr
	}

	cmdLogContext.WithField("entry", entry).Info("Entry persisted")

	entries := []*raft_rpc.Entry{&pb.Entry{Index: int32(entry.Index), Value: int32(entry.Value), Key: entry.Key, TermNumber: int32(entry.TermNumber)}}

	services.SyncRequestService.CancelOngoingRequests()

	for _, otherServer := range server.Others {
		go func(leaderId string, previousEntryIndex int, previousEntryTerm int, commitIndex int, otherServer string) {
			ctx, cancel := context.WithCancel(ctx)
			services.SyncRequestService.AddHandler(otherServer, cancel)
			appendEntriesRequest := &pb.AppendEntriesRequest{Term: int32(server.State.CurrentTerm), LeaderId: server.Id, PreviousLogIndex: int32(previousEntryIndex), PreviousLogTerm: int32(previousEntryTerm), Entries: entries, LeaderCommitIndex: int32(commitIndex)}
			appendEntriesLogContext := log.WithFields(log.Fields{"request": appendEntriesRequest, "sendTo": otherServer})
			client := client.Client.GetFor(otherServer)
			appendEntriesLogContext.Info("Sending append entries request")
			reply, cancelled := retryUntilNoErrorReceived(client, ctx, appendEntriesRequest, otherServer)
			if cancelled {
				appendEntriesLogContext.Info("Append entries request cancelled")
				return
			}
			for !reply.Success {
				appendEntriesLogContext.Info("Follower did not accept entry. Syncing log")
				validatePreviousLogIndex(appendEntriesLogContext, int(appendEntriesRequest.PreviousLogIndex))
				previousEntry, err := persistence.Repository.GetEntryAtIndex(ctx, int(appendEntriesRequest.PreviousLogIndex))
				if err != nil {
					appendEntriesLogContext.WithFields(log.Fields{"error": err, "index": appendEntriesRequest.PreviousLogIndex}).Error("Couldn't get entry at index")
					continue
				}
				prepareNextSyncRequest(previousEntry, appendEntriesRequest)
				reply, cancelled = retryUntilNoErrorReceived(client, ctx, appendEntriesRequest, otherServer)
				if cancelled {
					appendEntriesLogContext.Info("Append entries request cancelled")
					return
				}
			}

			if reply != nil {
				appendEntriesLogContext.WithField("reply", reply).Info("Follower responded to sync request")
			}

			defer func() {
				h.syncLogChannel <- struct{}{}
			}()
		}(server.Id, server.State.PreviousEntryIndex, server.State.PreviousEntryTerm, server.State.CommitIndex, otherServer)
	}
	for i := 0; i < (len(server.Others) / 2); i++ {
		cmdLogContext.Debug("Waiting for goroutine to finish work. i: ", i)
		<-h.syncLogChannel
	}

	server.Raft.ApplyEntryToState(entry)
	server.Raft.SetCommitIndex(entry.Index)
	return nil
}

func prepareNextSyncRequest(previousEntry *entry.Entry, appendEntriesRequest *pb.AppendEntriesRequest) {
	if appendEntriesRequest.PreviousLogIndex > int32(consts.FirstEntryIndex) {
		appendEntriesRequest.PreviousLogIndex -= 1
	}
	appendEntriesRequest.PreviousLogTerm = int32(previousEntry.TermNumber)
	appendEntriesRequest.Entries = append([]*pb.Entry{{Index: int32(previousEntry.Index), Value: int32(previousEntry.Value), Key: previousEntry.Key, TermNumber: int32(previousEntry.TermNumber)}}, appendEntriesRequest.Entries...)
}

func validatePreviousLogIndex(logCtx *log.Entry, previousIndex int) {
	if previousIndex == -1 {
		logCtx.Panic("Follower should accept entry, because leader log is empty. This should never happen")
	}
}

func retryUntilNoErrorReceived(client pb.RaftRpcClient, ctx context.Context, appendEntriesRequest *pb.AppendEntriesRequest, serverName string) (*pb.AppendEntriesReply, bool) {
	reply, rpcRequestError := client.AppendEntries(ctx, appendEntriesRequest, grpc.EmptyCallOption{})
	if rpcRequestError != nil {
		for rpcRequestError != nil {
			log.Print("Append entries request: ", appendEntriesRequest, "failed, because of rpc/network error:  ", rpcRequestError)
			select {
			case <-time.After(retryIntervalValue):
				log.Print("Retrying with delay...", retryIntervalValue)
				reply, rpcRequestError = client.AppendEntries(context.Background(), appendEntriesRequest, grpc.EmptyCallOption{})
				log.WithFields(log.Fields{"reply": reply, "error": rpcRequestError}).Info("Retried request")
			case <-ctx.Done():
				log.Print("New request arrived, cancelling sync request")
				services.SyncRequestService.DeleteHandlerFor(serverName)
				return nil, true
			}

		}
	}
	services.SyncRequestService.DeleteHandlerFor(serverName)
	log.Print("Append entries success: ", reply, " Error:", rpcRequestError)
	return reply, false
}
