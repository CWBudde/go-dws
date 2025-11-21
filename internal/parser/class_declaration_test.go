package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Class Declaration Parsing Tests
// ============================================================================

func TestSimpleClassDeclaration(t *testing.T) {
	input := `
type TPoint = class
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "TPoint" {
		t.Errorf("stmt.Name.Value not 'TPoint'. got=%s", stmt.Name.Value)
	}

	if stmt.Parent != nil {
		t.Errorf("stmt.Parent should be nil for root class. got=%v", stmt.Parent)
	}

	if len(stmt.Fields) != 0 {
		t.Errorf("stmt.Fields should be empty. got=%d", len(stmt.Fields))
	}

	if len(stmt.Methods) != 0 {
		t.Errorf("stmt.Methods should be empty. got=%d", len(stmt.Methods))
	}
}

func TestClassWithInheritance(t *testing.T) {
	input := `
type TChild = class(TParent)
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "TChild" {
		t.Errorf("stmt.Name.Value not 'TChild'. got=%s", stmt.Name.Value)
	}

	if stmt.Parent == nil {
		t.Fatal("stmt.Parent should not be nil")
	}

	if stmt.Parent.Value != "TParent" {
		t.Errorf("stmt.Parent.Value not 'TParent'. got=%s", stmt.Parent.Value)
	}
}

func TestClassWithFields(t *testing.T) {
	input := `
type TPoint = class
  X: Integer;
  Y: Integer;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Fields) != 2 {
		t.Fatalf("stmt.Fields should contain 2 fields. got=%d", len(stmt.Fields))
	}

	// Check first field (X: Integer)
	if stmt.Fields[0].Name.Value != "X" {
		t.Errorf("stmt.Fields[0].Name.Value not 'X'. got=%s", stmt.Fields[0].Name.Value)
	}
	typeAnnot1, ok := stmt.Fields[0].Type.(*ast.TypeAnnotation)
	if !ok || typeAnnot1.Name != "Integer" {
		t.Errorf("stmt.Fields[0].Type not 'Integer'. got=%v", stmt.Fields[0].Type)
	}

	// Check second field (Y: Integer)
	if stmt.Fields[1].Name.Value != "Y" {
		t.Errorf("stmt.Fields[1].Name.Value not 'Y'. got=%s", stmt.Fields[1].Name.Value)
	}
	typeAnnot2, ok := stmt.Fields[1].Type.(*ast.TypeAnnotation)
	if !ok || typeAnnot2.Name != "Integer" {
		t.Errorf("stmt.Fields[1].Type not 'Integer'. got=%v", stmt.Fields[1].Type)
	}
}

func TestClassWithMethod(t *testing.T) {
	input := `
type TCounter = class
  function GetValue(): Integer;
  begin
    Result := 0;
  end;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Methods) != 1 {
		t.Fatalf("stmt.Methods should contain 1 method. got=%d", len(stmt.Methods))
	}

	method := stmt.Methods[0]
	if method.Name.Value != "GetValue" {
		t.Errorf("method.Name.Value not 'GetValue'. got=%s", method.Name.Value)
	}

	if method.ReturnType == nil {
		t.Fatal("method.ReturnType should not be nil")
	}

	if method.ReturnType.String() != "Integer" {
		t.Errorf("method.ReturnType.String() not 'Integer'. got=%s", method.ReturnType.String())
	}
}

func TestClassWithMethodKeyword(t *testing.T) {
	input := `
type TPoint = class
  X, Y: Integer;

  method GetX(): Integer;
  begin
    Result := X;
  end;

  method SetX(value: Integer);
  begin
    X := value;
  end;

  function GetY(): Integer;
  begin
    Result := Y;
  end;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Methods) != 3 {
		t.Fatalf("stmt.Methods should contain 3 methods. got=%d", len(stmt.Methods))
	}

	// Test first method declared with 'method' keyword
	method1 := stmt.Methods[0]
	if method1.Name.Value != "GetX" {
		t.Errorf("method1.Name.Value not 'GetX'. got=%s", method1.Name.Value)
	}
	if method1.Token.Type != lexer.METHOD {
		t.Errorf("method1.Token.Type not METHOD. got=%s", method1.Token.Type)
	}
	if method1.ReturnType == nil {
		t.Fatal("method1.ReturnType should not be nil")
	}
	if method1.ReturnType.String() != "Integer" {
		t.Errorf("method1.ReturnType.String() not 'Integer'. got=%s", method1.ReturnType.String())
	}

	// Test second method declared with 'method' keyword (procedure-style)
	method2 := stmt.Methods[1]
	if method2.Name.Value != "SetX" {
		t.Errorf("method2.Name.Value not 'SetX'. got=%s", method2.Name.Value)
	}
	if method2.Token.Type != lexer.METHOD {
		t.Errorf("method2.Token.Type not METHOD. got=%s", method2.Token.Type)
	}
	if method2.ReturnType != nil {
		t.Errorf("method2.ReturnType should be nil for procedure-style method. got=%v", method2.ReturnType)
	}
	if len(method2.Parameters) != 1 {
		t.Fatalf("method2.Parameters should contain 1 parameter. got=%d", len(method2.Parameters))
	}

	// Test third method declared with 'function' keyword (should still work)
	method3 := stmt.Methods[2]
	if method3.Name.Value != "GetY" {
		t.Errorf("method3.Name.Value not 'GetY'. got=%s", method3.Name.Value)
	}
	if method3.Token.Type != lexer.FUNCTION {
		t.Errorf("method3.Token.Type not FUNCTION. got=%s", method3.Token.Type)
	}
}

// Test class with inline array type fields
func TestClassWithInlineArrayFields(t *testing.T) {
	input := `
type TBoard = class
  Pix : array of array of Integer;
  Data : array of String;
  Grid : array[1..10] of Integer;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Fields) != 3 {
		t.Fatalf("stmt.Fields should contain 3 fields. got=%d", len(stmt.Fields))
	}

	// Test first field: array of array of Integer
	field1 := stmt.Fields[0]
	if field1.Name.Value != "Pix" {
		t.Errorf("field1.Name.Value not 'Pix'. got=%s", field1.Name.Value)
	}
	arrayType1, ok := field1.Type.(*ast.ArrayTypeNode)
	if !ok {
		t.Fatalf("field1.Type is not *ast.ArrayTypeNode. got=%T", field1.Type)
	}
	// Check it's a nested array type
	innerArray, ok := arrayType1.ElementType.(*ast.ArrayTypeNode)
	if !ok {
		t.Fatalf("field1 element type is not *ast.ArrayTypeNode. got=%T", arrayType1.ElementType)
	}
	innerType, ok := innerArray.ElementType.(*ast.TypeAnnotation)
	if !ok || innerType.Name != "Integer" {
		t.Errorf("field1 inner element type not 'Integer'. got=%v", innerArray.ElementType)
	}

	// Test second field: array of String
	field2 := stmt.Fields[1]
	if field2.Name.Value != "Data" {
		t.Errorf("field2.Name.Value not 'Data'. got=%s", field2.Name.Value)
	}
	arrayType2, ok := field2.Type.(*ast.ArrayTypeNode)
	if !ok {
		t.Fatalf("field2.Type is not *ast.ArrayTypeNode. got=%T", field2.Type)
	}
	elementType2, ok := arrayType2.ElementType.(*ast.TypeAnnotation)
	if !ok || elementType2.Name != "String" {
		t.Errorf("field2 element type not 'String'. got=%v", arrayType2.ElementType)
	}

	// Test third field: static array[1..10] of Integer
	field3 := stmt.Fields[2]
	if field3.Name.Value != "Grid" {
		t.Errorf("field3.Name.Value not 'Grid'. got=%s", field3.Name.Value)
	}
	arrayType3, ok := field3.Type.(*ast.ArrayTypeNode)
	if !ok {
		t.Fatalf("field3.Type is not *ast.ArrayTypeNode. got=%T", field3.Type)
	}
	if arrayType3.LowBound == nil {
		t.Error("field3 LowBound should not be nil")
	}
	if lowBound, ok := arrayType3.LowBound.(*ast.IntegerLiteral); !ok || lowBound.Value != 1 {
		t.Errorf("field3 LowBound not IntegerLiteral with value 1. got=%v", arrayType3.LowBound)
	}
	if arrayType3.HighBound == nil {
		t.Error("field3 HighBound should not be nil")
	}
	if highBound, ok := arrayType3.HighBound.(*ast.IntegerLiteral); !ok || highBound.Value != 10 {
		t.Errorf("field3 HighBound not IntegerLiteral with value 10. got=%v", arrayType3.HighBound)
	}
}

// TestParseClassDeclaration tests the parseClassDeclaration function directly
func TestParseClassDeclaration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		checkName   string
		expectNil   bool
		expectError bool
	}{
		{
			name:        "valid class declaration",
			input:       "type TPoint = class end;",
			expectNil:   false,
			expectError: false,
			checkName:   "TPoint",
		},
		{
			name:        "valid class with parent",
			input:       "type TChild = class(TParent) end;",
			expectNil:   false,
			expectError: false,
			checkName:   "TChild",
		},
		{
			name:        "missing class name identifier",
			input:       "type = class end;",
			expectNil:   true,
			expectError: true,
		},
		{
			name:        "missing equals sign",
			input:       "type TPoint class end;",
			expectNil:   true,
			expectError: true,
		},
		{
			name:        "missing class keyword",
			input:       "type TPoint = end;",
			expectNil:   true,
			expectError: true,
		},
		{
			name:        "wrong token after equals (interface instead of class)",
			input:       "type TPoint = interface end;",
			expectNil:   true,
			expectError: true,
		},
		{
			name:        "number instead of identifier for class name",
			input:       "type 123 = class end;",
			expectNil:   true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			// parseClassDeclaration expects cursor to be positioned
			// at the token BEFORE the class name identifier
			// In this case, we position at TYPE token
			// (cursor is already at TYPE after New())

			result := p.parseClassDeclaration()

			if tt.expectNil && result != nil {
				t.Errorf("Expected nil result, got %T", result)
			}
			if !tt.expectNil && result == nil {
				t.Error("Expected result, got nil")
			}

			if result != nil && tt.checkName != "" {
				if result.Name.Value != tt.checkName {
					t.Errorf("Expected class name %q, got %q", tt.checkName, result.Name.Value)
				}
			}

			hasErrors := len(p.Errors()) > 0
			if hasErrors != tt.expectError {
				t.Errorf("Expected error = %v, got %v (errors: %d)", tt.expectError, hasErrors, len(p.Errors()))
				if hasErrors {
					for _, err := range p.Errors() {
						t.Logf("  Error: %s", err.Message)
					}
				}
			}
		})
	}
}
