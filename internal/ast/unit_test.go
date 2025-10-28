package ast

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestUsesClauseString(t *testing.T) {
	tests := []struct {
		name     string
		uses     *UsesClause
		expected string
	}{
		{
			name: "single unit",
			uses: &UsesClause{
				Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
				Units: []*Identifier{
					{Value: "System"},
				},
			},
			expected: "uses System;",
		},
		{
			name: "multiple units",
			uses: &UsesClause{
				Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
				Units: []*Identifier{
					{Value: "System"},
					{Value: "Math"},
					{Value: "Graphics"},
				},
			},
			expected: "uses System, Math, Graphics;",
		},
		{
			name: "two units",
			uses: &UsesClause{
				Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
				Units: []*Identifier{
					{Value: "System"},
					{Value: "SysUtils"},
				},
			},
			expected: "uses System, SysUtils;",
		},
		{
			name: "empty uses clause (edge case)",
			uses: &UsesClause{
				Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
				Units: []*Identifier{},
			},
			expected: "uses ;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.uses.String()
			if result != tt.expected {
				t.Errorf("UsesClause.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestUsesClauseNodeInterface(t *testing.T) {
	token := lexer.Token{
		Type:    lexer.USES,
		Literal: "uses",
		Pos:     lexer.Position{Line: 5, Column: 3},
	}

	uses := &UsesClause{
		Token: token,
		Units: []*Identifier{
			{Value: "System"},
		},
	}

	t.Run("TokenLiteral", func(t *testing.T) {
		if uses.TokenLiteral() != "uses" {
			t.Errorf("TokenLiteral() = %q, want %q", uses.TokenLiteral(), "uses")
		}
	})

	t.Run("Pos", func(t *testing.T) {
		pos := uses.Pos()
		if pos.Line != 5 || pos.Column != 3 {
			t.Errorf("Pos() = {Line: %d, Column: %d}, want {Line: 5, Column: 3}", pos.Line, pos.Column)
		}
	})

	t.Run("statementNode", func(t *testing.T) {
		// Just verify it implements Statement interface
		var _ Statement = uses
	})
}

func TestUnitDeclarationString(t *testing.T) {
	tests := []struct {
		name     string
		unit     *UnitDeclaration
		expected []string // Expected parts that should appear in output
	}{
		{
			name: "minimal unit",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "MyUnit"},
			},
			expected: []string{
				"unit MyUnit;",
				"end.",
			},
		},
		{
			name: "unit with interface section",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "MyLibrary"},
				InterfaceSection: &BlockStatement{
					Statements: []Statement{
						&UsesClause{
							Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
							Units: []*Identifier{
								{Value: "System"},
							},
						},
					},
				},
			},
			expected: []string{
				"unit MyLibrary;",
				"interface",
				"uses System;",
				"end.",
			},
		},
		{
			name: "unit with implementation section",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "TestUnit"},
				ImplementationSection: &BlockStatement{
					Statements: []Statement{
						&VarDeclStatement{
							Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
							Names: []*Identifier{{Value: "x"}},
							Type:  &TypeAnnotation{Name: "Integer"},
						},
					},
				},
			},
			expected: []string{
				"unit TestUnit;",
				"implementation",
				"var x: Integer",
				"end.",
			},
		},
		{
			name: "unit with interface and implementation",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "CompleteUnit"},
				InterfaceSection: &BlockStatement{
					Statements: []Statement{
						&UsesClause{
							Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
							Units: []*Identifier{
								{Value: "System"},
							},
						},
					},
				},
				ImplementationSection: &BlockStatement{
					Statements: []Statement{
						&VarDeclStatement{
							Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
							Names: []*Identifier{{Value: "count"}},
							Type:  &TypeAnnotation{Name: "Integer"},
						},
					},
				},
			},
			expected: []string{
				"unit CompleteUnit;",
				"interface",
				"uses System;",
				"implementation",
				"var count: Integer",
				"end.",
			},
		},
		{
			name: "unit with initialization",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "InitUnit"},
				InitSection: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: &Identifier{Value: "Setup"},
						},
					},
				},
			},
			expected: []string{
				"unit InitUnit;",
				"initialization",
				"Setup",
				"end.",
			},
		},
		{
			name: "unit with finalization",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "FinalUnit"},
				FinalSection: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: &Identifier{Value: "Cleanup"},
						},
					},
				},
			},
			expected: []string{
				"unit FinalUnit;",
				"finalization",
				"Cleanup",
				"end.",
			},
		},
		{
			name: "complete unit with all sections",
			unit: &UnitDeclaration{
				Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
				Name:  &Identifier{Value: "FullUnit"},
				InterfaceSection: &BlockStatement{
					Statements: []Statement{
						&UsesClause{
							Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
							Units: []*Identifier{
								{Value: "System"},
								{Value: "Math"},
							},
						},
					},
				},
				ImplementationSection: &BlockStatement{
					Statements: []Statement{
						&VarDeclStatement{
							Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
							Names: []*Identifier{{Value: "initialized"}},
							Type:  &TypeAnnotation{Name: "Boolean"},
						},
					},
				},
				InitSection: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: &Identifier{Value: "InitCode"},
						},
					},
				},
				FinalSection: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{
							Expression: &Identifier{Value: "FinalCode"},
						},
					},
				},
			},
			expected: []string{
				"unit FullUnit;",
				"interface",
				"uses System, Math;",
				"implementation",
				"var initialized: Boolean",
				"initialization",
				"InitCode",
				"finalization",
				"FinalCode",
				"end.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.unit.String()

			// Check that all expected parts appear in the output
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("UnitDeclaration.String() missing expected part %q\nGot:\n%s", expected, result)
				}
			}
		})
	}
}

func TestUnitDeclarationNodeInterface(t *testing.T) {
	token := lexer.Token{
		Type:    lexer.UNIT,
		Literal: "unit",
		Pos:     lexer.Position{Line: 1, Column: 1},
	}

	unit := &UnitDeclaration{
		Token: token,
		Name:  &Identifier{Value: "TestUnit"},
	}

	t.Run("TokenLiteral", func(t *testing.T) {
		if unit.TokenLiteral() != "unit" {
			t.Errorf("TokenLiteral() = %q, want %q", unit.TokenLiteral(), "unit")
		}
	})

	t.Run("Pos", func(t *testing.T) {
		pos := unit.Pos()
		if pos.Line != 1 || pos.Column != 1 {
			t.Errorf("Pos() = {Line: %d, Column: %d}, want {Line: 1, Column: 1}", pos.Line, pos.Column)
		}
	})

	t.Run("statementNode", func(t *testing.T) {
		// Just verify it implements Statement interface
		var _ Statement = unit
	})
}

func TestUnitDeclarationFields(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		unit := &UnitDeclaration{
			Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
			Name:  &Identifier{Value: "TestUnit"},
			InterfaceSection: &BlockStatement{
				Statements: []Statement{},
			},
			ImplementationSection: &BlockStatement{
				Statements: []Statement{},
			},
			InitSection: &BlockStatement{
				Statements: []Statement{},
			},
			FinalSection: &BlockStatement{
				Statements: []Statement{},
			},
		}

		if unit.Name.Value != "TestUnit" {
			t.Error("Name field not accessible")
		}

		if unit.InterfaceSection == nil {
			t.Error("InterfaceSection should not be nil")
		}

		if unit.ImplementationSection == nil {
			t.Error("ImplementationSection should not be nil")
		}

		if unit.InitSection == nil {
			t.Error("InitSection should not be nil")
		}

		if unit.FinalSection == nil {
			t.Error("FinalSection should not be nil")
		}
	})

	t.Run("optional sections can be nil", func(t *testing.T) {
		unit := &UnitDeclaration{
			Token: lexer.Token{Type: lexer.UNIT, Literal: "unit"},
			Name:  &Identifier{Value: "MinimalUnit"},
		}

		// These should be nil and String() should handle it
		if unit.InterfaceSection != nil {
			t.Error("InterfaceSection should be nil for minimal unit")
		}

		if unit.ImplementationSection != nil {
			t.Error("ImplementationSection should be nil for minimal unit")
		}

		if unit.InitSection != nil {
			t.Error("InitSection should be nil for minimal unit")
		}

		if unit.FinalSection != nil {
			t.Error("FinalSection should be nil for minimal unit")
		}

		// String() should work without panicking
		result := unit.String()
		if !strings.Contains(result, "unit MinimalUnit;") {
			t.Errorf("String() should handle nil sections gracefully, got: %s", result)
		}
	})
}

func TestUsesClauseFields(t *testing.T) {
	t.Run("units list accessible", func(t *testing.T) {
		uses := &UsesClause{
			Token: lexer.Token{Type: lexer.USES, Literal: "uses"},
			Units: []*Identifier{
				{Value: "System"},
				{Value: "Math"},
				{Value: "Graphics"},
			},
		}

		if len(uses.Units) != 3 {
			t.Errorf("expected 3 units, got %d", len(uses.Units))
		}

		if uses.Units[0].Value != "System" {
			t.Error("first unit should be System")
		}

		if uses.Units[1].Value != "Math" {
			t.Error("second unit should be Math")
		}

		if uses.Units[2].Value != "Graphics" {
			t.Error("third unit should be Graphics")
		}
	})
}
