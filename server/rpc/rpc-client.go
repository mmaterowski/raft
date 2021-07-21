package rpc

import (
	"log"

	. "github.com/mmaterowski/raft/helpers"
	pb "github.com/mmaterowski/raft/raft_rpc"
	"google.golang.org/grpc"
)

type Client struct {
	kimClient         pb.RaftRpcClient
	rickyClient       pb.RaftRpcClient
	laszloClient      pb.RaftRpcClient
	setUpSuccessfully bool
}

func (r *Client) SetupRpcClient(serverId string) {
	if serverId == "Kim" {
		rickyServerPort := "ricky:6960"
		rickyRpcClientConnection, err := grpc.Dial(rickyServerPort, grpc.WithInsecure())
		Check(err)
		r.rickyClient = pb.NewRaftRpcClient(rickyRpcClientConnection)

		laszloServerPort := "laszlo:6960"
		laszloRpcClientConnection, err := grpc.Dial(laszloServerPort, grpc.WithInsecure())
		Check(err)
		r.laszloClient = pb.NewRaftRpcClient(laszloRpcClientConnection)
		r.setUpSuccessfully = true

	}

	if serverId == "Ricky" {
		kimServerPort := "kim:6960"
		kimRpcClientConnection, err := grpc.Dial(kimServerPort, grpc.WithInsecure())
		Check(err)
		r.kimClient = pb.NewRaftRpcClient(kimRpcClientConnection)

		laszloServerPort := "laszlo:6960"
		laszloRpcClientConnection, err := grpc.Dial(laszloServerPort, grpc.WithInsecure())
		Check(err)
		r.laszloClient = pb.NewRaftRpcClient(laszloRpcClientConnection)
		r.setUpSuccessfully = true

	}

	if serverId == "Laszlo" {
		kimServerPort := "kim:6960"
		kimRpcClientConnection, err := grpc.Dial(kimServerPort, grpc.WithInsecure())
		Check(err)
		r.kimClient = pb.NewRaftRpcClient(kimRpcClientConnection)

		rickyServerPort := "ricky:6960"
		rickyRpcClientConnection, err := grpc.Dial(rickyServerPort, grpc.WithInsecure())
		Check(err)
		r.rickyClient = pb.NewRaftRpcClient(rickyRpcClientConnection)
		r.setUpSuccessfully = true

	}
	log.Printf("Raft clients set up succesfully: %t", r.setUpSuccessfully)

}

func (r Client) GetClientFor(serverId string) pb.RaftRpcClient {
	if !r.setUpSuccessfully {
		panic("Clients not set up!")
	}
	if serverId == "Kim" {
		return r.kimClient
	}
	if serverId == "Ricky" {
		return r.rickyClient
	}
	if serverId == "Laszlo" {
		return r.laszloClient
	}
	panic("Wrong server name")
}
