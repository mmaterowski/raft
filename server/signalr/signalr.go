package raft_signalr

import (
	"github.com/mmaterowski/raft/model/server"
	"github.com/philippseith/signalr"
	log "github.com/sirupsen/logrus"
)

type appHub struct {
	signalr.Hub
}

var AppHub = &appHub{}

const (
	ServerTypeChangedMessage = "serverTypeChanged"
)

func (h *appHub) SendServerTypeChanged(serverId string, serverType server.ServerType) {
	h.Clients().All().Send(ServerTypeChangedMessage, serverId, serverType)
}

func (h *appHub) OnConnected(connectionID string) {
	log.Info("SignalR: %s connected", connectionID)
}

func (h *appHub) OnDisconnected(connectionID string) {
	log.Info("SignalR: %s disconnected", connectionID)
}
