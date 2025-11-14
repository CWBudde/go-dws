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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"},
				},
				Name: &Identifier{
					Value: "TStringHelper",
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TStringHelper"},
				},
				ForType: &TypeAnnotation{
					Name: "String",
				},
				Methods: []*FunctionDecl{
					{
						Name: &Identifier{Value: "ToUpper"},
						ReturnType: &TypeAnnotation{
							Name: "String",
						},
						Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"},
				},
				Name: &Identifier{
					Value: "TIntHelper",
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TIntHelper"},
				},
				ForType: &TypeAnnotation{
					Name: "Integer",
				},
				Methods: []*FunctionDecl{
					{
						Name: &Identifier{Value: "IsEven"},
						ReturnType: &TypeAnnotation{
							Name: "Boolean",
						},
						Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"},
				},
				Name: &Identifier{
					Value: "TArrayHelper",
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TArrayHelper"},
				},
				ForType: &TypeAnnotation{
					Name: "TIntArray",
				},
				Properties: []*PropertyDecl{
					{
						BaseNode: BaseNode{
							Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
						},
						Name: &Identifier{Value: "Count"},
						Type: &TypeAnnotation{
							Name: "Integer",
						},
						Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"},
				},
				Name: &Identifier{
					Value: "THelper",
					Token: lexer.Token{Type: lexer.IDENT, Literal: "THelper"},
				},
				ForType: &TypeAnnotation{
					Name: "String",
				},
				ClassVars: []*FieldDecl{
					{
						Name: &Identifier{Value: "DefaultEncoding"},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"},
				},
				Name: &Identifier{
					Value: "TMathHelper",
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TMathHelper"},
				},
				ForType: &TypeAnnotation{
					Name: "Float",
				},
				ClassConsts: []*ConstDecl{
					{
						BaseNode: BaseNode{
							Token: lexer.Token{Type: lexer.CONST, Literal: "const"},
						},
						Name: &Identifier{Value: "PI"},
						Value: &FloatLiteral{
							Value: 3.14159,
							Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.14159"},
						},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.HELPER, Literal: "helper"},
				},
				Name: &Identifier{
					Value: "TComplexHelper",
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TComplexHelper"},
				},
				ForType: &TypeAnnotation{
					Name: "String",
				},
				PrivateMembers: []Statement{
					&FunctionDecl{
						Name: &Identifier{Value: "InternalMethod"},
						ReturnType: &TypeAnnotation{
							Name: "Integer",
						},
						Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
					},
				},
				PublicMembers: []Statement{
					&FunctionDecl{
						Name: &Identifier{Value: "ToUpper"},
						ReturnType: &TypeAnnotation{
							Name: "String",
						},
						Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
					},
					&PropertyDecl{
						BaseNode: BaseNode{
							Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
						},
						Name: &Identifier{Value: "Length"},
						Type: &TypeAnnotation{
							Name: "Integer",
						},
						Token: lexer.Token{Type: lexer.PROPERTY, Literal: "property"},
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
		BaseNode: BaseNode{
			Token: lexer.Token{
				Type:    lexer.HELPER,
				Literal: "helper",
				Pos:     lexer.Position{Line: 1, Column: 10},
			},
		},
		Name: &Identifier{
			Value: "TTestHelper",
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TTestHelper"},
		},
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
