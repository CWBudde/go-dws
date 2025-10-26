package dwscript

import (
	"io"
	"os"
)

// Options configures the behavior of the DWScript engine.
type Options struct {
	Output    io.Writer
	TypeCheck bool
	Trace     bool
}

// Option is a function that configures an Engine's Options.
type Option func(*Options) error

// defaultOptions returns the default options for a new engine.
func defaultOptions() Options {
	return Options{
		TypeCheck: true,
		Output:    os.Stdout,
		Trace:     false,
	}
}

// WithTypeCheck enables or disables type checking.
//
// Example:
//
//	engine, err := dwscript.New(dwscript.WithTypeCheck(false))
func WithTypeCheck(enabled bool) Option {
	return func(opts *Options) error {
		opts.TypeCheck = enabled
		return nil
	}
}

// WithOutput sets the output writer for program output.
//
// Example:
//
//	var buf bytes.Buffer
//	engine, err := dwscript.New(dwscript.WithOutput(&buf))
func WithOutput(w io.Writer) Option {
	return func(opts *Options) error {
		opts.Output = w
		return nil
	}
}

// WithTrace enables or disables execution tracing.
//
// Example:
//
//	engine, err := dwscript.New(dwscript.WithTrace(true))
func WithTrace(enabled bool) Option {
	return func(opts *Options) error {
		opts.Trace = enabled
		return nil
	}
}
