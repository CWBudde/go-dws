package interp

import (
	"strings"
	"testing"
)

// ============================================================================
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
	// Should work and return nil or zero value
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
