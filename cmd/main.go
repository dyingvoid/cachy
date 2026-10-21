package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/cachy/internal/config"
	"github.com/dyingvoid/cachy/internal/features"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_replication "github.com/dyingvoid/cachy/internal/features/replication"
	infrastructure_resp3 "github.com/dyingvoid/cachy/internal/infrastructure/resp3"
	infrastructure_storage "github.com/dyingvoid/cachy/internal/infrastructure/storage"
	transport_tcp_client "github.com/dyingvoid/cachy/internal/transport/tcp/client"
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
	storageCfg := infrastructure_storage.NewConfigMust()
	featuresCfg := features.NewConfigMust()

	log.Info(
		"config",
		slog.Int("server port", serverConfig.Port),
		slog.String("time_zone", appCfg.TimeZone.String()),
		slog.Any("storage", *storageCfg),
		slog.Any("features", featuresCfg),
	)

	storage := infrastructure_storage.NewMemoryStorage(log, *storageCfg)
	parser := features_commands.NewProtocolParser()
	serializer := &infrastructure_resp3.Serializer{}
	replicationHandler := features.NewReplicationHandler(featuresCfg.ReplicationBuffer)

	h := features.NewProtocolHandler(parser, storage, replicationHandler)

	if appCfg.Role == "slave" {
		replicaClient, err := handshakeReplication(
			ctx, *appCfg.ReplicaOf, serverConfig.Port,
		)
		if err != nil {
			panic(err)
		}
		defer replicaClient.Close()

		log.Info(
			"replication handshake completed",
			slog.String("master", *appCfg.ReplicaOf),
			slog.Int("listening_port", serverConfig.Port),
		)
	}

	server := transport_tcp_server.NewServer(
		serverConfig, log, h, serializer,
	)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}

func handshakeReplication(
	ctx context.Context,
	address string,
	port int,
) (*transport_tcp_client.Client, error) {
	cfg, err := transport_tcp_client.NewConfig(address, port)
	if err != nil {
		return nil, fmt.Errorf("replica client config: %w", err)
	}

	client, err := transport_tcp_client.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect replica: %w", err)
	}

	replicationClient := features_replication.NewReplicationClient(
		features_replication.NewTextCodec(),
		client,
	)
	if err := replicationClient.Handshake(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("handshake replica: %w", err)
	}

	return client, nil
}
