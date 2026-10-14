package server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port         int           `envconfig:"PORT"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("SERVER", &cfg); err != nil {
		return nil, fmt.Errorf("server: process config: %w", err)
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", cfg.Port)
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
