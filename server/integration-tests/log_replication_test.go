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
	"github.com/mmaterowski/raft/consts"
	"github.com/mmaterowski/raft/helpers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var compose *testcontainers.LocalDockerCompose

func up() error {
	log.Printf("Clearing up db from last test")
	e := os.RemoveAll("../../db/")
	if e != nil {
		log.Fatal(e)
	}

	composeFilePaths := []string{"../../docker-compose.test.yaml"}
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
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", consts.KimPort))
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
	resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key, value))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
	var r api.ValueResponse
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.KimPort, key))
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value {
		t.Errorf("Get request to: %d. Expected %d but got %d", consts.KimPort, value, r.Value)
	}
	time.Sleep(consts.HeartbeatInterval + 1*time.Second)
	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, key))

	err = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value {
		t.Errorf("Get request to: %d. Expected %d but got %d", consts.RickyPort, value, r.Value)
	}
}
func TestLogRebuiltProperlyAfterFailure(t *testing.T) {
	deleteEntriesFromAllServers()
	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	value := 1
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key1, value))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key2, value+1))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key3, value+2))

	execError := compose.WithCommand([]string{"restart", "-t", "30", "kim"}).Invoke()
	time.Sleep(5 * time.Second)
	if execError.Error != nil {
		t.Error("Error restarting service...", execError)
	}

	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.KimPort, key3))
	var r api.ValueResponse
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value+2 {
		t.Errorf("Get request to: %d. Expected %d but got %d. Log rebuild failed", consts.KimPort, value, r.Value)
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
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key1, value))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key2, value+1))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key3, value+2))

	var r api.ValueResponse
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.KimPort, key1))
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != value {
		t.Errorf("Get request to: %d. Expected %d but got %d", consts.KimPort, value, r.Value)
	}

	otherPort := "6970"
	_, _ = http.Get(fmt.Sprintf("http://localhost:%s/backdoor/put/%s/%d", otherPort, outOfSyncKey, value))
	time.Sleep(2 * time.Second)

	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, outOfSyncKey, expectedOutOfSyncKeyValue))
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
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key1, value))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key2, value+1))
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key3, value+2))

	execError := compose.WithCommand([]string{"stop", "-t", "30", "ricky"}).Invoke()
	if execError.Error != nil {
		t.Error("Error stopping service...", execError)
	}

	rickyAsleepNewValue := 420
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key3, rickyAsleepNewValue))

	execError = compose.WithCommand([]string{"start", "ricky"}).Invoke()
	time.Sleep(5 * time.Second)

	if execError.Error != nil {
		t.Error("Error starting service...", execError)
	}

	newValueAfterRickyIsUp := 666
	time.Sleep(consts.HeartbeatInterval + 1*time.Second)
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, key1, newValueAfterRickyIsUp))
	time.Sleep(consts.HeartbeatInterval + 1*time.Second)
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, key3))
	var r api.ValueResponse
	err := json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}

	if r.Value != rickyAsleepNewValue {
		t.Errorf("Get request to: %d. Expected %d but got %d.", consts.RickyPort, rickyAsleepNewValue, r.Value)
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, key1))
	err = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if err != nil {
		t.Error("Couldn't read response body")
	}
	if r.Value != newValueAfterRickyIsUp {
		t.Errorf("Get request to: %d. Expected %d but got %d.", consts.RickyPort, newValueAfterRickyIsUp, r.Value)
	}
}

func TestFollowerMissingOneEntry(t *testing.T) {
	swapTestDataAndRestartContainers("./test data/missing-one-entry.db")
	missingKey := "hejka"
	missingKeyValue := 12
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, missingKey))
	var r api.ValueResponse
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value == missingKeyValue {
		t.Errorf("Expected key to be missing")
	}

	newKey := "newKey"
	newKeyValue := 20
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, newKey, newKeyValue))
	time.Sleep(2 * time.Second)
	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.KimPort, newKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != newKeyValue {
		t.Errorf("Expected value to be updated")
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, missingKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != missingKeyValue {
		t.Errorf("Expected follower to sync missing value")
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, newKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != newKeyValue {
		t.Errorf("Expected follower to have correct last entry value")
	}
}

func TestFollowerMissingMultipleEntries(t *testing.T) {
	swapTestDataAndRestartContainers("./test data/missing-multiple-entries.db")
	missingKey := "hejka"
	missingKeyValue := 12
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, missingKey))
	var r api.ValueResponse
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value == missingKeyValue {
		t.Errorf("Expected key to be missing")
	}

	newKey := "newKey"
	newKeyValue := 20
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, newKey, newKeyValue))
	time.Sleep(2 * time.Second)
	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.KimPort, newKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != newKeyValue {
		t.Errorf("Expected value to be updated")
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, missingKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != missingKeyValue {
		t.Errorf("Expected follower to sync missing value")
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, newKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != newKeyValue {
		t.Errorf("Expected follower to have correct last entry value")
	}
}

func TestFollowerHasExtraEntry(t *testing.T) {
	swapTestDataAndRestartContainers("./test data/missing-multiple-entries.db")
	extraKey := "d"
	extraKeyValue := 19
	getKeyResponse, _ := http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, extraKey))
	var r api.ValueResponse
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value == extraKeyValue {
		t.Errorf("Expected extra key to be present")
	}

	newKey := "newKey"
	newKeyValue := 20
	_, _ = http.Get(fmt.Sprintf("http://localhost:%d/put/%s/%d", consts.KimPort, newKey, newKeyValue))
	time.Sleep(2 * time.Second)
	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.KimPort, newKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != newKeyValue {
		t.Errorf("Expected value to be updated")
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, extraKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != extraKeyValue {
		t.Errorf("Expected follower to delete extra value")
	}

	getKeyResponse, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", consts.RickyPort, newKey))
	_ = json.NewDecoder(getKeyResponse.Body).Decode(&r)
	if r.Value != newKeyValue {
		t.Errorf("Expected follower to have correct last entry value")
	}
}

func swapTestDataAndRestartContainers(dbToSwap string) {
	helpers.Copy(dbToSwap, "../../db/ricky/log.db")
	helpers.Copy("./test data/leader.db", "../../db/kim/log.db")
	helpers.Copy("./test data/leader.db", "../../db/laszlo/log.db")

	_ = compose.WithCommand([]string{"restart", "-t", "30", consts.RickyServiceName}).Invoke()
	_ = compose.WithCommand([]string{"restart", "-t", "30", consts.KimServiceName}).Invoke()
	_ = compose.WithCommand([]string{"restart", "-t", "30", consts.LaszloServiceName}).Invoke()

}
