package interp

import (
	"testing"
)

// ============================================================================

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
// SetLength() on String Tests
// ============================================================================

// TestBuiltinSetLength_String_Expand tests expanding a string with spaces.
func TestBuiltinSetLength_String_Expand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
	}{
		{
			name: "SetLength expands empty string with spaces",
			input: `
var s: String := '';
begin
	SetLength(s, 5);
	s;
end
			`,
			expectedResult: "     ", // 5 spaces
		},
		{
			name: "SetLength expands short string with spaces",
			input: `
var s: String := 'Hi';
begin
	SetLength(s, 10);
	s;
end
			`,
			expectedResult: "Hi        ", // "Hi" + 8 spaces
		},
		{
			name: "SetLength expands to exact length",
			input: `
var s: String := 'Test';
begin
	SetLength(s, 8);
	s;
end
			`,
			expectedResult: "Test    ", // "Test" + 4 spaces
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expectedResult {
				t.Errorf("String after SetLength = %q (len=%d), want %q (len=%d)",
					strVal.Value, len(strVal.Value), tt.expectedResult, len(tt.expectedResult))
			}
		})
	}
}

// TestBuiltinSetLength_String_Truncate tests truncating a string.
func TestBuiltinSetLength_String_Truncate(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
	}{
		{
			name: "SetLength truncates to zero",
			input: `
var s: String := 'Hello World';
begin
	SetLength(s, 0);
	s;
end
			`,
			expectedResult: "",
		},
		{
			name: "SetLength truncates to shorter length",
			input: `
var s: String := 'Hello World';
begin
	SetLength(s, 5);
	s;
end
			`,
			expectedResult: "Hello",
		},
		{
			name: "SetLength truncates multi-byte UTF-8",
			input: `
var s: String := 'Hello 世界';
begin
	SetLength(s, 6);
	s;
end
			`,
			expectedResult: "Hello ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expectedResult {
				t.Errorf("String after SetLength = %q, want %q", strVal.Value, tt.expectedResult)
			}
		})
	}
}

// TestBuiltinSetLength_String_SameLength tests SetLength with same length.
func TestBuiltinSetLength_String_SameLength(t *testing.T) {
	input := `
var s: String := 'Hello';
begin
	SetLength(s, 5);
	s;
end
	`

	result := testEval(input)

	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
	}

	if strVal.Value != "Hello" {
		t.Errorf("String after SetLength = %q, want 'Hello'", strVal.Value)
	}
}

// TestBuiltinSetLength_String_UTF8 tests SetLength with UTF-8 strings.
func TestBuiltinSetLength_String_UTF8(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
		expectedLen    int
	}{
		{
			name: "UTF-8 emoji expansion",
			input: `
var s: String := '😀😁';
begin
	SetLength(s, 5);
	s;
end
			`,
			expectedResult: "😀😁   ", // 2 emojis + 3 spaces
			expectedLen:    5,       // 5 runes
		},
		{
			name: "UTF-8 Chinese characters",
			input: `
var s: String := '你好';
begin
	SetLength(s, 6);
	s;
end
			`,
			expectedResult: "你好    ", // 2 chars + 4 spaces
			expectedLen:    6,
		},
		{
			name: "UTF-8 truncation",
			input: `
var s: String := '世界你好';
begin
	SetLength(s, 2);
	s;
end
			`,
			expectedResult: "世界",
			expectedLen:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expectedResult {
				t.Errorf("String after SetLength = %q, want %q", strVal.Value, tt.expectedResult)
			}

			// Verify the rune length matches expected
			runeLen := len([]rune(strVal.Value))
			if runeLen != tt.expectedLen {
				t.Errorf("Rune length = %d, want %d", runeLen, tt.expectedLen)
			}
		})
	}
}

// TestBuiltinSetLength_String_VarParam tests SetLength with var parameters.
func TestBuiltinSetLength_String_VarParam(t *testing.T) {
	input := `
procedure ModifyString(var s: String);
begin
	SetLength(s, 10);
end;

var myStr: String := 'Test';
begin
	ModifyString(myStr);
	myStr;
end
	`

	result := testEval(input)

	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
	}

	expectedResult := "Test      " // "Test" + 6 spaces
	if strVal.Value != expectedResult {
		t.Errorf("String after SetLength = %q (len=%d), want %q (len=%d)",
			strVal.Value, len(strVal.Value), expectedResult, len(expectedResult))
	}
}
