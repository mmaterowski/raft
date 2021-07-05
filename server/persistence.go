package main

import (
	"database/sql"
	"log"
	"time"
)

var db *sql.DB

func ConnectToSql() bool {
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

func InitTablesIfNeeded() {
	createEntriesTableIfNotExists()
	createTermTableIfNotExists()
	createVotedForIfNotExists()
}

func createEntriesTableIfNotExists() {
	createStatement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS "Entries" (
		"Index"	INTEGER,
		"Value"	INTEGER,
		"Key"	TEXT,
		"TermNumber"	INTEGER,
		PRIMARY KEY("Index" AUTOINCREMENT)
	)`)
	if err != nil {
		log.Print(err)
	}
	executeSafely(createStatement)

}

func createTermTableIfNotExists() {
	createStatement, err := db.Prepare(`CREATE TABLE "Term" (
		"currentTerm"	INTEGER
	, "UniqueEntryId"	INTEGER)`)
	if err != nil {
		log.Print(err)
	}
	executeSafely(createStatement)

}

func createVotedForIfNotExists() {
	createStatement, err := db.Prepare(`CREATE TABLE "VotedFor" (
		"Id"	TEXT,
		"UniqueEntryId"	TEXT
	)`)
	if err != nil {
		log.Print(err)
	}
	executeSafely(createStatement)

}

func executeSafely(sqlStatement *sql.Stmt, args ...interface{}) {
	if args != nil {
		_, errWithArgs := sqlStatement.Exec(args)
		if errWithArgs != nil {
			log.Print(errWithArgs)
		}
		return
	}
	_, err := sqlStatement.Exec()
	if err != nil {
		log.Print(err)
	}
}
