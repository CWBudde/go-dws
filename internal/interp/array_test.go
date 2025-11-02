package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// ArrayValue Tests
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
// Array Indexing Tests
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

// ============================================================================//
// Array Literal Evaluation Tests
// ============================================================================//

func TestArrayLiteralEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple integer literal",
			input: `
				var arr := [1, 2, 3];
				PrintLn(arr[1]);
			`,
			expected: "2",
		},
		{
			name: "numeric promotion to float",
			input: `
				var vec := [1, 2.5, 3];
				PrintLn(vec[1]);
			`,
			expected: "2.5",
		},
		{
			name: "float literal access",
			input: `
				var light := [-50.0, 30, 50];
				PrintLn(FloatToStr(light[0]));
			`,
			expected: "-50",
		},
		{
			name: "nested arrays",
			input: `
				var matrix := [[1, 2], [3, 4]];
				PrintLn(matrix[0][1]);
			`,
			expected: "2",
		},
		{
			name: "empty array with length helper",
			input: `
				var arrEmpty: array of Integer;
				begin
					arrEmpty := [];
					PrintLn(Length(arrEmpty));
				end.
			`,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, output := testEvalWithOutput(tt.input)
			if isError(val) {
				t.Fatalf("execution returned error value: %v", val)
			}
			got := strings.TrimSpace(output)
			if got != tt.expected {
				t.Fatalf("expected output %q, got %q", tt.expected, got)
			}
		})
	}
}

// ============================================================================
// Array Index Assignment Tests
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
// Comprehensive Array Assignment Tests
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
// Length() Built-in Function Tests
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
		// Note: We don't have SetLength implemented yet,
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

// ============================================================================
// Low() Built-in Function Tests
// ============================================================================

// TestBuiltinLow_StaticArrays tests Low() with static arrays.
func TestBuiltinLow_StaticArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Static array [1..5] returns 1",
			input: `
type TMyArray = array[1..5] of Integer;
var arr: TMyArray;
begin
	Low(arr);
end
			`,
			expected: 1,
		},
		{
			name: "Static array [0..9] returns 0",
			input: `
type TMyArray = array[0..9] of Integer;
var arr: TMyArray;
begin
	Low(arr);
end
			`,
			expected: 0,
		},
		{
			name: "Static array [100..105] returns 100",
			input: `
type TMyArray = array[100..105] of Integer;
var arr: TMyArray;
begin
	Low(arr);
end
			`,
			expected: 100,
		},
		{
			name: "Static array [10..20] returns 10",
			input: `
type TMyArray = array[10..20] of Integer;
var arr: TMyArray;
begin
	Low(arr);
end
			`,
			expected: 10,
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
				t.Errorf("Low() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLow_DynamicArrays tests Low() with dynamic arrays.
func TestBuiltinLow_DynamicArrays(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Low(arr);
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Dynamic arrays always have Low = 0
	if intVal.Value != 0 {
		t.Errorf("Low() for dynamic array = %d, want 0", intVal.Value)
	}
}

// ============================================================================
// High() Built-in Function Tests
// ============================================================================

// TestBuiltinHigh_StaticArrays tests High() with static arrays.
func TestBuiltinHigh_StaticArrays(t *testing.T) {
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
	High(arr);
end
			`,
			expected: 5,
		},
		{
			name: "Static array [0..9] returns 9",
			input: `
type TMyArray = array[0..9] of Integer;
var arr: TMyArray;
begin
	High(arr);
end
			`,
			expected: 9,
		},
		{
			name: "Static array [100..105] returns 105",
			input: `
type TMyArray = array[100..105] of Integer;
var arr: TMyArray;
begin
	High(arr);
end
			`,
			expected: 105,
		},
		{
			name: "Static array [10..20] returns 20",
			input: `
type TMyArray = array[10..20] of Integer;
var arr: TMyArray;
begin
	High(arr);
end
			`,
			expected: 20,
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
				t.Errorf("High() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinHigh_DynamicArrays tests High() with dynamic arrays.
func TestBuiltinHigh_DynamicArrays(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	High(arr);
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Empty dynamic array: High = Length - 1 = 0 - 1 = -1
	if intVal.Value != -1 {
		t.Errorf("High() for empty dynamic array = %d, want -1", intVal.Value)
	}
}

// TestBuiltinHigh_InForLoop tests High() used in a for loop.
func TestBuiltinHigh_InForLoop(t *testing.T) {
	input := `
type TMyArray = array[0..4] of Integer;
var arr: TMyArray;
var i: Integer;
var sum: Integer;
begin
	// Initialize array
	for i := 0 to 4 do
		arr[i] := i;

	// Sum using High()
	sum := 0;
	for i := Low(arr) to High(arr) do
		sum := sum + arr[i];

	sum;
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Expected: 0 + 1 + 2 + 3 + 4 = 10
	expected := int64(10)
	if intVal.Value != expected {
		t.Errorf("sum = %d, want %d", intVal.Value, expected)
	}
}

// ============================================================================
// SetLength() Built-in Function Tests
// ============================================================================

// TestBuiltinSetLength_Expand tests expanding a dynamic array.
func TestBuiltinSetLength_Expand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int64
	}{
		{
			name: "SetLength from 0 to 5",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	SetLength(arr, 5);
	Length(arr);
end
			`,
			expectedLen: 5,
		},
		{
			name: "SetLength from 0 to 10",
			input: `
type TDynArray = array of String;
var arr: TDynArray;
begin
	SetLength(arr, 10);
	Length(arr);
end
			`,
			expectedLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expectedLen {
				t.Errorf("Length after SetLength = %d, want %d", intVal.Value, tt.expectedLen)
			}
		})
	}
}

// TestBuiltinSetLength_Shrink tests shrinking a dynamic array.
func TestBuiltinSetLength_Shrink(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
var i: Integer;
begin
	// Expand to 10
	SetLength(arr, 10);

	// Fill with values
	for i := 0 to 9 do
		arr[i] := i * 10;

	// Shrink to 5
	SetLength(arr, 5);

	// Verify length
	Length(arr);
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	if intVal.Value != 5 {
		t.Errorf("Length after shrinking = %d, want 5", intVal.Value)
	}
}

// TestBuiltinSetLength_WithHighAndLow tests SetLength with High() and Low().
func TestBuiltinSetLength_WithHighAndLow(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	SetLength(arr, 5);
	Low(arr) + High(arr);
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Low = 0, High = 4, sum = 4
	expected := int64(4)
	if intVal.Value != expected {
		t.Errorf("Low + High = %d, want %d", intVal.Value, expected)
	}
}

// TestBuiltinSetLength_UseAfterResize tests using array elements after resize.
func TestBuiltinSetLength_UseAfterResize(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
var i: Integer;
var sum: Integer;
begin
	SetLength(arr, 5);

	// Fill array
	for i := 0 to High(arr) do
		arr[i] := i + 1;

	// Sum values
	sum := 0;
	for i := Low(arr) to High(arr) do
		sum := sum + arr[i];

	sum;
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Expected: 1 + 2 + 3 + 4 + 5 = 15
	expected := int64(15)
	if intVal.Value != expected {
		t.Errorf("sum = %d, want %d", intVal.Value, expected)
	}
}

// ============================================================================
// Error Cases for Low, High, SetLength
// ============================================================================

// TestBuiltinLowHighSetLength_ErrorCases tests error handling.
func TestBuiltinLowHighSetLength_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Low() with no arguments",
			input: `
begin
	Low();
end
			`,
		},
		{
			name: "High() with no arguments",
			input: `
begin
	High();
end
			`,
		},
		{
			name: "Low() with wrong type",
			input: `
var x: Integer := 42;
begin
	Low(x);
end
			`,
		},
		{
			name: "High() with wrong type",
			input: `
var s: String := "test";
begin
	High(s);
end
			`,
		},
		{
			name: "SetLength() with no arguments",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	SetLength();
end
			`,
		},
		{
			name: "SetLength() with one argument",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	SetLength(arr);
end
			`,
		},
		{
			name: "SetLength() with static array",
			input: `
type TStaticArray = array[1..5] of Integer;
var arr: TStaticArray;
begin
	SetLength(arr, 10);
end
			`,
		},
		{
			name: "SetLength() with negative length",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	SetLength(arr, -5);
end
			`,
		},
		{
			name: "SetLength() with non-integer length",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	SetLength(arr, "5");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			if _, ok := result.(*ErrorValue); !ok {
				t.Errorf("expected error, got %T: %+v", result, result)
			}
		})
	}
}

// ============================================================================
// Add() Built-in Function Tests
// ============================================================================

// TestBuiltinAdd_Basic tests basic Add() functionality.
func TestBuiltinAdd_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Add single element to empty array",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 42);
	Length(arr);
end
			`,
			expected: 1,
		},
		{
			name: "Add multiple elements",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Add(arr, 20);
	Add(arr, 30);
	Length(arr);
end
			`,
			expected: 3,
		},
		{
			name: "Add and retrieve element",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 100);
	arr[0];
end
			`,
			expected: 100,
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

// TestBuiltinAdd_InLoop tests Add() used in a loop.
func TestBuiltinAdd_InLoop(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
var i: Integer;
var sum: Integer;
begin
	// Add 5 elements
	for i := 1 to 5 do
		Add(arr, i * 10);

	// Sum all elements
	sum := 0;
	for i := 0 to High(arr) do
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

// TestBuiltinAdd_WithStrings tests Add() with string arrays.
func TestBuiltinAdd_WithStrings(t *testing.T) {
	input := `
type TStrArray = array of String;
var arr: TStrArray;
begin
	Add(arr, "hello");
	Add(arr, "world");
	Length(arr);
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	if intVal.Value != 2 {
		t.Errorf("Length = %d, want 2", intVal.Value)
	}
}

// ============================================================================
// Delete() Built-in Function Tests
// ============================================================================

// TestBuiltinDelete_Basic tests basic Delete() functionality.
func TestBuiltinDelete_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Delete first element",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Add(arr, 20);
	Add(arr, 30);
	Delete(arr, 0);
	Length(arr);
end
			`,
			expected: 2,
		},
		{
			name: "Delete middle element",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Add(arr, 20);
	Add(arr, 30);
	Delete(arr, 1);
	arr[1];
end
			`,
			expected: 30,
		},
		{
			name: "Delete last element",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Add(arr, 20);
	Add(arr, 30);
	Delete(arr, 2);
	Length(arr);
end
			`,
			expected: 2,
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

// TestBuiltinDelete_MultipleDeletes tests multiple Delete() operations.
func TestBuiltinDelete_MultipleDeletes(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
var i: Integer;
begin
	// Add 10 elements
	for i := 0 to 9 do
		Add(arr, i);

	// Delete every other element (from end to avoid index shifting issues)
	Delete(arr, 9);
	Delete(arr, 7);
	Delete(arr, 5);
	Delete(arr, 3);
	Delete(arr, 1);

	Length(arr);
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Started with 10, deleted 5
	expected := int64(5)
	if intVal.Value != expected {
		t.Errorf("Length after deletes = %d, want %d", intVal.Value, expected)
	}
}

// TestBuiltinDelete_VerifyContents tests that Delete() properly removes elements.
func TestBuiltinDelete_VerifyContents(t *testing.T) {
	input := `
type TDynArray = array of Integer;
var arr: TDynArray;
var sum: Integer;
var i: Integer;
begin
	Add(arr, 10);
	Add(arr, 20);
	Add(arr, 30);
	Add(arr, 40);
	Add(arr, 50);

	// Delete middle element (30)
	Delete(arr, 2);

	// Sum remaining elements
	sum := 0;
	for i := 0 to High(arr) do
		sum := sum + arr[i];

	sum;
end
	`

	result := testEval(input)

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
	}

	// Expected: 10 + 20 + 40 + 50 = 120 (30 was deleted)
	expected := int64(120)
	if intVal.Value != expected {
		t.Errorf("sum = %d, want %d", intVal.Value, expected)
	}
}

// ============================================================================
// Error Cases for Add and Delete
// ============================================================================

// TestBuiltinAddDelete_ErrorCases tests error handling for Add() and Delete().
func TestBuiltinAddDelete_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Add() with no arguments",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add();
end
			`,
		},
		{
			name: "Add() with one argument",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr);
end
			`,
		},
		{
			name: "Add() with static array",
			input: `
type TStaticArray = array[1..5] of Integer;
var arr: TStaticArray;
begin
	Add(arr, 42);
end
			`,
		},
		{
			name: "Add() with non-array",
			input: `
var x: Integer := 5;
begin
	Add(x, 10);
end
			`,
		},
		{
			name: "Delete() with no arguments",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Delete();
end
			`,
		},
		{
			name: "Delete() with one argument",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Delete(arr);
end
			`,
		},
		{
			name: "Delete() with static array",
			input: `
type TStaticArray = array[1..5] of Integer;
var arr: TStaticArray;
begin
	Delete(arr, 0);
end
			`,
		},
		{
			name: "Delete() with non-array",
			input: `
var x: Integer := 5;
begin
	Delete(x, 0);
end
			`,
		},
		{
			name: "Delete() with negative index",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Delete(arr, -1);
end
			`,
		},
		{
			name: "Delete() with out-of-bounds index",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Delete(arr, 5);
end
			`,
		},
		{
			name: "Delete() from empty array",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Delete(arr, 0);
end
			`,
		},
		{
			name: "Delete() with non-integer index",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
begin
	Add(arr, 10);
	Delete(arr, "0");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			if _, ok := result.(*ErrorValue); !ok {
				t.Errorf("expected error, got %T: %+v", result, result)
			}
		})
	}
}

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
// Task 9.71: Test IndexOf([1,2,3,2], 2) returns 1 (first occurrence at index 1)
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
// Task 9.71: Test IndexOf([1,2,3], 5) returns -1 (not found)
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
// Task 9.71: Test IndexOf([1,2,3,2], 2, 2) returns 3 (searches from index 2 onwards)
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
// Task 9.71: Test with strings
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
// Task 9.71: Test with empty array
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
// Task 9.71: Test edge cases with startIndex
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
// Task 9.71: Test error cases
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
// Task 9.73: Test Contains([1,2,3], 2) returns true
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
// Task 9.73: Test Contains([1,2,3], 5) returns false
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
// Task 9.73: Test with different types
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
// Task 9.73: Test with empty array returns false
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
// Task 9.75: Test var a := [1,2,3]; Reverse(a); → a = [3,2,1]
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
// Task 9.75: Test with even length array
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
// Task 9.75: Test with single element (no-op)
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
// Task 9.75: Test with empty array (no-op)
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
// Task 9.78: Test var a := [3,1,2]; Sort(a); → a = [1,2,3]
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
// Task 9.78: Test with strings: ['c','a','b'] → ['a','b','c']
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
// Task 9.78: Test with already sorted array
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
// Task 9.78: Test with single element
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
// Task 9.78: Test with duplicates
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
// Task 9.33: Sort(arr, lambda (a, b) => a - b)
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
// Task 9.33: Sort(arr, lambda (a, b) => b - a)
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
// Task 9.33: Sort strings by length instead of lexicographically
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
// Task 9.33: Sort(arr, @MyComparatorFunc)
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
// Task 9.33: Error cases
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
// Task 9.33: Edge case - empty array
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
// Task 9.33: Edge case - single element
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

// ============================================================================
// Inline Array Type Tests
// ============================================================================

func TestInlineArrayTypes_DynamicArrayVariable(t *testing.T) {
	input := `
var arr: array of Integer;
begin
	SetLength(arr, 3);
	arr[0] := 10;
	arr[1] := 20;
	arr[2] := 30;
	arr[0] + arr[1] + arr[2];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 60 {
		t.Errorf("expected 60, got %d", intVal.Value)
	}
}

func TestInlineArrayTypes_StaticArrayVariable(t *testing.T) {
	input := `
var arr: array[1..10] of Integer;
begin
	arr[1] := 100;
	arr[10] := 200;
	arr[1] + arr[10];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 300 {
		t.Errorf("expected 300, got %d", intVal.Value)
	}
}

func TestInlineArrayTypes_ZeroBasedStaticArray(t *testing.T) {
	input := `
var arr: array[0..9] of String;
begin
	arr[0] := 'first';
	arr[9] := 'last';
	arr[0] + ' and ' + arr[9];
end
	`

	result := testEval(input)
	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T: %+v", result, result)
	}

	if strVal.Value != "first and last" {
		t.Errorf("expected 'first and last', got %q", strVal.Value)
	}
}

func TestInlineArrayTypes_NegativeBounds(t *testing.T) {
	input := `
var arr: array[-5..5] of Integer;
begin
	arr[-5] := -50;
	arr[0] := 0;
	arr[5] := 50;
	arr[-5] + arr[0] + arr[5];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 0 {
		t.Errorf("expected 0, got %d", intVal.Value)
	}
}

func TestInlineArrayTypes_NestedStaticArrays(t *testing.T) {
	input := `
var matrix: array[1..3] of array[1..3] of Integer;
begin
	matrix[1][1] := 11;
	matrix[3][3] := 33;
	matrix[1][1] + matrix[3][3];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 44 {
		t.Errorf("expected 44, got %d", intVal.Value)
	}
}

func TestInlineArrayTypes_MixedStaticDynamic(t *testing.T) {
	input := `
var arr: array[1..2] of array of String;
begin
	SetLength(arr[1], 2);
	SetLength(arr[2], 2);
	arr[1][0] := 'hello';
	arr[2][1] := 'world';
	arr[1][0] + ' ' + arr[2][1];
end
	`

	result := testEval(input)
	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T: %+v", result, result)
	}

	if strVal.Value != "hello world" {
		t.Errorf("expected 'hello world', got %q", strVal.Value)
	}
}

func TestInlineArrayTypes_InFunctionParameter(t *testing.T) {
	input := `
function Sum(arr: array of Integer): Integer;
var i: Integer;
var total: Integer;
begin
	total := 0;
	for i := 0 to Length(arr) - 1 do
		total := total + arr[i];
	Result := total;
end;

var nums: array of Integer;
begin
	SetLength(nums, 3);
	nums[0] := 10;
	nums[1] := 20;
	nums[2] := 30;
	Sum(nums);
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 60 {
		t.Errorf("expected 60, got %d", intVal.Value)
	}
}

func TestInlineArrayTypes_StaticArrayInFunctionParameter(t *testing.T) {
	input := `
procedure Fill(var arr: array[1..5] of Integer; value: Integer);
var i: Integer;
begin
	for i := 1 to 5 do
		arr[i] := value;
end;

var numbers: array[1..5] of Integer;
begin
	Fill(numbers, 42);
	numbers[1] + numbers[5];
end
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 84 {
		t.Errorf("expected 84, got %d", intVal.Value)
	}
}

// ============================================================================
// Array Instantiation with new Keyword Tests
// ============================================================================

// TestNewArrayExpression_Integer1D tests creating a 1D array of integers.
// Example: new Integer[10] creates an array with 10 zero elements.
func TestNewArrayExpression_Integer1D(t *testing.T) {
	input := `
		var arr := new Integer[10];
		arr[0]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 0 {
		t.Errorf("expected 0 (zero value), got %d", intVal.Value)
	}
}

// TestNewArrayExpression_String1D tests creating a 1D array of strings.
// Example: new String[5] creates an array with 5 empty strings.
func TestNewArrayExpression_String1D(t *testing.T) {
	input := `
		var arr := new String[5];
		arr[2]
	`

	result := testEval(input)
	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T: %+v", result, result)
	}

	if strVal.Value != "" {
		t.Errorf("expected empty string, got '%s'", strVal.Value)
	}
}

// TestNewArrayExpression_Float1D tests creating a 1D array of floats.
func TestNewArrayExpression_Float1D(t *testing.T) {
	input := `
		var arr := new Float[3];
		arr[1]
	`

	result := testEval(input)
	floatVal, ok := result.(*FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T: %+v", result, result)
	}

	if floatVal.Value != 0.0 {
		t.Errorf("expected 0.0 (zero value), got %f", floatVal.Value)
	}
}

// TestNewArrayExpression_Boolean1D tests creating a 1D array of booleans.
func TestNewArrayExpression_Boolean1D(t *testing.T) {
	input := `
		var arr := new Boolean[4];
		arr[0]
	`

	result := testEval(input)
	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("expected BooleanValue, got %T: %+v", result, result)
	}

	if boolVal.Value != false {
		t.Errorf("expected false (zero value), got %t", boolVal.Value)
	}
}

// TestNewArrayExpression_2D tests creating a 2D array.
// Example: new Integer[3, 4] creates a 3×4 matrix.
func TestNewArrayExpression_2D(t *testing.T) {
	input := `
		var matrix := new Integer[3, 4];
		matrix[1][2]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 0 {
		t.Errorf("expected 0 (zero value), got %d", intVal.Value)
	}
}

// TestNewArrayExpression_3D tests creating a 3D array.
func TestNewArrayExpression_3D(t *testing.T) {
	input := `
		var cube := new Integer[2, 3, 4];
		cube[0][1][2]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 0 {
		t.Errorf("expected 0 (zero value), got %d", intVal.Value)
	}
}

// TestNewArrayExpression_WithExpression tests dimension expressions are evaluated.
// Example: new Integer[2*5] should create an array with 10 elements.
func TestNewArrayExpression_WithExpression(t *testing.T) {
	input := `
		var size := 5;
		var arr := new Integer[2 * size];
		arr[9]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 0 {
		t.Errorf("expected 0 (zero value), got %d", intVal.Value)
	}
}

// TestNewArrayExpression_ElementAssignment tests assigning to elements.
func TestNewArrayExpression_ElementAssignment(t *testing.T) {
	input := `
		var arr := new Integer[5];
		arr[2] := 42;
		arr[2]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 42 {
		t.Errorf("expected 42, got %d", intVal.Value)
	}
}

// TestNewArrayExpression_2DElementAssignment tests assigning to 2D array elements.
func TestNewArrayExpression_2DElementAssignment(t *testing.T) {
	input := `
		var matrix := new Integer[3, 4];
		matrix[1][2] := 99;
		matrix[1][2]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 99 {
		t.Errorf("expected 99, got %d", intVal.Value)
	}
}

// TestNewArrayExpression_StringAssignment tests string array assignments.
func TestNewArrayExpression_StringAssignment(t *testing.T) {
	input := `
		var arr := new String[3];
		arr[0] := 'hello';
		arr[1] := 'world';
		arr[0]
	`

	result := testEval(input)
	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T: %+v", result, result)
	}

	if strVal.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", strVal.Value)
	}
}

// TestNewArrayExpression_NegativeDimension tests error handling for negative dimensions.
func TestNewArrayExpression_NegativeDimension(t *testing.T) {
	input := `
		var arr := new Integer[-5];
	`

	result := testEval(input)
	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
	}

	if !strings.Contains(errVal.Message, "must be positive") {
		t.Errorf("expected error about positive dimension, got: %s", errVal.Message)
	}
}

// TestNewArrayExpression_ZeroDimension tests error handling for zero dimensions.
func TestNewArrayExpression_ZeroDimension(t *testing.T) {
	input := `
		var arr := new Integer[0];
	`

	result := testEval(input)
	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
	}

	if !strings.Contains(errVal.Message, "must be positive") {
		t.Errorf("expected error about positive dimension, got: %s", errVal.Message)
	}
}

// TestNewArrayExpression_NonIntegerDimension tests error handling for non-integer dimensions.
func TestNewArrayExpression_NonIntegerDimension(t *testing.T) {
	input := `
		var arr := new Integer[3.5];
	`

	result := testEval(input)
	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
	}

	if !strings.Contains(errVal.Message, "must be an integer") {
		t.Errorf("expected error about integer dimension, got: %s", errVal.Message)
	}
}

// TestNewArrayExpression_UnknownType tests error handling for unknown element types.
func TestNewArrayExpression_UnknownType(t *testing.T) {
	input := `
		var arr := new UnknownType[10];
	`

	result := testEval(input)
	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("expected ErrorValue, got %T: %+v", result, result)
	}

	if !strings.Contains(errVal.Message, "unknown element type") {
		t.Errorf("expected error about unknown type, got: %s", errVal.Message)
	}
}

// TestNewArrayExpression_LargeArray tests creating a larger array.
func TestNewArrayExpression_LargeArray(t *testing.T) {
	input := `
		var arr := new Integer[1000];
		arr[999] := 123;
		arr[999]
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	if intVal.Value != 123 {
		t.Errorf("expected 123, got %d", intVal.Value)
	}
}

// TestNewArrayExpression_NestedIteration tests iterating over a 2D array.
func TestNewArrayExpression_NestedIteration(t *testing.T) {
	input := `
		var matrix := new Integer[2, 3];
		matrix[0][0] := 1;
		matrix[0][1] := 2;
		matrix[0][2] := 3;
		matrix[1][0] := 4;
		matrix[1][1] := 5;
		matrix[1][2] := 6;

		var sum := 0;
		for var i := 0 to 1 do
		begin
			for var j := 0 to 2 do
			begin
				sum := sum + matrix[i][j];
			end;
		end;
		sum
	`

	result := testEval(input)
	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("expected IntegerValue, got %T: %+v", result, result)
	}

	// Sum should be 1+2+3+4+5+6 = 21
	if intVal.Value != 21 {
		t.Errorf("expected 21, got %d", intVal.Value)
	}
}
