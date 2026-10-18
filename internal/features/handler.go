package features

import (
	"context"
	"fmt"

	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

type ProtocolHandler struct {
	parser      Parser
	storage     Storage
	replication *ReplicationHandler
}

func NewProtocolHandler(
	parser Parser,
	storage Storage,
	replication *ReplicationHandler,
) *ProtocolHandler {
	return &ProtocolHandler{
		parser:      parser,
		storage:     storage,
		replication: replication,
	}
}

func (h *ProtocolHandler) Handle(
	ctx context.Context,
	state *features_conn.State,
	request []byte,
) ([]Result, error) {
	cmd, err := h.parser.Parse(string(request))
	if err != nil {
		return nil, err
	}

	return h.execute(ctx, state, cmd)
}

func (h *ProtocolHandler) execute(
	ctx context.Context,
	state *features_conn.State,
	cmd Command,
) ([]Result, error) {
	switch cmd := cmd.(type) {
	case PingCmd:
		return singleResult(h.executePing(cmd))
	case SetCmd:
		return singleResult(h.executeSet(cmd))
	case GetCmd:
		return singleResult(h.executeGet(cmd))
	case RPushCmd:
		return singleResult(h.executeRPush(cmd))
	case LPushCmd:
		return singleResult(h.executeLPush(cmd))
	case LRangeCmd:
		return singleResult(h.executeLRange(cmd))
	case LLenCmd:
		return singleResult(h.executeLLen(cmd))
	case LPopCmd:
		return singleResult(h.executeLPop(cmd))
	case InfoCmd:
		return singleResult(h.executeInfo(cmd))
	case ReplConfCmd:
		return singleResult(h.executeRepl(state, cmd))
	case PsyncCmd:
		return h.executePsync(state, cmd)
	default:
		return nil, fmt.Errorf("can't execute command: %w", ErrUnknownCommand)
	}
}

func (h *ProtocolHandler) replicate(cmd ReplicationCommand) {
	if h.replication == nil {
		return
	}

	h.replication.Publish(cmd)
}

func singleResult(result Result, err error) ([]Result, error) {
	if err != nil {
		return nil, err
	}

	return []Result{result}, nil
}
