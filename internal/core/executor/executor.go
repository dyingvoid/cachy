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
	switch cmd.Name {
	case domain.CommandSet:
		err := e.storage.Set(cmd.Args[0], cmd.Args[1])
		return "OK", err
	case domain.CommandGet:
		value, err := e.storage.Get(cmd.Args[0])
		if err != nil {
			return "", err
		}

		return domain.Result(value), nil
	default:
		return "", fmt.Errorf("%w: %s", domain_errors.ErrUnknownCommand, cmd.Name)
	}
}
