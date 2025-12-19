package errors

import (
	"fmt"
	"net/http"
	"sync"
)

// Coder defines an interface for error codes.
type Coder interface {
	// Code returns the error code.
	Code() int
	// HTTPStatus returns the HTTP status code.
	HTTPStatus() int
	// Message returns the user-facing message.
	Message() string
}

// coder is the default implementation of Coder.
type coder struct {
	code       int
	httpStatus int
	message    string
}

func (c coder) Code() int       { return c.code }
func (c coder) HTTPStatus() int { return c.httpStatus }
func (c coder) Message() string { return c.message }

var (
	registry   = make(map[int]coder)
	registryMu sync.RWMutex

	unknownCoder = coder{
		code:       1,
		httpStatus: http.StatusInternalServerError,
		message:    "Internal server error",
	}
)

// Register registers an error code with its HTTP status and message.
// Panics if the code is 0 or already registered.
func Register(code int, httpStatus int, message string) {
	if code == 0 {
		panic("error code 0 is reserved")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[code]; exists {
		panic(fmt.Sprintf("error code %d already registered", code))
	}

	registry[code] = coder{
		code:       code,
		httpStatus: httpStatus,
		message:    message,
	}
}

// MustRegister is like Register but allows overwriting existing codes.
func MustRegister(code int, httpStatus int, message string) {
	if code == 0 {
		panic("error code 0 is reserved")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	registry[code] = coder{
		code:       code,
		httpStatus: httpStatus,
		message:    message,
	}
}

// Lookup retrieves a Coder by code.
func Lookup(code int) (Coder, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	c, ok := registry[code]
	return c, ok
}

// HTTPStatus returns the HTTP status for an error code.
// Returns 500 if code is not registered.
func HTTPStatus(code int) int {
	registryMu.RLock()
	defer registryMu.RUnlock()

	if c, ok := registry[code]; ok {
		return c.httpStatus
	}
	return http.StatusInternalServerError
}

// codeError is an error with code.
type codeError struct {
	code    int
	message string
	cause   error
	stack   []uintptr
}

func (e *codeError) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e *codeError) Unwrap() error {
	return e.cause
}

func (e *codeError) Code() int {
	return e.code
}

func (e *codeError) StackTrace() []uintptr {
	return e.stack
}

func (e *codeError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "[%d] %s", e.code, e.message)
			if e.cause != nil {
				fmt.Fprintf(s, ": %+v", e.cause)
			}
			if len(e.stack) > 0 {
				formatStack(s, e.stack)
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// WithCode creates a new error with the given code.
func WithCode(code int, format string, args ...any) error {
	return &codeError{
		code:    code,
		message: fmt.Sprintf(format, args...),
		stack:   callers(),
	}
}

// WrapCode wraps an error with a code.
// If err is nil, returns nil.
func WrapCode(err error, code int, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &codeError{
		code:    code,
		message: fmt.Sprintf(format, args...),
		cause:   err,
		stack:   callers(),
	}
}

// CodedError is implemented by errors that have a code.
type CodedError interface {
	Code() int
}

// GetCode extracts the error code from an error.
// Returns 0 if no code is found.
func GetCode(err error) int {
	var coded CodedError
	if As(err, &coded) {
		return coded.Code()
	}
	return 0
}

// IsCode reports whether any error in err's chain has the given code.
func IsCode(err error, code int) bool {
	for err != nil {
		if coded, ok := err.(CodedError); ok {
			if coded.Code() == code {
				return true
			}
		}
		err = Unwrap(err)
	}
	return false
}

// ParseCoder extracts Coder from an error.
// Returns unknownCoder if no code is found.
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	code := GetCode(err)
	if code == 0 {
		return unknownCoder
	}

	if c, ok := Lookup(code); ok {
		return c
	}

	return unknownCoder
}
