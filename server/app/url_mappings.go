package app

import "github.com/mmaterowski/raft/controllers/ping"

func mapUrls() {
	router.GET("/ping", ping.Ping)
}
