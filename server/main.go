package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	api "github.com/mmaterowski/raft/api"
	"github.com/mmaterowski/raft/consts"
	helpers "github.com/mmaterowski/raft/helpers"
	"github.com/mmaterowski/raft/persistence"
	protoBuff "github.com/mmaterowski/raft/raft_rpc"
	raft "github.com/mmaterowski/raft/raft_server"
	rpc "github.com/mmaterowski/raft/rpc"
	"github.com/mmaterowski/raft/structs"
	"google.golang.org/grpc"
)

var server *raft.Server

func main() {
	helpers.PrintAsciiHelloString()
	env := getEnv()
	serverId := getServerId(env)
	config := persistence.GetDbConfig(env)
	db := persistence.NewDb(config)
	server = &raft.Server{AppRepository: &db}
	server.StartServer(serverId)
	others := api.IdentifyServer(server.Id, isLocalEnvironment(env))

	go func() {
		err := handleRPC(serverId, server)
		helpers.Check(err)
	}()
	client := rpc.Client{}

	go func(env string) {
		client.SetupRpcClient(server.Id)
		//Until election implemented kim is the leader
		SetupElection(client, others)
		if serverId == consts.KimId {
			StartHeartbeat(client, others)
		}
	}(env)

	api.RaftServerReference = server
	api.RpcClientReference = &client
	apiPort := getApiPort(env)

	api.HandleRequests(apiPort)

}

func SetupElection(c rpc.Client, others []string) {
	seed := rand.NewSource(time.Now().UnixNano())
	electionTimeout := rand.New(seed).Intn(100)*300 + 100
	log.Printf("Election timeout set to: %d", electionTimeout)
	server.ElectionTicker = time.NewTicker(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)

	go func() {
		select {
		case <-server.ElectionTicker.C:
			log.Print("Ticker timeout: Start election...")
			server.CurrentTerm++
			log.Printf("%s issues Election, incrementing current term to: %d", server.Id, server.CurrentTerm)
			server.ServerType = structs.Candidate

			for _, otherServer := range others {
				go func(serverId string) {
					log.Printf("Request vote from %s", serverId)
					// request := protoBuff.RequestVoteRequest{Term: int32(server.CurrentTerm), CandidateID: server.Id, LastLogIndex: int32(server.PreviousEntryIndex), LastLogTerm: int32(server.PreviousEntryTerm)}
					//  reply, err := c.GetClientFor(otherServer).RequestVote(context.Background(), &request)
				}(otherServer)
			}

		}
	}()
}

func StartHeartbeat(c rpc.Client, others []string) {
	ticker := time.NewTicker(consts.HeartbeatInterval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, otherServer := range others {
					go func(commitIndex int, term int, id string, otherServer string) {
						log.Print("Sending heartbeat. Leader commit index: ", commitIndex)
						request := protoBuff.AppendEntriesRequest{LeaderCommitIndex: int32(commitIndex), Term: int32(term), LeaderId: id}
						c.GetClientFor(otherServer).AppendEntries(context.Background(), &request)
					}(server.CommitIndex, server.CurrentTerm, server.Id, otherServer)
				}
			case <-quit:
				ticker.Stop()
			}
		}
	}()
}

func getApiPort(env string) string {
	port := os.Getenv("SERVER_PORT")
	if isLocalEnvironment(env) {
		port = "6969"
	}
	return port
}

func getEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = consts.Local
	}
	log.Printf("ENVIRONMENT: %s", env)
	return env
}

func getServerId(env string) string {
	serverId := ""
	if isLocalEnvironment(env) {
		serverId = "Kim"
	} else {
		serverId = os.Getenv("SERVER_ID")
	}

	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}
	return serverId
}

func isLocalEnvironment(env string) bool {
	return env == consts.LocalWithPersistence || env == consts.Local
}

func handleRPC(serverId string, serverReference *raft.Server) error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	grpcServer := grpc.NewServer()
	protoBuff.RegisterRaftRpcServer(grpcServer, &rpc.Server{Server: *serverReference})
	log.Printf("RPC listening on port: %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}
