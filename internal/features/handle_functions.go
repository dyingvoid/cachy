package features

import (
	"fmt"

	"github.com/dyingvoid/cachy/internal/config"
	core_state "github.com/dyingvoid/cachy/internal/core/state"
	"github.com/dyingvoid/cachy/internal/domain"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

func (h *ProtocolHandler) executePing(cmd features_commands.PingCmd) (StringResult, error) {
	return StringResult("PONG"), nil
}

func (h *ProtocolHandler) executeSet(cmd features_commands.SetCmd) (Result, error) {
	r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
	if err := h.storage.Set(cmd.Key, r); err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return StringResult("OK"), nil
}

func (h *ProtocolHandler) executeGet(cmd features_commands.GetCmd) (Result, error) {
	record, err := h.storage.Get(cmd.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return BulkStringResult(record.Value), nil
}

func (h *ProtocolHandler) executeRPush(cmd features_commands.RPushCmd) (Result, error) {
	n, err := h.storage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return IntResult(n), nil
}

func (h *ProtocolHandler) executeLPush(cmd features_commands.LPushCmd) (Result, error) {
	n, err := h.storage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return IntResult(n), nil
}

func (h *ProtocolHandler) executeLRange(cmd features_commands.LRangeCmd) (Result, error) {
	values, err := h.storage.LRange(cmd.Key, cmd.Start, cmd.Stop)
	if err != nil {
		return nil, err
	}

	return ListResult(values), nil
}

func (h *ProtocolHandler) executeLLen(cmd features_commands.LLenCmd) (Result, error) {
	n, err := h.storage.LLen(cmd.Key)
	if err != nil {
		return nil, err
	}

	return IntResult(n), nil
}

func (h *ProtocolHandler) executeLPop(cmd features_commands.LPopCmd) (Result, error) {
	values, err := h.storage.LPop(cmd.Key, cmd.Count)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return ListResult(values), nil
}

func (h *ProtocolHandler) executeInfo(cmd features_commands.InfoCmd) (Result, error) {
	cfg := config.Get()
	info := NewInfo(
		cfg.Role, cfg.ReplicaOf,
		core_state.ReplicationID(),
		core_state.ReplicationOffset(),
	)

	return MapResult(info), nil
}

func (h *ProtocolHandler) executeRepl(
	state *features_conn.State,
	cmd features_commands.ReplConfCmd,
) (Result, error) {
	if cmd.ListeningPort != nil {
		state.ListeningPort = cmd.ListeningPort
	}
	if cmd.Capa != nil {
		state.Capa = cmd.Capa
	}

	return StringResult("OK"), nil
}

func (h *ProtocolHandler) executePsync(
	state *features_conn.State,
	cmd features_commands.PsyncCmd,
) ([]Result, error) {
	state.ReplicationID = cmd.ReplicationID
	state.Offset = cmd.Offset

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
