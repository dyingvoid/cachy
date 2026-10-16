package core_logger

import (
	"context"
	"log/slog"
)

type contextKey struct{}

func ToContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(contextKey{}).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return log
}
