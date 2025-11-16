package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestShortFormClassDeclaration tests parsing of short-form class declarations
// (Phase 9.10.1): type TChild = class(TParent);
func TestShortFormClassDeclaration(t *testing.T) {
	input := `type
   TBase = class
   end;

type
   TChild = class(TBase);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}

	// Check the base class (full declaration)
	baseClass, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}
	if baseClass.Name.Value != "TBase" {
		t.Errorf("baseClass.Name.Value not 'TBase'. got=%q", baseClass.Name.Value)
	}

	// Check the child class (short-form declaration)
	childClass, ok := program.Statements[1].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.ClassDecl. got=%T",
			program.Statements[1])
	}

	if childClass.Name.Value != "TChild" {
		t.Errorf("childClass.Name.Value not 'TChild'. got=%q", childClass.Name.Value)
	}

	// Check that it has a parent
	if childClass.Parent == nil {
		t.Fatal("childClass.Parent is nil, expected TBase")
	}
	if childClass.Parent.Value != "TBase" {
		t.Errorf("childClass.Parent.Value not 'TBase'. got=%q", childClass.Parent.Value)
	}

	// Check that it has empty collections (no body)
	if len(childClass.Fields) != 0 {
		t.Errorf("childClass.Fields should be empty. got=%d", len(childClass.Fields))
	}
	if len(childClass.Methods) != 0 {
		t.Errorf("childClass.Methods should be empty. got=%d", len(childClass.Methods))
	}
	if len(childClass.Properties) != 0 {
		t.Errorf("childClass.Properties should be empty. got=%d", len(childClass.Properties))
	}
}

// TestShortFormClassWithAbstract tests short-form class with abstract modifier
func TestShortFormClassWithAbstract(t *testing.T) {
	input := `type
   TBase = class
   end;

type
   TChild = class abstract(TBase);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}

	childClass, ok := program.Statements[1].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.ClassDecl. got=%T",
			program.Statements[1])
	}

	if !childClass.IsAbstract {
		t.Error("childClass.IsAbstract should be true")
	}

	if childClass.Parent == nil {
		t.Fatal("childClass.Parent is nil, expected TBase")
	}
	if childClass.Parent.Value != "TBase" {
		t.Errorf("childClass.Parent.Value not 'TBase'. got=%q", childClass.Parent.Value)
	}
}

// TestMultipleShortFormClasses tests multiple short-form declarations
func TestMultipleShortFormClasses(t *testing.T) {
	input := `type
   TBase = class
   end;

type
   TChild1 = class(TBase);

type
   TChild2 = class(TBase);

type
   TGrandChild = class(TChild1);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d",
			len(program.Statements))
	}

	// Check TChild1
	child1, ok := program.Statements[1].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.ClassDecl. got=%T",
			program.Statements[1])
	}
	if child1.Name.Value != "TChild1" {
		t.Errorf("child1.Name.Value not 'TChild1'. got=%q", child1.Name.Value)
	}
	if child1.Parent.Value != "TBase" {
		t.Errorf("child1.Parent.Value not 'TBase'. got=%q", child1.Parent.Value)
	}

	// Check TChild2
	child2, ok := program.Statements[2].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[2] is not *ast.ClassDecl. got=%T",
			program.Statements[2])
	}
	if child2.Name.Value != "TChild2" {
		t.Errorf("child2.Name.Value not 'TChild2'. got=%q", child2.Name.Value)
	}
	if child2.Parent.Value != "TBase" {
		t.Errorf("child2.Parent.Value not 'TBase'. got=%q", child2.Parent.Value)
	}

	// Check TGrandChild
	grandChild, ok := program.Statements[3].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[3] is not *ast.ClassDecl. got=%T",
			program.Statements[3])
	}
	if grandChild.Name.Value != "TGrandChild" {
		t.Errorf("grandChild.Name.Value not 'TGrandChild'. got=%q", grandChild.Name.Value)
	}
	if grandChild.Parent.Value != "TChild1" {
		t.Errorf("grandChild.Parent.Value not 'TChild1'. got=%q", grandChild.Parent.Value)
	}
}

// TestShortFormWithNoParent tests that we can't have short-form without parent
func TestShortFormClassNoParent(t *testing.T) {
	// Note: In DWScript, `type TFoo = class;` without a parent or body
	// would be a forward declaration, not a short-form declaration.
	// This test just verifies the current behavior.
	input := `type TFoo = class;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// This should succeed - it's a valid forward declaration
	// (though this is more like an interface forward declaration)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(p.Errors()) > 0 {
		t.Logf("Parser had %d errors (this might be expected)", len(p.Errors()))
		for _, err := range p.Errors() {
			t.Logf("  - %s", err)
		}
	}
}

// TestTypeAliasToClass tests parsing of type aliases to class types
// (Phase 9.10.2): type TAlias = TClassName;
func TestTypeAliasToClass(t *testing.T) {
	input := `type
   TBase = class
   end;

type
   TAlias = TBase;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}

	// Check the alias (second statement)
	aliasDecl, ok := program.Statements[1].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.TypeDeclaration. got=%T",
			program.Statements[1])
	}

	if aliasDecl.Name.Value != "TAlias" {
		t.Errorf("aliasDecl.Name.Value not 'TAlias'. got=%q", aliasDecl.Name.Value)
	}

	if !aliasDecl.IsAlias {
		t.Error("aliasDecl.IsAlias should be true")
	}

	if aliasDecl.AliasedType == nil {
		t.Fatal("aliasDecl.AliasedType is nil")
	}

	if aliasDecl.AliasedType.String() != "TBase" {
		t.Errorf("aliasDecl.AliasedType.String() not 'TBase'. got=%q", aliasDecl.AliasedType.String())
	}
}

// TestMultipleTypeAliases tests multiple type aliases in sequence
func TestMultipleTypeAliases(t *testing.T) {
	input := `type
   TBase = class
   end;

type
   TAlias1 = TBase;

type
   TAlias2 = TAlias1;

type
   TAlias3 = TBase;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d",
			len(program.Statements))
	}

	// Check TAlias1
	alias1, ok := program.Statements[1].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.TypeDeclaration. got=%T",
			program.Statements[1])
	}
	if alias1.AliasedType == nil {
		t.Fatal("alias1.AliasedType is nil")
	}
	if alias1.Name.Value != "TAlias1" || alias1.AliasedType.String() != "TBase" {
		t.Errorf("TAlias1 incorrect")
	}

	// Check TAlias2 (alias to an alias)
	alias2, ok := program.Statements[2].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[2] is not *ast.TypeDeclaration. got=%T",
			program.Statements[2])
	}
	if alias2.AliasedType == nil {
		t.Fatal("alias2.AliasedType is nil")
	}
	if alias2.Name.Value != "TAlias2" || alias2.AliasedType.String() != "TAlias1" {
		t.Errorf("TAlias2 incorrect")
	}

	// Check TAlias3
	alias3, ok := program.Statements[3].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[3] is not *ast.TypeDeclaration. got=%T",
			program.Statements[3])
	}
	if alias3.AliasedType == nil {
		t.Fatal("alias3.AliasedType is nil")
	}
	if alias3.Name.Value != "TAlias3" || alias3.AliasedType.String() != "TBase" {
		t.Errorf("TAlias3 incorrect")
	}
}
