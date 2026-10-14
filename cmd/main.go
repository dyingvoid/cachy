package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	core_executor "github.com/dyingvoid/cachy/internal/core/executor"
	core_parser "github.com/dyingvoid/cachy/internal/core/parser"
	core_storage "github.com/dyingvoid/cachy/internal/core/storage"
	"github.com/dyingvoid/cachy/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	// TODO: log config
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	memStorage := core_storage.NewMemoryStorage()
	executor := core_executor.NewExecutor(memStorage)
	parser := core_parser.NewParser()

	server := server.NewServer(
		*server.NewConfigMust(), log,
		parser, executor,
	)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}
