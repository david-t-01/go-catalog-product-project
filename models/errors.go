package models

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Msg
}
