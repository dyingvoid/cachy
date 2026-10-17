package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/cachy/internal/config"
	core_handler "github.com/dyingvoid/cachy/internal/core/handler"
	core_parser "github.com/dyingvoid/cachy/internal/core/parser"
	core_serializer "github.com/dyingvoid/cachy/internal/core/serializer"
	core_storage "github.com/dyingvoid/cachy/internal/core/storage"
	transport_tcp "github.com/dyingvoid/cachy/internal/server/transport/tcp"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	appCfg := config.NewMust()
	config.Set(&appCfg)
	time.Local = appCfg.TimeZone
	serverConfig := transport_tcp.NewConfigMust()
	storageCfg := core_storage.NewConfigMust()

	log.Info(
		"config",
		slog.Int("server port", serverConfig.Port),
		slog.String("time_zone", appCfg.TimeZone.String()),
		slog.Any("storage", *storageCfg),
	)

	memStorage := core_storage.NewMemoryStorage(log, *storageCfg)
	listStorage := core_storage.NewListStorage(*storageCfg)
	parser := core_parser.NewParser()
	serializer := core_serializer.NewTextSerializer()

	h := core_handler.NewProtocolHandler(parser, memStorage, listStorage, serializer)

	server := transport_tcp.NewServer(
		serverConfig, log, h,
	)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}
