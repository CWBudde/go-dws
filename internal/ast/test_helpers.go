package ast

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Test helper functions for creating common AST nodes in tests.
// These reduce boilerplate and make test code more readable.

// NewTestIdentifier creates an Identifier with the given name.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestIdentifier(name string) *Identifier {
	return &Identifier{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.IDENT,
					Literal: name,
				},
			},
		},
		Value: name,
	}
}

// NewTestIntegerLiteral creates an IntegerLiteral with the given value.
func NewTestIntegerLiteral(value int64) *IntegerLiteral {
	return &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.INT,
					Literal: fmt.Sprintf("%d", value),
				},
			},
		},
		Value: value,
	}
}

// NewTestFloatLiteral creates a FloatLiteral with the given value.
func NewTestFloatLiteral(value float64) *FloatLiteral {
	return &FloatLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.FLOAT,
					Literal: fmt.Sprintf("%g", value),
				},
			},
		},
		Value: value,
	}
}

// NewTestStringLiteral creates a StringLiteral with the given value.
func NewTestStringLiteral(value string) *StringLiteral {
	return &StringLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.STRING,
					Literal: value,
				},
			},
		},
		Value: value,
	}
}

// NewTestBooleanLiteral creates a BooleanLiteral with the given value.
func NewTestBooleanLiteral(value bool) *BooleanLiteral {
	tokenType := lexer.TRUE
	literal := "true"
	if !value {
		tokenType = lexer.FALSE
		literal = "false"
	}
	return &BooleanLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    tokenType,
					Literal: literal,
				},
			},
		},
		Value: value,
	}
}

// NewTestTypeAnnotation creates a TypeAnnotation with the given type name.
func NewTestTypeAnnotation(typeName string) *TypeAnnotation {
	return &TypeAnnotation{
		Name: typeName,
	}
}

// NewTestArrayTypeAnnotation creates an array type annotation.
func NewTestArrayTypeAnnotation(elementType string) *TypeAnnotation {
	return &TypeAnnotation{
		Name: "array of " + elementType,
	}
}

// NewTestToken creates a token with the given type and literal.
func NewTestToken(tokenType lexer.TokenType, literal string) lexer.Token {
	return lexer.Token{
		Type:    tokenType,
		Literal: literal,
	}
}

// NewTestBaseNode creates a BaseNode with the given token.
func NewTestBaseNode(tokenType lexer.TokenType, literal string) BaseNode {
	return BaseNode{
		Token: NewTestToken(tokenType, literal),
	}
}
