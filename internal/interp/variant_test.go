package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Task 9.231: Runtime tests for Variant operations
// ============================================================================

// ============================================================================
// Boxing/Unboxing Tests
// ============================================================================

func TestVariantBoxingInteger(t *testing.T) {
	intVal := &IntegerValue{Value: 42}
	variant := boxVariant(intVal)

	if variant.Type() != "VARIANT" {
		t.Fatalf("expected VARIANT type, got %s", variant.Type())
	}

	if variant.Value != intVal {
		t.Fatalf("expected wrapped value to be %v, got %v", intVal, variant.Value)
	}

	if !variant.ActualType.Equals(types.INTEGER) {
		t.Fatalf("expected ActualType to be INTEGER, got %s", variant.ActualType)
	}
}

func TestVariantBoxingString(t *testing.T) {
	strVal := &StringValue{Value: "hello"}
	variant := boxVariant(strVal)

	if variant.Type() != "VARIANT" {
		t.Fatalf("expected VARIANT type, got %s", variant.Type())
	}

	if variant.Value != strVal {
		t.Fatalf("expected wrapped value to be %v, got %v", strVal, variant.Value)
	}

	if !variant.ActualType.Equals(types.STRING) {
		t.Fatalf("expected ActualType to be STRING, got %s", variant.ActualType)
	}
}

func TestVariantBoxingFloat(t *testing.T) {
	floatVal := &FloatValue{Value: 3.14}
	variant := boxVariant(floatVal)

	if variant.Type() != "VARIANT" {
		t.Fatalf("expected VARIANT type, got %s", variant.Type())
	}

	if variant.Value != floatVal {
		t.Fatalf("expected wrapped value to be %v, got %v", floatVal, variant.Value)
	}

	if !variant.ActualType.Equals(types.FLOAT) {
		t.Fatalf("expected ActualType to be FLOAT, got %s", variant.ActualType)
	}
}

func TestVariantBoxingBoolean(t *testing.T) {
	boolVal := &BooleanValue{Value: true}
	variant := boxVariant(boolVal)

	if variant.Type() != "VARIANT" {
		t.Fatalf("expected VARIANT type, got %s", variant.Type())
	}

	if variant.Value != boolVal {
		t.Fatalf("expected wrapped value to be %v, got %v", boolVal, variant.Value)
	}

	if !variant.ActualType.Equals(types.BOOLEAN) {
		t.Fatalf("expected ActualType to be BOOLEAN, got %s", variant.ActualType)
	}
}

func TestVariantBoxingNil(t *testing.T) {
	variant := boxVariant(nil)

	if variant.Type() != "VARIANT" {
		t.Fatalf("expected VARIANT type, got %s", variant.Type())
	}

	if variant.Value != nil {
		t.Fatalf("expected wrapped value to be nil, got %v", variant.Value)
	}

	if variant.ActualType != nil {
		t.Fatalf("expected ActualType to be nil, got %s", variant.ActualType)
	}
}

func TestVariantBoxingAlreadyVariant(t *testing.T) {
	// Boxing a Variant should return it as-is (no double-wrapping)
	intVal := &IntegerValue{Value: 42}
	variant1 := boxVariant(intVal)
	variant2 := boxVariant(variant1)

	if variant1 != variant2 {
		t.Fatal("expected boxing a Variant to return the same Variant (no double-wrapping)")
	}
}

func TestVariantUnboxing(t *testing.T) {
	intVal := &IntegerValue{Value: 42}
	variant := boxVariant(intVal)

	unboxed, ok := unboxVariant(variant)
	if !ok {
		t.Fatal("expected successful unboxing")
	}

	if unboxed != intVal {
		t.Fatalf("expected unboxed value to be %v, got %v", intVal, unboxed)
	}
}

func TestVariantUnboxingNonVariant(t *testing.T) {
	intVal := &IntegerValue{Value: 42}

	unboxed, ok := unboxVariant(intVal)
	if ok {
		t.Fatal("expected unboxing non-Variant to fail")
	}

	if unboxed != nil {
		t.Fatalf("expected nil for failed unboxing, got %v", unboxed)
	}
}

func TestVariantUnwrapping(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected Value
	}{
		{
			name:     "unwrap integer variant",
			input:    boxVariant(&IntegerValue{Value: 42}),
			expected: &IntegerValue{Value: 42},
		},
		{
			name:     "unwrap string variant",
			input:    boxVariant(&StringValue{Value: "hello"}),
			expected: &StringValue{Value: "hello"},
		},
		{
			name:     "unwrap non-variant returns as-is",
			input:    &IntegerValue{Value: 42},
			expected: &IntegerValue{Value: 42},
		},
		{
			name:     "unwrap nil variant returns NilValue",
			input:    boxVariant(nil),
			expected: &NilValue{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unwrapVariant(tt.input)

			// Compare by value, not pointer
			if result.Type() != tt.expected.Type() {
				t.Fatalf("expected type %s, got %s", tt.expected.Type(), result.Type())
			}

			if result.String() != tt.expected.String() {
				t.Fatalf("expected value %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// ============================================================================
// Arithmetic Operation Tests
// ============================================================================

func TestVariantArithmeticIntegerPlusInteger(t *testing.T) {
	input := `
		var v1: Variant := 10;
		var v2: Variant := 20;
		var result: Variant;
		begin
			result := v1 + v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectInteger(t, result, 30)
}

func TestVariantArithmeticIntegerMinusInteger(t *testing.T) {
	input := `
		var v1: Variant := 50;
		var v2: Variant := 30;
		var result: Variant;
		begin
			result := v1 - v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectInteger(t, result, 20)
}

func TestVariantArithmeticIntegerTimesInteger(t *testing.T) {
	input := `
		var v1: Variant := 6;
		var v2: Variant := 7;
		var result: Variant;
		begin
			result := v1 * v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectInteger(t, result, 42)
}

func TestVariantArithmeticIntegerDivInteger(t *testing.T) {
	input := `
		var v1: Variant := 20;
		var v2: Variant := 4;
		var result: Variant;
		begin
			result := v1 / v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectFloat(t, result, 5.0)
}

func TestVariantArithmeticIntegerDivOperator(t *testing.T) {
	input := `
		var v1: Variant := 20;
		var v2: Variant := 3;
		var result: Variant;
		begin
			result := v1 div v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectInteger(t, result, 6)
}

func TestVariantArithmeticIntegerModOperator(t *testing.T) {
	input := `
		var v1: Variant := 20;
		var v2: Variant := 3;
		var result: Variant;
		begin
			result := v1 mod v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectInteger(t, result, 2)
}

func TestVariantArithmeticFloatPlusFloat(t *testing.T) {
	input := `
		var v1: Variant := 3.14;
		var v2: Variant := 2.86;
		var result: Variant;
		begin
			result := v1 + v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectFloatClose(t, result, 6.0, 0.001)
}

func TestVariantArithmeticIntegerPlusFloat(t *testing.T) {
	input := `
		var v1: Variant := 10;
		var v2: Variant := 2.5;
		var result: Variant;
		begin
			result := v1 + v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectFloatClose(t, result, 12.5, 0.001)
}

func TestVariantArithmeticStringConcatenation(t *testing.T) {
	input := `
		var v1: Variant := "hello";
		var v2: Variant := " world";
		var result: Variant;
		begin
			result := v1 + v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectString(t, result, "hello world")
}

func TestVariantArithmeticStringPlusInteger(t *testing.T) {
	input := `
		var v1: Variant := "answer: ";
		var v2: Variant := 42;
		var result: Variant;
		begin
			result := v1 + v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectString(t, result, "answer: 42")
}

// ============================================================================
// Comparison Operation Tests
// ============================================================================

func TestVariantComparisonIntegerEquals(t *testing.T) {
	input := `
		var v1: Variant := 42;
		var v2: Variant := 42;
		var result: Boolean;
		begin
			result := v1 = v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonIntegerNotEquals(t *testing.T) {
	input := `
		var v1: Variant := 42;
		var v2: Variant := 43;
		var result: Boolean;
		begin
			result := v1 <> v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonIntegerLessThan(t *testing.T) {
	input := `
		var v1: Variant := 10;
		var v2: Variant := 20;
		var result: Boolean;
		begin
			result := v1 < v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonIntegerGreaterThan(t *testing.T) {
	input := `
		var v1: Variant := 30;
		var v2: Variant := 20;
		var result: Boolean;
		begin
			result := v1 > v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonFloatEquals(t *testing.T) {
	input := `
		var v1: Variant := 3.14;
		var v2: Variant := 3.14;
		var result: Boolean;
		begin
			result := v1 = v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonStringEquals(t *testing.T) {
	input := `
		var v1: Variant := "hello";
		var v2: Variant := "hello";
		var result: Boolean;
		begin
			result := v1 = v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonStringLessThan(t *testing.T) {
	input := `
		var v1: Variant := "apple";
		var v2: Variant := "banana";
		var result: Boolean;
		begin
			result := v1 < v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonBooleanEquals(t *testing.T) {
	input := `
		var v1: Variant := true;
		var v2: Variant := true;
		var result: Boolean;
		begin
			result := v1 = v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestVariantArithmeticIncompatibleTypes(t *testing.T) {
	input := `
		var v1: Variant := "hello";
		var v2: Variant := true;
		var result: Variant;
		begin
			result := v1 * v2;
		end.
	`
	expectRuntimeError(t, input)
}

func TestVariantArithmeticUnassignedVariant(t *testing.T) {
	input := `
		var v1: Variant;
		var v2: Variant := 10;
		var result: Variant;
		begin
			result := v1 + v2;
		end.
	`
	expectRuntimeError(t, input)
}

func TestVariantComparisonNilEquals(t *testing.T) {
	input := `
		var v1: Variant;
		var v2: Variant;
		var result: Boolean;
		begin
			result := v1 = v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantComparisonNilNotEquals(t *testing.T) {
	input := `
		var v1: Variant;
		var v2: Variant := 42;
		var result: Boolean;
		begin
			result := v1 <> v2;
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

// ============================================================================
// Array and Complex Type Tests
// ============================================================================

func TestVariantInArray(t *testing.T) {
	input := `
		var arr: array of Variant := [1, "hello", 3.14, true];
		var sum: Integer := 0;
		begin
			// Access array elements (they're Variants)
			sum := 1;  // Just verify it parses and runs
		end.
	`
	testRunProgram(t, input)
}

func TestVariantHeterogeneousArray(t *testing.T) {
	input := `
		var arr: array of Variant;
		begin
			arr := [10, "test", 2.5];
		end.
	`
	testRunProgram(t, input)
}

// ============================================================================
// Variant Introspection Function Tests
// ============================================================================

func TestVarTypeInteger(t *testing.T) {
	input := `
		var v: Variant := 42;
		var typeCode: Integer;
		begin
			typeCode := VarType(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "typeCode")
	expectInteger(t, result, 3) // varInteger = 3
}

func TestVarTypeFloat(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var typeCode: Integer;
		begin
			typeCode := VarType(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "typeCode")
	expectInteger(t, result, 5) // varDouble = 5
}

func TestVarTypeString(t *testing.T) {
	input := `
		var v: Variant := "hello";
		var typeCode: Integer;
		begin
			typeCode := VarType(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "typeCode")
	expectInteger(t, result, 256) // varString = 256
}

func TestVarTypeBoolean(t *testing.T) {
	input := `
		var v: Variant := true;
		var typeCode: Integer;
		begin
			typeCode := VarType(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "typeCode")
	expectInteger(t, result, 11) // varBoolean = 11
}

func TestVarTypeEmpty(t *testing.T) {
	input := `
		var v: Variant;
		var typeCode: Integer;
		begin
			typeCode := VarType(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "typeCode")
	expectInteger(t, result, 0) // varEmpty = 0
}

func TestVarTypeNonVariant(t *testing.T) {
	// VarType should also work with non-Variant values
	input := `
		var i: Integer := 100;
		var typeCode: Integer;
		begin
			typeCode := VarType(i);
		end.
	`
	result := testEvalAndGetVar(t, input, "typeCode")
	expectInteger(t, result, 3) // varInteger = 3
}

func TestVarIsNullUnassigned(t *testing.T) {
	input := `
		var v: Variant;
		var result: Boolean;
		begin
			result := VarIsNull(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVarIsNullAssignedInteger(t *testing.T) {
	input := `
		var v: Variant := 42;
		var result: Boolean;
		begin
			result := VarIsNull(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsNullAssignedString(t *testing.T) {
	input := `
		var v: Variant := "test";
		var result: Boolean;
		begin
			result := VarIsNull(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsNullNonVariant(t *testing.T) {
	// VarIsNull on non-Variant should return false
	input := `
		var i: Integer := 0;
		var result: Boolean;
		begin
			result := VarIsNull(i);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsEmptyUnassigned(t *testing.T) {
	input := `
		var v: Variant;
		var result: Boolean;
		begin
			result := VarIsEmpty(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVarIsEmptyAssigned(t *testing.T) {
	input := `
		var v: Variant := 123;
		var result: Boolean;
		begin
			result := VarIsEmpty(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsNumericInteger(t *testing.T) {
	input := `
		var v: Variant := 42;
		var result: Boolean;
		begin
			result := VarIsNumeric(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVarIsNumericFloat(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var result: Boolean;
		begin
			result := VarIsNumeric(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVarIsNumericString(t *testing.T) {
	input := `
		var v: Variant := "42";
		var result: Boolean;
		begin
			result := VarIsNumeric(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsNumericBoolean(t *testing.T) {
	input := `
		var v: Variant := true;
		var result: Boolean;
		begin
			result := VarIsNumeric(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsNumericUnassigned(t *testing.T) {
	input := `
		var v: Variant;
		var result: Boolean;
		begin
			result := VarIsNumeric(v);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, false)
}

func TestVarIsNumericNonVariant(t *testing.T) {
	// VarIsNumeric should work with non-Variant numeric values
	input := `
		var i: Integer := 100;
		var result: Boolean;
		begin
			result := VarIsNumeric(i);
		end.
	`
	result := testEvalAndGetVar(t, input, "result")
	expectBoolean(t, result, true)
}

func TestVariantIntrospectionCombined(t *testing.T) {
	// Test using multiple introspection functions together
	input := `
		var v1: Variant;
		var v2: Variant := 42;
		var v3: Variant := "hello";
		var results: array of Boolean;
		begin
			results := [
				VarIsNull(v1),           // true
				VarIsNumeric(v2),        // true
				VarIsNumeric(v3),        // false
				VarType(v2) = 3,         // true (varInteger)
				VarType(v3) = 256        // true (varString)
			];
		end.
	`
	testRunProgram(t, input)
}

// ============================================================================
// Variant Conversion Functions
// ============================================================================

func TestVarToStrInteger(t *testing.T) {
	input := `
		var v: Variant := 42;
		var s: String;
		begin
			s := VarToStr(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "s")
	stringVal, ok := val.(*StringValue)
	if !ok {
		t.Fatalf("expected *StringValue, got %T", val)
	}
	if stringVal.Value != "42" {
		t.Errorf("VarToStr(42) wrong. expected=%q, got=%q", "42", stringVal.Value)
	}
}

func TestVarToStrFloat(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var s: String;
		begin
			s := VarToStr(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "s")
	stringVal, ok := val.(*StringValue)
	if !ok {
		t.Fatalf("expected *StringValue, got %T", val)
	}
	if stringVal.Value != "3.14" {
		t.Errorf("VarToStr(3.14) wrong. expected=%q, got=%q", "3.14", stringVal.Value)
	}
}

func TestVarToStrBoolean(t *testing.T) {
	input := `
		var v: Variant := True;
		var s: String;
		begin
			s := VarToStr(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "s")
	stringVal, ok := val.(*StringValue)
	if !ok {
		t.Fatalf("expected *StringValue, got %T", val)
	}
	if stringVal.Value != "true" {
		t.Errorf("VarToStr(True) wrong. expected=%q, got=%q", "true", stringVal.Value)
	}
}

func TestVarToStrEmpty(t *testing.T) {
	input := `
		var v: Variant;
		var s: String;
		begin
			s := VarToStr(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "s")
	stringVal, ok := val.(*StringValue)
	if !ok {
		t.Fatalf("expected *StringValue, got %T", val)
	}
	if stringVal.Value != "" {
		t.Errorf("VarToStr(empty) wrong. expected=%q, got=%q", "", stringVal.Value)
	}
}

func TestVarToIntFromInteger(t *testing.T) {
	input := `
		var v: Variant := 42;
		var i: Integer;
		begin
			i := VarToInt(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "i")
	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", val)
	}
	if intVal.Value != 42 {
		t.Errorf("VarToInt(42) wrong. expected=%d, got=%d", 42, intVal.Value)
	}
}

func TestVarToIntFromFloat(t *testing.T) {
	input := `
		var v: Variant := 3.9;
		var i: Integer;
		begin
			i := VarToInt(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "i")
	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", val)
	}
	if intVal.Value != 3 {
		t.Errorf("VarToInt(3.9) wrong. expected=%d, got=%d", 3, intVal.Value)
	}
}

func TestVarToIntFromBoolean(t *testing.T) {
	input := `
		var v1: Variant := True;
		var v2: Variant := False;
		var i1, i2: Integer;
		begin
			i1 := VarToInt(v1);
			i2 := VarToInt(v2);
		end.
	`
	i1 := testEvalAndGetVar(t, input, "i1")
	intVal1, ok := i1.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", i1)
	}
	if intVal1.Value != 1 {
		t.Errorf("VarToInt(True) wrong. expected=%d, got=%d", 1, intVal1.Value)
	}

	i2 := testEvalAndGetVar(t, input, "i2")
	intVal2, ok := i2.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", i2)
	}
	if intVal2.Value != 0 {
		t.Errorf("VarToInt(False) wrong. expected=%d, got=%d", 0, intVal2.Value)
	}
}

func TestVarToIntFromString(t *testing.T) {
	input := `
		var v: Variant := '123';
		var i: Integer;
		begin
			i := VarToInt(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "i")
	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", val)
	}
	if intVal.Value != 123 {
		t.Errorf("VarToInt('123') wrong. expected=%d, got=%d", 123, intVal.Value)
	}
}

func TestVarToIntFromStringError(t *testing.T) {
	input := `
		var v: Variant := 'abc';
		var i: Integer;
		begin
			i := VarToInt(v);
		end.
	`
	val := testEval(input)
	if !isError(val) {
		t.Fatalf("expected error, got %T: %v", val, val)
	}
	if !strings.Contains(val.String(), "cannot convert string 'abc' to Integer") {
		t.Errorf("wrong error message. got=%s", val.String())
	}
}

func TestVarToIntFromEmpty(t *testing.T) {
	input := `
		var v: Variant;
		var i: Integer;
		begin
			i := VarToInt(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "i")
	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", val)
	}
	if intVal.Value != 0 {
		t.Errorf("VarToInt(empty) wrong. expected=%d, got=%d", 0, intVal.Value)
	}
}

func TestVarToFloatFromFloat(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var f: Float;
		begin
			f := VarToFloat(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "f")
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", val)
	}
	if floatVal.Value != 3.14 {
		t.Errorf("VarToFloat(3.14) wrong. expected=%f, got=%f", 3.14, floatVal.Value)
	}
}

func TestVarToFloatFromInteger(t *testing.T) {
	input := `
		var v: Variant := 42;
		var f: Float;
		begin
			f := VarToFloat(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "f")
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", val)
	}
	if floatVal.Value != 42.0 {
		t.Errorf("VarToFloat(42) wrong. expected=%f, got=%f", 42.0, floatVal.Value)
	}
}

func TestVarToFloatFromBoolean(t *testing.T) {
	input := `
		var v1: Variant := True;
		var v2: Variant := False;
		var f1, f2: Float;
		begin
			f1 := VarToFloat(v1);
			f2 := VarToFloat(v2);
		end.
	`
	f1 := testEvalAndGetVar(t, input, "f1")
	floatVal1, ok := f1.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", f1)
	}
	if floatVal1.Value != 1.0 {
		t.Errorf("VarToFloat(True) wrong. expected=%f, got=%f", 1.0, floatVal1.Value)
	}

	f2 := testEvalAndGetVar(t, input, "f2")
	floatVal2, ok := f2.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", f2)
	}
	if floatVal2.Value != 0.0 {
		t.Errorf("VarToFloat(False) wrong. expected=%f, got=%f", 0.0, floatVal2.Value)
	}
}

func TestVarToFloatFromString(t *testing.T) {
	input := `
		var v: Variant := '2.71';
		var f: Float;
		begin
			f := VarToFloat(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "f")
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", val)
	}
	if floatVal.Value != 2.71 {
		t.Errorf("VarToFloat('2.71') wrong. expected=%f, got=%f", 2.71, floatVal.Value)
	}
}

func TestVarToFloatFromStringError(t *testing.T) {
	input := `
		var v: Variant := 'not a number';
		var f: Float;
		begin
			f := VarToFloat(v);
		end.
	`
	val := testEval(input)
	if !isError(val) {
		t.Fatalf("expected error, got %T: %v", val, val)
	}
	if !strings.Contains(val.String(), "cannot convert string") {
		t.Errorf("wrong error message. got=%s", val.String())
	}
}

func TestVarToFloatFromEmpty(t *testing.T) {
	input := `
		var v: Variant;
		var f: Float;
		begin
			f := VarToFloat(v);
		end.
	`
	val := testEvalAndGetVar(t, input, "f")
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", val)
	}
	if floatVal.Value != 0.0 {
		t.Errorf("VarToFloat(empty) wrong. expected=%f, got=%f", 0.0, floatVal.Value)
	}
}

func TestVarAsTypeToInteger(t *testing.T) {
	input := `
		var v: Variant := '42';
		var result: Variant;
		var i: Integer;
		begin
			result := VarAsType(v, 3);  // 3 = varInteger
			i := VarToInt(result);
		end.
	`
	val := testEvalAndGetVar(t, input, "i")
	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", val)
	}
	if intVal.Value != 42 {
		t.Errorf("VarAsType('42', varInteger) wrong. expected=%d, got=%d", 42, intVal.Value)
	}
}

func TestVarAsTypeToFloat(t *testing.T) {
	input := `
		var v: Variant := 42;
		var result: Variant;
		var f: Float;
		begin
			result := VarAsType(v, 5);  // 5 = varDouble
			f := VarToFloat(result);
		end.
	`
	val := testEvalAndGetVar(t, input, "f")
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected *FloatValue, got %T", val)
	}
	if floatVal.Value != 42.0 {
		t.Errorf("VarAsType(42, varDouble) wrong. expected=%f, got=%f", 42.0, floatVal.Value)
	}
}

func TestVarAsTypeToString(t *testing.T) {
	input := `
		var v: Variant := 42;
		var result: Variant;
		var s: String;
		begin
			result := VarAsType(v, 256);  // 256 = varString
			s := VarToStr(result);
		end.
	`
	val := testEvalAndGetVar(t, input, "s")
	stringVal, ok := val.(*StringValue)
	if !ok {
		t.Fatalf("expected *StringValue, got %T", val)
	}
	if stringVal.Value != "42" {
		t.Errorf("VarAsType(42, varString) wrong. expected=%q, got=%q", "42", stringVal.Value)
	}
}

func TestVarAsTypeToBoolean(t *testing.T) {
	input := `
		var v1: Variant := 1;
		var v2: Variant := 0;
		var i1, i2: Integer;
		begin
			i1 := VarToInt(VarAsType(v1, 11));  // 11 = varBoolean, then convert to int
			i2 := VarToInt(VarAsType(v2, 11));
		end.
	`
	r1 := testEvalAndGetVar(t, input, "i1")
	intVal1, ok := r1.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", r1)
	}
	if intVal1.Value != 1 {
		t.Errorf("VarAsType(1, varBoolean) wrong. expected=1, got=%d", intVal1.Value)
	}

	r2 := testEvalAndGetVar(t, input, "i2")
	intVal2, ok := r2.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", r2)
	}
	if intVal2.Value != 0 {
		t.Errorf("VarAsType(0, varBoolean) wrong. expected=0, got=%d", intVal2.Value)
	}
}

func TestVarAsTypeEmptyToInteger(t *testing.T) {
	input := `
		var v: Variant;
		var result: Variant;
		var i: Integer;
		begin
			result := VarAsType(v, 3);  // 3 = varInteger
			i := VarToInt(result);
		end.
	`
	val := testEvalAndGetVar(t, input, "i")
	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Fatalf("expected *IntegerValue, got %T", val)
	}
	if intVal.Value != 0 {
		t.Errorf("VarAsType(empty, varInteger) wrong. expected=%d, got=%d", 0, intVal.Value)
	}
}

func TestVarAsTypeInvalidTypeCode(t *testing.T) {
	input := `
		var v: Variant := 42;
		var result: Variant;
		begin
			result := VarAsType(v, 999);  // invalid type code
		end.
	`
	val := testEval(input)
	if !isError(val) {
		t.Fatalf("expected error, got %T: %v", val, val)
	}
	if !strings.Contains(val.String(), "unsupported VarType code") {
		t.Errorf("wrong error message. got=%s", val.String())
	}
}

func TestVariantConversionCombined(t *testing.T) {
	input := `
		var v1: Variant := 42;
		var v2: Variant := 3.14;
		var v3: Variant := 'hello';
		begin
			PrintLn(VarToStr(v1));          // "42"
			PrintLn(VarToInt(v2));          // 3
			PrintLn(VarToFloat(v1));        // 42.0
			PrintLn(VarToStr(VarAsType(v1, 256)));  // "42"
		end.
	`
	testRunProgram(t, input)
}

// ============================================================================
// Helper Functions
// ============================================================================

func testEvalAndGetVar(t *testing.T, source, varName string) Value {
	t.Helper()
	val := testEval(source)
	if isError(val) {
		t.Fatalf("unexpected runtime error: %v", val)
	}

	// For Variant tests, we need to access the interpreter's environment
	// Use a different approach: evaluate and return result
	interp := testGetInterpreter(source)
	envVal, ok := interp.env.Get(varName)
	if !ok {
		t.Fatalf("variable %q not found in environment", varName)
	}

	// Unwrap if it's a Variant
	if variant, ok := envVal.(*VariantValue); ok {
		if variant.Value != nil {
			return variant.Value
		}
		return &NilValue{}
	}

	return envVal
}

func testGetInterpreter(source string) *Interpreter {
	interp := New(nil)
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		panic("parser errors")
	}
	interp.Eval(program)
	return interp
}

func expectInteger(t *testing.T, val Value, expected int64) {
	t.Helper()
	testIntegerValue(t, val, expected)
}

// ============================================================================
// VarClear Tests
// ============================================================================

func TestVarClearBasic(t *testing.T) {
	source := `
	var v: Variant := 42;
	var cleared: Variant;
	cleared := VarClear(v);
	PrintLn(VarIsNull(cleared));
	PrintLn(VarType(cleared));
	`
	val := testEval(source)
	if isError(val) {
		t.Fatalf("unexpected error: %v", val)
	}
}

func TestVarClearOnEmptyVariant(t *testing.T) {
	source := `
	var v: Variant;
	var cleared: Variant := VarClear(v);
	PrintLn(VarIsNull(cleared));
	`
	val := testEval(source)
	if isError(val) {
		t.Fatalf("unexpected error: %v", val)
	}
}

func TestVarClearReassignment(t *testing.T) {
	source := `
	var v: Variant := 'hello';
	PrintLn(VarType(v));  // Should be 256 (string)
	v := VarClear(v);
	PrintLn(VarIsNull(v));  // Should be True
	PrintLn(VarType(v));  // Should be 0 (empty)
	v := 99;
	PrintLn(VarType(v));  // Should be 3 (integer)
	`
	val := testEval(source)
	if isError(val) {
		t.Fatalf("unexpected error: %v", val)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func expectFloat(t *testing.T, val Value, expected float64) {
	t.Helper()
	testFloatValue(t, val, expected)
}

func expectFloatClose(t *testing.T, val Value, expected, tolerance float64) {
	t.Helper()
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", val)
	}
	diff := floatVal.Value - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("expected %f (±%f), got %f", expected, tolerance, floatVal.Value)
	}
}

func expectString(t *testing.T, val Value, expected string) {
	t.Helper()
	testStringValue(t, val, expected)
}

func expectBoolean(t *testing.T, val Value, expected bool) {
	t.Helper()
	testBooleanValue(t, val, expected)
}

func expectRuntimeError(t *testing.T, source string) {
	t.Helper()
	val := testEval(source)
	if !isError(val) {
		t.Fatalf("expected runtime error, got %T: %v", val, val)
	}
}

func testRunProgram(t *testing.T, source string) {
	t.Helper()
	val := testEval(source)
	if isError(val) {
		t.Fatalf("unexpected runtime error: %v", val)
	}
}
