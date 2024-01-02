package handler

import "errors"

var (
	ErrRecordNotFound = errors.New("no record found")
	ErrDuplicate      = errors.New("duplicate key value")
	ErrConflict       = errors.New("data conflict")
	ErrDatabase       = errors.New("database error")
)
