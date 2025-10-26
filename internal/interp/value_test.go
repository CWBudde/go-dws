package interp

import "testing"

// RED: Write failing test for IntegerValue
func TestIntegerValue(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		wantType string
		wantStr  string
	}{
		{"positive integer", 42, "INTEGER", "42"},
		{"negative integer", -123, "INTEGER", "-123"},
		{"zero", 0, "INTEGER", "0"},
		{"large integer", 9223372036854775807, "INTEGER", "9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := &IntegerValue{Value: tt.value}

			if got := val.Type(); got != tt.wantType {
				t.Errorf("IntegerValue.Type() = %v, want %v", got, tt.wantType)
			}

			if got := val.String(); got != tt.wantStr {
				t.Errorf("IntegerValue.String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

// RED: Write failing test for FloatValue
func TestFloatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		wantType string
		wantStr  string
	}{
		{"positive float", 3.14, "FLOAT", "3.14"},
		{"negative float", -2.5, "FLOAT", "-2.5"},
		{"zero", 0.0, "FLOAT", "0"},
		{"integer-like float", 42.0, "FLOAT", "42"},
		{"scientific notation", 1.23e10, "FLOAT", "1.23e+10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := &FloatValue{Value: tt.value}

			if got := val.Type(); got != tt.wantType {
				t.Errorf("FloatValue.Type() = %v, want %v", got, tt.wantType)
			}

			if got := val.String(); got != tt.wantStr {
				t.Errorf("FloatValue.String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

// RED: Write failing test for StringValue
func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantType string
		wantStr  string
	}{
		{"simple string", "hello", "STRING", "hello"},
		{"empty string", "", "STRING", ""},
		{"string with spaces", "hello world", "STRING", "hello world"},
		{"string with quotes", "it's", "STRING", "it's"},
		{"multiline string", "line1\nline2", "STRING", "line1\nline2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := &StringValue{Value: tt.value}

			if got := val.Type(); got != tt.wantType {
				t.Errorf("StringValue.Type() = %v, want %v", got, tt.wantType)
			}

			if got := val.String(); got != tt.wantStr {
				t.Errorf("StringValue.String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

// RED: Write failing test for BooleanValue
func TestBooleanValue(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		wantType string
		wantStr  string
	}{
		{"true value", true, "BOOLEAN", "true"},
		{"false value", false, "BOOLEAN", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := &BooleanValue{Value: tt.value}

			if got := val.Type(); got != tt.wantType {
				t.Errorf("BooleanValue.Type() = %v, want %v", got, tt.wantType)
			}

			if got := val.String(); got != tt.wantStr {
				t.Errorf("BooleanValue.String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

// RED: Write failing test for NilValue
func TestNilValue(t *testing.T) {
	val := &NilValue{}

	if got := val.Type(); got != "NIL" {
		t.Errorf("NilValue.Type() = %v, want NIL", got)
	}

	if got := val.String(); got != "nil" {
		t.Errorf("NilValue.String() = %v, want nil", got)
	}
}

// RED: Write failing test for helper conversion functions
func TestNewIntegerValue(t *testing.T) {
	val := NewIntegerValue(42)

	if val.Type() != "INTEGER" {
		t.Errorf("NewIntegerValue(42).Type() = %v, want INTEGER", val.Type())
	}

	if val.String() != "42" {
		t.Errorf("NewIntegerValue(42).String() = %v, want 42", val.String())
	}

	// Check that it returns an IntegerValue
	if _, ok := val.(*IntegerValue); !ok {
		t.Errorf("NewIntegerValue should return *IntegerValue, got %T", val)
	}
}

func TestNewFloatValue(t *testing.T) {
	val := NewFloatValue(3.14)

	if val.Type() != "FLOAT" {
		t.Errorf("NewFloatValue(3.14).Type() = %v, want FLOAT", val.Type())
	}

	if val.String() != "3.14" {
		t.Errorf("NewFloatValue(3.14).String() = %v, want 3.14", val.String())
	}

	// Check that it returns a FloatValue
	if _, ok := val.(*FloatValue); !ok {
		t.Errorf("NewFloatValue should return *FloatValue, got %T", val)
	}
}

func TestNewStringValue(t *testing.T) {
	val := NewStringValue("hello")

	if val.Type() != "STRING" {
		t.Errorf("NewStringValue('hello').Type() = %v, want STRING", val.Type())
	}

	if val.String() != "hello" {
		t.Errorf("NewStringValue('hello').String() = %v, want hello", val.String())
	}

	// Check that it returns a StringValue
	if _, ok := val.(*StringValue); !ok {
		t.Errorf("NewStringValue should return *StringValue, got %T", val)
	}
}

func TestNewBooleanValue(t *testing.T) {
	trueVal := NewBooleanValue(true)
	falseVal := NewBooleanValue(false)

	if trueVal.Type() != "BOOLEAN" {
		t.Errorf("NewBooleanValue(true).Type() = %v, want BOOLEAN", trueVal.Type())
	}

	if trueVal.String() != "true" {
		t.Errorf("NewBooleanValue(true).String() = %v, want true", trueVal.String())
	}

	if falseVal.String() != "false" {
		t.Errorf("NewBooleanValue(false).String() = %v, want false", falseVal.String())
	}

	// Check that it returns a BooleanValue
	if _, ok := trueVal.(*BooleanValue); !ok {
		t.Errorf("NewBooleanValue should return *BooleanValue, got %T", trueVal)
	}
}

func TestNewNilValue(t *testing.T) {
	val := NewNilValue()

	if val.Type() != "NIL" {
		t.Errorf("NewNilValue().Type() = %v, want NIL", val.Type())
	}

	if val.String() != "nil" {
		t.Errorf("NewNilValue().String() = %v, want nil", val.String())
	}

	// Check that it returns a NilValue
	if _, ok := val.(*NilValue); !ok {
		t.Errorf("NewNilValue should return *NilValue, got %T", val)
	}
}

// RED: Test that Value interface is correctly implemented
func TestValueInterface(t *testing.T) {
	// Verify all types implement Value interface
	var _ Value = &IntegerValue{}
	var _ Value = &FloatValue{}
	var _ Value = &StringValue{}
	var _ Value = &BooleanValue{}
	var _ Value = &NilValue{}
}

// Test GoInt conversion helper
func TestGoInt(t *testing.T) {
	// Success case
	intVal := NewIntegerValue(42)
	result, err := GoInt(intVal)
	if err != nil {
		t.Errorf("GoInt() unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("GoInt() = %v, want 42", result)
	}

	// Error cases
	tests := []struct {
		name  string
		value Value
	}{
		{"float value", NewFloatValue(3.14)},
		{"string value", NewStringValue("hello")},
		{"boolean value", NewBooleanValue(true)},
		{"nil value", NewNilValue()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GoInt(tt.value)
			if err == nil {
				t.Errorf("GoInt(%s) expected error, got nil", tt.name)
			}
		})
	}
}

// Test GoFloat conversion helper
func TestGoFloat(t *testing.T) {
	// Success case
	floatVal := NewFloatValue(3.14)
	result, err := GoFloat(floatVal)
	if err != nil {
		t.Errorf("GoFloat() unexpected error: %v", err)
	}
	if result != 3.14 {
		t.Errorf("GoFloat() = %v, want 3.14", result)
	}

	// Error cases
	tests := []struct {
		name  string
		value Value
	}{
		{"integer value", NewIntegerValue(42)},
		{"string value", NewStringValue("hello")},
		{"boolean value", NewBooleanValue(true)},
		{"nil value", NewNilValue()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GoFloat(tt.value)
			if err == nil {
				t.Errorf("GoFloat(%s) expected error, got nil", tt.name)
			}
		})
	}
}

// Test GoString conversion helper
func TestGoString(t *testing.T) {
	// Success case
	strVal := NewStringValue("hello")
	result, err := GoString(strVal)
	if err != nil {
		t.Errorf("GoString() unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("GoString() = %v, want hello", result)
	}

	// Error cases
	tests := []struct {
		name  string
		value Value
	}{
		{"integer value", NewIntegerValue(42)},
		{"float value", NewFloatValue(3.14)},
		{"boolean value", NewBooleanValue(true)},
		{"nil value", NewNilValue()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GoString(tt.value)
			if err == nil {
				t.Errorf("GoString(%s) expected error, got nil", tt.name)
			}
		})
	}
}

// Test GoBool conversion helper
func TestGoBool(t *testing.T) {
	// Success cases
	trueVal := NewBooleanValue(true)
	result, err := GoBool(trueVal)
	if err != nil {
		t.Errorf("GoBool(true) unexpected error: %v", err)
	}
	if result != true {
		t.Errorf("GoBool(true) = %v, want true", result)
	}

	falseVal := NewBooleanValue(false)
	result, err = GoBool(falseVal)
	if err != nil {
		t.Errorf("GoBool(false) unexpected error: %v", err)
	}
	if result != false {
		t.Errorf("GoBool(false) = %v, want false", result)
	}

	// Error cases
	tests := []struct {
		name  string
		value Value
	}{
		{"integer value", NewIntegerValue(42)},
		{"float value", NewFloatValue(3.14)},
		{"string value", NewStringValue("hello")},
		{"nil value", NewNilValue()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GoBool(tt.value)
			if err == nil {
				t.Errorf("GoBool(%s) expected error, got nil", tt.name)
			}
		})
	}
}
