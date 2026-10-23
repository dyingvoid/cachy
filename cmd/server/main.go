package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/cachy/internal/config"
	"github.com/dyingvoid/cachy/internal/features"
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
	featuresCfg := features_replication.NewConfigMust()

	log.Info(
		"config",
		slog.Int("server port", serverConfig.Port),
		slog.String("time_zone", appCfg.TimeZone.String()),
		slog.Any("storage", *storageCfg),
		slog.Any("features", featuresCfg),
	)

	storage := infrastructure_storage.NewMemoryStorage(log, *storageCfg)
	serializer := &infrastructure_resp3.Serializer{}
	newDeserializer := func(r io.Reader) transport_tcp_server.Deserializer {
		return infrastructure_resp3.NewDeserializer(r)
	}
	replicationHandler := features_replication.NewReplicationHandler(featuresCfg.ReplicationBuffer)

	h := features.NewProtocolHandler(storage, replicationHandler)

	if appCfg.Role == "slave" {
		replicaTransport, replicaClient, err := handshakeReplication(ctx)
		if err != nil {
			panic(err)
		}
		defer replicaTransport.Close()

		go func() {
			if err := replicaClient.Listen(ctx, h); err != nil && ctx.Err() == nil {
				log.Error("replication listen", slog.Any("error", err))
			}
		}()

		log.Info(
			"replication handshake completed",
			slog.String("master", *appCfg.ReplicaOf),
			slog.Int("listening_port", serverConfig.Port),
		)
	}

	server := transport_tcp_server.NewServer(
		serverConfig, log, h, serializer, newDeserializer,
	)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		panic(err)
	}
}

func handshakeReplication(
	ctx context.Context,
) (*transport_tcp_client.Client, *features_replication.ReplicationClient, error) {
	cfg, err := transport_tcp_client.NewConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("replica client config: %w", err)
	}

	client, err := transport_tcp_client.NewClient(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("connect replica: %w", err)
	}

	replicationClient := features_replication.NewReplicationClient(
		client,
		&infrastructure_resp3.Serializer{},
		infrastructure_resp3.NewDeserializer(client),
	)
	if err := replicationClient.Handshake(ctx); err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("handshake replica: %w", err)
	}

	return client, replicationClient, nil
}
