package features

import (
	"fmt"
	"strconv"

	core_state "github.com/dyingvoid/cachy/internal/core/state"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
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

func (h *ProtocolHandler) executePsync(
	state *features_conn.State,
	cmd PsyncCmd,
) ([]Result, error) {
	state.Replication.ReplicationID = cmd.ReplicationID
	state.Replication.Offset = cmd.Offset

	if !state.TryStartReplication() {
		return nil, ErrReplicationStateIncomplete
	}

	return []Result{
		StringResult(fmt.Sprintf(
			"FULLRESYNC %s %d",
			core_state.ReplicationID(),
			core_state.ReplicationOffset(),
		)),
		FileResult{Len: 0, Content: []byte{}},
		StreamingResult{Stream: h.replication.Subscribe()},
	}, nil
}
