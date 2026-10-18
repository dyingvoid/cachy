package features

import (
	"context"
	"fmt"
)

type ProtocolHandler struct {
	parser        Parser
	stringStorage StringStorage
	listStorage   ListStorage
	serializer    Serializer
}

func NewProtocolHandler(
	parser Parser,
	stringStorage StringStorage,
	listStorage ListStorage,
	serializer Serializer,
) *ProtocolHandler {
	return &ProtocolHandler{
		parser:        parser,
		stringStorage: stringStorage,
		listStorage:   listStorage,
		serializer:    serializer,
	}
}

func (h *ProtocolHandler) Handle(
	ctx context.Context,
	request []byte,
) ([]byte, error) {
	cmd, err := h.parser.Parse(string(request))
	if err != nil {
		return []byte(h.serializer.SerializeError(err)), nil
	}

	result, err := h.execute(ctx, cmd)
	if err != nil {
		return []byte(h.serializer.SerializeError(err)), nil
	}

	return []byte(h.serializer.Serialize(result)), nil
}

func (h *ProtocolHandler) execute(ctx context.Context, cmd Command) (Result, error) {
	switch cmd := cmd.(type) {
	case PingCmd:
		return StringResult("PONG"), nil
	case SetCmd:
		return h.executeSet(cmd)
	case GetCmd:
		return h.executeGet(cmd)
	case RPushCmd:
		return h.executeRPush(cmd)
	case LPushCmd:
		return h.executeLPush(cmd)
	case LRangeCmd:
		return h.executeLRange(cmd)
	case LLenCmd:
		return h.executeLLen(cmd)
	case LPopCmd:
		return h.executeLPop(cmd)
	case InfoCmd:
		return h.executeInfo(cmd)
	case ReplConfCmd:
		return StringResult("+OK"), nil
	default:
		return nil, fmt.Errorf("can't execute command: %w", ErrUnknownCommand)
	}
}

func (h *ProtocolHandler) ensureStringKey(key string) error {
	if h.listStorage.Exists(key) {
		return ErrWrongType
	}

	return nil
}

func (h *ProtocolHandler) ensureListKey(key string) error {
	if h.stringStorage.Exists(key) {
		return ErrWrongType
	}

	return nil
}
