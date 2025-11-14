package interp

import "testing"

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

// ============================================================================
// Task 9.45: Tests for Insert() and Delete() string functions
// ============================================================================

// TestBuiltinInsert_BasicUsage tests Insert() with basic string insertions.
// Insert(source, target, pos) - inserts source into target at 1-based position
// Task 9.45: Insert() tests

func TestBuiltinStringReplace_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Replace all occurrences",
			input: `
begin
	StringReplace("hello world", "l", "L");
end
			`,
			expected: "heLLo worLd",
		},
		{
			name: "Replace single occurrence",
			input: `
begin
	StringReplace("hello", "e", "a");
end
			`,
			expected: "hallo",
		},
		{
			name: "Replace with longer string",
			input: `
begin
	StringReplace("go", "o", "ood");
end
			`,
			expected: "good",
		},
		{
			name: "Replace with shorter string",
			input: `
begin
	StringReplace("hello", "ello", "i");
end
			`,
			expected: "hi",
		},
		{
			name: "Replace with variable",
			input: `
var s: String := "DWScript is great";
begin
	StringReplace(s, "great", "awesome");
end
			`,
			expected: "DWScript is awesome",
		},
		{
			name: "Replace all with explicit count -1",
			input: `
begin
	StringReplace("foo bar foo", "foo", "baz", -1);
end
			`,
			expected: "baz bar baz",
		},
		{
			name: "Replace first occurrence only (count 1)",
			input: `
begin
	StringReplace("foo bar foo", "foo", "baz", 1);
end
			`,
			expected: "baz bar foo",
		},
		{
			name: "Replace first two occurrences (count 2)",
			input: `
begin
	StringReplace("a a a a", "a", "b", 2);
end
			`,
			expected: "b b a a",
		},
		{
			name: "Multiple word replacement",
			input: `
begin
	StringReplace("the cat and the dog", "the", "a");
end
			`,
			expected: "a cat and a dog",
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
				t.Errorf("StringReplace() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStringReplace_EdgeCases tests StringReplace() with edge cases.
// Task 9.47: StringReplace() edge case tests
func TestBuiltinStringReplace_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Empty old string returns original",
			input: `
begin
	StringReplace("hello", "", "X");
end
			`,
			expected: "hello",
		},
		{
			name: "Empty new string removes occurrences",
			input: `
begin
	StringReplace("hello", "l", "");
end
			`,
			expected: "heo",
		},
		{
			name: "Empty string returns empty",
			input: `
begin
	StringReplace("", "x", "y");
end
			`,
			expected: "",
		},
		{
			name: "Old string not found returns original",
			input: `
begin
	StringReplace("hello", "xyz", "abc");
end
			`,
			expected: "hello",
		},
		{
			name: "Replace entire string",
			input: `
begin
	StringReplace("hello", "hello", "goodbye");
end
			`,
			expected: "goodbye",
		},
		{
			name: "Old and new are the same (no-op)",
			input: `
begin
	StringReplace("hello", "l", "l");
end
			`,
			expected: "hello",
		},
		{
			name: "Case sensitive - no match",
			input: `
begin
	StringReplace("Hello", "hello", "hi");
end
			`,
			expected: "Hello",
		},
		{
			name: "Count is 0 - no replacement",
			input: `
begin
	StringReplace("hello", "l", "L", 0);
end
			`,
			expected: "hello",
		},
		{
			name: "Count is negative (not -1) - no replacement",
			input: `
begin
	StringReplace("hello", "l", "L", -5);
end
			`,
			expected: "hello",
		},
		{
			name: "Count exceeds occurrences - replace all found",
			input: `
begin
	StringReplace("hello", "l", "L", 100);
end
			`,
			expected: "heLLo",
		},
		{
			name: "Overlapping patterns",
			input: `
begin
	StringReplace("aaa", "aa", "b");
end
			`,
			expected: "ba",
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
				t.Errorf("StringReplace() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStringReplace_InExpressions tests using StringReplace() in various contexts.
// Task 9.47: StringReplace() expression tests
func TestBuiltinStringReplace_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "StringReplace in concatenation",
			input: `
begin
	StringReplace("hello", "l", "L") + " world";
end
			`,
			expected: "heLLo world",
		},
		{
			name: "Nested StringReplace",
			input: `
begin
	StringReplace(StringReplace("foo bar", "foo", "baz"), "bar", "qux");
end
			`,
			expected: "baz qux",
		},
		{
			name: "StringReplace with UpperCase",
			input: `
begin
	UpperCase(StringReplace("hello world", "world", "go"));
end
			`,
			expected: "HELLO GO",
		},
		{
			name: "StringReplace with Copy",
			input: `
var s: String := "hello world";
begin
	StringReplace(Copy(s, 1, 5), "l", "L");
end
			`,
			expected: "heLLo",
		},
		{
			name: "StringReplace with expression as count",
			input: `
var n: Integer := 1;
begin
	StringReplace("foo foo foo", "foo", "bar", n + 1);
end
			`,
			expected: "bar bar foo",
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
				t.Errorf("StringReplace() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStringReplace_ErrorCases tests error handling for StringReplace().
// Task 9.47: StringReplace() error tests
func TestBuiltinStringReplace_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	StringReplace();
end
			`,
		},
		{
			name: "One argument only",
			input: `
begin
	StringReplace("hello");
end
			`,
		},
		{
			name: "Two arguments only",
			input: `
begin
	StringReplace("hello", "l");
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	StringReplace("hello", "l", "L", 1, 2);
end
			`,
		},
		{
			name: "First argument not a string",
			input: `
var x: Integer := 42;
begin
	StringReplace(x, "4", "5");
end
			`,
		},
		{
			name: "Second argument not a string",
			input: `
var x: Integer := 42;
begin
	StringReplace("hello", x, "L");
end
			`,
		},
		{
			name: "Third argument not a string",
			input: `
var x: Integer := 42;
begin
	StringReplace("hello", "l", x);
end
			`,
		},
		{
			name: "Fourth argument not an integer",
			input: `
begin
	StringReplace("hello", "l", "L", "1");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid StringReplace() call, got %T: %+v", result, result)
			}
		})
	}
}

// ============================================================================
// Task 9.48-9.50: Tests for Format() string function
// ============================================================================

// TestBuiltinFormat_BasicUsage tests Format() with basic string formatting.
// Format(fmt, args) - formats a string with placeholders replaced by args
// Task 9.50: Format() basic tests

func TestBuiltinASCIIUpperCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic ASCII lowercase to uppercase",
			input: `
begin
	ASCIIUpperCase("hello");
end
			`,
			expected: "HELLO",
		},
		{
			name: "Mixed case ASCII",
			input: `
begin
	ASCIIUpperCase("HeLLo");
end
			`,
			expected: "HELLO",
		},
		{
			name: "Already uppercase",
			input: `
begin
	ASCIIUpperCase("HELLO");
end
			`,
			expected: "HELLO",
		},
		{
			name: "With numbers and symbols",
			input: `
begin
	ASCIIUpperCase("hello123!@#");
end
			`,
			expected: "HELLO123!@#",
		},
		{
			name: "Empty string",
			input: `
begin
	ASCIIUpperCase("");
end
			`,
			expected: "",
		},
		{
			name: "Non-ASCII characters unchanged",
			input: `
begin
	ASCIIUpperCase("café");
end
			`,
			expected: "CAFé", // Only ASCII 'c', 'a', 'f' converted, é unchanged
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
				t.Errorf("ASCIIUpperCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinASCIILowerCase tests the ASCIILowerCase() built-in function.
func TestBuiltinASCIILowerCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic ASCII uppercase to lowercase",
			input: `
begin
	ASCIILowerCase("HELLO");
end
			`,
			expected: "hello",
		},
		{
			name: "Mixed case ASCII",
			input: `
begin
	ASCIILowerCase("HeLLo");
end
			`,
			expected: "hello",
		},
		{
			name: "Already lowercase",
			input: `
begin
	ASCIILowerCase("hello");
end
			`,
			expected: "hello",
		},
		{
			name: "With numbers and symbols",
			input: `
begin
	ASCIILowerCase("HELLO123!@#");
end
			`,
			expected: "hello123!@#",
		},
		{
			name: "Empty string",
			input: `
begin
	ASCIILowerCase("");
end
			`,
			expected: "",
		},
		{
			name: "Non-ASCII characters unchanged",
			input: `
begin
	ASCIILowerCase("CAFÉ");
end
			`,
			expected: "cafÉ", // Only ASCII 'C', 'A', 'F' converted, É unchanged
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
				t.Errorf("ASCIILowerCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinAnsiUpperCase tests the AnsiUpperCase() built-in function.
func TestBuiltinAnsiUpperCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic usage",
			input: `
begin
	AnsiUpperCase("hello");
end
			`,
			expected: "HELLO",
		},
		{
			name: "With Unicode",
			input: `
begin
	AnsiUpperCase("café");
end
			`,
			expected: "CAFÉ",
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
				t.Errorf("AnsiUpperCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinAnsiLowerCase tests the AnsiLowerCase() built-in function.
func TestBuiltinAnsiLowerCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic usage",
			input: `
begin
	AnsiLowerCase("HELLO");
end
			`,
			expected: "hello",
		},
		{
			name: "With Unicode",
			input: `
begin
	AnsiLowerCase("CAFÉ");
end
			`,
			expected: "café",
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
				t.Errorf("AnsiLowerCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinByteSizeToStr tests the ByteSizeToStr() built-in function.
