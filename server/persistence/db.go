package persistence

import structs "github.com/mmaterowski/raft/structs"

type Setup interface {
	Setup()
}

type Context interface {
	PersistValue(key string, value int, termNumber int) (bool, structs.Entry)
	PersistValues(entries []structs.Entry) (bool, structs.Entry)
	GetEntryAtIndex(index int) bool
	GetLastEntry() structs.Entry
	GetCurrentTerm() int
	GetVotedFor() string
	SetCurrentTerm(currentTerm int) bool
	SetVotedFor(votedForId string) bool
	GetLog() []structs.Entry
}

type Db struct {
	Context
	Path string
}

func NewDb(debug bool) Db {
	dbPath := "../data/raft-db/log.db"
	if debug {
		dbPath = "./log.db"
	}
	db := Db{}
	db.Path = dbPath
	db.Context = NewSqlLiteDbContext(dbPath)
	return db
}
