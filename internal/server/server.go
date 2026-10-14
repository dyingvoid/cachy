package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Server struct {
	cfg Config
	log *slog.Logger

	ln net.Listener
	wg sync.WaitGroup
}

func NewServer(cfg Config, log *slog.Logger) (*Server, error) {
	return &Server{
		cfg: cfg,
		log: log,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.log.Warn("start TCP server", slog.String("addr", addr))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s:")
	}
	s.ln = ln

	go func() {
		<-ctx.Done()
		s.log.Warn("server shutting down")
		if err := s.ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.log.Error("server shutdown failed", slog.Any("error", err))
		}
	}()

	return s.loop(ctx)
}

func (s *Server) loop(ctx context.Context) error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			var nerr net.Error
			if errors.As(err, &nerr) && nerr.Timeout() {
				s.log.Warn("accept timeout", slog.Any("error", err))
				continue
			}

			s.log.Error("accept connection", slog.Any("error", err))
			continue
		}

		s.wg.Go(func() {
			s.handle(ctx, conn)
		})
	}

	s.wg.Wait()
	s.log.Warn("TCP server stopped")
	return ctx.Err()
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))

		if !scanner.Scan() {
			break
		}

		_ = scanner.Text()

		conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
	}

	if err := scanner.Err(); err != nil {
		s.log.Error(
			"read error",
			slog.String("addr", conn.RemoteAddr().String()),
			slog.Any("error", err),
		)
	}
}
