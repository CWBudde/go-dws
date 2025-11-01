package interp

import (
	"bytes"
	"strings"
	"testing"
)

// ============================================================================
// Basic Lazy Parameter Tests
// ============================================================================

func TestLazyParameterBasicEvaluation(t *testing.T) {
	input := `
		function Test(lazy x: Integer): Integer;
		begin
			Result := x + 1;
		end;

		var a: Integer;
		begin
			a := 5;
			PrintLn(Test(a * 2));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "11\n" // (5 * 2) + 1 = 11
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestLazyParameterMultipleAccesses(t *testing.T) {
	input := `
		var evalCount: Integer;

		function GetValue: Integer;
		begin
			evalCount := evalCount + 1;
			Result := evalCount;
		end;

		function Triple(lazy x: Integer): Integer;
		begin
			// Access x three times - should evaluate GetValue() each time
			Result := x + x + x;
		end;

		begin
			evalCount := 0;
			PrintLn(Triple(GetValue()));
			// GetValue() is called 3 times: 1 + 2 + 3 = 6
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "6\n" // 1 + 2 + 3 = 6
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestLazyParameterWithVariableReference(t *testing.T) {
	input := `
		function Compute(lazy expr: Integer): Integer;
		begin
			Result := expr * 2;
		end;

		var value: Integer;
		begin
			value := 21;
			PrintLn(Compute(value));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Jensen's Device Tests
// ============================================================================

func TestLazyParameterJensensDevice(t *testing.T) {
	// NOTE: True Jensen's Device requires var parameter support (not yet implemented)
	// This test demonstrates the lazy evaluation pattern with a workaround
	input := `
		var globalI: Integer;

		// Simplified Jensen's Device using global variable
		function sum(lo, hi: Integer; lazy term: Float): Float;
		begin
			Result := 0.0;
			globalI := lo;
			while globalI <= hi do begin
				Result := Result + term;  // term is re-evaluated each iteration
				globalI := globalI + 1;
			end;
		end;

		var harmonic: Float;
		begin
			// Compute harmonic series: 1/1 + 1/2 + 1/3 + 1/4 + 1/5
			harmonic := sum(1, 5, 1.0 / globalI);
			PrintLn(harmonic);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	// Expected: 1/1 + 1/2 + 1/3 + 1/4 + 1/5 = 2.283333...
	output := out.String()
	if !strings.HasPrefix(output, "2.28") {
		t.Errorf("wrong output. expected to start with '2.28', got=%q", output)
	}
}

func TestLazyParameterJensensDeviceSquares(t *testing.T) {
	// NOTE: True Jensen's Device requires var parameter support (not yet implemented)
	// This test demonstrates the lazy evaluation pattern with a workaround
	input := `
		var globalI: Integer;

		function sum(lo, hi: Integer; lazy term: Integer): Integer;
		begin
			Result := 0;
			globalI := lo;
			while globalI <= hi do begin
				Result := Result + term;
				globalI := globalI + 1;
			end;
		end;

		var sumOfSquares: Integer;
		begin
			// Compute sum of squares: 1^2 + 2^2 + 3^2 + 4^2
			sumOfSquares := sum(1, 4, globalI * globalI);
			PrintLn(sumOfSquares);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "30\n" // 1 + 4 + 9 + 16 = 30
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Conditional Evaluation Tests
// ============================================================================

func TestLazyParameterConditionalEvaluation(t *testing.T) {
	input := `
		var evaluationCount: Integer;

		function ExpensiveComputation: Integer;
		begin
			evaluationCount := evaluationCount + 1;
			Result := 42;
		end;

		function IfThen(cond: Boolean; lazy trueVal: Integer; lazy falseVal: Integer): Integer;
		begin
			if cond then
				Result := trueVal
			else
				Result := falseVal;
		end;

		begin
			evaluationCount := 0;
			PrintLn(IfThen(true, ExpensiveComputation(), 99));
			// Only trueVal should be evaluated, so evaluationCount should be 1
			PrintLn(evaluationCount);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n1\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestLazyParameterConditionalEvaluationFalse(t *testing.T) {
	input := `
		var evaluationCount: Integer;

		function ExpensiveComputation: Integer;
		begin
			evaluationCount := evaluationCount + 1;
			Result := 42;
		end;

		function IfThen(cond: Boolean; lazy trueVal: Integer; lazy falseVal: Integer): Integer;
		begin
			if cond then
				Result := trueVal
			else
				Result := falseVal;
		end;

		begin
			evaluationCount := 0;
			PrintLn(IfThen(false, 99, ExpensiveComputation()));
			// Only falseVal should be evaluated, so evaluationCount should be 1
			PrintLn(evaluationCount);
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n1\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Mixed Parameter Tests
// ============================================================================

func TestLazyParameterMixedWithRegular(t *testing.T) {
	input := `
		function Process(name: String; lazy value: Integer; multiplier: Integer): Integer;
		begin
			PrintLn(name);
			Result := value * multiplier;
		end;

		var x: Integer;
		begin
			x := 7;
			PrintLn(Process('test', x + 3, 2));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "test\n20\n" // (7 + 3) * 2 = 20
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestLazyParameterMultipleLazy(t *testing.T) {
	input := `
		function Choose(cond: Boolean; lazy opt1: String; lazy opt2: String): String;
		begin
			if cond then
				Result := opt1
			else
				Result := opt2;
		end;

		begin
			PrintLn(Choose(true, 'first', 'second'));
			PrintLn(Choose(false, 'first', 'second'));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "first\nsecond\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Complex Expression Tests
// ============================================================================

func TestLazyParameterComplexExpression(t *testing.T) {
	input := `
		function Evaluate(lazy expr: Integer): Integer;
		begin
			Result := expr;
		end;

		var a: Integer;
		var b: Integer;
		begin
			a := 5;
			b := 3;
			PrintLn(Evaluate(a + b * 2 - 1));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "10\n" // 5 + 3*2 - 1 = 10
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Nested Scope Tests
// ============================================================================

func TestLazyParameterCapturesOuterVariable(t *testing.T) {
	input := `
		var outer: Integer;

		function UseLazy(lazy x: Integer): Integer;
		begin
			Result := x;
		end;

		begin
			outer := 42;
			PrintLn(UseLazy(outer + 8));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "50\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

func TestLazyParameterSeesVariableMutations(t *testing.T) {
	// NOTE: This test requires var parameter support (not yet implemented)
	// Modified to use global variable to demonstrate mutation visibility
	input := `
		var globalCounter: Integer;

		function AccessTwice(lazy value: Integer): Integer;
		begin
			// First access
			Result := value;
			// Increment global counter
			globalCounter := globalCounter + 1;
			// Second access - should see the incremented counter
			Result := Result + value;
		end;

		begin
			globalCounter := 5;
			// Passes globalCounter as lazy parameter "value"
			// First access: globalCounter = 5
			// Increment: globalCounter = 6
			// Second access: globalCounter = 6
			// Result: 5 + 6 = 11
			PrintLn(AccessTwice(globalCounter));
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "11\n" // 5 + 6 = 11
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}

// ============================================================================
// Lazy Logging Pattern Tests
// ============================================================================

func TestLazyParameterLogging(t *testing.T) {
	input := `
		var logLevel: Integer;
		var evaluations: Integer;

		function GetMessage: String;
		begin
			evaluations := evaluations + 1;
			Result := 'expensive log message';
		end;

		procedure Log(level: Integer; lazy msg: String);
		begin
			if level >= logLevel then
				PrintLn(msg);
		end;

		begin
			evaluations := 0;
			logLevel := 2;

			// This should NOT evaluate GetMessage() because level (1) < logLevel (2)
			Log(1, GetMessage());
			PrintLn(evaluations);  // Should be 0

			// This SHOULD evaluate GetMessage() because level (2) >= logLevel (2)
			Log(2, GetMessage());
			PrintLn(evaluations);  // Should be 1
		end.
	`

	var out bytes.Buffer
	interp := New(&out)
	result := interpret(interp, input)

	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "0\nexpensive log message\n1\n"
	if out.String() != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, out.String())
	}
}
