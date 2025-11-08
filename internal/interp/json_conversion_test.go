package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Tests for jsonValueToValue (JSON → DWScript runtime values)
// ============================================================================

func TestJSONValueToValue_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		input    *jsonvalue.Value
		wantType string
		wantStr  string
	}{
		{
			name:     "null",
			input:    jsonvalue.NewNull(),
			wantType: "NIL",
			wantStr:  "nil",
		},
		{
			name:     "undefined",
			input:    jsonvalue.NewUndefined(),
			wantType: "NIL",
			wantStr:  "nil",
		},
		{
			name:     "boolean true",
			input:    jsonvalue.NewBoolean(true),
			wantType: "BOOLEAN",
			wantStr:  "True",
		},
		{
			name:     "boolean false",
			input:    jsonvalue.NewBoolean(false),
			wantType: "BOOLEAN",
			wantStr:  "False",
		},
		{
			name:     "int64",
			input:    jsonvalue.NewInt64(42),
			wantType: "INTEGER",
			wantStr:  "42",
		},
		{
			name:     "int64 negative",
			input:    jsonvalue.NewInt64(-123),
			wantType: "INTEGER",
			wantStr:  "-123",
		},
		{
			name:     "number",
			input:    jsonvalue.NewNumber(3.14),
			wantType: "FLOAT",
			wantStr:  "3.14",
		},
		{
			name:     "number zero",
			input:    jsonvalue.NewNumber(0.0),
			wantType: "FLOAT",
			wantStr:  "0",
		},
		{
			name:     "string",
			input:    jsonvalue.NewString("hello"),
			wantType: "STRING",
			wantStr:  "hello",
		},
		{
			name:     "empty string",
			input:    jsonvalue.NewString(""),
			wantType: "STRING",
			wantStr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonValueToValue(tt.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if got := result.Type(); got != tt.wantType {
				t.Errorf("Type() = %v, want %v", got, tt.wantType)
			}
			if got := result.String(); got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestJSONValueToValue_Array(t *testing.T) {
	// Create JSON array: [1, 2, 3]
	arr := jsonvalue.NewArray()
	arr.ArrayAppend(jsonvalue.NewInt64(1))
	arr.ArrayAppend(jsonvalue.NewInt64(2))
	arr.ArrayAppend(jsonvalue.NewInt64(3))

	result := jsonValueToValue(arr)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Arrays should be kept as JSONValue for reference semantics
	if got := result.Type(); got != "JSON" {
		t.Errorf("Type() = %v, want JSON", got)
	}

	jsonVal, ok := result.(*JSONValue)
	if !ok {
		t.Fatal("expected JSONValue")
	}

	if jsonVal.Value.ArrayLen() != 3 {
		t.Errorf("ArrayLen() = %v, want 3", jsonVal.Value.ArrayLen())
	}
}

func TestJSONValueToValue_Object(t *testing.T) {
	// Create JSON object: {"name": "John", "age": 30}
	obj := jsonvalue.NewObject()
	obj.ObjectSet("name", jsonvalue.NewString("John"))
	obj.ObjectSet("age", jsonvalue.NewInt64(30))

	result := jsonValueToValue(obj)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Objects should be kept as JSONValue for reference semantics
	if got := result.Type(); got != "JSON" {
		t.Errorf("Type() = %v, want JSON", got)
	}

	jsonVal, ok := result.(*JSONValue)
	if !ok {
		t.Fatal("expected JSONValue")
	}

	name := jsonVal.Value.ObjectGet("name")
	if name == nil || name.StringValue() != "John" {
		t.Errorf("ObjectGet(name) = %v, want John", name)
	}

	age := jsonVal.Value.ObjectGet("age")
	if age == nil || age.Int64Value() != 30 {
		t.Errorf("ObjectGet(age) = %v, want 30", age)
	}
}

func TestJSONValueToValue_Nil(t *testing.T) {
	result := jsonValueToValue(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if got := result.Type(); got != "NIL" {
		t.Errorf("Type() = %v, want NIL", got)
	}
}

// ============================================================================
// Tests for valueToJSONValue (DWScript runtime values → JSON)
// ============================================================================

func TestValueToJSONValue_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		wantKind jsonvalue.Kind
	}{
		{
			name:     "nil value",
			input:    &NilValue{},
			wantKind: jsonvalue.KindNull,
		},
		{
			name:     "boolean true",
			input:    &BooleanValue{Value: true},
			wantKind: jsonvalue.KindBoolean,
		},
		{
			name:     "boolean false",
			input:    &BooleanValue{Value: false},
			wantKind: jsonvalue.KindBoolean,
		},
		{
			name:     "integer",
			input:    &IntegerValue{Value: 42},
			wantKind: jsonvalue.KindInt64,
		},
		{
			name:     "float",
			input:    &FloatValue{Value: 3.14},
			wantKind: jsonvalue.KindNumber,
		},
		{
			name:     "string",
			input:    &StringValue{Value: "hello"},
			wantKind: jsonvalue.KindString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToJSONValue(tt.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if got := result.Kind(); got != tt.wantKind {
				t.Errorf("Kind() = %v, want %v", got, tt.wantKind)
			}
		})
	}
}

func TestValueToJSONValue_BooleanValues(t *testing.T) {
	trueVal := &BooleanValue{Value: true}
	result := valueToJSONValue(trueVal)
	if result.Kind() != jsonvalue.KindBoolean {
		t.Errorf("Kind() = %v, want KindBoolean", result.Kind())
	}
	if !result.BoolValue() {
		t.Error("BoolValue() = false, want true")
	}

	falseVal := &BooleanValue{Value: false}
	result = valueToJSONValue(falseVal)
	if result.Kind() != jsonvalue.KindBoolean {
		t.Errorf("Kind() = %v, want KindBoolean", result.Kind())
	}
	if result.BoolValue() {
		t.Error("BoolValue() = true, want false")
	}
}

func TestValueToJSONValue_NumericValues(t *testing.T) {
	// Integer
	intVal := &IntegerValue{Value: 42}
	result := valueToJSONValue(intVal)
	if result.Kind() != jsonvalue.KindInt64 {
		t.Errorf("Kind() = %v, want KindInt64", result.Kind())
	}
	if result.Int64Value() != 42 {
		t.Errorf("Int64Value() = %v, want 42", result.Int64Value())
	}

	// Float
	floatVal := &FloatValue{Value: 3.14}
	result = valueToJSONValue(floatVal)
	if result.Kind() != jsonvalue.KindNumber {
		t.Errorf("Kind() = %v, want KindNumber", result.Kind())
	}
	if result.NumberValue() != 3.14 {
		t.Errorf("NumberValue() = %v, want 3.14", result.NumberValue())
	}
}

func TestValueToJSONValue_String(t *testing.T) {
	strVal := &StringValue{Value: "hello world"}
	result := valueToJSONValue(strVal)
	if result.Kind() != jsonvalue.KindString {
		t.Errorf("Kind() = %v, want KindString", result.Kind())
	}
	if result.StringValue() != "hello world" {
		t.Errorf("StringValue() = %v, want 'hello world'", result.StringValue())
	}
}

func TestValueToJSONValue_JSONValue(t *testing.T) {
	// When given a JSONValue, should unwrap and return the underlying jsonvalue.Value
	jv := jsonvalue.NewString("test")
	jsonVal := &JSONValue{Value: jv}

	result := valueToJSONValue(jsonVal)
	if result != jv {
		t.Error("expected same jsonvalue.Value instance")
	}
}

func TestValueToJSONValue_Array(t *testing.T) {
	// Create DWScript array: [1, "hello", true]
	arr := &ArrayValue{
		ArrayType: nil, // Not testing type system here
		Elements: []Value{
			&IntegerValue{Value: 1},
			&StringValue{Value: "hello"},
			&BooleanValue{Value: true},
		},
	}

	result := valueToJSONValue(arr)
	if result.Kind() != jsonvalue.KindArray {
		t.Errorf("Kind() = %v, want KindArray", result.Kind())
	}
	if result.ArrayLen() != 3 {
		t.Errorf("ArrayLen() = %v, want 3", result.ArrayLen())
	}

	// Check elements
	elem0 := result.ArrayGet(0)
	if elem0.Kind() != jsonvalue.KindInt64 || elem0.Int64Value() != 1 {
		t.Errorf("element 0 = %v, want int64(1)", elem0)
	}

	elem1 := result.ArrayGet(1)
	if elem1.Kind() != jsonvalue.KindString || elem1.StringValue() != "hello" {
		t.Errorf("element 1 = %v, want string(hello)", elem1)
	}

	elem2 := result.ArrayGet(2)
	if elem2.Kind() != jsonvalue.KindBoolean || !elem2.BoolValue() {
		t.Errorf("element 2 = %v, want bool(true)", elem2)
	}
}

func TestValueToJSONValue_Record(t *testing.T) {
	// Create DWScript record with fields
	rec := &RecordValue{
		RecordType: nil, // Not testing type system here
		Fields: map[string]Value{
			"name": &StringValue{Value: "Alice"},
			"age":  &IntegerValue{Value: 25},
		},
	}

	result := valueToJSONValue(rec)
	if result.Kind() != jsonvalue.KindObject {
		t.Errorf("Kind() = %v, want KindObject", result.Kind())
	}

	// Check fields
	name := result.ObjectGet("name")
	if name == nil || name.Kind() != jsonvalue.KindString || name.StringValue() != "Alice" {
		t.Errorf("field 'name' = %v, want string(Alice)", name)
	}

	age := result.ObjectGet("age")
	if age == nil || age.Kind() != jsonvalue.KindInt64 || age.Int64Value() != 25 {
		t.Errorf("field 'age' = %v, want int64(25)", age)
	}
}

func TestValueToJSONValue_Variant(t *testing.T) {
	// Variant wrapping an integer
	variant := &VariantValue{
		Value:      &IntegerValue{Value: 123},
		ActualType: types.INTEGER,
	}

	result := valueToJSONValue(variant)
	if result.Kind() != jsonvalue.KindInt64 {
		t.Errorf("Kind() = %v, want KindInt64", result.Kind())
	}
	if result.Int64Value() != 123 {
		t.Errorf("Int64Value() = %v, want 123", result.Int64Value())
	}
}

func TestValueToJSONValue_Nil(t *testing.T) {
	result := valueToJSONValue(nil)
	if result.Kind() != jsonvalue.KindNull {
		t.Errorf("Kind() = %v, want KindNull", result.Kind())
	}
}

// ============================================================================
// Tests for jsonValueToVariant
// ============================================================================

func TestJSONValueToVariant(t *testing.T) {
	jv := jsonvalue.NewString("test")
	variant := jsonValueToVariant(jv)

	if variant == nil {
		t.Fatal("expected non-nil variant")
	}

	if variant.Type() != "VARIANT" {
		t.Errorf("Type() = %v, want VARIANT", variant.Type())
	}

	// Check wrapped value is a JSONValue
	jsonVal, ok := variant.Value.(*JSONValue)
	if !ok {
		t.Fatal("expected JSONValue in variant")
	}

	if jsonVal.Value != jv {
		t.Error("expected same jsonvalue.Value instance")
	}
}

func TestJSONValueToVariant_Nil(t *testing.T) {
	variant := jsonValueToVariant(nil)

	if variant == nil {
		t.Fatal("expected non-nil variant")
	}

	if variant.Value != nil {
		t.Errorf("Value = %v, want nil", variant.Value)
	}
}

// ============================================================================
// Tests for variantToJSONValue
// ============================================================================

func TestVariantToJSONValue(t *testing.T) {
	jv := jsonvalue.NewInt64(42)
	jsonVal := &JSONValue{Value: jv}
	variant := &VariantValue{
		Value:      jsonVal,
		ActualType: nil,
	}

	result := variantToJSONValue(variant)
	if result != jv {
		t.Error("expected same jsonvalue.Value instance")
	}
}

func TestVariantToJSONValue_NonJSON(t *testing.T) {
	// Variant wrapping a regular value (not JSONValue)
	variant := &VariantValue{
		Value:      &StringValue{Value: "hello"},
		ActualType: types.STRING,
	}

	result := variantToJSONValue(variant)
	if result.Kind() != jsonvalue.KindString {
		t.Errorf("Kind() = %v, want KindString", result.Kind())
	}
	if result.StringValue() != "hello" {
		t.Errorf("StringValue() = %v, want hello", result.StringValue())
	}
}

func TestVariantToJSONValue_Nil(t *testing.T) {
	result := variantToJSONValue(nil)
	if result.Kind() != jsonvalue.KindNull {
		t.Errorf("Kind() = %v, want KindNull", result.Kind())
	}
}

// ============================================================================
// Tests for jsonKindToVarType
// ============================================================================

func TestJSONKindToVarType(t *testing.T) {
	tests := []struct {
		name     string
		kind     jsonvalue.Kind
		wantCode int64
	}{
		{"undefined", jsonvalue.KindUndefined, varEmpty},
		{"null", jsonvalue.KindNull, varNull},
		{"boolean", jsonvalue.KindBoolean, varBoolean},
		{"int64", jsonvalue.KindInt64, varInt64},
		{"number", jsonvalue.KindNumber, varDouble},
		{"string", jsonvalue.KindString, varString},
		{"array", jsonvalue.KindArray, varArray},
		{"object", jsonvalue.KindObject, varJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jsonKindToVarType(tt.kind)
			if got != tt.wantCode {
				t.Errorf("jsonKindToVarType(%v) = %v, want %v", tt.kind, got, tt.wantCode)
			}
		})
	}
}

// ============================================================================
// Round-trip tests
// ============================================================================

func TestRoundTrip_PrimitivesToJSON(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{"nil", &NilValue{}},
		{"boolean", &BooleanValue{Value: true}},
		{"integer", &IntegerValue{Value: 42}},
		{"float", &FloatValue{Value: 3.14}},
		{"string", &StringValue{Value: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to JSON
			jv := valueToJSONValue(tt.value)
			// Convert back
			result := jsonValueToValue(jv)

			// Type should match (except float might be integer if whole number)
			originalType := tt.value.Type()
			resultType := result.Type()

			// Special case: FLOAT 3.14 → KindNumber → FLOAT (OK)
			// Special case: INTEGER → KindInt64 → INTEGER (OK)
			if originalType != resultType {
				t.Errorf("Type mismatch: %v → %v", originalType, resultType)
			}
		})
	}
}

func TestRoundTrip_JSONToDWScript(t *testing.T) {
	tests := []struct {
		name     string
		jsonVal  *jsonvalue.Value
		wantType string
	}{
		{"null", jsonvalue.NewNull(), "NIL"},
		{"boolean", jsonvalue.NewBoolean(true), "BOOLEAN"},
		{"int64", jsonvalue.NewInt64(42), "INTEGER"},
		{"number", jsonvalue.NewNumber(3.14), "FLOAT"},
		{"string", jsonvalue.NewString("test"), "STRING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to DWScript value
			val := jsonValueToValue(tt.jsonVal)
			// Convert back to JSON
			jv := valueToJSONValue(val)

			// Kind should match original
			if jv.Kind() != tt.jsonVal.Kind() {
				t.Errorf("Kind mismatch: %v → %v", tt.jsonVal.Kind(), jv.Kind())
			}
		})
	}
}
