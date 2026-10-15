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

type LPushCmd struct {
	Key    string
	Values []string
}

type LRangeCmd struct {
	Key   string
	Start int
	Stop  int
}

type LLenCmd struct {
	Key string
}

type LPopCmd struct {
	Key   string
	Count int
}

func (SetCmd) isCommand() {}

func (GetCmd) isCommand() {}

func (RPushCmd) isCommand() {}

func (LPushCmd) isCommand() {}

func (LRangeCmd) isCommand() {}

func (LLenCmd) isCommand() {}

func (LPopCmd) isCommand() {}
