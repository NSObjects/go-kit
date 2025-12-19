// Package middleware provides Echo middleware implementations.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/NSObjects/go-kit/code"
	"github.com/NSObjects/go-kit/errors"
	"github.com/NSObjects/go-kit/resp"
	"github.com/labstack/echo/v4"
)

// ErrorHandler is the centralized error handler for Echo.
func ErrorHandler(err error, c echo.Context) {
	start := time.Now()

	// Handle different error types
	switch e := err.(type) {
	case *echo.HTTPError:
		handleHTTPError(e, c)
	case *ValidationError:
		handleValidationError(e, c)
	default:
		handleGenericError(err, c)
	}

	// Log handling duration
	slog.Debug("Error handled",
		slog.Duration("duration", time.Since(start)),
		slog.String("method", c.Request().Method),
		slog.String("uri", c.Request().RequestURI),
	)
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// handleHTTPError converts Echo HTTP errors to business errors.
func handleHTTPError(err *echo.HTTPError, c echo.Context) {
	message := extractErrorMessage(err.Message)

	var bizErr error
	switch err.Code {
	case http.StatusBadRequest:
		bizErr = code.NewErrorf(code.ErrBadRequest, "%s", message)
	case http.StatusUnauthorized:
		bizErr = code.NewErrorf(code.ErrUnauthorized, "%s", message)
	case http.StatusForbidden:
		bizErr = code.NewErrorf(code.ErrForbidden, "%s", message)
	case http.StatusNotFound:
		bizErr = code.NewErrorf(code.ErrNotFound, "%s", message)
	default:
		bizErr = code.NewErrorf(code.ErrInternalServer, "%s", message)
	}

	_ = resp.APIError(c, bizErr)
}

// handleValidationError handles validation errors.
func handleValidationError(err *ValidationError, c echo.Context) {
	slog.Warn("Validation Error",
		slog.String("field", err.Field),
		slog.String("message", err.Message),
		slog.Any("value", err.Value),
	)

	bizErr := code.NewValidationError(err.Field, err.Message)
	_ = resp.APIError(c, bizErr)
}

// handleGenericError handles generic errors.
func handleGenericError(err error, c echo.Context) {
	// Check if error has a code
	if errors.GetCode(err) != 0 {
		_ = resp.APIError(c, err)
		return
	}

	// Log unknown errors
	slog.Error("Generic Error",
		slog.String("error", err.Error()),
		slog.String("method", c.Request().Method),
		slog.String("uri", c.Request().RequestURI),
	)

	wrapped := code.WrapInternalServerError(err, "internal server error")
	_ = resp.APIError(c, wrapped)
}

// extractErrorMessage converts various types to string.
func extractErrorMessage(message any) string {
	switch v := message.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprint(v)
	}
}

// Recovery returns a panic recovery middleware.
func Recovery() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Panic recovered",
						slog.Any("panic", r),
						slog.String("method", c.Request().Method),
						slog.String("uri", c.Request().RequestURI),
					)

					err := code.NewError(code.ErrInternalServer, "internal server error")
					_ = resp.APIError(c, err)
				}
			}()

			return next(c)
		}
	}
}

// RequestLogger returns a request logging middleware.
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			slog.Info("Request",
				slog.String("method", c.Request().Method),
				slog.String("uri", c.Request().RequestURI),
				slog.Int("status", c.Response().Status),
				slog.Duration("latency", time.Since(start)),
			)

			return err
		}
	}
}
