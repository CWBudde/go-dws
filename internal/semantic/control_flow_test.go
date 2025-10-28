package semantic

import (
	"strings"
	"testing"
)

// ============================================================================
// Break Statement Semantic Analysis Tests (Task 8.235p)
// ============================================================================

// TestBreakOutsideLoop tests that break outside a loop produces a semantic error
func TestBreakOutsideLoop(t *testing.T) {
	input := `
		var x: Integer;
		x := 10;
		break;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for break outside loop, got nil")
	}

	if !strings.Contains(err.Error(), "break statement not allowed outside loop") {
		t.Errorf("Expected error about break outside loop, got: %v", err)
	}
}

// TestBreakInFinallyBlock tests that break in finally block produces a semantic error
func TestBreakInFinallyBlock(t *testing.T) {
	input := `
		var i: Integer;
		for i := 1 to 5 do
		begin
			try
				PrintLn(i);
			finally
				break;
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for break in finally block, got nil")
	}

	if !strings.Contains(err.Error(), "break statement not allowed in finally block") {
		t.Errorf("Expected error about break in finally block, got: %v", err)
	}
}

// TestBreakInForLoop tests that break in a for loop is valid
func TestBreakInForLoop(t *testing.T) {
	input := `
		var i: Integer;
		for i := 1 to 10 do
		begin
			if i = 5 then
				break;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for break in for loop, got: %v", err)
	}
}

// TestBreakInWhileLoop tests that break in a while loop is valid
func TestBreakInWhileLoop(t *testing.T) {
	input := `
		var i: Integer;
		i := 0;
		while i < 10 do
		begin
			i := i + 1;
			if i = 5 then
				break;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for break in while loop, got: %v", err)
	}
}

// TestBreakInRepeatLoop tests that break in a repeat loop is valid
func TestBreakInRepeatLoop(t *testing.T) {
	input := `
		var i: Integer;
		i := 0;
		repeat
		begin
			i := i + 1;
			if i = 5 then
				break;
		end
		until i >= 10;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for break in repeat loop, got: %v", err)
	}
}

// TestBreakInNestedLoops tests that break in nested loops is valid
func TestBreakInNestedLoops(t *testing.T) {
	input := `
		var i: Integer;
		var j: Integer;
		for i := 1 to 10 do
		begin
			for j := 1 to 10 do
			begin
				if j = 5 then
					break;
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for break in nested loops, got: %v", err)
	}
}

// ============================================================================
// Continue Statement Semantic Analysis Tests (Task 8.235p)
// ============================================================================

// TestContinueOutsideLoop tests that continue outside a loop produces a semantic error
func TestContinueOutsideLoop(t *testing.T) {
	input := `
		var x: Integer;
		x := 10;
		continue;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for continue outside loop, got nil")
	}

	if !strings.Contains(err.Error(), "continue statement not allowed outside loop") {
		t.Errorf("Expected error about continue outside loop, got: %v", err)
	}
}

// TestContinueInFinallyBlock tests that continue in finally block produces a semantic error
func TestContinueInFinallyBlock(t *testing.T) {
	input := `
		var i: Integer;
		for i := 1 to 5 do
		begin
			try
				PrintLn(i);
			finally
				continue;
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for continue in finally block, got nil")
	}

	if !strings.Contains(err.Error(), "continue statement not allowed in finally block") {
		t.Errorf("Expected error about continue in finally block, got: %v", err)
	}
}

// TestContinueInForLoop tests that continue in a for loop is valid
func TestContinueInForLoop(t *testing.T) {
	input := `
		var i: Integer;
		for i := 1 to 10 do
		begin
			if i = 5 then
				continue;
			PrintLn(i);
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for continue in for loop, got: %v", err)
	}
}

// TestContinueInWhileLoop tests that continue in a while loop is valid
func TestContinueInWhileLoop(t *testing.T) {
	input := `
		var i: Integer;
		i := 0;
		while i < 10 do
		begin
			i := i + 1;
			if i = 5 then
				continue;
			PrintLn(i);
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for continue in while loop, got: %v", err)
	}
}

// TestContinueInRepeatLoop tests that continue in a repeat loop is valid
func TestContinueInRepeatLoop(t *testing.T) {
	input := `
		var i: Integer;
		i := 0;
		repeat
		begin
			i := i + 1;
			if i = 5 then
				continue;
			PrintLn(i);
		end
		until i >= 10;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for continue in repeat loop, got: %v", err)
	}
}

// TestContinueInNestedLoops tests that continue in nested loops is valid
func TestContinueInNestedLoops(t *testing.T) {
	input := `
		var i: Integer;
		var j: Integer;
		for i := 1 to 10 do
		begin
			for j := 1 to 10 do
			begin
				if j = 5 then
					continue;
				PrintLn(j);
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for continue in nested loops, got: %v", err)
	}
}

// ============================================================================
// Exit Statement Semantic Analysis Tests (Task 8.235p)
// ============================================================================

// TestExitWithValueAtProgramLevel tests that exit with value at program level produces error
func TestExitWithValueAtProgramLevel(t *testing.T) {
	input := `
		var x: Integer;
		x := 10;
		exit(42);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for exit with value at program level, got nil")
	}

	if !strings.Contains(err.Error(), "exit with value not allowed at program level") {
		t.Errorf("Expected error about exit with value at program level, got: %v", err)
	}
}

// TestExitWithoutValueAtProgramLevel tests that exit without value at program level is valid
func TestExitWithoutValueAtProgramLevel(t *testing.T) {
	input := `
		var x: Integer;
		x := 10;
		if x = 10 then
			exit;
		PrintLn(x);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for exit without value at program level, got: %v", err)
	}
}

// TestExitInFinallyBlock tests that exit in finally block produces a semantic error
func TestExitInFinallyBlock(t *testing.T) {
	input := `
		function Test: Integer;
		begin
			try
				Result := 10;
			finally
				exit;
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for exit in finally block, got nil")
	}

	if !strings.Contains(err.Error(), "exit statement not allowed in finally block") {
		t.Errorf("Expected error about exit in finally block, got: %v", err)
	}
}

// TestExitInFunction tests that exit in a function is valid
func TestExitInFunction(t *testing.T) {
	input := `
		function GetValue(x: Integer): Integer;
		begin
			if x < 0 then
			begin
				Result := 0;
				exit;
			end;
			Result := x * 2;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for exit in function, got: %v", err)
	}
}

// TestExitInProcedure tests that exit in a procedure is valid
func TestExitInProcedure(t *testing.T) {
	input := `
		procedure DoSomething(x: Integer);
		begin
			if x < 0 then
				exit;
			PrintLn(x);
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for exit in procedure, got: %v", err)
	}
}

// ============================================================================
// Combined Tests (Task 8.235p)
// ============================================================================

// TestBreakContinueWithExceptionHandling tests break/continue with try-except
func TestBreakContinueWithExceptionHandling(t *testing.T) {
	input := `
		var i: Integer;
		for i := 1 to 10 do
		begin
			try
				if i = 5 then
					break;
			except
				on E: Exception do
					PrintLn(E.Message);
			end;
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for break with exception handling, got: %v", err)
	}
}

// ============================================================================
// For-In Statement Semantic Analysis Tests (Task 9.24)
// ============================================================================

// TestForInWithArray tests for-in loop with array type
func TestForInWithArray(t *testing.T) {
	input := `
		var arr: array of Integer;
		var x: Integer;
		for x in arr do
			PrintLn(x);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for for-in with array, got: %v", err)
	}
}

// TestForInWithInlineVar tests for-in loop with inline var declaration
func TestForInWithInlineVar(t *testing.T) {
	input := `
		var arr: array of Integer;
		for var x in arr do
			PrintLn(x);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for for-in with inline var, got: %v", err)
	}
}

// TestForInWithSet tests for-in loop with set type
func TestForInWithSet(t *testing.T) {
	input := `
		type TColor = (Red, Green, Blue);
		type TColorSet = set of TColor;
		var mySet: TColorSet;
		var color: TColor;
		for color in mySet do
			PrintLn(Integer(color));
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for for-in with set, got: %v", err)
	}
}

// TestForInWithString tests for-in loop with string type
func TestForInWithString(t *testing.T) {
	input := `
		var str: String;
		var ch: String;
		str := 'hello';
		for ch in str do
			Print(ch);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for for-in with string, got: %v", err)
	}
}

// TestForInWithStringLiteral tests for-in loop with string literal
func TestForInWithStringLiteral(t *testing.T) {
	input := `
		var ch: String;
		for ch in 'test' do
			Print(ch);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for for-in with string literal, got: %v", err)
	}
}

// TestForInWithNonEnumerableType tests for-in loop with non-enumerable type (should error)
func TestForInWithNonEnumerableType(t *testing.T) {
	input := `
		var x: Integer;
		var i: Integer;
		x := 42;
		for i in x do
			PrintLn(i);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for for-in with non-enumerable type, got nil")
	}

	if !strings.Contains(err.Error(), "not enumerable") {
		t.Errorf("Expected error about non-enumerable type, got: %v", err)
	}
}

// TestForInWithBoolean tests for-in loop with boolean type (should error)
func TestForInWithBoolean(t *testing.T) {
	input := `
		var flag: Boolean;
		var x: Boolean;
		flag := true;
		for x in flag do
			PrintLn(x);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for for-in with boolean type, got nil")
	}

	if !strings.Contains(err.Error(), "not enumerable") {
		t.Errorf("Expected error about non-enumerable type, got: %v", err)
	}
}

// TestForInLoopVariableScope tests that loop variable is scoped to the loop
func TestForInLoopVariableScope(t *testing.T) {
	input := `
		var arr: array of Integer;
		for var x in arr do
			PrintLn(x);
		// x should not be accessible here
		PrintLn(x);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("Expected semantic error for loop variable used outside scope, got nil")
	}

	if !strings.Contains(err.Error(), "undefined") {
		t.Errorf("Expected error about undefined variable, got: %v", err)
	}
}

// TestForInWithBreak tests break statement in for-in loop
func TestForInWithBreak(t *testing.T) {
	input := `
		var arr: array of Integer;
		for var x in arr do
		begin
			if x > 5 then
				break;
			PrintLn(x);
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for break in for-in loop, got: %v", err)
	}
}

// TestForInWithContinue tests continue statement in for-in loop
func TestForInWithContinue(t *testing.T) {
	input := `
		var arr: array of Integer;
		for var x in arr do
		begin
			if x = 0 then
				continue;
			PrintLn(x);
		end;
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for continue in for-in loop, got: %v", err)
	}
}

// TestNestedForInLoops tests nested for-in loops
func TestNestedForInLoops(t *testing.T) {
	input := `
		var matrix: array of array of Integer;
		for var row in matrix do
			for var cell in row do
				PrintLn(cell);
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for nested for-in loops, got: %v", err)
	}
}

// TestForInWithSetLiteral tests for-in loop with set literal
func TestForInWithSetLiteral(t *testing.T) {
	input := `
		type TColor = (Red, Green, Blue);
		var color: TColor;
		for color in [Red, Green, Blue] do
			PrintLn(Integer(color));
	`

	program := parseProgram(t, input)
	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no semantic error for for-in with set literal, got: %v", err)
	}
}

// Note: parseProgram helper function is defined in exceptions_test.go
