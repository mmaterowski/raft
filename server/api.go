package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"gopkg.in/matryer/respond.v1"
)

var laszloAddress = "laszlo:6971"
var rickyAddress = "ricky:6970"
var kimAddress = "ricky:6969"
var others []string

type StatusResponse struct {
	Status ServerType
}
type ValueResponse struct {
	Value int
}

type PutResponse struct {
	Success bool
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: serverType}
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

	entry := state[key]
	data := ValueResponse{Value: entry.Value}
	respond.With(w, r, http.StatusOK, data)
}

func AcceptLogEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValue(r)
	Check(err)
	success, entry := PersistValue(key, value, currentTerm)
	entries := make(map[string]Entry)
	entries[entry.Key] = entry
	makeSureLastEntryDataIsAvailable()

	// request := AppendEntriesRequest{
	// 	Term: currentTerm, LeaderId: serverId, PreviousLogIndex: previousEntryIndex, Entries: entries, LeaderCommitIndex: commitIndex,
	// }
	var wg sync.WaitGroup

	wg.Add((len(others) / 2) + 1)
	// for _, otherServer := range others {
	// go func(request AppendEntriesRequest, otherServer string) {
	// 	term, dataMatch := AppendEntriesRPC(request, otherServer)
	// 	log.Print(term, dataMatch)
	// 	defer wg.Done()
	// }(request, otherServer)
	// }
	wg.Wait()
	//majority accepted, can go on

	state[entry.Key] = entry
	commitIndex = entry.Index

	data := PutResponse{Success: success}
	respond.With(w, r, http.StatusOK, data)
}

func makeSureLastEntryDataIsAvailable() {
	if previousEntryIndex < 0 || previousEntryTerm < 0 {
		entry := GetLastEntry()
		previousEntryIndex = entry.Index
		previousEntryTerm = entry.TermNumber
	}
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
	fmt.Fprintf(w, "Raft module! Server: "+serverId)
}

func handleRequests() {
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
