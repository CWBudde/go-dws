// Package runtime provides runtime value types for the DWScript interpreter.
package runtime

import (
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// JSON Value Type
// ============================================================================
//
// Task 3.5.160a: JSONValue moved from interp/value.go to runtime package
// to enable direct creation by evaluator without circular imports.
// ============================================================================

// JSONValue represents a JSON value in DWScript.
// It wraps the jsonvalue.Value type from the internal jsonvalue package.
type JSONValue struct {
	Value *jsonvalue.Value // The underlying JSON value
}

// Type returns "JSON".
func (j *JSONValue) Type() string {
	return "JSON"
}

// String returns the string representation of the JSON value.
// For primitives, returns the value directly.
// For objects and arrays, returns a JSON-like representation.
func (j *JSONValue) String() string {
	if j.Value == nil {
		return "undefined"
	}

	switch j.Value.Kind() {
	case jsonvalue.KindUndefined:
		return "undefined"
	case jsonvalue.KindNull:
		return "null"
	case jsonvalue.KindBoolean:
		// Extract boolean value for string representation
		// We need to access the primitive payload, but it's not exported
		// For now, use a simple approach
		return j.jsonValueToString(j.Value)
	case jsonvalue.KindString:
		return j.jsonValueToString(j.Value)
	case jsonvalue.KindNumber, jsonvalue.KindInt64:
		return j.jsonValueToString(j.Value)
	case jsonvalue.KindObject:
		return j.jsonObjectToString(j.Value)
	case jsonvalue.KindArray:
		return j.jsonArrayToString(j.Value)
	default:
		return "unknown"
	}
}

// jsonValueToString converts a primitive JSON value to string.
// This is a helper for String() method.
func (j *JSONValue) jsonValueToString(v *jsonvalue.Value) string {
	switch v.Kind() {
	case jsonvalue.KindNull:
		return "null"
	case jsonvalue.KindBoolean:
		if v.BoolValue() {
			return "true"
		}
		return "false"
	case jsonvalue.KindString:
		return v.StringValue()
	case jsonvalue.KindNumber:
		return strconv.FormatFloat(v.NumberValue(), 'g', -1, 64)
	case jsonvalue.KindInt64:
		return strconv.FormatInt(v.Int64Value(), 10)
	default:
		return "undefined"
	}
}

// jsonObjectToString converts a JSON object to string representation.
func (j *JSONValue) jsonObjectToString(v *jsonvalue.Value) string {
	keys := v.ObjectKeys()
	if len(keys) == 0 {
		return "{}"
	}

	var sb strings.Builder
	sb.WriteString("{")
	for i, key := range keys {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(key)
		sb.WriteString(": ")
		child := v.ObjectGet(key)
		if child != nil {
			// Recursively stringify child values
			childJSON := &JSONValue{Value: child}
			sb.WriteString(childJSON.String())
		} else {
			sb.WriteString("undefined")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// jsonArrayToString converts a JSON array to string representation.
func (j *JSONValue) jsonArrayToString(v *jsonvalue.Value) string {
	length := v.ArrayLen()
	if length == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < length; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		child := v.ArrayGet(i)
		if child != nil {
			// Recursively stringify child values
			childJSON := &JSONValue{Value: child}
			sb.WriteString(childJSON.String())
		} else {
			sb.WriteString("undefined")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// NewJSONValue creates a new JSONValue wrapping a jsonvalue.Value.
func NewJSONValue(v *jsonvalue.Value) *JSONValue {
	return &JSONValue{Value: v}
}
