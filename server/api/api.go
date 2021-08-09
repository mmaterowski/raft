package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	syncRequest "github.com/mmaterowski/raft/cancel_service"
	command "github.com/mmaterowski/raft/command"
	"github.com/mmaterowski/raft/model/entry"
	"github.com/mmaterowski/raft/model/server"
	"github.com/mmaterowski/raft/rpc/client"
	raftServer "github.com/mmaterowski/raft/server"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/helpers"
	log "github.com/sirupsen/logrus"
	"gopkg.in/matryer/respond.v1"
)

var raftServerReference *raftServer.Server
var rpcClientReference *client.Client
var port string
var cancelService *syncRequest.SyncRequestService

type StatusResponse struct {
	Status server.ServerType
}

type ValueResponse struct {
	Value int
}

type PutResponse struct {
	Success bool
}

func InitApi(RaftServerReference *raftServer.Server, RpcClientReference *client.Client, CancelService *syncRequest.SyncRequestService, Port string) {
	if RaftServerReference == nil {
		log.Fatal("No raft server reference set")
	}

	if RpcClientReference == nil {
		log.Fatal("No rpc client reference set")
	}
	if Port == "" {
		log.Fatal("No api port specified")
	}

	if CancelService == nil {
		log.Fatal("No cancel service set")
	}

	raftServerReference = RaftServerReference
	rpcClientReference = RpcClientReference
	cancelService = CancelService
	port = Port
}

func HandleRequests() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", home)
	r.HandleFunc("/status", getStatus)
	r.HandleFunc("/get/{key}", getKeyValue)
	r.HandleFunc("/put/{key}/{value}", acceptLogEntry)
	r.HandleFunc("/backdoor/put/{key}/{value}", persistAndCommitValue)
	r.HandleFunc("/backdoor/deleteall", deleteAllEntries)
	log.Printf("API listens on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: raftServerReference.ServerType}
	respond.With(w, r, http.StatusOK, data)
}

func persistAndCommitValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValueFromRequestArgs(r)
	helpers.Check(err)
	entry, _ := raftServerReference.AppRepository.PersistValue(r.Context(), key, value, raftServerReference.CurrentTerm)
	(*raftServerReference.State)[entry.Key] = *entry
	raftServerReference.CommitIndex = entry.Index
	data := PutResponse{Success: !entry.IsEmpty()}
	log.Print("Backdooring entry. Persisted and commited ", entry)
	respond.With(w, r, http.StatusOK, data)
}

func deleteAllEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = raftServerReference.AppRepository.DeleteAllEntriesStartingFrom(r.Context(), 1)
	(*raftServerReference.State) = map[string]entry.Entry{}
	raftServerReference.CommitIndex = consts.LeaderCommitInitialValue
	raftServerReference.PreviousEntryIndex = consts.NoPreviousEntryValue
	raftServerReference.PreviousEntryTerm = consts.TermInitialValue
	data := PutResponse{Success: true}
	log.Print("Backdooring. Deleted all entries and cleared state")
	respond.With(w, r, http.StatusOK, data)
}

func getKeyValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	variables := mux.Vars(r)
	key := variables["key"]

	if key == "" {
		message := "Argument 'key' missing"
		log.Print(message)
		respond.With(w, r, http.StatusInternalServerError, message)
	}

	entry := (*raftServerReference.State)[key]
	data := ValueResponse{Value: entry.Value}
	log.Println("Getting key value")
	log.Println(helpers.PrettyPrint(entry))
	json.NewEncoder(w).Encode(data)
}

func acceptLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValueFromRequestArgs(r)
	if key == "" {
		return
	}
	helpers.Check(err)

	requestLogContext := log.WithFields(log.Fields{"request": r.Body, "key": key, "value": value})

	handler := command.NewAcceptLogEntryHandler(raftServerReference.AppRepository, raftServerReference, cancelService, rpcClientReference)
	cmd := command.AcceptLogEntry{Key: key, Value: value}
	err = handler.Handle(ctx, cmd)
	if err != nil {
		requestLogContext.WithField("error", err).Error("AcceptLogEntry failed")
		respond.With(w, r, http.StatusInternalServerError, PutResponse{Success: false})
	}
	respond.With(w, r, http.StatusOK, PutResponse{Success: true})
}

func getKeyAndValueFromRequestArgs(r *http.Request) (string, int, error) {
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
	fmt.Fprintf(w, "Raft module! RaftServerReference: "+raftServerReference.Id)
}
