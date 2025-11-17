package interp

import (
	"encoding/json"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/types"
)

// Ensure Interpreter implements builtins.Context interface at compile time.
var _ builtins.Context = (*Interpreter)(nil)

// NewError creates an error value with location information from the current node.
// This implements the builtins.Context interface.
func (i *Interpreter) NewError(format string, args ...interface{}) builtins.Value {
	return i.newErrorWithLocation(i.currentNode, format, args...)
}

// CurrentNode returns the AST node currently being evaluated.
// This implements the builtins.Context interface.
func (i *Interpreter) CurrentNode() ast.Node {
	return i.currentNode
}

// RandSource returns the random number generator for built-in functions.
// This implements the builtins.Context interface.
func (i *Interpreter) RandSource() *rand.Rand {
	return i.rand
}

// GetRandSeed returns the current random number generator seed value.
// This implements the builtins.Context interface.
func (i *Interpreter) GetRandSeed() int64 {
	return i.randSeed
}

// SetRandSeed sets the random number generator seed.
// This implements the builtins.Context interface.
func (i *Interpreter) SetRandSeed(seed int64) {
	i.randSeed = seed
	i.rand = rand.New(rand.NewSource(seed))
}

// UnwrapVariant returns the underlying value if input is a Variant, otherwise returns input as-is.
// This implements the builtins.Context interface.
// Task 9.4.5: Support for Variant arguments in built-in functions.
func (i *Interpreter) UnwrapVariant(value builtins.Value) builtins.Value {
	if value != nil {
		// Check if it's a VariantValue and unwrap it
		if variant, ok := value.(*VariantValue); ok {
			if variant.Value == nil {
				return &UnassignedValue{}
			}
			return variant.Value
		}
	}
	return value
}

// ToInt64 converts a Value to int64, handling SubrangeValue and EnumValue.
// This implements the builtins.Context interface.
// Task 3.7.3: Type helper for conversion functions.
func (i *Interpreter) ToInt64(value builtins.Value) (int64, bool) {
	switch v := value.(type) {
	case *IntegerValue:
		return v.Value, true
	case *SubrangeValue:
		return int64(v.Value), true
	case *EnumValue:
		return int64(v.OrdinalValue), true
	case *BooleanValue:
		if v.Value {
			return 1, true
		}
		return 0, true
	case *FloatValue:
		return int64(v.Value), true
	default:
		return 0, false
	}
}

// ToBool converts a Value to bool.
// This implements the builtins.Context interface.
// Task 3.7.3: Type helper for conversion functions.
func (i *Interpreter) ToBool(value builtins.Value) (bool, bool) {
	switch v := value.(type) {
	case *BooleanValue:
		return v.Value, true
	case *IntegerValue:
		return v.Value != 0, true
	case *SubrangeValue:
		return v.Value != 0, true
	case *EnumValue:
		return v.OrdinalValue != 0, true
	default:
		return false, false
	}
}

// ToFloat64 converts a Value to float64, handling integer types.
// This implements the builtins.Context interface.
// Task 3.7.3: Type helper for conversion functions.
func (i *Interpreter) ToFloat64(value builtins.Value) (float64, bool) {
	switch v := value.(type) {
	case *FloatValue:
		return v.Value, true
	case *IntegerValue:
		return float64(v.Value), true
	case *SubrangeValue:
		return float64(v.Value), true
	case *EnumValue:
		return float64(v.OrdinalValue), true
	default:
		return 0.0, false
	}
}

// ParseJSONString parses a JSON string and returns a Value (Variant containing JSONValue).
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ParseJSON function.
func (i *Interpreter) ParseJSONString(jsonStr string) (builtins.Value, error) {
	// Parse JSON using the existing helper function
	jsonVal, err := parseJSONString(jsonStr)
	if err != nil {
		return nil, err
	}

	// Convert to Variant containing JSONValue
	variant := jsonValueToVariant(jsonVal)
	return variant, nil
}

// ValueToJSON converts a DWScript Value to a JSON string.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for ToJSON and ToJSONFormatted functions.
func (i *Interpreter) ValueToJSON(value builtins.Value, formatted bool) (string, error) {
	// Convert Value to jsonvalue.Value using existing helper
	jsonVal := valueToJSONValue(value)

	// Serialize to JSON string using encoding/json
	var jsonBytes []byte
	var err error
	if formatted {
		jsonBytes, err = json.MarshalIndent(jsonVal, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(jsonVal)
	}

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetTypeOf returns the type name of a value.
// This implements the builtins.Context interface.
// Task 3.7.6: Type introspection helper for TypeOf function.
func (i *Interpreter) GetTypeOf(value builtins.Value) string {
	if value == nil {
		return "NULL"
	}
	return value.Type()
}

// GetClassOf returns the class name of an object value.
// This implements the builtins.Context interface.
// Task 3.7.6: Type introspection helper for TypeOfClass function.
func (i *Interpreter) GetClassOf(value builtins.Value) string {
	if objVal, ok := value.(*ObjectInstance); ok {
		if objVal.Class != nil {
			return objVal.Class.Name
		}
	}
	return ""
}

// JSONHasField checks if a JSON object value has a given field.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONHasField function.
func (i *Interpreter) JSONHasField(value builtins.Value, fieldName string) bool {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return false
	}

	// Check if it's an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		return false
	}

	// Check if field exists
	fieldValue := jsonVal.Value.ObjectGet(fieldName)
	return fieldValue != nil
}

// JSONGetKeys returns the keys of a JSON object in insertion order.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONKeys function.
func (i *Interpreter) JSONGetKeys(value builtins.Value) []string {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return []string{}
	}

	// Check if it's an object
	if jsonVal.Value == nil || jsonVal.Value.Kind() != 2 { // KindObject = 2
		return []string{}
	}

	// Get keys
	return jsonVal.Value.ObjectKeys()
}

// JSONGetValues returns the values of a JSON object/array.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONValues function.
func (i *Interpreter) JSONGetValues(value builtins.Value) []builtins.Value {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return []builtins.Value{}
	}

	if jsonVal.Value == nil {
		return []builtins.Value{}
	}

	// Handle objects
	if jsonVal.Value.Kind() == 2 { // KindObject = 2
		keys := jsonVal.Value.ObjectKeys()
		values := make([]builtins.Value, len(keys))
		for idx, key := range keys {
			fieldVal := jsonVal.Value.ObjectGet(key)
			// Wrap in JSONValue and then Variant
			values[idx] = jsonValueToVariant(fieldVal)
		}
		return values
	}

	// Handle arrays
	if jsonVal.Value.Kind() == 1 { // KindArray = 1
		arrayLen := jsonVal.Value.ArrayLen()
		values := make([]builtins.Value, arrayLen)
		for idx := 0; idx < arrayLen; idx++ {
			elemVal := jsonVal.Value.ArrayGet(idx)
			// Wrap in JSONValue and then Variant
			values[idx] = jsonValueToVariant(elemVal)
		}
		return values
	}

	return []builtins.Value{}
}

// JSONGetLength returns the length of a JSON array or object.
// This implements the builtins.Context interface.
// Task 3.7.6: JSON helper for JSONLength function.
func (i *Interpreter) JSONGetLength(value builtins.Value) int {
	// Unwrap variant
	val := unwrapVariant(value)

	// Check if it's a JSON value
	jsonVal, ok := val.(*JSONValue)
	if !ok {
		return 0
	}

	if jsonVal.Value == nil {
		return 0
	}

	// Handle objects - return number of keys
	if jsonVal.Value.Kind() == 2 { // KindObject = 2
		return len(jsonVal.Value.ObjectKeys())
	}

	// Handle arrays - return length
	if jsonVal.Value.Kind() == 1 { // KindArray = 1
		return jsonVal.Value.ArrayLen()
	}

	return 0
}

// CreateStringArray creates an array of strings from a slice of string values.
// This implements the builtins.Context interface.
// Task 3.7.6: Helper for creating string arrays in JSON functions.
func (i *Interpreter) CreateStringArray(values []string) builtins.Value {
	// Convert strings to StringValue elements
	elements := make([]Value, len(values))
	for idx, str := range values {
		elements[idx] = &StringValue{Value: str}
	}

	// Create array type: array of String
	arrayType := types.NewDynamicArrayType(types.STRING)

	return &ArrayValue{
		Elements:  elements,
		ArrayType: arrayType,
	}
}

// CreateVariantArray creates an array of Variants from a slice of values.
// This implements the builtins.Context interface.
// Task 3.7.6: Helper for creating variant arrays in JSON functions.
func (i *Interpreter) CreateVariantArray(values []builtins.Value) builtins.Value {
	// Create array type: array of Variant
	arrayType := types.NewDynamicArrayType(types.VARIANT)

	return &ArrayValue{
		Elements:  values,
		ArrayType: arrayType,
	}
}
