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

func TestBuiltinInsert_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Insert in middle - fix typo 'Helo' -> 'Hello'",
			input: `
var s: String := "Helo";
begin
	Insert("l", s, 3);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Insert at beginning",
			input: `
var s: String := "orld";
begin
	Insert("W", s, 1);
	s;
end
			`,
			expected: "World",
		},
		{
			name: "Insert at end",
			input: `
var s: String := "Hello";
begin
	Insert(" World", s, 6);
	s;
end
			`,
			expected: "Hello World",
		},
		{
			name: "Insert multiple characters",
			input: `
var s: String := "DWScript";
begin
	Insert("---", s, 3);
	s;
end
			`,
			expected: "DW---Script",
		},
		{
			name: "Insert empty string (no-op)",
			input: `
var s: String := "Hello";
begin
	Insert("", s, 3);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Insert into empty string",
			input: `
var s: String := "";
begin
	Insert("Hello", s, 1);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Multiple Insert operations",
			input: `
var s: String := "ac";
begin
	Insert("b", s, 2);
	Insert("d", s, 4);
	s;
end
			`,
			expected: "abcd",
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
				t.Errorf("Insert() result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinInsert_EdgeCases tests Insert() with edge cases and boundary conditions.
// Task 9.45: Insert() edge case tests
func TestBuiltinInsert_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Position < 1 - insert at beginning",
			input: `
var s: String := "Hello";
begin
	Insert("X", s, 0);
	s;
end
			`,
			expected: "XHello",
		},
		{
			name: "Position < 1 (negative) - insert at beginning",
			input: `
var s: String := "Hello";
begin
	Insert("X", s, -5);
	s;
end
			`,
			expected: "XHello",
		},
		{
			name: "Position > length - insert at end",
			input: `
var s: String := "Hello";
begin
	Insert("!", s, 100);
	s;
end
			`,
			expected: "Hello!",
		},
		{
			name: "Position exactly at length+1",
			input: `
var s: String := "Hello";
begin
	Insert("!", s, 6);
	s;
end
			`,
			expected: "Hello!",
		},
		{
			name: "Insert with variable source and position",
			input: `
var s: String := "Hello World";
var source: String := "Beautiful ";
var pos: Integer := 7;
begin
	Insert(source, s, pos);
	s;
end
			`,
			expected: "Hello Beautiful World",
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
				t.Errorf("Insert() result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinInsert_InExpressions tests using Insert() in various contexts.
// Task 9.45: Insert() expression tests
func TestBuiltinInsert_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Insert then concatenate",
			input: `
var s: String := "Hello";
begin
	Insert(" ", s, 6);
	s + "World";
end
			`,
			expected: "Hello World",
		},
		{
			name: "Insert with expression as position",
			input: `
var s: String := "abc";
var i: Integer := 2;
begin
	Insert("X", s, i + 1);
	s;
end
			`,
			expected: "abXc",
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
				t.Errorf("Insert() result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinInsert_ErrorCases tests error handling for Insert().
// Task 9.45: Insert() error tests
func TestBuiltinInsert_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Insert();
end
			`,
		},
		{
			name: "One argument only",
			input: `
var s: String := "Hello";
begin
	Insert("X", s);
end
			`,
		},
		{
			name: "Two arguments only",
			input: `
var s: String := "Hello";
begin
	Insert("X");
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
var s: String := "Hello";
begin
	Insert("X", s, 3, 4);
end
			`,
		},
		{
			name: "Non-string source",
			input: `
var s: String := "Hello";
var x: Integer := 42;
begin
	Insert(x, s, 3);
end
			`,
		},
		{
			name: "Non-identifier target (cannot modify literal)",
			input: `
begin
	Insert("X", "Hello", 3);
end
			`,
		},
		{
			name: "Non-integer position",
			input: `
var s: String := "Hello";
begin
	Insert("X", s, "3");
end
			`,
		},
		{
			name: "Target is not a string",
			input: `
var n: Integer := 42;
begin
	Insert("X", n, 1);
end
			`,
		},
		{
			name: "Undefined target variable",
			input: `
begin
	Insert("X", undefinedVar, 1);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid Insert() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinDelete_StringMode_BasicUsage tests Delete() for strings with basic operations.
// Delete(s, pos, count) - deletes count characters from s starting at 1-based position
// Task 9.45: Delete() string tests
func TestBuiltinDelete_StringMode_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Delete from middle",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, 2);
	s;
end
			`,
			expected: "Heo",
		},
		{
			name: "Delete from beginning",
			input: `
var s: String := "Hello";
begin
	Delete(s, 1, 2);
	s;
end
			`,
			expected: "llo",
		},
		{
			name: "Delete from end",
			input: `
var s: String := "Hello";
begin
	Delete(s, 4, 2);
	s;
end
			`,
			expected: "Hel",
		},
		{
			name: "Delete single character",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, 1);
	s;
end
			`,
			expected: "Helo",
		},
		{
			name: "Delete entire string",
			input: `
var s: String := "Hello";
begin
	Delete(s, 1, 5);
	s;
end
			`,
			expected: "",
		},
		{
			name: "Multiple Delete operations",
			input: `
var s: String := "abcdefgh";
begin
	Delete(s, 3, 2);
	Delete(s, 1, 1);
	s;
end
			`,
			expected: "befgh",
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
				t.Errorf("Delete() result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinDelete_StringMode_EdgeCases tests Delete() for strings with edge cases.
// Task 9.45: Delete() string edge case tests
func TestBuiltinDelete_StringMode_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Position < 1 - no-op",
			input: `
var s: String := "Hello";
begin
	Delete(s, 0, 2);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Position negative - no-op",
			input: `
var s: String := "Hello";
begin
	Delete(s, -5, 2);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Position > length - no-op",
			input: `
var s: String := "Hello";
begin
	Delete(s, 10, 2);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Count = 0 - no-op",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, 0);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Count negative - no-op",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, -5);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Count extends beyond string end",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, 100);
	s;
end
			`,
			expected: "He",
		},
		{
			name: "Delete from empty string - no-op",
			input: `
var s: String := "";
begin
	Delete(s, 1, 5);
	s;
end
			`,
			expected: "",
		},
		{
			name: "Delete with variable position and count",
			input: `
var s: String := "Hello World";
var pos: Integer := 6;
var cnt: Integer := 6;
begin
	Delete(s, pos, cnt);
	s;
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
				t.Errorf("Delete() result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinDelete_StringMode_InExpressions tests using Delete() in various contexts.
// Task 9.45: Delete() string expression tests
func TestBuiltinDelete_StringMode_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Delete then concatenate",
			input: `
var s: String := "HelloXXXWorld";
begin
	Delete(s, 6, 3);
	s;
end
			`,
			expected: "HelloWorld",
		},
		{
			name: "Delete with expression as position",
			input: `
var s: String := "abcXYZdef";
var i: Integer := 3;
begin
	Delete(s, i + 1, 3);
	s;
end
			`,
			expected: "abcdef",
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
				t.Errorf("Delete() result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinDelete_StringMode_ErrorCases tests error handling for Delete() string mode.
// Task 9.45: Delete() string error tests
func TestBuiltinDelete_StringMode_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Delete();
end
			`,
		},
		{
			name: "One argument only",
			input: `
var s: String := "Hello";
begin
	Delete(s);
end
			`,
		},
		{
			name: "Two arguments only (ambiguous - could be array or incomplete string)",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3);
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, 2, 1);
end
			`,
		},
		{
			name: "Non-identifier target (cannot modify literal)",
			input: `
begin
	Delete("Hello", 3, 2);
end
			`,
		},
		{
			name: "Non-integer position",
			input: `
var s: String := "Hello";
begin
	Delete(s, "3", 2);
end
			`,
		},
		{
			name: "Non-integer count",
			input: `
var s: String := "Hello";
begin
	Delete(s, 3, "2");
end
			`,
		},
		{
			name: "Target is not a string",
			input: `
var n: Integer := 42;
begin
	Delete(n, 1, 1);
end
			`,
		},
		{
			name: "Undefined target variable",
			input: `
begin
	Delete(undefinedVar, 1, 1);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid Delete() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinInsertAndDelete_Combined tests Insert() and Delete() used together.
// Task 9.45: Combined Insert() and Delete() tests
func TestBuiltinInsertAndDelete_Combined(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Insert then Delete",
			input: `
var s: String := "Hello";
begin
	Insert("XXX", s, 3);
	Delete(s, 3, 3);
	s;
end
			`,
			expected: "Hello",
		},
		{
			name: "Delete then Insert to repair",
			input: `
var s: String := "HelloXXXWorld";
begin
	Delete(s, 6, 3);
	Insert(" ", s, 6);
	s;
end
			`,
			expected: "Hello World",
		},
		{
			name: "Build string with Insert and Delete",
			input: `
var s: String := "";
begin
	Insert("abc", s, 1);
	Insert("def", s, 4);
	Insert("XXX", s, 4);
	Delete(s, 4, 3);
	s;
end
			`,
			expected: "abcdef",
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
				t.Errorf("Combined Insert/Delete result = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// ============================================================================
// Task 9.47: Tests for StringReplace() string function
// ============================================================================

// TestBuiltinStringReplace_BasicUsage tests StringReplace() with basic string replacements.
// StringReplace(str, old, new) - replaces all occurrences of old with new
// StringReplace(str, old, new, count) - replaces count occurrences (count=-1 means all)
// Task 9.47: StringReplace() basic tests

func TestBuiltinFormat_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple string formatting",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 1);
	arr[0] := "World";
	Format("Hello %s", arr);
end
			`,
			expected: "Hello World",
		},
		{
			name: "Integer formatting",
			input: `
type TIntArray = array of Integer;
var arr: TIntArray;
begin
	SetLength(arr, 1);
	arr[0] := 42;
	Format("Value: %d", arr);
end
			`,
			expected: "Value: 42",
		},
		{
			name: "Float formatting with precision",
			input: `
type TFloatArray = array of Float;
var arr: TFloatArray;
begin
	SetLength(arr, 1);
	arr[0] := 3.14159;
	Format("Pi: %.2f", arr);
end
			`,
			expected: "Pi: 3.14",
		},
		{
			name: "Multiple arguments with strings and integers",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 2);
	arr[0] := "Age";
	arr[1] := "25";
	Format("%s is %s", arr);
end
			`,
			expected: "Age is 25",
		},
		{
			name: "Multiple format specifiers",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 3);
	arr[0] := "John";
	arr[1] := "30";
	arr[2] := "5.9";
	Format("%s is %s years old and %s feet tall", arr);
end
			`,
			expected: "John is 30 years old and 5.9 feet tall",
		},
		{
			name: "Literal percent sign",
			input: `
type TIntArray = array of Integer;
var arr: TIntArray;
begin
	SetLength(arr, 1);
	arr[0] := 100;
	Format("100%% complete (%d)", arr);
end
			`,
			expected: "100% complete (100)",
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
				t.Errorf("Format() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinFormat_EdgeCases tests Format() with edge cases.
// Task 9.50: Format() edge case tests
func TestBuiltinFormat_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "No format specifiers",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 0);
	Format("Hello World", arr);
end
			`,
			expected: "Hello World",
		},
		{
			name: "Empty format string",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 0);
	Format("", arr);
end
			`,
			expected: "",
		},
		{
			name: "Zero value integer",
			input: `
type TIntArray = array of Integer;
var arr: TIntArray;
begin
	SetLength(arr, 1);
	arr[0] := 0;
	Format("Value: %d", arr);
end
			`,
			expected: "Value: 0",
		},
		{
			name: "Negative integer",
			input: `
type TIntArray = array of Integer;
var arr: TIntArray;
begin
	SetLength(arr, 1);
	arr[0] := -42;
	Format("Value: %d", arr);
end
			`,
			expected: "Value: -42",
		},
		{
			name: "Float with no decimal places",
			input: `
type TFloatArray = array of Float;
var arr: TFloatArray;
begin
	SetLength(arr, 1);
	arr[0] := 5.0;
	Format("Value: %.0f", arr);
end
			`,
			expected: "Value: 5",
		},
		{
			name: "Empty string argument",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 1);
	arr[0] := "";
	Format("Value: %s", arr);
end
			`,
			expected: "Value: ",
		},
		{
			name: "Width specifier for integer",
			input: `
type TIntArray = array of Integer;
var arr: TIntArray;
begin
	SetLength(arr, 1);
	arr[0] := 42;
	Format("Value: %5d", arr);
end
			`,
			expected: "Value:    42",
		},
		{
			name: "Width and precision for float",
			input: `
type TFloatArray = array of Float;
var arr: TFloatArray;
begin
	SetLength(arr, 1);
	arr[0] := 3.14159;
	Format("Value: %8.2f", arr);
end
			`,
			expected: "Value:     3.14",
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
				t.Errorf("Format() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinFormat_InExpressions tests using Format() in various contexts.
// Task 9.50: Format() expression tests
func TestBuiltinFormat_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Format in concatenation",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 1);
	arr[0] := "World";
	Format("Hello %s", arr) + "!";
end
			`,
			expected: "Hello World!",
		},
		{
			name: "Format with UpperCase",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 1);
	arr[0] := "world";
	UpperCase(Format("hello %s", arr));
end
			`,
			expected: "HELLO WORLD",
		},
		{
			name: "Nested Format calls",
			input: `
type TStrArray = array of String;
var arr1: TStrArray;
var arr2: TStrArray;
begin
	SetLength(arr1, 1);
	arr1[0] := "inner";
	SetLength(arr2, 1);
	arr2[0] := Format("(%s)", arr1);
	Format("outer %s", arr2);
end
			`,
			expected: "outer (inner)",
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
				t.Errorf("Format() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinFormat_ErrorCases tests error handling for Format().
// Task 9.50: Format() error tests
func TestBuiltinFormat_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	Format();
end
			`,
		},
		{
			name: "One argument only",
			input: `
begin
	Format("hello %s");
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 1);
	arr[0] := "test";
	Format("hello", arr, "extra");
end
			`,
		},
		{
			name: "First argument not a string",
			input: `
type TStrArray = array of String;
var x: Integer := 42;
var arr: TStrArray;
begin
	SetLength(arr, 0);
	Format(x, arr);
end
			`,
		},
		{
			name: "Second argument not an array",
			input: `
var x: String := "test";
begin
	Format("hello %s", x);
end
			`,
		},
		{
			name: "Too few arguments in array",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 0);
	Format("hello %s %s", arr);
end
			`,
		},
		{
			name: "Too many arguments in array",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 3);
	arr[0] := "a";
	arr[1] := "b";
	arr[2] := "c";
	Format("hello %s", arr);
end
			`,
		},
		{
			name: "Wrong type for format specifier",
			input: `
type TStrArray = array of String;
var arr: TStrArray;
begin
	SetLength(arr, 1);
	arr[0] := "not a number";
	Format("Value: %d", arr);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid Format() call, got %T: %+v", result, result)
			}
		})
	}
}

// ============================================================================
// Tests for SubStr() string function
// ============================================================================

// TestBuiltinSubStr_BasicUsage tests SubStr() with basic substring extraction.
// SubStr(str, start) - returns substring from start to end (1-based)
// SubStr(str, start, length) - returns length characters starting at start (1-based)
// Note: Different from Copy and SubString - SubStr takes length, SubString takes end position
func TestBuiltinSubStr_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "SubStr from beginning with length",
			input: `
begin
	SubStr("hello", 1, 2);
end
			`,
			expected: "he",
		},
		{
			name: "SubStr from middle with length",
			input: `
begin
	SubStr("hello", 2, 3);
end
			`,
			expected: "ell",
		},
		{
			name: "SubStr from position to end (no length)",
			input: `
begin
	SubStr("hello world", 7);
end
			`,
			expected: "world",
		},
		{
			name: "SubStr from beginning to end",
			input: `
begin
	SubStr("DWScript", 1);
end
			`,
			expected: "DWScript",
		},
		{
			name: "SubStr with string literal",
			input: `
begin
	SubStr("programming", 1, 7);
end
			`,
			expected: "program",
		},
		{
			name: "SubStr single character",
			input: `
begin
	SubStr("hello", 3, 1);
end
			`,
			expected: "l",
		},
		{
			name: "SubStr from factorize.pas use case",
			input: `
var s: String := " * 2 * 3";
begin
	SubStr(s, 4);
end
			`,
			expected: "2 * 3",
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
				t.Errorf("SubStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSubStr_EdgeCases tests SubStr() with edge cases and boundary conditions.
func TestBuiltinSubStr_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Empty string returns empty",
			input: `
begin
	SubStr("", 1, 5);
end
			`,
			expected: "",
		},
		{
			name: "Length beyond string length",
			input: `
begin
	SubStr("hello", 1, 100);
end
			`,
			expected: "hello",
		},
		{
			name: "Start beyond string length returns empty",
			input: `
begin
	SubStr("hello", 10, 5);
end
			`,
			expected: "",
		},
		{
			name: "Length is zero returns empty",
			input: `
begin
	SubStr("hello", 1, 0);
end
			`,
			expected: "",
		},
		{
			name: "Length is negative returns empty",
			input: `
begin
	SubStr("hello", 1, -5);
end
			`,
			expected: "",
		},
		{
			name: "Start at last character",
			input: `
begin
	SubStr("hello", 5, 1);
end
			`,
			expected: "o",
		},
		{
			name: "Start at last character with large length",
			input: `
begin
	SubStr("hello", 5, 100);
end
			`,
			expected: "o",
		},
		{
			name: "Start is zero returns empty",
			input: `
begin
	SubStr("hello", 0, 3);
end
			`,
			expected: "",
		},
		{
			name: "Start is negative returns empty",
			input: `
begin
	SubStr("hello", -1, 3);
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
				t.Errorf("SubStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSubStr_InExpressions tests using SubStr() in various expressions.
func TestBuiltinSubStr_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "SubStr result in string concatenation",
			input: `
begin
	SubStr("hello world", 1, 5) + "!";
end
			`,
			expected: "hello!",
		},
		{
			name: "SubStr with variable start and length",
			input: `
var s: String := "DWScript";
var start: Integer := 3;
var len: Integer := 6;
begin
	SubStr(s, start, len);
end
			`,
			expected: "Script",
		},
		{
			name: "SubStr with expression as start",
			input: `
var s: String := "testing";
var i: Integer := 2;
begin
	SubStr(s, i + 1, 4);
end
			`,
			expected: "stin",
		},
		{
			name: "Nested SubStr calls",
			input: `
begin
	SubStr(SubStr("hello world", 1, 5), 2, 3);
end
			`,
			expected: "ell",
		},
		{
			name: "SubStr with UpperCase",
			input: `
begin
	UpperCase(SubStr("hello world", 7));
end
			`,
			expected: "WORLD",
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
				t.Errorf("SubStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSubStr_ErrorCases tests error handling for SubStr().
func TestBuiltinSubStr_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	SubStr();
end
			`,
		},
		{
			name: "One argument only",
			input: `
begin
	SubStr("hello");
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	SubStr("hello", 1, 2, 3);
end
			`,
		},
		{
			name: "First argument not a string",
			input: `
var x: Integer := 42;
begin
	SubStr(x, 1, 2);
end
			`,
		},
		{
			name: "Second argument not an integer",
			input: `
begin
	SubStr("hello", "1", 2);
end
			`,
		},
		{
			name: "Third argument not an integer",
			input: `
begin
	SubStr("hello", 1, "2");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid SubStr() call, got %T: %+v", result, result)
			}
		})
	}
}

// ============================================================================
// Tests for PosEx() string function - edge cases for offset validation
// ============================================================================

// TestBuiltinPosEx_OffsetValidation tests PosEx() offset validation.
// This test verifies the fix for PR #40 comment - ensuring invalid offsets
// never return negative positions.

func TestBuiltinByteSizeToStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Small bytes",
			input: `
begin
	ByteSizeToStr(512);
end
			`,
			expected: "512 bytes",
		},
		{
			name: "Kilobytes",
			input: `
begin
	ByteSizeToStr(1536);
end
			`,
			expected: "1.50 KB",
		},
		{
			name: "Megabytes",
			input: `
begin
	ByteSizeToStr(1572864);
end
			`,
			expected: "1.50 MB",
		},
		{
			name: "Gigabytes",
			input: `
begin
	ByteSizeToStr(1610612736);
end
			`,
			expected: "1.50 GB",
		},
		{
			name: "Zero bytes",
			input: `
begin
	ByteSizeToStr(0);
end
			`,
			expected: "0 bytes",
		},
		{
			name: "One kilobyte exactly",
			input: `
begin
	ByteSizeToStr(1024);
end
			`,
			expected: "1.00 KB",
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
				t.Errorf("ByteSizeToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinGetText tests the GetText() built-in function.
func TestBuiltinGetText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic usage",
			input: `
begin
	GetText("Hello, World!");
end
			`,
			expected: "Hello, World!",
		},
		{
			name: "Empty string",
			input: `
begin
	GetText("");
end
			`,
			expected: "",
		},
		{
			name: "Unicode string",
			input: `
begin
	GetText("你好世界");
end
			`,
			expected: "你好世界",
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
				t.Errorf("GetText() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltin_ tests the _() built-in function (alias for GetText).
