package features

import (
	"fmt"

	core_state "github.com/dyingvoid/cachy/internal/core/state"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

func (h *ProtocolHandler) executePsync(
	state *features_conn.State,
	cmd features_commands.PsyncCmd,
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
