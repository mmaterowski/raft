package persistence

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/entry"
	"github.com/mmaterowski/raft/helpers"
)

type Setup interface {
	Setup()
}

type AppRepository interface {
	PersistValue(ctx context.Context, key string, value int, termNumber int) (*entry.Entry, error)
	PersistValues(ctx context.Context, entries []entry.Entry) (*entry.Entry, error)
	GetEntryAtIndex(ctx context.Context, index int) (*entry.Entry, error)
	GetLastEntry(ctx context.Context) (*entry.Entry, error)
	GetCurrentTerm(ctx context.Context) (int, error)
	GetVotedFor(ctx context.Context) (string, error)
	SetCurrentTerm(ctx context.Context, currentTerm int) error
	SetVotedFor(ctx context.Context, votedForId string) error
	GetLog(ctx context.Context) (*[]entry.Entry, error)
	DeleteAllEntriesStartingFrom(ctx context.Context, index int) error
}

type Db struct {
	AppRepository
	Path string
}

func NewDb(debug bool) Db {
	dbPath := "../data/raft-db/log.db"
	if debug {
		dbPath = "./log.db"
	}
	db := Db{}
	if debug {
		log.Print("Using in memory db")
		db.AppRepository = &InMemoryContext{}
	} else {
		log.Print("Using sqllite db")
		context, err := NewSqlLiteRepository(dbPath)
		helpers.Check(err)

		db = Db{Path: dbPath, AppRepository: context}
	}

	return db
}
