package persistence

import (
	"database/sql"
	"log"
	"time"
)

func (s *SqlLiteDbContext) setDbHandle(dbPath string) bool {

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

func (s SqlLiteDbContext) initTablesIfNeeded() bool {
	entriesCreated := s.createEntriesTableIfNotExists()
	termCreated := s.createTermTableIfNotExists()
	votedForCreated := s.createVotedForIfNotExists()
	return entriesCreated && termCreated && votedForCreated
}

func (s SqlLiteDbContext) createEntriesTableIfNotExists() bool {
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
	success, _ := executeSafely(createStatement)
	return success

}

func (s SqlLiteDbContext) createTermTableIfNotExists() bool {
	createStatement, err := s.handle.Prepare(`CREATE TABLE IF NOT EXISTS "Term" (
		"CurrentTerm"	INTEGER
	, "UniqueEntryId"	TEXT)`)
	if err != nil {
		log.Print(err)
		return false
	}
	success, _ := executeSafely(createStatement)
	return success

}

func (s SqlLiteDbContext) createVotedForIfNotExists() bool {
	createStatement, err := s.handle.Prepare(`CREATE TABLE IF NOT EXISTS "VotedFor" (
		"VotedForId"	TEXT,
		"UniqueEntryId"	TEXT
	)`)
	if err != nil {
		log.Print(err)
	}
	success, _ := executeSafely(createStatement)
	return success
}
