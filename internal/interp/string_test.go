package interp

import "testing"

// TestBuiltinCopy_BasicUsage tests Copy() with basic string operations.
// Copy(str, index, count) - index is 1-based, count is number of characters
func TestBuiltinCopy_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Copy from beginning",
			input: `
var s: String := "hello";
begin
	Copy(s, 1, 2);
end
			`,
			expected: "he",
		},
		{
			name: "Copy from middle",
			input: `
var s: String := "hello";
begin
	Copy(s, 2, 3);
end
			`,
			expected: "ell",
		},
		{
			name: "Copy from end",
			input: `
var s: String := "hello";
begin
	Copy(s, 4, 2);
end
			`,
			expected: "lo",
		},
		{
			name: "Copy entire string",
			input: `
var s: String := "hello";
begin
	Copy(s, 1, 5);
end
			`,
			expected: "hello",
		},
		{
			name: "Copy with string literal",
			input: `
begin
	Copy("DWScript", 1, 2);
end
			`,
			expected: "DW",
		},
		{
			name: "Copy single character",
			input: `
begin
	Copy("hello", 3, 1);
end
			`,
			expected: "l",
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
				t.Errorf("Copy() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCopy_EdgeCases tests Copy() with edge cases and boundary conditions.
func TestBuiltinCopy_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Empty string returns empty",
			input: `
var s: String := "";
begin
	Copy(s, 1, 5);
end
			`,
			expected: "",
		},
		{
			name: "Count beyond string length",
			input: `
begin
	Copy("hello", 1, 100);
end
			`,
			expected: "hello",
		},
		{
			name: "Index beyond string length returns empty",
			input: `
begin
	Copy("hello", 10, 5);
end
			`,
			expected: "",
		},
		{
			name: "Count is zero returns empty",
			input: `
begin
	Copy("hello", 1, 0);
end
			`,
			expected: "",
		},
		{
			name: "Count is negative returns empty",
			input: `
begin
	Copy("hello", 1, -5);
end
			`,
			expected: "",
		},
		{
			name: "Index at last character",
			input: `
begin
	Copy("hello", 5, 1);
end
			`,
			expected: "o",
		},
		{
			name: "Index at last character with large count",
			input: `
begin
	Copy("hello", 5, 100);
end
			`,
			expected: "o",
		},
		{
			name: "Index is zero returns empty (0 is invalid, 1-based)",
			input: `
begin
	Copy("hello", 0, 3);
end
			`,
			expected: "",
		},
		{
			name: "Index is negative returns empty",
			input: `
begin
	Copy("hello", -1, 3);
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
				t.Errorf("Copy() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCopy_InExpressions tests using Copy() in various expressions.
func TestBuiltinCopy_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Copy result in string concatenation",
			input: `
var s: String := "hello";
begin
	Copy(s, 1, 2) + " world";
end
			`,
			expected: "he world",
		},
		{
			name: "Copy with variable index and count",
			input: `
var s: String := "DWScript";
var idx: Integer := 3;
var cnt: Integer := 6;
begin
	Copy(s, idx, cnt);
end
			`,
			expected: "Script",
		},
		{
			name: "Copy with expression as index",
			input: `
var s: String := "testing";
var i: Integer := 2;
begin
	Copy(s, i + 1, 4);
end
			`,
			expected: "stin",
		},
		{
			name: "Nested Copy calls",
			input: `
begin
	Copy(Copy("hello world", 1, 5), 2, 3);
end
			`,
			expected: "ell",
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
				t.Errorf("Copy() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCopy_ErrorCases tests error handling for Copy().
func TestBuiltinCopy_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Copy();
end
			`,
		},
		{
			name: "One argument only",
			input: `
begin
	Copy("hello");
end
			`,
		},
		{
			name: "Two arguments only",
			input: `
begin
	Copy("hello", 1);
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	Copy("hello", 1, 2, 3);
end
			`,
		},
		{
			name: "First argument not a string",
			input: `
var x: Integer := 42;
begin
	Copy(x, 1, 2);
end
			`,
		},
		{
			name: "Second argument not an integer",
			input: `
begin
	Copy("hello", "1", 2);
end
			`,
		},
		{
			name: "Third argument not an integer",
			input: `
begin
	Copy("hello", 1, "2");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid Copy() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinConcat_BasicUsage tests Concat() with basic string concatenation.
// Concat(str1, str2, ...) - concatenates multiple strings
func TestBuiltinConcat_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Concat two strings",
			input: `
begin
	Concat("Hello", " World");
end
			`,
			expected: "Hello World",
		},
		{
			name: "Concat three strings",
			input: `
begin
	Concat("Hello", " ", "World");
end
			`,
			expected: "Hello World",
		},
		{
			name: "Concat four strings",
			input: `
begin
	Concat("DW", "Scr", "ipt", "!");
end
			`,
			expected: "DWScript!",
		},
		{
			name: "Concat with variables",
			input: `
var s1: String := "Hello";
var s2: String := "World";
begin
	Concat(s1, " ", s2);
end
			`,
			expected: "Hello World",
		},
		{
			name: "Concat single string",
			input: `
begin
	Concat("Hello");
end
			`,
			expected: "Hello",
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
				t.Errorf("Concat() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinConcat_EdgeCases tests Concat() with edge cases.
func TestBuiltinConcat_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Concat with empty strings",
			input: `
begin
	Concat("", "Hello", "");
end
			`,
			expected: "Hello",
		},
		{
			name: "Concat all empty strings",
			input: `
begin
	Concat("", "", "");
end
			`,
			expected: "",
		},
		{
			name: "Concat many strings",
			input: `
begin
	Concat("a", "b", "c", "d", "e", "f", "g");
end
			`,
			expected: "abcdefg",
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
				t.Errorf("Concat() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinConcat_ErrorCases tests error handling for Concat().
func TestBuiltinConcat_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Concat();
end
			`,
		},
		{
			name: "Non-string argument",
			input: `
var x: Integer := 42;
begin
	Concat("Hello", x);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid Concat() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinPos_BasicUsage tests Pos() with basic substring search.
// Pos(substr, str) - returns 1-based position of first occurrence (0 if not found)
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
func TestBuiltinUpperCase_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Lowercase to uppercase",
			input: `
begin
	UpperCase("hello");
end
			`,
			expected: "HELLO",
		},
		{
			name: "Mixed case to uppercase",
			input: `
begin
	UpperCase("Hello World");
end
			`,
			expected: "HELLO WORLD",
		},
		{
			name: "Already uppercase",
			input: `
begin
	UpperCase("HELLO");
end
			`,
			expected: "HELLO",
		},
		{
			name: "With numbers and symbols",
			input: `
begin
	UpperCase("hello123!@#");
end
			`,
			expected: "HELLO123!@#",
		},
		{
			name: "With variable",
			input: `
var s: String := "dwscript";
begin
	UpperCase(s);
end
			`,
			expected: "DWSCRIPT",
		},
		{
			name: "Empty string",
			input: `
begin
	UpperCase("");
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
				t.Errorf("UpperCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLowerCase_BasicUsage tests LowerCase() with basic string operations.
// LowerCase(str) - converts string to lowercase
func TestBuiltinLowerCase_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Uppercase to lowercase",
			input: `
begin
	LowerCase("HELLO");
end
			`,
			expected: "hello",
		},
		{
			name: "Mixed case to lowercase",
			input: `
begin
	LowerCase("Hello World");
end
			`,
			expected: "hello world",
		},
		{
			name: "Already lowercase",
			input: `
begin
	LowerCase("hello");
end
			`,
			expected: "hello",
		},
		{
			name: "With numbers and symbols",
			input: `
begin
	LowerCase("HELLO123!@#");
end
			`,
			expected: "hello123!@#",
		},
		{
			name: "With variable",
			input: `
var s: String := "DWSCRIPT";
begin
	LowerCase(s);
end
			`,
			expected: "dwscript",
		},
		{
			name: "Empty string",
			input: `
begin
	LowerCase("");
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
				t.Errorf("LowerCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinUpperCase_InExpressions tests using UpperCase() in expressions.
func TestBuiltinUpperCase_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "UpperCase in concatenation",
			input: `
begin
	UpperCase("hello") + " WORLD";
end
			`,
			expected: "HELLO WORLD",
		},
		{
			name: "UpperCase with Copy",
			input: `
var s: String := "hello world";
begin
	UpperCase(Copy(s, 1, 5));
end
			`,
			expected: "HELLO",
		},
		{
			name: "Nested UpperCase",
			input: `
begin
	UpperCase(UpperCase("hello"));
end
			`,
			expected: "HELLO",
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
				t.Errorf("UpperCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLowerCase_InExpressions tests using LowerCase() in expressions.
func TestBuiltinLowerCase_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "LowerCase in concatenation",
			input: `
begin
	LowerCase("HELLO") + " world";
end
			`,
			expected: "hello world",
		},
		{
			name: "LowerCase with Copy",
			input: `
var s: String := "HELLO WORLD";
begin
	LowerCase(Copy(s, 1, 5));
end
			`,
			expected: "hello",
		},
		{
			name: "Nested LowerCase",
			input: `
begin
	LowerCase(LowerCase("HELLO"));
end
			`,
			expected: "hello",
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
				t.Errorf("LowerCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinUpperCase_ErrorCases tests error handling for UpperCase().
func TestBuiltinUpperCase_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	UpperCase();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	UpperCase("hello", "world");
end
			`,
		},
		{
			name: "Non-string argument",
			input: `
var x: Integer := 42;
begin
	UpperCase(x);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid UpperCase() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinLowerCase_ErrorCases tests error handling for LowerCase().
func TestBuiltinLowerCase_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	LowerCase();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	LowerCase("HELLO", "WORLD");
end
			`,
		},
		{
			name: "Non-string argument",
			input: `
var x: Integer := 42;
begin
	LowerCase(x);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid LowerCase() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinTrim_BasicUsage tests Trim() with basic string operations.
// Trim(str) - removes leading and trailing whitespace
// Task 9.42: Tests for Trim() function
func TestBuiltinTrim_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Trim spaces from both ends",
			input: `
begin
	Trim("  hello  ");
end
			`,
			expected: "hello",
		},
		{
			name: "Trim leading spaces only",
			input: `
begin
	Trim("  hello");
end
			`,
			expected: "hello",
		},
		{
			name: "Trim trailing spaces only",
			input: `
begin
	Trim("hello  ");
end
			`,
			expected: "hello",
		},
		{
			name: "Trim with tabs",
			input: `
begin
	Trim("	hello	");
end
			`,
			expected: "hello",
		},
		{
			name: "Trim with newlines",
			input: `
begin
	Trim("
hello
");
end
			`,
			expected: "hello",
		},
		{
			name: "Trim with mixed whitespace",
			input: `
begin
	Trim("
hello
");
end
			`,
			expected: "hello",
		},
		{
			name: "No whitespace to trim",
			input: `
begin
	Trim("hello");
end
			`,
			expected: "hello",
		},
		{
			name: "Empty string",
			input: `
begin
	Trim("");
end
			`,
			expected: "",
		},
		{
			name: "Only whitespace",
			input: `
begin
	Trim("   ");
end
			`,
			expected: "",
		},
		{
			name: "Whitespace in middle is preserved",
			input: `
begin
	Trim("  hello world  ");
end
			`,
			expected: "hello world",
		},
		{
			name: "With variable",
			input: `
var s: String := "  DWScript  ";
begin
	Trim(s);
end
			`,
			expected: "DWScript",
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
				t.Errorf("Trim() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTrimLeft_BasicUsage tests TrimLeft() with basic string operations.
// TrimLeft(str) - removes leading whitespace only
// Task 9.42: Tests for TrimLeft() function
func TestBuiltinTrimLeft_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "TrimLeft removes leading spaces",
			input: `
begin
	TrimLeft("  hello");
end
			`,
			expected: "hello",
		},
		{
			name: "TrimLeft preserves trailing spaces",
			input: `
begin
	TrimLeft("  hello  ");
end
			`,
			expected: "hello  ",
		},
		{
			name: "TrimLeft with tabs",
			input: `
begin
	TrimLeft("		hello");
end
			`,
			expected: "hello",
		},
		{
			name: "TrimLeft with newlines",
			input: `
begin
	TrimLeft("
hello");
end
			`,
			expected: "hello",
		},
		{
			name: "TrimLeft with mixed whitespace",
			input: `
begin
	TrimLeft("
hello");
end
			`,
			expected: "hello",
		},
		{
			name: "No leading whitespace",
			input: `
begin
	TrimLeft("hello  ");
end
			`,
			expected: "hello  ",
		},
		{
			name: "Empty string",
			input: `
begin
	TrimLeft("");
end
			`,
			expected: "",
		},
		{
			name: "Only whitespace",
			input: `
begin
	TrimLeft("   ");
end
			`,
			expected: "",
		},
		{
			name: "With variable",
			input: `
var s: String := "  DWScript";
begin
	TrimLeft(s);
end
			`,
			expected: "DWScript",
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
				t.Errorf("TrimLeft() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTrimRight_BasicUsage tests TrimRight() with basic string operations.
// TrimRight(str) - removes trailing whitespace only
// Task 9.42: Tests for TrimRight() function
func TestBuiltinTrimRight_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "TrimRight removes trailing spaces",
			input: `
begin
	TrimRight("hello  ");
end
			`,
			expected: "hello",
		},
		{
			name: "TrimRight preserves leading spaces",
			input: `
begin
	TrimRight("  hello  ");
end
			`,
			expected: "  hello",
		},
		{
			name: "TrimRight with tabs",
			input: `
begin
	TrimRight("hello		");
end
			`,
			expected: "hello",
		},
		{
			name: "TrimRight with newlines",
			input: `
begin
	TrimRight("hello
");
end
			`,
			expected: "hello",
		},
		{
			name: "TrimRight with mixed whitespace",
			input: `
begin
	TrimRight("hello
");
end
			`,
			expected: "hello",
		},
		{
			name: "No trailing whitespace",
			input: `
begin
	TrimRight("  hello");
end
			`,
			expected: "  hello",
		},
		{
			name: "Empty string",
			input: `
begin
	TrimRight("");
end
			`,
			expected: "",
		},
		{
			name: "Only whitespace",
			input: `
begin
	TrimRight("   ");
end
			`,
			expected: "",
		},
		{
			name: "With variable",
			input: `
var s: String := "DWScript  ";
begin
	TrimRight(s);
end
			`,
			expected: "DWScript",
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
				t.Errorf("TrimRight() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTrim_InExpressions tests using Trim functions in expressions.
// Task 9.42: Tests for Trim() in expressions
func TestBuiltinTrim_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Trim in concatenation",
			input: `
begin
	Trim("  hello  ") + " world";
end
			`,
			expected: "hello world",
		},
		{
			name: "Trim with Copy",
			input: `
var s: String := "  hello world  ";
begin
	Trim(Copy(s, 3, 5));
end
			`,
			expected: "hello",
		},
		{
			name: "Nested Trim",
			input: `
begin
	Trim(Trim("  hello  "));
end
			`,
			expected: "hello",
		},
		{
			name: "TrimLeft then TrimRight",
			input: `
begin
	TrimRight(TrimLeft("  hello  "));
end
			`,
			expected: "hello",
		},
		{
			name: "Trim with UpperCase",
			input: `
begin
	UpperCase(Trim("  hello  "));
end
			`,
			expected: "HELLO",
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
				t.Errorf("result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTrim_ErrorCases tests error handling for Trim functions.
// Task 9.42: Error handling tests
func TestBuiltinTrim_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		function string
	}{
		{
			name:     "Trim: No arguments",
			function: "Trim",
			input: `
begin
	Trim();
end
			`,
		},
		{
			name:     "Trim: Too many arguments",
			function: "Trim",
			input: `
begin
	Trim("hello", "world");
end
			`,
		},
		{
			name:     "Trim: Non-string argument",
			function: "Trim",
			input: `
var x: Integer := 42;
begin
	Trim(x);
end
			`,
		},
		{
			name:     "TrimLeft: No arguments",
			function: "TrimLeft",
			input: `
begin
	TrimLeft();
end
			`,
		},
		{
			name:     "TrimLeft: Too many arguments",
			function: "TrimLeft",
			input: `
begin
	TrimLeft("hello", "world");
end
			`,
		},
		{
			name:     "TrimLeft: Non-string argument",
			function: "TrimLeft",
			input: `
var x: Integer := 42;
begin
	TrimLeft(x);
end
			`,
		},
		{
			name:     "TrimRight: No arguments",
			function: "TrimRight",
			input: `
begin
	TrimRight();
end
			`,
		},
		{
			name:     "TrimRight: Too many arguments",
			function: "TrimRight",
			input: `
begin
	TrimRight("hello", "world");
end
			`,
		},
		{
			name:     "TrimRight: Non-string argument",
			function: "TrimRight",
			input: `
var x: Integer := 42;
begin
	TrimRight(x);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid %s() call, got %T: %+v", tt.function, result, result)
			}
		})
	}
}
