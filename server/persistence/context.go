package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	helpers "github.com/mmaterowski/raft/helpers"
	structs "github.com/mmaterowski/raft/structs"
)

type SqlLiteDbContext struct {
	Context,
	handle *sql.DB
	uniqueEntryId string
}

func NewSqlLiteDbContext(dbPath string) SqlLiteDbContext {
	sqlContext := SqlLiteDbContext{uniqueEntryId: "ee9fdd8ac5b44fe5866e99bfc9e35932"}

	connected := sqlContext.setDbHandle(dbPath)

	if !connected {
		log.Panic("Error connecting to SQL")
	}
	success := sqlContext.initTablesIfNeeded()

	if !success {
		log.Panic("Db not initialized properly")
	}

	return sqlContext
}

func (db SqlLiteDbContext) PersistValue(key string, value int, termNumber int) (bool, structs.Entry) {
	var entry structs.Entry
	statement, _ := db.handle.Prepare("INSERT INTO Entries (Value, Key, TermNumber) VALUES (?, ?, ?)")
	success, result := executeSafely(statement, value, key, termNumber)

	if !success {
		return false, entry
	}

	if result != nil {
		resultId, _ := result.LastInsertId()
		previousEntryIndex := int(resultId)
		entry = structs.Entry{Index: previousEntryIndex, Value: value, Key: key, TermNumber: termNumber}
	}

	return success, entry
}

func (db SqlLiteDbContext) PersistValues(entries []structs.Entry) (bool, structs.Entry) {
	if len(entries) == 0 {
		return true, structs.Entry{}
	}

	var lastEntry structs.Entry = entries[len(entries)-1]
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
		lastEntry = structs.Entry{Index: index, Value: lastEntry.Value, Key: lastEntry.Key, TermNumber: lastEntry.TermNumber}

	}

	return success, lastEntry
}

func (db SqlLiteDbContext) GetEntryAtIndex(index int) bool {
	selectStatement := fmt.Sprintf(`SELECT * FROM Entries WHERE "Index"=%d`, index)
	statement, _ := db.handle.Prepare(selectStatement)
	success, _ := executeSafely(statement)
	return success
}

func (db SqlLiteDbContext) GetLastEntry() structs.Entry {
	selectStatement := `SELECT* FROM Entries WHERE "Index"=(SELECT MAX("Index") FROM Entries)`
	var entry structs.Entry
	err := db.handle.QueryRow(selectStatement).Scan(&entry.Index, &entry.Value, &entry.Key, &entry.TermNumber)
	helpers.Check(err)
	return entry
}

func (db SqlLiteDbContext) GetCurrentTerm() int {
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

func (db SqlLiteDbContext) SetCurrentTerm(currentTerm int) bool {
	insertStatement := fmt.Sprintf(`UPDATE Term SET CurrentTerm="%d" WHERE UniqueEntryId="%s"`, currentTerm, db.uniqueEntryId)
	statement, _ := db.handle.Prepare(insertStatement)
	success, _ := executeSafely(statement)
	return success
}

func (db SqlLiteDbContext) GetVotedFor() string {
	selectStatement := fmt.Sprintf(`SELECT VotedForId FROM VotedFor WHERE "UniqueEntryId"="%s"`, db.uniqueEntryId)
	foundId := ""
	err := db.handle.QueryRow(selectStatement).Scan(&foundId)
	if err != nil {
		log.Printf("%s", err.Error())
		return foundId
	}
	return foundId
}

func (db SqlLiteDbContext) SetVotedFor(votedForId string) bool {
	insertStatement := fmt.Sprintf(`UPDATE VotedFor SET VotedForId="%s" WHERE UniqueEntryId="%s"`, votedForId, db.uniqueEntryId)
	statement, _ := db.handle.Prepare(insertStatement)
	success, _ := executeSafely(statement)
	return success
}

func (db SqlLiteDbContext) GetLog() []structs.Entry {
	selectStatement := `SELECT * FROM Entries ORDER BY "Index"`
	rows, err := db.handle.Query(selectStatement)
	helpers.Check(err)
	entries := []structs.Entry{}
	for rows.Next() {
		var entry structs.Entry
		err = rows.Scan(&entry.Index, &entry.Value, &entry.Key, &entry.TermNumber)
		helpers.Check(err)
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
