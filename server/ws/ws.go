package ws

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type WsMessage struct {
	MessageType string `json:"type"`
}

type HeartbeatMessage struct {
	WsMessage,
	ServerId string `json:"serverId"`
}

func Reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		log.Println(string(p))
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}
