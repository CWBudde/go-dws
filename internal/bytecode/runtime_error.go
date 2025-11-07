package bytecode

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
)

// RuntimeError represents an error that occurred while executing bytecode.
// It includes a stack trace for easier debugging.
type RuntimeError struct {
	Message string
	Trace   errors.StackTrace
}

// Error implements the error interface.
func (r *RuntimeError) Error() string {
	if r == nil {
		return "<nil>"
	}
	if len(r.Trace) == 0 {
		return r.Message
	}
	return fmt.Sprintf("%s\nStack trace:\n%s", r.Message, r.Trace.String())
}
