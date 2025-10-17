package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// Helper function to create an analyzer and analyze source code
func analyzeSource(t *testing.T, input string) (*Analyzer, error) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)
	return analyzer, err
}

// Helper function to check that analysis succeeds
func expectNoErrors(t *testing.T, input string) {
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// Helper function to check that analysis fails with specific error
func expectError(t *testing.T, input string, expectedError string) {
	_, err := analyzeSource(t, input)
	if err == nil {
		t.Errorf("expected error containing '%s', got no error", expectedError)
		return
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error containing '%s', got: %v", expectedError, err)
	}
}

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
// Expression Type Checking Tests
// ============================================================================

func TestArithmeticExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"integer addition", "var x := 3 + 5;", true},
		{"float addition", "var x := 3.14 + 2.86;", true},
		{"mixed addition", "var x := 3 + 2.5;", true}, // Integer + Float -> Float
		{"integer subtraction", "var x := 10 - 5;", true},
		{"integer multiplication", "var x := 4 * 5;", true},
		{"float division", "var x := 10.0 / 2.0;", true},
		{"string + number", "var x: String; x := 'hello' + 5;", false},
		{"number + string", "var x: Integer; x := 5 + 'hello';", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `var s := 'hello' + ' ' + 'world';`
	expectNoErrors(t, input)
}

func TestStringConcatenationError(t *testing.T) {
	input := `var s := 'hello' + 42;`
	expectError(t, input, "string concatenation requires both operands to be strings")
}

func TestIntegerOperations(t *testing.T) {
	input := `
		var a := 10 div 3;
		var b := 10 mod 3;
	`
	expectNoErrors(t, input)
}

func TestIntegerOperationsError(t *testing.T) {
	input := `var x := 3.14 div 2.0;`
	expectError(t, input, "requires integer operands")
}

func TestComparisonOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"integer equality", "var b := 3 = 5;", true},
		{"integer inequality", "var b := 3 <> 5;", true},
		{"integer less than", "var b := 3 < 5;", true},
		{"integer greater than", "var b := 3 > 5;", true},
		{"string equality", "var b := 'a' = 'b';", true},
		{"string comparison", "var b := 'a' < 'b';", true},
		{"float comparison", "var b := 3.14 > 2.86;", true},
		{"mixed comparison", "var b := 3 < 2.5;", true}, // Integer vs Float
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestLogicalOperations(t *testing.T) {
	input := `
		var a := true and false;
		var b := true or false;
		var c := true xor false;
		var d := not true;
	`
	expectNoErrors(t, input)
}

func TestLogicalOperationsError(t *testing.T) {
	tests := []string{
		"var x := 3 and 5;",
		"var x := 'hello' or 'world';",
		"var x := not 42;",
	}

	for _, input := range tests {
		expectError(t, input, "boolean")
	}
}

func TestUnaryOperations(t *testing.T) {
	input := `
		var a := -5;
		var b := +3.14;
		var c := not true;
	`
	expectNoErrors(t, input)
}

func TestUnaryOperationsError(t *testing.T) {
	tests := []struct {
		input string
		error string
	}{
		{"var x := -'hello';", "numeric operand"},
		{"var x := not 42;", "boolean operand"},
	}

	for _, tt := range tests {
		expectError(t, tt.input, tt.error)
	}
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

// ============================================================================
// Function Declaration Tests
// ============================================================================

func TestFunctionDeclaration(t *testing.T) {
	input := `
		function Add(a: Integer; b: Integer): Integer;
		begin
			Result := a + b;
		end;
	`
	expectNoErrors(t, input)
}

func TestProcedureDeclaration(t *testing.T) {
	input := `
		procedure SayHello;
		begin
			PrintLn('Hello');
		end;
	`
	expectNoErrors(t, input)
}

func TestFunctionRedeclaration(t *testing.T) {
	input := `
		function Add(a: Integer; b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function Add(x: Integer; y: Integer): Integer;
		begin
			Result := x + y;
		end;
	`
	expectError(t, input, "already declared")
}

func TestFunctionWithUnknownParameterType(t *testing.T) {
	input := `
		function Add(a: UnknownType): Integer;
		begin
			Result := 0;
		end;
	`
	expectError(t, input, "unknown parameter type")
}

func TestFunctionWithUnknownReturnType(t *testing.T) {
	input := `
		function GetValue(): UnknownType;
		begin
			Result := nil;
		end;
	`
	expectError(t, input, "unknown return type")
}

func TestFunctionParameterScope(t *testing.T) {
	input := `
		function Add(a: Integer; b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var a: String := 'test';
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Function Call Tests
// ============================================================================

func TestFunctionCall(t *testing.T) {
	input := `
		function Add(a: Integer; b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var result: Integer := Add(2, 3);
	`
	expectNoErrors(t, input)
}

func TestProcedureCall(t *testing.T) {
	input := `
		procedure SayHello;
		begin
			PrintLn('Hello');
		end;

		begin
			SayHello;
		end;
	`
	expectNoErrors(t, input)
}

func TestUndefinedFunction(t *testing.T) {
	input := `var x := DoSomething(42);`
	expectError(t, input, "undefined function")
}

func TestFunctionCallWrongArgumentCount(t *testing.T) {
	input := `
		function Add(a: Integer; b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var x := Add(5);
	`
	expectError(t, input, "expects 2 arguments, got 1")
}

func TestFunctionCallWrongArgumentType(t *testing.T) {
	input := `
		function Add(a: Integer; b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var x := Add(5, 'hello');
	`
	expectError(t, input, "has type String, expected Integer")
}

func TestFunctionCallWithCoercion(t *testing.T) {
	input := `
		function AddFloat(a: Float; b: Float): Float;
		begin
			Result := a + b;
		end;

		var x := AddFloat(5, 3.14);
	`
	expectNoErrors(t, input)
}

func TestBuiltInFunctions(t *testing.T) {
	input := `
		begin
			PrintLn('Hello');
			PrintLn(42);
			Print('World');
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Return Statement Tests
// ============================================================================

func TestReturnStatement(t *testing.T) {
	input := `
		function GetValue(): Integer;
		begin
			return 42;
		end;
	`
	expectNoErrors(t, input)
}

func TestReturnTypeMismatch(t *testing.T) {
	input := `
		function GetValue(): Integer;
		begin
			return 'hello';
		end;
	`
	expectError(t, input, "return type String incompatible")
}

func TestReturnInProcedure(t *testing.T) {
	input := `
		procedure DoSomething;
		begin
			return 42;
		end;
	`
	expectError(t, input, "procedure cannot return a value")
}

func TestReturnOutsideFunction(t *testing.T) {
	input := `
		begin
			return 42;
		end;
	`
	expectError(t, input, "return statement outside of function")
}

// ============================================================================
// Complex Integration Tests
// ============================================================================

func TestComplexProgram(t *testing.T) {
	input := `
		function Factorial(n: Integer): Integer;
		begin
			if n <= 1 then
				Result := 1
			else
				Result := n * Factorial(n - 1);
		end;

		var result: Integer;
		for i := 1 to 5 do
		begin
			result := Factorial(i);
			PrintLn(result);
		end;
	`
	expectNoErrors(t, input)
}

func TestNestedScopes(t *testing.T) {
	input := `
		var x: Integer := 10;
		begin
			var x: String := 'outer';
			begin
				var x: Float := 3.14;
				PrintLn(x);
			end;
			PrintLn(x);
		end;
		PrintLn(x);
	`
	expectNoErrors(t, input)
}

func TestFunctionWithLocalVariables(t *testing.T) {
	input := `
		function Calculate(n: Integer): Integer;
		var temp: Integer;
		var result: Integer;
		begin
			temp := n * 2;
			result := temp + 10;
			Result := result;
		end;

		var x := Calculate(5);
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestEmptyProgram(t *testing.T) {
	input := ``
	expectNoErrors(t, input)
}

func TestOnlyComments(t *testing.T) {
	input := `
		// This is a comment
		(* This is another comment *)
		{ And another }
	`
	expectNoErrors(t, input)
}

func TestMultipleErrors(t *testing.T) {
	input := `
		var x: UnknownType;
		var y := z;
		x := 'hello';
	`
	analyzer, err := analyzeSource(t, input)
	if err == nil {
		t.Error("expected errors")
		return
	}

	if len(analyzer.Errors()) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(analyzer.Errors()))
	}
}
