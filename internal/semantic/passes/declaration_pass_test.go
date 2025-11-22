package passes_test

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestDeclarationPass_ClassRegistration(t *testing.T) {
	// Create a simple program with a class declaration
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClassDecl{
				Name: &ast.Identifier{Value: "TFoo"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
				Fields:     []*ast.FieldDecl{}, // Not nil, so not a forward declaration
				Methods:    []*ast.FunctionDecl{},
				Properties: []*ast.PropertyDecl{},
			},
		},
	}

	// Create pass context
	ctx := semantic.NewPassContext()

	// Run Pass 1
	pass := semantic.NewDeclarationPass()
	err := pass.Run(program, ctx)

	if err != nil {
		t.Fatalf("Pass 1 failed: %v", err)
	}

	// Verify class was registered
	classType, ok := ctx.TypeRegistry.Resolve("TFoo")
	if !ok {
		t.Fatal("Class 'TFoo' was not registered")
	}

	// Verify it's a ClassType
	ct, ok := classType.(*types.ClassType)
	if !ok {
		t.Fatalf("Expected ClassType, got %T", classType)
	}

	// Verify properties
	if ct.Name != "TFoo" {
		t.Errorf("Expected name 'TFoo', got '%s'", ct.Name)
	}

	if ct.IsForward {
		t.Error("Expected IsForward=false, got true")
	}

	if ct.Parent != nil {
		t.Errorf("Expected Parent=nil, got %v", ct.Parent)
	}
}

func TestDeclarationPass_ForwardDeclaration(t *testing.T) {
	// Create a program with a forward declaration
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
				Operators:  nil,
				Constants:  nil,
			},
		},
	}

	// Create pass context
	ctx := semantic.NewPassContext()

	// Run Pass 1
	pass := semantic.NewDeclarationPass()
	err := pass.Run(program, ctx)

	if err != nil {
		t.Fatalf("Pass 1 failed: %v", err)
	}

	// Verify forward declaration was registered
	classType, ok := ctx.TypeRegistry.Resolve("TForward")
	if !ok {
		t.Fatal("Forward-declared class 'TForward' was not registered")
	}

	// Verify it's marked as forward
	ct, ok := classType.(*types.ClassType)
	if !ok {
		t.Fatalf("Expected ClassType, got %T", classType)
	}

	if !ct.IsForward {
		t.Error("Expected IsForward=true, got false")
	}
}

func TestDeclarationPass_EnumRegistration(t *testing.T) {
	// Create a program with an enum declaration
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.EnumDecl{
				Name: &ast.Identifier{Value: "TColor"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
				Values: []ast.EnumValue{
					{Name: "Red", Value: nil},
					{Name: "Green", Value: nil},
					{Name: "Blue", Value: nil},
				},
				Scoped: false,
				Flags:  false,
			},
		},
	}

	// Create pass context
	ctx := semantic.NewPassContext()

	// Run Pass 1
	pass := semantic.NewDeclarationPass()
	err := pass.Run(program, ctx)

	if err != nil {
		t.Fatalf("Pass 1 failed: %v", err)
	}

	// Verify enum was registered
	enumType, ok := ctx.TypeRegistry.Resolve("TColor")
	if !ok {
		t.Fatal("Enum 'TColor' was not registered")
	}

	// Verify it's an EnumType
	et, ok := enumType.(*types.EnumType)
	if !ok {
		t.Fatalf("Expected EnumType, got %T", enumType)
	}

	// Verify enum values
	if len(et.Values) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(et.Values))
	}

	// Verify auto-assigned values (0, 1, 2)
	if et.Values["Red"] != 0 {
		t.Errorf("Expected Red=0, got %d", et.Values["Red"])
	}
	if et.Values["Green"] != 1 {
		t.Errorf("Expected Green=1, got %d", et.Values["Green"])
	}
	if et.Values["Blue"] != 2 {
		t.Errorf("Expected Blue=2, got %d", et.Values["Blue"])
	}
}

func TestDeclarationPass_DuplicateError(t *testing.T) {
	// Create a program with duplicate class declarations
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClassDecl{
				Name: &ast.Identifier{Value: "TFoo"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
				Fields: []*ast.FieldDecl{}, // Not a forward declaration
			},
			&ast.ClassDecl{
				Name: &ast.Identifier{Value: "TFoo"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 5, Column: 1}},
				},
				Fields: []*ast.FieldDecl{}, // Not a forward declaration
			},
		},
	}

	// Create pass context
	ctx := semantic.NewPassContext()

	// Run Pass 1
	pass := semantic.NewDeclarationPass()
	_ = pass.Run(program, ctx)

	// Verify error was collected
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected duplicate declaration error, got none")
	}

	// Verify error message mentions duplicate
	foundDuplicate := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "TFoo") && contains(errMsg, "already declared") {
			foundDuplicate = true
			break
		}
	}

	if !foundDuplicate {
		t.Errorf("Expected error about duplicate 'TFoo', got: %v", ctx.Errors)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
