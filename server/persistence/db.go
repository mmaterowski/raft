package persistence

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/model/entry"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/helpers"
)

var (
	Database   databaseInterface = &db{}
	Repository appRepository
	config     DbConfig = DbConfig{}
)

type appRepository interface {
	//NewEntryVM
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

type db struct {
}

type databaseInterface interface {
	Init(config DbConfig)
	GetStandardConfig(env string) DbConfig
}

func (db *db) GetStandardConfig(env string) DbConfig {
	config = DbConfig{}
	config.InMemory = env == consts.Local
	if env == consts.LocalWithPersistence {
		config.Path = "./log.db"
	} else {
		config.Path = "../data/raft-db/log.db"
	}
	return config
}

func (db *db) Init(config DbConfig) {
	if config.InMemory {
		log.Print("Using in memory db")
		Repository = &InMemoryContext{}
	} else {
		log.Print("Using sqllite db")
		log.Print("Db path: ", config.Path)
		context, err := NewSqlLiteRepository(config.Path)
		helpers.Check(err)
		Repository = context
	}
}
