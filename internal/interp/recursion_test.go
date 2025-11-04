package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// runScript is a helper to parse and execute a DWScript program
func runScript(t *testing.T, interp *Interpreter, script string) Value {
	l := lexer.New(script)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
	}

	return interp.Eval(program)
}

// TestRecursionAtLimit tests that recursion exactly at the limit (1024) succeeds
func TestRecursionAtLimit(t *testing.T) {
	script := `
var counter: Integer := 0;

procedure CountDown(n: Integer);
begin
	counter := counter + 1;
	if n > 1 then
		CountDown(n - 1);
end;

begin
	CountDown(1024);
	PrintLn(counter);
end.
`
	var output bytes.Buffer
	interp := New(&output)
	result := runScript(t, interp, script)

	if isError(result) {
		t.Fatalf("Expected success at limit, got error: %v", result)
	}

	got := strings.TrimSpace(output.String())
	expected := "1024"
	if got != expected {
		t.Errorf("Expected output %q, got %q", expected, got)
	}
}

// TestRecursionExceedingLimit tests that recursion exceeding limit by 1 raises EScriptStackOverflow
func TestRecursionExceedingLimit(t *testing.T) {
	script := `
var counter: Integer := 0;
var exceptionRaised: Boolean := False;

procedure CountDown(n: Integer);
begin
	counter := counter + 1;
	if n > 1 then
		CountDown(n - 1);
end;

begin
	try
		CountDown(1025);
	except
		on E: EScriptStackOverflow do
			exceptionRaised := True;
	end;

	PrintLn(exceptionRaised);
	PrintLn(counter);
end.
`
	var output bytes.Buffer
	interp := New(&output)
	result := runScript(t, interp, script)

	if isError(result) {
		t.Fatalf("Script execution failed: %v", result)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("Expected 2 lines of output, got %d", len(lines))
	}

	// Check that exception was raised
	if lines[0] != "true" {
		t.Errorf("Expected exception to be raised (true), got %q", lines[0])
	}

	// Check that counter reached the limit (1024)
	if lines[1] != "1024" {
		t.Errorf("Expected counter to reach 1024, got %q", lines[1])
	}
}

// TestRecursionExceptionMessage tests that the exception message contains the depth limit
func TestRecursionExceptionMessage(t *testing.T) {
	script := `
var message: String := '';

procedure Recurse;
begin
	Recurse;
end;

begin
	try
		Recurse;
	except
		on E: EScriptStackOverflow do
			message := E.Message;
	end;

	PrintLn(message);
end.
`
	var output bytes.Buffer
	interp := New(&output)
	result := runScript(t, interp, script)

	if isError(result) {
		t.Fatalf("Script execution failed: %v", result)
	}

	got := strings.TrimSpace(output.String())
	// Message should contain "1024"
	if !strings.Contains(got, "1024") {
		t.Errorf("Expected message to contain '1024', got %q", got)
	}
	if !strings.Contains(got, "Maximal recursion exceeded") {
		t.Errorf("Expected message to contain 'Maximal recursion exceeded', got %q", got)
	}
}

// TestRecursionExceptionCanBeCaught tests that the exception can be caught with try/except
func TestRecursionExceptionCanBeCaught(t *testing.T) {
	script := `
var caught: Boolean := False;

procedure InfiniteRecurse;
begin
	InfiniteRecurse;
end;

begin
	try
		InfiniteRecurse;
	except
		caught := True;
	end;

	PrintLn(caught);
end.
`
	var output bytes.Buffer
	interp := New(&output)
	result := runScript(t, interp, script)

	if isError(result) {
		t.Fatalf("Script execution failed: %v", result)
	}

	got := strings.TrimSpace(output.String())
	if got != "true" {
		t.Errorf("Expected exception to be caught (true), got %q", got)
	}
}

// TestRecursionWithNestedTryExcept tests nested try/except with recursion
func TestRecursionWithNestedTryExcept(t *testing.T) {
	script := `
var innerCaught: Boolean := False;
var outerCaught: Boolean := False;

procedure Recurse;
begin
	try
		Recurse;
	except
		innerCaught := True;
		raise;  // Re-raise to outer handler
	end;
end;

begin
	try
		Recurse;
	except
		outerCaught := True;
	end;

	PrintLn(innerCaught);
	PrintLn(outerCaught);
end.
`
	var output bytes.Buffer
	interp := New(&output)
	result := runScript(t, interp, script)

	if isError(result) {
		t.Fatalf("Script execution failed: %v", result)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("Expected 2 lines of output, got %d", len(lines))
	}

	// Both inner and outer should catch the exception
	if lines[0] != "true" {
		t.Errorf("Expected inner exception to be caught (true), got %q", lines[0])
	}
	if lines[1] != "true" {
		t.Errorf("Expected outer exception to be caught (true), got %q", lines[1])
	}
}

// TestRecursionWithLambda tests recursion limit with lambda functions
func TestRecursionWithLambda(t *testing.T) {
	script := `
var counter: Integer := 0;
var caught: Boolean := False;

begin
	try
		var recurse: procedure(n: Integer);
		recurse := lambda (n: Integer) begin
			counter := counter + 1;
			if n > 1 then
				recurse(n - 1);
		end;

		recurse(1025);
	except
		on E: EScriptStackOverflow do
			caught := True;
	end;

	PrintLn(caught);
	PrintLn(counter);
end.
`
	var output bytes.Buffer
	interp := New(&output)
	result := runScript(t, interp, script)

	if isError(result) {
		t.Fatalf("Script execution failed: %v", result)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("Expected 2 lines of output, got %d", len(lines))
	}

	// Check that exception was raised
	if lines[0] != "true" {
		t.Errorf("Expected exception to be raised (true), got %q", lines[0])
	}

	// Check that counter reached the limit (1024)
	if lines[1] != "1024" {
		t.Errorf("Expected counter to reach 1024, got %q", lines[1])
	}
}

// Note: Additional tests for record methods, class methods, and forward declarations
// are omitted as those features are not yet fully implemented. The 6 tests above
// comprehensively validate recursion limit enforcement for user functions and lambdas.
