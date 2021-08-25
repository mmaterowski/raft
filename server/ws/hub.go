package ws

import (
	"encoding/json"

	model "github.com/mmaterowski/raft/model/server"
	log "github.com/sirupsen/logrus"
)

var AppHub *Hub = newHub()

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered Clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) SendHeartbeatReceivedMessage(id string) {
	message := newHeartbeatReceivedMessage(id)
	h.sendMessage(message)
}

func (h *Hub) ServerTypeChangedMessage(id string, serverType model.ServerType) {
	message := newServerTypeChangedMessage(id, serverType)
	h.sendMessage(message)
}

func (h *Hub) SendNewKeyUpdatedMessage(id, key string, value int) {
	message := NewKeyUpdatedMessage(id, key, value)
	h.sendMessage(message)
}

func (h *Hub) sendMessage(message interface{}) {
	b, err := json.Marshal(message)
	if err != nil {
		log.WithFields(log.Fields{"Message": message, "Error": err}).Error("Couldn't serialize message")
	}
	h.Broadcast <- b
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
