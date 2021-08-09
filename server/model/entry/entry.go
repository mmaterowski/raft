package entry

import (
	"errors"

	"github.com/mmaterowski/raft/utils/guard"
)

type Entry struct {
	Index      int    `json:"index"`
	Value      int    `json:"value"`
	Key        string `json:"key"`
	TermNumber int    `json:"TermNumber"`
}

var (
	errIndexLowerThanZero      = errors.New("cannot create entry with index lower than zero")
	errKeyEmpty                = errors.New("cannot create entry without a key")
	errTermNumberLowerThanZero = errors.New("cannot create entry with term number lower than zero")
)

func New(index, value, termNumber int, key string) (*Entry, error) {
	if guard.AgainstNegativeValue(index) {
		return &Entry{}, errIndexLowerThanZero
	}

	if guard.AgainstEmptyString(key) {
		return &Entry{}, errKeyEmpty
	}

	if guard.AgainstNegativeValue(termNumber) {
		return &Entry{}, errTermNumberLowerThanZero
	}

	return &Entry{Index: index, Value: value, TermNumber: termNumber, Key: key}, nil
}

func (e Entry) IsEmpty() bool {
	return e == Entry{}
}
