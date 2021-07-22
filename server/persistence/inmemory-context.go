package persistence

import (
	"context"

	"github.com/mmaterowski/raft/entry"
)

type InMemoryContext struct {
	Context,
	entries []entry.Entry
	currentTerm int
	votedFor    string
}

func (c *InMemoryContext) PersistValue(ctx context.Context, key string, value int, termNumber int) (*entry.Entry, error) {
	e := entry.Entry{Index: len(c.entries), Value: value, Key: key, TermNumber: termNumber}
	c.entries = append(c.entries, e)
	return &e, nil
}

func (c *InMemoryContext) PersistValues(ctx context.Context, entries []entry.Entry) (*entry.Entry, error) {
	c.entries = append(c.entries, entries...)
	return &entries[len(entries)-1], nil
}

func (c InMemoryContext) GetEntryAtIndex(ctx context.Context, index int) (*entry.Entry, error) {
	if index >= len(c.entries) {
		return &entry.Entry{}, nil
	}

	return &c.entries[index], nil
}

func (c InMemoryContext) GetLastEntry(ctx context.Context) (*entry.Entry, error) {
	return &c.entries[len(c.entries)-1], nil
}

func (c InMemoryContext) GetCurrentTerm(ctx context.Context) (int, error) {
	return c.currentTerm, nil
}

func (c InMemoryContext) GetVotedFor(ctx context.Context) (string, error) {
	return c.votedFor, nil
}

func (c *InMemoryContext) SetCurrentTerm(ctx context.Context, currentTerm int) error {
	c.currentTerm = currentTerm
	return nil
}

func (c *InMemoryContext) SetVotedFor(ctx context.Context, votedForId string) error {
	c.votedFor = votedForId
	return nil
}

func (c InMemoryContext) GetLog(ctx context.Context) (*[]entry.Entry, error) {
	return &c.entries, nil
}

func (c *InMemoryContext) DeleteAllEntriesStartingFrom(ctx context.Context, index int) error {
	if index == 0 {
		c.entries = make([]entry.Entry, 0)
		return nil
	}
	c.entries = c.entries[index:len(c.entries)]
	return nil

}
