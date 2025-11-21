// Package ast provides test helper functions for creating common AST nodes.
// These reduce boilerplate and make test code more readable.
//
// Usage Examples:
//
//	// Instead of verbose struct initialization:
//	field := &FieldDecl{
//		BaseNode: BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "name"}},
//		Name: &Identifier{...},
//		Type: &TypeAnnotation{Name: "String"},
//		Visibility: VisibilityPublic,
//	}
//
//	// Use the helper:
//	field := NewTestFieldDecl("name", "String", VisibilityPublic)
//
//	// Create a function with parameters:
//	params := []*Parameter{
//		NewTestParameter("x", "Integer", false),
//		NewTestParameter("y", "String", true), // by reference
//	}
//	fn := NewTestFunctionDecl("Process", params, NewTestTypeAnnotation("Boolean"))
//
//	// Create expressions:
//	binary := NewTestBinaryExpression(
//		NewTestIdentifier("a"),
//		"+",
//		NewTestIntegerLiteral(42),
//	)
//	call := NewTestCallExpression(
//		NewTestIdentifier("PrintLn"),
//		[]Expression{NewTestStringLiteral("hello")},
//	)
package ast

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
)

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
// The tokenLiteral parameter should be the raw source code representation (e.g., "42").
func NewTestIntegerLiteral(value int64, tokenLiteral ...string) *IntegerLiteral {
	literal := fmt.Sprintf("%d", value)
	if len(tokenLiteral) > 0 {
		literal = tokenLiteral[0]
	}
	return &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.INT,
					Literal: literal,
				},
			},
		},
		Value: value,
	}
}

// NewTestFloatLiteral creates a FloatLiteral with the given value.
// The tokenLiteral parameter should be the raw source code representation (e.g., "3.14").
func NewTestFloatLiteral(value float64, tokenLiteral ...string) *FloatLiteral {
	literal := fmt.Sprintf("%g", value)
	if len(tokenLiteral) > 0 {
		literal = tokenLiteral[0]
	}
	return &FloatLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.FLOAT,
					Literal: literal,
				},
			},
		},
		Value: value,
	}
}

// NewTestStringLiteral creates a StringLiteral with the given value.
// The tokenLiteral parameter should be the raw source code representation (e.g., "'hello'").
func NewTestStringLiteral(value string, tokenLiteral ...string) *StringLiteral {
	literal := value
	if len(tokenLiteral) > 0 {
		literal = tokenLiteral[0]
	}
	return &StringLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.STRING,
					Literal: literal,
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

// NewTestFieldDecl creates a FieldDecl with the given name, type name, and visibility.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestFieldDecl(name, typeName string, visibility Visibility) *FieldDecl {
	return &FieldDecl{
		BaseNode: BaseNode{
			Token: NewTestToken(lexer.IDENT, name),
		},
		Name:       NewTestIdentifier(name),
		Type:       NewTestTypeAnnotation(typeName),
		Visibility: visibility,
		IsClassVar: false,
		InitValue:  nil,
	}
}

// NewTestParameter creates a Parameter with the given name, type name, and by-reference flag.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestParameter(name, typeName string, byRef bool) *Parameter {
	return &Parameter{
		Name:         NewTestIdentifier(name),
		Type:         NewTestTypeAnnotation(typeName),
		ByRef:        byRef,
		IsLazy:       false,
		IsConst:      false,
		DefaultValue: nil,
		Token:        NewTestToken(lexer.IDENT, name),
	}
}

// NewTestFunctionDecl creates a FunctionDecl with the given name, parameters, and return type.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestFunctionDecl(name string, params []*Parameter, returnType *TypeAnnotation) *FunctionDecl {
	tokenType := lexer.FUNCTION
	tokenLiteral := "function"
	if returnType == nil {
		tokenType = lexer.PROCEDURE
		tokenLiteral = "procedure"
	}
	return &FunctionDecl{
		BaseNode: BaseNode{
			Token: NewTestToken(tokenType, tokenLiteral),
		},
		Name:          NewTestIdentifier(name),
		Parameters:    params,
		ReturnType:    returnType,
		Body:          NewTestBlockStatement([]Statement{}), // Include empty body for String() output
		Visibility:    VisibilityPublic,
		IsClassMethod: false,
		IsVirtual:     false,
		IsOverride:    false,
		IsAbstract:    false,
		IsForward:     false,
		IsConstructor: false,
		IsDestructor:  false,
	}
}

// NewTestClassDecl creates a ClassDecl with the given name and optional parent.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestClassDecl(name string, parent *Identifier) *ClassDecl {
	return &ClassDecl{
		BaseNode: BaseNode{
			Token: NewTestToken(lexer.TYPE, "type"),
		},
		Name:       NewTestIdentifier(name),
		Parent:     parent,
		Fields:     []*FieldDecl{},
		Methods:    []*FunctionDecl{},
		Interfaces: []*Identifier{},
		Operators:  []*OperatorDecl{},
		Properties: []*PropertyDecl{},
		Constants:  []*ConstDecl{},
		IsAbstract: false,
		IsExternal: false,
		IsPartial:  false,
	}
}

// NewTestBlockStatement creates a BlockStatement with the given statements.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestBlockStatement(statements []Statement) *BlockStatement {
	return &BlockStatement{
		BaseNode:   BaseNode{Token: NewTestToken(lexer.BEGIN, "begin")},
		Statements: statements,
	}
}

// NewTestBinaryExpression creates a BinaryExpression with the given left operand, operator, and right operand.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestBinaryExpression(left Expression, operator string, right Expression) *BinaryExpression {
	return &BinaryExpression{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.IDENT, operator), // Use IDENT for operator token
			},
		},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

// NewTestUnaryExpression creates a UnaryExpression with the given operator and operand.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestUnaryExpression(operator string, operand Expression) *UnaryExpression {
	return &UnaryExpression{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: NewTestToken(lexer.IDENT, operator), // Use IDENT for operator token
			},
		},
		Right:    operand,
		Operator: operator,
	}
}

// NewTestCallExpression creates a CallExpression with the given function and arguments.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestCallExpression(function Expression, args []Expression) *CallExpression {
	return &CallExpression{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{Token: NewTestToken(lexer.LPAREN, "(")}, // Use LPAREN for call token
		},
		Function:  function,
		Arguments: args,
	}
}

// NewTestGroupedExpression creates a GroupedExpression with the given inner expression.
// This is a convenience helper for tests to avoid verbose struct initialization.
func NewTestGroupedExpression(expression Expression) *GroupedExpression {
	return &GroupedExpression{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{Token: NewTestToken(lexer.LPAREN, "(")},
		},
		Expression: expression,
	}
}

// NewTestIntegerLiteralWithPos creates an IntegerLiteral with the given value and position.
// This is a convenience helper for tests that need position information.
func NewTestIntegerLiteralWithPos(value int64, line, column int, tokenLiteral ...string) *IntegerLiteral {
	literal := fmt.Sprintf("%d", value)
	if len(tokenLiteral) > 0 {
		literal = tokenLiteral[0]
	}
	return &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.INT,
					Literal: literal,
					Pos:     lexer.Position{Line: line, Column: column},
				},
			},
		},
		Value: value,
	}
}

// NewTestUnaryExpressionWithPos creates a UnaryExpression with position information.
// This is a convenience helper for tests that need position information.
func NewTestUnaryExpressionWithPos(operator string, operand Expression, line, column int) *UnaryExpression {
	return &UnaryExpression{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: lexer.Token{
					Type:    lexer.IDENT,
					Literal: operator,
					Pos:     lexer.Position{Line: line, Column: column},
				},
			},
		},
		Right:    operand,
		Operator: operator,
	}
}

// NewTestBaseNodeWithPos creates a BaseNode with the given token type, literal, and position.
// This is a convenience helper for tests to avoid verbose BaseNode initialization.
func NewTestBaseNodeWithPos(tokenType lexer.TokenType, literal string, line, column int) BaseNode {
	return BaseNode{
		Token: lexer.Token{
			Type:    tokenType,
			Literal: literal,
			Pos:     lexer.Position{Line: line, Column: column},
		},
	}
}

// NewTestConstDeclBaseNode creates a BaseNode for const declarations with position information.
// This is a convenience helper for tests to avoid verbose BaseNode initialization.
func NewTestConstDeclBaseNode(line, column int) BaseNode {
	return NewTestBaseNodeWithPos(lexer.CONST, "const", line, column)
}

// NewTestTypeDeclBaseNode creates a BaseNode for type declarations with position information.
// This is a convenience helper for tests to avoid verbose BaseNode initialization.
func NewTestTypeDeclBaseNode(line, column int) BaseNode {
	return NewTestBaseNodeWithPos(lexer.TYPE, "type", line, column)
}
