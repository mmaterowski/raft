package ws

import model "github.com/mmaterowski/raft/model/server"

const (
	HeartbeatReceivedMessageType = "HeatbeatReceivedMessage"
	ServerTypeChangedMessageType = "ServerTypeChangedMessage"
	KeyUpdatedMessageType        = "KeyUpdatedMessage"
)

type WsMessage struct {
	MessageType string `json:"type"`
	ServerId    string `json:"serverId"`
}

type HeartbeatReceivedMessage struct {
	WsMessage
}

func newHeartbeatReceivedMessage(serverId string) *HeartbeatReceivedMessage {
	msg := &HeartbeatReceivedMessage{}
	msg.MessageType = HeartbeatReceivedMessageType
	msg.ServerId = serverId
	return msg
}

type ServerTypeChangedMessage struct {
	WsMessage
	Type model.ServerType `json:"serverType"`
}

func newServerTypeChangedMessage(serverId string, serverType model.ServerType) *ServerTypeChangedMessage {
	msg := &ServerTypeChangedMessage{}
	msg.MessageType = ServerTypeChangedMessageType
	msg.ServerId = serverId
	msg.Type = serverType
	return msg
}

type KeyUpdatedMessage struct {
	WsMessage
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func NewKeyUpdatedMessage(serverId, key string, value int) *KeyUpdatedMessage {
	msg := &KeyUpdatedMessage{}
	msg.ServerId = serverId
	msg.Key = key
	msg.Value = value
	msg.MessageType = KeyUpdatedMessageType
	return msg
}
