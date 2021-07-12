package main

import (
	"fmt"
	"log"
	"net"

	. "github.com/mmaterowski/raft/api"
	. "github.com/mmaterowski/raft/helpers"
	. "github.com/mmaterowski/raft/raft_rpc"
	. "github.com/mmaterowski/raft/raft_server"
	. "github.com/mmaterowski/raft/rpc"
	"google.golang.org/grpc"
)

var debug = true

func main() {
	PrintAsciiHelloString()
	server1 := "Kim"
	// server2 := "Ricky"
	// server3 := "Laszlo"
	server := RaftServer{}
	server.StartServer(server1, debug)
	IdentifyServer(server.Id, debug)

	go func() {
		err := handleRPC()
		Check(err)
	}()

	go func() {
		client := Client{}
		client.SetupRpcClient(server.Id)
	}()

	HandleRequests()

}

func handleRPC() error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	Check(err)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	RegisterRaftRpcServer(grpcServer, &Server{})
	log.Printf("RPC listening on port: %d", port)
	return grpcServer.Serve(lis)
}
