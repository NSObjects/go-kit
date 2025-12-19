package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileSink outputs logs to a file with rotation support.
type FileSink struct {
	mu       sync.Mutex
	file     *os.File
	filename string
	maxSize  int64 // bytes
	curSize  int64
	format   string
}

// FileSinkConfig configuration for file output.
type FileSinkConfig struct {
	Filename   string `json:"filename" yaml:"filename" toml:"filename"`
	MaxSize    int    `json:"max_size" yaml:"max_size" toml:"max_size"`          // MB
	MaxBackups int    `json:"max_backups" yaml:"max_backups" toml:"max_backups"` // not implemented in simple version
	MaxAge     int    `json:"max_age" yaml:"max_age" toml:"max_age"`             // not implemented in simple version
	Compress   bool   `json:"compress" yaml:"compress" toml:"compress"`          // not implemented in simple version
	Format     string `json:"format" yaml:"format" toml:"format"`                // json, text
}

// NewFileSink creates a file sink.
func NewFileSink(cfg FileSinkConfig) *FileSink {
	format := cfg.Format
	if format == "" {
		format = "json"
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	file, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file: %v", err))
	}

	info, _ := file.Stat()
	var curSize int64
	if info != nil {
		curSize = info.Size()
	}

	maxSize := int64(cfg.MaxSize) * 1024 * 1024 // MB to bytes
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // Default 100MB
	}

	return &FileSink{
		file:     file,
		filename: cfg.Filename,
		maxSize:  maxSize,
		curSize:  curSize,
		format:   format,
	}
}

func (f *FileSink) Write(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var data []byte
	var err error

	switch f.format {
	case "json":
		data, err = f.formatJSON(level, msg, attrs)
	default:
		data, err = f.formatText(level, msg, attrs)
	}

	if err != nil {
		return err
	}

	// Check rotation
	if f.curSize+int64(len(data)) > f.maxSize {
		if err := f.rotate(); err != nil {
			return err
		}
	}

	n, err := f.file.Write(data)
	f.curSize += int64(n)
	return err
}

func (f *FileSink) formatJSON(level slog.Level, msg string, attrs []slog.Attr) ([]byte, error) {
	entry := map[string]any{
		"time":  time.Now().Format(time.RFC3339),
		"level": level.String(),
		"msg":   msg,
	}

	for _, attr := range attrs {
		entry[attr.Key] = attr.Value.Any()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func (f *FileSink) formatText(level slog.Level, msg string, attrs []slog.Attr) ([]byte, error) {
	text := fmt.Sprintf("%s %s %s",
		time.Now().Format("2006-01-02 15:04:05"),
		level.String(),
		msg)

	for _, attr := range attrs {
		text += fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any())
	}
	text += "\n"

	return []byte(text), nil
}

func (f *FileSink) rotate() error {
	if err := f.file.Close(); err != nil {
		return err
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	newName := fmt.Sprintf("%s.%s", f.filename, timestamp)
	if err := os.Rename(f.filename, newName); err != nil {
		return err
	}

	// Create new file
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	f.file = file
	f.curSize = 0
	return nil
}

func (f *FileSink) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}

var _ io.Closer = (*FileSink)(nil)
