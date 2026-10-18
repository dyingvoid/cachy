package features

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ReplicationBuffer int `envconfig:"REPLICATION_BUFFER" default:"100"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("FEATURES", &cfg); err != nil {
		return Config{}, fmt.Errorf("features: process config: %w", err)
	}
	if cfg.ReplicationBuffer <= 0 {
		return Config{}, fmt.Errorf("features: invalid replication buffer: %d", cfg.ReplicationBuffer)
	}
	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
