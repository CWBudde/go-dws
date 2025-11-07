package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestIncInteger tests Inc() with integer variables
func TestIncInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Inc by 1",
			input: `
var x: Integer := 5;
Inc(x);
PrintLn(x);
`,
			expected: "6\n",
		},
		{
			name: "Inc multiple times",
			input: `
var x: Integer := 0;
Inc(x);
Inc(x);
Inc(x);
PrintLn(x);
`,
			expected: "3\n",
		},
		{
			name: "Inc negative number",
			input: `
var x: Integer := -5;
Inc(x);
PrintLn(x);
`,
			expected: "-4\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestIncIntegerDelta tests Inc() with custom delta values
func TestIncIntegerDelta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Inc by 3",
			input: `
var x: Integer := 5;
Inc(x, 3);
PrintLn(x);
`,
			expected: "8\n",
		},
		{
			name: "Inc by 10",
			input: `
var x: Integer := 0;
Inc(x, 10);
PrintLn(x);
`,
			expected: "10\n",
		},
		{
			name: "Inc by negative delta",
			input: `
var x: Integer := 10;
Inc(x, -5);
PrintLn(x);
`,
			expected: "5\n",
		},
		{
			name: "Inc by variable delta",
			input: `
var x: Integer := 5;
var delta: Integer := 7;
Inc(x, delta);
PrintLn(x);
`,
			expected: "12\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestIncEnum tests Inc() with enum values
func TestIncEnum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Inc enum basic",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
Inc(c);
PrintLn(Ord(c));
`,
			expected: "1\n",
		},
		{
			name: "Inc enum multiple times",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
Inc(c);
Inc(c);
PrintLn(Ord(c));
`,
			expected: "2\n",
		},
		{
			name: "Inc enum with explicit values",
			input: `
type TStatus = (Ok = 0, Warning = 5, Error = 10);
var s: TStatus := Ok;
Inc(s);
PrintLn(Ord(s));
`,
			expected: "5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestIncEnumBoundary tests error when incrementing beyond enum maximum
func TestIncEnumBoundary(t *testing.T) {
	input := `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
Inc(c);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if !isError(result) {
		t.Fatalf("expected error when incrementing enum beyond maximum, got: %s", result.String())
	}

	// Check that error message contains the expected text
	errorMsg := result.String()
	expectedError := "Inc() cannot increment enum beyond its maximum value"
	if !contains(errorMsg, expectedError) {
		t.Errorf("expected error to contain %q, got %q", expectedError, errorMsg)
	}
}

// TestIncErrors tests various error cases
func TestIncErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Inc undefined variable",
			input: `
Inc(x);
`,
			expectedError: "undefined variable: x",
		},
		{
			name: "Inc wrong type - string",
			input: `
var s: String := "hello";
Inc(s);
`,
			expectedError: "Inc() expects Integer or Enum, got STRING",
		},
		{
			name: "Inc wrong type - float",
			input: `
var f: Float := 3.14;
Inc(f);
`,
			expectedError: "Inc() expects Integer or Enum, got FLOAT",
		},
		{
			name: "Inc with non-integer delta",
			input: `
var x: Integer := 5;
Inc(x, 3.14);
`,
			expectedError: "Inc() delta must be Integer, got FLOAT",
		},
		{
			name: "Inc with string delta",
			input: `
var x: Integer := 5;
Inc(x, "hello");
`,
			expectedError: "Inc() delta must be Integer, got STRING",
		},
		{
			name: "Inc too many arguments",
			input: `
var x: Integer := 5;
Inc(x, 1, 2);
`,
			expectedError: "Inc() expects 1-2 arguments, got 3",
		},
		{
			name: "Inc no arguments",
			input: `
Inc();
`,
			expectedError: "Inc() expects 1-2 arguments, got 0",
		},
		{
			name: "Inc enum with delta",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
Inc(c, 2);
`,
			expectedError: "Inc() with delta not supported for enum types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				// Some tests might have parser errors, which is fine
				// Just check if the error message contains what we expect
				for _, err := range p.Errors() {
					if err.Error() == tt.expectedError {
						return
					}
				}
				// Continue to interpreter to see if it catches the error
			}

			interp := New(&out)
			result := interp.Eval(program)

			if !isError(result) {
				t.Fatalf("expected error, got: %s", result.String())
			}

			// Check that error message contains the expected text
			errorMsg := result.String()
			if !contains(errorMsg, tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, errorMsg)
			}
		})
	}
}

// TestIncInLoop tests Inc() in a loop context
func TestIncInLoop(t *testing.T) {
	input := `
var sum: Integer := 0;
var i: Integer := 0;
while i < 5 do
begin
	Inc(sum, i);
	Inc(i);
end;
PrintLn(sum);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "10\n" // 0+1+2+3+4 = 10
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestIncWithExpression tests that Inc() requires a variable, not an expression
func TestIncWithExpression(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Inc with literal",
			input: `
Inc(5);
`,
			expectedError: "Inc() first argument must be a variable, got *ast.IntegerLiteral",
		},
		{
			name: "Inc with expression",
			input: `
var x: Integer := 5;
Inc(x + 1);
`,
			expectedError: "Inc() first argument must be a variable, got *ast.BinaryExpression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			// Parser should succeed, but interpreter should fail
			interp := New(&out)
			result := interp.Eval(program)

			if !isError(result) {
				t.Fatalf("expected error, got: %s", result.String())
			}

			// Check that error message contains the expected text
			errorMsg := result.String()
			if !contains(errorMsg, tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, errorMsg)
			}
		})
	}
}

// TestDecInteger tests Dec() with integer variables
func TestDecInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Dec by 1",
			input: `
var x: Integer := 5;
Dec(x);
PrintLn(x);
`,
			expected: "4\n",
		},
		{
			name: "Dec multiple times",
			input: `
var x: Integer := 10;
Dec(x);
Dec(x);
Dec(x);
PrintLn(x);
`,
			expected: "7\n",
		},
		{
			name: "Dec negative number",
			input: `
var x: Integer := -5;
Dec(x);
PrintLn(x);
`,
			expected: "-6\n",
		},
		{
			name: "Dec to zero",
			input: `
var x: Integer := 1;
Dec(x);
PrintLn(x);
`,
			expected: "0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestDecIntegerDelta tests Dec() with custom delta values
func TestDecIntegerDelta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Dec by 3",
			input: `
var x: Integer := 10;
Dec(x, 3);
PrintLn(x);
`,
			expected: "7\n",
		},
		{
			name: "Dec by 10",
			input: `
var x: Integer := 20;
Dec(x, 10);
PrintLn(x);
`,
			expected: "10\n",
		},
		{
			name: "Dec by negative delta (adds)",
			input: `
var x: Integer := 5;
Dec(x, -5);
PrintLn(x);
`,
			expected: "10\n",
		},
		{
			name: "Dec by variable delta",
			input: `
var x: Integer := 20;
var delta: Integer := 7;
Dec(x, delta);
PrintLn(x);
`,
			expected: "13\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestDecEnum tests Dec() with enum values
func TestDecEnum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Dec enum basic",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
Dec(c);
PrintLn(Ord(c));
`,
			expected: "1\n",
		},
		{
			name: "Dec enum multiple times",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
Dec(c);
Dec(c);
PrintLn(Ord(c));
`,
			expected: "0\n",
		},
		{
			name: "Dec enum with explicit values",
			input: `
type TStatus = (Ok = 0, Warning = 5, Error = 10);
var s: TStatus := Error;
Dec(s);
PrintLn(Ord(s));
`,
			expected: "5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestDecEnumBoundary tests error when decrementing below enum minimum
func TestDecEnumBoundary(t *testing.T) {
	input := `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
Dec(c);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if !isError(result) {
		t.Fatalf("expected error when decrementing enum below minimum, got: %s", result.String())
	}

	// Check that error message contains the expected text
	errorMsg := result.String()
	expectedError := "Dec() cannot decrement enum below its minimum value"
	if !contains(errorMsg, expectedError) {
		t.Errorf("expected error to contain %q, got %q", expectedError, errorMsg)
	}
}

// TestDecErrors tests various error cases for Dec
func TestDecErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Dec undefined variable",
			input: `
Dec(x);
`,
			expectedError: "undefined variable: x",
		},
		{
			name: "Dec wrong type - string",
			input: `
var s: String := "hello";
Dec(s);
`,
			expectedError: "Dec() expects Integer or Enum, got STRING",
		},
		{
			name: "Dec wrong type - float",
			input: `
var f: Float := 3.14;
Dec(f);
`,
			expectedError: "Dec() expects Integer or Enum, got FLOAT",
		},
		{
			name: "Dec with non-integer delta",
			input: `
var x: Integer := 10;
Dec(x, 3.14);
`,
			expectedError: "Dec() delta must be Integer, got FLOAT",
		},
		{
			name: "Dec with string delta",
			input: `
var x: Integer := 10;
Dec(x, "hello");
`,
			expectedError: "Dec() delta must be Integer, got STRING",
		},
		{
			name: "Dec too many arguments",
			input: `
var x: Integer := 10;
Dec(x, 1, 2);
`,
			expectedError: "Dec() expects 1-2 arguments, got 3",
		},
		{
			name: "Dec no arguments",
			input: `
Dec();
`,
			expectedError: "Dec() expects 1-2 arguments, got 0",
		},
		{
			name: "Dec enum with delta",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
Dec(c, 2);
`,
			expectedError: "Dec() with delta not supported for enum types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				// Some tests might have parser errors, which is fine
				// Just check if the error message contains what we expect
				for _, err := range p.Errors() {
					if err.Error() == tt.expectedError {
						return
					}
				}
				// Continue to interpreter to see if it catches the error
			}

			interp := New(&out)
			result := interp.Eval(program)

			if !isError(result) {
				t.Fatalf("expected error, got: %s", result.String())
			}

			// Check that error message contains the expected text
			errorMsg := result.String()
			if !contains(errorMsg, tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, errorMsg)
			}
		})
	}
}

// TestDecInLoop tests Dec() in a loop context
func TestDecInLoop(t *testing.T) {
	input := `
var sum: Integer := 0;
var i: Integer := 5;
while i > 0 do
begin
	sum := sum + i;
	Dec(i);
end;
PrintLn(sum);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "15\n" // 5+4+3+2+1 = 15
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestIncAndDecTogether tests using Inc and Dec together
func TestIncAndDecTogether(t *testing.T) {
	input := `
var x: Integer := 10;
Inc(x, 5);
PrintLn(x);
Dec(x, 3);
PrintLn(x);
Inc(x);
PrintLn(x);
Dec(x);
PrintLn(x);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "15\n12\n13\n12\n"
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// TestSuccInteger tests Succ() with integer values
func TestSuccInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Succ of 5",
			input: `
var x: Integer := 5;
var y: Integer := Succ(x);
PrintLn(y);
`,
			expected: "6\n",
		},
		{
			name: "Succ of 0",
			input: `
PrintLn(Succ(0));
`,
			expected: "1\n",
		},
		{
			name: "Succ of negative",
			input: `
PrintLn(Succ(-10));
`,
			expected: "-9\n",
		},
		{
			name: "Succ chain",
			input: `
var x: Integer := 5;
PrintLn(Succ(Succ(x)));
`,
			expected: "7\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestPredInteger tests Pred() with integer values
func TestPredInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Pred of 5",
			input: `
var x: Integer := 5;
var y: Integer := Pred(x);
PrintLn(y);
`,
			expected: "4\n",
		},
		{
			name: "Pred of 1",
			input: `
PrintLn(Pred(1));
`,
			expected: "0\n",
		},
		{
			name: "Pred of negative",
			input: `
PrintLn(Pred(-10));
`,
			expected: "-11\n",
		},
		{
			name: "Pred chain",
			input: `
var x: Integer := 10;
PrintLn(Pred(Pred(x)));
`,
			expected: "8\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestSuccEnum tests Succ() with enum values
func TestSuccEnum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Succ enum basic",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
var next: TColor := Succ(c);
PrintLn(Ord(next));
`,
			expected: "1\n",
		},
		{
			name: "Succ enum chained",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
PrintLn(Ord(Succ(Succ(c))));
`,
			expected: "2\n",
		},
		{
			name: "Succ enum with explicit values",
			input: `
type TStatus = (Ok = 0, Warning = 5, Error = 10);
var s: TStatus := Ok;
PrintLn(Ord(Succ(s)));
`,
			expected: "5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestPredEnum tests Pred() with enum values
func TestPredEnum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Pred enum basic",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
var prev: TColor := Pred(c);
PrintLn(Ord(prev));
`,
			expected: "1\n",
		},
		{
			name: "Pred enum chained",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
PrintLn(Ord(Pred(Pred(c))));
`,
			expected: "0\n",
		},
		{
			name: "Pred enum with explicit values",
			input: `
type TStatus = (Ok = 0, Warning = 5, Error = 10);
var s: TStatus := Error;
PrintLn(Ord(Pred(s)));
`,
			expected: "5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestSuccEnumBoundary tests error when calling Succ on maximum enum value
func TestSuccEnumBoundary(t *testing.T) {
	input := `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
var next: TColor := Succ(c);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if !isError(result) {
		t.Fatalf("expected error when calling Succ on maximum enum value, got: %s", result.String())
	}

	errorMsg := result.String()
	expectedError := "Succ() cannot get successor of maximum enum value"
	if !contains(errorMsg, expectedError) {
		t.Errorf("expected error to contain %q, got %q", expectedError, errorMsg)
	}
}

// TestPredEnumBoundary tests error when calling Pred on minimum enum value
func TestPredEnumBoundary(t *testing.T) {
	input := `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
var prev: TColor := Pred(c);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if !isError(result) {
		t.Fatalf("expected error when calling Pred on minimum enum value, got: %s", result.String())
	}

	errorMsg := result.String()
	expectedError := "Pred() cannot get predecessor of minimum enum value"
	if !contains(errorMsg, expectedError) {
		t.Errorf("expected error to contain %q, got %q", expectedError, errorMsg)
	}
}

// TestSuccPredErrors tests error cases for Succ/Pred
func TestSuccPredErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Succ wrong type - string",
			input: `
var s: String := "hello";
PrintLn(Succ(s));
`,
			expectedError: "Succ() expects Integer or Enum, got STRING",
		},
		{
			name: "Pred wrong type - float",
			input: `
var f: Float := 3.14;
PrintLn(Pred(f));
`,
			expectedError: "Pred() expects Integer or Enum, got FLOAT",
		},
		{
			name: "Succ too many arguments",
			input: `
PrintLn(Succ(5, 10));
`,
			expectedError: "Succ() expects exactly 1 argument, got 2",
		},
		{
			name: "Pred no arguments",
			input: `
PrintLn(Pred());
`,
			expectedError: "Pred() expects exactly 1 argument, got 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				// Some tests might have parser errors
				for _, err := range p.Errors() {
					if err.Error() == tt.expectedError {
						return
					}
				}
			}

			interp := New(&out)
			result := interp.Eval(program)

			if !isError(result) {
				t.Fatalf("expected error, got: %s", result.String())
			}

			errorMsg := result.String()
			if !contains(errorMsg, tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, errorMsg)
			}
		})
	}
}

// TestSuccPredWithIncDec tests combining all ordinal functions
func TestSuccPredWithIncDec(t *testing.T) {
	input := `
var x: Integer := 10;
PrintLn(Succ(x));
PrintLn(Pred(x));

Inc(x);
PrintLn(x);
var y: Integer := Succ(x);
PrintLn(y);

Dec(x);
PrintLn(x);
var z: Integer := Pred(x);
PrintLn(z);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "11\n9\n11\n12\n10\n9\n"
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}

// ============================================================================
// Low/High Tests for Enums
// ============================================================================

// TestLowEnumBasic tests Low() with enum values
func TestLowEnumBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Low of enum variable returns first value",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Blue;
var low: TColor := Low(c);
PrintLn(low);
`,
			expected: "Red\n",
		},
		{
			name: "Low of enum with middle value",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Green;
PrintLn(Low(c));
`,
			expected: "Red\n",
		},
		{
			name: "Low of enum with explicit values",
			input: `
type TStatus = (Ok = 0, Warning = 5, Error = 10);
var s: TStatus := Error;
PrintLn(Ord(Low(s)));
`,
			expected: "0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestHighEnumBasic tests High() with enum values
func TestHighEnumBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "High of enum variable returns last value",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
var high: TColor := High(c);
PrintLn(high);
`,
			expected: "Blue\n",
		},
		{
			name: "High of enum with middle value",
			input: `
type TColor = (Red, Green, Blue);
var c: TColor := Green;
PrintLn(High(c));
`,
			expected: "Blue\n",
		},
		{
			name: "High of enum with explicit values",
			input: `
type TStatus = (Ok = 0, Warning = 5, Error = 10);
var s: TStatus := Ok;
PrintLn(Ord(High(s)));
`,
			expected: "10\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestLowHighEnumCompatibility tests Low/High still work for arrays
func TestLowHighEnumCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Low of dynamic array still returns 0",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
SetLength(arr, 5);
PrintLn(Low(arr));
`,
			expected: "0\n",
		},
		{
			name: "High of dynamic array still returns length-1",
			input: `
type TDynArray = array of Integer;
var arr: TDynArray;
SetLength(arr, 5);
PrintLn(High(arr));
`,
			expected: "4\n",
		},
		{
			name: "Low of static array",
			input: `
type TStaticArray = array[1..5] of Integer;
var arr: TStaticArray;
PrintLn(Low(arr));
`,
			expected: "1\n",
		},
		{
			name: "High of static array",
			input: `
type TStaticArray = array[1..5] of Integer;
var arr: TStaticArray;
PrintLn(High(arr));
`,
			expected: "5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(&out)
			result := interp.Eval(program)

			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}

			if out.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, out.String())
			}
		})
	}
}

// TestLowHighEnumInForLoop tests using Low/High in for loops with enums
func TestLowHighEnumInForLoop(t *testing.T) {
	input := `
type TColor = (Red, Green, Blue);
var c: TColor := Red;
var count: Integer := 0;

c := Low(c);
while Ord(c) <= Ord(High(c)) do
begin
	count := count + 1;
	if Ord(c) < Ord(High(c)) then
		Inc(c)
	else
		break;
end;

PrintLn(count);
`

	var out bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(&out)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "3\n" // Red, Green, Blue = 3 values
	if out.String() != expected {
		t.Errorf("expected output %q, got %q", expected, out.String())
	}
}
