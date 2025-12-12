package evaluator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// indexJSON performs JSON value indexing without importing the parent interp package.
// Supports both object property access (string index) and array element access (integer index).
//
// For JSON objects: obj['propertyName'] returns the value or nil if not found
// For JSON arrays: arr[index] returns the element or nil if out of bounds
func (e *Evaluator) indexJSON(base Value, index Value, node ast.Node) Value {
	// Extract the underlying jsonvalue.Value using reflection
	// This avoids importing the parent interp package
	jv := extractJSONValueViaReflection(base)

	// If we couldn't extract a JSON value, it's not a JSON type
	if jv == nil && base.Type() != "JSON" {
		return e.newError(node, "cannot index non-JSON value of type %s", base.Type())
	}

	// Handle nil/null JSON value
	if jv == nil {
		// nil/null JSON value - wrap in Variant directly
		return runtime.BoxVariantWithJSON(nil)
	}

	kind := jv.Kind()

	// JSON Object: support string indexing
	if kind == jsonvalue.KindObject {
		// Index must be a string for object property access
		// Check via Type() method to avoid importing interp package
		if index.Type() != "STRING" {
			return e.newError(node, "JSON object index must be a string, got %s", index.Type())
		}

		// Extract string value via String() method
		indexStr := index.String()

		// Get the property value (returns nil if not found)
		propValue := jv.ObjectGet(indexStr)

		// Wrap in Variant directly
		return runtime.BoxVariantWithJSON(propValue)
	}

	// JSON Array: support integer indexing
	if kind == jsonvalue.KindArray {
		// Extract integer index
		indexInt, ok := ExtractIntegerIndex(index)
		if !ok {
			return e.newError(node, "JSON array index must be an integer, got %s", index.Type())
		}

		// Get the array element (returns nil if out of bounds)
		elemValue := jv.ArrayGet(indexInt)

		// Wrap in Variant directly
		return runtime.BoxVariantWithJSON(elemValue)
	}

	// Not an object or array
	return e.newError(node, "cannot index JSON %s", kind.String())
}

// extractJSONValueViaReflection uses reflection to extract the internal jsonvalue.Value
// from a JSONValue struct, avoiding the need to import the parent interp package.
func extractJSONValueViaReflection(val Value) *jsonvalue.Value {
	if val == nil {
		return nil
	}

	// Check if this is a "JSON" type
	if val.Type() != "JSON" {
		return nil
	}

	// Use reflection to access the Value field
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	// Look for a field named "Value" of type *jsonvalue.Value
	valueField := rv.FieldByName("Value")
	if !valueField.IsValid() {
		return nil
	}

	// Try to convert to *jsonvalue.Value
	if jv, ok := valueField.Interface().(*jsonvalue.Value); ok {
		return jv
	}

	return nil
}

// ============================================================================
// JSON Conversion Helpers
// ============================================================================
//
// These standalone functions provide JSON conversion operations.
// They are used by both the Interpreter and Evaluator for implementing
// JSON parsing, serialization, and type conversions.
// ============================================================================

// VarType Constants
// These constants define the VarType codes used by DWScript's Variant type system.
// They match the standard Delphi VarType constants.
const (
	VarEmpty   = 0      // Unassigned/Empty
	VarNull    = 1      // Null (SQL NULL)
	VarInteger = 3      // 32-bit signed integer
	VarDouble  = 5      // Double precision float
	VarBoolean = 11     // Boolean
	VarInt64   = 20     // 64-bit signed integer
	VarString  = 256    // String value
	VarArray   = 0x2000 // Array flag (can be ORed with element type)
	VarJSON    = 0x4000 // Custom code for JSON objects
)

// JSONValueToValue converts a jsonvalue.Value to a DWScript runtime Value.
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
func JSONValueToValue(jv *jsonvalue.Value) Value {
	if jv == nil {
		return &runtime.NilValue{}
	}

	switch jv.Kind() {
	case jsonvalue.KindUndefined:
		return &runtime.NilValue{}
	case jsonvalue.KindNull:
		return &runtime.NilValue{}
	case jsonvalue.KindBoolean:
		return &runtime.BooleanValue{Value: jv.BoolValue()}
	case jsonvalue.KindInt64:
		return &runtime.IntegerValue{Value: jv.Int64Value()}
	case jsonvalue.KindNumber:
		return &runtime.FloatValue{Value: jv.NumberValue()}
	case jsonvalue.KindString:
		return &runtime.StringValue{Value: jv.StringValue()}
	case jsonvalue.KindArray, jsonvalue.KindObject:
		// Keep arrays/objects as JSONValue - need to create via reflection
		// to avoid importing interp package (circular dependency)
		// The caller will need to wrap this appropriately
		return createJSONValueViaReflection(jv)
	default:
		return &runtime.NilValue{}
	}
}

// createJSONValueViaReflection wraps a jsonvalue.Value in a runtime.JSONValue.
func createJSONValueViaReflection(jv *jsonvalue.Value) Value {
	if jv == nil {
		return &runtime.NilValue{}
	}
	return runtime.NewJSONValue(jv)
}

// ValueToJSONValue converts a DWScript runtime Value to a jsonvalue.Value.
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
func ValueToJSONValue(val Value) *jsonvalue.Value {
	if val == nil {
		return jsonvalue.NewNull()
	}

	// Unwrap Variant if present
	if wrapper, ok := val.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		if unwrapped == nil {
			return jsonvalue.NewNull()
		}
		val = unwrapped
	}

	switch v := val.(type) {
	case *runtime.NilValue:
		return jsonvalue.NewNull()
	case *runtime.BooleanValue:
		return jsonvalue.NewBoolean(v.Value)
	case *runtime.IntegerValue:
		return jsonvalue.NewInt64(v.Value)
	case *runtime.FloatValue:
		return jsonvalue.NewNumber(v.Value)
	case *runtime.StringValue:
		return jsonvalue.NewString(v.Value)
	case *runtime.ArrayValue:
		// Convert DWScript array to JSON array
		arr := jsonvalue.NewArray()
		for _, elem := range v.Elements {
			// Recursively convert each element
			jsonElem := ValueToJSONValue(elem)
			arr.ArrayAppend(jsonElem)
		}
		return arr
	case *runtime.RecordValue:
		// Convert DWScript record to JSON object
		obj := jsonvalue.NewObject()
		for fieldName, fieldValue := range v.Fields {
			// Recursively convert each field
			jsonField := ValueToJSONValue(fieldValue)
			obj.ObjectSet(fieldName, jsonField)
		}
		return obj
	default:
		// Check if it's a JSONValue by type name (avoid import)
		if val.Type() == "JSON" {
			// Extract using reflection
			jv := extractJSONValueViaReflection(val)
			if jv != nil {
				return jv
			}
		}
		// For unknown types, return null
		return jsonvalue.NewNull()
	}
}

// JSONKindToVarType maps a jsonvalue.Kind to a VarType code.
//
// This allows VarType(jsonVar) to return meaningful type codes.
// We use custom codes in the high range to avoid conflicts with standard VarType codes.
func JSONKindToVarType(kind jsonvalue.Kind) int64 {
	switch kind {
	case jsonvalue.KindUndefined:
		return VarEmpty // 0
	case jsonvalue.KindNull:
		return VarNull // 1
	case jsonvalue.KindBoolean:
		return VarBoolean // 11
	case jsonvalue.KindInt64:
		return VarInt64 // 20
	case jsonvalue.KindNumber:
		return VarDouble // 5
	case jsonvalue.KindString:
		return VarString // 256
	case jsonvalue.KindArray:
		return VarArray // 0x2000
	case jsonvalue.KindObject:
		return VarJSON // Custom code for JSON objects
	default:
		return VarEmpty
	}
}

// ParseJSONString parses a JSON string and returns a jsonvalue.Value.
// This is the core JSON parsing function used by ParseJSON and related functions.
func ParseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	// Decode JSON into Go interface{}
	var data interface{}
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber() // Preserve number precision

	err := decoder.Decode(&data)
	if err != nil {
		// Return detailed error with position
		return nil, FormatJSONError(err, jsonStr)
	}

	// Convert to jsonvalue.Value
	return GoValueToJSONValue(data), nil
}

// GoValueToJSONValue converts a Go interface{} value from encoding/json
// to our type-safe jsonvalue.Value representation.
//
// This handles the mapping:
//   - nil → JSON null
//   - bool → JSON boolean
//   - json.Number → JSON int64 or number
//   - string → JSON string
//   - []interface{} → JSON array (recursive)
//   - map[string]interface{} → JSON object (recursive)
func GoValueToJSONValue(data interface{}) *jsonvalue.Value {
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
			arr.ArrayAppend(GoValueToJSONValue(elem))
		}
		return arr

	case map[string]interface{}:
		// JSON object - recursively convert fields
		obj := jsonvalue.NewObject()
		for key, value := range v {
			obj.ObjectSet(key, GoValueToJSONValue(value))
		}
		return obj

	default:
		// Unknown type - return null
		return jsonvalue.NewNull()
	}
}

// FormatJSONError formats a JSON parsing error with position information.
// Extract line and column from syntax errors.
//
// Go's json.SyntaxError includes offset information which we convert
// to line:column format for better error messages.
func FormatJSONError(err error, jsonStr string) error {
	// Check if it's a syntax error with offset information
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		line, col := OffsetToLineCol(jsonStr, syntaxErr.Offset)
		return fmt.Errorf("invalid JSON at line %d, column %d: %s", line, col, syntaxErr.Error())
	}

	// Check if it's an unmarshal type error
	if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
		line, col := OffsetToLineCol(jsonStr, typeErr.Offset)
		return fmt.Errorf("invalid JSON type at line %d, column %d: %s", line, col, typeErr.Error())
	}

	// Other errors (e.g., unexpected EOF)
	return fmt.Errorf("invalid JSON: %s", err.Error())
}

// OffsetToLineCol converts a byte offset to line and column numbers.
// Helper for error position calculation.
//
// Lines and columns are 1-indexed for user-facing error messages.
func OffsetToLineCol(jsonStr string, offset int64) (line int, col int) {
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
