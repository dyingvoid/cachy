package config

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Role     string `envconfig:"role" required:"true"`
	TimeZone *time.Location
}

var global atomic.Pointer[Config]

func Get() *Config {
	return global.Load()
}

func Set(cfg *Config) {
	global.Store(cfg)
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("REPLICATION", &cfg); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	tz := os.Getenv("TIME_ZONE")
	if tz == "" {
		tz = "UTC"
	}

	location, err := time.LoadLocation(tz)
	if err != nil {
		return Config{}, fmt.Errorf("load time zone: %s: %w", tz, err)
	}
	cfg.TimeZone = location

	return cfg, nil
}

func NewMust() Config {
	cfg, err := New()
	if err != nil {
		panic(fmt.Errorf("get config: %w", err))
	}

	return cfg
}
