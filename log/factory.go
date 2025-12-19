package log

import (
	"log/slog"
	"strings"

	"github.com/NSObjects/go-kit/config"
)

// LogConfig extended configuration for logging.
type LogConfig struct {
	Level  string `json:"level" yaml:"level" toml:"level"`
	Format string `json:"format" yaml:"format" toml:"format"`

	Console       ConsoleSinkConfig       `json:"console" yaml:"console" toml:"console"`
	File          FileSinkConfig          `json:"file" yaml:"file" toml:"file"`
	Elasticsearch ElasticsearchSinkConfig `json:"elasticsearch" yaml:"elasticsearch" toml:"elasticsearch"`
	Loki          LokiSinkConfig          `json:"loki" yaml:"loki" toml:"loki"`
}

// New creates a logger from the base configuration.
func New(cfg config.LogConfig) Logger {
	level := parseLevel(cfg.Level)

	// Default: console sink with text format
	sink := NewConsoleSink(ConsoleSinkConfig{
		Format: cfg.Format,
		Output: cfg.Output,
	})

	return NewDefaultLogger(sink, level)
}

// NewFromLogConfig creates a logger from extended config with multiple sinks.
func NewFromLogConfig(cfg LogConfig, env string) Logger {
	level := parseLevel(cfg.Level)

	var sinks []Sink

	// Console sink (always enabled by default)
	if cfg.Console.Format != "" || cfg.Console.Output != "" {
		sinks = append(sinks, NewConsoleSink(cfg.Console))
	} else {
		// Default based on environment
		format := "color"
		if env == "prod" {
			format = "json"
		}
		sinks = append(sinks, NewConsoleSink(ConsoleSinkConfig{
			Format: format,
			Output: "stdout",
		}))
	}

	// File sink (production/test)
	if cfg.File.Filename != "" && (env == "prod" || env == "test") {
		sinks = append(sinks, NewFileSink(cfg.File))
	}

	// Elasticsearch sink
	if cfg.Elasticsearch.URL != "" {
		sinks = append(sinks, NewElasticsearchSink(cfg.Elasticsearch))
	}

	// Loki sink
	if cfg.Loki.URL != "" {
		sinks = append(sinks, NewLokiSink(cfg.Loki))
	}

	// Create sink
	var sink Sink
	if len(sinks) == 1 {
		sink = sinks[0]
	} else {
		sink = NewMultiSink(sinks...)
	}

	// Create logger and set as global
	logger := NewDefaultLogger(sink, level)
	SetGlobalLogger(logger)

	return logger
}

// parseLevel parses a string level to slog.Level.
func parseLevel(level string) slog.Level {
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
		return slog.LevelInfo
	}
}
