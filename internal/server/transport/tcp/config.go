package transport_tcp

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Port         int           `envconfig:"PORT"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT"`
}

func NewConfig() (ServerConfig, error) {
	var cfg ServerConfig
	if err := envconfig.Process("SERVER", &cfg); err != nil {
		return ServerConfig{}, fmt.Errorf("server: process config: %w", err)
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return ServerConfig{}, fmt.Errorf("invalid port: %d", cfg.Port)
	}
	return cfg, nil
}

func NewConfigMust() ServerConfig {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
