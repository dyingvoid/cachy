package domain

import "time"

type Command interface {
	isCommand()
}

type GetCmd struct {
	Key string
}

func (GetCmd) isCommand() {}

type SetCmd struct {
	Key    string
	Value  string
	Expiry *time.Duration
}

func (SetCmd) isCommand() {}
