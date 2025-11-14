package interp

import "testing"

func TestBuiltinPos_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Find substring at beginning",
			input: `
begin
	Pos("He", "Hello World");
end
			`,
			expected: 1,
		},
		{
			name: "Find substring in middle",
			input: `
begin
	Pos("lo", "Hello World");
end
			`,
			expected: 4,
		},
		{
			name: "Find substring at end",
			input: `
begin
	Pos("rld", "Hello World");
end
			`,
			expected: 9,
		},
		{
			name: "Find single character",
			input: `
begin
	Pos("o", "Hello World");
end
			`,
			expected: 5,
		},
		{
			name: "Substring not found returns 0",
			input: `
begin
	Pos("xyz", "Hello World");
end
			`,
			expected: 0,
		},
		{
			name: "Find with variables",
			input: `
var s: String := "DWScript";
var sub: String := "Scr";
begin
	Pos(sub, s);
end
			`,
			expected: 3,
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
				t.Errorf("Pos() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinPos_EdgeCases tests Pos() with edge cases.
func TestBuiltinPos_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Empty substring returns 1 (found at start)",
			input: `
begin
	Pos("", "Hello");
end
			`,
			expected: 1,
		},
		{
			name: "Empty string with non-empty substring returns 0",
			input: `
begin
	Pos("He", "");
end
			`,
			expected: 0,
		},
		{
			name: "Both empty returns 1",
			input: `
begin
	Pos("", "");
end
			`,
			expected: 1,
		},
		{
			name: "Exact match returns 1",
			input: `
begin
	Pos("Hello", "Hello");
end
			`,
			expected: 1,
		},
		{
			name: "Substring longer than string returns 0",
			input: `
begin
	Pos("Hello World!", "Hello");
end
			`,
			expected: 0,
		},
		{
			name: "Case sensitive search - no match",
			input: `
begin
	Pos("hello", "Hello World");
end
			`,
			expected: 0,
		},
		{
			name: "Multiple occurrences - returns first",
			input: `
begin
	Pos("o", "Hello World");
end
			`,
			expected: 5,
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
				t.Errorf("Pos() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinPos_InExpressions tests using Pos() in expressions.
func TestBuiltinPos_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Pos in comparison",
			input: `
begin
	if Pos("World", "Hello World") > 0 then
		42
	else
		0;
end
			`,
			expected: 42,
		},
		{
			name: "Pos with Copy",
			input: `
var s: String := "Hello World";
var pos: Integer := Pos("World", s);
begin
	if pos > 0 then
		pos
	else
		0;
end
			`,
			expected: 7,
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
				t.Errorf("Pos() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinPos_ErrorCases tests error handling for Pos().
func TestBuiltinPos_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Pos();
end
			`,
		},
		{
			name: "One argument only",
			input: `
begin
	Pos("Hello");
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	Pos("He", "Hello", "extra");
end
			`,
		},
		{
			name: "First argument not a string",
			input: `
var x: Integer := 42;
begin
	Pos(x, "Hello");
end
			`,
		},
		{
			name: "Second argument not a string",
			input: `
var x: Integer := 42;
begin
	Pos("Hello", x);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid Pos() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinUpperCase_BasicUsage tests UpperCase() with basic string operations.
// UpperCase(str) - converts string to uppercase

func TestBuiltinPosEx_OffsetValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Empty needle with negative offset returns 0 (not negative)",
			input: `
begin
	PosEx("", "hello", -3);
end
			`,
			expected: 0,
		},
		{
			name: "Empty needle with zero offset returns 0",
			input: `
begin
	PosEx("", "hello", 0);
end
			`,
			expected: 0,
		},
		{
			name: "Empty needle with valid offset returns 0 (matches DWScript behavior)",
			input: `
begin
	PosEx("", "hello", 1);
end
			`,
			expected: 0,
		},
		{
			name: "Non-empty needle with negative offset returns 0",
			input: `
begin
	PosEx("test", "hello test", -1);
end
			`,
			expected: 0,
		},
		{
			name: "Non-empty needle with zero offset returns 0",
			input: `
begin
	PosEx("test", "hello test", 0);
end
			`,
			expected: 0,
		},
		{
			name: "Valid search with offset 1 (normal case)",
			input: `
begin
	PosEx("ll", "hello", 1);
end
			`,
			expected: 3,
		},
		{
			name: "Valid search with offset > 1",
			input: `
begin
	PosEx("o", "hello world", 6);
end
			`,
			expected: 8,
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
				t.Errorf("PosEx() = %d, want %d", intVal.Value, tt.expected)
			}

			// Additional check: result should never be negative
			if intVal.Value < 0 {
				t.Errorf("PosEx() returned negative position %d, should return 0 or positive", intVal.Value)
			}
		})
	}
}

// TestBuiltinStrFind_OffsetValidation tests StrFind() offset validation.
// StrFind is an alias for PosEx with reordered arguments, so it should also
// never return negative positions.
func TestBuiltinStrFind_OffsetValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Empty substring with negative offset returns 0",
			input: `
begin
	StrFind("hello", "", -3);
end
			`,
			expected: 0,
		},
		{
			name: "Empty substring with zero offset returns 0",
			input: `
begin
	StrFind("hello", "", 0);
end
			`,
			expected: 0,
		},
		{
			name: "Non-empty substring with negative offset returns 0",
			input: `
begin
	StrFind("hello test", "test", -1);
end
			`,
			expected: 0,
		},
		{
			name: "Non-empty substring with zero offset returns 0",
			input: `
begin
	StrFind("hello test", "test", 0);
end
			`,
			expected: 0,
		},
		{
			name: "Valid search with offset 1",
			input: `
begin
	StrFind("hello", "ll", 1);
end
			`,
			expected: 3,
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
				t.Errorf("StrFind() = %d, want %d", intVal.Value, tt.expected)
			}

			// Additional check: result should never be negative
			if intVal.Value < 0 {
				t.Errorf("StrFind() returned negative position %d, should return 0 or positive", intVal.Value)
			}
		})
	}
}

// TestBuiltinASCIIUpperCase tests the ASCIIUpperCase() built-in function.

func TestBuiltin_(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic usage",
			input: `
begin
	_("Hello, World!");
end
			`,
			expected: "Hello, World!",
		},
		{
			name: "Empty string",
			input: `
begin
	_("");
end
			`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("_() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCharAt tests the CharAt() built-in function.
func TestBuiltinCharAt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Get first character",
			input: `
begin
	CharAt("hello", 1);
end
			`,
			expected: "h",
		},
		{
			name: "Get middle character",
			input: `
begin
	CharAt("hello", 3);
end
			`,
			expected: "l",
		},
		{
			name: "Get last character",
			input: `
begin
	CharAt("hello", 5);
end
			`,
			expected: "o",
		},
		{
			name: "With Unicode",
			input: `
begin
	CharAt("你好", 2);
end
			`,
			expected: "好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("CharAt() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}
