package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, vars["key"])
	fmt.Println("Endpoint Hit: getKey")
}

func putKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value := vars["value"]
	fmt.Fprintf(w, "Putting key!\n")
	fmt.Fprintf(w, "Key: "+key+"\nValue: "+value)
	fmt.Println("Endpoint Hit: putKey")
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func handleRequests() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", getKey)
	r.HandleFunc("/get/{key}", getKey)
	r.HandleFunc("/put/{key}/{value}", putKey)
	log.Fatal(http.ListenAndServe(":10000", nil))
}
