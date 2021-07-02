package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/matryer/respond.v1"
)

type StatusResponse struct {
	Status ServerType
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := StatusResponse{}
	data.Status = serverType
	respond.With(w, r, http.StatusOK, data)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Master server to accept requests and pass them to Raft!")
}

func handleRequests() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", home)
	r.HandleFunc("/status", getStatus)
	log.Fatal(http.ListenAndServe(":6969", nil))
}
