// Package code provides business error codes with HTTP status mapping.
//
// This package defines business-specific error codes and provides convenient
// functions for creating and wrapping errors with these codes.
//
// Usage:
//
//	import (
//	    "github.com/NSObjects/go-kit/errors"  // for low-level operations
//	    "github.com/NSObjects/go-kit/code"    // for business errors
//	)
//
//	// Create business errors
//	err := code.NewNotFoundError("user")
//	err := code.WrapDatabaseError(dbErr, "query")
//
//	// Low-level operations (use errors package)
//	code := errors.GetCode(err)
//	httpStatus := errors.HTTPStatus(code)
package code

import (
	"fmt"

	"github.com/NSObjects/go-kit/errors"
)

// ========== Error Creation ==========

// NewError creates a new error with the specified code and message.
func NewError(code int, message string) error {
	return errors.WithCode(code, "%s", message)
}

// NewErrorf creates a new error with the specified code and formatted message.
func NewErrorf(code int, format string, args ...any) error {
	return errors.WithCode(code, format, args...)
}

// ========== Error Wrapping ==========

// WrapError wraps an error with a code.
func WrapError(err error, code int, message string) error {
	return wrapOrNew(err, code, message)
}

// WrapErrorf wraps an error with a code and formatted message.
func WrapErrorf(err error, code int, format string, args ...any) error {
	return wrapOrNewf(err, code, format, args...)
}

func wrapIfError(err error, code int, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return errors.WrapCode(err, code, format, args...)
}

func wrapOrNew(err error, code int, message string) error {
	if err == nil {
		return NewError(code, message)
	}
	return errors.WrapCode(err, code, "%s", message)
}

func wrapOrNewf(err error, code int, format string, args ...any) error {
	if err == nil {
		return NewErrorf(code, format, args...)
	}
	return errors.WrapCode(err, code, format, args...)
}

// ========== Infrastructure Error Wrappers ==========

// WrapDatabaseError wraps a database error.
func WrapDatabaseError(err error, operation string) error {
	return wrapIfError(err, ErrDatabase, "database %s failed", operation)
}

// WrapRedisError wraps a Redis error.
func WrapRedisError(err error, operation string) error {
	return wrapIfError(err, ErrRedis, "redis %s failed", operation)
}

// WrapKafkaError wraps a Kafka error.
func WrapKafkaError(err error, operation string) error {
	return wrapIfError(err, ErrKafka, "kafka %s failed", operation)
}

// WrapExternalError wraps an external service error.
func WrapExternalError(err error, service, operation string) error {
	return wrapIfError(err, ErrExternalService, "external service %s %s failed", service, operation)
}

// ========== HTTP Error Wrappers ==========

// WrapBadRequestError wraps a 400 error.
func WrapBadRequestError(err error, message string) error {
	return wrapOrNew(err, ErrBadRequest, message)
}

// WrapUnauthorizedError wraps a 401 error.
func WrapUnauthorizedError(err error, message string) error {
	return wrapOrNew(err, ErrUnauthorized, message)
}

// WrapForbiddenError wraps a 403 error.
func WrapForbiddenError(err error, message string) error {
	return wrapOrNew(err, ErrForbidden, message)
}

// WrapNotFoundError wraps a 404 error.
func WrapNotFoundError(err error, message string) error {
	return wrapOrNew(err, ErrNotFound, message)
}

// WrapInternalServerError wraps a 500 error.
func WrapInternalServerError(err error, message string) error {
	return wrapOrNew(err, ErrInternalServer, message)
}

// WrapBindError wraps a request binding error.
func WrapBindError(err error, message string) error {
	return wrapOrNew(err, ErrBind, message)
}

// WrapValidationError wraps a validation error.
func WrapValidationError(err error, message string) error {
	return wrapOrNew(err, ErrValidation, message)
}

// ========== Error Creators ==========

// NewValidationError creates a validation error.
func NewValidationError(field, message string) error {
	return NewErrorf(ErrValidation, "validation failed for field %s: %s", field, message)
}

// NewPermissionDeniedError creates a permission denied error.
func NewPermissionDeniedError(resource, action string) error {
	return NewErrorf(ErrPermissionDenied, "permission denied for %s on %s", action, resource)
}

// NewTokenInvalidError creates a token invalid error.
func NewTokenInvalidError() error {
	return NewError(ErrTokenInvalid, "token is invalid")
}

// NewTokenExpiredError creates a token expired error.
func NewTokenExpiredError() error {
	return NewError(ErrExpired, "token is expired")
}

// NewUnauthorizedError creates an unauthorized error.
func NewUnauthorizedError() error {
	return NewError(ErrUnauthorized, "unauthorized")
}

// NewForbiddenError creates a forbidden error.
func NewForbiddenError() error {
	return NewError(ErrForbidden, "forbidden")
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(resource string) error {
	return NewErrorf(ErrNotFound, "%s not found", resource)
}

// NewBadRequestError creates a bad request error.
func NewBadRequestError(message string) error {
	return NewErrorf(ErrBadRequest, "bad request: %s", message)
}

// ========== Error Classification ==========

// IsClientError returns true if the error code represents a client error (4xx).
func IsClientError(errCode int) bool {
	status := errors.HTTPStatus(errCode)
	return status >= 400 && status < 500
}

// IsServerError returns true if the error code represents a server error (5xx).
func IsServerError(errCode int) bool {
	status := errors.HTTPStatus(errCode)
	return status >= 500
}

// ErrorType represents the type of error.
type ErrorType int

const (
	InternalError ErrorType = iota
	BusinessError
)

// ErrorCategory represents the category of error.
type ErrorCategory string

const (
	CategoryDatabase   ErrorCategory = "database"
	CategoryRedis      ErrorCategory = "redis"
	CategoryKafka      ErrorCategory = "kafka"
	CategoryExternal   ErrorCategory = "external"
	CategorySystem     ErrorCategory = "system"
	CategoryAuth       ErrorCategory = "auth"
	CategoryPermission ErrorCategory = "permission"
	CategoryValidation ErrorCategory = "validation"
	CategoryBusiness   ErrorCategory = "business"
)

// ErrorInfo provides structured error information.
type ErrorInfo struct {
	Type     ErrorType     `json:"type"`
	Category ErrorCategory `json:"category"`
	Code     int           `json:"code"`
	Message  string        `json:"message"`
	Details  string        `json:"details,omitempty"`
}

// NewErrorInfo creates ErrorInfo from an error.
func NewErrorInfo(err error) ErrorInfo {
	if err == nil {
		return ErrorInfo{}
	}

	errCode := errors.GetCode(err)
	info := ErrorInfo{
		Type:     classifyErrorType(errCode),
		Category: classifyErrorCategory(errCode),
		Code:     errCode,
		Message:  err.Error(),
		Details:  fmt.Sprintf("%+v", err),
	}

	// Don't expose internal details for business errors
	if info.Type == BusinessError {
		info.Details = ""
	}

	return info
}

// IsInternal returns true if this is an internal error.
func (e *ErrorInfo) IsInternal() bool {
	return e.Type == InternalError
}

// IsBusiness returns true if this is a business error.
func (e *ErrorInfo) IsBusiness() bool {
	return e.Type == BusinessError
}

func classifyErrorType(errCode int) ErrorType {
	status := errors.HTTPStatus(errCode)
	if status >= 500 {
		return InternalError
	}
	return BusinessError
}

func classifyErrorCategory(errCode int) ErrorCategory {
	switch errCode {
	case ErrDatabase:
		return CategoryDatabase
	case ErrRedis:
		return CategoryRedis
	case ErrKafka:
		return CategoryKafka
	case ErrExternalService:
		return CategoryExternal
	case ErrValidation, ErrBind, ErrBadRequest:
		return CategoryValidation
	case ErrUnauthorized, ErrTokenInvalid, ErrExpired, ErrInvalidAuthHeader, ErrMissingHeader, ErrSignatureInvalid, ErrPasswordIncorrect:
		return CategoryAuth
	case ErrForbidden, ErrPermissionDenied, ErrAccountLocked, ErrAccountDisabled, ErrTooManyAttempts:
		return CategoryPermission
	default:
		if errCode >= 100300 && errCode < 100400 {
			return CategorySystem
		}
		return CategoryBusiness
	}
}
