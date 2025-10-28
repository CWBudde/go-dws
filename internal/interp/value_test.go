package interp

import "testing"

// RED: Write failing test for IntegerValue
func TestIntegerValue(t *testing.T) {
	tests := []struct {
		name     string
		wantType string
		wantStr  string
		value    int64
	}{
		{name: "positive integer", wantType: "INTEGER", wantStr: "42", value: 42},
		{name: "negative integer", wantType: "INTEGER", wantStr: "-123", value: -123},
		{name: "zero", wantType: "INTEGER", wantStr: "0", value: 0},
		{name: "large integer", wantType: "INTEGER", wantStr: "9223372036854775807", value: 9223372036854775807},
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
		wantType string
		wantStr  string
		value    float64
	}{
		{name: "positive float", wantType: "FLOAT", wantStr: "3.14", value: 3.14},
		{name: "negative float", wantType: "FLOAT", wantStr: "-2.5", value: -2.5},
		{name: "zero", wantType: "FLOAT", wantStr: "0", value: 0.0},
		{name: "integer-like float", wantType: "FLOAT", wantStr: "42", value: 42.0},
		{name: "scientific notation", wantType: "FLOAT", wantStr: "1.23e+10", value: 1.23e10},
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
		{name: "simple string", value: "hello", wantType: "STRING", wantStr: "hello"},
		{name: "empty string", value: "", wantType: "STRING", wantStr: ""},
		{name: "string with spaces", value: "hello world", wantType: "STRING", wantStr: "hello world"},
		{name: "string with quotes", value: "it's", wantType: "STRING", wantStr: "it's"},
		{name: "multiline string", value: "line1\nline2", wantType: "STRING", wantStr: "line1\nline2"},
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
		wantType string
		wantStr  string
		value    bool
	}{
		{name: "true value", wantType: "BOOLEAN", wantStr: "true", value: true},
		{name: "false value", wantType: "BOOLEAN", wantStr: "false", value: false},
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
func TestValueInterface(_ *testing.T) {
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
		value Value
		name  string
	}{
		{name: "float value", value: NewFloatValue(3.14)},
		{name: "string value", value: NewStringValue("hello")},
		{name: "boolean value", value: NewBooleanValue(true)},
		{name: "nil value", value: NewNilValue()},
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
		value Value
		name  string
	}{
		{name: "integer value", value: NewIntegerValue(42)},
		{name: "string value", value: NewStringValue("hello")},
		{name: "boolean value", value: NewBooleanValue(true)},
		{name: "nil value", value: NewNilValue()},
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
		value Value
		name  string
	}{
		{name: "integer value", value: NewIntegerValue(42)},
		{name: "float value", value: NewFloatValue(3.14)},
		{name: "boolean value", value: NewBooleanValue(true)},
		{name: "nil value", value: NewNilValue()},
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
		value Value
		name  string
	}{
		{name: "integer value", value: NewIntegerValue(42)},
		{name: "float value", value: NewFloatValue(3.14)},
		{name: "string value", value: NewStringValue("hello")},
		{name: "nil value", value: NewNilValue()},
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
