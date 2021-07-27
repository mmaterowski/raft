package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mmaterowski/raft/consts"
	"github.com/mmaterowski/raft/entry"
	"github.com/mmaterowski/raft/helpers"
	. "github.com/mmaterowski/raft/helpers"
	"github.com/mmaterowski/raft/raft_rpc"
	pb "github.com/mmaterowski/raft/raft_rpc"
	raftServer "github.com/mmaterowski/raft/raft_server"
	. "github.com/mmaterowski/raft/rpc"
	structs "github.com/mmaterowski/raft/structs"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"gopkg.in/matryer/respond.v1"
)

var laszloId = "Laszlo"
var rickyId = "Ricky"
var kimId = "Kim"
var others []string
var RaftServerReference *raftServer.Server
var RpcClientReference *Client
var retryIntervalValue = 1 * time.Second
var cancelHandlesForOngoingSyncRequest []context.CancelFunc

type StatusResponse struct {
	Status structs.ServerType
}

type ValueResponse struct {
	Value int
}

type PutResponse struct {
	Success bool
}

func IdentifyServer(serverId string, local bool) []string {

	if local {
		others = append(others, laszloId, rickyId)
		return others
	}

	switch serverId {
	case consts.KimId:
		others = append(others, laszloId, rickyId)
	case consts.RickyId:
		others = append(others, laszloId, kimId)
	case consts.LaszloId:
		others = append(others, rickyId, kimId)
	default:
		log.Panic("Couldn't identify RaftServerReference")
	}
	return others
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: RaftServerReference.ServerType}
	respond.With(w, r, http.StatusOK, data)
}

func PersistAndCommitValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValue(r)
	Check(err)
	entry, _ := RaftServerReference.AppRepository.PersistValue(r.Context(), key, value, RaftServerReference.CurrentTerm)
	(*RaftServerReference.State)[entry.Key] = *entry
	RaftServerReference.CommitIndex = entry.Index
	data := PutResponse{Success: !entry.IsEmpty()}
	log.Print("Backdooring entry. Persisted and commited ", entry)
	respond.With(w, r, http.StatusOK, data)
}

func DeleteAllEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = RaftServerReference.AppRepository.DeleteAllEntriesStartingFrom(r.Context(), 1)
	(*RaftServerReference.State) = map[string]entry.Entry{}
	RaftServerReference.CommitIndex = consts.LeaderCommitInitialValue
	RaftServerReference.PreviousEntryIndex = consts.NoPreviousEntryValue
	RaftServerReference.PreviousEntryTerm = consts.TermInitialValue
	data := PutResponse{Success: true}
	log.Print("Backdooring. Deleted all entries and cleared state")
	respond.With(w, r, http.StatusOK, data)
}

func GetKeyValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	variables := mux.Vars(r)
	key := variables["key"]

	if key == "" {
		message := "Argument 'key' missing"
		log.Print(message)
		respond.With(w, r, http.StatusInternalServerError, message)
	}

	entry := (*RaftServerReference.State)[key]
	data := ValueResponse{Value: entry.Value}
	log.Println("Getting key value")
	log.Println(helpers.PrettyPrint(entry))
	respond.With(w, r, http.StatusOK, data)
}

func AcceptLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValue(r)
	Check(err)
	makeSureLastEntryDataIsAvailable(ctx)
	entry, persistErr := RaftServerReference.AppRepository.PersistValue(ctx, key, value, RaftServerReference.CurrentTerm)
	if persistErr != nil {
		log.Print("Error while persisting entry", persistErr)
	}
	log.Println("Leader persisted value: ", entry)
	if entry.IsEmpty() {
		respond.With(w, r, http.StatusOK, PutResponse{Success: false})
		return
	}

	entries := []*raft_rpc.Entry{}
	entries = append(entries, &pb.Entry{Index: int32(entry.Index), Value: int32(entry.Value), Key: entry.Key, TermNumber: int32(entry.TermNumber)})
	cancelOngoingSyncRequest()
	orderFollowersToSyncTheirLog(ctx, entries)

	(*RaftServerReference.State)[entry.Key] = *entry
	RaftServerReference.CommitIndex = entry.Index
	data := PutResponse{Success: true}
	respond.With(w, r, http.StatusOK, data)
}

func cancelOngoingSyncRequest() {
	if len(cancelHandlesForOngoingSyncRequest) > 0 {
		log.Printf("Cancelling sync requests")
		for _, cancelFunction := range cancelHandlesForOngoingSyncRequest {
			cancelFunction()
		}
		cancelHandlesForOngoingSyncRequest = []context.CancelFunc{}
	}
}

func orderFollowersToSyncTheirLog(ctx context.Context, entries []*raft_rpc.Entry) {
	var wg sync.WaitGroup

	wg.Add((len(others) / 2))

	for _, otherServer := range others {
		go func(leaderId string, previousEntryIndex int, previousEntryTerm int, commitIndex int, otherServer string) {
			ctx, cancel := context.WithCancel(ctx)
			cancelHandlesForOngoingSyncRequest = append(cancelHandlesForOngoingSyncRequest, cancel)
			appendEntriesRequest := buildAppenEntriesRequest(previousEntryIndex, previousEntryTerm, entries, commitIndex)
			client := RpcClientReference.GetClientFor(otherServer)
			log.Print("Sending append entries request to: ", otherServer)
			reply, cancelled := retryUntilNoErrorReceived(client, ctx, appendEntriesRequest)
			if cancelled {
				return
			}
			for !reply.Success {
				log.Printf("Follower %s did not accepted entry, syncing log", otherServer)
				if appendEntriesRequest.PreviousLogIndex == -1 {
					log.Panic("Follower should accept entry, because leader log is empty")
				}
				previousEntry, _ := RaftServerReference.AppRepository.GetEntryAtIndex(ctx, int(appendEntriesRequest.PreviousLogIndex))
				if appendEntriesRequest.PreviousLogIndex > int32(consts.FirstEntryIndex) {
					appendEntriesRequest.PreviousLogIndex -= 1
				}
				appendEntriesRequest.PreviousLogTerm = int32(previousEntry.TermNumber)
				appendEntriesRequest.Entries = append([]*pb.Entry{{Index: int32(previousEntry.Index), Value: int32(previousEntry.Value), Key: previousEntry.Key, TermNumber: int32(previousEntry.TermNumber)}}, appendEntriesRequest.Entries...)
				reply, cancelled = retryUntilNoErrorReceived(client, ctx, appendEntriesRequest)
				if cancelled {
					return
				}
			}

			if reply != nil {
				log.Print("Follower responded to sync request", reply.String())
			}
			defer wg.Done()
		}(RaftServerReference.Id, RaftServerReference.PreviousEntryIndex, RaftServerReference.PreviousEntryTerm, RaftServerReference.CommitIndex, otherServer)
	}
	wg.Wait()

}

func buildAppenEntriesRequest(previousEntryIndex int, previousEntryTerm int, entries []*pb.Entry, commitIndex int) *pb.AppendEntriesRequest {
	appendEntriesRequest := pb.AppendEntriesRequest{Term: int32(RaftServerReference.CurrentTerm), LeaderId: RaftServerReference.Id, PreviousLogIndex: int32(previousEntryIndex), PreviousLogTerm: int32(previousEntryTerm), Entries: entries, LeaderCommitIndex: int32(commitIndex)}
	return &appendEntriesRequest
}

func retryUntilNoErrorReceived(client pb.RaftRpcClient, ctx context.Context, appendEntriesRequest *pb.AppendEntriesRequest) (*pb.AppendEntriesReply, bool) {
	reply, rpcRequestError := client.AppendEntries(ctx, appendEntriesRequest, grpc.EmptyCallOption{})
	if rpcRequestError != nil {
		for rpcRequestError != nil {
			log.Print("Append entries request: ", appendEntriesRequest, "failed, because of rpc/network error:  ", rpcRequestError)
			select {
			case <-time.After(retryIntervalValue):
				log.Print("Retrying with delay...", retryIntervalValue)
				reply, rpcRequestError = client.AppendEntries(ctx, appendEntriesRequest, grpc.EmptyCallOption{})
			case <-ctx.Done():
				log.Print("New request arrived, cancelling sync request")
				return nil, true
			}

		}
	}
	log.Printf("Returtning from retryUntilNoErrorReceived")
	log.Print("Reply: ", reply, " Error:", rpcRequestError)
	return reply, false
}

func makeSureLastEntryDataIsAvailable(ctx context.Context) {
	entry, _ := RaftServerReference.AppRepository.GetLastEntry(ctx)
	if !entry.IsEmpty() {
		RaftServerReference.PreviousEntryIndex = entry.Index
		RaftServerReference.PreviousEntryTerm = entry.TermNumber
		return
	}
	RaftServerReference.PreviousEntryIndex = consts.NoPreviousEntryValue
	RaftServerReference.PreviousEntryTerm = consts.TermInitialValue
}

func getKeyAndValue(r *http.Request) (string, int, error) {
	var key string
	var value int
	var err error

	variables := mux.Vars(r)
	key = variables["key"]

	if key == "" {
		return key, value, errors.New("argument 'key' missing")
	}
	value, convError := strconv.Atoi(variables["value"])
	if convError != nil {
		return key, value, err
	}
	return key, value, err
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Raft module! RaftServerReference: "+RaftServerReference.Id)

}

func HandleRequests(port string) {
	if RaftServerReference == nil {
		log.Fatal("No raft server reference set")
	}

	if RpcClientReference == nil {
		log.Fatal("No rpc client reference set")
	}

	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", home)
	r.HandleFunc("/status", GetStatus)
	r.HandleFunc("/get/{key}", GetKeyValue)
	r.HandleFunc("/put/{key}/{value}", AcceptLogEntry)
	r.HandleFunc("/backdoor/put/{key}/{value}", PersistAndCommitValue)
	r.HandleFunc("/backdoor/deleteall", DeleteAllEntries)
	log.Printf("API listens on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
