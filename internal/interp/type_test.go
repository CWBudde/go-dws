package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Type Alias Runtime Tests (Task 9.21)
// ============================================================================

// TestTypeAliasBasicUsage tests that type aliases work at runtime
func TestTypeAliasBasicUsage(t *testing.T) {
	interp := New(new(bytes.Buffer))

	// Create AST:
	// type TUserID = Integer;
	// var id: TUserID := 42;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
				Name: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "id"},
					Value: "id",
				},
				Type: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
					Name:  "TUserID",
				},
				Value: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "42"},
					Value: 42,
				},
			},
		},
	}

	// Execute the program
	result := interp.Eval(program)

	// Should not return error
	if errVal, ok := result.(*ErrorValue); ok {
		t.Fatalf("Expected no error, got: %s", errVal.Message)
	}

	// Verify variable was created with correct value
	idVal, found := interp.env.Get("id")
	if !found {
		t.Fatal("Expected variable 'id' to be defined")
	}

	intVal, ok := idVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got: %T", idVal)
	}

	if intVal.Value != 42 {
		t.Errorf("Expected id to be 42, got: %d", intVal.Value)
	}
}

// TestTypeAliasResolveType tests that resolveType handles type aliases
func TestTypeAliasResolveType(t *testing.T) {
	interp := New(new(bytes.Buffer))

	// Create AST: type TUserID = Integer;
	typeDecl := &ast.TypeDeclaration{
		Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
			Value: "TUserID",
		},
		IsAlias: true,
		AliasedType: &ast.TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
			Name:  "Integer",
		},
	}

	// Execute the type declaration
	result := interp.Eval(typeDecl)
	if errVal, ok := result.(*ErrorValue); ok {
		t.Fatalf("Expected no error, got: %s", errVal.Message)
	}

	// Now resolve the type alias
	resolvedType, err := interp.resolveType("TUserID")
	if err != nil {
		t.Fatalf("Expected to resolve TUserID, got error: %v", err)
	}

	// Should resolve to Integer type
	if resolvedType.TypeKind() != "INTEGER" {
		t.Errorf("Expected TUserID to resolve to INTEGER, got: %s", resolvedType.TypeKind())
	}
}
