package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestInlineMethodInClass tests parsing of inline method implementations
// within a class declaration (Phase 9.9).
func TestInlineMethodInClass(t *testing.T) {
	input := `type
   TObj = class
      function GetName : String;
      begin
         Result := 'TObj';
      end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	// Check that we have a class declaration
	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if classDecl.Name.Value != "TObj" {
		t.Errorf("classDecl.Name.Value not 'TObj'. got=%q", classDecl.Name.Value)
	}

	// Check that the class has one method
	if len(classDecl.Methods) != 1 {
		t.Fatalf("classDecl.Methods does not contain 1 method. got=%d",
			len(classDecl.Methods))
	}

	method := classDecl.Methods[0]
	if method.Name.Value != "GetName" {
		t.Errorf("method.Name.Value not 'GetName'. got=%q", method.Name.Value)
	}

	// Check that the method has a body (inline implementation)
	if method.Body == nil {
		t.Fatal("method.Body is nil, expected inline method implementation")
	}

	// Check that the body contains the expected statement
	if len(method.Body.Statements) == 0 {
		t.Fatal("method.Body.Statements is empty")
	}
}

// TestMultipleInlineMethodsInClass tests parsing of multiple inline methods
func TestMultipleInlineMethodsInClass(t *testing.T) {
	input := `type
   TObj = class
      function GetName : String;
      begin
         Result := 'TObj';
      end;

      procedure SetValue(x: Integer);
      begin
         PrintLn(x);
      end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	// Check that the class has two methods
	if len(classDecl.Methods) != 2 {
		t.Fatalf("classDecl.Methods does not contain 2 methods. got=%d",
			len(classDecl.Methods))
	}

	// Check first method (function)
	method1 := classDecl.Methods[0]
	if method1.Name.Value != "GetName" {
		t.Errorf("method1.Name.Value not 'GetName'. got=%q", method1.Name.Value)
	}
	if method1.Body == nil {
		t.Error("method1.Body is nil, expected inline method implementation")
	}

	// Check second method (procedure)
	method2 := classDecl.Methods[1]
	if method2.Name.Value != "SetValue" {
		t.Errorf("method2.Name.Value not 'SetValue'. got=%q", method2.Name.Value)
	}
	if method2.Body == nil {
		t.Error("method2.Body is nil, expected inline method implementation")
	}
}

// TestInlineConstructorInClass tests parsing of inline constructor
func TestInlineConstructorInClass(t *testing.T) {
	input := `type
   TObj = class
      constructor Create;
      begin
      end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	// Check that the class has one method (constructor)
	if len(classDecl.Methods) != 1 {
		t.Fatalf("classDecl.Methods does not contain 1 method. got=%d",
			len(classDecl.Methods))
	}

	method := classDecl.Methods[0]
	if method.Name.Value != "Create" {
		t.Errorf("method.Name.Value not 'Create'. got=%q", method.Name.Value)
	}
	if !method.IsConstructor {
		t.Error("method.IsConstructor is false, expected true")
	}
	if method.Body == nil {
		t.Error("method.Body is nil, expected inline method implementation")
	}
}

// TestMixedInlineAndForwardMethods tests mixing inline and forward declaration methods
func TestMixedInlineAndForwardMethods(t *testing.T) {
	input := `type
   TObj = class
      function GetName : String;
      begin
         Result := 'TObj';
      end;

      procedure SetValue(x: Integer); // Forward declaration (no body)
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	// Check that the class has two methods
	if len(classDecl.Methods) != 2 {
		t.Fatalf("classDecl.Methods does not contain 2 methods. got=%d",
			len(classDecl.Methods))
	}

	// First method should have a body (inline)
	if classDecl.Methods[0].Body == nil {
		t.Error("first method.Body is nil, expected inline implementation")
	}

	// Second method should not have a body (forward declaration)
	if classDecl.Methods[1].Body != nil {
		t.Error("second method.Body is not nil, expected forward declaration")
	}
}
