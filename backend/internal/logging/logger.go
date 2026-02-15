// Package logging for logs integration
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	defaultLevel = "info"
)

func New(appEnv string, level string, out io.Writer) *slog.Logger {
	if out == nil {
		out = os.Stdout
	}

	logLevel := parseLevel(level)
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: isDevEnv(appEnv),
	}

	var handler slog.Handler
	if isDevEnv(appEnv) {
		handler = slog.NewTextHandler(out, opts)
	} else {
		handler = slog.NewJSONHandler(out, opts)
	}

	return slog.New(handler).With("service", "warranty_days", "env", normalizeEnv(appEnv))
}

func parseLevel(level string) slog.Level {
	if level == "" {
		level = defaultLevel
	}

	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func isDevEnv(appEnv string) bool {
	env := normalizeEnv(appEnv)
	return env == "" || env == "dev" || env == "development" || env == "local"
}

func normalizeEnv(appEnv string) string {
	return strings.ToLower(strings.TrimSpace(appEnv))
}
