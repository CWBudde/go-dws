package semantic

import (
	"testing"
)

// ============================================================================
// Const Declaration Tests
// ============================================================================

func TestConstDeclarationWithType(t *testing.T) {
	input := `
		const PI: Float = 3.14;
		const MAX: Integer = 100;
		const NAME: String = 'test';
		const FLAG: Boolean = true;
	`
	expectNoErrors(t, input)
}

func TestConstDeclarationWithTypeInference(t *testing.T) {
	input := `
		const PI = 3.14;
		const MAX = 100;
		const NAME = 'test';
		const FLAG = true;
	`
	expectNoErrors(t, input)
}

func TestConstDeclarationTypeMismatch(t *testing.T) {
	input := `const X: Integer = 'hello';`
	expectError(t, input, "cannot assign String to Integer")
}

func TestConstRedeclaration(t *testing.T) {
	input := `
		const PI = 3.14;
		const PI = 2.71;
	`
	expectError(t, input, "already declared")
}

func TestConstUsageInExpression(t *testing.T) {
	input := `
		const PI = 3.14;
		const RADIUS = 5.0;
		var area: Float := PI * RADIUS * RADIUS;
	`
	expectNoErrors(t, input)
}

func TestAssignmentToConst(t *testing.T) {
	input := `
		const MAX = 100;
		MAX := 200;
	`
	expectError(t, input, "Cannot assign to constant")
}

func TestConstWithVariableReference(t *testing.T) {
	input := `
		var x: Integer := 10;
		const MAX = x;
	`
	// Task 9.205: Const values must be compile-time constants
	// This should error because x is a variable, not a constant
	expectError(t, input, "identifier 'x' is not a constant")
}
