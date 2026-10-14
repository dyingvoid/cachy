package domain

import "time"

type Command interface {
	isCommand()
}

type GetCmd struct {
	Key string
}

type SetCmd struct {
	Key    string
	Value  string
	Expiry *time.Duration
}

type RPushCmd struct {
	Key    string
	Values []string
}

func (SetCmd) isCommand() {}

func (GetCmd) isCommand() {}

func (RPushCmd) isCommand() {}
