package semantic

import (
	"testing"
)

// ============================================================================
// Array Type Registration Tests
// ============================================================================

func TestArrayTypeRegistration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "static array type",
			input: `
				type TMyArray = array[1..10] of Integer;
			`,
		},
		{
			name: "dynamic array type",
			input: `
				type TStringArray = array of String;
			`,
		},
		{
			name: "array type with variable",
			input: `
				type TIntArray = array[0..99] of Integer;
				var numbers: TIntArray;
			`,
		},
		{
			name: "multiple array types",
			input: `
				type TIntArray = array[1..10] of Integer;
				type TStringArray = array of String;
				type TFloatArray = array[0..4] of Float;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestArrayTypeErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "duplicate array type declaration",
			input: `
				type TMyArray = array[1..10] of Integer;
				type TMyArray = array of String;
			`,
			expectedError: "type 'TMyArray' already declared",
		},
		{
			name: "undefined element type",
			input: `
				type TMyArray = array[1..10] of TUnknown;
			`,
			expectedError: "unknown type 'TUnknown'",
		},
		{
			name: "invalid array bounds (low > high)",
			input: `
				type TBadArray = array[10..1] of Integer;
			`,
			expectedError: "invalid array bounds: lower bound (10) > upper bound (1)",
		},
		{
			name: "undefined array type in variable",
			input: `
				var arr: TMyArray;
			`,
			expectedError: "unknown type 'TMyArray'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Array Indexing Tests
// ============================================================================

func TestArrayIndexing(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple array access",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var x: Integer;
				x := arr[5];
			`,
		},
		{
			name: "array access with variable index",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var i: Integer;
				var x: Integer;
				x := arr[i];
			`,
		},
		{
			name: "array access with expression index",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var i: Integer;
				var x: Integer;
				x := arr[i + 1];
			`,
		},
		{
			name: "nested array indexing",
			input: `
				type TRow = array[1..10] of Integer;
				type TMatrix = array[1..5] of TRow;
				var matrix: TMatrix;
				var x: Integer;
				x := matrix[1][2];
			`,
		},
		{
			name: "dynamic array access",
			input: `
				type TDynArray = array of Integer;
				var arr: TDynArray;
				var x: Integer;
				x := arr[0];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestArrayIndexingErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "index non-array type",
			input: `
				var x: Integer;
				var y: Integer;
				y := x[0];
			`,
			expectedError: "cannot index non-array type",
		},
		{
			name: "non-integer index",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var x: Integer;
				x := arr['hello'];
			`,
			expectedError: "array index must be integer",
		},
		{
			name: "float index",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var f: Float;
				var x: Integer;
				x := arr[f];
			`,
			expectedError: "array index must be integer",
		},
		{
			name: "undefined array variable",
			input: `
				var x: Integer;
				x := unknownArray[0];
			`,
			expectedError: "undefined variable 'unknownArray'",
		},
		{
			name: "type mismatch on assignment",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var s: String;
				s := arr[0];
			`,
			expectedError: "cannot assign Integer to String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Array Assignment Tests
// ============================================================================

// Note: Assignment to array elements (arr[i] := x) requires parser support
// for index expressions as lvalues. This is tracked separately.
// For now, we test reading from arrays, which exercises the semantic analysis.

func TestArrayElementAccess(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "read from array element",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var x: Integer;
				x := arr[3];
			`,
		},
		{
			name: "use array element in expression",
			input: `
				type TIntArray = array[1..10] of Integer;
				var arr: TIntArray;
				var x: Integer;
				x := arr[0] + arr[1];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// ============================================================================
// Inline Array Type Tests
// ============================================================================

func TestInlineArrayTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "inline dynamic array in variable",
			input: `
				var arr: array of Integer;
			`,
		},
		{
			name: "inline static array in variable",
			input: `
				var arr: array[1..10] of Integer;
			`,
		},
		{
			name: "inline static array zero-based",
			input: `
				var arr: array[0..99] of String;
			`,
		},
		{
			name: "inline static array negative bounds",
			input: `
				var arr: array[-10..10] of Integer;
			`,
		},
		{
			name: "inline nested dynamic arrays",
			input: `
				var matrix: array of array of Integer;
			`,
		},
		{
			name: "inline nested static arrays",
			input: `
				var matrix: array[1..5] of array[1..10] of Integer;
			`,
		},
		{
			name: "inline mixed static and dynamic",
			input: `
				var mixedA: array[1..10] of array of Integer;
				var mixedB: array of array[1..5] of String;
			`,
		},
		{
			name: "inline array in function parameter",
			input: `
				procedure Test(arr: array of Integer);
				begin
				end;
			`,
		},
		{
			name: "inline static array in function parameter",
			input: `
				procedure Test(arr: array[1..10] of Integer);
				begin
				end;
			`,
		},
		{
			name: "inline array in function return type (future)",
			input: `
				type TResult = array of Integer;
				function GetData(): TResult;
				begin
				end;
			`,
		},
		{
			name: "multiple inline array variables",
			input: `
				var numbers: array[1..10] of Integer;
				var names: array[0..4] of String;
				var statuses: array[1..100] of Boolean;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestInlineArrayTypeCompatibility(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "inline array element type matching",
			input: `
				var arr: array[1..10] of Integer;
				var x: Integer;
				x := arr[1];
			`,
		},
		{
			name: "nested inline array element access",
			input: `
				var matrix: array[1..5] of array[1..10] of Integer;
				var x: Integer;
				x := matrix[1][5];
			`,
		},
		{
			name: "inline dynamic array element access",
			input: `
				var arr: array of String;
				var s: String;
				s := arr[0];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestInlineArrayTypeErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "inline array with undefined element type",
			input: `
				var arr: array of TUnknown;
			`,
			expectedError: "unknown type 'array of TUnknown'",
		},
		{
			name: "inline static array with undefined element type",
			input: `
				var arr: array[1..10] of TUnknown;
			`,
			expectedError: "unknown type 'array[1..10] of TUnknown'",
		},
		{
			name: "type mismatch with inline array element",
			input: `
				var arr: array[1..10] of Integer;
				var s: String;
				s := arr[0];
			`,
			expectedError: "cannot assign Integer to String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Array Instantiation with 'new' Keyword Tests
// ============================================================================

func TestNewArrayExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple 1D array instantiation",
			input: `
				var a: array of Integer;
				a := new Integer[16];
			`,
		},
		{
			name: "2D array instantiation",
			input: `
				var matrix: array of array of Integer;
				matrix := new Integer[10, 20];
			`,
		},
		{
			name: "3D array instantiation",
			input: `
				var cube: array of array of array of Float;
				cube := new Float[5, 10, 15];
			`,
		},
		{
			name: "array instantiation with variable size",
			input: `
				var size: Integer;
				var arr: array of String;
				size := 10;
				arr := new String[size];
			`,
		},
		{
			name: "array instantiation with expression size",
			input: `
				var n: Integer;
				var arr: array of Integer;
				n := 5;
				arr := new Integer[n * 2];
			`,
		},
		{
			name: "array instantiation in var declaration",
			input: `
				var a := new Integer[16];
			`,
		},
		{
			name: "different element types",
			input: `
				var ints := new Integer[10];
				var strs := new String[5];
				var floats := new Float[8];
				var bools := new Boolean[3];
			`,
		},
		{
			name: "array instantiation with function call size",
			input: `
				function GetSize(): Integer;
				begin
					Result := 10;
				end;

				var arr := new Integer[GetSize()];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestNewArrayExpressionErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "non-integer dimension (float)",
			input: `
				var arr := new Integer[3.14];
			`,
			expectedError: "array dimension 1 must be integer, got Float",
		},
		{
			name: "non-integer dimension (string)",
			input: `
				var arr := new Integer['hello'];
			`,
			expectedError: "array dimension 1 must be integer, got String",
		},
		{
			name: "non-integer dimension (boolean)",
			input: `
				var b: Boolean;
				var arr := new Integer[b];
			`,
			expectedError: "array dimension 1 must be integer, got Boolean",
		},
		{
			name: "non-integer dimension in 2D array (first dimension)",
			input: `
				var arr := new Integer[3.14, 20];
			`,
			expectedError: "array dimension 1 must be integer, got Float",
		},
		{
			name: "non-integer dimension in 2D array (second dimension)",
			input: `
				var arr := new Integer[10, 'hello'];
			`,
			expectedError: "array dimension 2 must be integer, got String",
		},
		{
			name: "undefined element type",
			input: `
				var arr := new TUnknownType[10];
			`,
			expectedError: "unknown type 'TUnknownType'",
		},
		{
			name: "undefined variable in dimension",
			input: `
				var arr := new Integer[unknownVar];
			`,
			expectedError: "undefined variable 'unknownVar'",
		},
		{
			name: "type mismatch - assigning to wrong type",
			input: `
				var s: String;
				s := new Integer[10];
			`,
			expectedError: "cannot assign array of Integer to String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

func TestNewArrayExpressionTypeInference(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "1D array type matches variable",
			input: `
				var arr: array of Integer;
				arr := new Integer[10];
			`,
		},
		{
			name: "2D array type matches variable",
			input: `
				var matrix: array of array of String;
				matrix := new String[5, 10];
			`,
		},
		{
			name: "3D array type matches variable",
			input: `
				var cube: array of array of array of Boolean;
				cube := new Boolean[2, 3, 4];
			`,
		},
		{
			name: "array element access after instantiation",
			input: `
				var arr := new Integer[10];
				var x: Integer;
				x := arr[0];
			`,
		},
		{
			name: "nested array element access after instantiation",
			input: `
				var matrix := new Integer[5, 10];
				var x: Integer;
				x := matrix[0][5];
			`,
		},
		{
			name: "array used in expression after instantiation",
			input: `
				var arr := new Integer[10];
				var sum: Integer;
				sum := arr[0] + arr[1];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestNewArrayVsNewClassDistinction(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "new with parentheses is class instantiation",
			input: `
				type TPoint = class
				public
					X, Y: Integer;
					constructor Create(aX, aY: Integer);
				end;

				constructor TPoint.Create(aX, aY: Integer);
				begin
					X := aX;
					Y := aY;
				end;

				var p := new TPoint(10, 20);
			`,
		},
		{
			name: "new with brackets is array instantiation",
			input: `
				var arr := new Integer[10];
			`,
		},
		{
			name: "both can coexist",
			input: `
				type TIntArray = array of Integer;

				type TData = class
				public
					Values: TIntArray;
					constructor Create();
				end;

				constructor TData.Create();
				begin
					Values := new Integer[100];
				end;

				var obj := new TData();
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}
