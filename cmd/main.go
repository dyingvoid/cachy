package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/cachy/internal/config"
	core_serializer "github.com/dyingvoid/cachy/internal/core/serializer"
	core_storage "github.com/dyingvoid/cachy/internal/core/storage"
	"github.com/dyingvoid/cachy/internal/features"
	transport_tcp_server "github.com/dyingvoid/cachy/internal/transport/tcp/server"
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
	serverConfig := transport_tcp_server.NewConfigMust()
	storageCfg := core_storage.NewConfigMust()

	log.Info(
		"config",
		slog.Int("server port", serverConfig.Port),
		slog.String("time_zone", appCfg.TimeZone.String()),
		slog.Any("storage", *storageCfg),
	)

	memStorage := core_storage.NewMemoryStorage(log, *storageCfg)
	listStorage := core_storage.NewListStorage(*storageCfg)
	parser := features.NewProtocolParser()
	serializer := core_serializer.NewTextSerializer()

	h := features.NewProtocolHandler(parser, memStorage, listStorage, serializer)

	server := transport_tcp_server.NewServer(
		serverConfig, log, h,
	)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}
