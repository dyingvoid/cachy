package core_handler

import (
	"context"

	"github.com/dyingvoid/cachy/internal/domain"
)

type Parser interface {
	Parse(in string) (domain.Command, error)
}

type Executor interface {
	Execute(context.Context, domain.Command) (domain.Result, error)
}

type Serializer interface {
	Serialize(domain.Result) string
	SerializeError(error) string
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
