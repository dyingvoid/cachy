package domain_errors

import "errors"

var (
	ErrNotFound = errors.New("not found")

	// Parser errors.
	ErrEmptyCommand     = errors.New("empty command")
	ErrUnknownCommand   = errors.New("unknown command")
	ErrInvalidArguments = errors.New("invalid argument")
)
