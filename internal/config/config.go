package config

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Role      string  `envconfig:"role" required:"true"`
	ReplicaOf *string `envconfig:"replica_of"`
	TimeZone  *time.Location
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

	return cfg, cfg.validate()
}

func (c Config) validate() error {
	switch c.Role {
	case "master":
	case "slave":
		if c.ReplicaOf == nil || *c.ReplicaOf == "" {
			return errors.New("replica_of is required for slave role")
		}
	default:
		return errors.New("role must be master or slave")
	}

	return nil
}

func NewMust() Config {
	cfg, err := New()
	if err != nil {
		panic(fmt.Errorf("get config: %w", err))
	}

	return cfg
}
