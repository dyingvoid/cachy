package domain_errors

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

	ErrEmptyCommand     = errors.New("empty command")
	ErrUnknownCommand   = errors.New("unknown command")
	ErrInvalidArguments = errors.New("invalid argument")
)
