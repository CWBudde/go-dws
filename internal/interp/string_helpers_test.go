package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// String Helper Method Tests
// ============================================================================

func TestStringHelper_ToUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic ToUpper",
			input:    "PrintLn('hello'.ToUpper);",
			expected: "HELLO\n",
		},
		{
			name:     "ToUpper with variable",
			input:    "var s := 'world'; PrintLn(s.ToUpper);",
			expected: "WORLD\n",
		},
		{
			name:     "ToUpper with parens",
			input:    "PrintLn('test'.ToUpper());",
			expected: "TEST\n",
		},
		{
			name:     "ToUpper already uppercase",
			input:    "PrintLn('HELLO'.ToUpper);",
			expected: "HELLO\n",
		},
		{
			name:     "ToUpper mixed case",
			input:    "PrintLn('HeLLo WoRLd'.ToUpper);",
			expected: "HELLO WORLD\n",
		},
		{
			name:     "ToUpper empty string",
			input:    "PrintLn(''.ToUpper);",
			expected: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_ToLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic ToLower",
			input:    "PrintLn('HELLO'.ToLower);",
			expected: "hello\n",
		},
		{
			name:     "ToLower with variable",
			input:    "var s := 'WORLD'; PrintLn(s.ToLower);",
			expected: "world\n",
		},
		{
			name:     "ToLower already lowercase",
			input:    "PrintLn('hello'.ToLower);",
			expected: "hello\n",
		},
		{
			name:     "ToLower mixed case",
			input:    "PrintLn('HeLLo WoRLd'.ToLower);",
			expected: "hello world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_ToInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic ToInteger",
			input:    "PrintLn('123'.ToInteger);",
			expected: "123\n",
		},
		{
			name:     "ToInteger with variable",
			input:    "var s := '456'; PrintLn(s.ToInteger);",
			expected: "456\n",
		},
		{
			name:     "ToInteger negative",
			input:    "PrintLn('-789'.ToInteger);",
			expected: "-789\n",
		},
		{
			name:     "ToInteger in arithmetic",
			input:    "PrintLn('10'.ToInteger + '20'.ToInteger);",
			expected: "30\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_ToFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic ToFloat",
			input:    "PrintLn('3.14'.ToFloat);",
			expected: "3.14\n",
		},
		{
			name:     "ToFloat with variable",
			input:    "var s := '2.718'; PrintLn(s.ToFloat);",
			expected: "2.718\n",
		},
		{
			name:     "ToFloat from integer string",
			input:    "PrintLn('42'.ToFloat);",
			expected: "42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_ToString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic ToString",
			input:    "PrintLn('hello'.ToString);",
			expected: "hello\n",
		},
		{
			name:     "ToString identity",
			input:    "var s := 'world'; PrintLn(s.ToString);",
			expected: "world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_StartsWith(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "StartsWith true",
			input:    "PrintLn('hello'.StartsWith('he'));",
			expected: "True\n",
		},
		{
			name:     "StartsWith false",
			input:    "PrintLn('hello'.StartsWith('wo'));",
			expected: "False\n",
		},
		{
			name:     "StartsWith with variable",
			input:    "var s := 'world'; PrintLn(s.StartsWith('wo'));",
			expected: "True\n",
		},
		{
			name:     "StartsWith empty prefix",
			input:    "PrintLn('test'.StartsWith(''));",
			expected: "False\n",
		},
		{
			name:     "StartsWith longer than string",
			input:    "PrintLn('hi'.StartsWith('hello'));",
			expected: "False\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_EndsWith(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "EndsWith true",
			input:    "PrintLn('hello'.EndsWith('lo'));",
			expected: "True\n",
		},
		{
			name:     "EndsWith false",
			input:    "PrintLn('hello'.EndsWith('he'));",
			expected: "False\n",
		},
		{
			name:     "EndsWith with variable",
			input:    "var s := 'world'; PrintLn(s.EndsWith('ld'));",
			expected: "True\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_Contains(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Contains true",
			input:    "PrintLn('hello'.Contains('ll'));",
			expected: "True\n",
		},
		{
			name:     "Contains false",
			input:    "PrintLn('hello'.Contains('xyz'));",
			expected: "False\n",
		},
		{
			name:     "Contains with variable",
			input:    "var s := 'world'; PrintLn(s.Contains('or'));",
			expected: "True\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_IndexOf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IndexOf found",
			input:    "PrintLn('hello'.IndexOf('ll'));",
			expected: "3\n",
		},
		{
			name:     "IndexOf not found",
			input:    "PrintLn('hello'.IndexOf('xyz'));",
			expected: "0\n",
		},
		{
			name:     "IndexOf at start",
			input:    "PrintLn('hello'.IndexOf('he'));",
			expected: "1\n",
		},
		{
			name:     "IndexOf with variable",
			input:    "var s := 'world'; PrintLn(s.IndexOf('or'));",
			expected: "2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_Copy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Copy with 2 params",
			input:    "PrintLn('hello'.Copy(2, 3));",
			expected: "ell\n",
		},
		{
			name:     "Copy with 1 param",
			input:    "PrintLn('hello'.Copy(3));",
			expected: "llo\n",
		},
		{
			name:     "Copy from start",
			input:    "PrintLn('world'.Copy(1, 3));",
			expected: "wor\n",
		},
		{
			name:     "Copy with variable",
			input:    "var s := 'test'; PrintLn(s.Copy(2, 2));",
			expected: "es\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_Before(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Before found",
			input:    "PrintLn('hello world'.Before(' '));",
			expected: "hello\n",
		},
		{
			name:     "Before not found returns original string",
			input:    "PrintLn('hello'.Before('x'));",
			expected: "hello\n",
		},
		{
			name:     "Before with variable",
			input:    "var s := 'test@example.com'; PrintLn(s.Before('@'));",
			expected: "test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_After(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "After found",
			input:    "PrintLn('hello world'.After(' '));",
			expected: "world\n",
		},
		{
			name:     "After not found",
			input:    "PrintLn('hello'.After('x'));",
			expected: "\n",
		},
		{
			name:     "After with variable",
			input:    "var s := 'test@example.com'; PrintLn(s.After('@'));",
			expected: "example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_Split(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Split basic",
			input: `
				var parts := 'a,b,c'.Split(',');
				PrintLn(parts[0]);
				PrintLn(parts[1]);
				PrintLn(parts[2]);
			`,
			expected: "a\nb\nc\n",
		},
		{
			name: "Split with space",
			input: `
				var parts := 'hello world'.Split(' ');
				PrintLn(parts[0]);
				PrintLn(parts[1]);
			`,
			expected: "hello\nworld\n",
		},
		{
			name: "Split with variable",
			input: `
				var s := 'one;two;three';
				var parts := s.Split(';');
				PrintLn(parts.Length);
			`,
			expected: "3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_Chaining(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Copy then ToUpper",
			input:    "PrintLn('hello world'.Copy(1, 5).ToUpper);",
			expected: "HELLO\n",
		},
		{
			name:     "ToLower then Copy",
			input:    "PrintLn('HELLO'.ToLower.Copy(2, 3));",
			expected: "ell\n",
		},
		{
			name:     "Before then ToUpper",
			input:    "PrintLn('hello world'.Before(' ').ToUpper);",
			expected: "HELLO\n",
		},
		{
			name:     "After then ToLower",
			input:    "PrintLn('HELLO WORLD'.After(' ').ToLower);",
			expected: "world\n",
		},
		{
			name:     "Triple chain",
			input:    "PrintLn('HELLO WORLD'.Before(' ').Copy(2, 3).ToLower);",
			expected: "ell\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Helper in concatenation",
			input:    "PrintLn('hello'.ToUpper + ' ' + 'world'.ToLower);",
			expected: "HELLO world\n",
		},
		{
			name:     "Helper in if condition",
			input:    "if 'test'.StartsWith('te') then PrintLn('yes') else PrintLn('no');",
			expected: "yes\n",
		},
		{
			name: "Helper in while condition",
			input: `
				var s := 'test';
				var count := 0;
				while s.Contains('t') and (count < 1) do begin
					PrintLn('found');
					count := count + 1;
				end;
			`,
			expected: "found\n",
		},
		{
			name: "Helper in assignment",
			input: `
				var s: String;
				s := 'hello'.ToUpper;
				PrintLn(s);
			`,
			expected: "HELLO\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

func TestStringHelper_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string ToUpper",
			input:    "PrintLn(''.ToUpper);",
			expected: "\n",
		},
		{
			name:     "Empty string Copy",
			input:    "PrintLn(''.Copy(1, 1));",
			expected: "\n",
		},
		{
			name:     "StartsWith empty string",
			input:    "PrintLn('test'.StartsWith(''));",
			expected: "False\n",
		},
		{
			name:     "EndsWith empty string",
			input:    "PrintLn('test'.EndsWith(''));",
			expected: "False\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runInterpreterTest(t, tt.input, tt.expected)
		})
	}
}

// Helper function to run an interpreter test
func runInterpreterTest(t *testing.T, input, expected string) {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	val := interp.Eval(program)

	// Check for errors
	if errVal, ok := val.(*ErrorValue); ok {
		t.Fatalf("evaluation error: %s", errVal.Message)
	}

	output := buf.String()
	if output != expected {
		t.Errorf("output mismatch:\ngot:  %q\nwant: %q", output, expected)
	}
}
