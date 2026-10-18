package transport_tcp_client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrEmptyPayload = errors.New("empty payload")
	ErrClosed       = errors.New("client closed")
)

type Client struct {
	cfg  Config
	conn net.Conn

	readMu    sync.Mutex
	writeMu   sync.Mutex
	closeOnce sync.Once
	closeErr  error
	closed    atomic.Bool
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	dialer := net.Dialer{Timeout: cfg.DialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Address, err)
	}

	return &Client{
		cfg:  cfg,
		conn: conn,
	}, nil
}

func (c *Client) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if c.closed.Load() {
		return 0, fmt.Errorf("read: %w", ErrClosed)
	}

	if err := c.conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout)); err != nil {
		return 0, fmt.Errorf("set read deadline: %w", err)
	}

	n, err := c.conn.Read(p)
	if err != nil {
		return n, fmt.Errorf("read response: %w", err)
	}

	return n, nil
}

func (c *Client) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("write: %w", ErrEmptyPayload)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.closed.Load() {
		return 0, fmt.Errorf("write: %w", ErrClosed)
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout)); err != nil {
		return 0, fmt.Errorf("set write deadline: %w", err)
	}

	n, err := c.conn.Write(p)
	if err != nil {
		return n, fmt.Errorf("write payload: %w", err)
	}
	if n != len(p) {
		return n, fmt.Errorf("write payload: %w", io.ErrShortWrite)
	}

	return n, nil
}

func (c *Client) Port() int {
	return c.cfg.Port
}

func (c *Client) Close() error {
	c.closed.Store(true)

	c.closeOnce.Do(func() {
		c.closeErr = c.conn.Close()
	})

	if c.closeErr != nil {
		return fmt.Errorf("close connection: %w", c.closeErr)
	}

	return nil
}
