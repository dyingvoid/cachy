package domain

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)
