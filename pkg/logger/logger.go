package logger

import (
	"aloh-ssh/internal/config"
	"fmt"
	"io"
	"log/slog"
	"os"
)

const (
	localEnv = "local"
	devEnv   = "dev"
	prodEnv  = "prod"
	testEnv  = "test"
)

type Logger struct {
	Log *slog.Logger
}

func NewLogger(acfg config.App) *Logger {
	var log *slog.Logger
	env := acfg.Env
	writer := io.Writer(os.Stdout)
	if acfg.LogPath != "" {
		logFile, err := os.OpenFile(acfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(fmt.Errorf("faield to open log file: %w", err))
		}
		writer = io.MultiWriter(logFile, os.Stdout)
	}
	
	switch env {
	case localEnv:
		log = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case devEnv:
		log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case prodEnv:
		log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	logger := &Logger{
		Log: log.With(
			slog.String("type", "app"),
			slog.String("env", env),
			slog.String("app", acfg.Name),
			slog.String("version", acfg.Version),
		),
	}
	return logger
}

func (l *Logger) Info(msg string, args ...any) {
	l.Log.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Log.Error(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Log.Debug(msg, args...)
}

func (l *Logger) AddOp(op string) *Logger {
	logger := &Logger{
		Log: l.Log.With(slog.String("op", op)),
	}
	return logger
}

func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

func Attr(key string, val any) slog.Attr {
	return slog.Any(key, val)
}

func NewLogData(attrs ...any) []any {
	ld := make([]any, 0, len(attrs))
	ld = append(ld, attrs...)
	return ld
}
