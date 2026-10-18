package features

import (
	"context"
	"fmt"
)

type ProtocolHandler struct {
	parser     Parser
	storage    Storage
	serializer Serializer
}

func NewProtocolHandler(
	parser Parser,
	storage Storage,
	serializer Serializer,
) *ProtocolHandler {
	return &ProtocolHandler{
		parser:     parser,
		storage:    storage,
		serializer: serializer,
	}
}

func (h *ProtocolHandler) Handle(
	ctx context.Context,
	request []byte,
) ([]Result, error) {
	cmd, err := h.parser.Parse(string(request))
	if err != nil {
		return errorResults(h.serializer.SerializeError(err)), nil
	}

	results, err := h.execute(ctx, cmd)
	if err != nil {
		return errorResults(h.serializer.SerializeError(err)), nil
	}

	return results, nil
}

func (h *ProtocolHandler) execute(ctx context.Context, cmd Command) ([]Result, error) {
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
		return singleResult(h.executeRepl(cmd))
	case PsyncCmd:
		return h.executePsync(cmd)
	default:
		return nil, fmt.Errorf("can't execute command: %w", ErrUnknownCommand)
	}
}

func singleResult(result Result, err error) ([]Result, error) {
	if err != nil {
		return nil, err
	}

	return []Result{result}, nil
}

func errorResults(message string) []Result {
	return []Result{StringResult(message)}
}
