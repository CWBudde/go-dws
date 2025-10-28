package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestInlineFunctionPointerInParameter tests parsing function declarations
// with inline function pointer type parameters.
// Task 9.44: Inline function pointer types in parameters
func TestInlineFunctionPointerInParameter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedParam string // Expected parameter type string
		wantErr       bool
	}{
		{
			name:          "simple function pointer parameter",
			input:         "procedure Apply(f: function(x: Integer): Integer); begin end;",
			expectedParam: "function(x: Integer): Integer",
			wantErr:       false,
		},
		{
			name:          "procedure pointer parameter",
			input:         "procedure Run(callback: procedure(msg: String)); begin end;",
			expectedParam: "procedure(msg: String)",
			wantErr:       false,
		},
		{
			name:          "function pointer with multiple params",
			input:         "procedure Process(compare: function(a: Integer; b: Integer): Boolean); begin end;",
			expectedParam: "function(a: Integer; b: Integer): Boolean",
			wantErr:       false,
		},
		{
			name:          "function pointer with of object",
			input:         "procedure SetHandler(handler: procedure(Sender: TObject) of object); begin end;",
			expectedParam: "procedure(Sender: TObject) of object",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("expected FunctionDecl, got %T", program.Statements[0])
			}

			if len(funcDecl.Parameters) == 0 {
				t.Fatal("expected at least one parameter")
			}

			param := funcDecl.Parameters[0]
			if param.Type.Name != tt.expectedParam {
				t.Errorf("expected parameter type %q, got %q", tt.expectedParam, param.Type.Name)
			}
		})
	}
}

// TestArrayTypeInParameter tests parsing function declarations
// with array type parameters.
// Task 9.47: array of Type in function parameters
func TestArrayTypeInParameter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedParam string
		wantErr       bool
	}{
		{
			name:          "array of integer parameter",
			input:         "procedure PrintArray(arr: array of Integer); begin end;",
			expectedParam: "array of Integer",
			wantErr:       false,
		},
		{
			name:          "array of string parameter",
			input:         "function JoinStrings(items: array of String): String; begin end;",
			expectedParam: "array of String",
			wantErr:       false,
		},
		{
			name:          "nested array parameter",
			input:         "procedure Process(matrix: array of array of Integer); begin end;",
			expectedParam: "array of array of Integer",
			wantErr:       false,
		},
		{
			name:          "array of function pointers",
			input:         "procedure CallAll(handlers: array of procedure(msg: String)); begin end;",
			expectedParam: "array of procedure(msg: String)",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("expected FunctionDecl, got %T", program.Statements[0])
			}

			if len(funcDecl.Parameters) == 0 {
				t.Fatal("expected at least one parameter")
			}

			param := funcDecl.Parameters[0]
			if param.Type.Name != tt.expectedParam {
				t.Errorf("expected parameter type %q, got %q", tt.expectedParam, param.Type.Name)
			}
		})
	}
}

// TestInlineFunctionPointerInVariable tests parsing variable declarations
// with inline function pointer types.
// Task 9.45: Inline function pointer types in variable declarations
func TestInlineFunctionPointerInVariable(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		wantErr      bool
	}{
		{
			name:         "simple function pointer variable",
			input:        "var f: function(x: Integer): Integer;",
			expectedType: "function(x: Integer): Integer",
			wantErr:      false,
		},
		{
			name:         "procedure pointer variable",
			input:        "var callback: procedure(msg: String);",
			expectedType: "procedure(msg: String)",
			wantErr:      false,
		},
		{
			name:         "function pointer with no params",
			input:        "var getter: function(): String;",
			expectedType: "function(): String",
			wantErr:      false,
		},
		{
			name:         "function pointer of object",
			input:        "var handler: procedure(Sender: TObject) of object;",
			expectedType: "procedure(Sender: TObject) of object",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("expected VarDeclStatement, got %T", program.Statements[0])
			}

			if varDecl.Type == nil {
				t.Fatal("expected type annotation")
			}

			if varDecl.Type.Name != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, varDecl.Type.Name)
			}
		})
	}
}

// TestArrayTypeInVariable tests parsing variable declarations
// with array types.
// Task 9.46: array of Type in variable declarations
func TestArrayTypeInVariable(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		wantErr      bool
	}{
		{
			name:         "array of integer variable",
			input:        "var nums: array of Integer;",
			expectedType: "array of Integer",
			wantErr:      false,
		},
		{
			name:         "array of string variable",
			input:        "var items: array of String;",
			expectedType: "array of String",
			wantErr:      false,
		},
		{
			name:         "nested array variable",
			input:        "var matrix: array of array of Float;",
			expectedType: "array of array of Float",
			wantErr:      false,
		},
		{
			name:         "array of function pointers",
			input:        "var callbacks: array of procedure(msg: String);",
			expectedType: "array of procedure(msg: String)",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Errorf("expected error but got none")
				}
				return
			}

			if len(p.Errors()) > 0 {
				t.Errorf("unexpected errors: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("expected VarDeclStatement, got %T", program.Statements[0])
			}

			if varDecl.Type == nil {
				t.Fatal("expected type annotation")
			}

			if varDecl.Type.Name != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, varDecl.Type.Name)
			}
		})
	}
}
