package main

import (
	pb "raft/raft_rpc"

	"google.golang.org/grpc"
)

var kimRpcClient pb.RaftRpcClient
var rickyRpcClient pb.RaftRpcClient
var laszloRpcClient pb.RaftRpcClient
var rpcClientSetUpSuccessfull bool = false

func SetupRpcClient() {
	if serverId == "Kim" {
		rickyServerPort := "ricky:6960"
		rickyRpcClientConnection, err := grpc.Dial(rickyServerPort, grpc.WithInsecure())
		Check(err)
		rickyRpcClient = pb.NewRaftRpcClient(rickyRpcClientConnection)

		laszloServerPort := "laszlo:6960"
		laszloRpcClientConnection, err := grpc.Dial(laszloServerPort, grpc.WithInsecure())
		Check(err)
		laszloRpcClient = pb.NewRaftRpcClient(laszloRpcClientConnection)
		rpcClientSetUpSuccessfull = true

	}

	if serverId == "Ricky" {
		kimServerPort := "kim:6960"
		kimRpcClientConnection, err := grpc.Dial(kimServerPort, grpc.WithInsecure())
		Check(err)
		kimRpcClient = pb.NewRaftRpcClient(kimRpcClientConnection)

		laszloServerPort := "laszlo:6960"
		laszloRpcClientConnection, err := grpc.Dial(laszloServerPort, grpc.WithInsecure())
		Check(err)
		laszloRpcClient = pb.NewRaftRpcClient(laszloRpcClientConnection)
		rpcClientSetUpSuccessfull = true

	}

	if serverId == "Laszlo" {
		kimServerPort := "kim:6960"
		kimRpcClientConnection, err := grpc.Dial(kimServerPort, grpc.WithInsecure())
		Check(err)
		kimRpcClient = pb.NewRaftRpcClient(kimRpcClientConnection)

		rickyServerPort := "ricky:6960"
		rickyRpcClientConnection, err := grpc.Dial(rickyServerPort, grpc.WithInsecure())
		Check(err)
		rickyRpcClient = pb.NewRaftRpcClient(rickyRpcClientConnection)
		rpcClientSetUpSuccessfull = true

	}

}

func GetClientFor(serverId string) pb.RaftRpcClient {
	if !rpcClientSetUpSuccessfull {
		panic("Clients not set up!")
	}
	if serverId == "Kim" {
		return kimRpcClient
	}
	if serverId == "Ricky" {
		return rickyRpcClient
	}
	if serverId == "Laszlo" {
		return laszloRpcClient
	}
	panic("Wrong server name")
}
