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
	Set(key, value string) error
	Get(key string) (string, error)
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
		err := e.storage.Set(cmd.Key, cmd.Value)
		return "OK", err
	case domain.GetCmd:
		value, err := e.storage.Get(cmd.Key)
		if err != nil {
			return "", err
		}

		return domain.Result(value), nil
	default:
		return "", fmt.Errorf("%w", domain_errors.ErrUnknownCommand)
	}
}
