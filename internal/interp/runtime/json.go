// Package runtime provides runtime value types for the DWScript interpreter.
package runtime

import (
	"strconv"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// JSON Value Type
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

// String returns the string representation of the JSON value as produced when a
// JSONVariant is cast to a string / printed (matching DWScript's
// TBoxedJSONValue.ToString): a JSON string yields its raw content (no quotes), a
// container yields its compact JSON serialization, a boolean yields the Pascal
// "True"/"False", Null yields "Null", and Undefined/unassigned yields an empty
// string. Note this differs from the JSON serialization used inside a container
// (Stringify), where booleans and null are lowercase.
func (j *JSONValue) String() string {
	if j.Value == nil {
		return ""
	}

	switch j.Value.Kind() {
	case jsonvalue.KindUndefined:
		return ""
	case jsonvalue.KindNull:
		return "null"
	case jsonvalue.KindBoolean:
		if j.Value.BoolValue() {
			return "True"
		}
		return "False"
	case jsonvalue.KindString:
		return j.Value.StringValue()
	case jsonvalue.KindInt64:
		return strconv.FormatInt(j.Value.Int64Value(), 10)
	case jsonvalue.KindNumber:
		return (&FloatValue{Value: j.Value.NumberValue()}).String()
	case jsonvalue.KindObject, jsonvalue.KindArray:
		return jsonvalue.Stringify(j.Value)
	default:
		return ""
	}
}

// AsInteger converts a scalar JSON value to an integer (NumericValue interface).
// Objects, arrays, strings, null and undefined are not numeric.
func (j *JSONValue) AsInteger() (int64, bool) {
	if j.Value == nil {
		return 0, false
	}
	switch j.Value.Kind() {
	case jsonvalue.KindInt64:
		return j.Value.Int64Value(), true
	case jsonvalue.KindNumber:
		return int64(j.Value.NumberValue()), true
	case jsonvalue.KindBoolean:
		if j.Value.BoolValue() {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// AsFloat converts a scalar JSON value to a float (NumericValue interface).
func (j *JSONValue) AsFloat() (float64, bool) {
	if j.Value == nil {
		return 0, false
	}
	switch j.Value.Kind() {
	case jsonvalue.KindInt64:
		return float64(j.Value.Int64Value()), true
	case jsonvalue.KindNumber:
		return j.Value.NumberValue(), true
	case jsonvalue.KindBoolean:
		if j.Value.BoolValue() {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// IsUndefined reports whether the JSON value is Undefined (an unassigned /
// non-existent JSON value), which is how a JSONVariant reads as "empty".
func (j *JSONValue) IsUndefined() bool {
	return j.Value == nil || j.Value.Kind() == jsonvalue.KindUndefined
}

// NewJSONValue creates a new JSONValue wrapping a jsonvalue.Value.
func NewJSONValue(v *jsonvalue.Value) *JSONValue {
	return &JSONValue{Value: v}
}
