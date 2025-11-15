package interp

import (
	"testing"
)

// TestMultiDimensionalArrayCreation tests creating multi-dimensional arrays with new keyword
func TestMultiDimensionalArrayCreation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "2D array - basic creation",
			input: `
				function Test: Integer;
				var arr: array of array of Integer;
				begin
					arr := new Integer[3, 4];
					Result := Length(arr);
				end;
				Test();
			`,
			expected: 3,
		},
		{
			name: "2D array - inner length",
			input: `
				function Test: Integer;
				var arr: array of array of Integer;
				begin
					arr := new Integer[3, 4];
					Result := Length(arr[0]);
				end;
				Test();
			`,
			expected: 4,
		},
		{
			name: "2D array - set and get value",
			input: `
				function Test: Integer;
				var arr: array of array of Integer;
				begin
					arr := new Integer[3, 4];
					arr[1][2] := 42;
					Result := arr[1][2];
				end;
				Test();
			`,
			expected: 42,
		},
		{
			name: "2D array - with expressions for dimensions",
			input: `
				function Test: Integer;
				var M, N: Integer;
				var arr: array of array of Integer;
				begin
					M := 5;
					N := 10;
					arr := new Integer[M, N];
					Result := Length(arr) * Length(arr[0]);
				end;
				Test();
			`,
			expected: 50,
		},
		{
			name: "3D array - basic creation",
			input: `
				function Test: Integer;
				var arr: array of array of array of Integer;
				begin
					arr := new Integer[2, 3, 4];
					Result := Length(arr);
				end;
				Test();
			`,
			expected: 2,
		},
		{
			name: "3D array - nested access",
			input: `
				function Test: Integer;
				var arr: array of array of array of Integer;
				begin
					arr := new Integer[2, 3, 4];
					arr[1][2][3] := 99;
					Result := arr[1][2][3];
				end;
				Test();
			`,
			expected: 99,
		},
		{
			name: "2D array - Float type",
			input: `
				function Test: Integer;
				var arr: array of array of Float;
				begin
					arr := new Float[2, 3];
					arr[0][1] := 3.14;
					Result := Integer(arr[0][1] * 100);
				end;
				Test();
			`,
			expected: 314,
		},
		{
			name: "2D array - iteration",
			input: `
				function Test: Integer;
				var arr: array of array of Integer;
				var i, j, sum: Integer;
				begin
					arr := new Integer[3, 4];

					for i := 0 to 2 do
						for j := 0 to 3 do
							arr[i][j] := i * 10 + j;

					sum := 0;
					for i := 0 to 2 do
						for j := 0 to 3 do
							sum := sum + arr[i][j];

					Result := sum;
				end;
				Test();
			`,
			// Row 0: 0+1+2+3 = 6
			// Row 1: 10+11+12+13 = 46
			// Row 2: 20+21+22+23 = 86
			// Total: 138
			expected: 138,
		},
		{
			name: "2D array - different dimensions",
			input: `
				function Test: Integer;
				var arr: array of array of Integer;
				begin
					arr := new Integer[10, 5];
					Result := Length(arr) + Length(arr[0]);
				end;
				Test();
			`,
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			testIntegerValue(t, result, tt.expected)
		})
	}
}

// TestMultiDimensionalArrayErrors tests error cases for multi-dimensional arrays
func TestMultiDimensionalArrayErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name: "negative dimension",
			input: `
				var arr: array of array of Integer;
				arr := new Integer[3, -1];
			`,
			shouldError: true,
		},
		{
			name: "zero dimension",
			input: `
				var arr: array of array of Integer;
				arr := new Integer[3, 0];
			`,
			shouldError: true,
		},
		{
			name: "non-integer dimension",
			input: `
				var arr: array of array of Integer;
				arr := new Integer[3, 'hello'];
			`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			if tt.shouldError {
				if !isError(result) {
					t.Errorf("expected error, got %v", result)
				}
			}
		})
	}
}

// TestMultiDimensionalArrayWithStrings tests multi-dimensional string arrays
func TestMultiDimensionalArrayWithStrings(t *testing.T) {
	input := `
		function Test: Integer;
		var arr: array of array of String;
		var s: String;
		begin
			arr := new String[2, 3];
			arr[0][0] := 'hello';
			arr[1][2] := 'world';
			s := arr[0][0] + ' ' + arr[1][2];
			Result := Length(s);
		end;
		Test();
	`
	result := testEval(input)
	testIntegerValue(t, result, 11) // "hello world" = 11 chars
}
