package errors

import (
	"fmt"
)

// Wrap annotates err with a message and stack trace.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:   err,
		msg:     message,
		stack:   callers(),
		hasCode: false,
	}
}

// Wrapf annotates err with a formatted message and stack trace.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:   err,
		msg:     fmt.Sprintf(format, args...),
		stack:   callers(),
		hasCode: false,
	}
}

// WithMessage annotates err with a message (no stack trace).
// If err is nil, WithMessage returns nil.
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:   err,
		msg:     message,
		hasCode: false,
	}
}

// WithMessagef annotates err with a formatted message (no stack trace).
// If err is nil, WithMessagef returns nil.
func WithMessagef(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:   err,
		msg:     fmt.Sprintf(format, args...),
		hasCode: false,
	}
}

// withMessage is an error with additional context message.
type withMessage struct {
	cause   error
	msg     string
	stack   []uintptr
	hasCode bool
	code    int
}

func (w *withMessage) Error() string {
	if w.cause != nil {
		return w.msg + ": " + w.cause.Error()
	}
	return w.msg
}

func (w *withMessage) Unwrap() error {
	return w.cause
}

func (w *withMessage) StackTrace() []uintptr {
	return w.stack
}

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, w.Error())
			if len(w.stack) > 0 {
				formatStack(s, w.stack)
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

// Cause returns the root cause of the error, traversing the error chain.
// This is similar to errors.Unwrap but unwraps the entire chain.
func Cause(err error) error {
	for err != nil {
		unwrapped := Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
	return nil
}
