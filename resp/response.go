// Package resp provides unified API response formatting.
package resp

import (
	"log/slog"
	"net/http"

	"github.com/NSObjects/go-kit/code"
	"github.com/NSObjects/go-kit/errors"
	"github.com/labstack/echo/v4"
)

// Response is the unified API response structure.
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// ListResponse is the response structure for list endpoints.
type ListResponse struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
}

// SuccessJSON returns a success response with custom data.
func SuccessJSON(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// ListDataResponse returns a paginated list response.
func ListDataResponse(c echo.Context, list any, total int64) error {
	return SuccessJSON(c, ListResponse{
		List:  list,
		Total: total,
	})
}

// OneDataResponse returns a single item response.
func OneDataResponse(c echo.Context, data any) error {
	return SuccessJSON(c, data)
}

// OperateSuccess returns a success response for operations.
func OperateSuccess(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
	})
}

// APIError returns an error response.
func APIError(c echo.Context, err error) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Extract error code
	errorCode := errors.GetCode(err)
	if errorCode == 0 {
		errorCode = code.ErrInternalServer
	}

	httpStatus := errors.HTTPStatus(errorCode)
	message := err.Error()

	// Try to get registered message
	if coder, ok := errors.Lookup(errorCode); ok {
		message = coder.Message()
	}

	// Log the error
	logError(c, err, errorCode, message, requestID)

	return c.JSON(httpStatus, Response{
		Code: errorCode,
		Msg:  message,
	})
}

// logError logs an error with context.
func logError(c echo.Context, err error, errorCode int, message, requestID string) {
	logFields := []any{
		slog.String("request_id", requestID),
		slog.String("method", c.Request().Method),
		slog.String("uri", c.Request().RequestURI),
		slog.Int("code", errorCode),
		slog.String("message", message),
	}

	if code.IsServerError(errorCode) {
		slog.Error("Server error", append(logFields, slog.String("error", err.Error()))...)
	} else {
		slog.Warn("Client error", logFields...)
	}
}
