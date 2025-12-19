package code

import (
	"github.com/NSObjects/go-kit/errors"
)

//go:generate codegen -type=int
//go:generate codegen -type=int -doc -output ./error_code_generated.md

// Basic errors (100001-100099)
const (
	// ErrSuccess - 200: OK.
	ErrSuccess int = iota + 100001

	// ErrUnknown - 500: Internal server error.
	ErrUnknown

	// ErrBind - 400: Error occurred while binding the request body to the struct.
	ErrBind

	// ErrValidation - 400: Validation failed.
	ErrValidation

	// ErrTokenInvalid - 401: Token invalid.
	ErrTokenInvalid
)

// Database/Infrastructure errors (100101-100199)
const (
	// ErrDatabase - 500: Database error.
	ErrDatabase int = iota + 100101

	// ErrRedis - 500: Redis error.
	ErrRedis

	// ErrKafka - 500: Kafka error.
	ErrKafka

	// ErrExternalService - 500: External service error.
	ErrExternalService
)

// HTTP status code related errors (explicit values for clarity)
const (
	// ErrBadRequest - 400: Bad request.
	ErrBadRequest int = 100400
	// ErrUnauthorized - 401: Unauthorized.
	ErrUnauthorized int = 100401
	// ErrForbidden - 403: Forbidden.
	ErrForbidden int = 100403
	// ErrNotFound - 404: Not found.
	ErrNotFound int = 100404
	// ErrInternalServer - 500: Internal server error.
	ErrInternalServer int = 100500
)

// Authentication/Authorization errors (100201-100299)
const (
	// ErrEncrypt - 401: Error occurred while encrypting the user password.
	ErrEncrypt int = iota + 100201

	// ErrSignatureInvalid - 401: Signature is invalid.
	ErrSignatureInvalid

	// ErrExpired - 401: Token expired.
	ErrExpired

	// ErrInvalidAuthHeader - 401: Invalid authorization header.
	ErrInvalidAuthHeader

	// ErrMissingHeader - 401: The Authorization header was empty.
	ErrMissingHeader

	// ErrPasswordIncorrect - 401: Password was incorrect.
	ErrPasswordIncorrect

	// ErrPermissionDenied - 403: Permission denied.
	ErrPermissionDenied

	// ErrAccountLocked - 403: Account is locked.
	ErrAccountLocked

	// ErrAccountDisabled - 403: Account is disabled.
	ErrAccountDisabled

	// ErrTooManyAttempts - 403: Too many login attempts.
	ErrTooManyAttempts
)

// Encoding/Decoding errors (100301-100399)
const (
	// ErrEncodingFailed - 500: Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301

	// ErrDecodingFailed - 500: Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON - 500: Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON - 500: JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON - 500: JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml - 500: Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml - 500: Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml - 500: Yaml data could not be decoded.
	ErrDecodingYaml
)

func init() {
	// Register basic errors
	errors.Register(ErrSuccess, 200, "OK")
	errors.Register(ErrUnknown, 500, "Internal server error")
	errors.Register(ErrBind, 400, "Error binding request")
	errors.Register(ErrValidation, 400, "Validation failed")
	errors.Register(ErrTokenInvalid, 401, "Token invalid")

	// Register database errors
	errors.Register(ErrDatabase, 500, "Database error")
	errors.Register(ErrRedis, 500, "Redis error")
	errors.Register(ErrKafka, 500, "Kafka error")
	errors.Register(ErrExternalService, 500, "External service error")

	// Register HTTP status errors
	errors.Register(ErrBadRequest, 400, "Bad request")
	errors.Register(ErrUnauthorized, 401, "Unauthorized")
	errors.Register(ErrForbidden, 403, "Forbidden")
	errors.Register(ErrNotFound, 404, "Not found")
	errors.Register(ErrInternalServer, 500, "Internal server error")

	// Register auth errors
	errors.Register(ErrEncrypt, 401, "Encryption failed")
	errors.Register(ErrSignatureInvalid, 401, "Signature is invalid")
	errors.Register(ErrExpired, 401, "Token expired")
	errors.Register(ErrInvalidAuthHeader, 401, "Invalid authorization header")
	errors.Register(ErrMissingHeader, 401, "Authorization header missing")
	errors.Register(ErrPasswordIncorrect, 401, "Password incorrect")
	errors.Register(ErrPermissionDenied, 403, "Permission denied")
	errors.Register(ErrAccountLocked, 403, "Account locked")
	errors.Register(ErrAccountDisabled, 403, "Account disabled")
	errors.Register(ErrTooManyAttempts, 403, "Too many attempts")

	// Register encoding errors
	errors.Register(ErrEncodingFailed, 500, "Encoding failed")
	errors.Register(ErrDecodingFailed, 500, "Decoding failed")
	errors.Register(ErrInvalidJSON, 500, "Invalid JSON")
	errors.Register(ErrEncodingJSON, 500, "JSON encoding failed")
	errors.Register(ErrDecodingJSON, 500, "JSON decoding failed")
	errors.Register(ErrInvalidYaml, 500, "Invalid YAML")
	errors.Register(ErrEncodingYaml, 500, "YAML encoding failed")
	errors.Register(ErrDecodingYaml, 500, "YAML decoding failed")
}
