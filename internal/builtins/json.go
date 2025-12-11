package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// JSON Built-in Functions
// ============================================================================
//
// This file contains JSON manipulation functions that have been migrated
// from internal/interp to use the Context interface pattern.
//
// Functions in this file:
//   - ParseJSON: Parse JSON string to Variant
//   - ToJSON: Convert value to JSON string (compact)
//   - ToJSONFormatted: Convert value to JSON string (formatted)
//   - JSONHasField: Check if JSON object has field
//   - JSONKeys: Get keys of JSON object
//   - JSONValues: Get values of JSON object/array
//   - JSONLength: Get length of JSON array or object
//
// These functions use Context helper methods to access JSON functionality
// without creating circular dependencies with internal/interp types.

// ParseJSON parses a JSON string and returns a Variant containing the parsed value.
// ParseJSON(s: String): Variant
//
// Supported JSON types:
//   - null → JSON null
//   - boolean → JSON boolean
//   - number → JSON number (float64 or int64)
//   - string → JSON string
//   - array → JSON array
//   - object → JSON object
//
// Raises exception on invalid JSON with position information.
func ParseJSON(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("JSON.Parse() expects exactly 1 argument, got %d", len(args))
	}

	// Extract string argument
	jsonStr, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("JSON.Parse() expects String argument, got %s", args[0].Type())
	}

	// Parse JSON using Context helper
	result, err := ctx.ParseJSONString(jsonStr.Value)
	if err != nil {
		return ctx.NewError("JSON parse error: %s", err.Error())
	}

	return result
}

// ToJSON converts a DWScript value to a compact JSON string representation.
// ToJSON(value: Variant): String
//
// Supported types:
//   - Primitives (Integer, Float, String, Boolean, Nil) → JSON primitives
//   - Arrays → JSON arrays (recursive)
//   - Records → JSON objects (recursive)
//   - JSON values → preserved as-is
//   - Variants → unwrapped and converted
//
// Unsupported types (Enums, Sets, Objects, Functions) are converted to null.
func ToJSON(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ToJSON() expects exactly 1 argument, got %d", len(args))
	}

	// Convert value to JSON string using Context helper
	jsonStr, err := ctx.ValueToJSON(args[0], false)
	if err != nil {
		return ctx.NewError("ToJSON() serialization error: %s", err.Error())
	}

	return &runtime.StringValue{Value: jsonStr}
}

// ToJSONFormatted converts a DWScript value to a pretty-printed JSON string.
// ToJSONFormatted(value: Variant, indent: Integer): String
//
// The output is formatted with the specified indentation (default: 2 spaces).
// This is useful for debugging or displaying JSON in a human-readable format.
func ToJSONFormatted(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("ToJSONFormatted() expects 1 or 2 arguments, got %d", len(args))
	}

	// Extract optional indent parameter (default: 2)
	indent := int64(2)
	if len(args) == 2 {
		indentVal, ok := ctx.ToInt64(args[1])
		if !ok {
			return ctx.NewError("ToJSONFormatted() expects Integer as second argument, got %s", args[1].Type())
		}
		indent = indentVal
		if indent < 0 {
			indent = 0
		}
	}

	// Convert value to formatted JSON string using Context helper with indent
	jsonStr, err := ctx.ValueToJSONWithIndent(args[0], true, int(indent))
	if err != nil {
		return ctx.NewError("ToJSONFormatted() serialization error: %s", err.Error())
	}

	return &runtime.StringValue{Value: jsonStr}
}

// JSONHasField checks if a JSON object has a given field.
// JSONHasField(obj: Variant, fieldName: String): Boolean
//
// Returns true if obj is a JSON object and has the specified field.
// Returns false if obj is not a JSON object or field doesn't exist.
func JSONHasField(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("JSONHasField() expects exactly 2 arguments, got %d", len(args))
	}

	// Extract field name
	fieldName, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("JSONHasField() expects String as second argument, got %s", args[1].Type())
	}

	// Check if field exists using Context helper
	hasField := ctx.JSONHasField(args[0], fieldName.Value)
	return &runtime.BooleanValue{Value: hasField}
}

// JSONKeys returns an array containing all the keys of a JSON object.
// JSONKeys(obj: Variant): array of String
//
// Returns keys in insertion order.
// If the value is not a JSON object, returns an empty array.
func JSONKeys(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("JSONKeys() expects exactly 1 argument, got %d", len(args))
	}

	// Get keys using Context helper
	keys := ctx.JSONGetKeys(args[0])

	// Create array of strings using Context helper
	return ctx.CreateStringArray(keys)
}

// JSONValues returns an array containing all the values of a JSON object or array.
// JSONValues(obj: Variant): array of Variant
//
// For objects: returns values in the same order as JSONKeys()
// For arrays: returns elements in order
// If the value is not a JSON object or array, returns an empty array.
func JSONValues(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("JSONValues() expects exactly 1 argument, got %d", len(args))
	}

	// Get values using Context helper
	values := ctx.JSONGetValues(args[0])

	// Create array of variants using Context helper
	return ctx.CreateVariantArray(values)
}

// JSONLength returns the length of a JSON array or object.
// JSONLength(value: Variant): Integer
//
// For arrays: returns the number of elements
// For objects: returns the number of keys
// For other types: returns 0
func JSONLength(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("JSONLength() expects exactly 1 argument, got %d", len(args))
	}

	// Get length using Context helper
	length := ctx.JSONGetLength(args[0])

	return &runtime.IntegerValue{Value: int64(length)}
}
