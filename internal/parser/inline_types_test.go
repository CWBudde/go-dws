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

// TestArrayReturnTypes tests parsing function declarations with inline array return types.
// Task 9.59: Support inline array types in function return types
func TestArrayReturnTypes(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedReturn string // Expected return type string
		wantErr        bool
	}{
		{
			name:           "dynamic array return type",
			input:          "function GetData(): array of Integer; begin end;",
			expectedReturn: "array of Integer",
			wantErr:        false,
		},
		{
			name:           "static array return type",
			input:          "function CreateArray(): array[1..10] of Integer; begin end;",
			expectedReturn: "array[1..10] of Integer",
			wantErr:        false,
		},
		{
			name:           "nested dynamic array return type",
			input:          "function GetMatrix(): array of array of Float; begin end;",
			expectedReturn: "array of array of Float",
			wantErr:        false,
		},
		{
			name:           "nested static array return type",
			input:          "function CreateMatrix(): array[1..5] of array[1..10] of Integer; begin end;",
			expectedReturn: "array[1..5] of array[1..10] of Integer",
			wantErr:        false,
		},
		{
			name:           "array of strings return type",
			input:          "function GetNames(): array of String; begin end;",
			expectedReturn: "array of String",
			wantErr:        false,
		},
		{
			name:           "zero-based static array return type",
			input:          "function GetBytes(): array[0..255] of Integer; begin end;",
			expectedReturn: "array[0..255] of Integer",
			wantErr:        false,
		},
		{
			name:           "negative bounds array return type",
			input:          "function GetCentered(): array[(-10)..10] of Integer; begin end;",
			expectedReturn: "array[(-10)..10] of Integer",
			wantErr:        false,
		},
		{
			name:           "array of booleans return type",
			input:          "function GetFlags(): array[1..8] of Boolean; begin end;",
			expectedReturn: "array[1..8] of Boolean",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Error("expected parser errors, got none")
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

			fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("expected FunctionDecl, got %T", program.Statements[0])
			}

			if fnDecl.ReturnType == nil {
				t.Fatal("expected return type annotation")
			}

			if fnDecl.ReturnType.Name != tt.expectedReturn {
				t.Errorf("expected return type %q, got %q", tt.expectedReturn, fnDecl.ReturnType.Name)
			}
		})
	}
}

// TestComplexReturnTypes tests parsing function declarations with complex return types
// combining arrays and function pointers.
// Task 9.59: Support complex inline types in function return types
func TestComplexReturnTypes(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedReturn string // Expected return type string
		wantErr        bool
	}{
		{
			name:           "function pointer return type",
			input:          "function GetProcessor(): function(x: Integer): Integer; begin end;",
			expectedReturn: "function(x: Integer): Integer",
			wantErr:        false,
		},
		{
			name:           "procedure pointer return type",
			input:          "function GetCallback(): procedure(msg: String); begin end;",
			expectedReturn: "procedure(msg: String)",
			wantErr:        false,
		},
		{
			name:           "simple type return (baseline)",
			input:          "function GetNumber(): Integer; begin end;",
			expectedReturn: "Integer",
			wantErr:        false,
		},
		{
			name:           "string type return (baseline)",
			input:          "function GetText(): String; begin end;",
			expectedReturn: "String",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if tt.wantErr {
				if len(p.Errors()) == 0 {
					t.Error("expected parser errors, got none")
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

			fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
			if !ok {
				t.Fatalf("expected FunctionDecl, got %T", program.Statements[0])
			}

			if fnDecl.ReturnType == nil {
				t.Fatal("expected return type annotation")
			}

			if fnDecl.ReturnType.Name != tt.expectedReturn {
				t.Errorf("expected return type %q, got %q", tt.expectedReturn, fnDecl.ReturnType.Name)
			}
		})
	}
}

// TestMetaclassTypeDeclaration tests parsing metaclass type declarations.
// Task 9.73.4: Parser support for metaclass type aliases
func TestMetaclassTypeDeclaration(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedTypeName string
		expectedBaseType string
		wantErr          bool
	}{
		{
			name:             "simple metaclass type",
			input:            "type TBaseClass = class of TBase;",
			expectedTypeName: "TBaseClass",
			expectedBaseType: "TBase",
			wantErr:          false,
		},
		{
			name:             "metaclass of TObject",
			input:            "type TObjectClass = class of TObject;",
			expectedTypeName: "TObjectClass",
			expectedBaseType: "TObject",
			wantErr:          false,
		},
		{
			name:             "metaclass with longer name",
			input:            "type TMyControlClass = class of TMyControl;",
			expectedTypeName: "TMyControlClass",
			expectedBaseType: "TMyControl",
			wantErr:          false,
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

			typeDecl, ok := program.Statements[0].(*ast.TypeDeclaration)
			if !ok {
				t.Fatalf("expected TypeDeclaration, got %T", program.Statements[0])
			}

			if typeDecl.Name.Value != tt.expectedTypeName {
				t.Errorf("expected type name %q, got %q", tt.expectedTypeName, typeDecl.Name.Value)
			}

			if !typeDecl.IsAlias {
				t.Error("expected IsAlias to be true")
			}

			if typeDecl.AliasedType == nil {
				t.Fatal("expected AliasedType to be set")
			}

			if typeDecl.AliasedType.InlineType == nil {
				t.Fatal("expected InlineType to be set")
			}

			classOfNode, ok := typeDecl.AliasedType.InlineType.(*ast.ClassOfTypeNode)
			if !ok {
				t.Fatalf("expected ClassOfTypeNode, got %T", typeDecl.AliasedType.InlineType)
			}

			// Check the base class type
			classType, ok := classOfNode.ClassType.(*ast.TypeAnnotation)
			if !ok {
				t.Fatalf("expected TypeAnnotation for class type, got %T", classOfNode.ClassType)
			}

			if classType.Name != tt.expectedBaseType {
				t.Errorf("expected base type %q, got %q", tt.expectedBaseType, classType.Name)
			}
		})
	}
}

// TestMetaclassVariableDeclaration tests parsing variable declarations with metaclass types.
// Task 9.73.4: Parser support for metaclass variables
func TestMetaclassVariableDeclaration(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedType string
		wantErr      bool
	}{
		{
			name:         "metaclass variable",
			input:        "var meta: class of TBase;",
			expectedName: "meta",
			expectedType: "class of TBase",
			wantErr:      false,
		},
		{
			name:         "metaclass variable with longer name",
			input:        "var myClass: class of TMyControl;",
			expectedName: "myClass",
			expectedType: "class of TMyControl",
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
				t.Fatalf("expected VarDeclaration, got %T", program.Statements[0])
			}

			if len(varDecl.Names) != 1 {
				t.Fatalf("expected 1 variable name, got %d", len(varDecl.Names))
			}

			if varDecl.Names[0].Value != tt.expectedName {
				t.Errorf("expected variable name %q, got %q", tt.expectedName, varDecl.Names[0].Value)
			}

			if varDecl.Type.Name != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, varDecl.Type.Name)
			}

			// Verify InlineType is ClassOfTypeNode
			if varDecl.Type.InlineType == nil {
				t.Fatal("expected InlineType to be set")
			}

			_, ok = varDecl.Type.InlineType.(*ast.ClassOfTypeNode)
			if !ok {
				t.Fatalf("expected ClassOfTypeNode, got %T", varDecl.Type.InlineType)
			}
		})
	}
}

// TestEnumIndexedArrayType tests parsing enum-indexed arrays.
// Task 9.21.1: Support enum-indexed arrays
func TestEnumIndexedArrayType(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedEnum string
		expectedElem string
		wantErr      bool
	}{
		{
			name:         "enum-indexed array in var",
			input:        "type TDay = (Mon, Tue, Wed); var schedule: array[TDay] of Integer;",
			expectedType: "array[TDay] of Integer",
			expectedEnum: "TDay",
			expectedElem: "Integer",
			wantErr:      false,
		},
		{
			name:         "enum-indexed array with different enum",
			input:        "type TColor = (Red, Green, Blue); var colors: array[TColor] of String;",
			expectedType: "array[TColor] of String",
			expectedEnum: "TColor",
			expectedElem: "String",
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

			// Should have at least 2 statements (type decl + var)
			if len(program.Statements) < 2 {
				t.Fatalf("expected at least 2 statements, got %d", len(program.Statements))
			}

			// Get the second statement (var declaration)
			varDecl, ok := program.Statements[1].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("expected VarDeclStatement, got %T", program.Statements[1])
			}

			if varDecl.Type == nil || varDecl.Type.InlineType == nil {
				t.Fatal("expected type annotation with inline type")
			}

			arrayType, ok := varDecl.Type.InlineType.(*ast.ArrayTypeNode)
			if !ok {
				t.Fatalf("expected ArrayTypeNode, got %T", varDecl.Type.InlineType)
			}

			// Verify the array type
			if arrayType == nil {
				t.Fatal("arrayType is nil")
			}

			// Check that it's recognized as enum-indexed
			if !arrayType.IsEnumIndexed() {
				t.Error("expected IsEnumIndexed() to return true")
			}

			// Check that it's recognized as static (enum-indexed arrays are static)
			if !arrayType.IsStatic() {
				t.Error("expected IsStatic() to return true")
			}

			// Verify IndexType is set
			if arrayType.IndexType == nil {
				t.Fatal("expected IndexType to be set")
			}

			indexTypeAnnot, ok := arrayType.IndexType.(*ast.TypeAnnotation)
			if !ok {
				t.Fatalf("expected TypeAnnotation for IndexType, got %T", arrayType.IndexType)
			}

			if indexTypeAnnot.Name != tt.expectedEnum {
				t.Errorf("expected enum type %q, got %q", tt.expectedEnum, indexTypeAnnot.Name)
			}

			// Verify ElementType
			if arrayType.ElementType == nil {
				t.Fatal("expected ElementType to be set")
			}

			elemTypeAnnot, ok := arrayType.ElementType.(*ast.TypeAnnotation)
			if !ok {
				t.Fatalf("expected TypeAnnotation for ElementType, got %T", arrayType.ElementType)
			}

			if elemTypeAnnot.Name != tt.expectedElem {
				t.Errorf("expected element type %q, got %q", tt.expectedElem, elemTypeAnnot.Name)
			}

			// Verify String() representation
			if arrayType.String() != tt.expectedType {
				t.Errorf("expected String() %q, got %q", tt.expectedType, arrayType.String())
			}
		})
	}
}
