// Package metrics provides Prometheus metrics collection.
package metrics

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds application metrics.
type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestSize     *prometheus.SummaryVec
	ResponseSize    *prometheus.SummaryVec
}

// New creates and registers default metrics.
func New(namespace string) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		RequestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
			},
			[]string{"method", "path"},
		),
		ResponseSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
			},
			[]string{"method", "path"},
		),
	}

	prometheus.MustRegister(m.RequestsTotal)
	prometheus.MustRegister(m.RequestDuration)
	prometheus.MustRegister(m.RequestSize)
	prometheus.MustRegister(m.ResponseSize)

	return m
}

// Handler returns the Prometheus metrics handler.
func Handler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
