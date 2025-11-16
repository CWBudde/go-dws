package runtime

import (
	"fmt"
)

// ============================================================================
// Runtime Error Types
// ============================================================================
//
// These error types provide structured, informative errors for runtime
// operations. They include context about what went wrong and where.
//
// Error Types:
//   - ConversionError: Failed type conversion
//   - ArithmeticError: Arithmetic operation failed (overflow, division by zero)
//   - ComparisonError: Invalid comparison between values
//   - IndexError: Array/string index out of bounds
//   - NilError: Attempted to use a nil reference
//   - TypeError: Type mismatch or invalid operation for type
// ============================================================================

// ConversionError represents a failed type conversion.
type ConversionError struct {
	From   Value  // The source value (may be nil)
	To     string // Target type name
	Reason string // Why conversion failed
}

// Error implements the error interface.
func (e *ConversionError) Error() string {
	fromType := "nil"
	if e.From != nil {
		fromType = e.From.Type()
	}
	return fmt.Sprintf("cannot convert %s to %s: %s", fromType, e.To, e.Reason)
}

// NewConversionError creates a new conversion error.
func NewConversionError(from Value, to, reason string) error {
	return &ConversionError{
		From:   from,
		To:     to,
		Reason: reason,
	}
}

// ============================================================================

// ArithmeticError represents an arithmetic operation error.
type ArithmeticError struct {
	Operation string // e.g., "division by zero", "integer overflow"
}

// Error implements the error interface.
func (e *ArithmeticError) Error() string {
	return fmt.Sprintf("arithmetic error: %s", e.Operation)
}

// NewArithmeticError creates a new arithmetic error.
func NewArithmeticError(operation string) error {
	return &ArithmeticError{Operation: operation}
}

// ============================================================================

// ComparisonError represents an invalid comparison.
type ComparisonError struct {
	Left  Value
	Right Value
	Op    string // e.g., "=", "<", ">"
}

// Error implements the error interface.
func (e *ComparisonError) Error() string {
	leftType := "nil"
	rightType := "nil"
	if e.Left != nil {
		leftType = e.Left.Type()
	}
	if e.Right != nil {
		rightType = e.Right.Type()
	}
	return fmt.Sprintf("cannot compare %s %s %s", leftType, e.Op, rightType)
}

// NewComparisonError creates a new comparison error.
func NewComparisonError(left Value, right Value, op string) error {
	return &ComparisonError{
		Left:  left,
		Right: right,
		Op:    op,
	}
}

// ============================================================================

// IndexError represents an out-of-bounds index.
type IndexError struct {
	Type  string
	Index int64
	Min   int64
	Max   int64
}

// Error implements the error interface.
func (e *IndexError) Error() string {
	return fmt.Sprintf("index %d out of bounds for %s [%d..%d]", e.Index, e.Type, e.Min, e.Max)
}

// NewIndexError creates a new index error.
func NewIndexError(index, min, max int64, typ string) error {
	return &IndexError{
		Index: index,
		Min:   min,
		Max:   max,
		Type:  typ,
	}
}

// ============================================================================

// NilError represents an attempt to use a nil reference.
type NilError struct {
	Operation string // What was attempted (e.g., "access field", "call method")
	Type      string // Expected type (e.g., "object", "interface")
}

// Error implements the error interface.
func (e *NilError) Error() string {
	return fmt.Sprintf("nil %s: cannot %s", e.Type, e.Operation)
}

// NewNilError creates a new nil error.
func NewNilError(operation, typ string) error {
	return &NilError{
		Operation: operation,
		Type:      typ,
	}
}

// ============================================================================

// TypeError represents a type mismatch or invalid operation.
type TypeError struct {
	Expected string // Expected type(s)
	Got      Value  // Actual value
	Context  string // Where this happened (optional)
}

// Error implements the error interface.
func (e *TypeError) Error() string {
	gotType := "nil"
	if e.Got != nil {
		gotType = e.Got.Type()
	}

	if e.Context != "" {
		return fmt.Sprintf("type error in %s: expected %s, got %s", e.Context, e.Expected, gotType)
	}
	return fmt.Sprintf("type error: expected %s, got %s", e.Expected, gotType)
}

// NewTypeError creates a new type error.
func NewTypeError(expected string, got Value, context string) error {
	return &TypeError{
		Expected: expected,
		Got:      got,
		Context:  context,
	}
}

// ============================================================================
// Error Checking Utilities
// ============================================================================

// IsConversionError checks if an error is a ConversionError.
func IsConversionError(err error) bool {
	_, ok := err.(*ConversionError)
	return ok
}

// IsArithmeticError checks if an error is an ArithmeticError.
func IsArithmeticError(err error) bool {
	_, ok := err.(*ArithmeticError)
	return ok
}

// IsComparisonError checks if an error is a ComparisonError.
func IsComparisonError(err error) bool {
	_, ok := err.(*ComparisonError)
	return ok
}

// IsIndexError checks if an error is an IndexError.
func IsIndexError(err error) bool {
	_, ok := err.(*IndexError)
	return ok
}

// IsNilError checks if an error is a NilError.
func IsNilError(err error) bool {
	_, ok := err.(*NilError)
	return ok
}

// IsTypeError checks if an error is a TypeError.
func IsTypeError(err error) bool {
	_, ok := err.(*TypeError)
	return ok
}
