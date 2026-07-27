package store

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrCannotDelete = errors.New("cannot delete")
	ErrCannotUpdate = errors.New("cannot update")
	ErrInvalidInput = errors.New("invalid input")
)
