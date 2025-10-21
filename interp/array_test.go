package interp

import (
	"testing"

	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// ArrayValue Tests (Task 8.128)
// ============================================================================

// TestArrayValue_Creation tests creating an ArrayValue for both static and dynamic arrays.
func TestArrayValue_Creation(t *testing.T) {
	t.Run("dynamic array creation", func(t *testing.T) {
		// Create a dynamic array type: array of Integer
		elementType := types.INTEGER
		arrayType := types.NewDynamicArrayType(elementType)

		// Create an empty dynamic array
		arr := NewArrayValue(arrayType)

		// Verify Type() returns "ARRAY"
		if arr.Type() != "ARRAY" {
			t.Errorf("expected Type() = 'ARRAY', got '%s'", arr.Type())
		}

		// Verify it's empty
		if len(arr.Elements) != 0 {
			t.Errorf("expected empty array, got %d elements", len(arr.Elements))
		}

		// Verify String() for empty array
		if arr.String() != "[]" {
			t.Errorf("expected String() = '[]', got '%s'", arr.String())
		}
	})

	t.Run("static array creation", func(t *testing.T) {
		// Create a static array type: array[1..5] of Integer
		elementType := types.INTEGER
		lowBound := 1
		highBound := 5
		arrayType := types.NewStaticArrayType(elementType, lowBound, highBound)

		// Create a static array (should be pre-allocated with 5 elements)
		arr := NewArrayValue(arrayType)

		// Verify Type() returns "ARRAY"
		if arr.Type() != "ARRAY" {
			t.Errorf("expected Type() = 'ARRAY', got '%s'", arr.Type())
		}

		// Verify it has 5 elements (initialized to nil/zero values)
		expectedSize := highBound - lowBound + 1
		if len(arr.Elements) != expectedSize {
			t.Errorf("expected %d elements, got %d", expectedSize, len(arr.Elements))
		}
	})
}

// TestArrayValue_WithElements tests ArrayValue with pre-set elements.
func TestArrayValue_WithElements(t *testing.T) {
	// Create a dynamic array type: array of Integer
	elementType := types.INTEGER
	arrayType := types.NewDynamicArrayType(elementType)

	// Create an array with some elements
	arr := &ArrayValue{
		ArrayType: arrayType,
		Elements: []Value{
			&IntegerValue{Value: 10},
			&IntegerValue{Value: 20},
			&IntegerValue{Value: 30},
		},
	}

	// Verify Type()
	if arr.Type() != "ARRAY" {
		t.Errorf("expected Type() = 'ARRAY', got '%s'", arr.Type())
	}

	// Verify String() shows elements
	str := arr.String()
	expected := "[10, 20, 30]"
	if str != expected {
		t.Errorf("expected String() = '%s', got '%s'", expected, str)
	}

	// Verify element count
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

// ============================================================================
// Array Indexing Tests (Task 8.129 - Reading)
// ============================================================================

// TestArrayDeclaration_Basic tests that array type declarations work.
func TestArrayDeclaration_Basic(t *testing.T) {
	input := `
		type TIntArray = array[0..2] of Integer;
	`

	result := testEval(input)
	// Type declarations return nil
	if _, ok := result.(*NilValue); !ok {
		t.Errorf("expected NilValue from type declaration, got %T: %v", result, result)
	}
}

// TestArrayIndexing_StaticArray_NilValues tests reading from static arrays (initially nil).
func TestArrayIndexing_StaticArray_NilValues(t *testing.T) {
	// For now, just test that we can index a static array
	// Arrays are pre-allocated with nil values
	input := `
		type TIntArray = array[0..2] of Integer;
		var arr: TIntArray;
		arr[0]
	`

	result := testEval(input)
	// Should return nil initially (or we could default to zero)
	if result == nil {
		t.Errorf("expected a value, got nil")
	}
	// Accept either nil or zero value
	if _, ok := result.(*NilValue); ok {
		// nil is acceptable
		return
	}
	if intVal, ok := result.(*IntegerValue); ok && intVal.Value == 0 {
		// zero is acceptable
		return
	}
	t.Errorf("expected nil or zero value, got %v", result)
}

// TestArrayIndexing_DynamicArray tests dynamic array indexing.
func TestArrayIndexing_DynamicArray(t *testing.T) {
	input := `
		type TDynArray = array of String;
		var arr: TDynArray;
		arr[0]
	`

	result := testEval(input)
	// Dynamic arrays start empty, so indexing should error
	if _, ok := result.(*ErrorValue); !ok {
		t.Errorf("expected error for indexing empty dynamic array, got %T", result)
	}
}

// TestArrayIndexing_WithExpressionIndex tests indexing with expressions.
func TestArrayIndexing_WithExpressionIndex(t *testing.T) {
	input := `
		type TIntArray = array[0..5] of Integer;
		var arr: TIntArray;
		var i: Integer := 2;
		arr[i + 1]
	`

	result := testEval(input)
	// Should work and return nil
	if _, ok := result.(*NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

// TestArrayIndexing_OutOfBoundsStatic tests bounds checking for static arrays.
func TestArrayIndexing_OutOfBoundsStatic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "index below low bound",
			input: `
				type TArr = array[1..3] of Integer;
				var arr: TArr;
				arr[0]
			`,
		},
		{
			name: "index above high bound",
			input: `
				type TArr = array[1..3] of Integer;
				var arr: TArr;
				arr[10]
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			if _, ok := result.(*ErrorValue); !ok {
				t.Errorf("expected error for out of bounds access, got %T", result)
			}
		})
	}
}

// ============================================================================
// Array Index Assignment Tests (Task 8.139)
// ============================================================================

// TestArrayIndexAssignment_Static tests array index assignment with static arrays
func TestArrayIndexAssignment_Static(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Assign to first element",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[1] := 100;
	arr[1];
end
			`,
			expected: 100,
		},
		{
			name: "Assign to middle element",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[3] := 42;
	arr[3];
end
			`,
			expected: 42,
		},
		{
			name: "Assign to last element",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[5] := 99;
	arr[5];
end
			`,
			expected: 99,
		},
		{
			name: "Multiple assignments",
			input: `
type TMyArray = array[1..3] of Integer;
var arr: TMyArray;
var sum: Integer;
begin
	arr[1] := 10;
	arr[2] := 20;
	arr[3] := 30;
	sum := arr[1] + arr[2] + arr[3];
	sum;
end
			`,
			expected: 60,
		},
		{
			name: "Assignment with variable index",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
var i: Integer;
begin
	i := 2;
	arr[i] := 77;
	arr[2];
end
			`,
			expected: 77,
		},
		{
			name: "Assignment with expression index",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
var i: Integer;
begin
	i := 2;
	arr[i + 1] := 88;
	arr[3];
end
			`,
			expected: 88,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("value = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestArrayIndexAssignment_BoundsChecking tests bounds checking for array assignments
func TestArrayIndexAssignment_BoundsChecking(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Static array - index too low",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[0] := 42;
end
			`,
		},
		{
			name: "Static array - index too high",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[6] := 42;
end
			`,
		},
		{
			name: "Static array - negative index",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[-1] := 42;
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			if _, ok := result.(*ErrorValue); !ok {
				t.Errorf("expected error for out of bounds assignment, got %T", result)
			}
		})
	}
}

// TestArrayIndexAssignment_InLoop tests array assignment within loops
func TestArrayIndexAssignment_InLoop(t *testing.T) {
	input := `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
var i: Integer;
var sum: Integer;
begin
	// Fill array with values
	for i := 1 to 5 do
		arr[i] := i * 10;

	// Sum all values
	sum := 0;
	for i := 1 to 5 do
		sum := sum + arr[i];

	sum;
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Expected: 10 + 20 + 30 + 40 + 50 = 150
	expected := int64(150)
	if intVal.Value != expected {
		t.Errorf("sum = %d, want %d", intVal.Value, expected)
	}
}

// ============================================================================
// Comprehensive Array Assignment Tests (Task 8.140)
// ============================================================================

// TestArrayAssignment_StaticArraysDetailed tests comprehensive scenarios with static arrays
func TestArrayAssignment_StaticArraysDetailed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Zero-indexed static array assignment",
			input: `
type TMyArray = array[0..4] of Integer;
var arr: TMyArray;
begin
	arr[0] := 100;
	arr[4] := 500;
	arr[0] + arr[4];
end
			`,
			expected: 600,
		},
		{
			name: "Large index range",
			input: `
type TMyArray = array[100..105] of Integer;
var arr: TMyArray;
begin
	arr[100] := 10;
	arr[105] := 20;
	arr[100] + arr[105];
end
			`,
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("value = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestArrayAssignment_ComplexExpressions tests assignment with complex index expressions
func TestArrayAssignment_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Assignment with arithmetic index",
			input: `
type TMyArray = array[1..10] of Integer;
var arr: TMyArray;
var i: Integer;
begin
	i := 5;
	arr[i * 2] := 999;
	arr[10];
end
			`,
			expected: 999,
		},
		{
			name: "Assignment with function result as value",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;

function Double(x: Integer): Integer;
begin
	Result := x * 2;
end;

begin
	arr[1] := Double(21);
	arr[1];
end
			`,
			expected: 42,
		},
		{
			name: "Chain assignments using array elements",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[1] := 10;
	arr[2] := arr[1] + 5;
	arr[3] := arr[1] + arr[2];
	arr[3];
end
			`,
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("value = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestArrayAssignment_WithRecords tests arrays of records with field assignment
func TestArrayAssignment_WithRecords(t *testing.T) {
	input := `
type TPoint = record
	x: Integer;
	y: Integer;
end;

type TPoints = array[1..3] of TPoint;

var points: TPoints;
var p: TPoint;
begin
	// Create a point
	p.x := 10;
	p.y := 20;

	// Assign to array
	points[1] := p;

	// Modify through array
	points[2].x := 30;
	points[2].y := 40;

	// Read back
	points[1].x + points[2].x;
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Expected: 10 + 30 = 40
	expected := int64(40)
	if intVal.Value != expected {
		t.Errorf("value = %d, want %d", intVal.Value, expected)
	}
}

// TestArrayAssignment_EdgeCases tests edge cases and error conditions
func TestArrayAssignment_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name: "Reassign array element multiple times",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[1] := 1;
	arr[1] := 2;
	arr[1] := 3;
	arr[1] := 4;
	arr[1] := 5;
	arr[1];
end
			`,
			shouldError: false,
		},
		{
			name: "Assign to array element zero for 1-indexed array",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	arr[0] := 42;
end
			`,
			shouldError: true,
		},
		{
			name: "Negative index on static array",
			input: `
type TMyArray = array[0..5] of Integer;
var arr: TMyArray;
begin
	arr[-5] := 42;
end
			`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			if tt.shouldError {
				if _, ok := result.(*ErrorValue); !ok {
					t.Errorf("expected error, got %T: %+v", result, result)
				}
			} else {
				if _, ok := result.(*ErrorValue); ok {
					t.Errorf("unexpected error: %+v", result)
				}
			}
		})
	}
}

// ============================================================================
// Length() Built-in Function Tests (Task 8.130)
// ============================================================================

// TestBuiltinLength_StaticArrays tests Length() with static arrays of various bounds.
func TestBuiltinLength_StaticArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Static array [1..5] returns 5",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	Length(arr);
end
			`,
			expected: 5,
		},
		{
			name: "Static array [0..9] returns 10",
			input: `
type TMyArray = array[0..9] of Integer;
var arr: TMyArray;
begin
	Length(arr);
end
			`,
			expected: 10,
		},
		{
			name: "Static array [1..10] returns 10",
			input: `
type TMyArray = array[1..10] of Integer;
var arr: TMyArray;
begin
	Length(arr);
end
			`,
			expected: 10,
		},
		{
			name: "Static array [100..105] returns 6",
			input: `
type TMyArray = array[100..105] of Integer;
var arr: TMyArray;
begin
	Length(arr);
end
			`,
			expected: 6,
		},
		{
			name: "Static array [0..0] returns 1",
			input: `
type TMyArray = array[0..0] of Integer;
var arr: TMyArray;
begin
	Length(arr);
end
			`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("Length() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLength_DynamicArrays tests Length() with dynamic arrays.
func TestBuiltinLength_DynamicArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Empty dynamic array returns 0",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Length(arr);
end
			`,
			expected: 0,
		},
		// Note: We don't have SetLength implemented yet (task 8.131),
		// so we can't test non-empty dynamic arrays yet
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("Length() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLength_Strings tests Length() with string values.
func TestBuiltinLength_Strings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Empty string returns 0",
			input: `
var s: String := "";
begin
	Length(s);
end
			`,
			expected: 0,
		},
		{
			name: "String 'hello' returns 5",
			input: `
var s: String := "hello";
begin
	Length(s);
end
			`,
			expected: 5,
		},
		{
			name: "String 'DWScript' returns 8",
			input: `
var s: String := "DWScript";
begin
	Length(s);
end
			`,
			expected: 8,
		},
		{
			name: "String literal directly",
			input: `
begin
	Length("test");
end
			`,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("Length() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLength_InExpressions tests using Length() in expressions.
func TestBuiltinLength_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Length() in arithmetic expression",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	Length(arr) * 2;
end
			`,
			expected: 10,
		},
		{
			name: "Length() in for loop (Length - 1)",
			input: `
type TMyArray = array[0..4] of Integer;
var arr: TMyArray;
var i: Integer;
var count: Integer;
begin
	count := 0;
	for i := 0 to Length(arr) - 1 do
		count := count + 1;
	count;
end
			`,
			expected: 5,
		},
		{
			name: "Compare Length() result",
			input: `
type TMyArray = array[1..10] of Integer;
var arr: TMyArray;
var result: Integer;
begin
	if Length(arr) = 10 then
		result := 1
	else
		result := 0;
	result;
end
			`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("result = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLength_ErrorCases tests error handling for Length().
func TestBuiltinLength_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Length();
end
			`,
		},
		{
			name: "Multiple arguments",
			input: `
type TMyArray = array[1..5] of Integer;
var arr1: TMyArray;
var arr2: TMyArray;
begin
	Length(arr1, arr2);
end
			`,
		},
		{
			name: "Invalid argument type (integer)",
			input: `
var x: Integer := 42;
begin
	Length(x);
end
			`,
		},
		{
			name: "Invalid argument type (boolean)",
			input: `
var b: Boolean := true;
begin
	Length(b);
end
			`,
		},
		{
			name: "Invalid argument type (float)",
			input: `
var f: Float := 3.14;
begin
	Length(f);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			if _, ok := result.(*ErrorValue); !ok {
				t.Errorf("expected error for invalid Length() call, got %T: %+v", result, result)
			}
		})
	}
}
