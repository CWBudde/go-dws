package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestPreconditionSuccess tests that preconditions pass when conditions are met
func TestPreconditionSuccess(t *testing.T) {
	input := `
	function SafeDivide(a, b: Float): Float;
	require
		b <> 0.0;
	begin
		Result := a / b;
	end;

	begin
		PrintLn(SafeDivide(10.0, 2.0));
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "5\n" // Float division result
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestPreconditionFailure tests that preconditions fail when conditions are not met
func TestPreconditionFailure(t *testing.T) {
	input := `
	function SafeDivide(a, b: Float): Float;
	require
		b <> 0.0 : 'divisor cannot be zero';
	begin
		Result := a / b;
	end;

	begin
		SafeDivide(10.0, 0.0);
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if !isError(result) {
		t.Fatalf("Expected error for precondition failure, got %s", result.String())
	}

	errMsg := result.String()
	if !strings.Contains(errMsg, "Pre-condition failed") {
		t.Errorf("Expected 'Pre-condition failed' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "SafeDivide") {
		t.Errorf("Expected function name 'SafeDivide' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "divisor cannot be zero") {
		t.Errorf("Expected custom message in error, got: %s", errMsg)
	}
}

// TestPostconditionSuccess tests that postconditions pass when conditions are met
func TestPostconditionSuccess(t *testing.T) {
	input := `
	function AbsoluteValue(x: Integer): Integer;
	begin
		if x < 0 then
			Result := -x
		else
			Result := x;
	end;
	ensure
		Result >= 0 : 'absolute value must be non-negative';

	begin
		PrintLn(AbsoluteValue(-5));
		PrintLn(AbsoluteValue(5));
		PrintLn(AbsoluteValue(0));
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "5\n5\n0\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestPostconditionFailure tests that postconditions fail when conditions are not met
func TestPostconditionFailure(t *testing.T) {
	input := `
	function BrokenAbs(x: Integer): Integer;
	begin
		Result := x;  // Wrong implementation - doesn't actually take absolute value
	end;
	ensure
		Result >= 0 : 'absolute value must be non-negative';

	begin
		BrokenAbs(-5);
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if !isError(result) {
		t.Fatalf("Expected error for postcondition failure, got %s", result.String())
	}

	errMsg := result.String()
	if !strings.Contains(errMsg, "Post-condition failed") {
		t.Errorf("Expected 'Post-condition failed' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "BrokenAbs") {
		t.Errorf("Expected function name 'BrokenAbs' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "absolute value must be non-negative") {
		t.Errorf("Expected custom message in error, got: %s", errMsg)
	}
}

// TestOldExpressionInPostcondition tests the 'old' keyword in postconditions
func TestOldExpressionInPostcondition(t *testing.T) {
	input := `
	function Increment(x: Integer): Integer;
	begin
		Result := x + 1;
	end;
	ensure
		Result = old x + 1 : 'result must be one more than input';

	begin
		PrintLn(Increment(5));
		PrintLn(Increment(0));
		PrintLn(Increment(-1));
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "6\n1\n0\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestOldExpressionFailure tests that old expressions correctly detect violations
func TestOldExpressionFailure(t *testing.T) {
	input := `
	function BrokenIncrement(x: Integer): Integer;
	begin
		Result := x + 2;  // Wrong - adds 2 instead of 1
	end;
	ensure
		Result = old x + 1 : 'result must be one more than input';

	begin
		BrokenIncrement(5);
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if !isError(result) {
		t.Fatalf("Expected error for postcondition failure with old, got %s", result.String())
	}

	errMsg := result.String()
	if !strings.Contains(errMsg, "Post-condition failed") {
		t.Errorf("Expected 'Post-condition failed' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "result must be one more than input") {
		t.Errorf("Expected custom message in error, got: %s", errMsg)
	}
}

// TestMultipleConditions tests functions with multiple pre/postconditions
func TestMultipleConditions(t *testing.T) {
	input := `
	function Clamp(value, min, max: Integer): Integer;
	require
		min <= max : 'min must not exceed max';
	begin
		if value < min then
			Result := min
		else if value > max then
			Result := max
		else
			Result := value;
	end;
	ensure
		Result >= min : 'result must be >= min';
		Result <= max : 'result must be <= max';

	begin
		PrintLn(Clamp(5, 0, 10));   // 5
		PrintLn(Clamp(-5, 0, 10));  // 0
		PrintLn(Clamp(15, 0, 10));  // 10
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "5\n0\n10\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestMultipleConditionsFailure tests that the first failing condition is reported
func TestMultipleConditionsFailure(t *testing.T) {
	input := `
	function Clamp(value, min, max: Integer): Integer;
	require
		min <= max : 'min must not exceed max';
	begin
		Result := value;  // Wrong - doesn't clamp
	end;
	ensure
		Result >= min : 'result must be >= min';
		Result <= max : 'result must be <= max';

	begin
		Clamp(15, 0, 10);  // Should fail second postcondition
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if !isError(result) {
		t.Fatalf("Expected error for postcondition failure, got %s", result.String())
	}

	errMsg := result.String()
	if !strings.Contains(errMsg, "Post-condition failed") {
		t.Errorf("Expected 'Post-condition failed' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "result must be <= max") {
		t.Errorf("Expected second postcondition message in error, got: %s", errMsg)
	}
}

// TestContractsWithNestedCalls tests that old values are properly scoped in nested function calls
func TestContractsWithNestedCalls(t *testing.T) {
	input := `
	function Double(x: Integer): Integer;
	begin
		Result := x * 2;
	end;
	ensure
		Result = old x * 2;

	function AddDoubled(a, b: Integer): Integer;
	begin
		Result := Double(a) + Double(b);
	end;
	ensure
		Result = (old a * 2) + (old b * 2);

	begin
		PrintLn(AddDoubled(3, 4));  // Should print 14
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "14\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestContractWithNoMessage tests contracts without custom messages
func TestContractWithNoMessage(t *testing.T) {
	input := `
	function Positive(x: Integer): Integer;
	require
		x > 0;
	begin
		Result := x;
	end;

	begin
		Positive(-5);
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if !isError(result) {
		t.Fatalf("Expected error for precondition failure, got %s", result.String())
	}

	errMsg := result.String()
	if !strings.Contains(errMsg, "Pre-condition failed") {
		t.Errorf("Expected 'Pre-condition failed' in error, got: %s", errMsg)
	}
	// When no message is provided, should show the condition expression
	if !strings.Contains(errMsg, "x > 0") {
		t.Errorf("Expected condition expression in error, got: %s", errMsg)
	}
}

// TestContractDivisionExample tests the division example from testdata
func TestContractDivisionExample(t *testing.T) {
	input := `
	function SafeDivide(a, b: Float): Float;
	require
		b <> 0.0 : 'divisor cannot be zero';
	begin
		Result := a / b;
	end;
	ensure
		Result * b = a : 'division result verification failed';

	begin
		PrintLn(SafeDivide(10.0, 2.0));
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "5\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}

// TestContractWithProcedure tests that procedures (no return type) can have contracts
func TestContractWithProcedure(t *testing.T) {
	input := `
	var globalValue: Integer;

	procedure SetPositive(x: Integer);
	require
		x > 0 : 'value must be positive';
	begin
		globalValue := x;
	end;
	ensure
		globalValue = old x : 'global value must be set to parameter';

	begin
		SetPositive(42);
		PrintLn(globalValue);
	end.
	`

	output := &bytes.Buffer{}
	interp := New(output)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("Interpreter error: %s", result.String())
	}

	expected := "42\n"
	if output.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, output.String())
	}
}
