package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// DEPRECATED: This file contains the old LazyThunk implementation with direct
// Interpreter dependency. It has been replaced by runtime.LazyThunk
// which uses a callback-based pattern to break the dependency.
//
// This file is kept for reference during migration and will be deleted.
// New code should use the type alias: interp.LazyThunk = runtime.LazyThunk

// LazyThunk represents a deferred/lazy parameter - an unevaluated expression
// that is re-evaluated each time it's accessed. This enables patterns like
// Jensen's Device, conditional evaluation, and lightweight anonymous functions.
//
// Key characteristics:
// - Expression is NOT evaluated at call site
// - Expression is re-evaluated EACH time the parameter is accessed
// - Environment is captured from the call site (closure semantics)
// - No caching - each access triggers a fresh evaluation
//
// Example:
//
//	function sum(var i: Integer; lo, hi: Integer; lazy term: Float): Float;
//	begin
//	   i := lo;
//	   while i <= hi do begin
//	      Result += term;  // term is re-evaluated each iteration
//	      Inc(i);
//	   end;
//	end;
//
//	// Jensen's Device: Computes harmonic series 1/1 + 1/2 + ... + 1/100
//	var i: Integer;
//	PrintLn(sum(i, 1, 100, 1.0/i));  // Each access to 'term' evaluates 1.0/i with current i
type LazyThunk struct {
	// Expression is the unevaluated AST node to be evaluated on each access
	Expression ast.Expression

	// CapturedEnv is the environment from the call site where the expression was defined.
	// This enables the expression to reference variables from the caller's scope,
	// and see mutations to those variables (critical for Jensen's Device).
	CapturedEnv *Environment

	// interpreter is a reference to the interpreter instance for evaluation.
	// We need this to call Eval() on the expression.
	interpreter *Interpreter
}

// NewLazyThunk creates a new lazy parameter thunk with the given expression,
// captured environment, and interpreter reference.
func NewLazyThunk(expr ast.Expression, env *Environment, interp *Interpreter) *LazyThunk {
	return &LazyThunk{
		Expression:  expr,
		CapturedEnv: env,
		interpreter: interp,
	}
}

// Type returns "LAZY_THUNK" to identify this as a lazy parameter.
func (t *LazyThunk) Type() string {
	return "LAZY_THUNK"
}

// String returns a representation of the lazy thunk for debugging.
func (t *LazyThunk) String() string {
	return "<lazy thunk: " + t.Expression.String() + ">"
}

// Evaluate evaluates the lazy parameter's expression in the captured environment
// and returns the result. This method is called each time the parameter is accessed,
// ensuring that the expression is re-evaluated with the current state of variables.
//
// Critical behavior:
// - Switches to the captured environment before evaluation
// - Restores the previous environment after evaluation
// - NO caching - each call performs a fresh evaluation
// - Variable mutations in the captured environment are visible
func (t *LazyThunk) Evaluate() Value {
	// Save the current interpreter environment
	savedEnv := t.interpreter.Env()

	// Switch to the captured environment from the call site
	// This ensures the expression sees variables from the caller's scope
	t.interpreter.SetEnvironment(t.CapturedEnv)

	var result Value

	// Fast path: if the expression is an identifier bound to another lazy thunk,
	// delegate to that thunk to avoid self-referential evaluation loops.
	if identExpr, ok := t.Expression.(*ast.Identifier); ok {
		if val, ok := t.CapturedEnv.Get(identExpr.Value); ok {
			if lazyVal, ok := val.(interface{ Evaluate() Value }); ok && lazyVal != t {
				result = lazyVal.Evaluate()
			}
		}
	}

	// Evaluate the expression in the captured environment
	if result == nil {
		result = t.interpreter.Eval(t.Expression)
	}

	// Restore the previous environment
	t.interpreter.SetEnvironment(savedEnv)

	return result
}
