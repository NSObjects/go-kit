// Package utils provides common utility functions.
package utils

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

// ContextKey is the type for context keys (prevents collisions).
type ContextKey string

const (
	// KeyRequestID is the context key for request ID.
	KeyRequestID ContextKey = "request_id"
	// KeyUserID is the context key for user ID.
	KeyUserID ContextKey = "user_id"
	// KeyStartTime is the context key for request start time.
	KeyStartTime ContextKey = "start_time"
)

// TraceContext contains trace and request information from a request.
// TraceID and SpanID are extracted from OpenTelemetry context,
// RequestID is extracted from Echo middleware.
type TraceContext struct {
	TraceID   string    // TraceID from OpenTelemetry span context
	SpanID    string    // SpanID from OpenTelemetry span context
	RequestID string    // RequestID from Echo middleware (middleware.RequestID)
	UserID    string    // UserID from authenticated user (if available)
	StartTime time.Time // Request start time
}

// ExtractTraceContext extracts trace and request information from echo.Context.
//
// This function extracts:
//   - TraceID and SpanID from OpenTelemetry span context (requires otelecho middleware)
//   - RequestID from Echo middleware (requires middleware.RequestID)
//   - UserID from echo.Context if authenticated
//   - StartTime as current time
func ExtractTraceContext(c echo.Context) *TraceContext {
	tc := &TraceContext{
		StartTime: time.Now(),
	}

	// Extract RequestID from Echo middleware
	// Requires: e.Use(middleware.RequestID())
	tc.RequestID = c.Response().Header().Get(echo.HeaderXRequestID)

	// Extract TraceID and SpanID from OpenTelemetry context
	// Requires: e.Use(otelecho.Middleware("service-name"))
	ctx := c.Request().Context()
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		if spanContext := span.SpanContext(); spanContext.IsValid() {
			tc.TraceID = spanContext.TraceID().String()
			tc.SpanID = spanContext.SpanID().String()
		}
	}

	// Extract UserID from echo.Context (set by authentication middleware)
	if userID := c.Get("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			tc.UserID = uid
		}
	}

	return tc
}

// BuildContext creates a context.Context with request information from echo.Context.
//
// This function:
//   - Preserves the original context (including OpenTelemetry span context)
//   - Adds business-related values (RequestID, UserID, StartTime) to context
//   - Does NOT inject TraceID/SpanID as they are already available via OpenTelemetry
//
// Usage:
//
//	func handler(c echo.Context) error {
//	    ctx := utils.BuildContext(c)
//	    requestID := utils.GetRequestID(ctx)
//	    traceID := utils.GetTraceID(ctx) // From OpenTelemetry
//	    // ...
//	}
func BuildContext(c echo.Context) context.Context {
	ctx := c.Request().Context()
	tc := ExtractTraceContext(c)

	ctx = context.WithValue(ctx, KeyRequestID, tc.RequestID)
	ctx = context.WithValue(ctx, KeyUserID, tc.UserID)
	ctx = context.WithValue(ctx, KeyStartTime, tc.StartTime)

	return ctx
}

// GetTraceID returns TraceID from context.
// It extracts from OpenTelemetry span context if available.
//
// Returns empty string if no valid span context is found.
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ""
	}
	spanContext := span.SpanContext()
	if !spanContext.IsValid() {
		return ""
	}
	return spanContext.TraceID().String()
}

// GetSpanID returns SpanID from context.
// It extracts from OpenTelemetry span context if available.
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ""
	}
	spanContext := span.SpanContext()
	if !spanContext.IsValid() {
		return ""
	}
	return spanContext.SpanID().String()
}

// GetRequestID returns RequestID from context.
// The RequestID is set by BuildContext() which extracts it from Echo middleware.
// Returns empty string if not found.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(KeyRequestID).(string); ok {
		return v
	}
	return ""
}

// GetRequestIDFromEcho is a convenience function to get RequestID directly from echo.Context.
// This is useful when you don't have a context.Context but have echo.Context.
func GetRequestIDFromEcho(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}

// GetUserID returns UserID from context.
// The UserID should be set by BuildContext() which extracts it from echo.Context.
//
// Returns empty string if not found in context.
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(KeyUserID).(string); ok {
		return v
	}
	return ""
}

// GetStartTime returns start time from context.
// The StartTime should be set by BuildContext().
//
// Returns current time if not found in context.
func GetStartTime(ctx context.Context) time.Time {
	if v, ok := ctx.Value(KeyStartTime).(time.Time); ok {
		return v
	}
	return time.Now()
}

// WithRequestInfo adds request information to context.
// This is useful for testing or when you need to manually set request context.
//
// Note: TraceID and SpanID should be managed by OpenTelemetry, not manually set.
func WithRequestInfo(ctx context.Context, requestID, userID string) context.Context {
	ctx = context.WithValue(ctx, KeyRequestID, requestID)
	ctx = context.WithValue(ctx, KeyUserID, userID)
	ctx = context.WithValue(ctx, KeyStartTime, time.Now())
	return ctx
}
