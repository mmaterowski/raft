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

	"github.com/google/uuid"
	api "github.com/mmaterowski/raft/api"
	"github.com/testcontainers/testcontainers-go"
)

func up(identifier string) error {
	composeFilePaths := []string{"../docker-compose.test.yaml"}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.WithCommand([]string{"up", "-d"}).Invoke()
	//.WithEnv(map[string]string{
	// 	"key1": "value1",
	// 	"key2": "value2",
	// })

	err := execError.Error
	if err != nil {
		return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}
	return nil
}

func down(identifier string) error {
	log.Printf("Clearing up containers")
	composeFilePaths := []string{"../docker-compose.test.yaml"}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.Down()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}
	log.Printf("Shut down properly")
	return nil
}

func TestMain(m *testing.M) {
	identifier := strings.ToLower(uuid.New().String())
	err := up(identifier)
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	down(identifier)

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

func TestPutKeyIsReplicatedOnAllMachines(t *testing.T) {
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
