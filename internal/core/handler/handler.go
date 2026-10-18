package core_handler

import (
	"context"
	"fmt"

	"github.com/dyingvoid/cachy/internal/config"
	core_state "github.com/dyingvoid/cachy/internal/core/state"
	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
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

func (h *ProtocolHandler) execute(ctx context.Context, cmd domain.Command) (domain.Result, error) {
	switch cmd := cmd.(type) {
	case domain.PingCmd:
		return domain.StringResult("PONG"), nil
	case domain.SetCmd:
		return h.executeSet(cmd)
	case domain.GetCmd:
		return h.executeGet(cmd)
	case domain.RPushCmd:
		return h.executeRPush(cmd)
	case domain.LPushCmd:
		return h.executeLPush(cmd)
	case domain.LRangeCmd:
		return h.executeLRange(cmd)
	case domain.LLenCmd:
		return h.executeLLen(cmd)
	case domain.LPopCmd:
		return h.executeLPop(cmd)
	case domain.InfoCmd:
		return h.executeInfo(cmd)
	case domain.ReplConfCmd:
		return domain.StringResult("+OK"), nil
	default:
		return nil, fmt.Errorf("can't execute command: %w", domain_errors.ErrUnknownCommand)
	}
}

func (h *ProtocolHandler) executeInfo(cmd domain.InfoCmd) (domain.Result, error) {
	cfg := config.Get()
	info := domain.NewInfo(
		cfg.Role, cfg.ReplicaOf,
		core_state.ReplicationID(),
		core_state.ReplicationOffset(),
	)

	return domain.MapResult(info), nil
}

func (h *ProtocolHandler) executeLLen(cmd domain.LLenCmd) (domain.Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := h.listStorage.LLen(cmd.Key)
	if err != nil {
		return nil, err
	}

	return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
}

func (h *ProtocolHandler) executeLPop(cmd domain.LPopCmd) (domain.Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	values, err := h.listStorage.LPop(cmd.Key, cmd.Count)
	if err != nil {
		return nil, err
	}

	return domain.ListResult(values), nil
}

func (h *ProtocolHandler) executeLRange(cmd domain.LRangeCmd) (domain.Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	values, err := h.listStorage.LRange(cmd.Key, cmd.Start, cmd.Stop)
	if err != nil {
		return nil, err
	}

	return domain.ListResult(values), nil
}

func (h *ProtocolHandler) executeLPush(cmd domain.LPushCmd) (domain.Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := h.listStorage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
}

func (h *ProtocolHandler) executeRPush(cmd domain.RPushCmd) (domain.Result, error) {
	if err := h.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := h.listStorage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
}

func (h *ProtocolHandler) executeGet(cmd domain.GetCmd) (domain.Result, error) {
	if err := h.ensureStringKey(cmd.Key); err != nil {
		return nil, err
	}

	record, err := h.stringStorage.Get(cmd.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return domain.StringResult(record.Value), nil
}

func (h *ProtocolHandler) executeSet(cmd domain.SetCmd) (domain.Result, error) {
	if err := h.ensureStringKey(cmd.Key); err != nil {
		return nil, err
	}

	r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
	if err := h.stringStorage.Set(cmd.Key, r); err != nil {
		return nil, err
	}

	return domain.StringResult("OK"), nil
}

func (h *ProtocolHandler) ensureStringKey(key string) error {
	if h.listStorage.Exists(key) {
		return domain_errors.ErrWrongType
	}

	return nil
}

func (h *ProtocolHandler) ensureListKey(key string) error {
	if h.stringStorage.Exists(key) {
		return domain_errors.ErrWrongType
	}

	return nil
}
