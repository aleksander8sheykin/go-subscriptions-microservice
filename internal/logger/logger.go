package logger

import (
	"log/slog"
	"os"
	"strings"
)

var Log *slog.Logger

func init() {
	InitLogger()
}

func InitLogger() {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	Log = slog.New(handler)
}
