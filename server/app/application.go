package app

import (
	"fmt"
	"net"
	"os"
	"strconv"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	api "github.com/mmaterowski/raft/api"
	"github.com/mmaterowski/raft/persistence"
	rpcClient "github.com/mmaterowski/raft/rpc/client"
	protoBuff "github.com/mmaterowski/raft/rpc/raft_rpc"
	rpcServer "github.com/mmaterowski/raft/rpc/server"
	raftServer "github.com/mmaterowski/raft/server"
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
	config := persistence.Database.GetStandardConfig(env)
	persistence.Database.Init(config)
	rpcClient.Client.Setup(serverId)
	raftServer.Raft.StartServer(serverId, isLocalEnvironment(env), isIntegrationTesting(env))

	go handleRPC()
	// go handleSignalR(getSignalRPort(env))

	// if env == consts.Integrationtest {
	// 	if raftServer.Id == consts.KimId {
	// 		raftServer.Type = server_model.Leader
	// 	}
	// }

	// if env != consts.Integrationtest {
	// 	raftServer.Raft.SetupElection()
	// }
	raftServer.Raft.SetupElection()

	// _____________________________________START HERE________________________________________
	//HEARTBEAT NOT SENT
	//=INFINITE ELECTION
	raftServer.Raft.StartHeartbeat()

	// mapUrls()
	// router.Run(":8080")

	api.InitApi(getApiPort(env))
	api.HandleRequests()
}

// func handleSignalR(port string) {

// 	address := ":" + port
// 	hub := raft_signalr.AppHub

// 	server, _ := signalr.NewServer(context.TODO(),
// 		signalr.SimpleHubFactory(&hub.Hub),
// 		signalr.KeepAliveInterval(2*time.Second))

// 	router := http.NewServeMux()
// 	server.MapHTTP(router, "/signalr")
// 	log.WithField("port", address).Info("Listening for websocket connections.")
// 	if err := http.ListenAndServe(address, router); err != nil {
// 		log.WithField("Error", err).Fatal("SignalR ListenAndServe failed")
// 	}

// }

func getApiPort(env string) string {
	port := os.Getenv("SERVER_PORT")
	if isLocalEnvironment(env) {
		port = strconv.Itoa(consts.LocalApiPort)
	}
	return port
}

func getSignalRPort(env string) string {
	port := os.Getenv("SERVER_SIGNALR_PORT")
	if isLocalEnvironment(env) {
		port = strconv.Itoa(consts.LocalSignalRPort)
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

func handleRPC() {
	port := 6960
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	helpers.Check(err)
	grpcServer := grpc.NewServer()
	protoBuff.RegisterRaftRpcServer(grpcServer, &rpcServer.Server{})
	err = grpcServer.Serve(lis)
	helpers.Check(err)
	log.WithField("port", lis.Addr().String()).Info("RPC started listening")
}
