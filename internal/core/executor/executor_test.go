package core_executor

import (
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	core_storage "github.com/dyingvoid/cachy/internal/core/storage"
	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

func newTestExecutor() *Executor {
	cfg := core_storage.Config{
		ShardCount:       4,
		CleanupInterval:  time.Hour,
		CleanupBatchSize: 1000,
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewExecutor(
		core_storage.NewMemoryStorage(log, cfg),
		core_storage.NewListStorage(cfg),
	)
}

func TestExecutor_StringCommandOnListKey(t *testing.T) {
	cmds := []domain.Command{
		domain.GetCmd{Key: "key"},
		domain.SetCmd{Key: "key", Value: "value"},
	}

	for _, cmd := range cmds {
		t.Run(commandName(cmd), func(t *testing.T) {
			e := newTestExecutor()
			if _, err := e.Execute(domain.RPushCmd{Key: "key", Values: []string{"a"}}); err != nil {
				t.Fatalf("RPush returned unexpected error: %v", err)
			}

			res, err := e.Execute(cmd)
			if !errors.Is(err, domain_errors.ErrWrongType) {
				t.Fatalf("Execute(%s) error = %v, want ErrWrongType", commandName(cmd), err)
			}
			if res != nil {
				t.Fatalf("Execute(%s) result = %v, want nil", commandName(cmd), res)
			}
		})
	}
}

func TestExecutor_ListCommandOnStringKey(t *testing.T) {
	cmds := []domain.Command{
		domain.RPushCmd{Key: "key", Values: []string{"a"}},
		domain.LPushCmd{Key: "key", Values: []string{"a"}},
		domain.LRangeCmd{Key: "key", Start: 0, Stop: -1},
		domain.LLenCmd{Key: "key"},
		domain.LPopCmd{Key: "key", Count: 1},
	}

	for _, cmd := range cmds {
		t.Run(commandName(cmd), func(t *testing.T) {
			e := newTestExecutor()
			if _, err := e.Execute(domain.SetCmd{Key: "key", Value: "value"}); err != nil {
				t.Fatalf("Set returned unexpected error: %v", err)
			}

			res, err := e.Execute(cmd)
			if !errors.Is(err, domain_errors.ErrWrongType) {
				t.Fatalf("Execute(%s) error = %v, want ErrWrongType", commandName(cmd), err)
			}
			if res != nil {
				t.Fatalf("Execute(%s) result = %v, want nil", commandName(cmd), res)
			}
		})
	}
}

func TestExecutor_KeyKeepsSingleType(t *testing.T) {
	e := newTestExecutor()

	if _, err := e.Execute(domain.RPushCmd{Key: "key", Values: []string{"a", "b"}}); err != nil {
		t.Fatalf("RPush returned unexpected error: %v", err)
	}
	if _, err := e.Execute(domain.SetCmd{Key: "key", Value: "value"}); !errors.Is(err, domain_errors.ErrWrongType) {
		t.Fatalf("Set on list key error = %v, want ErrWrongType", err)
	}

	res, err := e.Execute(domain.LRangeCmd{Key: "key", Start: 0, Stop: -1})
	if err != nil {
		t.Fatalf("LRange returned unexpected error: %v", err)
	}
	want := domain.ListResult{"a", "b"}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("LRange = %v, want %v", res, want)
	}

	if _, err := e.Execute(domain.SetCmd{Key: "other", Value: "value"}); err != nil {
		t.Fatalf("Set returned unexpected error: %v", err)
	}
	if _, err := e.Execute(domain.RPushCmd{Key: "other", Values: []string{"a"}}); !errors.Is(err, domain_errors.ErrWrongType) {
		t.Fatalf("RPush on string key error = %v, want ErrWrongType", err)
	}

	res, err = e.Execute(domain.GetCmd{Key: "other"})
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if want := domain.StringResult("value"); res != want {
		t.Fatalf("Get = %v, want %v", res, want)
	}
}

func TestExecutor_LPop(t *testing.T) {
	e := newTestExecutor()

	if _, err := e.Execute(domain.RPushCmd{Key: "key", Values: []string{"a", "b", "c"}}); err != nil {
		t.Fatalf("RPush returned unexpected error: %v", err)
	}

	res, err := e.Execute(domain.LPopCmd{Key: "key", Count: 2})
	if err != nil {
		t.Fatalf("LPop returned unexpected error: %v", err)
	}
	if want := (domain.ListResult{"a", "b"}); !reflect.DeepEqual(res, want) {
		t.Fatalf("LPop = %v, want %v", res, want)
	}

	res, err = e.Execute(domain.LRangeCmd{Key: "key", Start: 0, Stop: -1})
	if err != nil {
		t.Fatalf("LRange returned unexpected error: %v", err)
	}
	if want := (domain.ListResult{"c"}); !reflect.DeepEqual(res, want) {
		t.Fatalf("LRange after LPop = %v, want %v", res, want)
	}

	if _, err := e.Execute(domain.LPopCmd{Key: "key", Count: 1}); err != nil {
		t.Fatalf("LPop returned unexpected error: %v", err)
	}

	if _, err := e.Execute(domain.SetCmd{Key: "key", Value: "value"}); err != nil {
		t.Fatalf("Set on emptied list key returned unexpected error: %v", err)
	}

	res, err = e.Execute(domain.LPopCmd{Key: "missing", Count: 2})
	if err != nil {
		t.Fatalf("LPop on missing key returned unexpected error: %v", err)
	}
	if want := (domain.ListResult{}); !reflect.DeepEqual(res, want) {
		t.Fatalf("LPop on missing key = %v, want %v", res, want)
	}
}

func commandName(cmd domain.Command) string {
	switch cmd.(type) {
	case domain.SetCmd:
		return "SET"
	case domain.GetCmd:
		return "GET"
	case domain.RPushCmd:
		return "RPUSH"
	case domain.LPushCmd:
		return "LPUSH"
	case domain.LRangeCmd:
		return "LRANGE"
	case domain.LLenCmd:
		return "LLEN"
	case domain.LPopCmd:
		return "LPOP"
	default:
		return "UNKNOWN"
	}
}
