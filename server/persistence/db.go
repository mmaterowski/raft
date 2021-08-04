package persistence

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/consts"
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

type DbConfig struct {
	Path     string
	InMemory bool
}

type Db struct {
	AppRepository
}

func GetDbConfig(env string) DbConfig {
	config := DbConfig{}
	config.InMemory = env == consts.Local
	if env == consts.LocalWithPersistence {
		config.Path = "./log.db"
	} else {
		config.Path = "../data/raft-db/log.db"
	}
	return config
}

func NewDb(config DbConfig) *Db {
	db := Db{}
	if config.InMemory {
		log.Print("Using in memory db")
		db.AppRepository = &InMemoryContext{}
	} else {
		log.Print("Using sqllite db")
		log.Print("Db path: ", config.Path)
		context, err := NewSqlLiteRepository(config.Path)
		helpers.Check(err)
		db = Db{AppRepository: context}
	}

	return &db
}
