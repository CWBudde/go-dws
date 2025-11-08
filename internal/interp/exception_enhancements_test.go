package interp

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestGetStackTrace tests the GetStackTrace() built-in function
func TestGetStackTrace(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected []string // Expected strings in the output
	}{
		{
			name: "GetStackTrace in nested function calls",
			script: `
procedure Level3();
begin
  PrintLn(GetStackTrace());
end;

procedure Level2();
begin
  Level3();
end;

procedure Level1();
begin
  Level2();
end;

Level1();
`,
			expected: []string{
				"Level3",
				"Level2",
				"Level1",
			},
		},
		{
			name: "GetStackTrace at top level returns empty",
			script: `
var trace := GetStackTrace();
PrintLn('Trace length: ' + IntToStr(Length(trace)));
`,
			expected: []string{
				"Trace length: 0",
			},
		},
		{
			name: "GetStackTrace in recursive function",
			script: `
procedure Recurse(n: Integer);
begin
  if n = 0 then
  begin
    PrintLn('Stack depth: ' + IntToStr(n));
    var trace := GetStackTrace();
    // Count how many times Recurse appears in the trace
    var lines := 0;
    var i := 1;
    while i <= Length(trace) do
    begin
      if Pos('Recurse', trace) > 0 then
        lines := lines + 1;
      i := i + 1;
    end;
  end
  else
    Recurse(n - 1);
end;

Recurse(3);
`,
			expected: []string{
				"Stack depth: 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			interp := New(output)

			l := lexer.New(tt.script)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
			}

			result := interp.Eval(program)

			if result != nil && result.Type() == "ERROR" {
				t.Fatalf("Script returned error: %s", result.String())
			}

			outputStr := output.String()
			for _, expected := range tt.expected {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, outputStr)
				}
			}
		})
	}
}

// TestExceptionStackTrace tests that exceptions capture stack traces
func TestExceptionStackTrace(t *testing.T) {
	tests := []struct {
		name            string
		script          string
		expectedInStack []string
		expectException bool
	}{
		{
			name: "Exception captures stack trace",
			script: `
procedure Level3();
begin
  raise Exception.Create('Error from Level3');
end;

procedure Level2();
begin
  Level3();
end;

procedure Level1();
begin
  Level2();
end;

Level1();
`,
			expectException: true,
			expectedInStack: []string{"Level3", "Level2", "Level1"},
		},
		{
			name: "Exception in try-except preserves stack trace",
			script: `
procedure ThrowError();
begin
  raise Exception.Create('Test error');
end;

procedure CallThrow();
begin
  ThrowError();
end;

try
  CallThrow();
except
  on E: Exception do
  begin
    PrintLn('Caught: ' + E.Message);
  end;
end;
`,
			expectException: false,
			expectedInStack: []string{}, // Exception is caught, so no unhandled exception
		},
		{
			name: "Deeply nested exception",
			script: `
procedure D();
begin
  raise Exception.Create('Deep error');
end;

procedure C();
begin
  D();
end;

procedure B();
begin
  C();
end;

procedure A();
begin
  B();
end;

A();
`,
			expectException: true,
			expectedInStack: []string{"D", "C", "B", "A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			interp := New(output)

			l := lexer.New(tt.script)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
			}

			result := interp.Eval(program)

			if tt.expectException {
				// Check that we have an exception
				exc := interp.GetException()
				if exc == nil {
					t.Fatal("Expected script to raise exception, but it didn't")
				}

				// Verify expected function names appear in the stack trace
				stackStr := exc.CallStack.String()
				for _, expected := range tt.expectedInStack {
					if !strings.Contains(stackStr, expected) {
						t.Errorf("Expected stack trace to contain %q, got:\n%s", expected, stackStr)
					}
				}
			} else {
				// No exception expected
				exc := interp.GetException()
				if exc != nil {
					t.Fatalf("Script raised unexpected exception: %s", exc.Message)
				}

				if result != nil && result.Type() == "ERROR" {
					t.Fatalf("Script returned error: %s", result.String())
				}
			}
		})
	}
}

// TestGetStackTraceNoArguments tests that GetStackTrace() rejects arguments
func TestGetStackTraceNoArguments(t *testing.T) {
	script := `
PrintLn(GetStackTrace(123));
`
	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(script)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	result := interp.Eval(program)

	// Should fail during execution with an error
	if result == nil || result.Type() != "ERROR" {
		t.Fatal("Expected GetStackTrace to reject arguments")
	}
}

// TestStackTraceFormat tests that stack traces are formatted correctly
func TestStackTraceFormat(t *testing.T) {
	script := `
procedure TestFunc();
begin
  var trace := GetStackTrace();
  PrintLn(trace);
end;

TestFunc();
`
	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(script)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Script returned error: %s", result.String())
	}

	outputStr := output.String()

	// Stack trace should contain the function name and position info
	if !strings.Contains(outputStr, "TestFunc") {
		t.Errorf("Expected stack trace to contain 'TestFunc', got:\n%s", outputStr)
	}

	// Should have line and column information
	if !strings.Contains(outputStr, "line:") {
		t.Errorf("Expected stack trace to contain position info, got:\n%s", outputStr)
	}
}

// TestGetCallStack tests the GetCallStack() built-in function
func TestGetCallStack(t *testing.T) {
	tests := []struct {
		checkOutput   func(string) error
		name          string
		script        string
		expectedDepth int
	}{
		{
			name: "GetCallStack in nested function calls",
			script: `
procedure Level3();
var
  stack: array of Variant;
begin
  stack := GetCallStack();
  PrintLn('Depth: ' + IntToStr(Length(stack)));
end;

procedure Level2();
begin
  Level3();
end;

procedure Level1();
begin
  Level2();
end;

Level1();
`,
			expectedDepth: 3,
			checkOutput: func(output string) error {
				if !strings.Contains(output, "Depth: 3") {
					return fmt.Errorf("expected output to contain 'Depth: 3', got: %s", output)
				}
				return nil
			},
		},
		{
			name: "GetCallStack at top level returns empty",
			script: `
var stack := GetCallStack();
PrintLn('Top depth: ' + IntToStr(Length(stack)));
`,
			expectedDepth: 0,
			checkOutput: func(output string) error {
				if !strings.Contains(output, "Top depth: 0") {
					return fmt.Errorf("expected output to contain 'Top depth: 0', got: %s", output)
				}
				return nil
			},
		},
		{
			name: "GetCallStack depth matches GetStackTrace",
			script: `
procedure TestFunc();
var
  stack: array of Variant;
  trace: String;
begin
  stack := GetCallStack();
  trace := GetStackTrace();
  PrintLn('Stack depth: ' + IntToStr(Length(stack)));
  if Length(trace) > 0 then
    PrintLn('Trace not empty')
  else
    PrintLn('Trace empty');
end;

TestFunc();
`,
			expectedDepth: 1,
			checkOutput: func(output string) error {
				if !strings.Contains(output, "Stack depth: 1") {
					return fmt.Errorf("expected output to contain 'Stack depth: 1', got: %s", output)
				}
				if !strings.Contains(output, "Trace not empty") {
					return fmt.Errorf("expected trace to not be empty, got: %s", output)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			interp := New(output)

			l := lexer.New(tt.script)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
			}

			result := interp.Eval(program)

			if result != nil && result.Type() == "ERROR" {
				t.Fatalf("Script returned error: %s", result.String())
			}

			if tt.checkOutput != nil {
				if err := tt.checkOutput(output.String()); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestGetCallStackNoArguments tests that GetCallStack() rejects arguments
func TestGetCallStackNoArguments(t *testing.T) {
	script := `
PrintLn(GetCallStack(123));
`
	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(script)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	result := interp.Eval(program)

	// Should fail during execution with an error
	if result == nil || result.Type() != "ERROR" {
		t.Fatal("Expected GetCallStack to reject arguments")
	}
}

// TestGetCallStackReturnsRecords tests that GetCallStack returns structured data
func TestGetCallStackReturnsRecords(t *testing.T) {
	script := `
procedure TestFunc();
var
  stack: array of Variant;
begin
  stack := GetCallStack();
  // Just verify we can get the array without errors
  if Length(stack) > 0 then
    PrintLn('Stack captured')
  else
    PrintLn('Empty stack');
end;

TestFunc();
`
	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(script)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Script returned error: %s", result.String())
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Stack captured") {
		t.Errorf("Expected 'Stack captured' in output, got: %s", outputStr)
	}
}
