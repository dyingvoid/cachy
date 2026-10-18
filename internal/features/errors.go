package features

import "errors"

var (
	ErrEmptyCommand               = errors.New("empty command")
	ErrUnknownCommand             = errors.New("unknown command")
	ErrInvalidArguments           = errors.New("invalid argument")
	ErrReplicationStateIncomplete = errors.New("replication state incomplete")
)
