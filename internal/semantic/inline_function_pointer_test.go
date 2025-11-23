package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Inline Function Pointer Semantic Analysis Tests
// ============================================================================

// TestInlineFunctionPointerInVariable tests inline function pointer types in variable declarations.
func TestInlineFunctionPointerInVariable(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple function pointer variable",
			input: `
				var f: function(x: Integer): Integer;
				begin
				end.
			`,
		},
		{
			name: "simple procedure pointer variable",
			input: `
				var callback: procedure(msg: String);
				begin
				end.
			`,
		},
		{
			name: "function pointer with no parameters",
			input: `
				var getter: function(): Boolean;
				begin
				end.
			`,
		},
		{
			name: "procedure pointer with no parameters",
			input: `
				var action: procedure();
				begin
				end.
			`,
		},
		{
			name: "method pointer (of object)",
			input: `
				var handler: procedure(Sender: TObject) of object;
				begin
				end.
			`,
		},
		{
			name: "function pointer with multiple parameters",
			input: `
				var processor: function(x, y, z: Integer): Boolean;
				begin
				end.
			`,
		},
		{
			name: "function pointer with different parameter types",
			input: `
				var converter: function(s: String; count: Integer): Boolean;
				begin
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

// TestInlineFunctionPointerInParameter tests inline function pointer types in function parameters.
func TestInlineFunctionPointerInParameter(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple function pointer parameter",
			input: `
				procedure Apply(f: function(x: Integer): Integer);
				begin
				end;
				begin
				end.
			`,
		},
		{
			name: "simple procedure pointer parameter",
			input: `
				procedure Run(callback: procedure(msg: String));
				begin
				end;
				begin
				end.
			`,
		},
		{
			name: "function pointer with no parameters",
			input: `
				procedure Execute(action: function(): Integer);
				begin
				end;
				begin
				end.
			`,
		},
		{
			name: "method pointer parameter (of object)",
			input: `
				procedure SetHandler(handler: procedure(Sender: TObject) of object);
				begin
				end;
				begin
				end.
			`,
		},
		{
			name: "multiple function pointer parameters",
			input: `
				procedure Transform(mapper: function(x: Integer): String; filter: function(s: String): Boolean);
				begin
				end;
				begin
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

// TestInlineFunctionPointerTypeErrors tests error cases for inline function pointer types.
func TestInlineFunctionPointerTypeErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "unknown parameter type in inline function pointer",
			input: `
				var f: function(x: UnknownType): Integer;
				begin
				end.
			`,
			expectedErr: "unknown type",
		},
		{
			name: "unknown return type in inline function pointer",
			input: `
				var f: function(x: Integer): UnknownType;
				begin
				end.
			`,
			expectedErr: "unknown type",
		},
		{
			name: "unknown parameter type in inline procedure pointer",
			input: `
				var p: procedure(x: UnknownType);
				begin
				end.
			`,
			expectedErr: "unknown type",
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

			if !ErrorMatches(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestInlineFunctionPointerAssignment tests assignment compatibility with inline function pointers.
func TestInlineFunctionPointerAssignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "assign function to inline function pointer variable",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				var op: function(x, y: Integer): Integer;
				begin
					op := @Add;
				end.
			`,
		},
		{
			name: "assign procedure to inline procedure pointer variable",
			input: `
				procedure PrintMsg(msg: String);
				begin
					PrintLn(msg);
				end;

				var callback: procedure(s: String);
				begin
					callback := @PrintMsg;
				end.
			`,
		},
		{
			name: "assign function with no params to inline function pointer",
			input: `
				function GetAnswer(): Integer;
				begin
					Result := 42;
				end;

				var getter: function(): Integer;
				begin
					getter := @GetAnswer;
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

// TestInlineFunctionPointerAssignmentErrors tests incompatible assignments with inline function pointers.
func TestInlineFunctionPointerAssignmentErrors(t *testing.T) {
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

				var op: function(x: Integer): Integer;
				begin
					op := @Add;
				end.
			`,
			expectedErr: "cannot assign",
		},
		{
			name: "incompatible parameter types",
			input: `
				function ProcessInt(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				var f: function(s: String): Integer;
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

				var getter: function(): String;
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

				var f: function(): Integer;
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

			if !ErrorMatches(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestInlineFunctionPointerCall tests calling inline function pointer variables.
func TestInlineFunctionPointerCall(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "call inline function pointer with correct arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				var op: function(x, y: Integer): Integer;
				var result: Integer;
				begin
					op := @Add;
					result := op(5, 3);
				end.
			`,
		},
		{
			name: "call inline procedure pointer",
			input: `
				procedure PrintMsg(msg: String);
				begin
					PrintLn(msg);
				end;

				var callback: procedure(s: String);
				begin
					callback := @PrintMsg;
					callback('Hello');
				end.
			`,
		},
		{
			name: "call inline function pointer with no arguments",
			input: `
				function GetAnswer(): Integer;
				begin
					Result := 42;
				end;

				var getter: function(): Integer;
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

// TestInlineFunctionPointerCallErrors tests error cases for calling inline function pointers.
func TestInlineFunctionPointerCallErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "wrong number of arguments to inline function pointer",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				var op: function(x, y: Integer): Integer;
				var result: Integer;
				begin
					op := @Add;
					result := op(5);
				end.
			`,
			expectedErr: "argument count mismatch",
		},
		{
			name: "wrong argument types to inline function pointer",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				var op: function(x, y: Integer): Integer;
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

			if !ErrorMatches(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestMixedInlineAndAliasedFunctionPointers tests using both inline and aliased function pointer types together.
func TestMixedInlineAndAliasedFunctionPointers(t *testing.T) {
	input := `
		// Aliased function pointer type
		type TComparator = function(a, b: Integer): Integer;

		function Compare(x, y: Integer): Integer;
		begin
			if x < y then Result := -1
			else if x > y then Result := 1
			else Result := 0;
		end;

		// Use both aliased and inline types
		var cmp1: TComparator;
		var cmp2: function(a, b: Integer): Integer;
		begin
			cmp1 := @Compare;
			cmp2 := @Compare;
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
}
