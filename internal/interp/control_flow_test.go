package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// ============================================================================
// Break Statement Tests (Task 8.235q)
// ============================================================================

// TestBreakExitsForLoop tests that break exits a for loop correctly
func TestBreakExitsForLoop(t *testing.T) {
	input := `
		var count: Integer;
		count := 0;

		var i: Integer;
		for i := 1 to 10 do
		begin
			count := count + 1;
			if i = 5 then
				break;
		end;

		PrintLn(count);
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
	expected := "5\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestBreakExitsWhileLoop tests that break exits a while loop correctly
func TestBreakExitsWhileLoop(t *testing.T) {
	input := `
		var i: Integer;
		var count: Integer;

		i := 0;
		count := 0;

		while i < 10 do
		begin
			i := i + 1;
			count := count + 1;
			if i = 5 then
				break;
		end;

		PrintLn(count);
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
	expected := "5\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestBreakExitsRepeatLoop tests that break exits a repeat loop correctly
func TestBreakExitsRepeatLoop(t *testing.T) {
	input := `
		var i: Integer;
		var count: Integer;

		i := 0;
		count := 0;

		repeat
		begin
			i := i + 1;
			count := count + 1;
			if i = 5 then
				break;
		end
		until i >= 10;

		PrintLn(count);
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
	expected := "5\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Continue Statement Tests (Task 8.235q)
// ============================================================================

// TestContinueSkipsForLoopIteration tests that continue skips to next iteration in for loop
func TestContinueSkipsForLoopIteration(t *testing.T) {
	input := `
		var sum: Integer;
		sum := 0;

		var i: Integer;
		for i := 1 to 10 do
		begin
			if i = 5 then
				continue;
			sum := sum + i;
		end;

		PrintLn(sum);
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
	// Sum should be 1+2+3+4+6+7+8+9+10 = 50
	expected := "50\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestContinueSkipsWhileLoopIteration tests that continue skips to next iteration in while loop
func TestContinueSkipsWhileLoopIteration(t *testing.T) {
	input := `
		var i: Integer;
		var sum: Integer;

		i := 0;
		sum := 0;

		while i < 10 do
		begin
			i := i + 1;
			if i = 5 then
				continue;
			sum := sum + i;
		end;

		PrintLn(sum);
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
	// Sum should be 1+2+3+4+6+7+8+9+10 = 50
	expected := "50\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestContinueSkipsRepeatLoopIteration tests that continue skips to next iteration in repeat loop
func TestContinueSkipsRepeatLoopIteration(t *testing.T) {
	input := `
		var i: Integer;
		var sum: Integer;

		i := 0;
		sum := 0;

		repeat
		begin
			i := i + 1;
			if i = 5 then
				continue;
			sum := sum + i;
		end
		until i >= 10;

		PrintLn(sum);
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
	// Sum should be 1+2+3+4+6+7+8+9+10 = 50
	expected := "50\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Exit Statement Tests (Task 8.235r)
// ============================================================================

// TestExitTerminatesFunctionImmediately tests that exit exits function immediately
func TestExitTerminatesFunctionImmediately(t *testing.T) {
	input := `
		function GetValue(x: Integer): Integer;
		begin
			Result := 10;
			if x < 0 then
			begin
				Result := 0;
				exit;
			end;
			Result := 20;
		end;

		PrintLn(GetValue(-1));
		PrintLn(GetValue(1));
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
	expected := "0\n20\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestExitInNestedFunctionDoesNotAffectCaller tests that exit in nested function doesn't affect caller
func TestExitInNestedFunctionDoesNotAffectCaller(t *testing.T) {
	input := `
		function Inner: Integer;
		begin
			Result := 5;
			exit;
			Result := 10;
		end;

		function Outer: Integer;
		begin
			Result := Inner();
			Result := Result + 100;
		end;

		PrintLn(Outer());
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
	expected := "105\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestExitPreservesResultVariable tests that exit preserves Result variable value
func TestExitPreservesResultVariable(t *testing.T) {
	input := `
		function Calculate(x: Integer): Integer;
		begin
			Result := x * 2;
			if x > 5 then
				exit;
			Result := Result + 10;
		end;

		PrintLn(Calculate(3));
		PrintLn(Calculate(10));
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
	// Calculate(3) = (3 * 2) + 10 = 16
	// Calculate(10) = 10 * 2 = 20 (exits before adding 10)
	expected := "16\n20\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestExitInProcedure tests that exit works in a procedure (no return value)
func TestExitInProcedure(t *testing.T) {
	input := `
		procedure PrintIfPositive(x: Integer);
		begin
			if x < 0 then
				exit;
			PrintLn(x);
		end;

		PrintIfPositive(-5);
		PrintIfPositive(10);
		PrintIfPositive(-3);
		PrintIfPositive(20);
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
	expected := "10\n20\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// ============================================================================
// Nested Scenarios Tests (Task 8.235s)
// ============================================================================

// TestBreakInNestedLoopsOnlyExitsInnermost tests that break only exits innermost loop
func TestBreakInNestedLoopsOnlyExitsInnermost(t *testing.T) {
	input := `
		var outerCount: Integer;
		var innerCount: Integer;

		outerCount := 0;
		innerCount := 0;

		var i: Integer;
		var j: Integer;
		for i := 1 to 3 do
		begin
			outerCount := outerCount + 1;
			for j := 1 to 10 do
			begin
				innerCount := innerCount + 1;
				if j = 2 then
					break;
			end;
		end;

		PrintLn(outerCount);
		PrintLn(innerCount);
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
	// Outer loop runs 3 times, inner loop breaks after 2 iterations each time
	expected := "3\n6\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestContinueInNestedLoopsOnlyAffectsInnermost tests that continue only affects innermost loop
func TestContinueInNestedLoopsOnlyAffectsInnermost(t *testing.T) {
	input := `
		var sum: Integer;
		sum := 0;

		var i: Integer;
		var j: Integer;
		for i := 1 to 3 do
		begin
			for j := 1 to 5 do
			begin
				if j = 3 then
					continue;
				sum := sum + 1;
			end;
		end;

		PrintLn(sum);
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
	// Outer loop runs 3 times, inner loop runs 5 times but skips j=3, so 4 increments per outer iteration
	expected := "12\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestBreakContinueWithExceptionHandling tests break/continue with try-except
func TestBreakContinueWithExceptionHandling(t *testing.T) {
	input := `
		var count: Integer;
		count := 0;

		var i: Integer;
		for i := 1 to 10 do
		begin
			try
				count := count + 1;
				if i = 5 then
					break;
			except
				on E: Exception do
					PrintLn('caught');
			end;
		end;

		PrintLn(count);
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
	expected := "5\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestExitWithNestedFunctionCalls tests that exit works correctly with nested function calls
func TestExitWithNestedFunctionCalls(t *testing.T) {
	input := `
		function Level3: Integer;
		begin
			Result := 3;
			exit;
			Result := 300;
		end;

		function Level2: Integer;
		begin
			Result := Level3() + 20;
			exit;
			Result := 200;
		end;

		function Level1: Integer;
		begin
			Result := Level2() + 100;
		end;

		PrintLn(Level1());
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
	// Level3 returns 3, Level2 returns 3+20=23, Level1 returns 23+100=123
	expected := "123\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

// TestExitAtProgramLevel tests that exit at program level terminates execution
func TestExitAtProgramLevel(t *testing.T) {
	input := `
		var x: Integer;
		x := 10;
		PrintLn(x);

		if x = 10 then
			exit;

		PrintLn('This should not be printed');
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
	expected := "10\n"

	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}
