package features

import (
	"context"

	"github.com/dyingvoid/cachy/internal/domain"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	"github.com/dyingvoid/cachy/internal/features/result"
)

type Executor interface {
	Execute(context.Context, features_commands.Command) ([]result.Result, error)
}

type Serializer interface {
	Serialize(result.Result) string
	SerializeCommand(features_commands.WriteCommand) string
	SerializeError(error) string
}

type Storage interface {
	Set(key string, value domain.StringRecord) error
	Get(key string) (domain.StringRecord, error)
	Delete(key string) error
	Exists(key string) bool

	RPush(key string, values ...string) (int, error)
	LPush(key string, values ...string) (int, error)
	LRange(key string, start, stop int) ([]string, error)
	LLen(key string) (int, error)
	LPop(key string, count int) ([]string, error)
}
