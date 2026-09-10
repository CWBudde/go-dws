package semantic

import (
	"testing"
)

// ============================================================================
// Indexing the result of an implicit (parenless) function call
// ============================================================================

// TestIndexImplicitFunctionCall covers `Test['toto']` where `Test` is a
// parameterless function whose result is an array, associative array or
// string (PLAN.md §3.3).
func TestIndexImplicitFunctionCall(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "parenless function returning dynamic array indexed by integer",
			input: `
				function Test : array of Integer;
				begin
					Result := [1, 2, 3];
				end;
				var x : Integer := Test[0];
			`,
		},
		{
			name: "parenless function returning static array indexed by integer",
			input: `
				type TIntArray = array[1..3] of Integer;
				function Test : TIntArray;
				begin
				end;
				var x : Integer := Test[1];
			`,
		},
		{
			name: "parenless function returning associative array indexed by string",
			input: `
				function Test : array[String] of Integer;
				begin
					Result['a'] := 1;
				end;
				var x : Integer := Test['a'];
			`,
		},
		{
			name: "parenless function returning associative array of records",
			input: `
				type TMyRec = record
					Value1 : String;
				end;
				function Test : array[String] of TMyRec;
				begin
				end;
				var r := Test['toto'];
				var s : String := r.Value1;
			`,
		},
		{
			name: "overload set whose parameterless overload returns an array",
			input: `
				type TIntArray = array of Integer;
				function Test : TIntArray; overload;
				begin
					Result := [1, 2, 3];
				end;
				function Test(i : Integer) : Integer; overload;
				begin
					Result := i;
				end;
				var x : Integer := Test[0];
			`,
		},
		{
			name: "parenless function returning string indexed by integer",
			input: `
				function Test : String;
				begin
					Result := 'abc';
				end;
				var c : String := Test[1];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// TestIndexImplicitFunctionCallErrors makes sure the implicit-call unwrap does
// not turn genuinely non-indexable expressions into accepted programs.
func TestIndexImplicitFunctionCallErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "parenless function returning non-indexable type",
			input: `
				function Test : Integer;
				begin
					Result := 1;
				end;
				var x : Integer := Test[0];
			`,
			expectedError: "Array expected",
		},
		{
			name: "function pointer with parameters is not indexable",
			input: `
				type TFn = function (i : Integer) : Integer;
				var f : TFn;
				var x : Integer := f[0];
			`,
			expectedError: "Array expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}
