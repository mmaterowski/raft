package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

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

type StatusResponse struct {
	Status structs.ServerType
}

type ValueResponse struct {
	Value int
}

type PutResponse struct {
	Success bool
}

func IdentifyServer(serverId string, debug bool) {

	if debug {
		others = append(others, laszloId, rickyId)
	}

	switch serverId {
	case "Kim":
		others = append(others, laszloId, rickyId)
	case "Ricky":
		others = append(others, laszloId, kimId)
	case "Laszlo":
		others = append(others, rickyId, kimId)
	default:
		log.Panic("Couldn't identify RaftServerReference")
	}
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: RaftServerReference.ServerType}
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

	entry := RaftServerReference.State[key]
	data := ValueResponse{Value: entry.Value}
	respond.With(w, r, http.StatusOK, data)
}

func AcceptLogEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValue(r)
	Check(err)
	entry, _ := RaftServerReference.Context.PersistValue(r.Context(), key, value, RaftServerReference.CurrentTerm)
	if !entry.IsEmpty() {
		respond.With(w, r, http.StatusOK, PutResponse{Success: false})
		return
	}

	entries := []*raft_rpc.Entry{}
	entries = append(entries, &pb.Entry{Index: int32(entry.Index), Value: int32(entry.Value), Key: entry.Key, TermNumber: int32(entry.TermNumber)})
	makeSureLastEntryDataIsAvailable(r.Context())
	orderFollowersToSyncTheirLog(entries)

	RaftServerReference.State[entry.Key] = *entry
	RaftServerReference.CommitIndex = entry.Index

	orderFollowersToCommitTheirEntries()

	data := PutResponse{Success: true}
	respond.With(w, r, http.StatusOK, data)
}

func orderFollowersToSyncTheirLog(entries []*raft_rpc.Entry) {
	var wg sync.WaitGroup

	wg.Add((len(others) / 2) + 1)
	for _, otherServer := range others {
		go func(leaderId string, previousEntryIndex int, previousEntryTerm int, commitIndex int, otherServer string) {
			appendEntriesRequest := pb.AppendEntriesRequest{Term: int32(RaftServerReference.CurrentTerm), LeaderId: RaftServerReference.Id, PreviousLogIndex: int32(previousEntryIndex), PreviousLogTerm: int32(previousEntryTerm), Entries: entries, LeaderCommitIndex: int32(commitIndex)}
			reply, err := RpcClientReference.GetClientFor(otherServer).AppendEntries(context.Background(), &appendEntriesRequest, grpc.EmptyCallOption{})
			//If the reply unsuccessfull force followers to sync their log
			//Find matching entry on follower and order it to replace a slice of log so it matches with leader

			//Init next index value with your last matching log
			//Decrement next index and send previous entry
			//Question: Maybe that's how you make call with multiple entries? Just send whole slice that increasingly grows
			// and when it's accepted it's synced
			//Repeat
			//Finally it will sync

			Check(err)
			log.Print(reply.String())
			defer wg.Done()
		}(RaftServerReference.Id, RaftServerReference.PreviousEntryIndex, RaftServerReference.PreviousEntryTerm, RaftServerReference.CommitIndex, otherServer)
	}
	wg.Wait()

}

func orderFollowersToCommitTheirEntries() {
	for _, otherServer := range others {
		go func(commitIndex int, otherServer string) {
			request := pb.CommitAvailableEntriesRequest{LeaderCommitIndex: int32(commitIndex)}
			RpcClientReference.GetClientFor(otherServer).CommitAvailableEntries(context.Background(), &request)
		}(RaftServerReference.CommitIndex, otherServer)
	}

}

func makeSureLastEntryDataIsAvailable(ctx context.Context) {
	if RaftServerReference.PreviousEntryIndex < 0 || RaftServerReference.PreviousEntryTerm < 0 {
		entry, _ := RaftServerReference.Context.GetLastEntry(ctx)
		if entry.IsEmpty() {
			RaftServerReference.PreviousEntryIndex = entry.Index
			RaftServerReference.PreviousEntryTerm = entry.TermNumber
		}

	}
	RaftServerReference.PreviousEntryIndex = 0
	RaftServerReference.PreviousEntryTerm = 0
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

func HandleRequests() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", home)
	r.HandleFunc("/status", GetStatus)
	r.HandleFunc("/get/{key}", GetKeyValue)
	r.HandleFunc("/put/{key}/{value}", AcceptLogEntry)
	port := os.Getenv("SERVER_PORT")
	log.Printf("API listens on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
