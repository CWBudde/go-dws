package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// ClassDecl Tests
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
				BaseNode: NewTestBaseNode(lexer.TYPE, "type"),
				Name:     NewTestIdentifier("TPoint"),
				Parent:   nil,
				Fields:   []*FieldDecl{},
				Methods:  []*FunctionDecl{},
			},
			expected: "type TPoint = class\nend",
		},
		{
			name: "class with parent",
			classDecl: &ClassDecl{
				BaseNode: NewTestBaseNode(lexer.TYPE, "type"),
				Name:     NewTestIdentifier("TChild"),
				Parent:   NewTestIdentifier("TParent"),
				Fields:   []*FieldDecl{},
				Methods:  []*FunctionDecl{},
			},
			expected: "type TChild = class(TParent)\nend",
		},
		{
			name: "class with fields",
			classDecl: &ClassDecl{
				BaseNode: NewTestBaseNode(lexer.TYPE, "type"),
				Name:     NewTestIdentifier("TPoint"),
				Parent:   nil,
				Fields: []*FieldDecl{
					{
						Name:       NewTestIdentifier("X"),
						Type:       NewTestTypeAnnotation("Integer"),
						Visibility: VisibilityPublic,
					},
					{
						Name:       NewTestIdentifier("Y"),
						Type:       NewTestTypeAnnotation("Integer"),
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
				BaseNode: NewTestBaseNode(lexer.TYPE, "type"),
				Name:     NewTestIdentifier("TCounter"),
				Parent:   nil,
				Fields:   []*FieldDecl{},
				Methods: []*FunctionDecl{
					{
						BaseNode:   NewTestBaseNode(lexer.FUNCTION, "function"),
						Name:       NewTestIdentifier("GetValue"),
						Parameters: []*Parameter{},
						ReturnType: NewTestTypeAnnotation("Integer"),
						Body:       NewTestBlockStatement([]Statement{}),
					},
				},
			},
			expected: "type TCounter = class\n  function GetValue(): Integer begin\n  end;\nend",
		},
		{
			name: "class with operator",
			classDecl: &ClassDecl{
				BaseNode: NewTestBaseNode(lexer.TYPE, "type"),
				Name:     NewTestIdentifier("TStream"),
				Parent:   nil,
				Fields:   []*FieldDecl{},
				Methods:  []*FunctionDecl{},
				Operators: []*OperatorDecl{
					{
						BaseNode:       NewTestBaseNode(lexer.CLASS, "class"),
						Kind:           OperatorKindClass,
						OperatorToken:  NewTestToken(lexer.LESS_LESS, "<<"),
						OperatorSymbol: "<<",
						Arity:          1,
						OperandTypes: []*TypeAnnotation{
							NewTestTypeAnnotation("String"),
						},
						Binding: NewTestIdentifier("Append"),
					},
				},
			},
			expected: "type TStream = class\n  class operator << String uses Append;\nend",
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
	classDecl := NewTestClassDecl("TTest", nil)

	expected := "type"
	result := classDecl.TokenLiteral()
	if result != expected {
		t.Errorf("ClassDecl.TokenLiteral() = %q, want %q", result, expected)
	}
}

func TestClassDeclPos(t *testing.T) {
	pos := lexer.Position{Line: 10, Column: 5, Offset: 100}
	classDecl := &ClassDecl{
		BaseNode: BaseNode{
			Token: lexer.Token{
				Type:    lexer.TYPE,
				Literal: "type",
				Pos:     pos,
			},
		},
	}

	result := classDecl.Pos()
	if result != pos {
		t.Errorf("ClassDecl.Pos() = %v, want %v", result, pos)
	}
}

// ============================================================================
// FieldDecl Tests
// ============================================================================

func TestFieldDeclString(t *testing.T) {
	tests := []struct {
		name      string
		fieldDecl *FieldDecl
		expected  string
	}{
		{
			name:      "public field",
			fieldDecl: NewTestFieldDecl("X", "Integer", VisibilityPublic),
			expected:  "X: Integer",
		},
		{
			name:      "private field",
			fieldDecl: NewTestFieldDecl("FValue", "String", VisibilityPrivate),
			expected:  "FValue: String",
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
			TypedExpressionBase: TypedExpressionBase{
				BaseNode: BaseNode{
					Token: lexer.Token{
						Type:    lexer.IDENT,
						Literal: "X",
						Pos:     pos,
					},
				},
			},
			Value: "X",
		},
		Type: &TypeAnnotation{
			Token: NewTestToken(lexer.IDENT, "Integer"),
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
// NewExpression Tests
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
				Token:     NewTestToken(lexer.IDENT, "TPoint"),
				ClassName: NewTestIdentifier("TPoint"),
				Arguments: []Expression{},
			},
			expected: "TPoint.Create()",
		},
		{
			name: "new with arguments",
			newExpr: &NewExpression{
				Token:     lexer.Token{Type: lexer.IDENT, Literal: "TPoint"},
				ClassName: NewTestIdentifier("TPoint"),
				Arguments: []Expression{
					NewTestIntegerLiteral(10),
					NewTestIntegerLiteral(20),
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
// MemberAccessExpression Tests
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
				Token:  NewTestToken(lexer.DOT, "."),
				Object: NewTestIdentifier("point"),
				Member: NewTestIdentifier("X"),
			},
			expected: "point.X",
		},
		{
			name: "chained member access",
			memAccess: &MemberAccessExpression{
				Token: NewTestToken(lexer.DOT, "."),
				Object: &MemberAccessExpression{
					Token:  NewTestToken(lexer.DOT, "."),
					Object: NewTestIdentifier("obj"),
					Member: NewTestIdentifier("field1"),
				},
				Member: NewTestIdentifier("field2"),
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
// MethodCallExpression Tests
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
				Token:     NewTestToken(lexer.DOT, "."),
				Object:    NewTestIdentifier("obj"),
				Method:    NewTestIdentifier("DoSomething"),
				Arguments: []Expression{},
			},
			expected: "obj.DoSomething()",
		},
		{
			name: "method call with arguments",
			methodCall: &MethodCallExpression{
				Token:  NewTestToken(lexer.DOT, "."),
				Object: NewTestIdentifier("point"),
				Method: NewTestIdentifier("MoveTo"),
				Arguments: []Expression{
					NewTestIntegerLiteral(10),
					NewTestIntegerLiteral(20),
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
