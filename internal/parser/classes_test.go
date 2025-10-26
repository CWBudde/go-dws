package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
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
	if stmt.Fields[0].Type.Name != "Integer" {
		t.Errorf("stmt.Fields[0].Type.Name not 'Integer'. got=%s", stmt.Fields[0].Type.Name)
	}

	// Check second field (Y: Integer)
	if stmt.Fields[1].Name.Value != "Y" {
		t.Errorf("stmt.Fields[1].Name.Value not 'Y'. got=%s", stmt.Fields[1].Name.Value)
	}
	if stmt.Fields[1].Type.Name != "Integer" {
		t.Errorf("stmt.Fields[1].Type.Name not 'Integer'. got=%s", stmt.Fields[1].Type.Name)
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

	if method.ReturnType.Name != "Integer" {
		t.Errorf("method.ReturnType.Name not 'Integer'. got=%s", method.ReturnType.Name)
	}
}

// ============================================================================
// New Expression Parsing Tests (Object Creation)
// ============================================================================

func TestNewExpression(t *testing.T) {
	input := `TPoint.Create(10, 20)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	newExpr, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.NewExpression. got=%T",
			stmt.Expression)
	}

	if newExpr.ClassName.Value != "TPoint" {
		t.Errorf("newExpr.ClassName.Value not 'TPoint'. got=%s", newExpr.ClassName.Value)
	}

	if len(newExpr.Arguments) != 2 {
		t.Fatalf("newExpr.Arguments should contain 2 arguments. got=%d", len(newExpr.Arguments))
	}
}

func TestNewExpressionNoArguments(t *testing.T) {
	input := `TObject.Create()`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	newExpr, ok := stmt.Expression.(*ast.NewExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.NewExpression. got=%T",
			stmt.Expression)
	}

	if newExpr.ClassName.Value != "TObject" {
		t.Errorf("newExpr.ClassName.Value not 'TObject'. got=%s", newExpr.ClassName.Value)
	}

	if len(newExpr.Arguments) != 0 {
		t.Fatalf("newExpr.Arguments should be empty. got=%d", len(newExpr.Arguments))
	}
}

// ============================================================================
// Member Access Parsing Tests
// ============================================================================

func TestMemberAccess(t *testing.T) {
	input := `point.X`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	memberExpr, ok := stmt.Expression.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MemberAccessExpression. got=%T",
			stmt.Expression)
	}

	// Check object is an identifier
	ident, ok := memberExpr.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("memberExpr.Object is not *ast.Identifier. got=%T",
			memberExpr.Object)
	}

	if ident.Value != "point" {
		t.Errorf("ident.Value not 'point'. got=%s", ident.Value)
	}

	if memberExpr.Member.Value != "X" {
		t.Errorf("memberExpr.Member.Value not 'X'. got=%s", memberExpr.Member.Value)
	}
}

func TestChainedMemberAccess(t *testing.T) {
	input := `obj.field1.field2`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	// The outer member access (field2)
	outerMember, ok := stmt.Expression.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MemberAccessExpression. got=%T",
			stmt.Expression)
	}

	if outerMember.Member.Value != "field2" {
		t.Errorf("outerMember.Member.Value not 'field2'. got=%s", outerMember.Member.Value)
	}

	// The inner member access (obj.field1)
	innerMember, ok := outerMember.Object.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("outerMember.Object is not *ast.MemberAccessExpression. got=%T",
			outerMember.Object)
	}

	if innerMember.Member.Value != "field1" {
		t.Errorf("innerMember.Member.Value not 'field1'. got=%s", innerMember.Member.Value)
	}

	// The base object
	ident, ok := innerMember.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("innerMember.Object is not *ast.Identifier. got=%T",
			innerMember.Object)
	}

	if ident.Value != "obj" {
		t.Errorf("ident.Value not 'obj'. got=%s", ident.Value)
	}
}

// ============================================================================
// Method Call Parsing Tests
// ============================================================================

func TestMethodCall(t *testing.T) {
	input := `obj.DoSomething(42, "hello")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	methodCall, ok := stmt.Expression.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MethodCallExpression. got=%T",
			stmt.Expression)
	}

	// Check object is an identifier
	ident, ok := methodCall.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("methodCall.Object is not *ast.Identifier. got=%T",
			methodCall.Object)
	}

	if ident.Value != "obj" {
		t.Errorf("ident.Value not 'obj'. got=%s", ident.Value)
	}

	if methodCall.Method.Value != "DoSomething" {
		t.Errorf("methodCall.Method.Value not 'DoSomething'. got=%s", methodCall.Method.Value)
	}

	if len(methodCall.Arguments) != 2 {
		t.Fatalf("methodCall.Arguments should contain 2 arguments. got=%d", len(methodCall.Arguments))
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestClassDeclarationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing class name",
			"type = class end;",
		},
		{
			"missing equals sign",
			"type TPoint class end;",
		},
		{
			"missing class keyword",
			"type TPoint = end;",
		},
		{
			"missing end keyword",
			"type TPoint = class X: Integer;",
		},
		{
			"missing semicolon after end",
			"type TPoint = class end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected parser errors but got none")
			}
		})
	}
}

func TestFieldDeclarationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing type in field",
			"type TPoint = class X:; end;",
		},
		{
			"missing semicolon after field",
			"type TPoint = class X: Integer end;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected parser errors but got none")
			}
		})
	}
}

func TestMemberAccessErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing identifier after dot",
			"obj.;",
		},
		{
			"number after dot",
			"obj.123;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected parser errors but got none")
			}
		})
	}
}

// ============================================================================
// Virtual/Override Method Tests (Task 7.64)
// ============================================================================

func TestVirtualMethodDeclaration(t *testing.T) {
	input := `
type TBase = class
  function DoWork(): Integer; virtual;
  begin
    Result := 1;
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
	if method.Name.Value != "DoWork" {
		t.Errorf("method.Name.Value not 'DoWork'. got=%s", method.Name.Value)
	}

	if !method.IsVirtual {
		t.Errorf("method.IsVirtual should be true. got=%v", method.IsVirtual)
	}

	if method.IsOverride {
		t.Errorf("method.IsOverride should be false. got=%v", method.IsOverride)
	}
}

func TestOverrideMethodDeclaration(t *testing.T) {
	input := `
type TChild = class(TBase)
  function DoWork(): Integer; override;
  begin
    Result := 2;
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
	if method.Name.Value != "DoWork" {
		t.Errorf("method.Name.Value not 'DoWork'. got=%s", method.Name.Value)
	}

	if method.IsVirtual {
		t.Errorf("method.IsVirtual should be false. got=%v", method.IsVirtual)
	}

	if !method.IsOverride {
		t.Errorf("method.IsOverride should be true. got=%v", method.IsOverride)
	}
}

func TestVirtualAndOverrideInSameClass(t *testing.T) {
	input := `
type TMixed = class
  function Method1(): Integer; virtual;
  begin
    Result := 1;
  end;

  function Method2(): String; virtual;
  begin
    Result := 'hello';
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

	if len(stmt.Methods) != 2 {
		t.Fatalf("stmt.Methods should contain 2 methods. got=%d", len(stmt.Methods))
	}

	// Check both methods are virtual
	for i, method := range stmt.Methods {
		if !method.IsVirtual {
			t.Errorf("method[%d].IsVirtual should be true. got=%v", i, method.IsVirtual)
		}
		if method.IsOverride {
			t.Errorf("method[%d].IsOverride should be false. got=%v", i, method.IsOverride)
		}
	}
}

// ============================================================================
// Abstract Class/Method Tests (Task 7.65)
// ============================================================================

func TestAbstractClassDeclaration(t *testing.T) {
	input := `
type TShape = class abstract
  FName: String;
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

	if stmt.Name.Value != "TShape" {
		t.Errorf("stmt.Name.Value not 'TShape'. got=%s", stmt.Name.Value)
	}

	if !stmt.IsAbstract {
		t.Errorf("stmt.IsAbstract should be true. got=%v", stmt.IsAbstract)
	}
}

func TestAbstractMethodDeclaration(t *testing.T) {
	input := `
type TShape = class abstract
  function GetArea(): Float; abstract;
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
	if method.Name.Value != "GetArea" {
		t.Errorf("method.Name.Value not 'GetArea'. got=%s", method.Name.Value)
	}

	if !method.IsAbstract {
		t.Errorf("method.IsAbstract should be true. got=%v", method.IsAbstract)
	}

	if method.Body != nil {
		t.Errorf("abstract method should have nil Body. got=%v", method.Body)
	}
}

func TestAbstractClassWithMixedMethods(t *testing.T) {
	input := `
type TShape = class abstract
  function GetArea(): Float; abstract;

  function GetName(): String;
  begin
    Result := 'Shape';
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

	if !stmt.IsAbstract {
		t.Errorf("stmt.IsAbstract should be true. got=%v", stmt.IsAbstract)
	}

	if len(stmt.Methods) != 2 {
		t.Fatalf("stmt.Methods should contain 2 methods. got=%d", len(stmt.Methods))
	}

	// First method should be abstract
	method1 := stmt.Methods[0]
	if method1.Name.Value != "GetArea" {
		t.Errorf("method1.Name.Value not 'GetArea'. got=%s", method1.Name.Value)
	}
	if !method1.IsAbstract {
		t.Errorf("method1.IsAbstract should be true. got=%v", method1.IsAbstract)
	}
	if method1.Body != nil {
		t.Errorf("abstract method should have nil Body. got=%v", method1.Body)
	}

	// Second method should be concrete
	method2 := stmt.Methods[1]
	if method2.Name.Value != "GetName" {
		t.Errorf("method2.Name.Value not 'GetName'. got=%s", method2.Name.Value)
	}
	if method2.IsAbstract {
		t.Errorf("method2.IsAbstract should be false. got=%v", method2.IsAbstract)
	}
	if method2.Body == nil {
		t.Errorf("concrete method should have Body. got=nil")
	}
}

func TestParseClassOperatorDeclarations(t *testing.T) {
	input := `
type TMyRange = class
   class operator += String uses AppendString;
   class operator IN array of Integer uses ContainsArray;
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

	if len(stmt.Operators) != 2 {
		t.Fatalf("stmt.Operators should contain 2 operators. got=%d", len(stmt.Operators))
	}

	first := stmt.Operators[0]
	if first.Kind != ast.OperatorKindClass {
		t.Fatalf("first operator kind expected OperatorKindClass, got %s", first.Kind)
	}
	if first.OperatorSymbol != "+=" {
		t.Fatalf("first operator symbol expected '+='; got %q", first.OperatorSymbol)
	}
	if first.Arity != 1 {
		t.Fatalf("first operator arity expected 1; got %d", first.Arity)
	}
	if len(first.OperandTypes) != 1 || first.OperandTypes[0].String() != "String" {
		t.Fatalf("first operator operand expected 'String'; got %v", first.OperandTypes)
	}
	if first.Binding == nil || first.Binding.Value != "AppendString" {
		t.Fatalf("first operator binding expected 'AppendString'; got %v", first.Binding)
	}

	second := stmt.Operators[1]
	if second.OperatorSymbol != "in" {
		t.Fatalf("second operator symbol expected 'in'; got %q", second.OperatorSymbol)
	}
	if len(second.OperandTypes) != 1 || second.OperandTypes[0].String() != "array of Integer" {
		t.Fatalf("second operator operand expected 'array of Integer'; got %v", second.OperandTypes)
	}
	if second.Binding == nil || second.Binding.Value != "ContainsArray" {
		t.Fatalf("second operator binding expected 'ContainsArray'; got %v", second.Binding)
	}
}

// ============================================================================
// Task 7.138: External Class Parsing Tests
// ============================================================================

func TestExternalClassParsing(t *testing.T) {
	t.Run("external class without name", func(t *testing.T) {
		input := `
type TExternal = class external
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ClassDecl)
		if !ok {
			t.Fatalf("Expected *ast.ClassDecl, got %T", program.Statements[0])
		}

		if stmt.Name.Value != "TExternal" {
			t.Errorf("Expected class name 'TExternal', got %q", stmt.Name.Value)
		}

		if !stmt.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if stmt.ExternalName != "" {
			t.Errorf("Expected empty ExternalName, got %q", stmt.ExternalName)
		}
	})

	t.Run("external class with name", func(t *testing.T) {
		input := `
type TExternal = class external 'MyExternalClass'
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ClassDecl)
		if !ok {
			t.Fatalf("Expected *ast.ClassDecl, got %T", program.Statements[0])
		}

		if !stmt.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if stmt.ExternalName != "MyExternalClass" {
			t.Errorf("Expected ExternalName 'MyExternalClass', got %q", stmt.ExternalName)
		}
	})

	t.Run("external class with parent", func(t *testing.T) {
		input := `
type TExternal = class(TParent) external
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ClassDecl)
		if !ok {
			t.Fatalf("Expected *ast.ClassDecl, got %T", program.Statements[0])
		}

		if stmt.Parent == nil || stmt.Parent.Value != "TParent" {
			t.Errorf("Expected parent 'TParent', got %v", stmt.Parent)
		}

		if !stmt.IsExternal {
			t.Error("Expected IsExternal to be true")
		}
	})

	t.Run("external class with methods", func(t *testing.T) {
		input := `
type TExternal = class external 'External'
  procedure DoSomething;
  function GetValue: Integer;
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ClassDecl)
		if !ok {
			t.Fatalf("Expected *ast.ClassDecl, got %T", program.Statements[0])
		}

		if !stmt.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if stmt.ExternalName != "External" {
			t.Errorf("Expected ExternalName 'External', got %q", stmt.ExternalName)
		}

		if len(stmt.Methods) != 2 {
			t.Fatalf("Expected 2 methods, got %d", len(stmt.Methods))
		}
	})

	t.Run("regular class is not external", func(t *testing.T) {
		input := `
type TRegular = class
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ClassDecl)
		if !ok {
			t.Fatalf("Expected *ast.ClassDecl, got %T", program.Statements[0])
		}

		if stmt.IsExternal {
			t.Error("Regular class should not be external")
		}

		if stmt.ExternalName != "" {
			t.Errorf("Regular class should have empty ExternalName, got %q", stmt.ExternalName)
		}
	})
}

// ============================================================================
// Task 7.140: External Method Parsing Tests
// ============================================================================

func TestExternalMethodParsing(t *testing.T) {
	t.Run("external method without name", func(t *testing.T) {
		input := `
type TExternal = class external
  procedure Hello; external;
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		if len(stmt.Methods) != 1 {
			t.Fatalf("Expected 1 method, got %d", len(stmt.Methods))
		}

		method := stmt.Methods[0]
		if !method.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if method.ExternalName != "" {
			t.Errorf("Expected empty ExternalName, got %q", method.ExternalName)
		}
	})

	t.Run("external method with name", func(t *testing.T) {
		input := `
type TExternal = class external
  procedure Hello; external 'world';
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		method := stmt.Methods[0]

		if !method.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if method.ExternalName != "world" {
			t.Errorf("Expected ExternalName 'world', got %q", method.ExternalName)
		}
	})

	t.Run("external function with name", func(t *testing.T) {
		input := `
type TExternal = class external
  function GetValue: Integer; external 'getValue';
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		method := stmt.Methods[0]

		if !method.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if method.ExternalName != "getValue" {
			t.Errorf("Expected ExternalName 'getValue', got %q", method.ExternalName)
		}

		if method.ReturnType == nil || method.ReturnType.Name != "Integer" {
			t.Error("Expected return type Integer")
		}
	})

	t.Run("regular method is not external", func(t *testing.T) {
		input := `
type TRegular = class
  procedure DoSomething;
  begin
  end;
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		method := stmt.Methods[0]

		if method.IsExternal {
			t.Error("Regular method should not be external")
		}

		if method.ExternalName != "" {
			t.Errorf("Regular method should have empty ExternalName, got %q", method.ExternalName)
		}
	})
}
