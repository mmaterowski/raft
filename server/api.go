package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gopkg.in/matryer/respond.v1"
)

type StatusResponse struct {
	Status ServerType
}
type ValueResponse struct {
	Value string
}

type PutResponse struct {
	Success bool
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{Status: serverType}
	respond.With(w, r, http.StatusOK, data)
}

func getKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	variables := mux.Vars(r)
	key := variables["key"]

	valueFromRedis, err := redisClient.Get(key).Result()
	if err != nil {
		log.Println(err)
	}

	data := ValueResponse{Value: valueFromRedis}
	respond.With(w, r, http.StatusOK, data)
}

func putKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	variables := mux.Vars(r)
	key := variables["key"]
	value := variables["value"]

	_, err := redisClient.Set(key, value, 0).Result()
	if err != nil {
		log.Println(err)
	}

	data := PutResponse{Success: true}
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
	r.HandleFunc("/get/{key}", getKey)
	r.HandleFunc("/put/{key}/{value}", putKey)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), nil))
}
