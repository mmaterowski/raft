package persistence

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmaterowski/raft/consts"
	"github.com/mmaterowski/raft/entry"
)

func TestSqlLiteGetEntryAtIndex(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 0, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 1, Value: 2, Key: "key2", TermNumber: 35})

	_, _ = db.PersistValues(ctx, entries)
	ent, _ := db.GetEntryAtIndex(ctx, 2)
	if ent.Key != "key2" {
		t.Error("Got invalid entry")
	}
}

func TestSqlLitePersistValueDoesNotFail(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	termNumber := 1
	_, err := db.PersistValue(ctx, key, value, termNumber)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}

	entry, getErr := db.GetEntryAtIndex(ctx, 1)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != value {
		t.Errorf("Expected to get %d, but got %d", value, entry.Value)
	}
}

func TestSqlLitePersistValuesDoesNotFail(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 1, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 2, Value: 2, Key: "key2", TermNumber: 35})
	_, err := db.PersistValues(ctx, entries)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}

	entry, getErr := db.GetEntryAtIndex(ctx, 1)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != 1 {
		t.Errorf("Got invalid value")
	}

	secondEntry, secondGetErr := db.GetEntryAtIndex(ctx, 2)
	if getErr != nil {
		t.Errorf("Expected no error: %s", secondGetErr)
	}
	if secondEntry.Value != 2 {
		t.Errorf("Got invalid value")
	}
}

func TestSqLiteDeletingAllEntries(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	index := 0
	_, _ = db.PersistValue(ctx, key, value, index)
	_, _ = db.PersistValue(ctx, key, value+1, index)
	_, _ = db.PersistValue(ctx, key, value+2, index)
	db.DeleteAllEntriesStartingFrom(ctx, 1)
	entries, _ := db.GetLog(ctx)
	if (len(*entries)) != 0 {
		t.Errorf("Expected no entries")
	}
}

func TestSqlLiteDeletingEntryOutsideOfRange(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	index := 1
	_, err := db.PersistValue(ctx, key, value, index)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	err = db.DeleteAllEntriesStartingFrom(ctx, 23)
	if err != nil {
		t.Errorf("Expected no err")
	}

	entries, _ := db.GetLog(ctx)
	if (len(*entries)) != 1 {
		t.Errorf("Expected delete operation not to affect log")
	}
}

func TestSqlLiteDeletingSlice(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, _ = db.PersistValue(ctx, key, value, term)
	_, _ = db.PersistValue(ctx, key, value+1, term)
	_, _ = db.PersistValue(ctx, key, value+2, term)

	_ = db.DeleteAllEntriesStartingFrom(ctx, 2)
	entries, _ := db.GetLog(ctx)
	if (len(*entries)) != 1 {
		t.Errorf("Expected single entry left")
	}
}

func TestSqlLiteDeletingAndReapendingEntryWorks(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, err := db.PersistValue(ctx, key, value, term)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	entries, _ := db.GetLog(ctx)
	if (len(*entries)) != 0 {
		t.Errorf("Expected no entries")
	}

	_, _ = db.PersistValue(ctx, key, value, term)
	entry, getErr := db.GetEntryAtIndex(ctx, 1)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != value {
		t.Errorf("Got invalid value")
	}

}

func TestSqlLiteSettingTerm(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)

	db.SetCurrentTerm(ctx, 2)
	term, _ := db.GetCurrentTerm(ctx)
	if term != 2 {
		t.Error("Got invalid term value")
	}
}

func TestSqlLiteSettingVotedFor(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)

	db.SetVotedFor(ctx, "kim")
	votedFor, _ := db.GetVotedFor(ctx)
	if votedFor != "kim" {
		t.Error("Got invalid voted for value")
	}
}

func TestSqlLiteGettingLastEntry(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 1, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 2, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = db.PersistValues(ctx, entries)
	ent, _ := db.GetLastEntry(ctx)
	if ent.Key != "key2" {
		t.Error("Got invalid last entry")
	}
}

func TestSqlLiteGetLog(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	db := NewDb(config)
	//Cleanup
	db.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 1, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 2, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = db.PersistValues(ctx, entries)
	raftLog, _ := db.GetLog(ctx)
	if (*raftLog)[1].Key != "key2" {
		t.Error("Goit invalid value from log")
	}
}
