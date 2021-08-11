package persistence

import (
	"context"
	"log"

	"github.com/mmaterowski/raft/model/entry"
	"github.com/mmaterowski/raft/utils/consts"
	"github.com/mmaterowski/raft/utils/helpers"
)

var (
	Database   DatabaseInterface = &Db{}
	Repository AppRepository
	config     DbConfig = DbConfig{}
)

type AppRepository interface {
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

type Db struct {
}

type DatabaseInterface interface {
	Init(config DbConfig)
	GetConfig(env string) DbConfig
}

func (db *Db) GetConfig(env string) DbConfig {
	config = DbConfig{}
	config.InMemory = env == consts.Local
	if env == consts.LocalWithPersistence {
		config.Path = "./log.db"
	} else {

		config.Path = "../data/raft-db/log.db"
	}
	return config
}

func (db *Db) Init(config DbConfig) {
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
