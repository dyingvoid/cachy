package core_storage

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ShardCount       int           `envconfig:"SHARD_COUNT" default:"256"`
	CleanupInterval  time.Duration `envconfig:"CLEANUP_INTERVAL" default:"1s"`
	CleanupBatchSize int           `envconfig:"CLEANUP_BATCH_SIZE" default:"1000"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("STORAGE", &cfg); err != nil {
		return nil, fmt.Errorf("storage: process config: %w", err)
	}
	if cfg.ShardCount <= 0 {
		return nil, fmt.Errorf("storage: invalid shard count: %d", cfg.ShardCount)
	}
	if cfg.CleanupInterval <= 0 {
		return nil, fmt.Errorf("storage: invalid cleanup interval: %s", cfg.CleanupInterval)
	}
	if cfg.CleanupBatchSize <= 0 {
		return nil, fmt.Errorf("storage: invalid cleanup batch size: %d", cfg.CleanupBatchSize)
	}
	return &cfg, nil
}

func NewConfigMust() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
