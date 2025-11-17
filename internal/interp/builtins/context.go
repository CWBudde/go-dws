// Package builtins provides built-in function implementations for DWScript.
//
// This package is organized to avoid circular dependencies with the main
// interpreter package. Built-in functions are implemented as regular functions
// that take a Context interface, rather than methods on the Interpreter.
//
// The Context interface provides the minimal functionality that built-ins need:
// - Error reporting with location information
// - Access to the current AST node (for error messages)
//
// This allows both the legacy Interpreter and the new Evaluator to use the same
// built-in implementations by implementing the Context interface.
package builtins

import (
	"math/rand"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// Value represents a runtime value in the DWScript interpreter.
// This is aliased from the runtime package to avoid circular imports.
// All built-in functions work with Value types.
type Value = runtime.Value

// Context provides the minimal interface that built-in functions need
// to interact with the interpreter/evaluator.
//
// The Interpreter implements this interface to provide error reporting
// and AST node tracking for built-in functions.
//
// Design rationale:
// - Avoids circular dependency (builtins → interp → builtins)
// - Enables code reuse between Interpreter and Evaluator
// - Keeps built-in functions focused and testable
type Context interface {
	// NewError creates an error value with location information from the current node.
	// It formats the message using fmt.Sprintf semantics.
	NewError(format string, args ...interface{}) Value

	// CurrentNode returns the AST node currently being evaluated.
	// This is used for error reporting to provide source location context.
	CurrentNode() ast.Node

	// RandSource returns the random number generator for built-in functions
	// like Random(), RandomInt(), and RandG().
	RandSource() *rand.Rand

	// GetRandSeed returns the current random number generator seed value.
	// Used by the RandSeed() built-in function.
	GetRandSeed() int64

	// SetRandSeed sets the random number generator seed.
	// Used by the SetRandSeed() and Randomize() built-in functions.
	SetRandSeed(seed int64)
}

// BuiltinFunc is the signature for all built-in function implementations.
// Each built-in receives:
// - ctx: Context for error reporting and AST node access
// - args: Slice of argument values passed to the function
//
// Returns:
// - A Value result (may be an error value if the function fails)
type BuiltinFunc func(ctx Context, args []Value) Value
