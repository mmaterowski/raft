package app

import (
	"fmt"
	"net"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	api "github.com/mmaterowski/raft/api"
	cancelService "github.com/mmaterowski/raft/cancel_service"
	server_model "github.com/mmaterowski/raft/model/server"
	"github.com/mmaterowski/raft/persistence"
	rpcClient "github.com/mmaterowski/raft/rpc/client"
	protoBuff "github.com/mmaterowski/raft/rpc/raft_rpc"
	rpc "github.com/mmaterowski/raft/rpc/raft_rpc"
	raft "github.com/mmaterowski/raft/server"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/helpers"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var router = gin.Default()

func StartApplication() {

	log.SetFormatter(&nested.Formatter{})
	helpers.PrintAsciiHelloString()
	env := getEnv()
	serverId := getServerId(env)
	config := persistence.GetDbConfig(env)
	db := persistence.NewDb(config)
	client := rpcClient.NewClient(serverId)
	server := &raft.Server{AppRepository: db, RpcClient: *client}
	server.StartServer(serverId, isLocalEnvironment(env), isIntegrationTesting(env))

	go handleRPC(server)

	if env == consts.Integrationtest {
		if server.Id == consts.KimId {
			server.ServerType = server_model.Leader
		}
	}

	if env != consts.Integrationtest {
		server.SetupElection()
	}

	server.StartHeartbeat()

	mapUrls()
	router.Run("8080")

	api.InitApi(server, client, cancelService.NewSyncRequestService(), getApiPort(env))
	api.HandleRequests()
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

func isIntegrationTesting(env string) bool {
	return env == consts.Integrationtest
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
