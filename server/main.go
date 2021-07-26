package main

import (
	"fmt"
	"log"
	"net"
	"os"

	api "github.com/mmaterowski/raft/api"
	"github.com/mmaterowski/raft/consts"
	helpers "github.com/mmaterowski/raft/helpers"
	"github.com/mmaterowski/raft/persistence"
	protoBuff "github.com/mmaterowski/raft/raft_rpc"
	raft "github.com/mmaterowski/raft/raft_server"
	rpc "github.com/mmaterowski/raft/rpc"
	"google.golang.org/grpc"
)

func main() {
	helpers.PrintAsciiHelloString()
	env := getEnv()
	serverId := getServerId(env)
	config := persistence.GetDbConfig(env)
	db := persistence.NewDb(config)
	server := raft.Server{AppRepository: &db}
	server.StartServer(serverId)
	api.IdentifyServer(server.Id, isLocalEnvironment(env))

	go func() {
		err := handleRPC(serverId, server)
		helpers.Check(err)
	}()
	client := rpc.Client{}
	go func() {
		client.SetupRpcClient(server.Id)
	}()

	api.RaftServerReference = &server
	api.RpcClientReference = &client
	apiPort := getApiPort(env)
	api.HandleRequests(apiPort)

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

func handleRPC(serverId string, serverReference raft.Server) error {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	grpcServer := grpc.NewServer()
	protoBuff.RegisterRaftRpcServer(grpcServer, &rpc.Server{Server: serverReference})
	log.Printf("RPC listening on port: %s", lis.Addr().String())
	return grpcServer.Serve(lis)
}
