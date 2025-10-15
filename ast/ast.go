// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/lexer"
)

// Node is the base interface for all AST nodes.
// Every node in the AST must be able to provide its token literal
// and a string representation for debugging.
type Node interface {
	// TokenLiteral returns the literal value of the token this node is associated with.
	TokenLiteral() string

	// String returns a string representation of the node for debugging and testing.
	String() string
}

// Expression represents any node that produces a value.
// All expression nodes must implement this interface.
type Expression interface {
	Node
	expressionNode()
}

// Statement represents a node that performs an action but doesn't produce a value.
// All statement nodes must implement this interface.
type Statement interface {
	Node
	statementNode()
}

// Program is the root node of the AST.
// It contains a slice of statements that make up the entire program.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, stmt := range p.Statements {
		out.WriteString(stmt.String())
	}

	return out.String()
}

// Identifier represents an identifier (variable name, function name, etc.)
type Identifier struct {
	Token lexer.Token // The IDENT token
	Value string      // The actual identifier name
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer literal value.
type IntegerLiteral struct {
	Token lexer.Token // The INT token
	Value int64       // The parsed integer value
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a floating-point literal value.
type FloatLiteral struct {
	Token lexer.Token // The FLOAT token
	Value float64     // The parsed float value
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal value.
type StringLiteral struct {
	Token lexer.Token // The STRING token
	Value string      // The parsed string value (without quotes)
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// BooleanLiteral represents a boolean literal value (true or false).
type BooleanLiteral struct {
	Token lexer.Token // The TRUE or FALSE token
	Value bool        // The boolean value
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// BinaryExpression represents a binary operation (e.g., a + b, x < y).
type BinaryExpression struct {
	Token    lexer.Token // The operator token
	Left     Expression  // The left operand
	Operator string      // The operator as a string (+, -, *, /, =, <>, etc.)
	Right    Expression  // The right operand
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" " + be.Operator + " ")
	out.WriteString(be.Right.String())
	out.WriteString(")")

	return out.String()
}

// UnaryExpression represents a unary operation (e.g., -x, not b).
type UnaryExpression struct {
	Token    lexer.Token // The operator token
	Operator string      // The operator as a string (-, not, +)
	Right    Expression  // The operand
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }
func (ue *UnaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ue.Operator)

	// Add space after operator if it's a word (like "not")
	if len(ue.Operator) > 0 && (ue.Operator[0] >= 'a' && ue.Operator[0] <= 'z' || ue.Operator[0] >= 'A' && ue.Operator[0] <= 'Z') {
		out.WriteString(" ")
	}

	out.WriteString(ue.Right.String())
	out.WriteString(")")

	return out.String()
}

// GroupedExpression represents an expression wrapped in parentheses.
type GroupedExpression struct {
	Token      lexer.Token // The '(' token
	Expression Expression  // The expression inside the parentheses
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupedExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ge.Expression.String())
	out.WriteString(")")

	return out.String()
}

// ExpressionStatement represents a statement that consists of a single expression.
// This is used when an expression appears in a statement context.
type ExpressionStatement struct {
	Token      lexer.Token // The first token of the expression
	Expression Expression  // The expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// NilLiteral represents a nil literal value.
type NilLiteral struct {
	Token lexer.Token // The NIL token
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return "nil" }

// BlockStatement represents a block of statements (begin...end).
type BlockStatement struct {
	Token      lexer.Token // The 'begin' token
	Statements []Statement // The statements in the block
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString("begin\n")
	for _, stmt := range bs.Statements {
		out.WriteString("  ")
		out.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
		out.WriteString("\n")
	}
	out.WriteString("end")

	return out.String()
}
