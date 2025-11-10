package semantic

import (
	"testing"
)

// ============================================================================
// Variable Declaration Tests
// ============================================================================

func TestVariableDeclarationWithType(t *testing.T) {
	input := `
		var x: Integer;
		var s: String;
		var f: Float;
		var b: Boolean;
	`
	expectNoErrors(t, input)
}

func TestVariableDeclarationWithInitializer(t *testing.T) {
	input := `
		var x: Integer := 42;
		var s: String := 'hello';
		var f: Float := 3.14;
		var b: Boolean := true;
	`
	expectNoErrors(t, input)
}

func TestVariableDeclarationWithTypeInference(t *testing.T) {
	input := `
		var x := 42;
		var s := 'hello';
		var f := 3.14;
		var b := true;
	`
	expectNoErrors(t, input)
}

func TestVariableDeclarationTypeMismatch(t *testing.T) {
	input := `var x: Integer := 'hello';`
	expectError(t, input, "cannot assign String to Integer")
}

func TestVariableDeclarationNoTypeNoInitializer(t *testing.T) {
	input := `var x;`
	expectError(t, input, "variable declaration requires a type or initializer")
}

func TestVariableRedeclaration(t *testing.T) {
	input := `
		var x: Integer;
		var x: String;
	`
	expectError(t, input, "already declared")
}

func TestVariableRedeclarationInDifferentScope(t *testing.T) {
	// Same variable name in different scopes should be allowed
	input := `
		var x: Integer;
		begin
			var x: String;
		end;
	`
	expectNoErrors(t, input)
}

func TestUnknownType(t *testing.T) {
	input := `var x: UnknownType;`
	expectError(t, input, "unknown type")
}

// ============================================================================
// Undefined Variable Tests
// ============================================================================

func TestUndefinedVariable(t *testing.T) {
	input := `
		var x: Integer;
		y := 10;
	`
	expectError(t, input, "undefined variable 'y'")
}

func TestUndefinedVariableInExpression(t *testing.T) {
	input := `
		var x: Integer := y + 5;
	`
	expectError(t, input, "undefined variable 'y'")
}

// ============================================================================
// Assignment Tests
// ============================================================================

func TestAssignment(t *testing.T) {
	input := `
		var x: Integer;
		x := 42;
	`
	expectNoErrors(t, input)
}

func TestAssignmentTypeMismatch(t *testing.T) {
	input := `
		var x: Integer;
		x := 'hello';
	`
	expectError(t, input, "cannot assign String to Integer")
}

func TestAssignmentWithCoercion(t *testing.T) {
	// Integer can be assigned to Float
	input := `
		var f: Float;
		f := 42;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Control Flow Tests
// ============================================================================

func TestIfStatement(t *testing.T) {
	input := `
		var x: Integer := 10;
		if x > 5 then
			x := x + 1;
	`
	expectNoErrors(t, input)
}

func TestIfElseStatement(t *testing.T) {
	input := `
		var x: Integer := 10;
		if x > 5 then
			x := x + 1
		else
			x := x - 1;
	`
	expectNoErrors(t, input)
}

func TestIfConditionNotBoolean(t *testing.T) {
	input := `
		var x: Integer := 10;
		if x then
			x := x + 1;
	`
	expectError(t, input, "if condition must be boolean")
}

func TestWhileStatement(t *testing.T) {
	input := `
		var x: Integer := 0;
		while x < 10 do
			x := x + 1;
	`
	expectNoErrors(t, input)
}

func TestWhileConditionNotBoolean(t *testing.T) {
	input := `
		var x: Integer := 10;
		while x do
			x := x - 1;
	`
	expectError(t, input, "while condition must be boolean")
}

func TestRepeatStatement(t *testing.T) {
	input := `
		var x: Integer := 0;
		repeat
			x := x + 1
		until x >= 10;
	`
	expectNoErrors(t, input)
}

func TestRepeatConditionNotBoolean(t *testing.T) {
	input := `
		var x: Integer := 0;
		repeat
			x := x + 1
		until x;
	`
	expectError(t, input, "repeat-until condition must be boolean")
}

func TestForStatement(t *testing.T) {
	input := `
		var sum: Integer := 0;
		for i := 1 to 10 do
			sum := sum + i;
	`
	expectNoErrors(t, input)
}

func TestForStatementDownto(t *testing.T) {
	input := `
		var sum: Integer := 0;
		for i := 10 downto 1 do
			sum := sum + i;
	`
	expectNoErrors(t, input)
}

func TestForStatementNonOrdinalType(t *testing.T) {
	input := `
		for i := 1.0 to 10.0 do
			PrintLn(i);
	`
	expectError(t, input, "ordinal type")
}

// Semantic tests for for-loop step feature
func TestForStatementWithStepInteger(t *testing.T) {
	input := `
		var sum: Integer := 0;
		for i := 1 to 10 step 2 do
			sum := sum + i;
	`
	expectNoErrors(t, input)
}

func TestForStatementWithStepExpression(t *testing.T) {
	input := `
		var x: Integer := 1;
		var sum: Integer := 0;
		for i := 1 to 10 step (x + 1) do
			sum := sum + i;
	`
	expectNoErrors(t, input)
}

func TestForStatementWithStepVariable(t *testing.T) {
	input := `
		var stepValue: Integer := 2;
		var sum: Integer := 0;
		for i := 1 to 10 step stepValue do
			sum := sum + i;
	`
	expectNoErrors(t, input)
}

func TestForStatementStepFloatError(t *testing.T) {
	input := `
		for i := 1 to 10 step 2.5 do
			PrintLn(i);
	`
	expectError(t, input, "step must be Integer")
}

func TestForStatementStepStringError(t *testing.T) {
	input := `
		for i := 1 to 10 step "text" do
			PrintLn(i);
	`
	expectError(t, input, "step must be Integer")
}

func TestForStatementStepZeroError(t *testing.T) {
	input := `
		for i := 1 to 10 step 0 do
			PrintLn(i);
	`
	expectError(t, input, "step must be strictly positive")
}

func TestForStatementStepNegativeError(t *testing.T) {
	input := `
		for i := 1 to 10 step -1 do
			PrintLn(i);
	`
	expectError(t, input, "step must be strictly positive")
}

func TestForStatementDowntoWithStep(t *testing.T) {
	input := `
		var sum: Integer := 0;
		for i := 10 downto 1 step 3 do
			sum := sum + i;
	`
	expectNoErrors(t, input)
}

func TestCaseStatement(t *testing.T) {
	input := `
		var x: Integer := 5;
		case x of
			1: PrintLn('one');
			2: PrintLn('two');
		else
			PrintLn('other');
		end;
	`
	expectNoErrors(t, input)
}

func TestCaseTypeMismatch(t *testing.T) {
	input := `
		var x: Integer := 5;
		case x of
			1: PrintLn('one');
			'two': PrintLn('two');
		end;
	`
	expectError(t, input, "incompatible")
}

// ============================================================================
// Compound Assignment Tests
// ============================================================================

func TestCompoundAssignmentInteger(t *testing.T) {
	input := `
		var x: Integer := 10;
		x += 5;
		x -= 2;
		x *= 3;
		x /= 2;
	`
	expectNoErrors(t, input)
}

func TestCompoundAssignmentFloat(t *testing.T) {
	input := `
		var f: Float := 3.14;
		f += 1.0;
		f -= 0.5;
		f *= 2.0;
		f /= 1.5;
	`
	expectNoErrors(t, input)
}

func TestCompoundAssignmentStringConcat(t *testing.T) {
	input := `
		var s: String := 'hello';
		s += ' world';
	`
	expectNoErrors(t, input)
}

func TestCompoundAssignmentBooleanError(t *testing.T) {
	input := `
		var b: Boolean := true;
		b += false;
	`
	expectError(t, input, "operator += not supported for type Boolean")
}

func TestCompoundAssignmentStringMultiplyError(t *testing.T) {
	input := `
		var s: String := 'hello';
		s *= 2;
	`
	expectError(t, input, "operator *= not supported for type String")
}

func TestCompoundAssignmentConstError(t *testing.T) {
	input := `
		const MYPI = 3.14;
		PI += 1.0;
	`
	expectError(t, input, "Cannot assign to constant")
}

func TestCompoundAssignmentArrayElement(t *testing.T) {
	input := `
		var arr: array[0..4] of Integer;
		arr[0] := 10;
		arr[0] += 5;
	`
	expectNoErrors(t, input)
}

func TestCompoundAssignmentMemberAccess(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;
		end;

		var p: TPoint;
		p.X := 10;
		p.X += 5;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Exit Statement Tests
// ============================================================================

func TestExitWithValueInFunction(t *testing.T) {
	input := `
		function GetValue: Integer;
		begin
			Exit 42;
		end;
	`
	expectNoErrors(t, input)
}

func TestExitWithWrongTypeInFunction(t *testing.T) {
	input := `
		function GetValue: Integer;
		begin
			Exit 'hello';
		end;
	`
	expectError(t, input, "exit value type String incompatible with function return type Integer")
}

func TestExitWithValueAtProgramLevelError(t *testing.T) {
	input := `
		var x: Integer := 5;
		Exit 42;
	`
	expectError(t, input, "exit with value not allowed at program level")
}

func TestExitWithoutValueInFunctionSemantic(t *testing.T) {
	input := `
		function GetValue: Integer;
		begin
			Exit;
		end;
	`
	expectNoErrors(t, input)
}

func TestExitInProcedureNoValue(t *testing.T) {
	input := `
		procedure DoSomething;
		begin
			Exit;
		end;
	`
	expectNoErrors(t, input)
}
