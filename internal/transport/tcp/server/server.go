package transport_tcp_server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	core_logger "github.com/dyingvoid/cachy/internal/core/logger"
	"github.com/dyingvoid/cachy/internal/features"
	features_commands "github.com/dyingvoid/cachy/internal/features/commands"
	features_conn "github.com/dyingvoid/cachy/internal/features/conn"
)

type Server struct {
	cfg ServerConfig
	log *slog.Logger

	ln net.Listener
	wg sync.WaitGroup

	handler Handler

	serializer Serializer
}

type Handler interface {
	Handle(ctx context.Context, state *features_conn.State, request []byte) (
		results []features.Result,
		err error,
	)
}

type Serializer interface {
	Serialize(features.Result) ([]byte, error)
	SerializeCommand(features_commands.WriteCommand) []byte
	BulkError(string) []byte
}

func NewServer(
	cfg ServerConfig,
	log *slog.Logger,
	handler Handler,
	serializer Serializer,
) *Server {
	return &Server{
		cfg:        cfg,
		log:        log,
		handler:    handler,
		serializer: serializer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.log.Warn("start TCP server", slog.String("addr", addr))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s:", addr)
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
	log := s.log.With(slog.String("addr", conn.RemoteAddr().String()))
	log.Info("New connection")

	ctx = core_logger.ToContext(ctx, log)

	var writeMu sync.Mutex
	var stream *features.ReplicationStream
	var streamDone chan struct{}

	defer func() {
		conn.Close()
		if stream != nil {
			stream.Close()
			<-streamDone
		}
	}()

	state := features_conn.NewState()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))
		results, err := s.handler.Handle(ctx, state, scanner.Bytes())
		if err != nil {
			if err := s.write(ctx, conn, &writeMu, s.serializer.BulkError(err.Error())); err != nil {
				return
			}
			continue
		}

		for _, result := range results {
			streaming, ok := result.(features.StreamingResult)
			if ok {
				if stream == nil && streaming.Stream != nil {
					stream = streaming.Stream
					streamDone = make(chan struct{})
					go s.writeStream(ctx, conn, &writeMu, stream, streamDone)
				}
				continue
			}

			bytes, err := s.serializer.Serialize(result)
			if err != nil {
				if err := s.write(ctx, conn, &writeMu, s.serializer.BulkError(err.Error())); err != nil {
					return
				}
				continue
			}

			if err := s.write(ctx, conn, &writeMu, bytes); err != nil {
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		core_logger.FromContext(ctx).Error(
			"read error",
			slog.String("addr", conn.RemoteAddr().String()),
			slog.Any("error", err),
		)
	}
}

func (s *Server) writeStream(
	ctx context.Context,
	conn net.Conn,
	mu *sync.Mutex,
	stream *features.ReplicationStream,
	done chan<- struct{},
) {
	defer close(done)

	for cmd := range stream.Commands() {
		if err := s.write(ctx, conn, mu, s.serializer.SerializeCommand(cmd)); err != nil {
			conn.Close()
			return
		}
	}
}

func (s *Server) write(
	ctx context.Context,
	conn net.Conn,
	mu *sync.Mutex,
	data []byte,
) error {
	mu.Lock()
	defer mu.Unlock()

	if err := conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout)); err != nil {
		return err
	}

	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			core_logger.FromContext(ctx).Error("write error", slog.Any("error", err))
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}

	return nil
}
