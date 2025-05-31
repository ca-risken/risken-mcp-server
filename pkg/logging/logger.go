package logging

import (
	"log/slog"
	"os"
)

func NewStdioLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{
			Level: level,
		},
	))
}

func NewHTTPLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	))
}
