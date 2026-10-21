package transport_tcp_client

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestClientReadWrite(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen returned unexpected error: %v", err)
	}
	defer listener.Close()

	requests := make(chan string, 1)
	serverErrors := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErrors <- err
			return
		}
		defer conn.Close()

		request, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			serverErrors <- err
			return
		}
		requests <- request

		if _, err := conn.Write([]byte("OK\r\n")); err != nil {
			serverErrors <- err
		}
	}()

	cfg := Config{
		Host:         listener.Addr().String(),
		Port:         8082,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient returned unexpected error: %v", err)
	}
	defer client.Close()

	if n, err := client.Write([]byte("PING\n")); err != nil {
		t.Fatalf("Write returned unexpected error: %v", err)
	} else if n != len("PING\n") {
		t.Fatalf("Write = %d, want %d", n, len("PING\n"))
	}

	raw := make([]byte, 4)
	if _, err := io.ReadFull(client, raw); err != nil {
		t.Fatalf("Read returned unexpected error: %v", err)
	}
	if want := "OK\r\n"; string(raw) != want {
		t.Fatalf("Read = %q, want %q", raw, want)
	}

	select {
	case request := <-requests:
		if want := "PING\n"; request != want {
			t.Fatalf("server received %q, want %q", request, want)
		}
	case err := <-serverErrors:
		t.Fatalf("server returned unexpected error: %v", err)
	case <-time.After(time.Second):
		t.Fatal("server did not receive request")
	}

	if got := client.Port(); got != cfg.Port {
		t.Fatalf("Port = %d, want %d", got, cfg.Port)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
}

func TestClientWriteAfterClose(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen returned unexpected error: %v", err)
	}
	defer listener.Close()

	cfg := Config{
		Host:         listener.Addr().String(),
		Port:         8082,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient returned unexpected error: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}

	if _, err := client.Write([]byte("PING\n")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Write error = %v, want %v", err, ErrClosed)
	}

	if _, err := client.Read(make([]byte, 4)); !errors.Is(err, ErrClosed) {
		t.Fatalf("Read error = %v, want %v", err, ErrClosed)
	}
}

func TestNewClientValidation(t *testing.T) {
	valid := Config{
		Host:         "127.0.0.1:6379",
		Port:         8082,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{
			name:   "empty address",
			mutate: func(cfg *Config) { cfg.Host = " " },
		},
		{
			name:   "zero port",
			mutate: func(cfg *Config) { cfg.Port = 0 },
		},
		{
			name:   "port out of range",
			mutate: func(cfg *Config) { cfg.Port = 65536 },
		},
		{
			name:   "zero dial timeout",
			mutate: func(cfg *Config) { cfg.DialTimeout = 0 },
		},
		{
			name:   "zero read timeout",
			mutate: func(cfg *Config) { cfg.ReadTimeout = 0 },
		},
		{
			name:   "zero write timeout",
			mutate: func(cfg *Config) { cfg.WriteTimeout = 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)

			if _, err := NewClient(context.Background(), cfg); err == nil {
				t.Fatal("NewClient returned nil error")
			}
		})
	}
}
