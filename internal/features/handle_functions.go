package features

import (
	"fmt"

	"github.com/dyingvoid/cachy/internal/config"
	"github.com/dyingvoid/cachy/internal/domain"
	"github.com/dyingvoid/cachy/internal/features/commands"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
	features_replication "github.com/dyingvoid/cachy/internal/features/replication"
	"github.com/dyingvoid/cachy/internal/features/result"
	core_state "github.com/dyingvoid/cachy/internal/features/state"
)

func (h *ProtocolHandler) executePing(cmd features_commands.PingCmd) (result.StringResult, error) {
	return result.StringResult("PONG"), nil
}

func (h *ProtocolHandler) executeSet(cmd features_commands.SetCmd) (result.StringResult, error) {
	r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
	if err := h.storage.Set(cmd.Key, r); err != nil {
		return "", err
	}

	h.replicate(cmd)

	return result.StringResult("OK"), nil
}

func (h *ProtocolHandler) executeGet(cmd features_commands.GetCmd) (result.StringResult, error) {
	record, err := h.storage.Get(cmd.Key)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}

	return result.StringResult(record.Value), nil
}

func (h *ProtocolHandler) executeRPush(cmd features_commands.RPushCmd) (result.IntResult, error) {
	n, err := h.storage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return result.IntResult(0), err
	}

	h.replicate(cmd)

	return result.IntResult(n), nil
}

func (h *ProtocolHandler) executeLPush(cmd features_commands.LPushCmd) (result.IntResult, error) {
	n, err := h.storage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return result.IntResult(0), err
	}

	h.replicate(cmd)

	return result.IntResult(n), nil
}

func (h *ProtocolHandler) executeLRange(cmd features_commands.LRangeCmd) (result.ListResult, error) {
	values, err := h.storage.LRange(cmd.Key, cmd.Start, cmd.Stop)
	if err != nil {
		return nil, err
	}

	return result.ListResult(values), nil
}

func (h *ProtocolHandler) executeLLen(cmd features_commands.LLenCmd) (result.Result, error) {
	n, err := h.storage.LLen(cmd.Key)
	if err != nil {
		return nil, err
	}

	return result.IntResult(n), nil
}

func (h *ProtocolHandler) executeLPop(cmd features_commands.LPopCmd) (result.ListResult, error) {
	values, err := h.storage.LPop(cmd.Key, cmd.Count)
	if err != nil {
		return nil, err
	}

	h.replicate(cmd)

	return result.ListResult(values), nil
}

func (h *ProtocolHandler) executeInfo(cmd features_commands.InfoCmd) (result.MapResult, error) {
	cfg := config.Get()
	str := "todo"
	// TODO: move info elsewhere
	info := result.NewInfo(
		cfg.Role, &str,
		core_state.ReplicationID(),
		core_state.ReplicationOffset(),
	)

	return result.MapResult(info), nil
}

func (h *ProtocolHandler) executeRepl(
	state *features_conn.State,
	cmd features_commands.ReplConfCmd,
) (result.StringResult, error) {
	if cmd.ListeningPort != nil {
		state.ListeningPort = cmd.ListeningPort
	}
	if cmd.Capa != nil {
		state.Capa = cmd.Capa
	}

	return result.StringResult("OK"), nil
}

func (h *ProtocolHandler) executePsync(
	state *features_conn.State,
	cmd features_commands.PsyncCmd,
) ([]result.Result, error) {
	state.ReplicationID = cmd.ReplicationID
	state.Offset = cmd.Offset

	if !state.TryStartReplication() {
		return nil, features_replication.ErrReplicationStateIncomplete
	}

	return []result.Result{
		result.StringResult(fmt.Sprintf(
			"FULLRESYNC %s %d",
			core_state.ReplicationID(),
			core_state.ReplicationOffset(),
		)),
		result.FileResult{Len: 0, Content: []byte{}},
		result.StreamingResult[commands.WriteCommand]{Stream: h.replication.Subscribe()},
	}, nil
}
