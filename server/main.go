package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
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
var mu sync.Mutex

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

	client := &rpc.Client{}

	go func(env string) {
		client.SetupRpcClient(server.Id)
		//Until election implemented kim is the leader
		SetupElection(client, others)
		if serverId == consts.KimId {
			StartHeartbeat(client, others)
		}
	}(env)

	api.RaftServerReference = server
	api.RpcClientReference = client
	apiPort := getApiPort(env)

	api.HandleRequests(apiPort)

}

func SetupElection(c *rpc.Client, others []string) {
	go func() {
		for {
			select {
			case <-server.ElectionTicker.C:
				log.Print("Ticker timeout: Start election...")
				server.CurrentTerm++
				server.VotedFor = server.Id
				server.SetVotedFor(context.Background(), server.Id)
				log.Printf("%s issues Election, incrementing current term to: %d", server.Id, server.CurrentTerm)

				mu.Lock()
				server.ServerType = structs.Candidate
				server.HeartbeatTicker.Stop()
				mu.Unlock()

				for _, otherServer := range others {
					go func(serverId string) {
						request := protoBuff.RequestVoteRequest{Term: int32(server.CurrentTerm), CandidateID: server.Id, LastLogIndex: int32(server.PreviousEntryIndex), LastLogTerm: int32(server.PreviousEntryTerm)}
						reply, err := c.GetClientFor(serverId).RequestVote(context.Background(), &request)
						if err != nil {
							log.Print("Request vote error: ", err)
						}
						if reply == nil {
							log.Print("No errors but nil reply, something werid happened")
							return
						}

						if reply.Term > int32(server.CurrentTerm) {
							log.Print("Response term higher than candidate's, setting current term to: ", reply.Term)
							server.CurrentTerm = int(reply.Term)
							server.SetCurrentTerm(context.Background(), int(reply.Term))
						}

						log.Print("Got RequestVoteResponse, voteGranted: ", reply.VoteGranted)
						log.Print("Reply term: ", reply.Term, "Server current term: ", server.CurrentTerm)
						log.Print("Server state: ", server.ServerType)
						if reply.VoteGranted && reply.Term <= int32(server.CurrentTerm) && server.ServerType != structs.Leader {
							log.Printf("%s becoming a leader", server.Id)
							server.ServerType = structs.Leader
							server.HeartbeatTicker = time.NewTicker(consts.HeartbeatInterval)
							server.TriggerHeartbeat <- struct{}{}
							server.ResetElectionTicker <- struct{}{}
						}
					}(otherServer)
				}
			case <-server.ResetElectionTicker:
				seed := rand.NewSource(time.Now().UnixNano())
				electionTimeout := rand.New(seed).Intn(100)*300 + 100
				log.Printf("Resetting ticker, election timeout: %d + %d", electionTimeout, consts.HeartbeatInterval)
				server.ElectionTicker.Reset(consts.HeartbeatInterval + time.Duration(electionTimeout)*time.Millisecond)
			}
		}
	}()
}

func StartHeartbeat(c *rpc.Client, others []string) {
	server.HeartbeatTicker.Stop()
	go func() {
		for {
			select {
			case <-server.HeartbeatTicker.C:
				server.ResetElectionTicker <- struct{}{}
				for _, otherServer := range others {
					go func(commitIndex int, term int, id string, otherServer string) {
						log.Print("Sending heartbeat. Leader commit index: ", commitIndex)
						request := protoBuff.AppendEntriesRequest{LeaderCommitIndex: int32(commitIndex), Term: int32(term), LeaderId: id}
						c.GetClientFor(otherServer).AppendEntries(context.Background(), &request)
					}(server.CommitIndex, server.CurrentTerm, server.Id, otherServer)
				}
			case <-server.TriggerHeartbeat:
				log.Printf("Heartbeat was resetted, expect to send hearbeat after: %d", consts.HeartbeatInterval)
				//just to recalculate select values
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
