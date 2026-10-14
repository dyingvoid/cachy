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
	RPush(key string, value string) error
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
		return "OK", err
	case domain.GetCmd:
		record, err := e.stringStorage.Get(cmd.Key)
		if err != nil {
			return domain.Result(record.Value), fmt.Errorf("failed to fetch: %w", err)
		}

		return domain.Result(record.Value), nil
	default:
		return "", fmt.Errorf("can't execute command: %w", domain_errors.ErrUnknownCommand)
	}
}
