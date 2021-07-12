package main

import (
	"fmt"
	"log"
	"net"

	api "github.com/mmaterowski/raft/api"
	helpers "github.com/mmaterowski/raft/helpers"
	raftRpc "github.com/mmaterowski/raft/raft_rpc"
	raftServer "github.com/mmaterowski/raft/raft_server"
	rpc "github.com/mmaterowski/raft/rpc"
	"google.golang.org/grpc"
)

var debug = true

func main() {
	helpers.PrintAsciiHelloString()
	server1 := "Kim"
	// server2 := "Ricky"
	// server3 := "Laszlo"
	server := raftServer.RaftServer{}
	server.StartServer(server1, debug)
	api.IdentifyServer(server.Id, debug)

	go func() {
		err := handleRPC()
		helpers.Check(err)
	}()

	go func() {
		client := rpc.Client{}
		client.SetupRpcClient(server.Id)
	}()

	api.HandleRequests()

}

func handleRPC() error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	raftRpc.RegisterRaftRpcServer(grpcServer, &rpc.Server{})
	log.Printf("RPC listening on port: %d", port)
	return grpcServer.Serve(lis)
}
