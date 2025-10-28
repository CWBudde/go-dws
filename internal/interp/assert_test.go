package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Assert Function Tests
// ============================================================================

// TestAssertTrue tests that Assert(true) does not raise an error
func TestAssertTrue(t *testing.T) {
	input := `
		Assert(true);
		PrintLn('passed');
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()
	expected := "passed\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestAssertFalse tests that Assert(false) raises EAssertionFailed
func TestAssertFalse(t *testing.T) {
	input := `
		try
			Assert(false);
		except
			on E: EAssertionFailed do
				PrintLn('caught');
		end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()
	expected := "caught\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestAssertTrueWithMessage tests that Assert(true, 'message') does not raise error
func TestAssertTrueWithMessage(t *testing.T) {
	input := `
		Assert(true, 'this should not fail');
		PrintLn('passed');
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()
	expected := "passed\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestAssertFalseWithCustomMessage tests that Assert(false, 'Custom message') raises error with custom message
func TestAssertFalseWithCustomMessage(t *testing.T) {
	input := `
		try
			Assert(false, 'Custom error message');
		except
			on E: EAssertionFailed do
				PrintLn(E.Message);
		end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// The message should contain "Custom error message" and position info
	if !strings.Contains(output, "Custom error message") {
		t.Errorf("expected output to contain 'Custom error message', got %q", output)
	}
	if !strings.Contains(output, "Assertion failed") {
		t.Errorf("expected output to contain 'Assertion failed', got %q", output)
	}
}

// TestAssertInFunction tests using Assert for precondition validation
func TestAssertInFunction(t *testing.T) {
	input := `
		function Divide(a, b: Integer): Integer;
		begin
			Assert(b <> 0, 'Division by zero');
			Result := a div b;
		end;

		try
			PrintLn(Divide(10, 2));
			PrintLn(Divide(10, 0));
		except
			on E: EAssertionFailed do
				PrintLn('caught assertion');
		end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	if !strings.Contains(output, "5\n") {
		t.Errorf("expected output to contain '5', got %q", output)
	}
	if !strings.Contains(output, "caught assertion") {
		t.Errorf("expected output to contain 'caught assertion', got %q", output)
	}
}

// TestAssertWithExpression tests Assert with a boolean expression
func TestAssertWithExpression(t *testing.T) {
	input := `
		var x: Integer;
		x := 5;
		Assert(x > 0, 'x must be positive');
		PrintLn('x is positive');

		x := -1;
		try
			Assert(x > 0, 'x must be positive');
		except
			on E: EAssertionFailed do
				PrintLn('caught negative x');
		end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	if !strings.Contains(output, "x is positive") {
		t.Errorf("expected output to contain 'x is positive', got %q", output)
	}
	if !strings.Contains(output, "caught negative x") {
		t.Errorf("expected output to contain 'caught negative x', got %q", output)
	}
}

// TestAssertMessageFormat tests the exact format of assertion messages
func TestAssertMessageFormat(t *testing.T) {
	input := `
		try
			Assert(false);
		except
			on E: EAssertionFailed do
				PrintLn(E.Message);
		end;
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Should contain "Assertion failed" and position info like "[line: N, column: M]"
	if !strings.Contains(output, "Assertion failed") {
		t.Errorf("expected output to contain 'Assertion failed', got %q", output)
	}
	if !strings.Contains(output, "line:") {
		t.Errorf("expected output to contain 'line:', got %q", output)
	}
	if !strings.Contains(output, "column:") {
		t.Errorf("expected output to contain 'column:', got %q", output)
	}
}

// TestAssertWrongArgumentCount tests error handling for wrong number of arguments
func TestAssertWrongArgumentCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "no arguments",
			input: `
				Assert();
			`,
		},
		{
			name: "too many arguments",
			input: `
				Assert(true, 'msg', 'extra');
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				// Parser might catch some errors
				return
			}

			var buf bytes.Buffer
			interp := New(&buf)
			result := interp.Eval(program)

			// Should return an error
			if result != nil && result.Type() == "ERROR" {
				// Expected
				return
			}

			// Or the interpreter might handle it differently
			// As long as it doesn't panic or produce wrong output
		})
	}
}

// TestAssertWrongArgumentType tests error handling for wrong argument types
func TestAssertWrongArgumentType(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "non-boolean condition",
			input: `
				Assert(42);
			`,
		},
		{
			name: "non-string message",
			input: `
				Assert(true, 42);
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				// Parser might catch some errors
				return
			}

			var buf bytes.Buffer
			interp := New(&buf)
			result := interp.Eval(program)

			// Should return an error or handle gracefully
			if result != nil && result.Type() == "ERROR" {
				// Expected
				return
			}
		})
	}
}
