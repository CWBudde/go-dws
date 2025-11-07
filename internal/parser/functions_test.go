package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestFunctionDeclarations(t *testing.T) {
	tests := []struct {
		expected func(*testing.T, ast.Statement)
		name     string
		input    string
	}{
		{
			name:  "simple function with no parameters",
			input: "function GetValue: Integer; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "GetValue" {
					t.Errorf("function name = %q, want %q", fn.Name.Value, "GetValue")
				}
				if fn.ReturnType == nil || fn.ReturnType.Name != "Integer" {
					t.Errorf("return type = %q, want %q", fn.ReturnType, "Integer")
				}
				if len(fn.Parameters) != 0 {
					t.Errorf("parameters count = %d, want 0", len(fn.Parameters))
				}
			},
		},
		{
			name:  "procedure with no parameters",
			input: "procedure Hello; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "Hello" {
					t.Errorf("function name = %q, want %q", fn.Name.Value, "Hello")
				}
				if fn.ReturnType != nil {
					t.Errorf("return type = %q, want empty string for procedure", fn.ReturnType)
				}
			},
		},
		{
			name:  "function with single parameter",
			input: "function Double(x: Integer): Integer; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "Double" {
					t.Errorf("function name = %q, want %q", fn.Name.Value, "Double")
				}
				if len(fn.Parameters) != 1 {
					t.Fatalf("parameters count = %d, want 1", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "x" {
					t.Errorf("parameter name = %q, want %q", param.Name.Value, "x")
				}
				if param.Type == nil || param.Type.Name != "Integer" {
					t.Errorf("parameter type = %q, want %q", param.Type, "Integer")
				}
				if param.ByRef {
					t.Errorf("parameter ByRef = true, want false")
				}
			},
		},
		{
			name:  "function with multiple parameters",
			input: "function Add(a: Integer; b: Integer): Integer; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if len(fn.Parameters) != 2 {
					t.Fatalf("parameters count = %d, want 2", len(fn.Parameters))
				}
				if fn.Parameters[0].Name.Value != "a" {
					t.Errorf("first param name = %q, want %q", fn.Parameters[0].Name.Value, "a")
				}
				if fn.Parameters[1].Name.Value != "b" {
					t.Errorf("second param name = %q, want %q", fn.Parameters[1].Name.Value, "b")
				}
			},
		},
		{
			name:  "function with var parameter",
			input: "function Process(var data: String): Boolean; begin end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if len(fn.Parameters) != 1 {
					t.Fatalf("parameters count = %d, want 1", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.ByRef {
					t.Errorf("parameter ByRef = false, want true")
				}
				if param.Name.Value != "data" {
					t.Errorf("parameter name = %q, want %q", param.Name.Value, "data")
				}
				if param.Type == nil || param.Type.Name != "String" {
					t.Errorf("parameter type = %q, want %q", param.Type, "String")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			tt.expected(t, program.Statements[0])
		})
	}
}

// TestParameters tests parameter parsing in function declarations - Task 5.14
func TestParameters(t *testing.T) {
	tests := []struct {
		expected func(*testing.T, *ast.FunctionDecl)
		name     string
		input    string
	}{
		{
			name:  "single parameter",
			input: "function Test(x: Integer): Boolean; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "x" {
					t.Errorf("param name = %q, want 'x'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "Integer" {
					t.Errorf("param type = %q, want 'Integer'", param.Type)
				}
				if param.ByRef {
					t.Error("param should not be by reference")
				}
			},
		},
		{
			name:  "multiple parameters with different types",
			input: "function Calculate(x: Integer; y: Float; name: String): Float; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// Check first parameter
				if fn.Parameters[0].Name.Value != "x" {
					t.Errorf("param[0] name = %q, want 'x'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[0].Type == nil || fn.Parameters[0].Type.Name != "Integer" {
					t.Errorf("param[0] type = %q, want 'Integer'", fn.Parameters[0].Type)
				}

				// Check second parameter
				if fn.Parameters[1].Name.Value != "y" {
					t.Errorf("param[1] name = %q, want 'y'", fn.Parameters[1].Name.Value)
				}
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "Float" {
					t.Errorf("param[1] type = %q, want 'Float'", fn.Parameters[1].Type)
				}

				// Check third parameter
				if fn.Parameters[2].Name.Value != "name" {
					t.Errorf("param[2] name = %q, want 'name'", fn.Parameters[2].Name.Value)
				}
				if fn.Parameters[2].Type == nil || fn.Parameters[2].Type.Name != "String" {
					t.Errorf("param[2] type = %q, want 'String'", fn.Parameters[2].Type)
				}
			},
		},
		{
			name:  "var parameter by reference",
			input: "procedure Swap(var a: Integer; var b: Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// Both parameters should be by reference
				if !fn.Parameters[0].ByRef {
					t.Error("param[0] should be by reference")
				}
				if !fn.Parameters[1].ByRef {
					t.Error("param[1] should be by reference")
				}

				if fn.Parameters[0].Name.Value != "a" {
					t.Errorf("param[0] name = %q, want 'a'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[1].Name.Value != "b" {
					t.Errorf("param[1] name = %q, want 'b'", fn.Parameters[1].Name.Value)
				}
			},
		},
		{
			name:  "mixed var and value parameters",
			input: "procedure Update(var x: Integer; y: Integer; var z: String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// Check ByRef flags
				if !fn.Parameters[0].ByRef {
					t.Error("param[0] should be by reference")
				}
				if fn.Parameters[1].ByRef {
					t.Error("param[1] should not be by reference")
				}
				if !fn.Parameters[2].ByRef {
					t.Error("param[2] should be by reference")
				}
			},
		},
		{
			name:  "function with no parameters",
			input: "function GetRandom: Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 0 {
					t.Fatalf("expected 0 parameters, got %d", len(fn.Parameters))
				}
			},
		},
		{
			name:  "procedure with no parameters",
			input: "procedure PrintHello; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 0 {
					t.Fatalf("expected 0 parameters, got %d", len(fn.Parameters))
				}
			},
		},
		{
			name:  "lazy parameter - basic",
			input: "function Test(lazy x: Integer): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.IsLazy {
					t.Error("param should be lazy")
				}
				if param.ByRef {
					t.Error("param should not be by reference")
				}
				if param.Name.Value != "x" {
					t.Errorf("param name = %q, want 'x'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "Integer" {
					t.Errorf("param type = %q, want 'Integer'", param.Type)
				}
			},
		},
		{
			name:  "lazy parameter - mixed with regular parameters",
			input: "procedure Log(level: Integer; lazy msg: String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// First parameter should be regular (not lazy, not by reference)
				if fn.Parameters[0].IsLazy {
					t.Error("param[0] should not be lazy")
				}
				if fn.Parameters[0].ByRef {
					t.Error("param[0] should not be by reference")
				}
				if fn.Parameters[0].Name.Value != "level" {
					t.Errorf("param[0] name = %q, want 'level'", fn.Parameters[0].Name.Value)
				}

				// Second parameter should be lazy
				if !fn.Parameters[1].IsLazy {
					t.Error("param[1] should be lazy")
				}
				if fn.Parameters[1].ByRef {
					t.Error("param[1] should not be by reference")
				}
				if fn.Parameters[1].Name.Value != "msg" {
					t.Errorf("param[1] name = %q, want 'msg'", fn.Parameters[1].Name.Value)
				}
			},
		},
		{
			name:  "lazy parameter - multiple lazy parameters with shared type",
			input: "function If(cond: Boolean; lazy trueVal, falseVal: Integer): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// First parameter (cond) should be regular
				if fn.Parameters[0].IsLazy {
					t.Error("param[0] should not be lazy")
				}
				if fn.Parameters[0].Name.Value != "cond" {
					t.Errorf("param[0] name = %q, want 'cond'", fn.Parameters[0].Name.Value)
				}

				// Second parameter (trueVal) should be lazy
				if !fn.Parameters[1].IsLazy {
					t.Error("param[1] should be lazy")
				}
				if fn.Parameters[1].Name.Value != "trueVal" {
					t.Errorf("param[1] name = %q, want 'trueVal'", fn.Parameters[1].Name.Value)
				}

				// Third parameter (falseVal) should be lazy (shares type with trueVal)
				if !fn.Parameters[2].IsLazy {
					t.Error("param[2] should be lazy")
				}
				if fn.Parameters[2].Name.Value != "falseVal" {
					t.Errorf("param[2] name = %q, want 'falseVal'", fn.Parameters[2].Name.Value)
				}

				// Both lazy parameters should have Integer type
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "Integer" {
					t.Errorf("param[1] type = %q, want 'Integer'", fn.Parameters[1].Type)
				}
				if fn.Parameters[2].Type == nil || fn.Parameters[2].Type.Name != "Integer" {
					t.Errorf("param[2] type = %q, want 'Integer'", fn.Parameters[2].Type)
				}
			},
		},
		{
			name:  "const parameter - basic",
			input: "procedure Process(const data: array of Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.IsConst {
					t.Error("param should be const")
				}
				if param.ByRef {
					t.Error("param should not be by reference")
				}
				if param.IsLazy {
					t.Error("param should not be lazy")
				}
				if param.Name.Value != "data" {
					t.Errorf("param name = %q, want 'data'", param.Name.Value)
				}
			},
		},
		{
			name:  "const parameter - with string type",
			input: "procedure Display(const message: String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if !param.IsConst {
					t.Error("param should be const")
				}
				if param.Name.Value != "message" {
					t.Errorf("param name = %q, want 'message'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "String" {
					t.Errorf("param type = %q, want 'String'", param.Type)
				}
			},
		},
		{
			name:  "const parameter - mixed with var and regular parameters",
			input: "procedure Update(const src: String; var dest: String; count: Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(fn.Parameters))
				}

				// First parameter should be const
				if !fn.Parameters[0].IsConst {
					t.Error("param[0] should be const")
				}
				if fn.Parameters[0].ByRef {
					t.Error("param[0] should not be by reference")
				}
				if fn.Parameters[0].Name.Value != "src" {
					t.Errorf("param[0] name = %q, want 'src'", fn.Parameters[0].Name.Value)
				}

				// Second parameter should be var (by reference)
				if fn.Parameters[1].IsConst {
					t.Error("param[1] should not be const")
				}
				if !fn.Parameters[1].ByRef {
					t.Error("param[1] should be by reference")
				}
				if fn.Parameters[1].Name.Value != "dest" {
					t.Errorf("param[1] name = %q, want 'dest'", fn.Parameters[1].Name.Value)
				}

				// Third parameter should be regular (not const, not by reference)
				if fn.Parameters[2].IsConst {
					t.Error("param[2] should not be const")
				}
				if fn.Parameters[2].ByRef {
					t.Error("param[2] should not be by reference")
				}
				if fn.Parameters[2].Name.Value != "count" {
					t.Errorf("param[2] name = %q, want 'count'", fn.Parameters[2].Name.Value)
				}
			},
		},
		{
			name:  "const parameter - multiple const parameters with shared type",
			input: "procedure Compare(const a, b: Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// Both parameters should be const
				if !fn.Parameters[0].IsConst {
					t.Error("param[0] should be const")
				}
				if !fn.Parameters[1].IsConst {
					t.Error("param[1] should be const")
				}

				if fn.Parameters[0].Name.Value != "a" {
					t.Errorf("param[0] name = %q, want 'a'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[1].Name.Value != "b" {
					t.Errorf("param[1] name = %q, want 'b'", fn.Parameters[1].Name.Value)
				}

				// Both should have Integer type
				if fn.Parameters[0].Type == nil || fn.Parameters[0].Type.Name != "Integer" {
					t.Errorf("param[0] type = %q, want 'Integer'", fn.Parameters[0].Type)
				}
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "Integer" {
					t.Errorf("param[1] type = %q, want 'Integer'", fn.Parameters[1].Type)
				}
			},
		},
		{
			name:  "variadic parameter: array of const",
			input: "procedure Test(const a: array of const); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "a" {
					t.Errorf("param name = %q, want 'a'", param.Name.Value)
				}
				if !param.IsConst {
					t.Error("param should be const")
				}
				// Type should be "array of const" (parsed as array type with const element)
				if param.Type == nil {
					t.Fatal("param type is nil")
				}
				// The type name should contain "array" since it's a synthetic TypeAnnotation
				// from ArrayTypeNode.String()
				if param.Type.Name != "array of const" {
					t.Errorf("param type = %q, want 'array of const'", param.Type.Name)
				}
			},
		},
		{
			name:  "variadic parameter: array of Integer",
			input: "procedure ProcessValues(const values: array of Integer); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "values" {
					t.Errorf("param name = %q, want 'values'", param.Name.Value)
				}
				if !param.IsConst {
					t.Error("param should be const")
				}
				if param.Type == nil {
					t.Fatal("param type is nil")
				}
				if param.Type.Name != "array of Integer" {
					t.Errorf("param type = %q, want 'array of Integer'", param.Type.Name)
				}
			},
		},
		{
			name:  "mixed fixed and variadic parameters",
			input: "function Format(fmt: String; const args: array of const): String; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("expected 2 parameters, got %d", len(fn.Parameters))
				}

				// First parameter: fixed
				if fn.Parameters[0].Name.Value != "fmt" {
					t.Errorf("param[0] name = %q, want 'fmt'", fn.Parameters[0].Name.Value)
				}
				if fn.Parameters[0].Type == nil || fn.Parameters[0].Type.Name != "String" {
					t.Errorf("param[0] type = %q, want 'String'", fn.Parameters[0].Type)
				}
				if fn.Parameters[0].IsConst {
					t.Error("param[0] should not be const")
				}

				// Second parameter: variadic
				if fn.Parameters[1].Name.Value != "args" {
					t.Errorf("param[1] name = %q, want 'args'", fn.Parameters[1].Name.Value)
				}
				if !fn.Parameters[1].IsConst {
					t.Error("param[1] should be const")
				}
				if fn.Parameters[1].Type == nil || fn.Parameters[1].Type.Name != "array of const" {
					t.Errorf("param[1] type = %q, want 'array of const'", fn.Parameters[1].Type)
				}
			},
		},
		{
			name:  "variadic parameter: array of String",
			input: "procedure PrintAll(const items: array of String); begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "items" {
					t.Errorf("param name = %q, want 'items'", param.Name.Value)
				}
				if param.Type == nil || param.Type.Name != "array of String" {
					t.Errorf("param type = %q, want 'array of String'", param.Type.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			fn, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
			}

			tt.expected(t, fn)
		})
	}
}

// TestParameterErrors tests error cases for parameter parsing
func TestParameterErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
	}{
		{
			name:          "lazy and var are mutually exclusive",
			input:         "function Test(lazy var x: Integer): Integer; begin end;",
			errorContains: "parameter modifiers are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			_ = p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Fatalf("expected parser error, got none")
			}

			// Check that error message contains expected text
			found := false
			for _, err := range p.Errors() {
				if strings.Contains(err, tt.errorContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got %v", tt.errorContains, p.Errors())
			}
		})
	}
}

// TestUserDefinedFunctionCallsWithArguments tests calling user-defined functions with arguments - Task 5.15
func TestUserDefinedFunctionCallsWithArguments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "call user function with integer arguments",
			input: `
				function Add(a: Integer; b: Integer): Integer;
				begin
					end;

				begin
					Add(1, 2);
				end;
			`,
		},
		{
			name: "call user function with mixed argument types",
			input: `
				function Format(name: String; age: Integer): String;
				begin
				end;

				begin
					Format('John', 25);
				end;
			`,
		},
		{
			name: "call user function with expression arguments",
			input: `
				function Calculate(x: Integer; y: Integer): Integer;
				begin
				end;

				begin
					Calculate(1 + 2, 3 * 4);
				end;
			`,
		},
		{
			name: "call user function with no arguments",
			input: `
				function GetValue: Integer;
				begin
				end;

				begin
					GetValue();
				end;
			`,
		},
		{
			name: "call procedure with arguments",
			input: `
				procedure PrintValue(x: Integer);
				begin
				end;

				begin
					PrintValue(42);
				end;
			`,
		},
		{
			name: "nested function calls as arguments",
			input: `
				function Double(x: Integer): Integer;
				begin
				end;

				function Triple(x: Integer): Integer;
				begin
				end;

				begin
					Double(Triple(5));
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			// The program should parse successfully
			if len(program.Statements) < 2 {
				t.Fatalf("expected at least 2 statements (function + main block), got %d", len(program.Statements))
			}

			// First statement(s) should be function declarations
			for i := 0; i < len(program.Statements)-1; i++ {
				if _, ok := program.Statements[i].(*ast.FunctionDecl); !ok {
					t.Errorf("statement %d is not *ast.FunctionDecl, got %T", i, program.Statements[i])
				}
			}

			// Last statement should be the main block containing the call
			lastStmt := program.Statements[len(program.Statements)-1]
			if _, ok := lastStmt.(*ast.BlockStatement); !ok {
				t.Errorf("last statement is not *ast.BlockStatement, got %T", lastStmt)
			}
		})
	}
}

// TestNestedFunctions tests nested function declarations - Task 5.16
// Note: DWScript may or may not support nested functions. This test documents current behavior.
func TestNestedFunctions(t *testing.T) {
	input := `
		function Outer(x: Integer): Integer;
		begin
			function Inner(y: Integer): Integer;
			begin
			end;
		end;
	`

	p := testParser(input)
	program := p.ParseProgram()

	// Check if parser supports nested functions
	// If there are parser errors, nested functions are not yet supported
	errors := p.Errors()
	if len(errors) > 0 {
		t.Skip("Nested functions not yet supported - this is expected per PLAN.md task 5.11")
		return
	}

	// If we get here, nested functions ARE supported
	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	outerFn, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
	}

	if outerFn.Name.Value != "Outer" {
		t.Errorf("outer function name = %q, want 'Outer'", outerFn.Name.Value)
	}

	// Check if the body contains the nested function
	// This would require the AST to support nested function declarations
	// For now, we just verify the outer function parses correctly
	if outerFn.Body == nil {
		t.Error("outer function body is nil")
	}
}

// TestNewKeywordExpression tests parsing of 'new' keyword expressions
// The 'new' keyword creates a NewExpression: new T(args) -> NewExpression{ClassName: T, Arguments: args}
func TestNewKeywordExpression(t *testing.T) {
	tests := []struct {
		input    string
		typeName string
		numArgs  int
	}{
		{"new Exception('test');", "Exception", 1},
		{"new TPoint(10, 20);", "TPoint", 2},
		{"new TMyClass();", "TMyClass", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
			}

			// new T(args) should create a NewExpression
			newExpr, ok := stmt.Expression.(*ast.NewExpression)
			if !ok {
				t.Fatalf("expression is not ast.NewExpression. got=%T", stmt.Expression)
			}

			// Check class name
			if newExpr.ClassName.Value != tt.typeName {
				t.Fatalf("wrong class name. expected=%s, got=%s", tt.typeName, newExpr.ClassName.Value)
			}

			// Check number of arguments
			if len(newExpr.Arguments) != tt.numArgs {
				t.Fatalf("wrong number of arguments. expected=%d, got=%d", tt.numArgs, len(newExpr.Arguments))
			}
		})
	}
}

// TestContextualKeywordStep tests that 'step' can be used both as a keyword in for loops
// and as a variable name in other contexts.
func TestContextualKeywordStep(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldParse bool
	}{
		{
			name: "step as variable name in declaration",
			input: `
				var step: Integer;
			`,
			shouldParse: true,
		},
		{
			name: "step as variable name with initialization",
			input: `
				var step := 0;
			`,
			shouldParse: true,
		},
		{
			name: "step in assignment statement",
			input: `
				var step := 1;
				step := 2;
			`,
			shouldParse: true,
		},
		{
			name: "step in expression",
			input: `
				var step := 1;
				var result := step + 5;
			`,
			shouldParse: true,
		},
		{
			name: "step as keyword in for loop",
			input: `
				for i := 1 to 10 step 2 do
					PrintLn(i);
			`,
			shouldParse: true,
		},
		{
			name: "step used in both contexts",
			input: `
				var step := 2;
				for i := 1 to 10 step step do
					PrintLn(i);
			`,
			shouldParse: true,
		},
		{
			name: "step in function parameter",
			input: `
				function Process(step: Integer): Integer;
				begin
					Result := step * 2;
				end;
			`,
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.shouldParse {
				checkParserErrors(t, p)
				if program == nil {
					t.Fatal("ParseProgram() returned nil")
				}
				if len(program.Statements) == 0 {
					t.Fatal("ParseProgram() returned empty statements")
				}
			} else {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected parser errors, but got none")
				}
			}
		})
	}
}

func TestMethodDeclarations(t *testing.T) {
	tests := []struct {
		expected func(*testing.T, ast.Statement)
		name     string
		input    string
	}{
		{
			name:  "method implementation with qualified name",
			input: "method TMyClass.DoSomething : String; begin Result := 'test'; end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "DoSomething" {
					t.Errorf("method name = %q, want %q", fn.Name.Value, "DoSomething")
				}
				if fn.ClassName == nil || fn.ClassName.Value != "TMyClass" {
					className := ""
					if fn.ClassName != nil {
						className = fn.ClassName.Value
					}
					t.Errorf("class name = %q, want %q", className, "TMyClass")
				}
				if fn.ReturnType == nil || fn.ReturnType.Name != "String" {
					t.Errorf("return type = %v, want %q", fn.ReturnType, "String")
				}
			},
		},
		{
			name:  "method implementation with parameters",
			input: "method TBottlesSinger.Sing : String; begin PrepareLine; Result := Line; end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "Sing" {
					t.Errorf("method name = %q, want %q", fn.Name.Value, "Sing")
				}
				if fn.ClassName == nil || fn.ClassName.Value != "TBottlesSinger" {
					className := ""
					if fn.ClassName != nil {
						className = fn.ClassName.Value
					}
					t.Errorf("class name = %q, want %q", className, "TBottlesSinger")
				}
				if fn.ReturnType == nil || fn.ReturnType.Name != "String" {
					t.Errorf("return type = %v, want %q", fn.ReturnType, "String")
				}
			},
		},
		{
			name:  "method procedure implementation",
			input: "method TMyClass.Initialize; begin Value := 0; end;",
			expected: func(t *testing.T, stmt ast.Statement) {
				fn, ok := stmt.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("stmt is not *ast.FunctionDecl. got=%T", stmt)
				}
				if fn.Name.Value != "Initialize" {
					t.Errorf("method name = %q, want %q", fn.Name.Value, "Initialize")
				}
				if fn.ClassName == nil || fn.ClassName.Value != "TMyClass" {
					className := ""
					if fn.ClassName != nil {
						className = fn.ClassName.Value
					}
					t.Errorf("class name = %q, want %q", className, "TMyClass")
				}
				if fn.ReturnType != nil {
					t.Errorf("return type = %v, want nil for procedure", fn.ReturnType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			tt.expected(t, program.Statements[0])
		})
	}
}

// TestOverloadDirective tests parsing of the overload directive on functions
func TestOverloadDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*testing.T, *ast.FunctionDecl)
	}{
		{
			name:  "function with overload directive",
			input: "function Test(x: Integer): Float; overload; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if fn.Name.Value != "Test" {
					t.Errorf("function name = %q, want 'Test'", fn.Name.Value)
				}
				if !fn.IsOverload {
					t.Error("fn.IsOverload should be true")
				}
				if fn.IsVirtual || fn.IsOverride || fn.IsAbstract {
					t.Error("no other directives should be set")
				}
			},
		},
		{
			name:  "procedure with overload directive",
			input: "procedure Print(msg: String); overload; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if fn.Name.Value != "Print" {
					t.Errorf("function name = %q, want 'Print'", fn.Name.Value)
				}
				if !fn.IsOverload {
					t.Error("fn.IsOverload should be true")
				}
				if fn.ReturnType != nil {
					t.Errorf("procedure should have no return type, got %v", fn.ReturnType)
				}
			},
		},
		{
			name:  "function with virtual and overload directives",
			input: "function DoWork(): Integer; virtual; overload; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if fn.Name.Value != "DoWork" {
					t.Errorf("function name = %q, want 'DoWork'", fn.Name.Value)
				}
				if !fn.IsVirtual {
					t.Error("fn.IsVirtual should be true")
				}
				if !fn.IsOverload {
					t.Error("fn.IsOverload should be true")
				}
				if fn.IsOverride || fn.IsAbstract {
					t.Error("IsOverride and IsAbstract should be false")
				}
			},
		},
		{
			name:  "function with overload only (no body - forward declaration)",
			input: "function Helper(s: String): Boolean; overload;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if fn.Name.Value != "Helper" {
					t.Errorf("function name = %q, want 'Helper'", fn.Name.Value)
				}
				if !fn.IsOverload {
					t.Error("fn.IsOverload should be true")
				}
				if fn.Body != nil {
					t.Error("forward declaration should have no body")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			fn, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
			}

			tt.expected(t, fn)
		})
	}
}

func TestOptionalParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*testing.T, *ast.FunctionDecl)
	}{
		{
			name:  "function with single optional parameter",
			input: "function Greet(name: String = 'World'): String; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if fn.Name.Value != "Greet" {
					t.Errorf("function name = %q, want 'Greet'", fn.Name.Value)
				}
				if len(fn.Parameters) != 1 {
					t.Fatalf("parameters count = %d, want 1", len(fn.Parameters))
				}
				param := fn.Parameters[0]
				if param.Name.Value != "name" {
					t.Errorf("parameter name = %q, want 'name'", param.Name.Value)
				}
				if param.Type.Name != "String" {
					t.Errorf("parameter type = %q, want 'String'", param.Type.Name)
				}
				if param.DefaultValue == nil {
					t.Fatal("expected default value, got nil")
				}
				// Check default value is a string literal
				strLit, ok := param.DefaultValue.(*ast.StringLiteral)
				if !ok {
					t.Fatalf("default value is not *ast.StringLiteral, got %T", param.DefaultValue)
				}
				if strLit.Value != "World" {
					t.Errorf("default value = %q, want 'World'", strLit.Value)
				}
			},
		},
		{
			name:  "function with required and optional parameters",
			input: "function Add(a: Integer; b: Integer = 0): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("parameters count = %d, want 2", len(fn.Parameters))
				}
				// First parameter should be required
				if fn.Parameters[0].DefaultValue != nil {
					t.Error("first parameter should be required (no default value)")
				}
				// Second parameter should be optional
				if fn.Parameters[1].DefaultValue == nil {
					t.Fatal("second parameter should have default value")
				}
				intLit, ok := fn.Parameters[1].DefaultValue.(*ast.IntegerLiteral)
				if !ok {
					t.Fatalf("default value is not *ast.IntegerLiteral, got %T", fn.Parameters[1].DefaultValue)
				}
				if intLit.Value != 0 {
					t.Errorf("default value = %d, want 0", intLit.Value)
				}
			},
		},
		{
			name:  "function with multiple optional parameters",
			input: "function Format(text: String; prefix: String = '['; suffix: String = ']'): String; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 3 {
					t.Fatalf("parameters count = %d, want 3", len(fn.Parameters))
				}
				// First parameter is required
				if fn.Parameters[0].DefaultValue != nil {
					t.Error("first parameter should be required")
				}
				// Second and third are optional
				if fn.Parameters[1].DefaultValue == nil {
					t.Error("second parameter should have default value")
				}
				if fn.Parameters[2].DefaultValue == nil {
					t.Error("third parameter should have default value")
				}
			},
		},
		{
			name:  "function with numeric default value",
			input: "function Power(base: Float; exponent: Float = 2.0): Float; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("parameters count = %d, want 2", len(fn.Parameters))
				}
				param := fn.Parameters[1]
				if param.DefaultValue == nil {
					t.Fatal("expected default value")
				}
				floatLit, ok := param.DefaultValue.(*ast.FloatLiteral)
				if !ok {
					t.Fatalf("default value is not *ast.FloatLiteral, got %T", param.DefaultValue)
				}
				if floatLit.Value != 2.0 {
					t.Errorf("default value = %f, want 2.0", floatLit.Value)
				}
			},
		},
		{
			name:  "function with expression as default value",
			input: "function Calculate(x: Integer; multiplier: Integer = 2 * 3): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("parameters count = %d, want 2", len(fn.Parameters))
				}
				param := fn.Parameters[1]
				if param.DefaultValue == nil {
					t.Fatal("expected default value")
				}
				// Default value should be a binary expression (2 * 3)
				binExpr, ok := param.DefaultValue.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("default value is not *ast.BinaryExpression, got %T", param.DefaultValue)
				}
				if binExpr.Operator != "*" {
					t.Errorf("operator = %q, want '*'", binExpr.Operator)
				}
			},
		},
		{
			name:  "comma-separated parameters with default value",
			input: "function Test(a, b: Integer = 5): Integer; begin end;",
			expected: func(t *testing.T, fn *ast.FunctionDecl) {
				if len(fn.Parameters) != 2 {
					t.Fatalf("parameters count = %d, want 2", len(fn.Parameters))
				}
				// Both parameters should have the same default value
				if fn.Parameters[0].DefaultValue == nil {
					t.Error("first parameter should have default value")
				}
				if fn.Parameters[1].DefaultValue == nil {
					t.Error("second parameter should have default value")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			fn, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
			}

			tt.expected(t, fn)
		})
	}
}

func TestOptionalParametersErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "var parameter with default value",
			input:         "function Test(var x: Integer = 5): Integer; begin end;",
			expectedError: "optional parameters cannot have lazy, var, or const modifiers",
		},
		{
			name:          "lazy parameter with default value",
			input:         "function Test(lazy x: Integer = 5): Integer; begin end;",
			expectedError: "optional parameters cannot have lazy, var, or const modifiers",
		},
		{
			name:          "const parameter with default value",
			input:         "function Test(const x: Integer = 5): Integer; begin end;",
			expectedError: "optional parameters cannot have lazy, var, or const modifiers",
		},
		{
			name:          "empty default value",
			input:         "function Test(x: Integer = ): Integer; begin end;",
			expectedError: "expected default value expression after '='",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			p.ParseProgram()

			if len(p.errors) == 0 {
				t.Fatal("expected parser error, got none")
			}

			found := false
			for _, err := range p.errors {
				if strings.Contains(err, tt.expectedError) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got errors: %v", tt.expectedError, p.errors)
			}
		})
	}
}
