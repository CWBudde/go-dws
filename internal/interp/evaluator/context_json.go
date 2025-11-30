package evaluator

import (
	"encoding/json"
)

// ============================================================================
// JSON Parsing & Conversion Methods (Task 3.5.143t)
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
	// We need to use the adapter because the evaluator package can't create
	// interp.JSONValue or interp.VariantValue (circular dependency).
	// The adapter's WrapJSONValueInVariant method handles this for us.
	variant := e.adapter.WrapJSONValueInVariant(jsonVal)
	return variant, nil
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
