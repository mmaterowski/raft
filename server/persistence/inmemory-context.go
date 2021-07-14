package persistence

import structs "github.com/mmaterowski/raft/structs"

type InMemoryContext struct {
	Context,
	entries []structs.Entry
	currentTerm int
	votedFor    string
}

func (c *InMemoryContext) PersistValue(key string, value int, termNumber int) (bool, structs.Entry) {
	e := structs.Entry{Index: len(c.entries) - 1, Value: value, Key: key, TermNumber: termNumber}
	c.entries = append(c.entries, e)
	return true, e
}

func (c *InMemoryContext) PersistValues(entries []structs.Entry) (bool, structs.Entry) {
	c.entries = append(c.entries, entries...)
	return true, entries[len(entries)-1]
}

func (c InMemoryContext) GetEntryAtIndex(index int) (structs.Entry, error) {
	return c.entries[index], nil
}

func (c InMemoryContext) GetLastEntry() structs.Entry {
	return c.entries[len(c.entries)-1]
}

func (c InMemoryContext) GetCurrentTerm() int {
	return c.currentTerm
}

func (c InMemoryContext) GetVotedFor() string {
	return c.votedFor
}

func (c *InMemoryContext) SetCurrentTerm(currentTerm int) bool {
	c.currentTerm = currentTerm
	return true
}

func (c *InMemoryContext) SetVotedFor(votedForId string) bool {
	c.votedFor = votedForId
	return true
}

func (c InMemoryContext) GetLog() []structs.Entry {
	return c.entries
}

func (c *InMemoryContext) DeleteAllEntriesStartingFrom(index int) bool {
	c.entries = c.entries[index:len(c.entries)]
	return true
}
