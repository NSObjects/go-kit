// Package errors provides error handling primitives for Go 1.21+.
//
// This is a modern reimplementation of error handling, designed for Go 1.21+.
// It leverages the standard library's errors package (errors.Is, errors.As,
// errors.Join, fmt.Errorf with %w) while providing error codes, stack traces,
// and a registration system for API error responses.
//
// Key features:
//   - Error codes with HTTP status mapping
//   - Stack traces for debugging
//   - Unwrap support for error chain traversal
//   - Compatible with errors.Is and errors.As
//
// Basic usage:
//
//	err := errors.New("something went wrong")
//	err = errors.Wrap(err, "context added")
//	err = errors.WithCode(ErrNotFound, "user %d not found", userId)
//
// Checking error codes:
//
//	if errors.IsCode(err, ErrNotFound) {
//	    // handle not found
//	}
package errors

import (
	"errors"
	"fmt"
)

// Re-export standard library functions for convenience
var (
	// Is reports whether any error in err's tree matches target.
	Is = errors.Is
	// As finds the first error in err's tree that matches target.
	As = errors.As
	// Join returns an error that wraps the given errors.
	Join = errors.Join
	// Unwrap returns the result of calling the Unwrap method on err.
	Unwrap = errors.Unwrap
)

// New returns an error with the supplied message and a stack trace.
func New(message string) error {
	return &fundamental{
		msg:   message,
		stack: callers(),
	}
}

// Errorf formats according to a format specifier and returns an error
// with a stack trace.
func Errorf(format string, args ...any) error {
	return &fundamental{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// fundamental is a basic error with message and stack.
type fundamental struct {
	msg   string
	stack []uintptr
}

func (f *fundamental) Error() string {
	return f.msg
}

// StackTrace returns the call stack where the error was created.
func (f *fundamental) StackTrace() []uintptr {
	return f.stack
}

// Format implements fmt.Formatter for rich error output.
func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, f.msg)
			formatStack(s, f.stack)
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, f.msg)
	case 'q':
		fmt.Fprintf(s, "%q", f.msg)
	}
}
