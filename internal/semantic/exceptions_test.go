package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Exception Class Registration Tests
// ============================================================================

// Test that Exception base class is registered as a built-in type
func TestExceptionClassRegistered(t *testing.T) {
	analyzer := NewAnalyzer()

	// Exception should be registered as a built-in class
	// Use lowercase for case-insensitive lookup
	exceptionClass, exists := analyzer.GetClasses()["exception"]
	if !exists {
		t.Fatal("Exception class should be registered as a built-in type")
	}

	if exceptionClass.Name != "Exception" {
		t.Errorf("exceptionClass.Name = %s, want 'Exception'", exceptionClass.Name)
	}

	// Exception should have a Message field
	// Fields are stored in lowercase for case-insensitive lookup
	messageField, exists := exceptionClass.Fields["message"]
	if !exists {
		t.Fatal("Exception should have a 'message' field")
	}

	if messageField.String() != "String" {
		t.Errorf("Message field type = %s, want 'String'", messageField.String())
	}
}

// Test that standard exception types are registered
func TestStandardExceptionTypesRegistered(t *testing.T) {
	analyzer := NewAnalyzer()

	standardExceptions := []string{
		"EConvertError",
		"ERangeError",
		"EDivByZero",
		"EAssertionFailed",
		"EInvalidOp",
	}

	for _, excName := range standardExceptions {
		// Use lowercase for case-insensitive lookup
		excClass, exists := analyzer.GetClasses()[strings.ToLower(excName)]
		if !exists {
			t.Errorf("%s should be registered as a built-in exception type", excName)
			continue
		}

		// All standard exceptions should inherit from Exception
		if excClass.Parent == nil {
			t.Errorf("%s should have Exception as parent class", excName)
			continue
		}

		if excClass.Parent.Name != "Exception" {
			t.Errorf("%s parent class = %s, want 'Exception'", excName, excClass.Parent.Name)
		}
	}
}

// ============================================================================
// Raise Statement Semantic Analysis Tests
// ============================================================================

// Test raising an exception with constructor call
func TestRaiseExceptionWithConstructor(t *testing.T) {
	input := `
		raise Exception.Create('error message');
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic errors, got: %v", err)
	}
}

// Test raising an exception variable
func TestRaiseExceptionVariable(t *testing.T) {
	input := `
		var exc: Exception;
		exc := Exception.Create('error');
		raise exc;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic errors, got: %v", err)
	}
}

// Test bare raise validation (moved to semantic analysis)
func TestBareRaiseOutsideHandler(t *testing.T) {
	input := `
		raise;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Bare raise now validated at semantic analysis time
	if err == nil {
		t.Fatal("Expected semantic error for bare raise outside handler")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "bare raise") || !strings.Contains(errMsg, "exception handler") {
		t.Errorf("Expected error about bare raise context, got: %s", errMsg)
	}
}

// Test raising non-exception type (should error)
func TestRaiseNonExceptionType(t *testing.T) {
	input := `
		var x: Integer;
		x := 42;
		raise x;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for raising non-exception type")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Exception") {
		t.Errorf("Expected error about Exception type, got: %s", errMsg)
	}
}

// ============================================================================
// Try/Except Semantic Analysis Tests
// ============================================================================

// Test basic try/except structure
func TestTryExceptBasic(t *testing.T) {
	input := `
		try
			var x: Integer;
			x := 42;
		except
			on E: Exception do
				PrintLn('error');
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic errors, got: %v", err)
	}
}

// Test exception variable scoping in handler
func TestExceptionVariableScoping(t *testing.T) {
	input := `
		try
			raise Exception.Create('error');
		except
			on E: Exception do
				PrintLn(E.Message);
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic errors, got: %v", err)
	}
}

// Test invalid exception type in handler (should error)
func TestInvalidExceptionTypeInHandler(t *testing.T) {
	input := `
		try
			raise Exception.Create('error');
		except
			on E: Integer do
				PrintLn('error');
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for non-exception type in handler")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Exception") || !strings.Contains(errMsg, "Integer") {
		t.Errorf("Expected error about Exception type compatibility, got: %s", errMsg)
	}
}

// Test try without except or finally (should error)
func TestTryWithoutExceptOrFinally(t *testing.T) {
	// This should be caught by the parser, but let's verify
	// the semantic analyzer handles it gracefully
	input := `
		try
			var x: Integer;
		end;
	`

	// Parse should fail, but if it somehow succeeds, semantic should catch it
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Parser should already catch this error
	if len(p.Errors()) == 0 {
		t.Fatal("Expected parser error for try without except/finally")
	}

	// Still verify analyzer handles malformed AST gracefully
	_ = program
}

// Test duplicate exception handlers (should error)
func TestDuplicateExceptionHandlers(t *testing.T) {
	input := `
		try
			raise Exception.Create('error');
		except
			on E1: ERangeError do
				PrintLn('first handler');
			on E2: ERangeError do
				PrintLn('second handler');
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for duplicate exception handlers")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "duplicate") || !strings.Contains(errMsg, "ERangeError") {
		t.Errorf("Expected error about duplicate ERangeError handler, got: %s", errMsg)
	}
}

// Test exception variable is read-only (cannot reassign)
func TestExceptionVariableReadOnly(t *testing.T) {
	input := `
		try
			raise Exception.Create('error');
		except
			on E: Exception do begin
				E := Exception.Create('new error'); // Should fail - E is read-only
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for assigning to read-only exception variable")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "read-only") {
		t.Errorf("Expected error about read-only variable, got: %s", errMsg)
	}
}

// Test bare raise outside handler now produces semantic error
func TestBareRaiseOutsideHandlerSemanticError(t *testing.T) {
	input := `
		raise;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for bare raise outside handler")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "bare raise") || !strings.Contains(errMsg, "exception handler") {
		t.Errorf("Expected error about bare raise context, got: %s", errMsg)
	}
}

// Test raise is allowed in finally block (exception to the rule)
func TestRaiseInFinallyBlockAllowed(t *testing.T) {
	input := `
		try
			PrintLn('try');
		finally
			raise Exception.Create('cleanup error');  // Raise is allowed in finally
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Raise should be allowed in finally block, got error: %v", err)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func parseProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	return program
}
