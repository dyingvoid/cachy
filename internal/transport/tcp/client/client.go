package transport_tcp_client

import (
	"bufio"
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
	cfg    Config
	conn   net.Conn
	reader *bufio.Reader

	sendMu    sync.Mutex
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
		cfg:    cfg,
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) Send(ctx context.Context, payload []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("send: %w", ErrEmptyPayload)
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	if c.closed.Load() {
		return nil, fmt.Errorf("send: %w", ErrClosed)
	}

	if err := c.conn.SetWriteDeadline(deadline(ctx, c.cfg.WriteTimeout)); err != nil {
		return nil, fmt.Errorf("set write deadline: %w", err)
	}

	n, err := c.conn.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("write payload: %w", err)
	}
	if n != len(payload) {
		return nil, fmt.Errorf("write payload: %w", io.ErrShortWrite)
	}

	if err := c.conn.SetReadDeadline(deadline(ctx, c.cfg.ReadTimeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	raw, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return raw, nil
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

func deadline(ctx context.Context, timeout time.Duration) time.Time {
	limit := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(limit) {
		return ctxDeadline
	}

	return limit
}
