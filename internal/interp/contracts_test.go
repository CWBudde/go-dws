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

// TestPreconditionArrayLength verifies helper member access in preconditions
func TestPreconditionArrayLength(t *testing.T) {
	input := `
	function First(arr: array of Integer): Integer;
	require
		arr.Length > 0;
	begin
		Result := arr[0];
	end;

	var data: array of Integer;
	begin
		SetLength(data, 1);
		data[0] := 42;
		PrintLn(First(data));
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

	if got := output.String(); got != "42\n" {
		t.Fatalf("expected output 42, got %q", got)
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

// ============================================================================
// Method contracts: naming and inheritance (PLAN.md §3.2.3)
// ============================================================================

// TestContract_InlineMethodNameIsClassQualified verifies that a method whose
// body is written inline in the class declaration reports its contract
// failures class-qualified, exactly as an out-of-line implementation does.
// Inline methods carry no ClassName on the declaration, which used to drop the
// prefix (SimpleScripts/method_condition).
func TestContract_InlineMethodNameIsClassQualified(t *testing.T) {
	runScriptTestWithSemantic(t, `
		type TTest = class
			Field : Integer;
			procedure Bump;
			require
				Field <= 0;
			begin
				Field += 1;
			end;
		end;

		var t := new TTest;
		t.Bump;
		try
			t.Bump;
		except
			on E: Exception do PrintLn(E.Message);
		end;
	`, "Pre-condition failed in TTest.Bump [line: 6, column: 5], Field <= 0")
}

// TestContract_FreeFunctionCalledFromMethodKeepsBareName verifies that a free
// function called from inside a method body is not mistaken for a method of the
// enclosing class. The callee still sees the caller's class binding, so the
// chain resolver must confirm that the class actually declares the routine.
func TestContract_FreeFunctionCalledFromMethodKeepsBareName(t *testing.T) {
	runScriptTestWithSemantic(t, `
		procedure RequirePositive(i : Integer);
		require
			i > 0;
		begin
			PrintLn(i);
		end;

		type TTest = class
			procedure Run(i : Integer);
			begin
				RequirePositive(i);
			end;
		end;

		var t := new TTest;
		try
			t.Run(-1);
		except
			on E: Exception do PrintLn(E.Message);
		end;
	`, "Pre-condition failed in RequirePositive [line: 4, column: 4], i > 0")
}

// TestContract_OverrideInheritsPrecondition verifies that an override with no
// `require` of its own runs the ancestor's, and that the failure names the
// class that declares the condition rather than the receiver's dynamic class
// (SimpleScripts/method_contracts).
func TestContract_OverrideInheritsPrecondition(t *testing.T) {
	runScriptTestWithSemantic(t, `
		type TBase = class
			procedure Check(i : Integer); virtual;
			require
				i > 0;
			begin
				PrintLn('base ' + IntToStr(i));
			end;
		end;

		type TChild = class (TBase)
			procedure Check(i : Integer); override;
			begin
				PrintLn('child ' + IntToStr(i));
			end;
		end;

		var c := TChild.Create;
		c.Check(1);
		try
			c.Check(-1);
		except
			on E: Exception do PrintLn(E.Message);
		end;
	`, "child 1\nPre-condition failed in TBase.Check [line: 5, column: 5], i > 0")
}

// TestContract_InheritedPreconditionThroughBaseReference verifies that the
// inherited precondition also runs when the call is dispatched virtually
// through a base-typed reference.
func TestContract_InheritedPreconditionThroughBaseReference(t *testing.T) {
	runScriptTestWithSemantic(t, `
		type TBase = class
			procedure Check(i : Integer); virtual;
			require
				i > 0;
			begin
				PrintLn('base ' + IntToStr(i));
			end;
		end;

		type TChild = class (TBase)
			procedure Check(i : Integer); override;
			begin
				PrintLn('child ' + IntToStr(i));
			end;
		end;

		var b : TBase := TChild.Create;
		try
			b.Check(0);
		except
			on E: Exception do PrintLn(E.Message);
		end;
	`, "Pre-condition failed in TBase.Check [line: 5, column: 5], i > 0")
}

// TestContract_InheritedPreconditionWithRenamedParameter verifies that an
// ancestor condition is evaluated against the call's arguments by position,
// even when the override gives its parameters different names.
func TestContract_InheritedPreconditionWithRenamedParameter(t *testing.T) {
	runScriptTestWithSemantic(t, `
		type TBase = class
			procedure Check(value : Integer); virtual;
			require
				value > 0;
			begin
				PrintLn('base ' + IntToStr(value));
			end;
		end;

		type TChild = class (TBase)
			procedure Check(amount : Integer); override;
			begin
				PrintLn('child ' + IntToStr(amount));
			end;
		end;

		var c := TChild.Create;
		try
			c.Check(-3);
		except
			on E: Exception do PrintLn(E.Message);
		end;
	`, "Pre-condition failed in TBase.Check [line: 5, column: 5], value > 0")
}
