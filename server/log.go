package main

type Entry struct {
	Index      int
	Value      int
	Key        string
	TermNumber int
}

func RebuildStateFromLog() bool {
	entries := GetLog()
	for _, entry := range entries {
		server.state[entry.Key] = entry
	}
	return true
}
