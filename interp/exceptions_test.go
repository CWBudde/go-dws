package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// ============================================================================
// Basic Exception Handling Tests (Task 8.219)
// ============================================================================

// TestBasicTryExcept tests raising and catching a basic exception
// TDD: RED - This test should fail because exception handling is not yet implemented
func TestBasicTryExcept(t *testing.T) {
	input := `
		var caught: Boolean;
		caught := false;

		try
			raise Exception.Create('test error');
		except
			on E: Exception do
				caught := true;
		end;

		PrintLn(caught);
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
	expected := "true\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestRaiseWithMessage tests creating and raising an exception with a message
func TestRaiseWithMessage(t *testing.T) {
	input := `
		var msg: String;
		msg := '';

		try
			raise Exception.Create('custom error message');
		except
			on E: Exception do
				msg := E.Message;
		end;

		PrintLn(msg);
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
	expected := "custom error message\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestSpecificExceptionType tests catching a specific exception type
func TestSpecificExceptionType(t *testing.T) {
	input := `
		var caught: String;
		caught := 'none';

		try
			raise ERangeError.Create('out of range');
		except
			on E: ERangeError do
				caught := 'ERangeError';
		end;

		PrintLn(caught);
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
	expected := "ERangeError\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Multiple Exception Handlers Tests (Task 8.219)
// ============================================================================

// TestMultipleHandlers tests catching different exception types with multiple handlers
func TestMultipleHandlers(t *testing.T) {
	input := `
		var caught: String;
		caught := 'none';

		try
			raise EConvertError.Create('conversion error');
		except
			on E: ERangeError do
				caught := 'ERangeError';
			on E: EConvertError do
				caught := 'EConvertError';
			on E: Exception do
				caught := 'Exception';
		end;

		PrintLn(caught);
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
	expected := "EConvertError\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Bare Except Tests (Task 8.219)
// ============================================================================

// TestBareExcept tests bare except (catch-all) without specific handler
func TestBareExcept(t *testing.T) {
	input := `
		var caught: Boolean;
		caught := false;

		try
			raise Exception.Create('any error');
		except
			caught := true;
		end;

		PrintLn(caught);
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
	expected := "true\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestUncaughtException tests that uncaught exceptions propagate to top level
func TestUncaughtException(t *testing.T) {
	input := `
		var executed: Boolean;
		executed := false;

		try
			raise EConvertError.Create('conversion error');
		except
			on E: ERangeError do
				executed := true;
		end;

		PrintLn('should not reach here');
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	err := interp.Eval(program)

	// Should have an error (uncaught exception)
	if err == nil {
		t.Fatal("expected uncaught exception error, got nil")
	}

	// Output should not include "should not reach here"
	output := buf.String()
	if strings.Contains(output, "should not reach here") {
		t.Error("code after uncaught exception should not execute")
	}
}

// ============================================================================
// Finally Block Tests (Task 8.220)
// ============================================================================

// TestTryFinallyNoException tests try/finally when no exception occurs
func TestTryFinallyNoException(t *testing.T) {
	input := `
		var finallyExecuted: Boolean;
		finallyExecuted := false;

		try
			PrintLn('try block');
		finally
			finallyExecuted := true;
			PrintLn('finally block');
		end;

		PrintLn(finallyExecuted);
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
	expected := "try block\nfinally block\ntrue\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestTryFinallyWithException tests try/finally when exception occurs (uncaught)
func TestTryFinallyWithException(t *testing.T) {
	input := `
		var finallyExecuted: Boolean;
		finallyExecuted := false;

		try
			PrintLn('try block');
			raise Exception.Create('error');
			PrintLn('after raise');
		finally
			finallyExecuted := true;
			PrintLn('finally block');
		end;

		PrintLn('after try-finally');
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	err := interp.Eval(program)

	// Should have an error (exception propagates after finally)
	if err == nil {
		t.Fatal("expected exception to propagate after finally, got nil")
	}

	// Finally block should have executed
	output := buf.String()
	if !strings.Contains(output, "finally block") {
		t.Error("finally block should have executed even with exception")
	}

	// Code after raise should not execute
	if strings.Contains(output, "after raise") {
		t.Error("code after raise should not execute")
	}

	// Code after try-finally should not execute
	if strings.Contains(output, "after try-finally") {
		t.Error("code after try-finally should not execute when exception propagates")
	}
}

// TestTryExceptFinallyCombined tests try/except/finally all together
func TestTryExceptFinallyCombined(t *testing.T) {
	input := `
		var caught: Boolean;
		var finallyExecuted: Boolean;
		caught := false;
		finallyExecuted := false;

		try
			raise Exception.Create('error');
		except
			on E: Exception do begin
				caught := true;
				PrintLn('exception caught');
			end;
		finally
			finallyExecuted := true;
			PrintLn('finally executed');
		end;

		PrintLn(caught);
		PrintLn(finallyExecuted);
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
	expected := "exception caught\nfinally executed\ntrue\ntrue\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Exception Propagation Tests (Task 8.221)
// ============================================================================

// TestExceptionPropagatesAcrossFunctions tests exception propagation through function calls
func TestExceptionPropagatesAcrossFunctions(t *testing.T) {
	input := `
		function Level3: Integer;
		begin
			raise Exception.Create('error in level 3');
			Result := 0;
		end;

		function Level2: Integer;
		begin
			Result := Level3();
		end;

		function Level1: Integer;
		begin
			try
				Result := Level2();
			except
				on E: Exception do begin
					PrintLn('caught in level 1');
					Result := -1;
				end;
			end;
		end;

		var x: Integer;
		x := Level1();
		PrintLn(x);
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
	expected := "caught in level 1\n-1\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestNestedTryBlocks tests nested try/except blocks
func TestNestedTryBlocks(t *testing.T) {
	input := `
		var innerCaught: Boolean;
		var outerCaught: Boolean;
		innerCaught := false;
		outerCaught := false;

		try
			PrintLn('outer try');
			try
				PrintLn('inner try');
				raise EConvertError.Create('inner error');
			except
				on E: EConvertError do begin
					innerCaught := true;
					PrintLn('inner caught');
				end;
			end;
			PrintLn('after inner');
		except
			on E: Exception do begin
				outerCaught := true;
				PrintLn('outer caught');
			end;
		end;

		PrintLn(innerCaught);
		PrintLn(outerCaught);
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
	expected := "outer try\ninner try\ninner caught\nafter inner\ntrue\nfalse\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestNestedTryOuterCatches tests nested try where inner doesn't catch, outer does
func TestNestedTryOuterCatches(t *testing.T) {
	input := `
		var innerCaught: Boolean;
		var outerCaught: Boolean;
		innerCaught := false;
		outerCaught := false;

		try
			try
				raise EConvertError.Create('error');
			except
				on E: ERangeError do
					innerCaught := true;
			end;
		except
			on E: Exception do begin
				outerCaught := true;
				PrintLn('outer caught');
			end;
		end;

		PrintLn(innerCaught);
		PrintLn(outerCaught);
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
	expected := "outer caught\nfalse\ntrue\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Bare Raise Tests (Task 8.222)
// ============================================================================

// TestBareRaiseReThrows tests that bare raise re-throws the current exception
func TestBareRaiseReThrows(t *testing.T) {
	input := `
		var innerCaught: Boolean;
		var outerCaught: Boolean;
		innerCaught := false;
		outerCaught := false;

		try
			try
				raise Exception.Create('original error');
			except
				on E: Exception do begin
					innerCaught := true;
					PrintLn('inner caught');
					raise;  // Re-throw
				end;
			end;
		except
			on E: Exception do begin
				outerCaught := true;
				PrintLn('outer caught');
			end;
		end;

		PrintLn(innerCaught);
		PrintLn(outerCaught);
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
	expected := "inner caught\nouter caught\ntrue\ntrue\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Exception Type Hierarchy Tests (Task 8.223)
// ============================================================================

// TestExceptionCatchesAllTypes tests that catching Exception catches all exception types
func TestExceptionCatchesAllTypes(t *testing.T) {
	input := `
		var caught1: Boolean;
		var caught2: Boolean;
		var caught3: Boolean;
		caught1 := false;
		caught2 := false;
		caught3 := false;

		try
			raise ERangeError.Create('range error');
		except
			on E: Exception do
				caught1 := true;
		end;

		try
			raise EConvertError.Create('convert error');
		except
			on E: Exception do
				caught2 := true;
		end;

		try
			raise EDivByZero.Create('div by zero');
		except
			on E: Exception do
				caught3 := true;
		end;

		PrintLn(caught1);
		PrintLn(caught2);
		PrintLn(caught3);
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
	expected := "true\ntrue\ntrue\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestSpecificTypeDoesNotCatchOthers tests that specific handlers don't catch wrong types
func TestSpecificTypeDoesNotCatchOthers(t *testing.T) {
	input := `
		var caught: Boolean;
		caught := false;

		try
			raise EConvertError.Create('convert error');
		except
			on E: ERangeError do
				caught := true;
		end;

		PrintLn('should not reach here');
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	err := interp.Eval(program)

	// Should have an error (uncaught exception)
	if err == nil {
		t.Fatal("expected uncaught exception, got nil")
	}

	// Handler should not have executed
	output := buf.String()
	if strings.Contains(output, "should not reach here") {
		t.Error("code after uncaught exception should not execute")
	}
}

// TestHandlerOrderMatters tests that first matching handler wins
func TestHandlerOrderMatters(t *testing.T) {
	input := `
		var handler: String;
		handler := 'none';

		try
			raise ERangeError.Create('range error');
		except
			on E: Exception do
				handler := 'Exception';
			on E: ERangeError do
				handler := 'ERangeError';
		end;

		PrintLn(handler);
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
	// First handler (Exception) should match, not the more specific ERangeError
	expected := "Exception\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}
