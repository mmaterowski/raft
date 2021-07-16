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

var debug = true

func main() {
	helpers.PrintAsciiHelloString()

	serverId := os.Getenv("SERVER_ID")
	if debug {
		serverId = "Kim"
	}
	if serverId == "" {
		log.Fatal("Server id not set. Check Your environmental variable 'SERVER_ID'")
	}

	db := persistence.NewDb(debug)

	server := raft.Server{AppRepository: &db}
	server.StartServer(serverId)
	api.IdentifyServer(server.Id, debug)

	go func() {
		err := handleRPC()
		helpers.Check(err)
	}()

	go func() {
		client := rpc.Client{}
		client.SetupRpcClient(server.Id)
	}()
	port := os.Getenv("SERVER_PORT")
	if debug {
		port = "6969"
	}
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
