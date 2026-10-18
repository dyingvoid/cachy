package features

import (
	"fmt"
	"strconv"

	core_state "github.com/dyingvoid/cachy/internal/core/state"
)

type PsyncCmd struct {
	ReplicationID string
	Offset        int64
	commandMarker
}

func parsePsyncArgs(args []string) (PsyncCmd, error) {
	if len(args) != 2 {
		return PsyncCmd{}, ErrInvalidArguments
	}

	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return PsyncCmd{}, ErrInvalidArguments
	}

	return PsyncCmd{
		ReplicationID: args[0],
		Offset:        offset,
	}, nil
}

func (h *ProtocolHandler) executePsync(cmd PsyncCmd) ([]Result, error) {
	return []Result{
		StringResult(fmt.Sprintf(
			"FULLRESYNC %s %d",
			core_state.ReplicationID(),
			core_state.ReplicationOffset(),
		)),
		FileResult{Len: 0, Content: []byte{}},
	}, nil
}
