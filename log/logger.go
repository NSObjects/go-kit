// Package log provides structured logging with multiple sink support.
package log

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Logger is the unified logging interface.
type Logger interface {
	Debug(msg string, attrs ...slog.Attr)
	Info(msg string, attrs ...slog.Attr)
	Warn(msg string, attrs ...slog.Attr)
	Error(msg string, attrs ...slog.Attr)
	Fatal(msg string, attrs ...slog.Attr)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)

	With(attrs ...slog.Attr) Logger
	WithGroup(name string) Logger
}

// Sink is the log output target abstraction.
type Sink interface {
	Write(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) error
	Close() error
}

// MultiSink writes to multiple sinks.
type MultiSink struct {
	sinks []Sink
}

// NewMultiSink creates a sink that writes to multiple targets.
func NewMultiSink(sinks ...Sink) *MultiSink {
	return &MultiSink{sinks: sinks}
}

func (m *MultiSink) Write(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) error {
	for _, sink := range m.sinks {
		if err := sink.Write(ctx, level, msg, attrs); err != nil {
			continue // Log error but don't stop other sinks
		}
	}
	return nil
}

func (m *MultiSink) Close() error {
	for _, sink := range m.sinks {
		sink.Close()
	}
	return nil
}

// DefaultLogger is the default logger implementation.
type DefaultLogger struct {
	slog *slog.Logger
	sink Sink
	mu   sync.RWMutex
}

// NewDefaultLogger creates a logger with the given sink and level.
func NewDefaultLogger(sink Sink, level slog.Level) *DefaultLogger {
	handler := &SinkHandler{sink: sink, level: level}
	return &DefaultLogger{
		slog: slog.New(handler),
		sink: sink,
	}
}

func (l *DefaultLogger) Debug(msg string, attrs ...slog.Attr) {
	l.slog.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

func (l *DefaultLogger) Info(msg string, attrs ...slog.Attr) {
	l.slog.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

func (l *DefaultLogger) Warn(msg string, attrs ...slog.Attr) {
	l.slog.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

func (l *DefaultLogger) Error(msg string, attrs ...slog.Attr) {
	l.slog.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

func (l *DefaultLogger) Fatal(msg string, attrs ...slog.Attr) {
	l.slog.LogAttrs(context.Background(), slog.LevelError+1, msg, attrs...)
}

func (l *DefaultLogger) Debugf(format string, args ...any) {
	l.slog.Debug(fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Infof(format string, args ...any) {
	l.slog.Info(fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Warnf(format string, args ...any) {
	l.slog.Warn(fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Errorf(format string, args ...any) {
	l.slog.Error(fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) Fatalf(format string, args ...any) {
	l.slog.Log(context.Background(), slog.LevelError+1, fmt.Sprintf(format, args...))
}

func (l *DefaultLogger) With(attrs ...slog.Attr) Logger {
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return &DefaultLogger{
		slog: l.slog.With(args...),
		sink: l.sink,
	}
}

func (l *DefaultLogger) WithGroup(name string) Logger {
	return &DefaultLogger{
		slog: l.slog.WithGroup(name),
		sink: l.sink,
	}
}

// SinkHandler implements slog.Handler.
type SinkHandler struct {
	sink  Sink
	level slog.Level
}

func (h *SinkHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SinkHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	return h.sink.Write(ctx, r.Level, r.Message, attrs)
}

func (h *SinkHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SinkHandler{sink: h.sink, level: h.level}
}

func (h *SinkHandler) WithGroup(name string) slog.Handler {
	return &SinkHandler{sink: h.sink, level: h.level}
}
