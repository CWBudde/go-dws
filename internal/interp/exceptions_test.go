package interp

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Basic Exception Handling Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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

// TestEHostCreateSetsFields ensures the EHost constructor assigns both Message and ExceptionClass.
func TestEHostCreateSetsFields(t *testing.T) {
	input := `
		var host: EHost;
		host := EHost.Create('my.Err', 'boom');

		PrintLn(host.ExceptionClass);
		PrintLn(host.Message);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()
	expected := "my.Err\nboom\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Multiple Exception Handlers Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
// Bare Except Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
// Finally Block Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
// Exception Propagation Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
// Bare Raise Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
// Exception Type Hierarchy Tests
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
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

// ============================================================================
// Additional Missing Tests
// ============================================================================

// TestTryFinallyWithReturn tests that finally executes even when returning from try block
// Note: Using implicit return (Result assignment) since 'exit' keyword is not yet implemented
func TestTryFinallyWithReturn(t *testing.T) {
	input := `
		var finallyExecuted: Boolean;
		finallyExecuted := false;

		function TestFunction(): Integer;
		begin
			try
				PrintLn('in try');
				Result := 42;
			finally
				finallyExecuted := true;
				PrintLn('finally executed');
			end;
		end;

		var result: Integer;
		result := TestFunction();
		PrintLn(finallyExecuted);
		PrintLn(result);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Finally block should have executed
	if !strings.Contains(output, "finally executed") {
		t.Error("finally block should execute even on return")
	}

	// Should have set finallyExecuted to true
	if !strings.Contains(output, "true") {
		t.Error("finallyExecuted should be true")
	}

	// Should have returned 42
	if !strings.Contains(output, "42") {
		t.Error("should have returned 42")
	}
}

// TestRaiseCustomException tests raising a custom exception class
// Note: This test verifies that custom exception classes properly inherit from Exception
func TestRaiseCustomException(t *testing.T) {
	input := `
		type ECustomError = class(Exception)
		end;

		var caught: Boolean;
		var exceptionType: String;
		caught := false;

		try
			raise ECustomError.Create('custom error message');
		except
			on E: ECustomError do begin
				caught := true;
				exceptionType := 'ECustomError';
			end;
			on E: Exception do begin
				caught := true;
				exceptionType := 'Exception';
			end;
		end;

		PrintLn(caught);
		PrintLn(exceptionType);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Should have caught the exception
	if !strings.Contains(output, "true") {
		t.Error("custom exception should have been caught")
	}

	// Should have caught with the specific handler (not the base Exception handler)
	if !strings.Contains(output, "ECustomError") {
		t.Errorf("expected to catch with ECustomError handler, got output: %s", output)
	}
}

// TestBareRaiseOutsideHandler tests that bare raise outside a handler causes runtime error
func TestBareRaiseOutsideHandler(t *testing.T) {
	input := `
		raise;  // Bare raise with no active exception
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)

	// This should panic with "bare raise with no active exception"
	defer func() {
		if r := recover(); r != nil {
			// Check that the panic message is what we expect
			errMsg := r.(string)
			if !strings.Contains(errMsg, "bare raise") {
				t.Errorf("expected panic about bare raise, got: %v", r)
			}
		} else {
			t.Error("expected panic for bare raise outside handler, but didn't panic")
		}
	}()

	interp.Eval(program)
}

// ============================================================================
// Ported DWScript Exception Tests
// ============================================================================

// TestPortedBasicExceptions tests basic exception handling with specific types
// Ported from: reference/dwscript-original/Test/SimpleScripts/exceptions.pas
func TestPortedBasicExceptions(t *testing.T) {
	source, err := os.ReadFile("../../testdata/exceptions/basic_exceptions.dws")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Expected output based on the script
	expectedOutputs := []string{
		"MyException: exception message",
		"OtherException: exception message",
		"Else",
		"MyException: exception message 2",
		"Finally",
		"Except",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}

// TestPortedReRaise tests re-raising exceptions
// Ported from: reference/dwscript-original/Test/SimpleScripts/exceptions2.pas
func TestPortedReRaise(t *testing.T) {
	source, err := os.ReadFile("../../testdata/exceptions/re_raise.dws")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Expected output based on the script
	expectedOutputs := []string{
		"Caught once",
		"Caught again",
		"Ended",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}

// TestPortedTryExceptFinally tests combined try-except-finally blocks
// Ported from: reference/dwscript-original/Test/SimpleScripts/try_except_finally.pas
func TestPortedTryExceptFinally(t *testing.T) {
	source, err := os.ReadFile("../../testdata/exceptions/try_except_finally.dws")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Expected output based on the script
	expectedOutputs := []string{
		"Hello World",
		"Exception World",
		"Bye World",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}

// TestPortedExceptObject tests the ExceptObject built-in variable
// Ported from: reference/dwscript-original/Test/SimpleScripts/exceptobj.pas
func TestPortedExceptObject(t *testing.T) {
	source, err := os.ReadFile("../../testdata/exceptions/except_object.dws")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Expected output based on the script
	expectedOutputs := []string{
		"Exception: hello",
		"EMyExcept: world",
		"EMyExcept: world",
		"Exception: hello",
		"EMyExcept: hello world",
		"EMyExcept: hello world",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}

// TestPortedNestedCalls tests exceptions propagating through nested function calls
// This test expects an unhandled exception
// Ported from: reference/dwscript-original/Test/SimpleScripts/exception_nested_call.pas
func TestPortedNestedCalls(t *testing.T) {
	source, err := os.ReadFile("../../testdata/exceptions/nested_calls.dws")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	output := buf.String()

	// Expected partial output before the unhandled exception
	if !strings.Contains(output, "Exception1 caught") {
		t.Errorf("Expected output to contain 'Exception1 caught', but it didn't.\nActual output:\n%s", output)
	}

	// Should have an unhandled exception (ERROR result)
	if result == nil || result.Type() != "ERROR" {
		t.Errorf("Expected an ERROR result for unhandled exception, got: %v", result)
	}

	// The error message should contain "Error message 2"
	if result != nil && result.Type() == "ERROR" {
		errorMsg := result.String()
		if !strings.Contains(errorMsg, "Error message 2") {
			t.Errorf("Expected error to contain 'Error message 2', got: %s", errorMsg)
		}
	}
}

// TestExceptionPropagationFromFunction tests that exceptions propagate from function calls
func TestExceptionPropagationFromFunction(t *testing.T) {
	input := `
		procedure Trigger;
		begin
			PrintLn('Before raise');
			raise Exception.Create('from function');
			PrintLn('After raise'); // Should not print
		end;

		PrintLn('Defined Trigger');

		var result: String;
		result := '';

		PrintLn('Before try');
		try
			PrintLn('In try');
			PrintLn('About to call Trigger');
			Trigger;
			PrintLn('After Trigger'); // Should not print
		except
			PrintLn('In except');
			result := 'Caught';
		end;
		PrintLn('After try');

		PrintLn(result);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Just check what we got
	t.Logf("Actual output:\n%s", output)

	// Check that except block executed
	if !strings.Contains(output, "In except") {
		t.Error("Expected except block to execute ('In except' not found in output)")
	}
	if !strings.Contains(output, "Caught") {
		t.Error("Expected 'Caught' in output")
	}
}

// TestBareExceptBlock tests that bare except blocks (with no handlers) catch all exceptions
func TestBareExceptBlock(t *testing.T) {
	input := `
		var result: String;
		result := '';

		try
			raise Exception.Create('test error');
		except
			result := 'Caught';
		end;

		PrintLn(result);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()
	expected := "Caught\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestElseClauseWhenNoHandlerMatches tests that else clause executes when exception raised but no handler matches
// This is a TDD test for fixing the else clause bug
func TestElseClauseWhenNoHandlerMatches(t *testing.T) {
	input := `
		var result: String;
		result := '';

		try
			raise Exception.Create('test error');
		except
			on E: ERangeError do
				result := 'ERangeError';
			on E: EConvertError do
				result := 'EConvertError';
		else
			result := 'Else';
		end;

		PrintLn(result);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()
	expected := "Else\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestPortedBreakInExcept tests break statements inside exception handlers
// Ported from: reference/dwscript-original/Test/SimpleScripts/break_in_except_block.pas
func TestPortedBreakInExcept(t *testing.T) {
	source, err := os.ReadFile("../../testdata/exceptions/break_in_except.dws")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.Eval(program)

	output := buf.String()

	// Expected output based on the script
	expectedOutputs := []string{
		"hello",
		"world",
		"duh",
		"finally",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}
