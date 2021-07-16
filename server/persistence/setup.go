package persistence

import (
	"database/sql"
	"log"
	"time"
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

func (s SqlLiteRepository) initTablesIfNeeded() bool {
	entriesCreated := s.createEntriesTableIfNotExists()
	termCreated := s.createTermTableIfNotExists()
	votedForCreated := s.createVotedForIfNotExists()
	return entriesCreated && termCreated && votedForCreated
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
	return exceuteError != nil

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
	return exceuteError != nil

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
	return exceuteError != nil
}
