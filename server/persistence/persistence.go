package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	. "github.com/mmaterowski/raft/helpers"
	. "github.com/mmaterowski/raft/structs"
)

type SqlLiteDb struct {
	handle        *sql.DB
	uniqueEntryId string
}

func NewDb(debug bool) SqlLiteDb {
	db := SqlLiteDb{}
	db.uniqueEntryId = "ee9fdd8ac5b44fe5866e99bfc9e35932"
	success := db.setup(debug)
	if !success {
		log.Panic("Db not initialized properly")
	}
	return db
}

func (s *SqlLiteDb) connectToSql(debug bool) bool {
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
	s.handle = database
	return database != nil
}

func (s *SqlLiteDb) setup(debug bool) bool {
	connected := s.connectToSql(debug)
	if !connected {
		log.Print("Error connecting to SQL")
		return false
	}
	tablesInitialized := s.initTablesIfNeeded()
	return tablesInitialized

}

func (s SqlLiteDb) initTablesIfNeeded() bool {
	entriesCreated := s.createEntriesTableIfNotExists()
	termCreated := s.createTermTableIfNotExists()
	votedForCreated := s.createVotedForIfNotExists()
	return entriesCreated && termCreated && votedForCreated
}

func (s SqlLiteDb) createEntriesTableIfNotExists() bool {
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

func (s SqlLiteDb) createTermTableIfNotExists() bool {
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

func (s SqlLiteDb) createVotedForIfNotExists() bool {
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

func (db SqlLiteDb) PersistValue(key string, value int, termNumber int) (bool, Entry) {
	var entry Entry
	statement, _ := db.handle.Prepare("INSERT INTO Entries (Value, Key, TermNumber) VALUES (?, ?, ?)")
	success, result := executeSafely(statement, value, key, termNumber)

	if !success {
		return false, entry
	}

	if result != nil {
		resultId, _ := result.LastInsertId()
		previousEntryIndex := int(resultId)
		entry = Entry{Index: previousEntryIndex, Value: value, Key: key, TermNumber: termNumber}
	}

	return success, entry
}

func (db SqlLiteDb) PersistValues(entries []Entry) (bool, Entry) {
	if len(entries) == 0 {
		return true, Entry{}
	}

	var lastEntry Entry = entries[len(entries)-1]
	insert := "INSERT INTO Entries (Value, Key, TermNumber) VALUES "
	for _, entry := range entries {
		insert += fmt.Sprintf(`(%d,"%s",%d),`, entry.Value, entry.Key, entry.TermNumber)
	}
	insert = strings.TrimSuffix(insert, ",")
	statement, _ := db.handle.Prepare(insert)

	success, result := executeSafely(statement)

	if !success {
		return false, lastEntry
	}

	if result != nil {
		resultId, _ := result.LastInsertId()
		index := int(resultId)
		lastEntry = Entry{Index: index, Value: lastEntry.Value, Key: lastEntry.Key, TermNumber: lastEntry.TermNumber}

	}

	return success, lastEntry
}

func (db SqlLiteDb) GetEntryAtIndex(index int) bool {
	selectStatement := fmt.Sprintf(`SELECT * FROM Entries WHERE "Index"=%d`, index)
	statement, _ := db.handle.Prepare(selectStatement)
	success, _ := executeSafely(statement)
	return success
}

func (db SqlLiteDb) GetLastEntry() Entry {
	selectStatement := `SELECT* FROM Entries WHERE "Index"=(SELECT MAX("Index") FROM Entries)`
	var entry Entry
	err := db.handle.QueryRow(selectStatement).Scan(&entry.Index, &entry.Value, &entry.Key, &entry.TermNumber)
	Check(err)
	return entry
}

func (db SqlLiteDb) GetCurrentTerm() int {
	selectStatement := fmt.Sprintf(`SELECT CurrentTerm FROM Term WHERE "UniqueEntryId"="%s"`, db.uniqueEntryId)
	currentTermNumber := 0
	sqlRes := ""

	err := db.handle.QueryRow(selectStatement).Scan(&sqlRes)
	if err != nil {
		log.Printf("%s", err.Error())
		return currentTermNumber
	}

	currentTermNumber, convError := strconv.Atoi(sqlRes)
	if convError != nil {
		log.Print(convError)
		return currentTermNumber
	}

	return currentTermNumber
}

func (db SqlLiteDb) SetCurrentTerm(currentTerm int) bool {
	insertStatement := fmt.Sprintf(`UPDATE Term SET CurrentTerm="%d" WHERE UniqueEntryId="%s"`, currentTerm, db.uniqueEntryId)
	statement, _ := db.handle.Prepare(insertStatement)
	success, _ := executeSafely(statement)
	return success
}

func (db SqlLiteDb) GetVotedFor() string {
	selectStatement := fmt.Sprintf(`SELECT VotedForId FROM VotedFor WHERE "UniqueEntryId"="%s"`, db.uniqueEntryId)
	foundId := ""
	err := db.handle.QueryRow(selectStatement).Scan(&foundId)
	if err != nil {
		log.Printf("%s", err.Error())
		return foundId
	}
	return foundId
}

func (db SqlLiteDb) SetVotedFor(votedForId string) bool {
	insertStatement := fmt.Sprintf(`UPDATE VotedFor SET VotedForId="%s" WHERE UniqueEntryId="%s"`, votedForId, db.uniqueEntryId)
	statement, _ := db.handle.Prepare(insertStatement)
	success, _ := executeSafely(statement)
	return success
}

func (db SqlLiteDb) GetLog() []Entry {
	selectStatement := `SELECT * FROM Entries ORDER BY "Index"`
	rows, err := db.handle.Query(selectStatement)
	Check(err)
	entries := []Entry{}
	for rows.Next() {
		var entry Entry
		err = rows.Scan(&entry.Index, &entry.Value, &entry.Key, &entry.TermNumber)
		Check(err)
		entries = append(entries, entry)
	}
	return entries
}

func executeSafely(sqlStatement *sql.Stmt, args ...interface{}) (bool, sql.Result) {
	if args != nil {
		result, errWithArgs := sqlStatement.Exec(args...)
		if errWithArgs != nil {
			log.Print(errWithArgs)
			return false, result

		}
		return true, result
	}
	result, err := sqlStatement.Exec()
	if err != nil {
		log.Print(err)
		return false, result
	}
	return true, result
}
