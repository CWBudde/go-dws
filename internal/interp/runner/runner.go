package runner

import (
	"io"

	"github.com/cwbudde/go-dws/internal/interp"
)

// New creates a new Interpreter with a fresh global environment.
func New(output io.Writer) *interp.Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions delegates runtime construction to the interpreter package.
func NewWithOptions(output io.Writer, opts interp.Options) *interp.Interpreter {
	return interp.NewWithOptions(output, opts)
}
