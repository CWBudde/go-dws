package dwscript

import (
	"io"
	"os"

	"github.com/cwbudde/go-dws/internal/interp"
)

// CompileMode selects which execution engine the DWScript runtime uses.
type CompileMode int

const (
	// CompileModeAST executes programs using the existing AST interpreter.
	CompileModeAST CompileMode = iota
	// CompileModeBytecode compiles programs to bytecode and executes them on the VM.
	CompileModeBytecode
)

func (m CompileMode) String() string {
	switch m {
	case CompileModeAST:
		return "ast"
	case CompileModeBytecode:
		return "bytecode"
	default:
		return "unknown"
	}
}

// Options configures the behavior of the DWScript engine.
type Options struct {
	Output            io.Writer
	TypeCheck         bool
	Trace             bool
	ExternalFunctions *interp.ExternalFunctionRegistry
	MaxRecursionDepth int // Maximum recursion depth (default: 1024)
	CompileMode       CompileMode
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
		CompileMode:       CompileModeAST,
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

// WithCompileMode selects which execution engine should be used (AST or bytecode VM).
func WithCompileMode(mode CompileMode) Option {
	return func(opts *Options) error {
		opts.CompileMode = mode
		return nil
	}
}
