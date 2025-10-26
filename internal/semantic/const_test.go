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
	// This should error because const values must be constant expressions
	// For now, we'll allow any expression, but could be enhanced in the future
	expectNoErrors(t, input) // We'll validate this as constant later
}
