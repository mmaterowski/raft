package persistence

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmaterowski/raft/model/entry"
	"github.com/mmaterowski/raft/utils/consts"
)

func TestSqlLiteGetEntryAtIndex(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 0, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 1, Value: 2, Key: "key2", TermNumber: 35})

	_, _ = Repository.PersistValues(ctx, entries)
	ent, _ := Repository.GetEntryAtIndex(ctx, 2)
	if ent.Key != "key2" {
		t.Error("Got invalid entry")
	}
}

func TestSqlLitePersistValueDoesNotFail(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	termNumber := 1
	_, err := Repository.PersistValue(ctx, key, value, termNumber)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}

	entry, getErr := Repository.GetEntryAtIndex(ctx, 1)
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
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 1, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 2, Value: 2, Key: "key2", TermNumber: 35})
	_, err := Repository.PersistValues(ctx, entries)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}

	entry, getErr := Repository.GetEntryAtIndex(ctx, 1)
	if getErr != nil {
		t.Errorf("Expected no error: %s", getErr)
	}
	if entry.Value != 1 {
		t.Errorf("Got invalid value")
	}

	secondEntry, secondGetErr := Repository.GetEntryAtIndex(ctx, 2)
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
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	index := 0
	_, _ = Repository.PersistValue(ctx, key, value, index)
	_, _ = Repository.PersistValue(ctx, key, value+1, index)
	_, _ = Repository.PersistValue(ctx, key, value+2, index)
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)
	entries, _ := Repository.GetLog(ctx)
	if (len(*entries)) != 0 {
		t.Errorf("Expected no entries")
	}
}

func TestSqlLiteDeletingEntryOutsideOfRange(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	index := 1
	_, err := Repository.PersistValue(ctx, key, value, index)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	err = Repository.DeleteAllEntriesStartingFrom(ctx, 23)
	if err != nil {
		t.Errorf("Expected no err")
	}

	entries, _ := Repository.GetLog(ctx)
	if (len(*entries)) != 1 {
		t.Errorf("Expected delete operation not to affect log")
	}
}

func TestSqlLiteDeletingSlice(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, _ = Repository.PersistValue(ctx, key, value, term)
	_, _ = Repository.PersistValue(ctx, key, value+1, term)
	_, _ = Repository.PersistValue(ctx, key, value+2, term)

	_ = Repository.DeleteAllEntriesStartingFrom(ctx, 2)
	entries, _ := Repository.GetLog(ctx)
	if (len(*entries)) != 1 {
		t.Errorf("Expected single entry left")
	}
}

func TestSqlLiteDeletingAndReapendingEntryWorks(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	key := "testVal"
	value := 12
	term := consts.TermInitialValue
	_, err := Repository.PersistValue(ctx, key, value, term)
	if err != nil {
		t.Errorf("Expected no error: %s", err)
	}
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	entries, _ := Repository.GetLog(ctx)
	if (len(*entries)) != 0 {
		t.Errorf("Expected no entries")
	}

	_, _ = Repository.PersistValue(ctx, key, value, term)
	entry, getErr := Repository.GetEntryAtIndex(ctx, 1)
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
	Database.Init(config)

	Repository.SetCurrentTerm(ctx, 2)
	term, _ := Repository.GetCurrentTerm(ctx)
	if term != 2 {
		t.Error("Got invalid term value")
	}
}

func TestSqlLiteSettingVotedFor(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)

	Repository.SetVotedFor(ctx, "kim")
	votedFor, _ := Repository.GetVotedFor(ctx)
	if votedFor != "kim" {
		t.Error("Got invalid voted for value")
	}
}

func TestSqlLiteGettingLastEntry(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 1, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 2, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = Repository.PersistValues(ctx, entries)
	ent, _ := Repository.GetLastEntry(ctx)
	if ent.Key != "key2" {
		t.Error("Got invalid last entry")
	}
}

func TestSqlLiteGetLog(t *testing.T) {
	ctx := context.Background()
	config := DbConfig{Path: "./unit_test.db", InMemory: false}
	Database.Init(config)
	//Cleanup
	Repository.DeleteAllEntriesStartingFrom(ctx, 1)

	var entries []entry.Entry
	entries = append(entries, entry.Entry{Index: 1, Value: 1, Key: "key1", TermNumber: 34}, entry.Entry{Index: 2, Value: 2, Key: "key2", TermNumber: 35})
	_, _ = Repository.PersistValues(ctx, entries)
	raftLog, _ := Repository.GetLog(ctx)
	if (*raftLog)[1].Key != "key2" {
		t.Error("Goit invalid value from log")
	}
}
