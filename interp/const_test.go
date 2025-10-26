package interp

import (
	"testing"
)

// ============================================================================
// Const Declaration Tests (Task 8.258)
// ============================================================================

func TestConstDeclarationInteger(t *testing.T) {
	input := `
		const MAX = 100;
		MAX
	`
	val := testEval(input)
	testIntegerValue(t, val, 100)
}

func TestConstDeclarationFloat(t *testing.T) {
	input := `
		const PI = 3.14;
		PI
	`
	val := testEval(input)
	testFloatValue(t, val, 3.14)
}

func TestConstDeclarationString(t *testing.T) {
	input := `
		const APP_NAME = 'MyApp';
		APP_NAME
	`
	val := testEval(input)
	testStringValue(t, val, "MyApp")
}

func TestConstDeclarationBoolean(t *testing.T) {
	input := `
		const FLAG = true;
		FLAG
	`
	val := testEval(input)
	testBooleanValue(t, val, true)
}

func TestConstInExpression(t *testing.T) {
	input := `
		const X = 5;
		const Y = 10;
		X * 2 + Y
	`
	val := testEval(input)
	testIntegerValue(t, val, 20)
}

func TestMultipleConstDeclarations(t *testing.T) {
	input := `
		const PI = 3.14;
		const RADIUS = 5.0;
		var area: Float := PI * RADIUS * RADIUS;
		area
	`
	val := testEval(input)
	testFloatValue(t, val, 78.5)
}

func TestConstUsedInVariableInitializer(t *testing.T) {
	input := `
		const MAX = 100;
		var x: Integer := MAX * 2;
		x
	`
	val := testEval(input)
	testIntegerValue(t, val, 200)
}

func TestConstTypedDeclaration(t *testing.T) {
	input := `
		const MAX: Integer = 100;
		MAX
	`
	val := testEval(input)
	testIntegerValue(t, val, 100)
}

func TestConstExpressionValue(t *testing.T) {
	input := `
		const RESULT = 10 + 20 * 2;
		RESULT
	`
	val := testEval(input)
	testIntegerValue(t, val, 50)
}
