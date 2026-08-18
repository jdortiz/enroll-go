package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"enroll/app"
	"enroll/token"

	"github.com/joho/godotenv"
)

var version = "undefined"

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn("could not load .env file", "error", err)
	}

	fmt.Printf("KommKorp enrollment application v.%s\n", version)

	opts := &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	}
	logHandler := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	slog.SetDefault(logHandler)

	encodingKey, err := token.GetEncodingKey()
	if err != nil {
		slog.Error("failed to load encoding key", "error", err)
		os.Exit(1)
	}
	users := seedUsers()
	a := app.New(users, encodingKey)
	slog.Error("Application error: HTTP server", "error", a.Start())
}

func seedUsers() *app.UserInMemoryRepo {
	users := app.NewUserInMemoryRepo()
	verifyPhone := os.Getenv("VERIFY_PHONE")
	if len(verifyPhone) == 0 {
		verifyPhone = "34600123456"
	}
	user, err := app.NewUser("alice", "secret123", verifyPhone)
	if err != nil {
		slog.Error("failed to create seed user", "error", err)
		os.Exit(1)
	}
	if err := users.Add(user); err != nil {
		slog.Error("failed to seed user", "error", err)
		os.Exit(1)
	}
	return users
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
