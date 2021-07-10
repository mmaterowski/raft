package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

var db *sql.DB
var uniqueEntryId = "ee9fdd8ac5b44fe5866e99bfc9e35932"

func PersistValue(key string, value int, termNumber int) (bool, Entry) {
	var entry Entry
	statement, _ := db.Prepare("INSERT INTO Entries (Value, Key, TermNumber) VALUES (?, ?, ?)")
	success, result := executeSafely(statement, value, key, termNumber)

	if !success {
		return false, entry
	}

	if result != nil {
		resultId, _ := result.LastInsertId()
		previousEntryIndex = int(resultId)
	}

	entry = Entry{Index: previousEntryIndex, Value: value, Key: key, TermNumber: termNumber}
	return success, entry
}

func PersistValues(entries []Entry) (bool, Entry) {
	if len(entries) == 0 {
		return true, Entry{}
	}

	var lastEntry Entry = entries[len(entries)-1]
	insert := "INSERT INTO Entries (Value, Key, TermNumber) VALUES "
	for _, entry := range entries {
		insert += fmt.Sprintf(`(%d,"%s",%d),`, entry.Value, entry.Key, entry.TermNumber)
	}
	insert = strings.TrimSuffix(insert, ",")
	statement, _ := db.Prepare(insert)

	success, result := executeSafely(statement)

	if !success {
		return false, lastEntry
	}

	if result != nil {
		resultId, _ := result.LastInsertId()
		previousEntryIndex = int(resultId)
	}

	lastEntry = Entry{Index: previousEntryIndex, Value: lastEntry.Value, Key: lastEntry.Key, TermNumber: lastEntry.TermNumber}
	return success, lastEntry
}

func GetEntryAtIndex(index int) bool {
	selectStatement := fmt.Sprintf(`SELECT * FROM Entries WHERE "Index"=%d`, index)
	statement, _ := db.Prepare(selectStatement)
	success, _ := executeSafely(statement)
	return success
}

func GetLastEntry() Entry {
	selectStatement := `SELECT* FROM Entries WHERE "Index"=(SELECT MAX("Index") FROM Entries)`
	var entry Entry
	err := db.QueryRow(selectStatement).Scan(&entry.Index, &entry.Value, &entry.Key, &entry.TermNumber)
	Check(err)
	return entry
}

func GetCurrentTerm() int {
	selectStatement := fmt.Sprintf(`SELECT CurrentTerm FROM Term WHERE "UniqueEntryId"="%s"`, uniqueEntryId)
	currentTermNumber := 0
	sqlRes := ""

	err := db.QueryRow(selectStatement).Scan(&sqlRes)
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

func SetCurrentTerm(currentTerm int) bool {
	insertStatement := fmt.Sprintf(`UPDATE Term SET CurrentTerm="%d" WHERE UniqueEntryId="%s"`, currentTerm, uniqueEntryId)
	statement, _ := db.Prepare(insertStatement)
	success, _ := executeSafely(statement)
	return success
}

func GetVotedFor() string {
	selectStatement := fmt.Sprintf(`SELECT VotedForId FROM VotedFor WHERE "UniqueEntryId"="%s"`, uniqueEntryId)
	foundId := ""
	err := db.QueryRow(selectStatement).Scan(&foundId)
	if err != nil {
		log.Printf("%s", err.Error())
		return foundId
	}
	return foundId
}

func SetVotedFor(votedForId string) bool {
	insertStatement := fmt.Sprintf(`UPDATE VotedFor SET VotedForId="%s" WHERE UniqueEntryId="%s"`, votedForId, uniqueEntryId)
	statement, _ := db.Prepare(insertStatement)
	success, _ := executeSafely(statement)
	return success
}

func GetLog() []Entry {
	selectStatement := `SELECT * FROM Entries ORDER BY "Index"`
	rows, err := db.Query(selectStatement)
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
