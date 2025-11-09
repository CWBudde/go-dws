package semantic

import (
	"testing"
)

// ============================================================================
// Built-in Array Functions Tests
// ============================================================================
// These tests cover the built-in array manipulation functions to improve
// coverage of analyze_builtin_array.go (currently at 0-50% coverage)

// Low function tests
func TestBuiltinLow_Array(t *testing.T) {
	input := `
		var arr: array[1..10] of Integer;
		var lowIndex := Low(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLow_DynamicArray(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3];
		var lowIndex := Low(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLow_String(t *testing.T) {
	input := `
		var s := 'hello';
		var lowIndex := Low(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLow_InvalidType(t *testing.T) {
	input := `
		var x := 42;
		var lowIndex := Low(x);
	`
	expectError(t, input, "array")
}

// High function tests
func TestBuiltinHigh_Array(t *testing.T) {
	input := `
		var arr: array[1..10] of Integer;
		var highIndex := High(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinHigh_DynamicArray(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3];
		var highIndex := High(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinHigh_String(t *testing.T) {
	input := `
		var s := 'hello';
		var highIndex := High(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinHigh_InvalidType(t *testing.T) {
	input := `
		var x := 42;
		var highIndex := High(x);
	`
	expectError(t, input, "array")
}

// SetLength function tests
func TestBuiltinSetLength_DynamicArray(t *testing.T) {
	input := `
		var arr: array of Integer;
		SetLength(arr, 10);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSetLength_Resize(t *testing.T) {
	input := `
		var arr: array of String := ['a', 'b', 'c'];
		SetLength(arr, 5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSetLength_Zero(t *testing.T) {
	input := `
		var arr: array of Integer;
		SetLength(arr, 0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSetLength_StaticArray(t *testing.T) {
	input := `
		var arr: array[1..10] of Integer;
		SetLength(arr, 5);
	`
	expectError(t, input, "dynamic array")
}

func TestBuiltinSetLength_InvalidArgCount(t *testing.T) {
	input := `
		var arr: array of Integer;
		SetLength(arr);
	`
	expectError(t, input, "argument")
}

func TestBuiltinSetLength_InvalidLengthType(t *testing.T) {
	input := `
		var arr: array of Integer;
		SetLength(arr, 'ten');
	`
	expectError(t, input, "integer")
}

// Add function tests (for dynamic arrays)
func TestBuiltinAdd_Integer(t *testing.T) {
	input := `
		var arr: array of Integer;
		Add(arr, 42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinAdd_String(t *testing.T) {
	input := `
		var arr: array of String;
		Add(arr, 'hello');
	`
	expectNoErrors(t, input)
}

func TestBuiltinAdd_Object(t *testing.T) {
	input := `
		type TMyClass = class
		end;
		var arr: array of TMyClass;
		var obj: TMyClass;
		Add(arr, obj);
	`
	expectNoErrors(t, input)
}

func TestBuiltinAdd_InvalidArgCount(t *testing.T) {
	input := `
		var arr: array of Integer;
		Add(arr);
	`
	expectError(t, input, "argument")
}

func TestBuiltinAdd_NotArray(t *testing.T) {
	input := `
		var x := 5;
		Add(x, 10);
	`
	expectError(t, input, "array")
}

func TestBuiltinAdd_TypeMismatch(t *testing.T) {
	input := `
		var arr: array of Integer;
		Add(arr, 'hello');
	`
	expectError(t, input, "type")
}

// Delete function tests
func TestBuiltinDelete_SingleElement(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3, 4, 5];
		Delete(arr, 2, 1);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDelete_MultipleElements(t *testing.T) {
	input := `
		var arr: array of String := ['a', 'b', 'c', 'd'];
		Delete(arr, 1, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDelete_InvalidArgCount(t *testing.T) {
	input := `
		var arr: array of Integer;
		Delete(arr, 0);
	`
	expectError(t, input, "argument")
}

func TestBuiltinDelete_NotArray(t *testing.T) {
	input := `
		var s := 'hello';
		Delete(s, 1, 2);
	`
	// Delete can work on strings in some Pascal variants
	// The error handling depends on implementation
	expectNoErrors(t, input)
}

func TestBuiltinDelete_InvalidIndexType(t *testing.T) {
	input := `
		var arr: array of Integer;
		Delete(arr, 'first', 1);
	`
	expectError(t, input, "integer")
}

// Combined array operations tests
func TestBuiltinArray_LowHighLength(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3, 4, 5];
		var low := Low(arr);
		var high := High(arr);
		var len := Length(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_DynamicResize(t *testing.T) {
	input := `
		var arr: array of Integer;
		SetLength(arr, 5);
		Add(arr, 42);
		var len := Length(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_IterationWithBounds(t *testing.T) {
	input := `
		var arr: array of Integer := [10, 20, 30, 40, 50];
		for i := Low(arr) to High(arr) do
		begin
			PrintLn(arr[i]);
		end;
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_AddMultiple(t *testing.T) {
	input := `
		var arr: array of String;
		Add(arr, 'first');
		Add(arr, 'second');
		Add(arr, 'third');
		var count := Length(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_DeleteFromMiddle(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3, 4, 5];
		Delete(arr, 2, 2);
		var newLen := Length(arr);
	`
	expectNoErrors(t, input)
}

// Multidimensional arrays
func TestBuiltinArray_MultidimensionalLength(t *testing.T) {
	input := `
		var matrix: array of array of Integer;
		SetLength(matrix, 3);
		SetLength(matrix[0], 4);
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinArray_EmptyArray(t *testing.T) {
	input := `
		var arr: array of Integer;
		var low := Low(arr);
		var high := High(arr);
		var len := Length(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_LargeResize(t *testing.T) {
	input := `
		var arr: array of Integer;
		SetLength(arr, 10000);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_ResizeToZero(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3];
		SetLength(arr, 0);
		var len := Length(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_InFunction(t *testing.T) {
	input := `
		function CreateArray(size: Integer): array of Integer;
		begin
			SetLength(Result, size);
			for i := Low(Result) to High(Result) do
				Result[i] := i * 2;
		end;

		var arr := CreateArray(10);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArray_AsParameter(t *testing.T) {
	input := `
		procedure PrintArray(arr: array of Integer);
		begin
			for i := Low(arr) to High(arr) do
				PrintLn(arr[i]);
		end;

		var myArray: array of Integer := [1, 2, 3];
		PrintArray(myArray);
	`
	expectNoErrors(t, input)
}

// Array with records
func TestBuiltinArray_Records(t *testing.T) {
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var points: array of TPoint;
		SetLength(points, 3);
	`
	expectNoErrors(t, input)
}

// Array with classes
func TestBuiltinArray_Classes(t *testing.T) {
	input := `
		type TBase = class
		end;
		var objects: array of TBase;
		SetLength(objects, 5);
	`
	expectNoErrors(t, input)
}
