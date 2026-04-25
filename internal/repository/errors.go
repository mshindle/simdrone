package repository

import (
	"errors"
	"fmt"
)

const (
	dataNotFound = iota
	customError
)

var (
	ErrNotFound = &DBError{code: dataNotFound, Message: "no data found"}
)

// NewDBError creates new instance of DBError
func NewDBError(message string) *DBError {
	return &DBError{
		code:    customError,
		Message: message,
	}
}

// DBError represents an error that occurred while handling a request.
type DBError struct {
	code    int
	Message string `json:"message"`
	err     error
}

// Error makes it compatible with `error` interface.
func (de *DBError) Error() string {
	if de.err == nil {
		return fmt.Sprintf("code=%d, message=%v", de.code, de.Message)
	}
	return fmt.Sprintf("code=%d, message=%v, err=%v", de.code, de.Message, de.err.Error())
}

// Wrap returns a new DBError with given errors wrapped inside
func (de *DBError) Wrap(err error) error {
	return &DBError{
		code:    de.code,
		Message: de.Message,
		err:     err,
	}
}

func (de *DBError) Unwrap() error {
	return de.err
}

func (de *DBError) Is(target error) bool {
	var t *DBError
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	if de.code == t.code {
		return true
	}
	return de.Message == t.Message
}
