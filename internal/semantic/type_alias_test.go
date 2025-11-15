package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Type Alias Tests
// ============================================================================

// TestTypeAliasRegistration tests that type aliases are registered in the analyzer
func TestTypeAliasRegistration(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create AST: type TUserID = Integer;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
		},
	}

	// Analyze the program
	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("Expected no errors, got: %v", err)
	}

	// Verify type alias was registered
	resolvedType, err := analyzer.resolveType("TUserID")
	if err != nil {
		t.Fatalf("Expected TUserID to be registered, but got error: %v", err)
	}

	// Should resolve to Integer type
	if resolvedType.TypeKind() != "INTEGER" {
		t.Errorf("Expected TUserID to resolve to INTEGER, got: %s", resolvedType.TypeKind())
	}
}

// TestTypeAliasInVariableDeclaration tests using a type alias in a variable declaration
func TestTypeAliasInVariableDeclaration(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create AST:
	// type TUserID = Integer;
	// var id: TUserID;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.VAR, Literal: "var"}},
				Names: []*ast.Identifier{{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "id"},
						},
					},
					Value: "id",
				}},
				Type: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
					Name:  "TUserID",
				},
				Value: nil,
			},
		},
	}

	// Analyze the program
	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("Expected no errors, got: %v", err)
	}

	// Verify variable was declared with the alias type
	varSymbol, found := analyzer.symbols.Resolve("id")
	if !found {
		t.Fatal("Expected variable 'id' to be declared")
	}

	// The variable should have the alias type, which resolves to INTEGER
	if varSymbol.Type.TypeKind() != "INTEGER" {
		t.Errorf("Expected variable 'id' to have INTEGER type, got: %s", varSymbol.Type.TypeKind())
	}
}

// TestTypeAliasCompatibility tests that type aliases are compatible with their underlying types
func TestTypeAliasCompatibility(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create AST:
	// type TUserID = Integer;
	// var id: TUserID := 42;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.VAR, Literal: "var"}},
				Names: []*ast.Identifier{{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "id"},
						},
					},
					Value: "id",
				}},
				Type: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
					Name:  "TUserID",
				},
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
		},
	}

	// Analyze the program
	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("Expected no errors when assigning Integer to TUserID, got: %v", err)
	}
}

// TestTypeAliasUndefinedType tests error handling for undefined aliased types
func TestTypeAliasUndefinedType(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create AST: type TUserID = UndefinedType;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 1, Column: 1}},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "UndefinedType"},
					Name:  "UndefinedType",
				},
			},
		},
	}

	// Analyze the program
	err := analyzer.Analyze(program)
	if err == nil {
		t.Fatal("Expected error for undefined aliased type, got none")
	}

	// Verify the error message
	if len(analyzer.Errors()) == 0 {
		t.Fatal("Expected analyzer to accumulate errors")
	}

	errorMsg := analyzer.Errors()[0]
	if !containsSubstring(errorMsg, "unknown type") && !containsSubstring(errorMsg, "UndefinedType") {
		t.Errorf("Expected error about unknown type 'UndefinedType', got: %s", errorMsg)
	}
}

// TestTypeAliasNestedAliases tests chained type aliases
func TestTypeAliasNestedAliases(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create AST:
	// type TInt = Integer;
	// type TUserID = TInt;
	// var id: TUserID;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TInt"},
						},
					},
					Value: "TInt",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TInt"},
					Name:  "TInt",
				},
			},
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.VAR, Literal: "var"}},
				Names: []*ast.Identifier{{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "id"},
						},
					},
					Value: "id",
				}},
				Type: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
					Name:  "TUserID",
				},
				Value: nil,
			},
		},
	}

	// Analyze the program
	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("Expected no errors for nested type aliases, got: %v", err)
	}

	// Verify the nested alias resolves correctly
	varSymbol, found := analyzer.symbols.Resolve("id")
	if !found {
		t.Fatal("Expected variable 'id' to be declared")
	}

	if varSymbol.Type.TypeKind() != "INTEGER" {
		t.Errorf("Expected nested alias to resolve to INTEGER, got: %s", varSymbol.Type.TypeKind())
	}
}

// TestTypeAliasDuplicateDeclaration tests error handling for duplicate type declarations
func TestTypeAliasDuplicateDeclaration(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create AST:
	// type TUserID = Integer;
	// type TUserID = String;  // Duplicate!
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			&ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: lexer.Position{Line: 2, Column: 1}},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "TUserID"},
						},
					},
					Value: "TUserID",
				},
				IsAlias: true,
				AliasedType: &ast.TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
					Name:  "String",
				},
			},
		},
	}

	// Analyze the program
	err := analyzer.Analyze(program)
	if err == nil {
		t.Fatal("Expected error for duplicate type declaration, got none")
	}

	// Verify error message
	if len(analyzer.Errors()) == 0 {
		t.Fatal("Expected analyzer to accumulate errors")
	}

	errorMsg := analyzer.Errors()[0]
	if !containsSubstring(errorMsg, "already declared") {
		t.Errorf("Expected error about duplicate type declaration, got: %s", errorMsg)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
