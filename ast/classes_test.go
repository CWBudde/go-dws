package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/lexer"
)

// ============================================================================
// ClassDecl Tests (Task 7.7)
// ============================================================================

func TestClassDeclString(t *testing.T) {
	tests := []struct {
		name      string
		classDecl *ClassDecl
		expected  string
	}{
		{
			name: "simple class without parent",
			classDecl: &ClassDecl{
				Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
					Value: "TPoint",
				},
				Parent:  nil,
				Fields:  []*FieldDecl{},
				Methods: []*FunctionDecl{},
			},
			expected: "type TPoint = class\nend",
		},
		{
			name: "class with parent",
			classDecl: &ClassDecl{
				Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TChild"},
					Value: "TChild",
				},
				Parent: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TParent"},
					Value: "TParent",
				},
				Fields:  []*FieldDecl{},
				Methods: []*FunctionDecl{},
			},
			expected: "type TChild = class(TParent)\nend",
		},
		{
			name: "class with fields",
			classDecl: &ClassDecl{
				Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
					Value: "TPoint",
				},
				Parent: nil,
				Fields: []*FieldDecl{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "X"},
							Value: "X",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
						Visibility: VisibilityPublic,
					},
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Y"},
							Value: "Y",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
						Visibility: VisibilityPublic,
					},
				},
				Methods: []*FunctionDecl{},
			},
			expected: "type TPoint = class\n  X: Integer;\n  Y: Integer;\nend",
		},
		{
			name: "class with method",
			classDecl: &ClassDecl{
				Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TCounter"},
					Value: "TCounter",
				},
				Parent: nil,
				Fields: []*FieldDecl{},
				Methods: []*FunctionDecl{
					{
						Token: lexer.Token{Type: lexer.FUNCTION, Literal: "function"},
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
							Value: "GetValue",
						},
						Parameters: []*Parameter{},
						ReturnType: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
						Body: &BlockStatement{
							Token:      lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
							Statements: []Statement{},
						},
					},
				},
			},
			expected: "type TCounter = class\n  function GetValue(): Integer begin\n  end;\nend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.classDecl.String()
			if result != tt.expected {
				t.Errorf("ClassDecl.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestClassDeclTokenLiteral(t *testing.T) {
	classDecl := &ClassDecl{
		Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TTest"},
			Value: "TTest",
		},
	}

	expected := "type"
	result := classDecl.TokenLiteral()
	if result != expected {
		t.Errorf("ClassDecl.TokenLiteral() = %q, want %q", result, expected)
	}
}

func TestClassDeclPos(t *testing.T) {
	pos := lexer.Position{Line: 10, Column: 5, Offset: 100}
	classDecl := &ClassDecl{
		Token: lexer.Token{Type: lexer.TYPE, Literal: "type", Pos: pos},
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "TTest"},
			Value: "TTest",
		},
	}

	result := classDecl.Pos()
	if result != pos {
		t.Errorf("ClassDecl.Pos() = %v, want %v", result, pos)
	}
}

// ============================================================================
// FieldDecl Tests (Task 7.8)
// ============================================================================

func TestFieldDeclString(t *testing.T) {
	tests := []struct {
		name      string
		fieldDecl *FieldDecl
		expected  string
	}{
		{
			name: "public field",
			fieldDecl: &FieldDecl{
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "X"},
					Value: "X",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
				Visibility: VisibilityPublic,
			},
			expected: "X: Integer",
		},
		{
			name: "private field",
			fieldDecl: &FieldDecl{
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "FValue"},
					Value: "FValue",
				},
				Type: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "String"},
					Name:  "String",
				},
				Visibility: VisibilityPrivate,
			},
			expected: "FValue: String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldDecl.String()
			if result != tt.expected {
				t.Errorf("FieldDecl.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFieldDeclMethods(t *testing.T) {
	pos := lexer.Position{Line: 5, Column: 10, Offset: 50}
	fieldDecl := &FieldDecl{
		Name: &Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "X", Pos: pos},
			Value: "X",
		},
		Type: &TypeAnnotation{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
			Name:  "Integer",
		},
		Visibility: VisibilityPublic,
	}

	// Test TokenLiteral
	if got := fieldDecl.TokenLiteral(); got != "X" {
		t.Errorf("FieldDecl.TokenLiteral() = %q, want %q", got, "X")
	}

	// Test Pos
	if got := fieldDecl.Pos(); got != pos {
		t.Errorf("FieldDecl.Pos() = %v, want %v", got, pos)
	}
}

// ============================================================================
// NewExpression Tests (Task 7.9)
// ============================================================================

func TestNewExpressionString(t *testing.T) {
	tests := []struct {
		name     string
		newExpr  *NewExpression
		expected string
	}{
		{
			name: "new without arguments",
			newExpr: &NewExpression{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
				ClassName: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
					Value: "TPoint",
				},
				Arguments: []Expression{},
			},
			expected: "TPoint.Create()",
		},
		{
			name: "new with arguments",
			newExpr: &NewExpression{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
				ClassName: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
					Value: "TPoint",
				},
				Arguments: []Expression{
					&IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "10"},
						Value: 10,
					},
					&IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "20"},
						Value: 20,
					},
				},
			},
			expected: "TPoint.Create(10, 20)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.newExpr.String()
			if result != tt.expected {
				t.Errorf("NewExpression.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// MemberAccessExpression Tests (Task 7.10)
// ============================================================================

func TestMemberAccessString(t *testing.T) {
	tests := []struct {
		name      string
		memAccess *MemberAccessExpression
		expected  string
	}{
		{
			name: "simple field access",
			memAccess: &MemberAccessExpression{
				Token: lexer.Token{Type: lexer.DOT, Literal: "."},
				Object: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "point"},
					Value: "point",
				},
				Member: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "X"},
					Value: "X",
				},
			},
			expected: "point.X",
		},
		{
			name: "chained member access",
			memAccess: &MemberAccessExpression{
				Token: lexer.Token{Type: lexer.DOT, Literal: "."},
				Object: &MemberAccessExpression{
					Token: lexer.Token{Type: lexer.DOT, Literal: "."},
					Object: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "obj"},
						Value: "obj",
					},
					Member: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "field1"},
						Value: "field1",
					},
				},
				Member: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "field2"},
					Value: "field2",
				},
			},
			expected: "obj.field1.field2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.memAccess.String()
			if result != tt.expected {
				t.Errorf("MemberAccessExpression.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// MethodCallExpression Tests (Task 7.11)
// ============================================================================

func TestMethodCallString(t *testing.T) {
	tests := []struct {
		name       string
		methodCall *MethodCallExpression
		expected   string
	}{
		{
			name: "method call without arguments",
			methodCall: &MethodCallExpression{
				Token: lexer.Token{Type: lexer.DOT, Literal: "."},
				Object: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "obj"},
					Value: "obj",
				},
				Method: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "DoSomething"},
					Value: "DoSomething",
				},
				Arguments: []Expression{},
			},
			expected: "obj.DoSomething()",
		},
		{
			name: "method call with arguments",
			methodCall: &MethodCallExpression{
				Token: lexer.Token{Type: lexer.DOT, Literal: "."},
				Object: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "point"},
					Value: "point",
				},
				Method: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "MoveTo"},
					Value: "MoveTo",
				},
				Arguments: []Expression{
					&IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "10"},
						Value: 10,
					},
					&IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "20"},
						Value: 20,
					},
				},
			},
			expected: "point.MoveTo(10, 20)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.methodCall.String()
			if result != tt.expected {
				t.Errorf("MethodCallExpression.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
