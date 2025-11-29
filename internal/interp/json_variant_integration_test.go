package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// ============================================================================
// Integration tests for JSON values with Variant system
// ============================================================================

// TestJSONValue_String verifies that JSONValue.String() produces correct output
func TestJSONValue_String(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *JSONValue
		wantStr  string
		wantType string
	}{
		{
			name: "null",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewNull())
			},
			wantStr:  "null",
			wantType: "JSON",
		},
		{
			name: "boolean true",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewBoolean(true))
			},
			wantStr:  "true",
			wantType: "JSON",
		},
		{
			name: "boolean false",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewBoolean(false))
			},
			wantStr:  "false",
			wantType: "JSON",
		},
		{
			name: "integer",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewInt64(42))
			},
			wantStr:  "42",
			wantType: "JSON",
		},
		{
			name: "negative integer",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewInt64(-123))
			},
			wantStr:  "-123",
			wantType: "JSON",
		},
		{
			name: "float",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewNumber(3.14))
			},
			wantStr:  "3.14",
			wantType: "JSON",
		},
		{
			name: "string",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewString("hello"))
			},
			wantStr:  "hello",
			wantType: "JSON",
		},
		{
			name: "empty string",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewString(""))
			},
			wantStr:  "",
			wantType: "JSON",
		},
		{
			name: "empty array",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewArray())
			},
			wantStr:  "[]",
			wantType: "JSON",
		},
		{
			name: "array with elements",
			setup: func() *JSONValue {
				arr := jsonvalue.NewArray()
				arr.ArrayAppend(jsonvalue.NewInt64(1))
				arr.ArrayAppend(jsonvalue.NewInt64(2))
				arr.ArrayAppend(jsonvalue.NewInt64(3))
				return NewJSONValue(arr)
			},
			wantStr:  "[1, 2, 3]",
			wantType: "JSON",
		},
		{
			name: "empty object",
			setup: func() *JSONValue {
				return NewJSONValue(jsonvalue.NewObject())
			},
			wantStr:  "{}",
			wantType: "JSON",
		},
		{
			name: "object with fields",
			setup: func() *JSONValue {
				obj := jsonvalue.NewObject()
				obj.ObjectSet("name", jsonvalue.NewString("John"))
				obj.ObjectSet("age", jsonvalue.NewInt64(30))
				return NewJSONValue(obj)
			},
			// Note: Object string representation depends on insertion order
			wantStr:  "{name: John, age: 30}",
			wantType: "JSON",
		},
		{
			name: "nested object",
			setup: func() *JSONValue {
				inner := jsonvalue.NewObject()
				inner.ObjectSet("x", jsonvalue.NewInt64(1))
				inner.ObjectSet("y", jsonvalue.NewInt64(2))

				outer := jsonvalue.NewObject()
				outer.ObjectSet("point", inner)
				return NewJSONValue(outer)
			},
			wantStr:  "{point: {x: 1, y: 2}}",
			wantType: "JSON",
		},
		{
			name: "nested array",
			setup: func() *JSONValue {
				inner := jsonvalue.NewArray()
				inner.ArrayAppend(jsonvalue.NewInt64(1))
				inner.ArrayAppend(jsonvalue.NewInt64(2))

				outer := jsonvalue.NewArray()
				outer.ArrayAppend(inner)
				outer.ArrayAppend(jsonvalue.NewInt64(3))
				return NewJSONValue(outer)
			},
			wantStr:  "[[1, 2], 3]",
			wantType: "JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonVal := tt.setup()

			if got := jsonVal.Type(); got != tt.wantType {
				t.Errorf("Type() = %v, want %v", got, tt.wantType)
			}

			if got := jsonVal.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

// TestJSONValue_VariantBoxing verifies JSON values can be boxed in Variants
func TestJSONValue_VariantBoxing(t *testing.T) {
	// Create a JSON value
	jv := jsonvalue.NewString("test")
	jsonVal := NewJSONValue(jv)

	// Box in Variant
	variant := BoxVariant(jsonVal)

	// Verify it's a Variant
	if variant.Type() != "VARIANT" {
		t.Errorf("Type() = %v, want VARIANT", variant.Type())
	}

	// Verify the wrapped value is the JSON value
	wrappedJSON, ok := variant.Value.(*JSONValue)
	if !ok {
		t.Fatal("expected wrapped value to be JSONValue")
	}

	if wrappedJSON.Value != jv {
		t.Error("wrapped JSON value mismatch")
	}

	// Verify String() delegation works
	if variant.String() != "test" {
		t.Errorf("String() = %v, want test", variant.String())
	}
}

// TestJSONValue_VarTypeIntegration verifies VarType works with JSON values
func TestJSONValue_VarTypeIntegration(t *testing.T) {
	tests := []struct {
		setup       func() *jsonvalue.Value
		name        string
		wantVarType int64
	}{
		{setup: jsonvalue.NewNull, name: "null", wantVarType: varNull},
		{setup: func() *jsonvalue.Value { return jsonvalue.NewBoolean(true) }, name: "boolean", wantVarType: varBoolean},
		{setup: func() *jsonvalue.Value { return jsonvalue.NewInt64(42) }, name: "int64", wantVarType: varInt64},
		{setup: func() *jsonvalue.Value { return jsonvalue.NewNumber(3.14) }, name: "number", wantVarType: varDouble},
		{setup: func() *jsonvalue.Value { return jsonvalue.NewString("test") }, name: "string", wantVarType: varString},
		{setup: jsonvalue.NewArray, name: "array", wantVarType: varArray},
		{setup: jsonvalue.NewObject, name: "object", wantVarType: varJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jv := tt.setup()
			_ = NewJSONValue(jv) // Verify we can create a JSONValue

			// Get VarType code
			gotCode := jsonKindToVarType(jv.Kind())
			if gotCode != tt.wantVarType {
				t.Errorf("jsonKindToVarType() = %v, want %v", gotCode, tt.wantVarType)
			}
		})
	}
}

// TestJSONValue_NestedStructures verifies complex nested JSON structures
func TestJSONValue_NestedStructures(t *testing.T) {
	// Create: {"users": [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}]}
	alice := jsonvalue.NewObject()
	alice.ObjectSet("name", jsonvalue.NewString("Alice"))
	alice.ObjectSet("age", jsonvalue.NewInt64(25))

	bob := jsonvalue.NewObject()
	bob.ObjectSet("name", jsonvalue.NewString("Bob"))
	bob.ObjectSet("age", jsonvalue.NewInt64(30))

	users := jsonvalue.NewArray()
	users.ArrayAppend(alice)
	users.ArrayAppend(bob)

	root := jsonvalue.NewObject()
	root.ObjectSet("users", users)

	jsonVal := NewJSONValue(root)

	// Verify type
	if jsonVal.Type() != "JSON" {
		t.Errorf("Type() = %v, want JSON", jsonVal.Type())
	}

	// Verify structure
	usersField := root.ObjectGet("users")
	if usersField == nil {
		t.Fatal("users field is nil")
	}
	if usersField.ArrayLen() != 2 {
		t.Errorf("users array length = %v, want 2", usersField.ArrayLen())
	}

	// Verify first user
	firstUser := usersField.ArrayGet(0)
	if firstUser == nil {
		t.Fatal("first user is nil")
	}
	name := firstUser.ObjectGet("name")
	if name == nil || name.StringValue() != "Alice" {
		t.Errorf("first user name = %v, want Alice", name)
	}

	// Verify string representation includes nested structure
	str := jsonVal.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	t.Logf("Nested structure string: %s", str)
}

// TestJSONValue_Conversion_RoundTrip verifies bidirectional conversion
func TestJSONValue_Conversion_RoundTrip(t *testing.T) {
	// Create a complex structure
	obj := jsonvalue.NewObject()
	obj.ObjectSet("name", jsonvalue.NewString("Test"))
	obj.ObjectSet("count", jsonvalue.NewInt64(42))
	obj.ObjectSet("active", jsonvalue.NewBoolean(true))

	arr := jsonvalue.NewArray()
	arr.ArrayAppend(jsonvalue.NewInt64(1))
	arr.ArrayAppend(jsonvalue.NewInt64(2))
	obj.ObjectSet("items", arr)

	// Convert to DWScript Value
	val := jsonValueToValue(obj)

	// Should be a JSONValue
	_, ok := val.(*JSONValue)
	if !ok {
		t.Fatal("expected JSONValue")
	}

	// Convert back to jsonvalue.Value
	result := valueToJSONValue(val)

	// Should be the same object
	if result != obj {
		t.Error("round-trip conversion should preserve identity for objects")
	}
}

// TestJSONValue_VariantConversion verifies Variant wrapper conversion
func TestJSONValue_VariantConversion(t *testing.T) {
	// Create JSON value
	jv := jsonvalue.NewInt64(123)

	// Wrap in Variant using jsonValueToVariant
	variant := jsonValueToVariant(jv)

	// Verify it's a Variant
	if variant.Type() != "VARIANT" {
		t.Errorf("Type() = %v, want VARIANT", variant.Type())
	}

	// Extract back using variantToJSONValue
	extracted := variantToJSONValue(variant)

	// Should get back the original jsonvalue
	if extracted.Kind() != jsonvalue.KindInt64 {
		t.Errorf("Kind() = %v, want KindInt64", extracted.Kind())
	}
	if extracted.Int64Value() != 123 {
		t.Errorf("Int64Value() = %v, want 123", extracted.Int64Value())
	}
}

// TestJSONValue_WithInterpreter simulates using JSON values in the interpreter
func TestJSONValue_WithInterpreter(t *testing.T) {
	// This test verifies that JSON values integrate properly with the interpreter's
	// type system and can be used like any other Value.

	// Create a JSON object
	obj := jsonvalue.NewObject()
	obj.ObjectSet("x", jsonvalue.NewInt64(10))
	obj.ObjectSet("y", jsonvalue.NewInt64(20))

	jsonVal := NewJSONValue(obj)

	// Verify it implements Value interface
	var _ Value = jsonVal

	// Verify Type() and String() work
	if jsonVal.Type() != "JSON" {
		t.Errorf("Type() = %v, want JSON", jsonVal.Type())
	}

	expectedStr := "{x: 10, y: 20}"
	if jsonVal.String() != expectedStr {
		t.Errorf("String() = %v, want %v", jsonVal.String(), expectedStr)
	}

	// Box in Variant
	variant := BoxVariant(jsonVal)

	// Unwrap
	unwrapped := unwrapVariant(variant)

	// Should get back the JSONValue
	unwrappedJSON, ok := unwrapped.(*JSONValue)
	if !ok {
		t.Fatal("unwrapped value should be JSONValue")
	}

	if unwrappedJSON.Value != obj {
		t.Error("unwrapped JSON value should be the same object")
	}
}
