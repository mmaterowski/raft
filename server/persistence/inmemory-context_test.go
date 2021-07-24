package persistence

import (
	"context"
	"testing"

	"github.com/mmaterowski/raft/consts"
	"github.com/mmaterowski/raft/entry"
)

func TestInMemGetEntryAtIndex(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 0, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 1, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = inMemContext.PersistValues(ctx, entries)
	ent, _ := inMemContext.GetEntryAtIndex(ctx, 1)
	if ent.Key != "key2" {
		t.Error("Got invalid entry")
	}
}

func TestInMemPersistValueDoesNotFail(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, err := inMemContext.PersistValue(ctx, key, value, term)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}

	entry, getErr := inMemContext.GetEntryAtIndex(ctx, 0)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != value {
		t.Errorf("Expected to get %d, but got %d", value, entry.Value)
	}
}

func TestInMemPersistValuesDoesNotFail(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 0, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 1, Value: 2, Key: "key2", TermNumber: 35})
	_, err := inMemContext.PersistValues(ctx, entries)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}

	entry, getErr := inMemContext.GetEntryAtIndex(ctx, 0)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != 1 {
		t.Errorf("Got invalid value")
	}

	secondEntry, secondGetErr := inMemContext.GetEntryAtIndex(ctx, 1)
	if getErr != nil {
		t.Errorf("Expected no error: %s", secondGetErr)
	}
	if secondEntry.Value != 2 {
		t.Errorf("Got invalid value")
	}
}

func TestInMemDeletingAllEntries(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	key := "testVal"
	value := 12
	index := 0
	_, _ = inMemContext.PersistValue(ctx, key, value, index)
	_, _ = inMemContext.PersistValue(ctx, key, value+1, index)
	_, _ = inMemContext.PersistValue(ctx, key, value+2, index)
	inMemContext.DeleteAllEntriesStartingFrom(ctx, 0)
	if (len(inMemContext.entries)) != 0 {
		t.Errorf("Expected no entries")
	}
}

func TestInMemDeletingEntryOutsideOfRange(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, err := inMemContext.PersistValue(ctx, key, value, term)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	err = inMemContext.DeleteAllEntriesStartingFrom(ctx, 23)
	if err == nil {
		t.Errorf("Expected err")
	}
}

func TestInMemDeletingSlice(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, _ = inMemContext.PersistValue(ctx, key, value, term)
	_, _ = inMemContext.PersistValue(ctx, key, value+1, term)
	_, _ = inMemContext.PersistValue(ctx, key, value+2, term)

	_ = inMemContext.DeleteAllEntriesStartingFrom(ctx, 1)
	if (len(inMemContext.entries)) != 1 {
		t.Errorf("Expected single entry left")
	}
}

func TestInMemDeletingAndReapendingEntryWorks(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, err := inMemContext.PersistValue(ctx, key, value, term)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	inMemContext.DeleteAllEntriesStartingFrom(ctx, 0)
	if (len(inMemContext.entries)) != 0 {
		t.Errorf("Expected no entries")
	}

	_, _ = inMemContext.PersistValue(ctx, key, value, term)
	entry, getErr := inMemContext.GetEntryAtIndex(ctx, 0)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != value {
		t.Errorf("Got invalid value")
	}

}

func TestInMemSettingTerm(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	inMemContext.SetCurrentTerm(ctx, 2)
	term, _ := inMemContext.GetCurrentTerm(ctx)
	if term != 2 {
		t.Error("Got invalid term value")
	}
}

func TestInMemSettingVotedFor(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	inMemContext.SetVotedFor(ctx, "kim")
	votedFor, _ := inMemContext.GetVotedFor(ctx)
	if votedFor != "kim" {
		t.Error("Got invalid voted for value")
	}
}

func TestInMemGettingLastEntry(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 0, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 1, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = inMemContext.PersistValues(ctx, entries)
	ent, _ := inMemContext.GetLastEntry(ctx)
	if ent.Key != "key2" {
		t.Error("Got invalid last entry")
	}
}

func TestInMemGetLog(t *testing.T) {
	ctx := context.Background()
	inMemContext := InMemoryContext{}
	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 0, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 1, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = inMemContext.PersistValues(ctx, entries)
	raftLog, _ := inMemContext.GetLog(ctx)
	if (*raftLog)[1].Key != "key2" {
		t.Error("Goit invalid value from log")
	}
}
