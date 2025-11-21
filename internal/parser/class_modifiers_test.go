package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Virtual/Override Method Tests
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
// Abstract Class/Method Tests
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

// ============================================================================
// External Class Parsing Tests
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
// Forward Class Declaration Tests
// ============================================================================

func TestForwardClassDeclaration(t *testing.T) {
	input := `type TChild = class;`

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

	// Forward declarations have no body, so all slices should be nil or empty
	if len(stmt.Fields) != 0 {
		t.Errorf("Forward declaration should have no fields. got=%d", len(stmt.Fields))
	}

	if len(stmt.Methods) != 0 {
		t.Errorf("Forward declaration should have no methods. got=%d", len(stmt.Methods))
	}

	if len(stmt.Properties) != 0 {
		t.Errorf("Forward declaration should have no properties. got=%d", len(stmt.Properties))
	}
}

func TestForwardClassDeclarationWithParent(t *testing.T) {
	input := `type TChild = class(TBase);`

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

	if stmt.Parent.Value != "TBase" {
		t.Errorf("stmt.Parent.Value not 'TBase'. got=%s", stmt.Parent.Value)
	}

	// Forward declarations have no body
	if len(stmt.Fields) != 0 {
		t.Errorf("Forward declaration should have no fields. got=%d", len(stmt.Fields))
	}

	if len(stmt.Methods) != 0 {
		t.Errorf("Forward declaration should have no methods. got=%d", len(stmt.Methods))
	}
}

func TestMultipleForwardClassDeclarations(t *testing.T) {
	input := `
type TChild = class;
type TBase = class;
type TDerived = class(TBase);
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	// Check first forward declaration
	stmt1, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}
	if stmt1.Name.Value != "TChild" {
		t.Errorf("stmt1.Name.Value not 'TChild'. got=%s", stmt1.Name.Value)
	}

	// Check second forward declaration
	stmt2, ok := program.Statements[1].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.ClassDecl. got=%T",
			program.Statements[1])
	}
	if stmt2.Name.Value != "TBase" {
		t.Errorf("stmt2.Name.Value not 'TBase'. got=%s", stmt2.Name.Value)
	}

	// Check third forward declaration with parent
	stmt3, ok := program.Statements[2].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[2] is not *ast.ClassDecl. got=%T",
			program.Statements[2])
	}
	if stmt3.Name.Value != "TDerived" {
		t.Errorf("stmt3.Name.Value not 'TDerived'. got=%s", stmt3.Name.Value)
	}
	if stmt3.Parent == nil || stmt3.Parent.Value != "TBase" {
		t.Errorf("stmt3.Parent should be 'TBase'. got=%v", stmt3.Parent)
	}
}

func TestForwardDeclarationFollowedByImplementation(t *testing.T) {
	input := `
type TChild = class;

type TBase = class
   function Stuff : TChild; virtual; abstract;
end;

type TChild = class (TBase)
   function Stuff : TChild; override;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	// Check forward declaration
	forward, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}
	if forward.Name.Value != "TChild" {
		t.Errorf("forward.Name.Value not 'TChild'. got=%s", forward.Name.Value)
	}
	if len(forward.Methods) != 0 {
		t.Errorf("Forward declaration should have no methods. got=%d", len(forward.Methods))
	}

	// Check TBase class
	base, ok := program.Statements[1].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.ClassDecl. got=%T",
			program.Statements[1])
	}
	if base.Name.Value != "TBase" {
		t.Errorf("base.Name.Value not 'TBase'. got=%s", base.Name.Value)
	}
	if len(base.Methods) != 1 {
		t.Fatalf("TBase should have 1 method. got=%d", len(base.Methods))
	}

	// Check TChild full implementation
	impl, ok := program.Statements[2].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[2] is not *ast.ClassDecl. got=%T",
			program.Statements[2])
	}
	if impl.Name.Value != "TChild" {
		t.Errorf("impl.Name.Value not 'TChild'. got=%s", impl.Name.Value)
	}
	if impl.Parent == nil || impl.Parent.Value != "TBase" {
		t.Errorf("impl.Parent should be 'TBase'. got=%v", impl.Parent)
	}
	if len(impl.Methods) != 1 {
		t.Fatalf("TChild implementation should have 1 method. got=%d", len(impl.Methods))
	}
}

// ============================================================================
// Partial Class Tests
// ============================================================================

func TestPartialClassDeclaration(t *testing.T) {
	input := `
type TTest = partial class
  Field : Integer;
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

	if stmt.Name.Value != "TTest" {
		t.Errorf("stmt.Name.Value not 'TTest'. got=%s", stmt.Name.Value)
	}

	if !stmt.IsPartial {
		t.Errorf("stmt.IsPartial should be true. got=%v", stmt.IsPartial)
	}

	if len(stmt.Fields) != 1 {
		t.Fatalf("stmt.Fields should contain 1 field. got=%d", len(stmt.Fields))
	}

	if stmt.Fields[0].Name.Value != "Field" {
		t.Errorf("field name should be 'Field'. got=%s", stmt.Fields[0].Name.Value)
	}
}

func TestPartialClassAfterClass(t *testing.T) {
	input := `
type TTest = class partial
  Field : Integer;
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

	if stmt.Name.Value != "TTest" {
		t.Errorf("stmt.Name.Value not 'TTest'. got=%s", stmt.Name.Value)
	}

	if !stmt.IsPartial {
		t.Errorf("stmt.IsPartial should be true. got=%v", stmt.IsPartial)
	}
}

func TestMultiplePartialClassDeclarations(t *testing.T) {
	input := `
type TTest = partial class
  Field : Integer;
  procedure PrintMe; begin PrintLn(Field); end;
end;

type TTest = partial class
  procedure Inc; begin Field+=1; end;
end;

type TTest = class partial
  procedure Dec; begin Field-=1; end;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	// Check all three are partial classes with the same name
	for i, stmt := range program.Statements {
		classDecl, ok := stmt.(*ast.ClassDecl)
		if !ok {
			t.Fatalf("program.Statements[%d] is not *ast.ClassDecl. got=%T",
				i, stmt)
		}

		if classDecl.Name.Value != "TTest" {
			t.Errorf("statement %d: class name not 'TTest'. got=%s", i, classDecl.Name.Value)
		}

		if !classDecl.IsPartial {
			t.Errorf("statement %d: IsPartial should be true. got=%v", i, classDecl.IsPartial)
		}
	}

	// Check first partial class has field and method
	first := program.Statements[0].(*ast.ClassDecl)
	if len(first.Fields) != 1 {
		t.Errorf("first partial class should have 1 field. got=%d", len(first.Fields))
	}
	if len(first.Methods) != 1 {
		t.Errorf("first partial class should have 1 method. got=%d", len(first.Methods))
	}

	// Check second partial class has method
	second := program.Statements[1].(*ast.ClassDecl)
	if len(second.Methods) != 1 {
		t.Errorf("second partial class should have 1 method. got=%d", len(second.Methods))
	}

	// Check third partial class has method
	third := program.Statements[2].(*ast.ClassDecl)
	if len(third.Methods) != 1 {
		t.Errorf("third partial class should have 1 method. got=%d", len(third.Methods))
	}
}

func TestPartialClassWithParent(t *testing.T) {
	input := `
type TChild = partial class(TParent)
  Field : Integer;
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

	if !stmt.IsPartial {
		t.Errorf("stmt.IsPartial should be true. got=%v", stmt.IsPartial)
	}

	if stmt.Parent == nil {
		t.Fatal("stmt.Parent should not be nil")
	}

	if stmt.Parent.Value != "TParent" {
		t.Errorf("stmt.Parent.Value not 'TParent'. got=%s", stmt.Parent.Value)
	}
}

func TestPartialClassString(t *testing.T) {
	input := `type TTest = partial class end;`

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

	// Check that String() includes "partial"
	str := stmt.String()
	if !contains(str, "partial") {
		t.Errorf("stmt.String() should contain 'partial'. got=%s", str)
	}
}

func TestPartialClassWithAbstract(t *testing.T) {
	input := `
type TTest = partial class abstract
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

	if !stmt.IsPartial {
		t.Errorf("stmt.IsPartial should be true. got=%v", stmt.IsPartial)
	}

	if !stmt.IsAbstract {
		t.Errorf("stmt.IsAbstract should be true. got=%v", stmt.IsAbstract)
	}
}
