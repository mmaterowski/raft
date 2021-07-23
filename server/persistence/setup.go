package persistence

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mmaterowski/raft/consts"
)

func (s *SqlLiteRepository) setDbHandle(dbPath string) bool {

	handle, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Print(err)
	}
	handle.SetMaxOpenConns(25)
	handle.SetMaxIdleConns(25)
	handle.SetConnMaxLifetime(5 * time.Minute)
	s.handle = handle
	return handle != nil
}

func (s SqlLiteRepository) InitTablesIfNeeded() bool {
	entriesCreated := s.createEntriesTableIfNotExists()
	termCreated := s.createTermTableIfNotExists()
	votedForCreated := s.createVotedForIfNotExists()
	entriesInitialized := insertUniqueEntriesForVotedForAndCurrentTerm(s)
	return entriesCreated && termCreated && votedForCreated && entriesInitialized
}

func insertUniqueEntriesForVotedForAndCurrentTerm(s SqlLiteRepository) bool {
	currentTerm, _ := s.GetCurrentTerm(context.Background())
	if currentTerm == consts.TermUninitializedValue {
		termInsertStatement, _ := s.handle.Prepare("INSERT INTO Term (CurrentTerm, UniqueEntryId) VALUES (?, ?)")
		_, termErr := termInsertStatement.Exec(consts.TermInitialValue, s.uniqueEntryId)
		if termErr != nil {
			return false
		}

	}

	_, err := s.GetVotedFor(context.Background())
	if err == sql.ErrNoRows {
		votedForInsertStatement, _ := s.handle.Prepare("INSERT INTO VotedFor (VotedForId, UniqueEntryId) VALUES (?, ?)")
		_, votedForErr := votedForInsertStatement.Exec("", s.uniqueEntryId)
		if votedForErr != nil {
			return false
		}
	}

	return true

}

func (s SqlLiteRepository) createEntriesTableIfNotExists() bool {
	createStatement, err := s.handle.Prepare(`CREATE TABLE IF NOT EXISTS "Entries" (
		"Index"	INTEGER,
		"Value"	INTEGER,
		"Key"	TEXT,
		"TermNumber"	INTEGER,
		PRIMARY KEY("Index" AUTOINCREMENT)
	)`)
	if err != nil {
		log.Print(err)
		return false
	}
	_, exceuteError := createStatement.Exec()
	return exceuteError == nil

}

func (s SqlLiteRepository) createTermTableIfNotExists() bool {
	createStatement, err := s.handle.Prepare(`CREATE TABLE IF NOT EXISTS "Term" (
		"CurrentTerm"	INTEGER
	, "UniqueEntryId"	TEXT)`)
	if err != nil {
		log.Print(err)
		return false
	}
	_, exceuteError := createStatement.Exec()
	return exceuteError == nil

}

func (s SqlLiteRepository) createVotedForIfNotExists() bool {
	createStatement, err := s.handle.Prepare(`CREATE TABLE IF NOT EXISTS "VotedFor" (
		"VotedForId"	TEXT,
		"UniqueEntryId"	TEXT
	)`)
	if err != nil {
		log.Print(err)
	}
	_, exceuteError := createStatement.Exec()
	return exceuteError == nil
}
