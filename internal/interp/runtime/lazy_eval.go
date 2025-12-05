// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains callback types and values for lazy evaluation.
package runtime

// EvalCallback is a function that evaluates an expression and returns a value.
// This callback is used by LazyThunk to defer expression evaluation without
// storing a direct reference to the Interpreter.
//
// The callback captures:
// - The interpreter's Eval method
// - The expression to evaluate
// - The captured environment from the call site
//
// This enables LazyThunk to be moved to the runtime package while still being
// able to evaluate expressions. The callback captures all necessary context
// (interpreter instance, expression, environment) in a closure, breaking the
// direct dependency on the interp package.
//
// Example usage:
//
//	evalCallback := func() Value {
//	    savedEnv := interpreter.env
//	    interpreter.env = capturedEnv
//	    result := interpreter.Eval(expression)
//	    interpreter.env = savedEnv
//	    return result
//	}
//	lazyThunk := runtime.NewLazyThunk(expr, evalCallback)
type EvalCallback func() Value

// GetterCallback reads the value of a variable from an environment.
// This callback is used by ReferenceValue to dereference var parameters
// without storing a direct reference to Environment.
//
// The callback captures:
// - The environment containing the variable
// - The variable name
// - The Get operation
//
// Returns the current value of the variable, or an error if the variable
// is not found in the environment.
//
// This enables ReferenceValue to be moved to the runtime package while still
// being able to read variable values. The callback captures the environment
// and variable name in a closure, breaking the direct dependency on the
// interp.Environment type.
//
// Example usage:
//
//	getter := func() (Value, error) {
//	    return capturedEnv.Get(varName)
//	}
//	refValue := runtime.NewReferenceValue(varName, getter, setter)
type GetterCallback func() (Value, error)

// SetterCallback writes a value to a variable in an environment.
// This callback is used by ReferenceValue to assign through var parameters
// without storing a direct reference to Environment.
//
// The callback captures:
// - The environment containing the variable
// - The variable name
// - The Set operation
//
// Returns an error if the variable cannot be assigned (e.g., if the variable
// doesn't exist or type constraints prevent the assignment).
//
// This enables ReferenceValue to be moved to the runtime package while still
// being able to write variable values. The callback captures the environment
// and variable name in a closure, breaking the direct dependency on the
// interp.Environment type.
//
// Example usage:
//
//	setter := func(val Value) error {
//	    return capturedEnv.Set(varName, val)
//	}
//	refValue := runtime.NewReferenceValue(varName, getter, setter)
type SetterCallback func(Value) error

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
	// expression is the unevaluated AST node stored for debugging/display purposes.
	// The actual evaluation happens through the evaluator callback.
	// We store this as `any` to avoid importing the ast package into runtime.
	expression any

	// evaluator is a callback function that evaluates the expression and returns a value.
	// This callback captures the interpreter, expression, and captured environment,
	// enabling LazyThunk to live in the runtime package without direct dependencies.
	evaluator EvalCallback
}

// NewLazyThunk creates a new lazy parameter thunk with the given expression and evaluator callback.
//
// The evaluator callback should capture:
// - The interpreter's Eval method
// - The expression to evaluate
// - The captured environment from the call site
//
// Example usage:
//
//	evalCallback := func() Value {
//	    savedEnv := interpreter.env
//	    interpreter.env = capturedEnv
//	    result := interpreter.Eval(expression)
//	    interpreter.env = savedEnv
//	    return result
//	}
//	lazyThunk := runtime.NewLazyThunk(expr, evalCallback)
func NewLazyThunk(expr any, evaluator EvalCallback) *LazyThunk {
	return &LazyThunk{
		expression: expr,
		evaluator:  evaluator,
	}
}

// Type returns "LAZY_THUNK" to identify this as a lazy parameter.
func (t *LazyThunk) Type() string {
	return "LAZY_THUNK"
}

// String returns a representation of the lazy thunk for debugging.
// If the expression implements String(), we use that; otherwise we use a generic message.
func (t *LazyThunk) String() string {
	if stringer, ok := t.expression.(interface{ String() string }); ok {
		return "<lazy thunk: " + stringer.String() + ">"
	}
	return "<lazy thunk>"
}

// Evaluate evaluates the lazy parameter's expression in the captured environment
// and returns the result. This method is called each time the parameter is accessed,
// ensuring that the expression is re-evaluated with the current state of variables.
//
// Critical behavior:
// - Delegates to the evaluator callback (which handles environment switching)
// - NO caching - each call performs a fresh evaluation
// - Variable mutations in the captured environment are visible
func (t *LazyThunk) Evaluate() Value {
	return t.evaluator()
}

// ReferenceValue represents a reference to a variable in another environment.
//
// When a function has a var parameter, instead of copying the argument value,
// we create a ReferenceValue that points to the original variable in the caller's
// environment. This allows the function to modify the caller's variable.
//
// Example:
//
//	procedure Increment(var x: Integer);
//	begin
//	  x := x + 1;  // Modifies the caller's variable through the reference
//	end;
//
//	var n := 5;
//	Increment(n);  // n becomes 6
//
// Implementation:
//   - When calling Increment(n), instead of passing IntegerValue{5}, we pass
//     ReferenceValue{VarName: "n", getter: func..., setter: func...}
//   - When the function reads x, it calls getter() to get current value from caller's env
//   - When the function assigns to x, it calls setter() to write to caller's env
type ReferenceValue struct {
	// getter is a callback function that reads the current value of the variable.
	// This callback captures the environment and variable name, enabling
	// ReferenceValue to live in the runtime package without direct dependencies.
	getter GetterCallback

	// setter is a callback function that writes a new value to the variable.
	// This callback captures the environment and variable name, enabling
	// ReferenceValue to live in the runtime package without direct dependencies.
	setter SetterCallback

	// VarName is the name of the variable being referenced.
	// Stored for debugging/display purposes.
	VarName string
}

// NewReferenceValue creates a new var parameter reference with the given variable name
// and getter/setter callbacks.
//
// The getter callback should capture:
// - The environment containing the variable
// - The variable name
// - The Get operation
//
// The setter callback should capture:
// - The environment containing the variable
// - The variable name
// - The Set operation
//
// Example usage:
//
//	getter := func() (Value, error) {
//	    return capturedEnv.Get(varName)
//	}
//	setter := func(val Value) error {
//	    return capturedEnv.Set(varName, val)
//	}
//	refValue := runtime.NewReferenceValue(varName, getter, setter)
func NewReferenceValue(varName string, getter GetterCallback, setter SetterCallback) *ReferenceValue {
	return &ReferenceValue{
		VarName: varName,
		getter:  getter,
		setter:  setter,
	}
}

// Type returns "REFERENCE".
func (r *ReferenceValue) Type() string {
	return "REFERENCE"
}

// String returns a description of the reference.
func (r *ReferenceValue) String() string {
	return "&" + r.VarName
}

// Dereference returns the current value of the referenced variable.
func (r *ReferenceValue) Dereference() (Value, error) {
	return r.getter()
}

// Assign sets the value of the referenced variable.
func (r *ReferenceValue) Assign(value Value) error {
	return r.setter(value)
}
