package log

import (
	"log/slog"
	"sync"
)

var (
	globalLogger Logger
	mu           sync.RWMutex
)

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(logger Logger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() Logger {
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger
}

// Global logging functions

func Debug(msg string, attrs ...slog.Attr) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Debug(msg, attrs...)
	}
}

func Info(msg string, attrs ...slog.Attr) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Info(msg, attrs...)
	}
}

func Warn(msg string, attrs ...slog.Attr) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Warn(msg, attrs...)
	}
}

func Error(msg string, attrs ...slog.Attr) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Error(msg, attrs...)
	}
}

func Fatal(msg string, attrs ...slog.Attr) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Fatal(msg, attrs...)
	}
}

func Debugf(format string, args ...any) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Debugf(format, args...)
	}
}

func Infof(format string, args ...any) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Infof(format, args...)
	}
}

func Warnf(format string, args ...any) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Warnf(format, args...)
	}
}

func Errorf(format string, args ...any) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Errorf(format, args...)
	}
}

func Fatalf(format string, args ...any) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Fatalf(format, args...)
	}
}

func With(attrs ...slog.Attr) Logger {
	if logger := GetGlobalLogger(); logger != nil {
		return logger.With(attrs...)
	}
	return nil
}

func WithGroup(name string) Logger {
	if logger := GetGlobalLogger(); logger != nil {
		return logger.WithGroup(name)
	}
	return nil
}
