package core_executor

import "github.com/dyingvoid/cachy/internal/domain"

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
	// TODO: constants and enums for cmds
	switch cmd.Name {
	case "SET":
		err := e.storage.Set(cmd.Args[0], cmd.Args[1])
		return "OK", err
	case "GET":
		value, err := e.storage.Get(cmd.Args[0])
		if err != nil {
			return "", err
		}

		return domain.Result(value), nil
	default:
		panic("Unknown command: " + cmd.Name)
	}
}
