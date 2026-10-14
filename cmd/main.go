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

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	serverCfg := server.NewConfigMust()
	storageCfg := core_storage.NewConfigMust()
	log.Info("config", slog.Any("server", *serverCfg), slog.Any("storage", *storageCfg))

	memStorage := core_storage.NewMemoryStorage(*storageCfg)
	executor := core_executor.NewExecutor(memStorage)
	parser := core_parser.NewParser()

	server := server.NewServer(
		*serverCfg, log,
		parser, executor,
	)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}
