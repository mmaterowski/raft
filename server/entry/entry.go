package entry

import (
	"errors"

	"github.com/mmaterowski/raft/guard"
)

type Entry struct {
	Index      int
	Value      int
	Key        string
	TermNumber int
}

var (
	ErrIndexLowerThanZero      = errors.New("cannot create entry with index lower than zero")
	ErrKeyEmpty                = errors.New("cannot create entry without a key")
	ErrTermNumberLowerThanZero = errors.New("cannot create entry with term number lower than zero")
)

func New(index, value, termNumber int, key string) (*Entry, error) {
	if guard.AgainstNegativeValue(termNumber) {
		return &Entry{}, ErrIndexLowerThanZero
	}

	if guard.AgainstEmptyString(key) {
		return &Entry{}, ErrKeyEmpty
	}

	if guard.AgainstNegativeValue(termNumber) {
		return &Entry{}, ErrTermNumberLowerThanZero
	}

	return &Entry{Index: index, Value: value, TermNumber: termNumber, Key: key}, nil
}

func (e Entry) IsEmpty() bool {
	return e == Entry{}
}
