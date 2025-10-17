package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
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
