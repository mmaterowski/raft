package main

import (
	"context"
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
	Local           = "Local"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "Local"
	}
	log.Printf("ENVIRONMENT: %s", env)
	helpers.PrintAsciiHelloString()
	serverId := os.Getenv("SERVER_ID")
	if env == Local {
		serverId = "Kim"
	}

	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}
	useInMemoryDb := env == Debug || env == Integrationtest || env == Local
	db := persistence.NewDb(useInMemoryDb)
	if useInMemoryDb {
		db.SetCurrentTerm(context.Background(), 1)
	}
	server := raft.Server{AppRepository: &db}
	server.StartServer(serverId)
	api.IdentifyServer(server.Id, env == Local)

	go func() {
		err := handleRPC(serverId, server)
		helpers.Check(err)
	}()
	client := rpc.Client{}
	go func() {
		client.SetupRpcClient(server.Id)
	}()
	port := os.Getenv("SERVER_PORT")
	if env == Local {
		port = "6969"
	}
	api.RaftServerReference = &server
	api.RpcClientReference = &client
	api.HandleRequests(port)

}

func handleRPC(serverId string, serverReference raft.Server) error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	grpcServer := grpc.NewServer()
	protoBuff.RegisterRaftRpcServer(grpcServer, &rpc.Server{Server: serverReference})
	log.Printf("RPC listening on port: %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}
