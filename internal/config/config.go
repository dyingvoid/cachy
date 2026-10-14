package config

import (
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TimeZone *time.Location
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
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
