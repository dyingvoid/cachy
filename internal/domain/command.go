package domain

import "time"

type Command interface {
	isCommand()
}

type commandMarker struct{}

func (commandMarker) isCommand() {}

type PingCmd struct {
	commandMarker
}

type GetCmd struct {
	Key string
	commandMarker
}

type SetCmd struct {
	Key    string
	Value  string
	Expiry *time.Duration
	commandMarker
}

type RPushCmd struct {
	Key    string
	Values []string
	commandMarker
}

type LPushCmd struct {
	Key    string
	Values []string
	commandMarker
}

type LRangeCmd struct {
	Key   string
	Start int
	Stop  int
	commandMarker
}

type LLenCmd struct {
	Key string
	commandMarker
}

type LPopCmd struct {
	Key   string
	Count int
	commandMarker
}

type InfoCmd struct {
	commandMarker
}

type ReplConfCmd struct {
	ListeningPort *int
	Capa          *string
	commandMarker
}
