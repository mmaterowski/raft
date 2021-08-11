package client

import (
	log "github.com/sirupsen/logrus"

	"github.com/mmaterowski/raft/rpc/raft_rpc"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/helpers"
	"google.golang.org/grpc"
)

type client struct {
	kimClient         raft_rpc.RaftRpcClient
	rickyClient       raft_rpc.RaftRpcClient
	laszloClient      raft_rpc.RaftRpcClient
	setUpSuccessfully bool
}

var Client clientInterface = &client{}

type clientInterface interface {
	Setup(serverId string)
	GetFor(serverId string) raft_rpc.RaftRpcClient
}

func (r *client) Setup(serverId string) {
	if serverId == consts.KimId {
		rickyServerPort := "ricky:6960"
		rickyRpcClientConnection, err := grpc.Dial(rickyServerPort, grpc.WithInsecure())
		helpers.Check(err)
		r.rickyClient = raft_rpc.NewRaftRpcClient(rickyRpcClientConnection)

		laszloServerPort := "laszlo:6960"
		laszloRpcClientConnection, err := grpc.Dial(laszloServerPort, grpc.WithInsecure())
		helpers.Check(err)
		r.laszloClient = raft_rpc.NewRaftRpcClient(laszloRpcClientConnection)
		r.setUpSuccessfully = true

	}

	if serverId == consts.RickyId {
		kimServerPort := "kim:6960"
		kimRpcClientConnection, err := grpc.Dial(kimServerPort, grpc.WithInsecure())
		helpers.Check(err)
		r.kimClient = raft_rpc.NewRaftRpcClient(kimRpcClientConnection)

		laszloServerPort := "laszlo:6960"
		laszloRpcClientConnection, err := grpc.Dial(laszloServerPort, grpc.WithInsecure())
		helpers.Check(err)
		r.laszloClient = raft_rpc.NewRaftRpcClient(laszloRpcClientConnection)
		r.setUpSuccessfully = true

	}

	if serverId == consts.LaszloId {
		kimServerPort := "kim:6960"
		kimRpcClientConnection, err := grpc.Dial(kimServerPort, grpc.WithInsecure())
		helpers.Check(err)
		r.kimClient = raft_rpc.NewRaftRpcClient(kimRpcClientConnection)

		rickyServerPort := "ricky:6960"
		rickyRpcClientConnection, err := grpc.Dial(rickyServerPort, grpc.WithInsecure())
		helpers.Check(err)
		r.rickyClient = raft_rpc.NewRaftRpcClient(rickyRpcClientConnection)
		r.setUpSuccessfully = true

	}
	log.Printf("Raft clients set up succesfully: %t", r.setUpSuccessfully)

}

func (r *client) GetFor(serverId string) raft_rpc.RaftRpcClient {
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
