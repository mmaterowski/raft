package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	command "github.com/mmaterowski/raft/command"
	"github.com/mmaterowski/raft/model/entry"
	server_model "github.com/mmaterowski/raft/model/server"
	"github.com/mmaterowski/raft/persistence"
	"github.com/mmaterowski/raft/server"
	"github.com/mmaterowski/raft/utils/consts"
	rest_errors "github.com/mmaterowski/raft/utils/errors"
	"github.com/mmaterowski/raft/utils/helpers"
	raft_ws "github.com/mmaterowski/raft/ws/"
	log "github.com/sirupsen/logrus"
	"gopkg.in/matryer/respond.v1"
)

var port string

type StatusResponse struct {
	Status server_model.ServerType
}

type ValueResponse struct {
	Value int
}

type PutResponse struct {
	Success bool
}

func InitApi(Port string) {

	if Port == "" {
		log.Fatal("No api port specified")
	}

	port = Port
}

func HandleRequests() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", home)
	r.HandleFunc("/ws", wsEndpoint)
	r.HandleFunc("/status", getStatus)
	r.HandleFunc("/get/{key}", getKeyValue)
	r.HandleFunc("/put/{key}/{value}", acceptLogEntry)
	r.HandleFunc("/backdoor/put/{key}/{value}", persistAndCommitValue)
	r.HandleFunc("/backdoor/deleteall", deleteAllEntries)
	log.Printf("API listens on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")
	_ = ws.WriteJSON(`{"message":"Hi Client!"}`)
	raft_ws.Reader(ws)

}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: server.Type}
	respond.With(w, r, http.StatusOK, data)
}

func persistAndCommitValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key, value, err := getKeyAndValueFromRequestArgs(r)
	helpers.Check(err)
	entry, _ := persistence.Repository.PersistValue(r.Context(), key, value, server.State.CurrentTerm)
	(*server.State.Entries)[entry.Key] = *entry
	server.State.CommitIndex = entry.Index
	data := PutResponse{Success: !entry.IsEmpty()}
	log.Print("Backdooring entry. Persisted and commited ", entry)
	respond.With(w, r, http.StatusOK, data)
}

func deleteAllEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = persistence.Repository.DeleteAllEntriesStartingFrom(r.Context(), 1)
	(*server.State.Entries) = map[string]entry.Entry{}
	server.State.CommitIndex = consts.LeaderCommitInitialValue
	server.State.PreviousEntryIndex = consts.NoPreviousEntryValue
	server.State.PreviousEntryTerm = consts.TermInitialValue
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

	entry := (*server.State.Entries)[key]
	data := ValueResponse{Value: entry.Value}
	log.Println("Getting key value")
	log.Println(helpers.PrettyPrint(entry))
	json.NewEncoder(w).Encode(data)
}

func acceptLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	//change http to accept json key+value
	//make service accept new struct NewEntryViewModel
	//validate using separate method

	key, value, err := getKeyAndValueFromRequestArgs(r)

	if err != nil {
		restErr := rest_errors.NewBadRequestError("Error getting key and value from request body")
		respond.With(w, r, restErr.Status, restErr)
		return
	}

	requestLogContext := log.WithFields(log.Fields{"request": r.Body, "key": key, "value": value})

	handler := command.NewAcceptLogEntryHandler()
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
	fmt.Fprintf(w, "Raft module! RaftServerReference: "+server.Id)
}
