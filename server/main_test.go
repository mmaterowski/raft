// +build integration
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	api "github.com/mmaterowski/raft/api"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var compose *testcontainers.LocalDockerCompose

func up() error {
	log.Printf("Clearing up db from last test")
	e := os.RemoveAll("../db/")
	if e != nil {
		log.Fatal(e)
	}

	composeFilePaths := []string{"../docker-compose.test.yaml"}
	identifier := strings.ToLower(uuid.New().String())
	compose = testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.WithCommand([]string{"up", "-d"}).Invoke()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}

	kimPort := 6969
	laszloPort := 6970
	rickyPort := 6971
	wait.ForHTTP(fmt.Sprintf("http://localhost:%d/", kimPort))
	wait.ForHTTP(fmt.Sprintf("http://localhost:%d/", laszloPort))
	wait.ForHTTP(fmt.Sprintf("http://localhost:%d/", rickyPort))

	return nil
}

func down() error {
	execError := compose.Down()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("Could not run compose file: - %v", err)
	}

	log.Printf("Shut down properly")
	return nil
}

func TestMain(m *testing.M) {
	err := up()
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	down()

	os.Exit(code)
}

func TestCheckAvailability(t *testing.T) {
	port := "6969"
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s", port))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}

func deleteEntriesFromAllServers() {
	kimPort := 6969
	laszloPort := 6970
	rickyPort := 6971
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/backdoor/deleteall", kimPort))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/backdoor/deleteall", laszloPort))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/backdoor/deleteall", rickyPort))
}

func TestPutKeyIsReplicatedOnAllMachines(t *testing.T) {
	deleteEntriesFromAllServers()
	key := "key"
	value := 3
	kimPort := "6969"
	resp, _ := http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key, value))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	var r api.ValueResponse
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%s/get/%s", kimPort, key))
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value {
		t.Errorf("Get request to: %s. Expected %d but got %d", kimPort, value, r.Value)
	}

	otherPort := "6970"
	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%s/get/%s", otherPort, key))

	err = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value {
		t.Errorf("Get request to: %s. Expected %d but got %d", otherPort, value, r.Value)
	}

}
func TestLogRebuiltProperlyAfterFailure(t *testing.T) {
	deleteEntriesFromAllServers()
	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	value := 1
	kimPort := "6969"
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key1, value))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key2, value+1))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key3, value+2))

	execError := compose.WithCommand([]string{"restart", "-t", "30", "kim"}).Invoke()
	time.Sleep(5 * time.Second)
	if execError.Error != nil {
		t.Error("Error restarting service...", execError)
	}

	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%s/get/%s", kimPort, key3))
	var r api.ValueResponse
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value+2 {
		t.Errorf("Get request to: %s. Expected %d but got %d. Log rebuild failed", kimPort, value, r.Value)
	}

}

func TestLeaderForcingFollowerToSyncLog(t *testing.T) {
	deleteEntriesFromAllServers()
	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	outOfSyncKey := "key4"
	expectedOutOfSyncKeyValue := 10
	value := 1
	kimPort := "6969"
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key1, value))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key2, value+1))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key3, value+2))

	var r api.ValueResponse
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%s/get/%s", kimPort, key1))
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value {
		t.Errorf("Get request to: %s. Expected %d but got %d", kimPort, value, r.Value)
	}

	otherPort := "6970"
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/backdoor/put/%s/%d", otherPort, outOfSyncKey, value))
	time.Sleep(2 * time.Second)

	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, outOfSyncKey, expectedOutOfSyncKeyValue))
	time.Sleep(2 * time.Second)

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%s/get/%s", otherPort, outOfSyncKey))
	err = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != expectedOutOfSyncKeyValue {
		t.Errorf("Get request to: %s. Expected %d but got %d", otherPort, value, r.Value)
	}

}

func TestLogSyncedAfterServiceNotWorkingForAWhile(t *testing.T) {
	deleteEntriesFromAllServers()
	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	value := 1
	kimPort := "6969"
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key1, value))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key2, value+1))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key3, value+2))

	execError := compose.WithCommand([]string{"stop", "-t", "30", "ricky"}).Invoke()
	if execError.Error != nil {
		t.Error("Error stopping service...", execError)
	}

	rickyAsleepNewValue := 420
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key3, rickyAsleepNewValue))

	execError = compose.WithCommand([]string{"start", "ricky"}).Invoke()
	time.Sleep(5 * time.Second)

	if execError.Error != nil {
		t.Error("Error starting service...", execError)
	}

	newValueAfterRickyIsUp := 666
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/put/%s/%d", kimPort, key1, newValueAfterRickyIsUp))
	time.Sleep(2 * time.Second)
	rickyPort := 6970
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", rickyPort, key3))
	var r api.ValueResponse
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != rickyAsleepNewValue {
		t.Errorf("Get request to: %d. Expected %d but got %d.", rickyPort, rickyAsleepNewValue, r.Value)
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", rickyPort, key1))
	err = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}
	if r.Value != newValueAfterRickyIsUp {
		t.Errorf("Get request to: %d. Expected %d but got %d.", rickyPort, newValueAfterRickyIsUp, r.Value)
	}

}
