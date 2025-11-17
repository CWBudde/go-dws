package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Type Introspection Built-in Functions (Task 3.7.6)
// ============================================================================
//
// This file contains type introspection functions that have been migrated
// from internal/interp to use the Context interface pattern.
//
// Functions in this file:
//   - TypeOf: Get the type name of a value
//   - TypeOfClass: Get the class name of an object
//
// These functions use Context helper methods to access type information
// without creating circular dependencies with internal/interp types.

// TypeOf returns the type name of a value.
// TypeOf(value: Variant): String
//
// Returns type names like:
//   - "INTEGER" for integers
//   - "FLOAT" for floats
//   - "STRING" for strings
//   - "BOOLEAN" for booleans
//   - "ARRAY" for arrays
//   - "RECORD" for records
//   - "TMyClass" for objects (returns class name)
//   - "ENUM" for enums
//   - "NULL" for nil/null values
//
// This is useful for runtime type checking and debugging.
func TypeOf(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("TypeOf() expects exactly 1 argument, got %d", len(args))
	}

	// Get type name using Context helper
	typeName := ctx.GetTypeOf(args[0])

	return &runtime.StringValue{Value: typeName}
}

// TypeOfClass returns the class name of an object value.
// TypeOfClass(obj: TObject): String
//
// Returns the class name (e.g., "TMyClass") if the value is an object.
// Returns empty string if the value is not an object.
//
// This is more specific than TypeOf() for objects - it only works with
// object instances and returns their exact class name.
func TypeOfClass(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("TypeOfClass() expects exactly 1 argument, got %d", len(args))
	}

	// Get class name using Context helper
	className := ctx.GetClassOf(args[0])

	return &runtime.StringValue{Value: className}
}
