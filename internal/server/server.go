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

	"github.com/dyingvoid/cachy/internal/domain"
	"github.com/dyingvoid/cachy/internal/logger"
)

type Server struct {
	cfg Config
	log *slog.Logger

	ln net.Listener
	wg sync.WaitGroup

	parser   Parser
	executor Executor
}

type Parser interface {
	Parse(in string) (domain.Command, error)
}

type Executor interface {
	Execute(domain.Command) (domain.Result, error)
}

func NewServer(
	cfg Config,
	log *slog.Logger,
	parser Parser,
	executor Executor,
) *Server {
	return &Server{
		cfg:      cfg,
		log:      log,
		parser:   parser,
		executor: executor,
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
	defer conn.Close()

	log := slog.With(slog.String("addr", conn.RemoteAddr().String()))
	log.Info("New connection")

	ctx = logger.ToContext(ctx, log)

	scanner := bufio.NewScanner(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))
		if !scanner.Scan() {
			log.Info("break", slog.String("addr", conn.RemoteAddr().String()))
			break
		}

		cmd, err := s.parser.Parse(scanner.Text())
		if err != nil {
			if err := s.writeError(ctx, conn, err); err != nil {
				return
			}
			continue
		}

		res, err := s.executor.Execute(cmd)
		if err != nil {
			if err := s.writeError(ctx, conn, err); err != nil {
				return
			}
			continue
		}

		if err := s.writeRes(ctx, conn, res); err != nil {
			return
		}

		conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
	}

	if err := scanner.Err(); err != nil {
		logger.FromContext(ctx).Error(
			"read error",
			slog.String("addr", conn.RemoteAddr().String()),
			slog.Any("error", err),
		)
	}
}

func (s *Server) writeError(ctx context.Context, conn net.Conn, err error) error {
	conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
	_, werr := fmt.Fprintf(conn, "ERR %s\n", err)
	if werr != nil {
		logger.FromContext(ctx).Error("write error", slog.Any("error", werr))
		return werr
	}

	return nil
}

func (s *Server) writeRes(ctx context.Context, conn net.Conn, res domain.Result) error {
	conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
	_, werr := fmt.Fprintf(conn, "%s\n", res)
	if werr != nil {
		logger.FromContext(ctx).Error("write error", slog.Any("error", werr))
		return werr
	}

	return nil
}
