package core_parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(in string) (domain.Command, error) {
	fields := strings.Fields(in)
	if len(fields) == 0 {
		return nil, domain_errors.ErrEmptyCommand
	}

	name := fields[0]
	args := fields[1:]

	switch strings.ToUpper(name) {
	case "GET":
		return parseGetArgs(args)
	case "SET":
		return parseSetArgs(args)
	case "RPUSH":
		return parseRPushArgs(args)
	case "LRANGE":
		return parseLRangeArgs(args)
	default:
		return nil, fmt.Errorf("can't parse %s: %w", name, domain_errors.ErrUnknownCommand)
	}
}

func parseGetArgs(args []string) (domain.GetCmd, error) {
	if len(args) != 1 {
		return domain.GetCmd{}, domain_errors.ErrInvalidArguments
	}

	return domain.GetCmd{Key: args[0]}, nil
}

func parseSetArgs(args []string) (domain.SetCmd, error) {
	l := len(args)
	if l < 2 || l > 4 || l%2 != 0 {
		return domain.SetCmd{}, domain_errors.ErrInvalidArguments
	}

	cmd := domain.SetCmd{
		Key:   args[0],
		Value: args[1],
	}

	if l == 4 {
		n, err := strconv.Atoi(args[3])
		if err != nil || n <= 0 {
			return domain.SetCmd{}, domain_errors.ErrInvalidArguments
		}
		if strings.ToUpper(args[2]) == "EX" {
			d := time.Duration(n) * time.Second
			cmd.Expiry = &d
		} else if strings.ToUpper(args[2]) == "PX" {
			d := time.Duration(n) * time.Millisecond
			cmd.Expiry = &d
		} else {
			return domain.SetCmd{}, domain_errors.ErrInvalidArguments
		}
	}

	return cmd, nil
}

func parseRPushArgs(args []string) (domain.RPushCmd, error) {
	if len(args) < 2 {
		return domain.RPushCmd{}, domain_errors.ErrInvalidArguments
	}

	return domain.RPushCmd{Key: args[0], Values: args[1:]}, nil
}

func parseLRangeArgs(args []string) (domain.LRangeCmd, error) {
	if len(args) != 3 {
		return domain.LRangeCmd{}, domain_errors.ErrInvalidArguments
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return domain.LRangeCmd{}, domain_errors.ErrInvalidArguments
	}

	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return domain.LRangeCmd{}, domain_errors.ErrInvalidArguments
	}

	return domain.LRangeCmd{Key: args[0], Start: start, Stop: stop}, nil
}
