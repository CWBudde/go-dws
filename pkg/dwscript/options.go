package dwscript

import (
	"io"
	"os"

	"github.com/cwbudde/go-dws/internal/interp"
)

// Options configures the behavior of the DWScript engine.
type Options struct {
	Output            io.Writer
	TypeCheck         bool
	Trace             bool
	ExternalFunctions *interp.ExternalFunctionRegistry
	MaxRecursionDepth int // Maximum recursion depth (default: 1024)
}

// Option is a function that configures an Engine's Options.
type Option func(*Options) error

// defaultOptions returns the default options for a new engine.
func defaultOptions() Options {
	return Options{
		TypeCheck:         true,
		Output:            os.Stdout,
		Trace:             false,
		MaxRecursionDepth: 1024, // Default matches DWScript's cDefaultMaxRecursionDepth
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

// WithMaxRecursionDepth sets the maximum recursion depth for function calls.
// This prevents infinite recursion and stack overflow errors. When the call
// stack reaches this depth, the interpreter raises an EScriptStackOverflow exception.
//
// The default value is 1024, which matches DWScript's default limit.
//
// Example:
//
//	engine, err := dwscript.New(dwscript.WithMaxRecursionDepth(2048))
func WithMaxRecursionDepth(depth int) Option {
	return func(opts *Options) error {
		opts.MaxRecursionDepth = depth
		return nil
	}
}
