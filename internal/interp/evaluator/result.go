package evaluator

// EvalResult wraps a Value to provide cleaner error propagation patterns.
// It reduces the boilerplate of repetitive "if isError(val) { return val }" checks
// while maintaining backward compatibility with existing code.
//
// Usage patterns:
//
//	// Pattern 1: Check and return early
//	result := NewResult(i.Eval(expr))
//	if err := result.Error(); err != nil {
//	    return err
//	}
//	value := result.Value()
//
//	// Pattern 2: Chain with OrReturn for early exit
//	value, err := NewResult(i.Eval(expr)).OrReturn()
//	if err != nil {
//	    return err
//	}
//
//	// Pattern 3: Check multiple values
//	left := NewResult(i.Eval(expr.Left))
//	right := NewResult(i.Eval(expr.Right))
//	if err := FirstError(left, right); err != nil {
//	    return err
//	}
type EvalResult struct {
	value Value
}

// NewResult creates a new EvalResult wrapping the given value.
func NewResult(value Value) *EvalResult {
	return &EvalResult{value: value}
}

// Val returns the wrapped value.
// If the result contains an error, this will return the ErrorValue.
func (r *EvalResult) Val() Value {
	return r.value
}

// Error returns the error if this result contains an ErrorValue, nil otherwise.
// This provides a cleaner way to check for errors compared to isError(val).
func (r *EvalResult) Error() Value {
	if r.value != nil && r.value.Type() == "ERROR" {
		return r.value
	}
	return nil
}

// IsError returns true if this result contains an error.
func (r *EvalResult) IsError() bool {
	return r.Error() != nil
}

// IsOk returns true if this result does not contain an error.
func (r *EvalResult) IsOk() bool {
	return r.Error() == nil
}

// OrReturn provides a convenient pattern for early error returns.
// It returns (value, nil) if there's no error, or (error, error) if there is.
//
// Usage:
//
//	value, err := NewResult(i.Eval(expr)).OrReturn()
//	if err != nil {
//	    return err
//	}
func (r *EvalResult) OrReturn() (Value, Value) {
	if err := r.Error(); err != nil {
		return err, err
	}
	return r.value, nil
}

// Unwrap returns the value if ok, or returns the error.
// This is useful when you want to propagate errors immediately.
func (r *EvalResult) Unwrap() Value {
	if err := r.Error(); err != nil {
		return err
	}
	return r.value
}

// Map applies a function to the value if there's no error.
// If there's an error, it returns the error unchanged.
// This enables functional-style error handling.
//
// Usage:
//
//	result := NewResult(i.Eval(expr)).Map(func(v Value) Value {
//	    // transform v
//	    return transformed
//	})
func (r *EvalResult) Map(fn func(Value) Value) *EvalResult {
	if r.IsError() {
		return r
	}
	return NewResult(fn(r.value))
}

// AndThen chains another evaluation if this result is ok.
// If this result has an error, it returns the error.
// This enables monadic-style chaining.
//
// Usage:
//
//	result := NewResult(i.Eval(expr1)).AndThen(func(v Value) Value {
//	    return i.Eval(expr2)
//	})
func (r *EvalResult) AndThen(fn func(Value) Value) *EvalResult {
	if r.IsError() {
		return r
	}
	return NewResult(fn(r.value))
}

// FirstError returns the first error from a list of results, or nil if all are ok.
// This is useful when evaluating multiple expressions and you want to fail fast.
//
// Usage:
//
//	left := NewResult(i.Eval(expr.Left))
//	right := NewResult(i.Eval(expr.Right))
//	if err := FirstError(left, right); err != nil {
//	    return err
//	}
func FirstError(results ...*EvalResult) Value {
	for _, result := range results {
		if err := result.Error(); err != nil {
			return err
		}
	}
	return nil
}

// AllValues returns all values if all results are ok, or the first error.
// This is useful for evaluating multiple expressions and collecting all values.
//
// Usage:
//
//	values, err := AllValues(
//	    NewResult(i.Eval(expr1)),
//	    NewResult(i.Eval(expr2)),
//	    NewResult(i.Eval(expr3)),
//	)
//	if err != nil {
//	    return err
//	}
func AllValues(results ...*EvalResult) ([]Value, Value) {
	values := make([]Value, len(results))
	for i, result := range results {
		if err := result.Error(); err != nil {
			return nil, err
		}
		values[i] = result.Val()
	}
	return values, nil
}

// Collect evaluates multiple expressions and collects their values.
// It returns the first error encountered, or all values if successful.
// This is a helper that combines evaluation and collection.
//
// Usage:
//
//	values, err := Collect(func(collect func(Value)) {
//	    collect(i.Eval(expr1))
//	    collect(i.Eval(expr2))
//	    collect(i.Eval(expr3))
//	})
//	if err != nil {
//	    return err
//	}
func Collect(fn func(collect func(Value))) ([]Value, Value) {
	var values []Value
	var firstError Value

	collect := func(value Value) {
		if firstError != nil {
			return // Already have an error, skip
		}
		result := NewResult(value)
		if err := result.Error(); err != nil {
			firstError = err
			return
		}
		values = append(values, value)
	}

	fn(collect)

	if firstError != nil {
		return nil, firstError
	}
	return values, nil
}
