package transport_tcp_client

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Address      string
	Port         int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c Config) validate() error {
	if strings.TrimSpace(c.Address) == "" {
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
