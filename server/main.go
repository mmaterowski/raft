package main

import (
	"fmt"
	"log"
	"net"
	"os"

	api "github.com/mmaterowski/raft/api"
	helpers "github.com/mmaterowski/raft/helpers"
	"github.com/mmaterowski/raft/persistence"
	protoBuff "github.com/mmaterowski/raft/raft_rpc"
	raft "github.com/mmaterowski/raft/raft_server"
	rpc "github.com/mmaterowski/raft/rpc"
	"google.golang.org/grpc"
)

const (
	Debug           = "Debug"
	Integrationtest = "IntegrationTest"
	Prod            = "Prod"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "Debug"
	}
	helpers.PrintAsciiHelloString()
	serverId := os.Getenv("SERVER_ID")
	if env == Debug {
		serverId = "Kim"
	}

	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}
	useInMemoryDb := env == Debug || env == Integrationtest
	db := persistence.NewDb(useInMemoryDb)

	server := raft.Server{AppRepository: &db}
	server.StartServer(serverId)
	api.IdentifyServer(server.Id, env == Debug)

	go func() {
		err := handleRPC()
		helpers.Check(err)
	}()

	go func() {
		client := rpc.Client{}
		client.SetupRpcClient(server.Id)
	}()
	port := os.Getenv("SERVER_PORT")
	if env == Debug {
		port = "6969"
	}
	api.RaftServerReference = &server
	api.HandleRequests(port)

}

func handleRPC() error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	grpcServer := grpc.NewServer()
	protoBuff.RegisterRaftRpcServer(grpcServer, &rpc.Server{})
	log.Printf("RPC listening on port: %d", port)
	return grpcServer.Serve(lis)
}
