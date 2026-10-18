package features

import (
	"fmt"
	"strings"
)

type ProtocolParser struct {
}

func NewProtocolParser() *ProtocolParser {
	return &ProtocolParser{}
}

func (p *ProtocolParser) Parse(in string) (Command, error) {
	fields := strings.Fields(in)
	if len(fields) == 0 {
		return nil, ErrEmptyCommand
	}

	name := fields[0]
	args := fields[1:]

	switch strings.ToUpper(name) {
	case "PING":
		return parsePingArgs(args)
	case "GET":
		return parseGetArgs(args)
	case "SET":
		return parseSetArgs(args)
	case "RPUSH":
		return parseRPushArgs(args)
	case "LPUSH":
		return parseLPushArgs(args)
	case "LRANGE":
		return parseLRangeArgs(args)
	case "LLEN":
		return parseLLenArgs(args)
	case "LPOP":
		return parseLPopArgs(args)
	case "INFO":
		return parseInfoArgs(args)
	case "REPLCONF":
		return parseReplConfArgs(args)
	case "PSYNC":
		return parsePsyncArgs(args)
	default:
		return nil, fmt.Errorf("can't parse %s: %w", name, ErrUnknownCommand)
	}
}
