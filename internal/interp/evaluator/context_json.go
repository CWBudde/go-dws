package evaluator

import (
	"encoding/json"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// JSON Parsing & Conversion Methods
// ============================================================================
//
// This file implements the JSON parsing and conversion methods of the
// builtins.Context interface for the Evaluator:
// - ParseJSONString(): Parse JSON string and return Variant-wrapped JSONValue
// - ValueToJSON(): Convert DWScript value to JSON string
// - ValueToJSONWithIndent(): Convert DWScript value to formatted JSON string
//
// Used by JSON built-in functions: ParseJSON, ToJSON, ToJSONFormatted.
//
// Phase 3.5.143 - Phase IV: Complex Methods
// ============================================================================
//
// ============================================================================
// JSON Inspection Methods
// ============================================================================
//
// This file also implements JSON inspection methods:
// - JSONHasField(): Check if JSON object has a field
// - JSONGetKeys(): Get all keys from JSON object
// - JSONGetValues(): Get all values from JSON object/array
// - JSONGetLength(): Get length of JSON object/array
//
// Used by JSON built-in functions: JSONHasField, JSONKeys, JSONValues, JSONLength.
//
// Phase 3.5.143 - Phase IV: Complex Methods
// ============================================================================

// ParseJSONString parses a JSON string and returns a Value (Variant containing JSONValue).
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ParseJSON function.
func (e *Evaluator) ParseJSONString(jsonStr string) (Value, error) {
	// Parse JSON using the existing helper function from json_helpers.go
	jsonVal, err := ParseJSONString(jsonStr)
	if err != nil {
		return nil, err
	}

	// Convert to Variant containing JSONValue
	return runtime.BoxVariantWithJSON(jsonVal), nil
}

// ValueToJSON converts a DWScript Value to a JSON string.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ToJSON and ToJSONFormatted functions.
func (e *Evaluator) ValueToJSON(value Value, formatted bool) (string, error) {
	return e.ValueToJSONWithIndent(value, formatted, 2)
}

// ValueToJSONWithIndent converts a DWScript Value to a JSON string with custom indentation.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ToJSONFormatted function with custom indent.
func (e *Evaluator) ValueToJSONWithIndent(value Value, formatted bool, indent int) (string, error) {
	// Convert Value to jsonvalue.Value using existing helper from json_helpers.go
	jsonVal := ValueToJSONValue(value)

	// Serialize to JSON string using encoding/json
	var jsonBytes []byte
	var err error
	if formatted {
		// Build indent string
		indentStr := ""
		for j := 0; j < indent; j++ {
			indentStr += " "
		}
		jsonBytes, err = json.MarshalIndent(jsonVal, "", indentStr)
	} else {
		jsonBytes, err = json.Marshal(jsonVal)
	}

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// ============================================================================
// JSON Inspection Methods
// ============================================================================

// JSONHasField checks if a JSON object has a specific field.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONHasField function.
func (e *Evaluator) JSONHasField(value Value, fieldName string) bool {
	// Unwrap variant if present
	val := unwrapVariant(value)

	// Extract the underlying jsonvalue.Value using reflection
	// (can't type-assert to interp.JSONValue due to circular dependency)
	jsonVal := extractJSONValueViaReflection(val)
	if jsonVal == nil {
		return false
	}

	// Check if it's an object
	if jsonVal.Kind() != jsonvalue.KindObject {
		return false
	}

	// Check if field exists
	fieldValue := jsonVal.ObjectGet(fieldName)
	return fieldValue != nil
}

// JSONGetKeys returns the keys of a JSON object in insertion order.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONKeys function.
func (e *Evaluator) JSONGetKeys(value Value) []string {
	// Unwrap variant if present
	val := unwrapVariant(value)

	// Extract the underlying jsonvalue.Value using reflection
	jsonVal := extractJSONValueViaReflection(val)
	if jsonVal == nil {
		return []string{}
	}

	// Check if it's an object
	if jsonVal.Kind() != jsonvalue.KindObject {
		return []string{}
	}

	// Get keys
	return jsonVal.ObjectKeys()
}

// JSONGetValues returns the values of a JSON object/array.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONValues function.
func (e *Evaluator) JSONGetValues(value Value) []Value {
	// Unwrap variant if present
	val := unwrapVariant(value)

	// Extract the underlying jsonvalue.Value using reflection
	jsonVal := extractJSONValueViaReflection(val)
	if jsonVal == nil {
		return []Value{}
	}

	// Handle objects
	if jsonVal.Kind() == jsonvalue.KindObject {
		keys := jsonVal.ObjectKeys()
		values := make([]Value, len(keys))
		for idx, key := range keys {
			fieldVal := jsonVal.ObjectGet(key)
			// Wrap in Variant directly
			values[idx] = runtime.BoxVariantWithJSON(fieldVal)
		}
		return values
	}

	// Handle arrays
	if jsonVal.Kind() == jsonvalue.KindArray {
		arrayLen := jsonVal.ArrayLen()
		values := make([]Value, arrayLen)
		for idx := 0; idx < arrayLen; idx++ {
			elemVal := jsonVal.ArrayGet(idx)
			// Wrap in Variant directly
			values[idx] = runtime.BoxVariantWithJSON(elemVal)
		}
		return values
	}

	// Not an object or array
	return []Value{}
}

// JSONGetLength returns the length of a JSON array or object.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONLength function.
func (e *Evaluator) JSONGetLength(value Value) int {
	// Unwrap variant if present
	val := unwrapVariant(value)

	// Extract the underlying jsonvalue.Value using reflection
	jsonVal := extractJSONValueViaReflection(val)
	if jsonVal == nil {
		return 0
	}

	// Handle objects - return number of keys
	if jsonVal.Kind() == jsonvalue.KindObject {
		return len(jsonVal.ObjectKeys())
	}

	// Handle arrays - return length
	if jsonVal.Kind() == jsonvalue.KindArray {
		return jsonVal.ArrayLen()
	}

	// Not an object or array
	return 0
}
