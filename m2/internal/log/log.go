package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	logger *slog.Logger
)

func Error(msg string, args ...any) {
	log(slog.LevelError, msg, args...)
}

func Info(msg string, args ...any) {
	log(slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	log(slog.LevelDebug, msg, args...)
}

// log() implementation from:
// https://cs.opensource.google/go/go/+/master:src/log/slog/logger.go;drc=691af6ca28dad9c72e51346fe10c6aaadc3f940b;l=240
// Accessed Aug 13, 2025
func log(level slog.Level, msg string, args ...any) {
	ctx := context.Background()
	l := logger

	if !l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}

func InitTerminal(logLevel string) {
	initLogger(os.Stdout, logLevel, false)
}

func InitFile(logFilePath, logLevel string) (closeLogFile func()) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic("Could not open log file: " + err.Error())
	}
	initLogger(logFile, logLevel, true)

	closeLogFile = func() {
		Debug("Closing log", "path", logFilePath)
		logFile.Close()
	}
	return
}

func initLogger(writer io.Writer, logLevel string, json bool) {
	level := parseLogLevel(logLevel)
	options := &slog.HandlerOptions{
		Level: level,
	}
	if json {
		options.AddSource = true
		logger = slog.New(slog.NewJSONHandler(writer, options))
	} else {
		logger = slog.New(slog.NewTextHandler(writer, options))
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelError
	}
}
