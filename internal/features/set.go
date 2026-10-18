package features

import (
	"strconv"
	"strings"
	"time"

	"github.com/dyingvoid/cachy/internal/domain"
)

type SetCmd struct {
	Key    string
	Value  string
	Expiry *time.Duration
	commandMarker
}

func parseSetArgs(args []string) (SetCmd, error) {
	l := len(args)
	if l < 2 || l > 4 || l%2 != 0 {
		return SetCmd{}, ErrInvalidArguments
	}

	cmd := SetCmd{
		Key:   args[0],
		Value: args[1],
	}

	if l == 4 {
		n, err := strconv.Atoi(args[3])
		if err != nil || n <= 0 {
			return SetCmd{}, ErrInvalidArguments
		}
		if strings.ToUpper(args[2]) == "EX" {
			d := time.Duration(n) * time.Second
			cmd.Expiry = &d
		} else if strings.ToUpper(args[2]) == "PX" {
			d := time.Duration(n) * time.Millisecond
			cmd.Expiry = &d
		} else {
			return SetCmd{}, ErrInvalidArguments
		}
	}

	return cmd, nil
}

func (h *ProtocolHandler) executeSet(cmd SetCmd) (Result, error) {
	r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
	if err := h.storage.Set(cmd.Key, r); err != nil {
		return nil, err
	}

	return StringResult("OK"), nil
}
