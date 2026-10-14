package core_executor

import (
	"fmt"

	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

type Executor struct {
	storage Storage
}

type Storage interface {
	Set(key string, value domain.Record) error
	Get(key string) (domain.Record, error)
	Delete(key string) error
}

func NewExecutor(
	storage Storage,
) *Executor {
	return &Executor{
		storage: storage,
	}
}

func (e *Executor) Execute(cmd domain.Command) (domain.Result, error) {
	switch cmd := cmd.(type) {
	case domain.SetCmd:
		r := domain.NewRecord(cmd.Value, cmd.Expiry)
		err := e.storage.Set(cmd.Key, r)
		return "OK", err
	case domain.GetCmd:
		record, err := e.storage.Get(cmd.Key)
		if err != nil {
			return domain.Result(record.Value), fmt.Errorf("failed to fetch: %w", err)
		}

		return domain.Result(record.Value), nil
	default:
		return "", fmt.Errorf("can't execute command: %w", domain_errors.ErrUnknownCommand)
	}
}
