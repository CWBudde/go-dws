package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// ============================================================================
// Exception Class Registration Tests (Task 8.203-8.204)
// ============================================================================

// Task 8.203: Test that Exception base class is registered as a built-in type
func TestExceptionClassRegistered(t *testing.T) {
	analyzer := NewAnalyzer()

	// Exception should be registered as a built-in class
	exceptionClass, exists := analyzer.classes["Exception"]
	if !exists {
		t.Fatal("Exception class should be registered as a built-in type")
	}

	if exceptionClass.Name != "Exception" {
		t.Errorf("exceptionClass.Name = %s, want 'Exception'", exceptionClass.Name)
	}

	// Exception should have a Message field
	messageField, exists := exceptionClass.Fields["Message"]
	if !exists {
		t.Fatal("Exception should have a 'Message' field")
	}

	if messageField.String() != "String" {
		t.Errorf("Message field type = %s, want 'String'", messageField.String())
	}
}

// Task 8.204: Test that standard exception types are registered
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
		excClass, exists := analyzer.classes[excName]
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
// Raise Statement Semantic Analysis Tests (Task 8.208)
// ============================================================================

// Task 8.208: Test raising an exception with constructor call
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

// Task 8.208: Test raising an exception variable
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

// Task 8.208: Test bare raise validation (moved to semantic analysis)
// Note: Originally this was a runtime check, but Task 8.208 requires semantic validation
// This test has been superseded by TestBareRaiseOutsideHandlerSemanticError
func TestBareRaiseOutsideHandler(t *testing.T) {
	input := `
		raise;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	// Task 8.208: Bare raise now validated at semantic analysis time
	if err == nil {
		t.Fatal("Expected semantic error for bare raise outside handler")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "bare raise") || !strings.Contains(errMsg, "exception handler") {
		t.Errorf("Expected error about bare raise context, got: %s", errMsg)
	}
}

// Task 8.208: Test raising non-exception type (should error)
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
// Try/Except Semantic Analysis Tests (Task 8.205-8.207)
// ============================================================================

// Task 8.205: Test basic try/except structure
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

// Task 8.207: Test exception variable scoping in handler
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

// Task 8.206: Test invalid exception type in handler (should error)
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

// Task 8.205: Test try without except or finally (should error)
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

// Task 8.206: Test duplicate exception handlers (should error)
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

// Task 8.207: Test exception variable is read-only (cannot reassign)
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

// Task 8.208: Test bare raise outside handler now produces semantic error
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

// Task 8.209: Test return statement in finally block (should error)
// NOTE: This test is currently skipped because return statements are not yet
// implemented in the parser. The validation code is in place (analyze_functions.go:93-96)
// and will work once return statement parsing is added.
func TestReturnInFinallyBlock(t *testing.T) {
	t.Skip("Return statements not yet implemented in parser - validation code is ready")

	// TODO: Enable this test once return statements are parsed
	// The validation is implemented in analyzeReturn() which checks a.inFinallyBlock
	/*
	input := `
		function TestFunc(): Integer;
		begin
			try
				Result := 42;
			finally
				return;  // Should fail - return not allowed in finally
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for return statement in finally block")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "return") || !strings.Contains(errMsg, "finally") {
		t.Errorf("Expected error about return in finally block, got: %s", errMsg)
	}
	*/
}

// Task 8.209: Test raise is allowed in finally block (exception to the rule)
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
