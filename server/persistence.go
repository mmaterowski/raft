package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
)

var db *sql.DB
var uniqueEntryId = "ee9fdd8ac5b44fe5866e99bfc9e35932"

func SetupDB() bool {
	connected := connectToSql()
	if !connected {
		log.Print("Error connecting to SQL")
		return false
	}
	tablesInitialized := initTablesIfNeeded()
	return tablesInitialized
}

func putValue(key string, value int, termNumber int) bool {
	statement, _ := db.Prepare("INSERT INTO Entries (Value, Key, TermNumber) VALUES (?, ?, ?)")
	success := executeSafely(statement, key, value, termNumber)
	return success
}

func GetEntryAtIndex(index int) bool {
	selectStatement := fmt.Sprintf(`SELECT * FROM Entries WHERE "Index"=%d`, index)
	statement, _ := db.Prepare(selectStatement)
	success := executeSafely(statement)
	return success
}

func GetCurrentTerm() int {
	selectStatement := fmt.Sprintf(`SELECT CurrentTerm FROM Term WHERE "UniqueEntryId"="%s"`, uniqueEntryId)
	log.Print(selectStatement)
	currentTermNumber := 0
	sqlRes := ""
	err := db.QueryRow(selectStatement).Scan(&sqlRes)
	//TODO Check for null
	currentTermNumber, _ = strconv.Atoi(sqlRes)
	log.Print(currentTermNumber)
	if err != nil {
		log.Printf("%s", err.Error())
		return currentTermNumber
	}
	return currentTermNumber
}

func GetVotedFor() string {
	selectStatement := fmt.Sprintf(`SELECT VotedForId FROM VotedFor WHERE "UniqueEntryId"="%s"`, uniqueEntryId)
	log.Print(selectStatement)
	foundId := ""
	err := db.QueryRow(selectStatement).Scan(&foundId)
	if err != nil {
		log.Printf("%s", err.Error())
		return foundId
	}
	return foundId
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
	createStatement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS "Term" (
		"CurrentTerm"	INTEGER
	, "UniqueEntryId"	TEXT)`)
	if err != nil {
		log.Print(err)
		return false
	}
	success := executeSafely(createStatement)
	return success

}

func createVotedForIfNotExists() bool {
	createStatement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS "VotedFor" (
		"VotedForId"	TEXT,
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
