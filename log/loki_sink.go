package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// LokiSink outputs logs to Grafana Loki.
type LokiSink struct {
	client  *http.Client
	url     string
	labels  map[string]string
	timeout time.Duration
}

// LokiSinkConfig configuration for Loki output.
type LokiSinkConfig struct {
	URL     string            `json:"url" yaml:"url" toml:"url"`
	Labels  map[string]string `json:"labels" yaml:"labels" toml:"labels"`
	Timeout time.Duration     `json:"timeout" yaml:"timeout" toml:"timeout"`
}

// NewLokiSink creates a Loki sink.
func NewLokiSink(cfg LokiSinkConfig) *LokiSink {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	labels := cfg.Labels
	if labels == nil {
		labels = map[string]string{
			"service": "app",
		}
	}

	return &LokiSink{
		client:  &http.Client{Timeout: timeout},
		url:     cfg.URL,
		labels:  labels,
		timeout: timeout,
	}
}

func (l *LokiSink) Write(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) error {
	// Build log entry
	entry := map[string]any{
		"level":   level.String(),
		"message": msg,
	}

	for _, attr := range attrs {
		entry[attr.Key] = attr.Value.Any()
	}

	entryJSON, _ := json.Marshal(entry)

	// Build Loki push API request
	lokiEntry := map[string]any{
		"stream": l.labels,
		"values": [][]string{
			{fmt.Sprintf("%d", time.Now().UnixNano()), string(entryJSON)},
		},
	}

	payload := map[string]any{
		"streams": []any{lokiEntry},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", l.url+"/loki/api/v1/push", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("loki request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (l *LokiSink) Close() error {
	return nil
}
