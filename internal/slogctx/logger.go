package slogctx

import (
	"context"
	"log/slog"
)

type ctxLoggerKey struct{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

func Ctx(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
