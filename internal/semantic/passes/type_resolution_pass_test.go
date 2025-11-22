package passes_test

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestTypeResolutionPass_ClassHierarchy(t *testing.T) {
	// Create a program with a class hierarchy
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClassDecl{
				Name: &ast.Identifier{Value: "TBase"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
				Fields: []*ast.FieldDecl{}, // Not a forward declaration
			},
			&ast.ClassDecl{
				Name:   &ast.Identifier{Value: "TDerived"},
				Parent: &ast.Identifier{Value: "TBase"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 3, Column: 1}},
				},
				Fields: []*ast.FieldDecl{},
			},
		},
	}

	// Run Pass 1
	ctx := semantic.NewPassContext()
	pass1 := semantic.NewDeclarationPass()
	err := pass1.Run(program, ctx)
	if err != nil {
		t.Fatalf("Pass 1 failed: %v", err)
	}

	// Run Pass 2
	pass2 := semantic.NewTypeResolutionPass()
	err = pass2.Run(program, ctx)
	if err != nil {
		t.Fatalf("Pass 2 failed: %v", err)
	}

	// Verify TBase has no parent
	baseType, ok := ctx.TypeRegistry.Resolve("TBase")
	if !ok {
		t.Fatal("TBase not found in registry")
	}
	baseClass, ok := baseType.(*types.ClassType)
	if !ok {
		t.Fatalf("TBase is not a ClassType, got %T", baseType)
	}
	if baseClass.Parent != nil {
		t.Errorf("Expected TBase.Parent=nil, got %v", baseClass.Parent)
	}

	// Verify TDerived has TBase as parent
	derivedType, ok := ctx.TypeRegistry.Resolve("TDerived")
	if !ok {
		t.Fatal("TDerived not found in registry")
	}
	derivedClass, ok := derivedType.(*types.ClassType)
	if !ok {
		t.Fatalf("TDerived is not a ClassType, got %T", derivedType)
	}
	if derivedClass.Parent == nil {
		t.Fatal("Expected TDerived.Parent to be set")
	}
	if derivedClass.Parent.Name != "TBase" {
		t.Errorf("Expected TDerived.Parent=TBase, got %s", derivedClass.Parent.Name)
	}
}

func TestTypeResolutionPass_CircularDependency(t *testing.T) {
	// Create a program with circular class hierarchy
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClassDecl{
				Name:   &ast.Identifier{Value: "TA"},
				Parent: &ast.Identifier{Value: "TB"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
				Fields: []*ast.FieldDecl{},
			},
			&ast.ClassDecl{
				Name:   &ast.Identifier{Value: "TB"},
				Parent: &ast.Identifier{Value: "TA"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 3, Column: 1}},
				},
				Fields: []*ast.FieldDecl{},
			},
		},
	}

	// Run Pass 1
	ctx := semantic.NewPassContext()
	pass1 := semantic.NewDeclarationPass()
	_ = pass1.Run(program, ctx)

	// Run Pass 2
	pass2 := semantic.NewTypeResolutionPass()
	_ = pass2.Run(program, ctx)

	// Verify circular dependency error was detected
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected circular dependency error, got none")
	}

	foundCircular := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "circular dependency") {
			foundCircular = true
			break
		}
	}

	if !foundCircular {
		t.Errorf("Expected circular dependency error, got: %v", ctx.Errors)
	}
}

func TestTypeResolutionPass_ForwardDeclarationValidation(t *testing.T) {
	// Create a program with an unresolved forward declaration
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClassDecl{
				Name: &ast.Identifier{Value: "TForward"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
				// All nil -> forward declaration
				Fields:     nil,
				Methods:    nil,
				Properties: nil,
			},
		},
	}

	// Run Pass 1
	ctx := semantic.NewPassContext()
	pass1 := semantic.NewDeclarationPass()
	_ = pass1.Run(program, ctx)

	// Run Pass 2
	pass2 := semantic.NewTypeResolutionPass()
	_ = pass2.Run(program, ctx)

	// Verify forward declaration error was detected
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected forward declaration error, got none")
	}

	foundForward := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "forward-declared") && contains(errMsg, "TForward") {
			foundForward = true
			break
		}
	}

	if !foundForward {
		t.Errorf("Expected forward declaration error, got: %v", ctx.Errors)
	}
}

func TestTypeResolutionPass_InterfaceHierarchy(t *testing.T) {
	// Create a program with an interface hierarchy
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.InterfaceDecl{
				Name: &ast.Identifier{Value: "IBase"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
			&ast.InterfaceDecl{
				Name:   &ast.Identifier{Value: "IDerived"},
				Parent: &ast.Identifier{Value: "IBase"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 3, Column: 1}},
				},
			},
		},
	}

	// Run Pass 1
	ctx := semantic.NewPassContext()
	pass1 := semantic.NewDeclarationPass()
	err := pass1.Run(program, ctx)
	if err != nil {
		t.Fatalf("Pass 1 failed: %v", err)
	}

	// Run Pass 2
	pass2 := semantic.NewTypeResolutionPass()
	err = pass2.Run(program, ctx)
	if err != nil {
		t.Fatalf("Pass 2 failed: %v", err)
	}

	// Verify IBase has no parent
	baseType, ok := ctx.TypeRegistry.Resolve("IBase")
	if !ok {
		t.Fatal("IBase not found in registry")
	}
	baseInterface, ok := baseType.(*types.InterfaceType)
	if !ok {
		t.Fatalf("IBase is not an InterfaceType, got %T", baseType)
	}
	if baseInterface.Parent != nil {
		t.Errorf("Expected IBase.Parent=nil, got %v", baseInterface.Parent)
	}

	// Verify IDerived has IBase as parent
	derivedType, ok := ctx.TypeRegistry.Resolve("IDerived")
	if !ok {
		t.Fatal("IDerived not found in registry")
	}
	derivedInterface, ok := derivedType.(*types.InterfaceType)
	if !ok {
		t.Fatalf("IDerived is not an InterfaceType, got %T", derivedType)
	}
	if derivedInterface.Parent == nil {
		t.Fatal("Expected IDerived.Parent to be set")
	}
	if derivedInterface.Parent.Name != "IBase" {
		t.Errorf("Expected IDerived.Parent=IBase, got %s", derivedInterface.Parent.Name)
	}
}

func TestTypeResolutionPass_BuiltinTypesRegistered(t *testing.T) {
	// Create an empty program
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	// Run Pass 2
	ctx := semantic.NewPassContext()
	pass2 := semantic.NewTypeResolutionPass()
	err := pass2.Run(program, ctx)
	if err != nil {
		t.Fatalf("Pass 2 failed: %v", err)
	}

	// Verify built-in types are registered
	builtinNames := []string{"Integer", "Float", "String", "Boolean", "Variant", "Void"}
	for _, name := range builtinNames {
		if !ctx.TypeRegistry.Has(name) {
			t.Errorf("Expected built-in type '%s' to be registered", name)
		}
	}
}
