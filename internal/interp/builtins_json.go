package interp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// JSON Built-in Functions
// Task 9.91-9.93: JSON parsing with error handling
// ============================================================================

// builtinParseJSON implements the JSON.Parse() built-in function.
// Task 9.91: Parse JSON string and return as Variant.
//
// Syntax: JSON.Parse(s: String): Variant
//
// Parses a JSON string and returns the result as a Variant containing a JSONValue.
// Supported JSON types:
//   - null → JSON null
//   - boolean → JSON boolean
//   - number → JSON number (float64 or int64)
//   - string → JSON string
//   - array → JSON array
//   - object → JSON object
//
// Task 9.92: Raises exception on invalid JSON with position information.
//
// Example:
//   var obj := JSON.Parse('{"name": "John", "age": 30}');
//   PrintLn(obj); // Outputs: {name: John, age: 30}
func (i *Interpreter) builtinParseJSON(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "JSON.Parse() expects exactly 1 argument, got %d", len(args))
	}

	// Extract string argument
	arg := args[0]
	jsonStr, ok := arg.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "JSON.Parse() expects String argument, got %s", arg.Type())
	}

	// Parse JSON using Go's encoding/json
	// Task 9.92: Parse errors are caught and reported with position information
	jsonVal, err := parseJSONString(jsonStr.Value)
	if err != nil {
		// Format error message with position information if available
		return i.newErrorWithLocation(i.currentNode, "JSON parse error: %s", err.Error())
	}

	// Wrap in Variant using the conversion function from Task 9.89
	variant := jsonValueToVariant(jsonVal)
	return variant
}

// parseJSONString parses a JSON string into a jsonvalue.Value.
// Task 9.91: Uses Go's encoding/json internally.
//
// This function converts from Go's interface{} representation to our
// type-safe jsonvalue.Value representation.
func parseJSONString(jsonStr string) (*jsonvalue.Value, error) {
	// Decode JSON into Go interface{}
	var data interface{}
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber() // Preserve number precision

	err := decoder.Decode(&data)
	if err != nil {
		// Task 9.92: Return detailed error with position
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
// Task 9.92: Extract line and column from syntax errors.
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
// Task 9.92: Helper for error position calculation.
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
