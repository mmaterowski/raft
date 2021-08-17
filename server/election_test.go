package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mmaterowski/raft/api"
	"github.com/mmaterowski/raft/model/server"
	"github.com/mmaterowski/raft/utils/consts"
)

func TestLogRebuiltProperlyAfterWonElection(t *testing.T) {
	useTestDbForAll()
	time.Sleep(5 * time.Second)

	var r api.StatusResponse
	leaderId := ""
	resp, _ := http.Get(fmt.Sprintf("http://localhost:%d/status", consts.KimPort))
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if r.Status == server.Leader {
		leaderId = consts.KimId
	}

	resp, _ = http.Get(fmt.Sprintf("http://localhost:%d/status", consts.RickyPort))
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if r.Status == server.Leader {
		leaderId = consts.RickyId
	}
	resp, _ = http.Get(fmt.Sprintf("http://localhost:%d/status", consts.LaszloPort))
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if r.Status == server.Leader {
		leaderId = consts.LaszloId
	}

	if leaderId == "" {
		t.Errorf("Leader expected, but none of servers is.")
	}

	var vr api.ValueResponse
	key := "hejka"
	expectedValue := 12
	resp, _ = http.Get(fmt.Sprintf("http://localhost:%d/get/%s", getPortFor(leaderId), key))
	_ = json.NewDecoder(resp.Body).Decode(&vr)

	if vr.Value != expectedValue {
		t.Errorf("Expected %d, but got %d", expectedValue, vr.Value)
	}

}

func getPortFor(id string) int {
	if id == consts.KimId {
		return consts.KimPort
	}
	if id == consts.RickyId {
		return consts.RickyPort
	}
	if id == consts.LaszloId {
		return consts.LaszloPort
	}
	return 0
}
