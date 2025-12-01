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

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
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
	UnwrapVariant(value Value) Value

	// ToInt64 converts a Value to int64, handling SubrangeValue and EnumValue.
	// Returns the integer value and true if successful, 0 and false otherwise.
	ToInt64(value Value) (int64, bool)

	// ToBool converts a Value to bool.
	// Returns the boolean value and true if successful, false and false otherwise.
	ToBool(value Value) (bool, bool)

	// ToFloat64 converts a Value to float64, handling integer types.
	// Returns the float value and true if successful, 0.0 and false otherwise.
	ToFloat64(value Value) (float64, bool)

	// ParseJSONString parses a JSON string and returns a Value (typically a Variant containing JSONValue).
	// Returns an error if the JSON is invalid.
	ParseJSONString(jsonStr string) (Value, error)

	// ValueToJSON converts a DWScript Value to a JSON string.
	// If formatted is true, the output is pretty-printed with indentation.
	ValueToJSON(value Value, formatted bool) (string, error)

	// ValueToJSONWithIndent converts a DWScript Value to a JSON string with custom indentation.
	// If formatted is true, the output is pretty-printed with the specified indent spaces.
	ValueToJSONWithIndent(value Value, formatted bool, indent int) (string, error)

	// GetTypeOf returns the type name of a value (e.g., "INTEGER", "STRING", "TMyClass").
	GetTypeOf(value Value) string

	// GetClassOf returns the class name of an object value, or empty string if not an object.
	GetClassOf(value Value) string

	// JSONHasField checks if a JSON object value has a given field.
	// Returns false if value is not a JSON object or field doesn't exist.
	JSONHasField(value Value, fieldName string) bool

	// JSONGetKeys returns the keys of a JSON object in insertion order.
	// Returns empty array if value is not a JSON object.
	JSONGetKeys(value Value) []string

	// JSONGetValues returns the values of a JSON object/array.
	// Returns empty array if value is not a JSON object or array.
	JSONGetValues(value Value) []Value

	// JSONGetLength returns the length of a JSON array or object (number of keys).
	// Returns 0 if value is not a JSON array or object.
	JSONGetLength(value Value) int

	// CreateStringArray creates an array of strings from a slice of string values.
	CreateStringArray(values []string) Value

	// CreateVariantArray creates an array of Variants from a slice of values.
	CreateVariantArray(values []Value) Value

	// Write writes a string to the output without a newline.
	Write(s string)

	// WriteLine writes a string to the output followed by a newline.
	WriteLine(s string)

	// GetEnumOrdinal returns the ordinal value of an enum Value.
	// Returns (ordinal, true) if the value is an enum, (0, false) otherwise.
	GetEnumOrdinal(value Value) (int64, bool)

	// GetJSONVarType returns the VarType code for a JSON value based on its kind.
	// Returns (varType, true) if the value is a JSONValue, (0, false) otherwise.
	GetJSONVarType(value Value) (int64, bool)

	// GetBuiltinArrayLength returns the number of elements in an array.
	// Returns (length, true) if the value is an array, (0, false) otherwise.
	GetBuiltinArrayLength(value Value) (int64, bool)

	// SetArrayLength resizes a dynamic array to the specified length.
	// Returns an error if the value is not a dynamic array or the length is invalid.
	SetArrayLength(array Value, newLength int) error

	// ArrayCopy creates a deep copy of an array value.
	// Returns a new array with the same elements and type.
	ArrayCopy(array Value) Value

	// ArrayReverse reverses the elements of an array in place.
	// Returns nil on success, or an error value on failure.
	ArrayReverse(array Value) Value

	// ArraySort sorts the elements of an array in place using default comparison.
	// Returns nil on success, or an error value on failure.
	ArraySort(array Value) Value

	// EvalFunctionPointer calls a function pointer with the given arguments.
	// Returns the result value from the function call, or an error value on failure.
	EvalFunctionPointer(funcPtr Value, args []Value) Value

	// GetCallStackString returns a formatted string representation of the current call stack.
	GetCallStackString() string

	// GetCallStackArray returns the current call stack as an array of records.
	// Each record contains FunctionName, Line, and Column fields.
	GetCallStackArray() Value

	// IsAssigned checks if a value is assigned (not nil).
	// Returns true for non-nil values, including objects, variants, and other reference types.
	IsAssigned(value Value) bool

	// RaiseAssertionFailed raises an EAssertionFailed exception with an optional custom message.
	// The exception includes position information from the current node.
	RaiseAssertionFailed(customMessage string)

	// GetEnumSuccessor returns the successor of an enum value.
	// Returns (successor value, nil) on success, or (nil, error) if at maximum.
	GetEnumSuccessor(enumVal Value) (Value, error)

	// GetEnumPredecessor returns the predecessor of an enum value.
	// Returns (predecessor value, nil) on success, or (nil, error) if at minimum.
	GetEnumPredecessor(enumVal Value) (Value, error)

	// ParseInt parses a string to an integer with the specified base (2-36).
	// Returns (value, true) on success, or (0, false) on error.
	ParseInt(s string, base int) (int64, bool)

	// ParseFloat parses a string to a float64.
	// Returns (value, true) on success, or (0.0, false) on error.
	ParseFloat(s string) (float64, bool)

	// FormatString formats a string using Go fmt.Sprintf semantics with DWScript values.
	// Supports %s, %d, %f, %v, %x, %X, %o format verbs.
	// Returns (formatted string, nil) on success, or ("", error) on formatting error.
	FormatString(format string, args []Value) (string, error)

	// GetLowBound returns the lower bound for arrays, enums, or type meta-values.
	// Returns (low value, nil) on success, or (nil, error) on failure.
	GetLowBound(value Value) (Value, error)

	// GetHighBound returns the upper bound for arrays, enums, or type meta-values.
	// Returns (high value, nil) on success, or (nil, error) on failure.
	GetHighBound(value Value) (Value, error)

	// ConcatStrings concatenates multiple string values into a single string.
	// Returns the concatenated string value.
	ConcatStrings(args []Value) Value
}

// BuiltinFunc is the signature for all built-in function implementations.
// Each built-in receives:
// - ctx: Context for error reporting and AST node access
// - args: Slice of argument values passed to the function
//
// Returns:
// - A Value result (may be an error value if the function fails)
type BuiltinFunc func(ctx Context, args []Value) Value
