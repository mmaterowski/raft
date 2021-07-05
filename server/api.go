package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/matryer/respond.v1"
)

type StatusResponse struct {
	Status ServerType
}
type ValueResponse struct {
	Value int
}

type PutResponse struct {
	Success bool
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: serverType}
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

	valueFromState := state[key]
	data := ValueResponse{Value: valueFromState}
	respond.With(w, r, http.StatusOK, data)
}

func putKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	variables := mux.Vars(r)
	key := variables["key"]

	if key == "" {
		message := "Argument 'key' missing"
		log.Print(message)
		respond.With(w, r, http.StatusInternalServerError, message)
	}

	value, convError := strconv.Atoi(variables["value"])
	if convError != nil {
		log.Print(convError)
		respond.With(w, r, http.StatusInternalServerError, convError)
		return
	}

	success := putValue(key, value, currentTerm)
	data := PutResponse{Success: success}
	respond.With(w, r, http.StatusOK, data)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Raft module! Server: "+serverId)
}

func handleRequests() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", home)
	r.HandleFunc("/status", getStatus)
	r.HandleFunc("/get/{key}", getKeyValue)
	r.HandleFunc("/put/{key}/{value}", putKey)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), nil))
}
