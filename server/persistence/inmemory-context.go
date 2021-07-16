package persistence

import (
	"github.com/mmaterowski/raft/entry"
)

type InMemoryContext struct {
	Context,
	entries []entry.Entry
	currentTerm int
	votedFor    string
}

func (c *InMemoryContext) PersistValue(key string, value int, termNumber int) (bool, entry.Entry) {
	e := entry.Entry{Index: len(c.entries), Value: value, Key: key, TermNumber: termNumber}
	c.entries = append(c.entries, e)
	return true, e
}

func (c *InMemoryContext) PersistValues(entries []entry.Entry) (bool, entry.Entry) {
	c.entries = append(c.entries, entries...)
	return true, entries[len(entries)-1]
}

func (c InMemoryContext) GetEntryAtIndex(index int) (entry.Entry, error) {
	if index >= len(c.entries) {
		return entry.Entry{}, nil
	}

	return c.entries[index], nil
}

func (c InMemoryContext) GetLastEntry() entry.Entry {
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

func (c InMemoryContext) GetLog() []entry.Entry {
	return c.entries
}

func (c *InMemoryContext) DeleteAllEntriesStartingFrom(index int) bool {
	c.entries = c.entries[index:len(c.entries)]
	return true
}
