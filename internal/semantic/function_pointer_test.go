package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Task 9.163: Function Pointer Semantic Analysis Tests
// ============================================================================

// TestFunctionPointerTypeDeclaration tests valid function pointer type declarations.
// Task 9.163: Test valid type declarations
func TestFunctionPointerTypeDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple function pointer",
			input: `
				type TComparator = function(a, b: Integer): Integer;
			`,
		},
		{
			name: "simple procedure pointer",
			input: `
				type TCallback = procedure(msg: String);
			`,
		},
		{
			name: "function pointer with no parameters",
			input: `
				type TGetter = function(): Integer;
			`,
		},
		{
			name: "procedure pointer with no parameters",
			input: `
				type TAction = procedure();
			`,
		},
		{
			name: "method pointer (of object)",
			input: `
				type TNotifyEvent = procedure(Sender: TObject) of object;
			`,
		},
		{
			name: "function pointer with multiple parameters",
			input: `
				type TProcessor = function(x, y, z: Integer): Boolean;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestFunctionPointerTypeDeclarationErrors tests invalid function pointer type declarations.
// Task 9.163: Test invalid type declarations
func TestFunctionPointerTypeDeclarationErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "duplicate parameter names",
			input: `
				type TBadFunc = function(x, x: Integer): Integer;
			`,
			expectedErr: "duplicate parameter name",
		},
		{
			name: "non-existent parameter type",
			input: `
				type TBadFunc = function(x: UnknownType): Integer;
			`,
			expectedErr: "unknown parameter type",
		},
		{
			name: "non-existent return type",
			input: `
				type TBadFunc = function(x: Integer): UnknownType;
			`,
			expectedErr: "unknown return type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err == nil {
				t.Errorf("expected semantic error containing '%s', got no error", tt.expectedErr)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestAddressOfExpression tests valid address-of expressions.
// Task 9.163: Test address-of expression analysis
func TestAddressOfExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "address of function",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TBinaryOp = function(x, y: Integer): Integer;
				var f: TBinaryOp;
				begin
					f := @Add;
				end.
			`,
		},
		{
			name: "address of procedure",
			input: `
				procedure PrintMsg(msg: String);
				begin
					PrintLn(msg);
				end;

				type TCallback = procedure(msg: String);
				var p: TCallback;
				begin
					p := @PrintMsg;
				end.
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestAddressOfExpressionErrors tests invalid address-of expressions.
// Task 9.163: Test address-of expression validation
func TestAddressOfExpressionErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "address of undefined function",
			input: `
				type TFunc = function(): Integer;
				var f: TFunc;
				begin
					f := @UndefinedFunc;
				end.
			`,
			expectedErr: "undefined function",
		},
		{
			name: "address of variable (not a function)",
			input: `
				type TFunc = function(): Integer;
				var x: Integer;
				var f: TFunc;
				begin
					x := 42;
					f := @x;
				end.
			`,
			expectedErr: "not a function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err == nil {
				t.Errorf("expected semantic error containing '%s', got no error", tt.expectedErr)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestFunctionPointerAssignment tests function pointer assignment compatibility.
// Task 9.163: Test assignment validation
func TestFunctionPointerAssignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "compatible function pointer assignment",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TBinaryOp = function(x, y: Integer): Integer;
				var op: TBinaryOp;
				begin
					op := @Add;
				end.
			`,
		},
		{
			name: "compatible procedure pointer assignment",
			input: `
				procedure PrintMsg(msg: String);
				begin
					PrintLn(msg);
				end;

				type TCallback = procedure(s: String);
				var callback: TCallback;
				begin
					callback := @PrintMsg;
				end.
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestFunctionPointerAssignmentErrors tests incompatible function pointer assignments.
// Task 9.163: Test assignment validation errors
func TestFunctionPointerAssignmentErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "incompatible parameter count",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TUnaryOp = function(x: Integer): Integer;
				var op: TUnaryOp;
				begin
					op := @Add;
				end.
			`,
			expectedErr: "cannot assign", // Generic assignment error
		},
		{
			name: "incompatible parameter types",
			input: `
				function ProcessInt(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				type TStringFunc = function(s: String): Integer;
				var f: TStringFunc;
				begin
					f := @ProcessInt;
				end.
			`,
			expectedErr: "cannot assign",
		},
		{
			name: "incompatible return types",
			input: `
				function GetInt(): Integer;
				begin
					Result := 42;
				end;

				type TStringGetter = function(): String;
				var getter: TStringGetter;
				begin
					getter := @GetInt;
				end.
			`,
			expectedErr: "cannot assign",
		},
		{
			name: "function vs procedure mismatch",
			input: `
				procedure DoNothing();
				begin
				end;

				type TIntFunc = function(): Integer;
				var f: TIntFunc;
				begin
					f := @DoNothing;
				end.
			`,
			expectedErr: "cannot assign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err == nil {
				t.Errorf("expected semantic error containing '%s', got no error", tt.expectedErr)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestFunctionPointerCall tests valid function pointer calls.
// Task 9.163: Test function pointer call validation
func TestFunctionPointerCall(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "function pointer call with correct arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TBinaryOp = function(x, y: Integer): Integer;
				var op: TBinaryOp;
				var result: Integer;
				begin
					op := @Add;
					result := op(5, 3);
				end.
			`,
		},
		{
			name: "procedure pointer call",
			input: `
				procedure PrintMsg(msg: String);
				begin
					PrintLn(msg);
				end;

				type TCallback = procedure(s: String);
				var callback: TCallback;
				begin
					callback := @PrintMsg;
					callback('Hello');
				end.
			`,
		},
		{
			name: "function pointer call with no arguments",
			input: `
				function GetAnswer(): Integer;
				begin
					Result := 42;
				end;

				type TGetter = function(): Integer;
				var getter: TGetter;
				var answer: Integer;
				begin
					getter := @GetAnswer;
					answer := getter();
				end.
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestFunctionPointerCallErrors tests invalid function pointer calls.
// Task 9.163: Test function pointer call validation errors
func TestFunctionPointerCallErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "wrong number of arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TBinaryOp = function(x, y: Integer): Integer;
				var op: TBinaryOp;
				var result: Integer;
				begin
					op := @Add;
					result := op(5);
				end.
			`,
			expectedErr: "argument count mismatch",
		},
		{
			name: "wrong argument types",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				type TBinaryOp = function(x, y: Integer): Integer;
				var op: TBinaryOp;
				var result: Integer;
				begin
					op := @Add;
					result := op('hello', 3);
				end.
			`,
			expectedErr: "type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err == nil {
				t.Errorf("expected semantic error containing '%s', got no error", tt.expectedErr)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestFunctionPointerTypeInference tests that function pointer types are correctly inferred.
// Task 9.163: Test type inference
func TestFunctionPointerTypeInference(t *testing.T) {
	input := `
		function Multiply(a, b: Integer): Integer;
		begin
			Result := a * b;
		end;

		type TBinaryOp = function(x, y: Integer): Integer;
		var op: TBinaryOp;
		var x: Integer;
		var y: Integer;
		var result: Integer;
		begin
			op := @Multiply;
			x := 5;
			y := 3;
			result := op(x, y);
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	err := a.Analyze(program)
	if err != nil {
		t.Errorf("expected no semantic errors, got: %v", err)
	}
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}

	// Verify that the function pointer type was registered
	if _, exists := a.functionPointers["TBinaryOp"]; !exists {
		t.Errorf("expected function pointer type 'TBinaryOp' to be registered")
	}
}

// TestMultipleFunctionPointerTypes tests using multiple function pointer types.
// Task 9.163: Test multiple function pointer type declarations
func TestMultipleFunctionPointerTypes(t *testing.T) {
	input := `
		type TComparator = function(a, b: Integer): Integer;
		type TCallback = procedure(msg: String);
		type TGetter = function(): Boolean;

		function Compare(x, y: Integer): Integer;
		begin
			if x < y then Result := -1
			else if x > y then Result := 1
			else Result := 0;
		end;

		procedure Notify(s: String);
		begin
			PrintLn(s);
		end;

		function IsReady(): Boolean;
		begin
			Result := true;
		end;

		var cmp: TComparator;
		var cb: TCallback;
		var getter: TGetter;
		begin
			cmp := @Compare;
			cb := @Notify;
			getter := @IsReady;
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	err := a.Analyze(program)
	if err != nil {
		t.Errorf("expected no semantic errors, got: %v", err)
	}
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}

	// Verify all function pointer types were registered
	expectedTypes := []string{"TComparator", "TCallback", "TGetter"}
	for _, typeName := range expectedTypes {
		if _, exists := a.functionPointers[typeName]; !exists {
			t.Errorf("expected function pointer type '%s' to be registered", typeName)
		}
	}
}

// TestFunctionPointerVarDeclaration tests declaring variables with function pointer types.
// Task 9.163: Test variable declarations with function pointer types
func TestFunctionPointerVarDeclaration(t *testing.T) {
	input := `
		type TFunc = function(x: Integer): Integer;
		var f: TFunc;
		begin
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	a := NewAnalyzer()
	err := a.Analyze(program)
	if err != nil {
		t.Errorf("expected no semantic errors, got: %v", err)
	}
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}

	// Verify the variable was registered with the correct type
	sym, ok := a.symbols.Resolve("f")
	if !ok {
		t.Errorf("expected variable 'f' to be in symbol table")
	}
	// The type should be a TypeAlias that wraps a FunctionPointerType
	if sym != nil && sym.Type != nil {
		// Just verify it's not nil - detailed type checking is done in other tests
		if sym.Type.TypeKind() != "FUNCTION_POINTER" && sym.Type.TypeKind() != "TYPE_ALIAS" {
			t.Errorf("expected variable 'f' to have function pointer or alias type, got: %s", sym.Type.TypeKind())
		}
	}
}
