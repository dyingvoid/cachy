package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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
	server, err := server.NewServer(*server.NewConfigMust(), log)
	if err != nil {
		panic(err)
	}

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}
