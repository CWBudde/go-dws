package semantic

import (
	"testing"
)

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
			Result := 42;
		end;
	`
	expectNoErrors(t, input)
}

func TestReturnTypeMismatch(t *testing.T) {
	input := `
		function GetValue(): Integer;
		begin
			Result := 'hello';
		end;
	`
	expectError(t, input, "cannot assign String to Integer")
}

func TestReturnInProcedure(t *testing.T) {
	input := `
		procedure DoSomething;
		begin
			Result := 42;
		end;
	`
	expectError(t, input, "undefined variable 'Result'")
}

func TestReturnOutsideFunction(t *testing.T) {
	input := `
		begin
			Result := 42;
		end;
	`
	expectError(t, input, "undefined variable 'Result'")
}
