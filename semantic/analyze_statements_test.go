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
	expectError(t, input, "must have either a type annotation or an initializer")
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
