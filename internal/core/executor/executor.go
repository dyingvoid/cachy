package core_executor

import (
	"fmt"

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
}

type ListStorage interface {
	RPush(key string, values ...string) (int, error)
	LPush(key string, values ...string) (int, error)
	LRange(key string, start, stop int) ([]string, error)
	LLen(key string) (int, error)
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
		r := domain.NewStringRecord(cmd.Value, cmd.Expiry)
		err := e.stringStorage.Set(cmd.Key, r)
		return domain.StringResult("OK"), err
	case domain.GetCmd:
		record, err := e.stringStorage.Get(cmd.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch: %w", err)
		}

		return domain.StringResult(record.Value), nil
	case domain.RPushCmd:
		n, err := e.listStorage.RPush(cmd.Key, cmd.Values...)
		if err != nil {
			return nil, err
		}

		return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
	case domain.LPushCmd:
		n, err := e.listStorage.LPush(cmd.Key, cmd.Values...)
		if err != nil {
			return nil, err
		}

		return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
	case domain.LRangeCmd:
		values, err := e.listStorage.LRange(cmd.Key, cmd.Start, cmd.Stop)
		if err != nil {
			return nil, err
		}

		return domain.ListResult(values), nil
	case domain.LLenCmd:
		n, err := e.listStorage.LLen(cmd.Key)
		if err != nil {
			return nil, err
		}

		return domain.StringResult(fmt.Sprintf("(integer) %d", n)), nil
	default:
		return nil, fmt.Errorf("can't execute command: %w", domain_errors.ErrUnknownCommand)
	}
}
