package interp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// VarType Constants
// ============================================================================
// These constants define the VarType codes used by DWScript's Variant type system.
// They match the standard Delphi VarType constants.

const (
	varEmpty   = 0      // Unassigned/Empty
	varNull    = 1      // Null (SQL NULL)
	varInteger = 3      // 32-bit signed integer
	varDouble  = 5      // Double precision float
	varBoolean = 11     // Boolean
	varInt64   = 20     // 64-bit signed integer
	varString  = 256    // String value
	varArray   = 0x2000 // Array flag (can be ORed with element type)
	varJSON    = 0x4000 // Custom code for JSON objects
)

// ============================================================================
// JSON Value Conversions
// ============================================================================

// jsonValueToValue converts a jsonvalue.Value to a DWScript runtime Value.
// This is used when accessing JSON properties/elements to get native DWScript values.
//
// Conversion rules:
//   - JSON null → NilValue
//   - JSON boolean → BooleanValue
//   - JSON number → FloatValue
//   - JSON int64 → IntegerValue
//   - JSON string → StringValue
//   - JSON array → JSONValue (keep as reference for mutation)
//   - JSON object → JSONValue (keep as reference for mutation)
//
// Containers (arrays/objects) remain as JSONValue to preserve reference semantics
// and allow mutations to be visible.
func jsonValueToValue(jv *jsonvalue.Value) Value {
	if jv == nil {
		return &NilValue{}
	}

	switch jv.Kind() {
	case jsonvalue.KindUndefined:
		return &NilValue{}
	case jsonvalue.KindNull:
		return &NilValue{}
	case jsonvalue.KindBoolean:
		return &BooleanValue{Value: jv.BoolValue()}
	case jsonvalue.KindInt64:
		return &IntegerValue{Value: jv.Int64Value()}
	case jsonvalue.KindNumber:
		return &FloatValue{Value: jv.NumberValue()}
	case jsonvalue.KindString:
		return &StringValue{Value: jv.StringValue()}
	case jsonvalue.KindArray:
		// Keep arrays as JSONValue for reference semantics
		return &JSONValue{Value: jv}
	case jsonvalue.KindObject:
		// Keep objects as JSONValue for reference semantics
		return &JSONValue{Value: jv}
	default:
		return &NilValue{}
	}
}

// valueToJSONValue converts a DWScript runtime Value to a jsonvalue.Value.
// This is used when building JSON from DWScript values (for ToJSON/Stringify).
//
// Conversion rules:
//   - NilValue → JSON null
//   - BooleanValue → JSON boolean
//   - IntegerValue → JSON int64
//   - FloatValue → JSON number
//   - StringValue → JSON string
//   - ArrayValue → JSON array (recursive conversion)
//   - RecordValue → JSON object (recursive conversion)
//   - JSONValue → unwrap and return underlying jsonvalue.Value
//   - VariantValue → unwrap and convert the wrapped value
func valueToJSONValue(val Value) *jsonvalue.Value {
	if val == nil {
		return jsonvalue.NewNull()
	}

	// Unwrap Variant if present
	val = unwrapVariant(val)

	switch v := val.(type) {
	case *NilValue:
		return jsonvalue.NewNull()
	case *BooleanValue:
		return jsonvalue.NewBoolean(v.Value)
	case *IntegerValue:
		return jsonvalue.NewInt64(v.Value)
	case *FloatValue:
		return jsonvalue.NewNumber(v.Value)
	case *StringValue:
		return jsonvalue.NewString(v.Value)
	case *JSONValue:
		// Already a JSON value, return its underlying value
		return v.Value
	case *ArrayValue:
		// Convert DWScript array to JSON array
		arr := jsonvalue.NewArray()
		for _, elem := range v.Elements {
			// Recursively convert each element
			jsonElem := valueToJSONValue(elem)
			arr.ArrayAppend(jsonElem)
		}
		return arr
	case *RecordValue:
		// Convert DWScript record to JSON object
		obj := jsonvalue.NewObject()
		for fieldName, fieldValue := range v.Fields {
			// Recursively convert each field
			jsonField := valueToJSONValue(fieldValue)
			obj.ObjectSet(fieldName, jsonField)
		}
		return obj
	default:
		// For unknown types, return null
		return jsonvalue.NewNull()
	}
}

// jsonValueToVariant wraps a jsonvalue.Value in a VariantValue.
//
// This creates a JSONValue wrapper and boxes it in a Variant.
// The Variant's ActualType is set to nil since JSON is a dynamic type.
func jsonValueToVariant(jv *jsonvalue.Value) *VariantValue {
	if jv == nil {
		return &VariantValue{Value: nil, ActualType: nil}
	}

	// Wrap in JSONValue
	jsonVal := &JSONValue{Value: jv}

	// Box in Variant
	// Note: We don't have a specific types.Type for JSON, so use nil
	return &VariantValue{
		Value:      jsonVal,
		ActualType: nil, // JSON is a dynamic type
	}
}

// variantToJSONValue extracts the underlying jsonvalue.Value from a Variant.
// If the Variant contains a JSONValue, returns its underlying jsonvalue.Value.
// Otherwise, converts the Variant's value to JSON.
func variantToJSONValue(variant *VariantValue) *jsonvalue.Value {
	if variant == nil || variant.Value == nil {
		return jsonvalue.NewNull()
	}

	// If it's already a JSONValue, extract it
	if jsonVal, ok := variant.Value.(*JSONValue); ok {
		return jsonVal.Value
	}

	// Otherwise, convert the value to JSON
	return valueToJSONValue(variant.Value)
}

// jsonKindToVarType maps a jsonvalue.Kind to a VarType code.
//
// This allows VarType(jsonVar) to return meaningful type codes.
// We use custom codes in the high range to avoid conflicts with standard VarType codes.
func jsonKindToVarType(kind jsonvalue.Kind) int64 {
	switch kind {
	case jsonvalue.KindUndefined:
		return varEmpty // 0
	case jsonvalue.KindNull:
		return varNull // 1
	case jsonvalue.KindBoolean:
		return varBoolean // 11
	case jsonvalue.KindInt64:
		return varInt64 // 20
	case jsonvalue.KindNumber:
		return varDouble // 5
	case jsonvalue.KindString:
		return varString // 256
	case jsonvalue.KindArray:
		return varArray // 0x2000
	case jsonvalue.KindObject:
		return varJSON // Custom code for JSON objects (to be defined)
	default:
		return varEmpty
	}
}

// getJSONValueType returns the semantic type for a JSONValue.
// This is used for type checking in the semantic analyzer.
//
// Since JSON is a dynamic type, we return a generic JSONType.
// Individual elements (when accessed) have their actual types.
func getJSONValueType(jv *JSONValue) types.Type {
	// For now, return nil to indicate dynamic typing
	// TODO: Create a proper JSONType in types package if needed for semantic analysis
	return nil
}

// ============================================================================
// JSON Parsing Helpers
// ============================================================================

// parseJSONString parses a JSON string and returns a jsonvalue.Value.
// This is the core JSON parsing function used by ParseJSON and related functions.
func parseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	// Decode JSON into Go interface{}
	var data interface{}
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber() // Preserve number precision

	err := decoder.Decode(&data)
	if err != nil {
		// Return detailed error with position
		return nil, formatJSONError(err, jsonStr)
	}

	// Convert to jsonvalue.Value
	return goValueToJSONValue(data), nil
}

// goValueToJSONValue converts a Go interface{} value from encoding/json
// to our type-safe jsonvalue.Value representation.
//
// This handles the mapping:
//   - nil → JSON null
//   - bool → JSON boolean
//   - json.Number → JSON int64 or number
//   - string → JSON string
//   - []interface{} → JSON array (recursive)
//   - map[string]interface{} → JSON object (recursive)
func goValueToJSONValue(data interface{}) *jsonvalue.Value {
	if data == nil {
		return jsonvalue.NewNull()
	}

	switch v := data.(type) {
	case bool:
		return jsonvalue.NewBoolean(v)

	case json.Number:
		// Try to parse as int64 first (for whole numbers)
		if i64, err := v.Int64(); err == nil {
			return jsonvalue.NewInt64(i64)
		}
		// Otherwise parse as float64
		if f64, err := v.Float64(); err == nil {
			return jsonvalue.NewNumber(f64)
		}
		// Fallback to string if parsing fails
		return jsonvalue.NewString(v.String())

	case float64:
		// Direct float64 (when UseNumber is not used)
		return jsonvalue.NewNumber(v)

	case string:
		return jsonvalue.NewString(v)

	case []interface{}:
		// JSON array - recursively convert elements
		arr := jsonvalue.NewArray()
		for _, elem := range v {
			arr.ArrayAppend(goValueToJSONValue(elem))
		}
		return arr

	case map[string]interface{}:
		// JSON object - recursively convert fields
		obj := jsonvalue.NewObject()
		for key, value := range v {
			obj.ObjectSet(key, goValueToJSONValue(value))
		}
		return obj

	default:
		// Unknown type - return null
		return jsonvalue.NewNull()
	}
}

// formatJSONError formats a JSON parsing error with position information.
// Extract line and column from syntax errors.
//
// Go's json.SyntaxError includes offset information which we convert
// to line:column format for better error messages.
func formatJSONError(err error, jsonStr string) error {
	// Check if it's a syntax error with offset information
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		line, col := offsetToLineCol(jsonStr, syntaxErr.Offset)
		return fmt.Errorf("invalid JSON at line %d, column %d: %s", line, col, syntaxErr.Error())
	}

	// Check if it's an unmarshal type error
	if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
		line, col := offsetToLineCol(jsonStr, typeErr.Offset)
		return fmt.Errorf("invalid JSON type at line %d, column %d: %s", line, col, typeErr.Error())
	}

	// Other errors (e.g., unexpected EOF)
	return fmt.Errorf("invalid JSON: %s", err.Error())
}

// offsetToLineCol converts a byte offset to line and column numbers.
// Helper for error position calculation.
//
// Lines and columns are 1-indexed for user-facing error messages.
func offsetToLineCol(jsonStr string, offset int64) (line int, col int) {
	line = 1
	col = 1

	for i := 0; i < int(offset) && i < len(jsonStr); i++ {
		if jsonStr[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}
