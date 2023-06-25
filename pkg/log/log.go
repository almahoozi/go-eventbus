package log

import (
	"context"

	"golang.org/x/exp/slog"
)

var (
	LogLevel = -5
	ErrLevel = -4
)

func Log(ctx context.Context, msg string, args ...any) {
	slog.Log(ctx, slog.Level(LogLevel), msg, args...)
}

func LogErr(ctx context.Context, msg string, args ...any) {
	slog.Log(ctx, slog.Level(ErrLevel), msg, args...)
}
