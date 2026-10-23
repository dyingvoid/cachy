package transport_tcp_client

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host         string        `envconfig:"HOST" default:"localhost"`
	Port         int           `envconfig:"PORT" default:"8081"`
	DialTimeout  time.Duration `envconfig:"DIAL_TIMEOUT" default:"100s"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"100s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"100s"`
}

func NewConfig() (Config, error) {
	var cfg Config

	if err := envconfig.Process("CLIENT", &cfg); err != nil {
		return Config{}, fmt.Errorf("replica: process config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("replica: validate config: %w", err)
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

func (c Config) validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("address is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.DialTimeout <= 0 {
		return fmt.Errorf("invalid dial timeout: %s", c.DialTimeout)
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %s", c.ReadTimeout)
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %s", c.WriteTimeout)
	}

	return nil
}
