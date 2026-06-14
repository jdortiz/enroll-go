package main

import (
	"log/slog"
	"os"
	"strings"

	"enroll/app"
)

func main() {
	opts := &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	}
	logHandler := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logHandler)

	app := app.New()
	app.Start()
}

// Returs the slog.Level configured in the environment variable.
func getLogLevelFromEnv() slog.Level {
	levelStr := os.Getenv("GO_LOG")
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn // Errors and warnings are displayed

	}
}
