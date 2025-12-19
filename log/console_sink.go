package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// ConsoleSink outputs logs to console with formatting options.
type ConsoleSink struct {
	writer io.Writer
	format string // "json", "text", "color"
}

// ConsoleSinkConfig configuration for console output.
type ConsoleSinkConfig struct {
	Format string `json:"format" yaml:"format" toml:"format"` // json, text, color
	Output string `json:"output" yaml:"output" toml:"output"` // stdout, stderr
}

// NewConsoleSink creates a console sink.
func NewConsoleSink(cfg ConsoleSinkConfig) *ConsoleSink {
	format := cfg.Format
	if format == "" {
		format = "text"
	}

	writer := os.Stdout
	if cfg.Output == "stderr" {
		writer = os.Stderr
	}

	return &ConsoleSink{
		writer: writer,
		format: format,
	}
}

func (c *ConsoleSink) Write(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) error {
	switch c.format {
	case "json":
		return c.writeJSON(level, msg, attrs)
	case "text":
		return c.writeText(level, msg, attrs)
	case "color":
		return c.writeColor(level, msg, attrs)
	default:
		return c.writeText(level, msg, attrs)
	}
}

func (c *ConsoleSink) writeJSON(level slog.Level, msg string, attrs []slog.Attr) error {
	json := fmt.Sprintf(`{"time":"%s","level":"%s","msg":"%s"`,
		time.Now().Format(time.RFC3339),
		level.String(),
		msg)

	for _, attr := range attrs {
		json += fmt.Sprintf(`,"%s":"%v"`, attr.Key, attr.Value.Any())
	}
	json += "}\n"

	_, err := c.writer.Write([]byte(json))
	return err
}

func (c *ConsoleSink) writeText(level slog.Level, msg string, attrs []slog.Attr) error {
	text := fmt.Sprintf("%s %s %s",
		time.Now().Format("2006-01-02 15:04:05"),
		strings.ToUpper(level.String()),
		msg)

	for _, attr := range attrs {
		text += fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any())
	}
	text += "\n"

	_, err := c.writer.Write([]byte(text))
	return err
}

func (c *ConsoleSink) writeColor(level slog.Level, msg string, attrs []slog.Attr) error {
	// ANSI color codes
	var color string
	switch level {
	case slog.LevelDebug:
		color = "\033[36m" // Cyan
	case slog.LevelInfo:
		color = "\033[32m" // Green
	case slog.LevelWarn:
		color = "\033[33m" // Yellow
	case slog.LevelError:
		color = "\033[31m" // Red
	default:
		color = "\033[35m" // Magenta for Fatal
	}
	reset := "\033[0m"

	text := fmt.Sprintf("%s%s %s%s %s",
		color,
		time.Now().Format("15:04:05"),
		strings.ToUpper(level.String()),
		reset,
		msg)

	for _, attr := range attrs {
		text += fmt.Sprintf(" \033[2m%s=%v\033[0m", attr.Key, attr.Value.Any())
	}
	text += "\n"

	_, err := c.writer.Write([]byte(text))
	return err
}

func (c *ConsoleSink) Close() error {
	return nil
}
