package main

import (
	"fmt"
	"net"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	api "github.com/mmaterowski/raft/api"
	"github.com/mmaterowski/raft/consts"
	helpers "github.com/mmaterowski/raft/helpers"
	"github.com/mmaterowski/raft/persistence"
	protoBuff "github.com/mmaterowski/raft/raft_rpc"
	raft "github.com/mmaterowski/raft/raft_server"
	rpc "github.com/mmaterowski/raft/rpc"
	rpcClient "github.com/mmaterowski/raft/rpc_client"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var server *raft.Server

func main() {

	log.SetFormatter(&nested.Formatter{})

	helpers.PrintAsciiHelloString()
	env := getEnv()
	serverId := getServerId(env)
	config := persistence.GetDbConfig(env)
	db := persistence.NewDb(config)
	client := rpcClient.NewClient(serverId)
	server = &raft.Server{AppRepository: db, RpcClient: *client}
	server.StartServer(serverId, isLocalEnvironment(env))

	go handleRPC(server)
	server.SetupElection()
	server.StartHeartbeat()

	api.RaftServerReference = server
	api.RpcClientReference = client
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
	log.WithField("value", env).Info("ENVIRONMENT")
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

func handleRPC(serverReference *raft.Server) {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	grpcServer := grpc.NewServer()
	protoBuff.RegisterRaftRpcServer(grpcServer, &rpc.Server{Server: *serverReference})
	err = grpcServer.Serve(lis)
	helpers.Check(err)
	log.WithField("port", lis.Addr().String()).Info("RPC started listening")
}
