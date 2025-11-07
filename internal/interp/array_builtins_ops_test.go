package interp

import (
	"strings"
	"testing"
)

// ============================================================================
// ============================================================================
// Array Copy Tests
// ============================================================================

// TestArrayCopy_DynamicArray tests copying a dynamic array and verifying mutation isolation
func TestArrayCopy_DynamicArray(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var a1: TDynArray;
var a2: TDynArray;
begin
	Add(a1, 1);
	Add(a1, 2);
	Add(a1, 3);
	a2 := Copy(a1);
	a2[0] := 99;
	a1[0];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Original array should be unchanged (a1[0] = 1)
	if intVal.Value != 1 {
		t.Errorf("expected a1[0] = 1 (unchanged), got %d", intVal.Value)
	}

	// Verify the copy was modified
	input2 := `
type TDynArray = array of Integer;
var a1: TDynArray;
var a2: TDynArray;
begin
	Add(a1, 1);
	Add(a1, 2);
	Add(a1, 3);
	a2 := Copy(a1);
	a2[0] := 99;
	a2[0];
end
	`

	result2 := testEval(input2)
	intVal2, ok := result2.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result2, result2)
	}

	// Copy should have new value (a2[0] = 99)
	if intVal2.Value != 99 {
		t.Errorf("expected a2[0] = 99 (modified), got %d", intVal2.Value)
	}
}

// TestArrayCopy_StaticArray tests copying a static array
func TestArrayCopy_StaticArray(t *testing.T) {
	input := `
type TStaticArray = array[1..3] of Integer;
var a1: TStaticArray;
var a2: TStaticArray;
begin
	a1[1] := 10;
	a1[2] := 20;
	a1[3] := 30;
	a2 := Copy(a1);
	a2[1] := 999;
	a1[1];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Original array should be unchanged (a1[1] = 10)
	if intVal.Value != 10 {
		t.Errorf("expected a1[1] = 10 (unchanged), got %d", intVal.Value)
	}
}

// TestArrayCopy_PreservesElementTypes tests that copy preserves element types
func TestArrayCopy_PreservesElementTypes(t *testing.T) {
	tests := []struct {
		expected any
		name     string
		input    string
	}{
		{
			name: "Integer array",
			input: `
type TIntArray = array of Integer;
var a1: TIntArray;
var a2: TIntArray;
begin
	Add(a1, 42);
	a2 := Copy(a1);
	a2[0];
end
			`,
			expected: int64(42),
		},
		{
			name: "String array",
			input: `
type TStrArray = array of String;
var a1: TStrArray;
var a2: TStrArray;
begin
	Add(a1, "hello");
	a2 := Copy(a1);
	a2[0];
end
			`,
			expected: "hello",
		},
		{
			name: "Float array",
			input: `
type TFloatArray = array of Float;
var a1: TFloatArray;
var a2: TFloatArray;
begin
	Add(a1, 3.14);
	a2 := Copy(a1);
	a2[0];
end
			`,
			expected: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			switch exp := tt.expected.(type) {
			case int64:
				intVal, ok := result.(*IntegerValue)
				if !ok {
					t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
				}
				if intVal.Value != exp {
					t.Errorf("expected %d, got %d", exp, intVal.Value)
				}
			case string:
				strVal, ok := result.(*StringValue)
				if !ok {
					t.Fatalf("expected StringValue, got %T: %+v", result, result)
				}
				if strVal.Value != exp {
					t.Errorf("expected %s, got %s", exp, strVal.Value)
				}
			case float64:
				floatVal, ok := result.(*FloatValue)
				if !ok {
					t.Fatalf("expected FloatValue, got %T: %+v", result, result)
				}
				if floatVal.Value != exp {
					t.Errorf("expected %f, got %f", exp, floatVal.Value)
				}
			}
		})
	}
}

// TestArrayCopy_EmptyArray tests copying an empty array
func TestArrayCopy_EmptyArray(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var a1: TDynArray;
var a2: TDynArray;
begin
	a2 := Copy(a1);
	Length(a2);
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Empty array should have length 0
	if intVal.Value != 0 {
		t.Errorf("expected length = 0, got %d", intVal.Value)
	}
}

// ============================================================================
// IndexOf Tests
// ============================================================================

// TestArrayIndexOf_BasicFound tests IndexOf finding the first occurrence.
func TestArrayIndexOf_BasicFound(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 4);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	a[3] := 2;
	IndexOf(a, 2);
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Should return 1 (0-based index: value 2 is at a[1])
	if intVal.Value != 1 {
		t.Errorf("expected IndexOf to return 1, got %d", intVal.Value)
	}
}

// TestArrayIndexOf_NotFound tests IndexOf returning -1 when value not found.
func TestArrayIndexOf_NotFound(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	IndexOf(a, 5);
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Should return -1 (not found)
	if intVal.Value != -1 {
		t.Errorf("expected IndexOf to return -1, got %d", intVal.Value)
	}
}

// TestArrayIndexOf_WithStartIndex tests IndexOf with optional startIndex parameter.
func TestArrayIndexOf_WithStartIndex(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 4);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	a[3] := 2;
	IndexOf(a, 2, 2);
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Should return 3 (skips first 2 elements, finds at a[3])
	if intVal.Value != 3 {
		t.Errorf("expected IndexOf to return 3, got %d", intVal.Value)
	}
}

// TestArrayIndexOf_StringArray tests IndexOf with string arrays.
func TestArrayIndexOf_StringArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "string found",
			input: `
type TStringArray = array of String;
var a: TStringArray;
begin
	SetLength(a, 3);
	a[0] := 'a';
	a[1] := 'b';
	a[2] := 'c';
	IndexOf(a, 'b');
end
			`,
			expected: 1,
		},
		{
			name: "string not found",
			input: `
type TStringArray = array of String;
var a: TStringArray;
begin
	SetLength(a, 2);
	a[0] := 'hello';
	a[1] := 'world';
	IndexOf(a, 'foo');
end
			`,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("expected IndexOf to return %d, got %d", tt.expected, intVal.Value)
			}
		})
	}
}

// TestArrayIndexOf_EmptyArray tests IndexOf with an empty array.
func TestArrayIndexOf_EmptyArray(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	IndexOf(a, 42);
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Should return -1 (empty array has no elements)
	if intVal.Value != -1 {
		t.Errorf("expected IndexOf to return -1 for empty array, got %d", intVal.Value)
	}
}

// TestArrayIndexOf_EdgeCases tests IndexOf boundary conditions.
func TestArrayIndexOf_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "startIndex at 0",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	IndexOf(a, 2, 0);
end
			`,
			expected: 1, // Finds at a[1], returns index 1
		},
		{
			name: "negative startIndex",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	IndexOf(a, 2, -1);
end
			`,
			expected: -1, // Invalid index returns -1
		},
		{
			name: "startIndex beyond bounds",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	IndexOf(a, 2, 10);
end
			`,
			expected: -1, // Beyond bounds returns -1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("expected IndexOf to return %d, got %d", tt.expected, intVal.Value)
			}
		})
	}
}

// TestArrayIndexOf_ErrorCases tests IndexOf error handling.
func TestArrayIndexOf_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "wrong argument count (1 arg)",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	IndexOf(a);
end
			`,
			expectedErr: "IndexOf() expects 2 or 3 arguments, got 1",
		},
		{
			name: "non-array first argument",
			input: `
begin
	IndexOf(42, 1);
end
			`,
			expectedErr: "IndexOf() expects array as first argument",
		},
		{
			name: "non-integer third argument",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 2);
	IndexOf(a, 1, 'bad');
end
			`,
			expectedErr: "IndexOf() expects integer as third argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, errVal.Message)
			}
		})
	}
}

// ============================================================================
// Contains Tests
// ============================================================================

// TestArrayContains_Found tests Contains returning true when value exists.
func TestArrayContains_Found(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	Contains(a, 2);
end
	`

	result := testEval(input)
	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("expected BooleanValue, got %T: %+v", result, result)
	}

	// Should return true (value 2 is in the array)
	if !boolVal.Value {
		t.Errorf("expected Contains to return true, got false")
	}
}

// TestArrayContains_NotFound tests Contains returning false when value doesn't exist.
func TestArrayContains_NotFound(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	Contains(a, 5);
end
	`

	result := testEval(input)
	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("expected BooleanValue, got %T: %+v", result, result)
	}

	// Should return false (value 5 is not in the array)
	if boolVal.Value {
		t.Errorf("expected Contains to return false, got true")
	}
}

// TestArrayContains_DifferentTypes tests Contains with different value types.
func TestArrayContains_DifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "string array - found",
			input: `
type TStringArray = array of String;
var a: TStringArray;
begin
	SetLength(a, 3);
	a[0] := 'apple';
	a[1] := 'banana';
	a[2] := 'cherry';
	Contains(a, 'banana');
end
			`,
			expected: true,
		},
		{
			name: "string array - not found",
			input: `
type TStringArray = array of String;
var a: TStringArray;
begin
	SetLength(a, 2);
	a[0] := 'hello';
	a[1] := 'world';
	Contains(a, 'goodbye');
end
			`,
			expected: false,
		},
		{
			name: "float array - found",
			input: `
type TFloatArray = array of Float;
var a: TFloatArray;
begin
	SetLength(a, 3);
	a[0] := 1.5;
	a[1] := 2.5;
	a[2] := 3.5;
	Contains(a, 2.5);
end
			`,
			expected: true,
		},
		{
			name: "boolean array - found",
			input: `
type TBoolArray = array of Boolean;
var a: TBoolArray;
begin
	SetLength(a, 2);
	a[0] := True;
	a[1] := False;
	Contains(a, False);
end
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			boolVal, ok := result.(*BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T: %+v", result, result)
			}

			if boolVal.Value != tt.expected {
				t.Errorf("expected Contains to return %v, got %v", tt.expected, boolVal.Value)
			}
		})
	}
}

// TestArrayContains_EmptyArray tests Contains with empty array.
func TestArrayContains_EmptyArray(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	Contains(a, 42);
end
	`

	result := testEval(input)
	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("expected BooleanValue, got %T: %+v", result, result)
	}

	// Should return false (empty array cannot contain any value)
	if boolVal.Value {
		t.Errorf("expected Contains to return false for empty array, got true")
	}
}

// TestArrayContains_ErrorCases tests Contains error handling.
func TestArrayContains_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "wrong argument count",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	Contains(a);
end
			`,
			expectedErr: "Contains() expects 2 arguments, got 1",
		},
		{
			name: "non-array first argument",
			input: `
begin
	Contains(42, 1);
end
			`,
			expectedErr: "Contains() expects array as first argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, errVal.Message)
			}
		})
	}
}

// ============================================================================
// Reverse Tests
// ============================================================================

// TestArrayReverse_OddLength tests reversing an odd-length array.
func TestArrayReverse_OddLength(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	Reverse(a);

	// Check reversed values
	if a[0] <> 3 then
		PrintLn('FAIL: a[0] expected 3');
	if a[1] <> 2 then
		PrintLn('FAIL: a[1] expected 2');
	if a[2] <> 1 then
		PrintLn('FAIL: a[2] expected 1');

	a[0];  // Return first element to verify
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// After reversal, a[0] should be 3
	if intVal.Value != 3 {
		t.Errorf("expected a[0] = 3 after Reverse, got %d", intVal.Value)
	}
}

// TestArrayReverse_EvenLength tests reversing an even-length array.
func TestArrayReverse_EvenLength(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 4);
	a[0] := 10;
	a[1] := 20;
	a[2] := 30;
	a[3] := 40;
	Reverse(a);

	// After reversal: [40, 30, 20, 10]
	if a[0] <> 40 then
		PrintLn('FAIL: a[0] expected 40');
	if a[1] <> 30 then
		PrintLn('FAIL: a[1] expected 30');
	if a[2] <> 20 then
		PrintLn('FAIL: a[2] expected 20');
	if a[3] <> 10 then
		PrintLn('FAIL: a[3] expected 10');

	a[0];  // Return first element to verify
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// After reversal, a[0] should be 40
	if intVal.Value != 40 {
		t.Errorf("expected a[0] = 40 after Reverse, got %d", intVal.Value)
	}
}

// TestArrayReverse_SingleElement tests reversing a single-element array (no-op).
func TestArrayReverse_SingleElement(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 1);
	a[0] := 42;
	Reverse(a);

	a[0];  // Should still be 42
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Single element should remain unchanged
	if intVal.Value != 42 {
		t.Errorf("expected a[0] = 42 after Reverse, got %d", intVal.Value)
	}
}

// TestArrayReverse_EmptyArray tests reversing an empty array (no-op).
func TestArrayReverse_EmptyArray(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	Reverse(a);
	Length(a);  // Should still be 0
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Empty array should remain empty
	if intVal.Value != 0 {
		t.Errorf("expected length = 0 after Reverse, got %d", intVal.Value)
	}
}

// TestArrayReverse_ErrorCases tests Reverse error handling.
func TestArrayReverse_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "wrong argument count",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	Reverse(a, a);
end
			`,
			expectedErr: "Reverse() expects 1 argument, got 2",
		},
		{
			name: "non-array argument",
			input: `
begin
	Reverse(42);
end
			`,
			expectedErr: "Reverse() expects array as argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, errVal.Message)
			}
		})
	}
}

// ============================================================================
// Sort Tests
// ============================================================================

// TestArraySort_Integers tests sorting an integer array.
func TestArraySort_Integers(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 5);
	a[0] := 3;
	a[1] := 1;
	a[2] := 4;
	a[3] := 2;
	a[4] := 5;
	Sort(a);

	// After sort: [1, 2, 3, 4, 5]
	if a[0] <> 1 then
		PrintLn('FAIL: a[0] expected 1');
	if a[1] <> 2 then
		PrintLn('FAIL: a[1] expected 2');
	if a[2] <> 3 then
		PrintLn('FAIL: a[2] expected 3');
	if a[3] <> 4 then
		PrintLn('FAIL: a[3] expected 4');
	if a[4] <> 5 then
		PrintLn('FAIL: a[4] expected 5');

	a[0];  // Return first element to verify
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// After sort, a[0] should be 1
	if intVal.Value != 1 {
		t.Errorf("expected a[0] = 1 after Sort, got %d", intVal.Value)
	}
}

// TestArraySort_Strings tests sorting a string array.
func TestArraySort_Strings(t *testing.T) {
	input := `
type TStringArray = array of String;
var a: TStringArray;
begin
	SetLength(a, 3);
	a[0] := 'cherry';
	a[1] := 'apple';
	a[2] := 'banana';
	Sort(a);

	// After sort: ['apple', 'banana', 'cherry']
	if a[0] <> 'apple' then
		PrintLn('FAIL: a[0] expected apple');
	if a[1] <> 'banana' then
		PrintLn('FAIL: a[1] expected banana');
	if a[2] <> 'cherry' then
		PrintLn('FAIL: a[2] expected cherry');

	a[0];  // Return first element to verify
end
	`

	result := testEval(input)
	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T: %+v", result, result)
	}

	// After sort, a[0] should be 'apple'
	if strVal.Value != "apple" {
		t.Errorf("expected a[0] = 'apple' after Sort, got %s", strVal.Value)
	}
}

// TestArraySort_AlreadySorted tests sorting an already sorted array (no-op).
func TestArraySort_AlreadySorted(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 4);
	a[0] := 1;
	a[1] := 2;
	a[2] := 3;
	a[3] := 4;
	Sort(a);

	// Should remain: [1, 2, 3, 4]
	if a[0] <> 1 then
		PrintLn('FAIL: a[0] expected 1');
	if a[3] <> 4 then
		PrintLn('FAIL: a[3] expected 4');

	a[0];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Array should remain unchanged
	if intVal.Value != 1 {
		t.Errorf("expected a[0] = 1 after Sort, got %d", intVal.Value)
	}
}

// TestArraySort_SingleElement tests sorting a single-element array (no-op).
func TestArraySort_SingleElement(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 1);
	a[0] := 99;
	Sort(a);

	a[0];  // Should still be 99
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Single element should remain unchanged
	if intVal.Value != 99 {
		t.Errorf("expected a[0] = 99 after Sort, got %d", intVal.Value)
	}
}

// TestArraySort_Duplicates tests sorting an array with duplicate values.
func TestArraySort_Duplicates(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 7);
	a[0] := 3;
	a[1] := 1;
	a[2] := 4;
	a[3] := 1;
	a[4] := 5;
	a[5] := 9;
	a[6] := 2;
	Sort(a);

	// After sort: [1, 1, 2, 3, 4, 5, 9]
	if a[0] <> 1 then
		PrintLn('FAIL: a[0] expected 1');
	if a[1] <> 1 then
		PrintLn('FAIL: a[1] expected 1');
	if a[2] <> 2 then
		PrintLn('FAIL: a[2] expected 2');
	if a[3] <> 3 then
		PrintLn('FAIL: a[3] expected 3');
	if a[4] <> 4 then
		PrintLn('FAIL: a[4] expected 4');
	if a[5] <> 5 then
		PrintLn('FAIL: a[5] expected 5');
	if a[6] <> 9 then
		PrintLn('FAIL: a[6] expected 9');

	a[0];  // Return first element
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// After sort, a[0] should be 1
	if intVal.Value != 1 {
		t.Errorf("expected a[0] = 1 after Sort, got %d", intVal.Value)
	}
}

// TestArraySort_EmptyArray tests sorting an empty array (no-op).
func TestArraySort_EmptyArray(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	Sort(a);
	Length(a);  // Should still be 0
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Empty array should remain empty
	if intVal.Value != 0 {
		t.Errorf("expected length = 0 after Sort, got %d", intVal.Value)
	}
}

// TestArraySort_ErrorCases tests Sort error handling.
func TestArraySort_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "too many arguments (3 args)",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	Sort(a, a, a);
end
			`,
			expectedErr: "Sort() expects 1 or 2 arguments",
		},
		{
			name: "non-array argument",
			input: `
begin
	Sort(42);
end
			`,
			expectedErr: "Sort() expects array as first argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, errVal.Message)
			}
		})
	}
}

// TestArraySort_WithLambdaComparator_Ascending tests Sort() with a lambda comparator for ascending order.
func TestArraySort_WithLambdaComparator_Ascending(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 5);
	a[0] := 3;
	a[1] := 1;
	a[2] := 4;
	a[3] := 2;
	a[4] := 5;

	Sort(a, lambda (x, y): Integer => x - y);  // Ascending

	// After sort: [1, 2, 3, 4, 5]
	if a[0] <> 1 then PrintLn('FAIL: a[0]');
	if a[1] <> 2 then PrintLn('FAIL: a[1]');
	if a[2] <> 3 then PrintLn('FAIL: a[2]');
	if a[3] <> 4 then PrintLn('FAIL: a[3]');
	if a[4] <> 5 then PrintLn('FAIL: a[4]');

	a[0];  // Return first element
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 1 {
		t.Errorf("expected a[0] = 1 after ascending sort, got %d", intVal.Value)
	}
}

// TestArraySort_WithLambdaComparator_Descending tests Sort() with a lambda comparator for descending order.
func TestArraySort_WithLambdaComparator_Descending(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 5);
	a[0] := 3;
	a[1] := 1;
	a[2] := 4;
	a[3] := 2;
	a[4] := 5;

	Sort(a, lambda (x, y): Integer => y - x);  // Descending

	// After sort: [5, 4, 3, 2, 1]
	if a[0] <> 5 then PrintLn('FAIL: a[0]');
	if a[1] <> 4 then PrintLn('FAIL: a[1]');
	if a[2] <> 3 then PrintLn('FAIL: a[2]');
	if a[3] <> 2 then PrintLn('FAIL: a[3]');
	if a[4] <> 1 then PrintLn('FAIL: a[4]');

	a[0];  // Return first element
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 5 {
		t.Errorf("expected a[0] = 5 after descending sort, got %d", intVal.Value)
	}
}

// TestArraySort_WithLambdaComparator_CustomLogic tests Sort() with custom sorting logic.
func TestArraySort_WithLambdaComparator_CustomLogic(t *testing.T) {
	input := `
type TStringArray = array of String;
var a: TStringArray;
begin
	SetLength(a, 4);
	a[0] := 'apple';
	a[1] := 'hi';
	a[2] := 'banana';
	a[3] := 'cat';

	// Sort by string length: shorter strings first
	Sort(a, lambda (x, y): Integer => Length(x) - Length(y));

	// After sort by length: ['hi', 'cat', 'apple', 'banana']
	if a[0] <> 'hi' then PrintLn('FAIL: a[0]');
	if a[1] <> 'cat' then PrintLn('FAIL: a[1]');
	if a[2] <> 'apple' then PrintLn('FAIL: a[2]');
	if a[3] <> 'banana' then PrintLn('FAIL: a[3]');

	a[0];  // Return first element
end
	`

	result := testEval(input)
	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T: %+v", result, result)
	}

	if strVal.Value != "hi" {
		t.Errorf("expected a[0] = 'hi' after sort by length, got '%s'", strVal.Value)
	}
}

// TestArraySort_WithNamedFunctionComparator tests Sort() with a named function pointer.
func TestArraySort_WithNamedFunctionComparator(t *testing.T) {
	input := `
type TIntArray = array of Integer;

// Comparator function for descending order
function CompareDesc(a: Integer; b: Integer): Integer;
begin
	Result := b - a;  // Descending
end;

var arr: TIntArray;
begin
	SetLength(arr, 5);
	arr[0] := 3;
	arr[1] := 1;
	arr[2] := 4;
	arr[3] := 2;
	arr[4] := 5;

	Sort(arr, @CompareDesc);  // Named function pointer

	// After sort: [5, 4, 3, 2, 1]
	if arr[0] <> 5 then PrintLn('FAIL: arr[0]');
	if arr[1] <> 4 then PrintLn('FAIL: arr[1]');
	if arr[2] <> 3 then PrintLn('FAIL: arr[2]');
	if arr[3] <> 2 then PrintLn('FAIL: arr[3]');
	if arr[4] <> 1 then PrintLn('FAIL: arr[4]');

	arr[0];  // Return first element
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 5 {
		t.Errorf("expected arr[0] = 5 after descending sort with named function, got %d", intVal.Value)
	}
}

// TestArraySort_WithComparator_ErrorCases tests error handling for Sort() with comparators.
func TestArraySort_WithComparator_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "too many parameters in comparator",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 3; a[1] := 1; a[2] := 2;
	Sort(a, lambda (x, y, z) => x - y);
end
			`,
			expectedErr: "comparator must accept 2 parameters",
		},
		{
			name: "too few parameters in comparator",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 3; a[1] := 1; a[2] := 2;
	Sort(a, lambda (x) => x);
end
			`,
			expectedErr: "comparator must accept 2 parameters",
		},
		{
			name: "comparator returns wrong type",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 3; a[1] := 1; a[2] := 2;
	Sort(a, lambda (x, y) => 'string');  // Returns String instead of Integer
end
			`,
			expectedErr: "comparator must return Integer",
		},
		{
			name: "second argument is not a function pointer",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 3; a[1] := 1; a[2] := 2;
	Sort(a, 42);  // Integer instead of function pointer
end
			`,
			expectedErr: "function pointer as second argument",
		},
		{
			name: "too many arguments",
			input: `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 3);
	a[0] := 3; a[1] := 1; a[2] := 2;
	Sort(a, lambda (x, y) => x - y, 42);  // 3 arguments
end
			`,
			expectedErr: "expects 1 or 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedErr) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, errVal.Message)
			}
		})
	}
}

// TestArraySort_WithComparator_EmptyArray tests Sort() with comparator on empty array.
func TestArraySort_WithComparator_EmptyArray(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 0);
	Sort(a, lambda (x, y): Integer => x - y);  // Should handle empty array gracefully
	Length(a);  // Return length (should be 0)
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 0 {
		t.Errorf("expected length 0 for empty array, got %d", intVal.Value)
	}
}

// TestArraySort_WithComparator_SingleElement tests Sort() with comparator on single element array.
func TestArraySort_WithComparator_SingleElement(t *testing.T) {
	input := `
type TIntArray = array of Integer;
var a: TIntArray;
begin
	SetLength(a, 1);
	a[0] := 42;
	Sort(a, lambda (x, y): Integer => x - y);  // Should handle single element gracefully
	a[0];  // Return element
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 42 {
		t.Errorf("expected a[0] = 42, got %d", intVal.Value)
	}
}
