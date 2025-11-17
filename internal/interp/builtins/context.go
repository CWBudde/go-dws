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

	// UnwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
	// This allows built-in functions to work with both direct values and Variant-wrapped values.
	// Task 9.4.5: Support for Variant arguments in built-in functions.
	UnwrapVariant(value Value) Value

	// ToInt64 converts a Value to int64, handling SubrangeValue and EnumValue.
	// Returns the integer value and true if successful, 0 and false otherwise.
	// Task 3.7.3: Type helper for conversion functions.
	ToInt64(value Value) (int64, bool)

	// ToBool converts a Value to bool.
	// Returns the boolean value and true if successful, false and false otherwise.
	// Task 3.7.3: Type helper for conversion functions.
	ToBool(value Value) (bool, bool)

	// ToFloat64 converts a Value to float64, handling integer types.
	// Returns the float value and true if successful, 0.0 and false otherwise.
	// Task 3.7.3: Type helper for conversion functions.
	ToFloat64(value Value) (float64, bool)

	// ParseJSONString parses a JSON string and returns a Value (typically a Variant containing JSONValue).
	// Returns an error if the JSON is invalid.
	// Task 3.7.6: JSON helper for ParseJSON function.
	ParseJSONString(jsonStr string) (Value, error)

	// ValueToJSON converts a DWScript Value to a JSON string.
	// If formatted is true, the output is pretty-printed with indentation.
	// Task 3.7.6: JSON helper for ToJSON and ToJSONFormatted functions.
	ValueToJSON(value Value, formatted bool) (string, error)

	// GetTypeOf returns the type name of a value (e.g., "INTEGER", "STRING", "TMyClass").
	// Task 3.7.6: Type introspection helper for TypeOf function.
	GetTypeOf(value Value) string

	// GetClassOf returns the class name of an object value, or empty string if not an object.
	// Task 3.7.6: Type introspection helper for TypeOfClass function.
	GetClassOf(value Value) string

	// JSONHasField checks if a JSON object value has a given field.
	// Returns false if value is not a JSON object or field doesn't exist.
	// Task 3.7.6: JSON helper for JSONHasField function.
	JSONHasField(value Value, fieldName string) bool

	// JSONGetKeys returns the keys of a JSON object in insertion order.
	// Returns empty array if value is not a JSON object.
	// Task 3.7.6: JSON helper for JSONKeys function.
	JSONGetKeys(value Value) []string

	// JSONGetValues returns the values of a JSON object/array.
	// Returns empty array if value is not a JSON object or array.
	// Task 3.7.6: JSON helper for JSONValues function.
	JSONGetValues(value Value) []Value

	// JSONGetLength returns the length of a JSON array or object (number of keys).
	// Returns 0 if value is not a JSON array or object.
	// Task 3.7.6: JSON helper for JSONLength function.
	JSONGetLength(value Value) int

	// CreateStringArray creates an array of strings from a slice of string values.
	// Task 3.7.6: Helper for creating string arrays in JSON functions.
	CreateStringArray(values []string) Value

	// CreateVariantArray creates an array of Variants from a slice of values.
	// Task 3.7.6: Helper for creating variant arrays in JSON functions.
	CreateVariantArray(values []Value) Value

	// Write writes a string to the output without a newline.
	// Task 3.7.4: I/O helper for Print function.
	Write(s string)

	// WriteLine writes a string to the output followed by a newline.
	// Task 3.7.4: I/O helper for PrintLn function.
	WriteLine(s string)

	// GetEnumOrdinal returns the ordinal value of an enum Value.
	// Returns (ordinal, true) if the value is an enum, (0, false) otherwise.
	// Task 3.7.5: Helper for Ord() function.
	GetEnumOrdinal(value Value) (int64, bool)
}

// BuiltinFunc is the signature for all built-in function implementations.
// Each built-in receives:
// - ctx: Context for error reporting and AST node access
// - args: Slice of argument values passed to the function
//
// Returns:
// - A Value result (may be an error value if the function fails)
type BuiltinFunc func(ctx Context, args []Value) Value
