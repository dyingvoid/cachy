package core_storage

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)


type Config struct {
	ShardCount int `envconfig:"SHARD_COUNT" default:"256"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("STORAGE", &cfg); err != nil {
		return nil, fmt.Errorf("storage: process config: %w", err)
	}
	if cfg.ShardCount <= 0 {
		return nil, fmt.Errorf("storage: invalid shard count: %d", cfg.ShardCount)
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
