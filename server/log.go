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
		state[entry.Key] = entry.Value
	}
	return true
}
