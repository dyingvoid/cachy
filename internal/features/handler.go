package features

import (
	"context"
	"fmt"

	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
	features_replication "github.com/dyingvoid/cachy/internal/features/replication"
	"github.com/dyingvoid/cachy/internal/features/result"
)

type ProtocolHandler struct {
	storage     Storage
	replication *features_replication.ReplicationHandler
}

func NewProtocolHandler(
	storage Storage,
	replication *features_replication.ReplicationHandler,
) *ProtocolHandler {
	return &ProtocolHandler{
		storage:     storage,
		replication: replication,
	}
}

func (h *ProtocolHandler) Handle(
	ctx context.Context,
	state *features_conn.State,
	cmd features_commands.Command,
) ([]result.Result, error) {
	switch cmd := cmd.(type) {
	case features_commands.PingCmd:
		return singleResult(h.executePing(cmd))
	case features_commands.SetCmd:
		return singleResult(h.executeSet(cmd))
	case features_commands.GetCmd:
		return singleResult(h.executeGet(cmd))
	case features_commands.RPushCmd:
		return singleResult(h.executeRPush(cmd))
	case features_commands.LPushCmd:
		return singleResult(h.executeLPush(cmd))
	case features_commands.LRangeCmd:
		return singleResult(h.executeLRange(cmd))
	case features_commands.LLenCmd:
		return singleResult(h.executeLLen(cmd))
	case features_commands.LPopCmd:
		return singleResult(h.executeLPop(cmd))
	case features_commands.InfoCmd:
		return singleResult(h.executeInfo(cmd))
	case features_commands.ReplConfCmd:
		return singleResult(h.executeRepl(state, cmd))
	case features_commands.PsyncCmd:
		return h.executePsync(state, cmd)
	default:
		return nil, fmt.Errorf("can't execute command: %w", features_commands.ErrUnknownCommand)
	}
}

func (h *ProtocolHandler) replicate(cmd features_commands.WriteCommand) {
	if h.replication == nil {
		return
	}

	h.replication.Publish(cmd)
}

func singleResult(res result.Result, err error) ([]result.Result, error) {
	if err != nil {
		return nil, err
	}

	return []result.Result{res}, nil
}
