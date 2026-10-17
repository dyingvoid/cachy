package core_executor

import (
	"fmt"

	"github.com/dyingvoid/cachy/internal/config"
	core_state "github.com/dyingvoid/cachy/internal/core/state"
	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

type Executor struct {
	stringStorage StringStorage
	listStorage   ListStorage
}

type StringStorage interface {
	Set(key string, value domain.StringRecord) error
	Get(key string) (domain.StringRecord, error)
	Delete(key string) error
	Exists(key string) bool
}

type ListStorage interface {
	RPush(key string, values ...string) (int, error)
	LPush(key string, values ...string) (int, error)
	LRange(key string, start, stop int) ([]string, error)
	LLen(key string) (int, error)
	LPop(key string, count int) ([]string, error)
	Exists(key string) bool
}

func NewExecutor(
	stringStorage StringStorage,
	listStorage ListStorage,
) *Executor {
	return &Executor{
		stringStorage: stringStorage,
		listStorage:   listStorage,
	}
}

func (e *Executor) Execute(cmd domain.Command) (domain.Result, error) {
	switch cmd := cmd.(type) {
	case domain.SetCmd:
		return e.executeSet(cmd)
	case domain.GetCmd:
		return e.executeGet(cmd)
	case domain.RPushCmd:
		return e.executeRPush(cmd)
	case domain.LPushCmd:
		return e.executeLPush(cmd)
	case domain.LRangeCmd:
		return e.executeLRange(cmd)
	case domain.LLenCmd:
		return e.executeLLen(cmd)
	case domain.LPopCmd:
		return e.executeLPop(cmd)
	case domain.InfoCmd:
		return e.executeInfo(cmd)
	default:
		return nil, fmt.Errorf("can't execute command: %w", domain_errors.ErrUnknownCommand)
	}
}

func (e *Executor) executeInfo(cmd domain.InfoCmd) (domain.Result, error) {
	cfg := config.Get()
	info := domain.NewInfo(
		cfg.Role, cfg.ReplicaOf,
		core_state.ReplicationID(),
		core_state.ReplicationOffset(),
	)

	return domain.MapResult(info), nil
}

func (e *Executor) executeLLen(cmd domain.LLenCmd) (domain.Result, error) {
	if err := e.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := e.listStorage.LLen(cmd.Key)
	if err != nil {
		return nil, err
	}

	return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
}

func (e *Executor) executeLPop(cmd domain.LPopCmd) (domain.Result, error) {
	if err := e.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	values, err := e.listStorage.LPop(cmd.Key, cmd.Count)
	if err != nil {
		return nil, err
	}

	return domain.ListResult(values), nil
}

func (e *Executor) executeLRange(cmd domain.LRangeCmd) (domain.Result, error) {
	if err := e.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	values, err := e.listStorage.LRange(cmd.Key, cmd.Start, cmd.Stop)
	if err != nil {
		return nil, err
	}

	return domain.ListResult(values), nil
}

func (e *Executor) executeLPush(cmd domain.LPushCmd) (domain.Result, error) {
	if err := e.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := e.listStorage.LPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
}

func (e *Executor) executeRPush(cmd domain.RPushCmd) (domain.Result, error) {
	if err := e.ensureListKey(cmd.Key); err != nil {
		return nil, err
	}

	n, err := e.listStorage.RPush(cmd.Key, cmd.Values...)
	if err != nil {
		return nil, err
	}

	return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
}

func (e *Executor) executeGet(cmd domain.GetCmd) (domain.Result, error) {
	if err := e.ensureStringKey(cmd.Key); err != nil {
		return nil, err
	}

	record, err := e.stringStorage.Get(cmd.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	return domain.StringResult(record.Value), nil
}

func (e *Executor) executeSet(cmd domain.SetCmd) (domain.Result, error) {
	if err := e.ensureStringKey(cmd.Key); err != nil {
		return nil, err
	}

	r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
	if err := e.stringStorage.Set(cmd.Key, r); err != nil {
		return nil, err
	}

	return domain.StringResult("OK"), nil
}

func (e *Executor) ensureStringKey(key string) error {
	if e.listStorage.Exists(key) {
		return domain_errors.ErrWrongType
	}

	return nil
}

func (e *Executor) ensureListKey(key string) error {
	if e.stringStorage.Exists(key) {
		return domain_errors.ErrWrongType
	}

	return nil
}
