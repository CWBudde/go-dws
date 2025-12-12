package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestInlineMethodInClass tests parsing of inline method implementations within a class declaration.
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

// TestInlineMethodWithCallingConvention tests parsing of inline methods with calling conventions
func TestInlineMethodWithCallingConvention(t *testing.T) {
	input := `type
   TMyClass = class
      procedure Test1; safecall; begin PrintLn('test'); end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(classDecl.Methods))
	}

	method := classDecl.Methods[0]
	if method.Name.Value != "Test1" {
		t.Errorf("Expected method name 'Test1', got '%s'", method.Name.Value)
	}

	if method.CallingConvention != "safecall" {
		t.Errorf("Expected calling convention 'safecall', got '%s'", method.CallingConvention)
	}

	if method.Body == nil {
		t.Fatal("Expected method body, got nil")
	}

	if len(method.Body.Statements) == 0 {
		t.Error("Expected method body to have statements")
	}
}

// TestClassMethodWithCallingConvention tests parsing of class methods with calling conventions
func TestClassMethodWithCallingConvention(t *testing.T) {
	input := `type
   TMyClass = class
      class procedure Test1; safecall; begin PrintLn('test'); end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(classDecl.Methods))
	}

	method := classDecl.Methods[0]
	if method.Name.Value != "Test1" {
		t.Errorf("Expected method name 'Test1', got '%s'", method.Name.Value)
	}

	if !method.IsClassMethod {
		t.Error("Expected IsClassMethod to be true")
	}

	if method.CallingConvention != "safecall" {
		t.Errorf("Expected calling convention 'safecall', got '%s'", method.CallingConvention)
	}

	if method.Body == nil {
		t.Fatal("Expected method body, got nil")
	}

	if len(method.Body.Statements) == 0 {
		t.Error("Expected method body to have statements")
	}
}

// TestInlineMethodWithVirtualAndCallingConvention tests parsing methods with multiple directives
func TestInlineMethodWithVirtualAndCallingConvention(t *testing.T) {
	input := `type
   TMyClass = class
      method Test3; virtual; cdecl; begin end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(classDecl.Methods))
	}

	method := classDecl.Methods[0]
	if method.Name.Value != "Test3" {
		t.Errorf("Expected method name 'Test3', got '%s'", method.Name.Value)
	}

	if !method.IsVirtual {
		t.Error("Expected IsVirtual to be true")
	}

	if method.CallingConvention != "cdecl" {
		t.Errorf("Expected calling convention 'cdecl', got '%s'", method.CallingConvention)
	}

	if method.Body == nil {
		t.Fatal("Expected method body, got nil")
	}
}

// TestMultipleInlineMethodsWithCallingConventions tests parsing multiple methods with different calling conventions
func TestMultipleInlineMethodsWithCallingConventions(t *testing.T) {
	input := `type
   TMyClass = class
      class procedure Test1; safecall; begin end;
      procedure Test2; stdcall; begin end;
      method Test3; virtual; cdecl; begin end;
   end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 3 {
		t.Fatalf("Expected 3 methods, got %d", len(classDecl.Methods))
	}

	// Check first method
	method1 := classDecl.Methods[0]
	if method1.Name.Value != "Test1" {
		t.Errorf("Expected method 1 name 'Test1', got '%s'", method1.Name.Value)
	}
	if !method1.IsClassMethod {
		t.Error("Expected method 1 IsClassMethod to be true")
	}
	if method1.CallingConvention != "safecall" {
		t.Errorf("Expected method 1 calling convention 'safecall', got '%s'", method1.CallingConvention)
	}
	if method1.Body == nil {
		t.Error("Expected method 1 to have a body")
	}

	// Check second method
	method2 := classDecl.Methods[1]
	if method2.Name.Value != "Test2" {
		t.Errorf("Expected method 2 name 'Test2', got '%s'", method2.Name.Value)
	}
	if method2.CallingConvention != "stdcall" {
		t.Errorf("Expected method 2 calling convention 'stdcall', got '%s'", method2.CallingConvention)
	}
	if method2.Body == nil {
		t.Error("Expected method 2 to have a body")
	}

	// Check third method
	method3 := classDecl.Methods[2]
	if method3.Name.Value != "Test3" {
		t.Errorf("Expected method 3 name 'Test3', got '%s'", method3.Name.Value)
	}
	if !method3.IsVirtual {
		t.Error("Expected method 3 IsVirtual to be true")
	}
	if method3.CallingConvention != "cdecl" {
		t.Errorf("Expected method 3 calling convention 'cdecl', got '%s'", method3.CallingConvention)
	}
	if method3.Body == nil {
		t.Error("Expected method 3 to have a body")
	}
}

// TestStandaloneFunctionWithCallingConvention tests parsing standalone functions with calling conventions
func TestStandaloneFunctionWithCallingConvention(t *testing.T) {
	input := `procedure Test4; register;
begin
end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name.Value != "Test4" {
		t.Errorf("Expected function name 'Test4', got '%s'", funcDecl.Name.Value)
	}

	if funcDecl.CallingConvention != "register" {
		t.Errorf("Expected calling convention 'register', got '%s'", funcDecl.CallingConvention)
	}

	if funcDecl.Body == nil {
		t.Fatal("Expected function body, got nil")
	}
}

// TestStandaloneFunctionWithOverloadAndCallingConvention tests multiple directives
func TestStandaloneFunctionWithOverloadAndCallingConvention(t *testing.T) {
	input := `procedure Test5; overload; pascal;
begin
end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name.Value != "Test5" {
		t.Errorf("Expected function name 'Test5', got '%s'", funcDecl.Name.Value)
	}

	if !funcDecl.IsOverload {
		t.Error("Expected IsOverload to be true")
	}

	if funcDecl.CallingConvention != "pascal" {
		t.Errorf("Expected calling convention 'pascal', got '%s'", funcDecl.CallingConvention)
	}

	if funcDecl.Body == nil {
		t.Fatal("Expected function body, got nil")
	}
}
