package main

import (
	"database/sql"
	"log"
	"time"
)

var db *sql.DB

func setupDB() bool {
	connected := connectToSql()
	if !connected {
		log.Print("Error connecting to SQL")
		return false
	}
	tablesInitialized := initTablesIfNeeded()
	return tablesInitialized
}

func connectToSql() bool {
	dbPath := "../data/raft-db/log.db"
	if debug {
		dbPath = "./log.db"
	}
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Print(err)
	}
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(25)
	database.SetConnMaxLifetime(5 * time.Minute)
	db = database
	return database != nil
}

func initTablesIfNeeded() bool {
	entriesCreated := createEntriesTableIfNotExists()
	termCreated := createTermTableIfNotExists()
	votedForCreated := createVotedForIfNotExists()
	return entriesCreated && termCreated && votedForCreated
}

func createEntriesTableIfNotExists() bool {
	createStatement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS "Entries" (
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
	success := executeSafely(createStatement)
	return success

}

func createTermTableIfNotExists() bool {
	createStatement, err := db.Prepare(`CREATE TABLE "Term" (
		"currentTerm"	INTEGER
	, "UniqueEntryId"	INTEGER)`)
	if err != nil {
		log.Print(err)
		return false
	}
	success := executeSafely(createStatement)
	return success

}

func createVotedForIfNotExists() bool {
	createStatement, err := db.Prepare(`CREATE TABLE "VotedFor" (
		"Id"	TEXT,
		"UniqueEntryId"	TEXT
	)`)
	if err != nil {
		log.Print(err)
	}
	success := executeSafely(createStatement)
	return success
}

func executeSafely(sqlStatement *sql.Stmt, args ...interface{}) bool {
	if args != nil {
		_, errWithArgs := sqlStatement.Exec(args)
		if errWithArgs != nil {
			log.Print(errWithArgs)
			return false

		}
		return true
	}
	_, err := sqlStatement.Exec()
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}
