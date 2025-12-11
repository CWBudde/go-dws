package semantic

import (
	"testing"
)

// ============================================================================
// Const Declaration Tests
// ============================================================================

func TestConstDeclarationWithType(t *testing.T) {
	input := `
		const MYPI: Float = 3.14;
		const MAX: Integer = 100;
		const NAME: String = 'test';
		const FLAG: Boolean = true;
	`
	expectNoErrors(t, input)
}

func TestConstDeclarationWithTypeInference(t *testing.T) {
	input := `
		const MYPI = 3.14;
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
		const MYPI = 3.14;
		const MYPI = 2.71;
	`
	expectError(t, input, "already exists")
}

func TestConstUsageInExpression(t *testing.T) {
	input := `
		const MYPI = 3.14;
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
	// Const values must be compile-time constants
	// This should error because x is a variable, not a constant
	expectError(t, input, "identifier 'x' is not a constant")
}

// ============================================================================
// Const with Built-in Function Calls
// ============================================================================

func TestConstWithHighFunction(t *testing.T) {
	input := `const MAX = High(Integer);`
	expectNoErrors(t, input)
}

func TestConstWithFloorFunction(t *testing.T) {
	input := `const VAL = Floor(3.7);`
	expectNoErrors(t, input)
}

func TestConstWithLog2Function(t *testing.T) {
	input := `const VAL = Log2(16.0);`
	expectNoErrors(t, input)
}

func TestConstWithNestedFunctions(t *testing.T) {
	input := `const VAL = Floor(Log2(256.0));`
	expectNoErrors(t, input)
}

func TestConstWithComplexExpression(t *testing.T) {
	// This is the exact pattern from lucas_lehmer.pas
	input := `const upperBound = Floor(Log2(High(Integer))/2) - 1;`
	expectNoErrors(t, input)
}

func TestConstWithFunctionAndArithmetic(t *testing.T) {
	input := `const VAL = High(Integer) / 2;`
	expectNoErrors(t, input)
}

func TestConstWithUnsupportedFunction(t *testing.T) {
	input := `
		function MyFunc(): Integer;
		begin
			Result := 42;
		end;
		const VAL = MyFunc();
	`
	// User-defined functions are not compile-time evaluable
	expectError(t, input, "not a compile-time constant")
}

// ============================================================================
// Const Expression Evaluator Tests
// ============================================================================

// Test string concatenation in const expressions
func TestConstWithStringConcatenation(t *testing.T) {
	input := `const GREETING = 'Hello' + ' ' + 'World';`
	expectNoErrors(t, input)
}

func TestConstWithEmptyStringConcatenation(t *testing.T) {
	input := `const CRLF = '' + #13#10;`
	expectNoErrors(t, input)
}

// Test character literals in const expressions
func TestConstWithCharacterLiteral(t *testing.T) {
	input := `const CR = #13;`
	expectNoErrors(t, input)
}

func TestConstWithMultipleCharacterLiterals(t *testing.T) {
	input := `
		const CR = #13;
		const LF = #10;
		const CRLF = CR + LF;
	`
	expectNoErrors(t, input)
}

func TestConstWithCharacterLiteralConcatenation(t *testing.T) {
	input := `const NEWLINE = #13 + #10;`
	expectNoErrors(t, input)
}

func TestConstBlockSharesScope(t *testing.T) {
	input := `
		const
			C1 = 1;
			C2 = C1 + 1;

		var v: Integer := C2;

		begin
			PrintLn(C1 + v);
		end;
	`
	expectNoErrors(t, input)
}

// Test Ord/Chr in const context
func TestConstWithOrdFunction(t *testing.T) {
	input := `const CODE = Ord('A');`
	expectNoErrors(t, input)
}

func TestConstWithChrFunction(t *testing.T) {
	input := `const LETTER = Chr(65);`
	expectNoErrors(t, input)
}

func TestConstWithOrdAndArithmetic(t *testing.T) {
	input := `
		const MIN_CHAR = Ord('A');
		const MAX_CHAR = Ord('Z');
		const RANGE = MAX_CHAR - MIN_CHAR + 1;
	`
	expectNoErrors(t, input)
}

// Test function-local const declarations
func TestFunctionLocalConstDeclaration(t *testing.T) {
	input := `
		function Test(): Integer;
		begin
			const LOCAL_CONST = 42;
			Result := LOCAL_CONST;
		end;
	`
	expectNoErrors(t, input)
}

func TestFunctionLocalConstWithOrd(t *testing.T) {
	input := `
		function Vigenere(src: String): Integer;
		begin
			const cOrdMinChar = Ord('A');
			const cOrdMaxChar = Ord('Z');
			const cCharRangeCount = cOrdMaxChar - cOrdMinChar + 1;
			Result := cCharRangeCount;
		end;
	`
	expectNoErrors(t, input)
}
