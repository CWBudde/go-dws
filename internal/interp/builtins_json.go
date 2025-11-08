package interp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// JSON Built-in Functions
// ============================================================================

// builtinParseJSON implements the JSON.Parse() built-in function.
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
// Raises exception on invalid JSON with position information.
//
// Example:
//
//	var obj := JSON.Parse('{"name": "John", "age": 30}');
//	PrintLn(obj); // Outputs: {name: John, age: 30}
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
	jsonVal, err := parseJSONString(jsonStr.Value)
	if err != nil {
		// Format error message with position information if available
		return i.newErrorWithLocation(i.currentNode, "JSON parse error: %s", err.Error())
	}

	variant := jsonValueToVariant(jsonVal)
	return variant
}

// parseJSONString parses a JSON string into a jsonvalue.Value.
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

// ============================================================================
// JSON Serialization Built-in Functions
// ============================================================================

// builtinToJSON implements the ToJSON() built-in function.
//
// Syntax: ToJSON(value: Variant): String
//
// Converts a DWScript value to a JSON string representation.
// Supported types:
//   - Primitives (Integer, Float, String, Boolean, Nil) → JSON primitives
//   - Arrays → JSON arrays (recursive)
//   - Records → JSON objects (recursive)
//   - JSON values → preserved as-is
//   - Variants → unwrapped and converted
//
// Unsupported types (Enums, Sets, Objects, Functions) are converted to null.
//
// Example:
//
//	var obj := NewRecord();
//	obj.name := 'John';
//	obj.age := 30;
//	var json := ToJSON(obj);
//	PrintLn(json); // Outputs: {"name":"John","age":30}
func (i *Interpreter) builtinToJSON(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ToJSON() expects exactly 1 argument, got %d", len(args))
	}

	// Convert the value to JSON using the existing valueToJSONValue helper
	jsonVal := valueToJSONValue(args[0])

	// Serialize to compact JSON using encoding/json
	jsonBytes, err := json.Marshal(jsonVal)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "ToJSON() serialization error: %s", err.Error())
	}

	return &StringValue{Value: string(jsonBytes)}
}

// builtinToJSONFormatted implements the ToJSONFormatted() built-in function.
//
// Syntax: ToJSONFormatted(value: Variant, indent: Integer): String
//
// Converts a DWScript value to a pretty-printed JSON string with the specified
// indentation level. Each nesting level is indented by 'indent' spaces.
//
// Parameters:
//   - value: The value to serialize
//   - indent: Number of spaces per indentation level (typically 2 or 4)
//
// Example:
//
//	var obj := NewRecord();
//	obj.name := 'John';
//	obj.age := 30;
//	var json := ToJSONFormatted(obj, 2);
//	PrintLn(json);
//	// Outputs:
//	// {
//	//   "name": "John",
//	//   "age": 30
//	// }
func (i *Interpreter) builtinToJSONFormatted(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "ToJSONFormatted() expects exactly 2 arguments, got %d", len(args))
	}

	// Extract indent parameter
	indentVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ToJSONFormatted() expects Integer as second argument, got %s", args[1].Type())
	}

	// Validate indent (must be non-negative)
	indent := int(indentVal.Value)
	if indent < 0 {
		return i.newErrorWithLocation(i.currentNode, "ToJSONFormatted() indent must be non-negative, got %d", indent)
	}

	// Convert the value to JSON using the existing valueToJSONValue helper
	jsonVal := valueToJSONValue(args[0])

	// Build indent string (e.g., "  " for indent=2)
	indentStr := strings.Repeat(" ", indent)

	// Serialize to formatted JSON using encoding/json.MarshalIndent
	jsonBytes, err := json.MarshalIndent(jsonVal, "", indentStr)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "ToJSONFormatted() serialization error: %s", err.Error())
	}

	return &StringValue{Value: string(jsonBytes)}
}

// ============================================================================
// JSON Object Access Built-in Functions
// ============================================================================

// builtinJSONHasField implements the JSONHasField() built-in function.
//
// Syntax: JSONHasField(obj: Variant, field: String): Boolean
//
// Returns True if the JSON object contains the specified field, False otherwise.
// If the value is not a JSON object, returns False.
//
// Example:
//
//	var obj := JSON.Parse('{"name": "John", "age": 30}');
//	PrintLn(JSONHasField(obj, 'name'));    // Outputs: true
//	PrintLn(JSONHasField(obj, 'email'));   // Outputs: false
func (i *Interpreter) builtinJSONHasField(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "JSONHasField() expects exactly 2 arguments, got %d", len(args))
	}

	// Unwrap the first argument to get the JSON value
	objVal := unwrapVariant(args[0])

	// Check if it's a JSON value
	jsonVal, ok := objVal.(*JSONValue)
	if !ok {
		// Not a JSON value - return false
		return &BooleanValue{Value: false}
	}

	// Check if the JSON value is an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		// Not a JSON object - return false
		return &BooleanValue{Value: false}
	}

	// Extract the field name
	fieldName, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "JSONHasField() expects String as second argument, got %s", args[1].Type())
	}

	// Check if the field exists
	fieldValue := jsonVal.Value.ObjectGet(fieldName.Value)
	return &BooleanValue{Value: fieldValue != nil}
}

// builtinJSONKeys implements the JSONKeys() built-in function.
//
// Syntax: JSONKeys(obj: Variant): array of String
//
// Returns an array containing all the keys of the JSON object in insertion order.
// If the value is not a JSON object, returns an empty array.
//
// Example:
//
//	var obj := JSON.Parse('{"name": "John", "age": 30, "city": "NYC"}');
//	var keys := JSONKeys(obj);
//	for var i := 0 to Length(keys) - 1 do
//	    PrintLn(keys[i]);
//	// Outputs: name, age, city (in insertion order)
func (i *Interpreter) builtinJSONKeys(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "JSONKeys() expects exactly 1 argument, got %d", len(args))
	}

	// Unwrap the argument to get the JSON value
	objVal := unwrapVariant(args[0])

	// Check if it's a JSON value
	jsonVal, ok := objVal.(*JSONValue)
	if !ok {
		// Not a JSON value - return empty array
		return i.createEmptyStringArray()
	}

	// Check if the JSON value is an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		// Not a JSON object - return empty array
		return i.createEmptyStringArray()
	}

	// Get the keys in insertion order
	keys := jsonVal.Value.ObjectKeys()

	// Convert to array of StringValue
	elements := make([]Value, len(keys))
	for idx, key := range keys {
		elements[idx] = &StringValue{Value: key}
	}

	// Create array type: array of String
	arrayType := types.NewDynamicArrayType(types.STRING)

	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// builtinJSONValues implements the JSONValues() built-in function.
//
// Syntax: JSONValues(obj: Variant): array of Variant
//
// Returns an array containing all the values of the JSON object in the same order as JSONKeys().
// Each value is wrapped in a Variant. If the value is not a JSON object, returns an empty array.
//
// Example:
//
//	var obj := JSON.Parse('{"name": "John", "age": 30}');
//	var values := JSONValues(obj);
//	for var i := 0 to Length(values) - 1 do
//	    PrintLn(values[i]);
//	// Outputs: John, 30
func (i *Interpreter) builtinJSONValues(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "JSONValues() expects exactly 1 argument, got %d", len(args))
	}

	// Unwrap the argument to get the JSON value
	objVal := unwrapVariant(args[0])

	// Check if it's a JSON value
	jsonVal, ok := objVal.(*JSONValue)
	if !ok {
		// Not a JSON value - return empty array
		return i.createEmptyVariantArray()
	}

	// Check if the JSON value is an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		// Not a JSON object - return empty array
		return i.createEmptyVariantArray()
	}

	// Get the keys in insertion order
	keys := jsonVal.Value.ObjectKeys()

	// Get the values in the same order
	elements := make([]Value, len(keys))
	for idx, key := range keys {
		fieldValue := jsonVal.Value.ObjectGet(key)
		// Convert JSON value to Variant
		elements[idx] = jsonValueToVariant(fieldValue)
	}

	// Create array type: array of Variant
	arrayType := types.NewDynamicArrayType(types.VARIANT)

	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// createEmptyStringArray creates an empty dynamic array of String.
func (i *Interpreter) createEmptyStringArray() *ArrayValue {
	arrayType := types.NewDynamicArrayType(types.STRING)
	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  []Value{},
	}
}

// createEmptyVariantArray creates an empty dynamic array of Variant.
func (i *Interpreter) createEmptyVariantArray() *ArrayValue {
	arrayType := types.NewDynamicArrayType(types.VARIANT)
	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  []Value{},
	}
}

// ============================================================================
// JSON Array Length Built-in Function
// ============================================================================

// builtinJSONLength implements the JSONLength() built-in function.
//
// Syntax: JSONLength(arr: Variant): Integer
//
// Returns the number of elements in a JSON array.
// If the value is not a JSON array, returns 0.
//
// Example:
//
//	var arr := ParseJSON('[10, 20, 30, 40, 50]');
//	PrintLn(JSONLength(arr));  // Outputs: 5
//
//	var obj := ParseJSON('{"name": "John"}');
//	PrintLn(JSONLength(obj));  // Outputs: 0 (not an array)
func (i *Interpreter) builtinJSONLength(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "JSONLength() expects exactly 1 argument, got %d", len(args))
	}

	// Unwrap the argument to get the JSON value
	arrVal := unwrapVariant(args[0])

	// Check if it's a JSON value
	jsonVal, ok := arrVal.(*JSONValue)
	if !ok {
		// Not a JSON value - return 0
		return &IntegerValue{Value: 0}
	}

	// Check if the JSON value is an array
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 3 { // KindArray = 3
		// Not a JSON array - return 0
		return &IntegerValue{Value: 0}
	}

	// Get the array length
	length := jsonVal.Value.ArrayLen()
	return &IntegerValue{Value: int64(length)}
}
