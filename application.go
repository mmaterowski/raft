package app

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	api "github.com/mmaterowski/raft/api"
	"github.com/mmaterowski/raft/persistence"
	rpcClient "github.com/mmaterowski/raft/rpc/client"
	protoBuff "github.com/mmaterowski/raft/rpc/raft_rpc"
	rpcServer "github.com/mmaterowski/raft/rpc/server"
	raftServer "github.com/mmaterowski/raft/server"
	raft_signalr "github.com/mmaterowski/raft/signalr"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/helpers"
	"github.com/rs/cors"
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

	//SETUP SOCKETSIO WITH CORS PROPERLY
	go handleSocketIo(getSocketPort(env))

	raftServer.Raft.SetupElection()

	raftServer.Raft.StartHeartbeat()

	api.InitApi(getApiPort(env))
	api.HandleRequests()
}

func handleSocketIo(port string) {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", raft_signalr.HeartbeatReceivedMessage, func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	log.Println("Serving ws at: " + port)
	handler := cors.Default().Handler(server)
	log.Fatal(http.ListenAndServeTLS(":"+port, "./localhost.crt", "./localhost.key", handler))
}

func getApiPort(env string) string {
	port := os.Getenv("SERVER_PORT")
	if isLocalEnvironment(env) {
		port = strconv.Itoa(consts.LocalApiPort)
	}
	return port
}

func getSocketPort(env string) string {
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
