// +build integration
package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

var identifier string

func up() error {
	composeFilePaths := []string{"../docker-compose.test.yaml"}

	identifier = strings.ToLower(uuid.New().String())

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

func down() error {
	composeFilePaths := []string{"../docker-compose.test.yaml"}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.Down()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}
	return nil
}

func TestNginxLatestReturn(t *testing.T) {

	err := up()
	defer down()
	if err != nil {
		t.Error(err)
	}
	port := "6969"
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s", port))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
}
