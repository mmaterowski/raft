package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/mmaterowski/raft/consts"
	"github.com/mmaterowski/raft/entry"
	"github.com/mmaterowski/raft/guard"
)

type SqlLiteRepository struct {
	Context,
	handle *sql.DB
	uniqueEntryId string
}

var lock sync.Mutex

var (
	errCantConnect             = errors.New("error connecting to SQL")
	errDbNotInitialized        = errors.New("db not initialized properly")
	errTermNumberLowerThanZero = errors.New("cannot create entry with term number lower than zero")
	errNoRowsUpdated           = errors.New("no rows were updated")
	errCouldNotReconstructLog  = errors.New("couldnt reconstruct log. Probably there's invalid entry in log")
	errInvalidArgument         = errors.New("argument passed to method was invalid")
	errDeleteOutsideOfRange    = errors.New("could not delete entries that are outside of entries length")
	errQueryingRow             = errors.New("something bad happened")
)

func NewSqlLiteRepository(dbPath string) (SqlLiteRepository, error) {
	repo := SqlLiteRepository{uniqueEntryId: "ee9fdd8ac5b44fe5866e99bfc9e35932"}
	connected := repo.setDbHandle(dbPath)

	if !connected {
		return SqlLiteRepository{}, errCantConnect
	}

	success := repo.InitTablesIfNeeded()

	if !success {
		return SqlLiteRepository{}, errDbNotInitialized
	}

	return repo, nil
}

func (s SqlLiteRepository) PersistValue(ctx context.Context, key string, value int, termNumber int) (*entry.Entry, error) {
	lock.Lock()
	defer lock.Unlock()
	insertStatement, _ := s.handle.Prepare("INSERT INTO Entries (Value, Key, TermNumber) VALUES (?, ?, ?)")
	defer insertStatement.Close()
	insertResult, err := insertStatement.Exec(value, key, termNumber)

	if err != nil {
		return &entry.Entry{}, err
	}

	lastInsertedId, err := insertResult.LastInsertId()
	entry, createErr := entry.New(int(lastInsertedId), value, termNumber, key)
	return entry, createErr
}

func (s SqlLiteRepository) PersistValues(ctx context.Context, entries []entry.Entry) (*entry.Entry, error) {
	lock.Lock()
	defer lock.Unlock()
	if len(entries) == 0 {
		return &entry.Entry{}, nil
	}

	insert := "INSERT INTO Entries (Value, Key, TermNumber) VALUES "
	for _, entry := range entries {
		insert += fmt.Sprintf(`(%d,"%s",%d),`, entry.Value, entry.Key, entry.TermNumber)
	}
	insert = strings.TrimSuffix(insert, ",")
	statement, _ := s.handle.Prepare(insert)
	insertResult, err := statement.Exec()

	if err != nil {
		return &entry.Entry{}, err
	}

	lastEntryId, _ := insertResult.LastInsertId()
	lastEntry := entries[len(entries)-1]
	return entry.New(int(lastEntryId), lastEntry.Value, lastEntry.TermNumber, lastEntry.Key)
}

func (db SqlLiteRepository) GetEntryAtIndex(ctx context.Context, index int) (*entry.Entry, error) {
	if guard.AgainstNegativeValue(index) {
		return &entry.Entry{}, errInvalidArgument
	}

	selectStatement := fmt.Sprintf(`SELECT * FROM Entries WHERE "Index"=%d`, index)
	var createdIndex, value, term int
	var key string
	err := db.handle.QueryRow(selectStatement).Scan(&createdIndex, &value, &key, &term)
	if err != nil {
		return &entry.Entry{}, err
	}
	return entry.New(createdIndex, value, term, key)
}

func (db SqlLiteRepository) GetLastEntry(ctx context.Context) (*entry.Entry, error) {
	selectStatement := `SELECT* FROM Entries WHERE "Index"=(SELECT MAX("Index") FROM Entries)`
	var createdIndex, value, term int
	var key string
	err := db.handle.QueryRow(selectStatement).Scan(&createdIndex, &value, &key, &term)
	if err != nil {
		return &entry.Entry{}, err
	}
	return entry.New(createdIndex, value, term, key)
}

func (db SqlLiteRepository) GetCurrentTerm(ctx context.Context) (int, error) {
	selectStatement := fmt.Sprintf(`SELECT CurrentTerm FROM Term WHERE "UniqueEntryId"="%s"`, db.uniqueEntryId)
	currentTermNumber := consts.TermUninitializedValue
	sqlRes := ""

	err := db.handle.QueryRow(selectStatement).Scan(&sqlRes)
	if err == sql.ErrNoRows {
		return currentTermNumber, err
	}
	if err != nil {
		log.Print(errQueryingRow, err)
	}
	currentTermNumber, convError := strconv.Atoi(sqlRes)
	if convError != nil {
		return currentTermNumber, convError
	}

	return currentTermNumber, nil
}

func (db SqlLiteRepository) SetCurrentTerm(ctx context.Context, currentTerm int) error {
	if guard.AgainstNegativeValue(currentTerm) {
		return errInvalidArgument
	}

	insertStatement := fmt.Sprintf(`UPDATE Term SET CurrentTerm="%d" WHERE UniqueEntryId="%s"`, currentTerm, db.uniqueEntryId)
	statement, _ := db.handle.Prepare(insertStatement)
	sqlRes, err := statement.Exec()
	if err != nil {
		return err
	}

	if atLeastOneRowWasUpdated(sqlRes) {
		return errNoRowsUpdated
	}

	return nil
}

func (db SqlLiteRepository) GetVotedFor(ctx context.Context) (string, error) {
	selectStatement := fmt.Sprintf(`SELECT VotedForId FROM VotedFor WHERE "UniqueEntryId"="%s"`, db.uniqueEntryId)
	foundId := ""
	err := db.handle.QueryRow(selectStatement).Scan(&foundId)
	if err == sql.ErrNoRows {
		return foundId, err
	} else if err != nil {
		log.Print(errQueryingRow, err)
	}

	return foundId, nil
}

func (db SqlLiteRepository) SetVotedFor(ctx context.Context, votedForId string) error {
	insertStatement := fmt.Sprintf(`UPDATE VotedFor SET VotedForId="%s" WHERE UniqueEntryId="%s"`, votedForId, db.uniqueEntryId)
	statement, _ := db.handle.Prepare(insertStatement)
	sqlRes, err := statement.Exec()
	if err != nil {
		return err
	}

	if atLeastOneRowWasUpdated(sqlRes) {
		return errNoRowsUpdated
	}

	return nil
}

func (db SqlLiteRepository) GetLog(ctx context.Context) (*[]entry.Entry, error) {
	selectStatement := `SELECT * FROM Entries ORDER BY "Index"`
	rows, err := db.handle.Query(selectStatement)
	if err != nil {
		return &[]entry.Entry{}, nil
	}

	entries := []entry.Entry{}
	couldntReconstructLog := false
	for rows.Next() {
		var index, value, term int
		var key string
		err = rows.Scan(&index, &value, &key, &term)
		if err != nil {
			log.Print("Error while scanning log entries:", err)
			couldntReconstructLog = true
			break
		}
		entry, createError := entry.New(index, value, term, key)
		if createError != nil {
			couldntReconstructLog = true
			break
		}
		entries = append(entries, *entry)
	}

	if couldntReconstructLog {
		return &[]entry.Entry{}, errCouldNotReconstructLog
	}

	return &entries, nil
}

func (db SqlLiteRepository) DeleteAllEntriesStartingFrom(ctx context.Context, index int) error {
	if guard.AgainstZeroOrNegativeValue(index) {
		return errInvalidArgument
	}

	lock.Lock()
	defer lock.Unlock()

	tx, err := db.handle.Begin()
	if err != nil {
		return err
	}
	delete := fmt.Sprintf(`DELETE FROM Entries Where "Index" >= %d`, index)
	statement, _ := tx.Prepare(delete)
	_, deleteErr := statement.Exec()
	if deleteErr != nil {
		return deleteErr
	}
	autoincrementResetStatement := `UPDATE "sqlite_sequence"
	SET "seq" = (SELECT COUNT("Index")-1 FROM "Entries")-1
	WHERE name = "Entries"`
	statement, _ = tx.Prepare(autoincrementResetStatement)
	_, incErr := statement.Exec()
	tx.Commit()
	return incErr
}

func atLeastOneRowWasUpdated(result sql.Result) bool {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Print(err)
		return false
	}
	if int(rowsAffected) <= 0 {
		return false
	}
	return true
}
