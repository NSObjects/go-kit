// Package utils provides common utility functions.
package utils

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

// ContextKey is the type for context keys (prevents collisions).
type ContextKey string

const (
	KeyTraceID   ContextKey = "trace_id"
	KeySpanID    ContextKey = "span_id"
	KeyRequestID ContextKey = "request_id"
	KeyUserID    ContextKey = "user_id"
	KeyStartTime ContextKey = "start_time"
)

// TraceContext contains trace information from a request.
type TraceContext struct {
	TraceID   string
	SpanID    string
	RequestID string
	UserID    string
	StartTime time.Time
}

// ExtractTraceContext extracts trace information from echo.Context.
func ExtractTraceContext(c echo.Context) *TraceContext {
	tc := &TraceContext{
		StartTime: time.Now(),
	}

	// Extract Request ID
	if requestID := c.Request().Header.Get("X-Request-ID"); requestID != "" {
		tc.RequestID = requestID
	} else if requestID = c.Response().Header().Get("X-Request-ID"); requestID != "" {
		tc.RequestID = requestID
	}

	// Extract TraceID and SpanID (OpenTelemetry format)
	if traceID := c.Request().Header.Get("X-Trace-ID"); traceID != "" {
		tc.TraceID = traceID
	}
	if spanID := c.Request().Header.Get("X-Span-ID"); spanID != "" {
		tc.SpanID = spanID
	}

	// Extract User ID (if authenticated)
	if userID := c.Get("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			tc.UserID = uid
		}
	}

	return tc
}

// BuildContext creates a context.Context with trace information from echo.Context.
func BuildContext(c echo.Context) context.Context {
	ctx := c.Request().Context()
	tc := ExtractTraceContext(c)

	ctx = context.WithValue(ctx, KeyTraceID, tc.TraceID)
	ctx = context.WithValue(ctx, KeySpanID, tc.SpanID)
	ctx = context.WithValue(ctx, KeyRequestID, tc.RequestID)
	ctx = context.WithValue(ctx, KeyUserID, tc.UserID)
	ctx = context.WithValue(ctx, KeyStartTime, tc.StartTime)

	return ctx
}

// GetTraceID returns TraceID from context.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(KeyTraceID).(string); ok {
		return v
	}
	return ""
}

// GetSpanID returns SpanID from context.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(KeySpanID).(string); ok {
		return v
	}
	return ""
}

// GetRequestID returns RequestID from context.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(KeyRequestID).(string); ok {
		return v
	}
	return ""
}

// GetUserID returns UserID from context.
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(KeyUserID).(string); ok {
		return v
	}
	return ""
}

// GetStartTime returns start time from context.
func GetStartTime(ctx context.Context) time.Time {
	if v, ok := ctx.Value(KeyStartTime).(time.Time); ok {
		return v
	}
	return time.Now()
}

// WithTraceInfo adds trace information to context.
func WithTraceInfo(ctx context.Context, traceID, spanID, requestID, userID string) context.Context {
	ctx = context.WithValue(ctx, KeyTraceID, traceID)
	ctx = context.WithValue(ctx, KeySpanID, spanID)
	ctx = context.WithValue(ctx, KeyRequestID, requestID)
	ctx = context.WithValue(ctx, KeyUserID, userID)
	ctx = context.WithValue(ctx, KeyStartTime, time.Now())
	return ctx
}
