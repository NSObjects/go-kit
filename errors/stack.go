package errors

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

// callers returns a stack trace starting from the caller of callers.
func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}

// formatStack writes a formatted stack trace to the writer.
func formatStack(w io.Writer, stack []uintptr) {
	if len(stack) == 0 {
		return
	}

	frames := runtime.CallersFrames(stack)
	for {
		frame, more := frames.Next()
		// Skip runtime frames
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}

		fmt.Fprintf(w, "\n    %s\n        %s:%d", frame.Function, frame.File, frame.Line)

		if !more {
			break
		}
	}
}

// FormatStack returns a formatted stack trace as a string.
func FormatStack(stack []uintptr) string {
	var b strings.Builder
	formatStack(&b, stack)
	return b.String()
}

// StackTracer is implemented by errors that have a stack trace.
type StackTracer interface {
	StackTrace() []uintptr
}

// GetStackTrace extracts stack trace from an error if available.
func GetStackTrace(err error) []uintptr {
	var tracer StackTracer
	if As(err, &tracer) {
		return tracer.StackTrace()
	}
	return nil
}
