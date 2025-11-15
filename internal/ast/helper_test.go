package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestHelperDeclString(t *testing.T) {
	tests := []struct {
		name     string
		helper   *HelperDecl
		expected string
	}{
		{
			name: "simple record helper with method",
			helper: &HelperDecl{
				BaseNode: NewTestBaseNode(lexer.HELPER, "helper"),
				Name:     NewTestIdentifier("TStringHelper"),
				ForType: &TypeAnnotation{
					Name: "String",
				},
				Methods: []*FunctionDecl{
					{
						BaseNode: NewTestBaseNode(lexer.FUNCTION, "function"),
						Name:     &Identifier{Value: "ToUpper"},
						ReturnType: &TypeAnnotation{
							Name: "String",
						},
					},
				},
				IsRecordHelper: true,
			},
			expected: `type TStringHelper = record helper for String
  function ToUpper(): String;
end`,
		},
		{
			name: "simple helper without record keyword",
			helper: &HelperDecl{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"}},
				Name:     NewTestIdentifier("TIntHelper"),
				ForType: &TypeAnnotation{
					Name: "Integer",
				},
				Methods: []*FunctionDecl{
					{
						BaseNode: BaseNode{Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"}},
						Name:     &Identifier{Value: "IsEven"},
						ReturnType: &TypeAnnotation{
							Name: "Boolean",
						},
					},
				},
				IsRecordHelper: false,
			},
			expected: `type TIntHelper = helper for Integer
  function IsEven(): Boolean;
end`,
		},
		{
			name: "helper with property",
			helper: &HelperDecl{
				BaseNode: NewTestBaseNode(lexer.HELPER, "helper"),
				Name:     NewTestIdentifier("TArrayHelper"),
				ForType: &TypeAnnotation{
					Name: "TIntArray",
				},
				Properties: []*PropertyDecl{
					{
						BaseNode: NewTestBaseNode(lexer.PROPERTY, "property"),
						Name:     &Identifier{Value: "Count"},
						Type: &TypeAnnotation{
							Name: "Integer",
						},
					},
				},
				IsRecordHelper: false,
			},
			expected: `type TArrayHelper = helper for TIntArray
  property Count: Integer;;
end`,
		},
		{
			name: "helper with class var",
			helper: &HelperDecl{
				BaseNode: NewTestBaseNode(lexer.HELPER, "helper"),
				Name:     NewTestIdentifier("THelper"),
				ForType: &TypeAnnotation{
					Name: "String",
				},
				ClassVars: []*FieldDecl{
					{
						BaseNode: NewTestBaseNode(lexer.IDENT, "DefaultEncoding"),
						Name:     &Identifier{Value: "DefaultEncoding"},
						Type: &TypeAnnotation{
							Name: "String",
						},
						IsClassVar: true,
					},
				},
				IsRecordHelper: true,
			},
			expected: `type THelper = record helper for String
  class var DefaultEncoding: String;
end`,
		},
		{
			name: "helper with class const",
			helper: &HelperDecl{
				BaseNode: NewTestBaseNode(lexer.HELPER, "helper"),
				Name:     NewTestIdentifier("TMathHelper"),
				ForType: &TypeAnnotation{
					Name: "Float",
				},
				ClassConsts: []*ConstDecl{
					{
						BaseNode: NewTestBaseNode(lexer.CONST, "const"),
						Name:     &Identifier{Value: "PI"},
						Value:    &FloatLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: NewTestBaseNode(lexer.FLOAT, "3.14159")}, Value: 3.14159},
					},
				},
				IsRecordHelper: false,
			},
			expected: `type TMathHelper = helper for Float
  class const PI = 3.14159;
end`,
		},
		{
			name: "helper with private and public sections",
			helper: &HelperDecl{
				BaseNode: NewTestBaseNode(lexer.HELPER, "helper"),
				Name:     NewTestIdentifier("TComplexHelper"),
				ForType: &TypeAnnotation{
					Name: "String",
				},
				PrivateMembers: []Statement{
					&FunctionDecl{
						BaseNode: NewTestBaseNode(lexer.FUNCTION, "function"),
						Name:     &Identifier{Value: "InternalMethod"},
						ReturnType: &TypeAnnotation{
							Name: "Integer",
						},
					},
				},
				PublicMembers: []Statement{
					&FunctionDecl{
						BaseNode: NewTestBaseNode(lexer.FUNCTION, "function"),
						Name:     &Identifier{Value: "ToUpper"},
						ReturnType: &TypeAnnotation{
							Name: "String",
						},
					},
					&PropertyDecl{
						BaseNode: NewTestBaseNode(lexer.PROPERTY, "property"),
						Name:     &Identifier{Value: "Length"},
						Type: &TypeAnnotation{
							Name: "Integer",
						},
					},
				},
				IsRecordHelper: true,
			},
			expected: `type TComplexHelper = record helper for String
  private
    function InternalMethod(): Integer;
  public
    function ToUpper(): String;
    property Length: Integer;;
end`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.helper.String()
			if result != tt.expected {
				t.Errorf("HelperDecl.String() failed\nExpected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestHelperDeclNodeInterface(t *testing.T) {
	helper := &HelperDecl{
		BaseNode: BaseNode{Token: lexer.Token{
			Type:    lexer.HELPER,
			Literal: "helper",
			Pos:     lexer.Position{Line: 1, Column: 10},
		}},
		Name: NewTestIdentifier("TTestHelper"),
	}

	// Test that HelperDecl implements Statement interface
	var _ Statement = helper

	// Test TokenLiteral
	if helper.TokenLiteral() != "helper" {
		t.Errorf("TokenLiteral() = %q, want %q", helper.TokenLiteral(), "helper")
	}

	// Test Pos
	pos := helper.Pos()
	if pos.Line != 1 || pos.Column != 10 {
		t.Errorf("Pos() = {Line: %d, Column: %d}, want {Line: 1, Column: 10}", pos.Line, pos.Column)
	}
}
