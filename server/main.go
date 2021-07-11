package main

var server raftServer

func main() {
	PrintAsciiHelloString()
	server1 := "Kim"
	// server2 := "Ricky"
	// server3 := "Laszlo"
	server = raftServer{}
	server.startServer(server1)
}
