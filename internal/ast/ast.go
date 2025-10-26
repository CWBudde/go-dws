// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Node is the base interface for all AST nodes.
// Every node in the AST must be able to provide its token literal,
// position information, and a string representation for debugging.
type Node interface {
	// TokenLiteral returns the literal value of the token this node is associated with.
	TokenLiteral() string

	// String returns a string representation of the node for debugging and testing.
	String() string

	// Pos returns the position of the node in the source code for error reporting.
	Pos() lexer.Position
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

func (p *Program) Pos() lexer.Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return lexer.Position{Line: 1, Column: 1, Offset: 0}
}

// Identifier represents an identifier (variable name, function name, etc.)
type Identifier struct {
	Token lexer.Token     // The IDENT token
	Value string          // The actual identifier name
	Type  *TypeAnnotation // The inferred or annotated type (set by semantic analyzer)
}

func (i *Identifier) expressionNode()             {}
func (i *Identifier) TokenLiteral() string        { return i.Token.Literal }
func (i *Identifier) String() string              { return i.Value }
func (i *Identifier) Pos() lexer.Position         { return i.Token.Pos }
func (i *Identifier) GetType() *TypeAnnotation    { return i.Type }
func (i *Identifier) SetType(typ *TypeAnnotation) { i.Type = typ }

// IntegerLiteral represents an integer literal value.
type IntegerLiteral struct {
	Token lexer.Token     // The INT token
	Value int64           // The parsed integer value
	Type  *TypeAnnotation // The type (always Integer for integer literals)
}

func (il *IntegerLiteral) expressionNode()             {}
func (il *IntegerLiteral) TokenLiteral() string        { return il.Token.Literal }
func (il *IntegerLiteral) String() string              { return il.Token.Literal }
func (il *IntegerLiteral) Pos() lexer.Position         { return il.Token.Pos }
func (il *IntegerLiteral) GetType() *TypeAnnotation    { return il.Type }
func (il *IntegerLiteral) SetType(typ *TypeAnnotation) { il.Type = typ }

// FloatLiteral represents a floating-point literal value.
type FloatLiteral struct {
	Token lexer.Token     // The FLOAT token
	Value float64         // The parsed float value
	Type  *TypeAnnotation // The type (always Float for float literals)
}

func (fl *FloatLiteral) expressionNode()             {}
func (fl *FloatLiteral) TokenLiteral() string        { return fl.Token.Literal }
func (fl *FloatLiteral) String() string              { return fl.Token.Literal }
func (fl *FloatLiteral) Pos() lexer.Position         { return fl.Token.Pos }
func (fl *FloatLiteral) GetType() *TypeAnnotation    { return fl.Type }
func (fl *FloatLiteral) SetType(typ *TypeAnnotation) { fl.Type = typ }

// StringLiteral represents a string literal value.
type StringLiteral struct {
	Token lexer.Token     // The STRING token
	Value string          // The parsed string value (without quotes)
	Type  *TypeAnnotation // The type (always String for string literals)
}

func (sl *StringLiteral) expressionNode()             {}
func (sl *StringLiteral) TokenLiteral() string        { return sl.Token.Literal }
func (sl *StringLiteral) String() string              { return "\"" + sl.Value + "\"" }
func (sl *StringLiteral) Pos() lexer.Position         { return sl.Token.Pos }
func (sl *StringLiteral) GetType() *TypeAnnotation    { return sl.Type }
func (sl *StringLiteral) SetType(typ *TypeAnnotation) { sl.Type = typ }

// BooleanLiteral represents a boolean literal value (true or false).
type BooleanLiteral struct {
	Token lexer.Token     // The TRUE or FALSE token
	Value bool            // The boolean value
	Type  *TypeAnnotation // The type (always Boolean for boolean literals)
}

func (bl *BooleanLiteral) expressionNode()             {}
func (bl *BooleanLiteral) TokenLiteral() string        { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string              { return bl.Token.Literal }
func (bl *BooleanLiteral) Pos() lexer.Position         { return bl.Token.Pos }
func (bl *BooleanLiteral) GetType() *TypeAnnotation    { return bl.Type }
func (bl *BooleanLiteral) SetType(typ *TypeAnnotation) { bl.Type = typ }

// BinaryExpression represents a binary operation (e.g., a + b, x < y).
type BinaryExpression struct {
	Token    lexer.Token     // The operator token
	Left     Expression      // The left operand
	Operator string          // The operator as a string (+, -, *, /, =, <>, etc.)
	Right    Expression      // The right operand
	Type     *TypeAnnotation // The result type (determined by semantic analyzer)
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) Pos() lexer.Position  { return be.Token.Pos }
func (be *BinaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" " + be.Operator + " ")
	out.WriteString(be.Right.String())
	out.WriteString(")")

	return out.String()
}
func (be *BinaryExpression) GetType() *TypeAnnotation    { return be.Type }
func (be *BinaryExpression) SetType(typ *TypeAnnotation) { be.Type = typ }

// UnaryExpression represents a unary operation (e.g., -x, not b).
type UnaryExpression struct {
	Token    lexer.Token     // The operator token
	Operator string          // The operator as a string (-, not, +)
	Right    Expression      // The operand
	Type     *TypeAnnotation // The result type (determined by semantic analyzer)
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }
func (ue *UnaryExpression) Pos() lexer.Position  { return ue.Token.Pos }
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
func (ue *UnaryExpression) GetType() *TypeAnnotation    { return ue.Type }
func (ue *UnaryExpression) SetType(typ *TypeAnnotation) { ue.Type = typ }

// GroupedExpression represents an expression wrapped in parentheses.
type GroupedExpression struct {
	Token      lexer.Token     // The '(' token
	Expression Expression      // The expression inside the parentheses
	Type       *TypeAnnotation // The type (same as inner expression)
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupedExpression) Pos() lexer.Position  { return ge.Token.Pos }
func (ge *GroupedExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ge.Expression.String())
	out.WriteString(")")

	return out.String()
}
func (ge *GroupedExpression) GetType() *TypeAnnotation    { return ge.Type }
func (ge *GroupedExpression) SetType(typ *TypeAnnotation) { ge.Type = typ }

// RangeExpression represents a range expression (e.g., A..C in set literals).
// Used primarily in set literals to specify a range of enum values.
// Example: [A..C] or [one..five]
type RangeExpression struct {
	Token lexer.Token     // The '..' token
	Start Expression      // The start of the range
	End   Expression      // The end of the range
	Type  *TypeAnnotation // The type (determined by semantic analyzer)
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) Pos() lexer.Position  { return re.Token.Pos }
func (re *RangeExpression) String() string {
	var out bytes.Buffer

	out.WriteString(re.Start.String())
	out.WriteString("..")
	out.WriteString(re.End.String())

	return out.String()
}
func (re *RangeExpression) GetType() *TypeAnnotation    { return re.Type }
func (re *RangeExpression) SetType(typ *TypeAnnotation) { re.Type = typ }

// ExpressionStatement represents a statement that consists of a single expression.
// This is used when an expression appears in a statement context.
type ExpressionStatement struct {
	Token      lexer.Token // The first token of the expression
	Expression Expression  // The expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) Pos() lexer.Position  { return es.Token.Pos }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// NilLiteral represents a nil literal value.
type NilLiteral struct {
	Token lexer.Token     // The NIL token
	Type  *TypeAnnotation // The type (always Nil for nil literals)
}

func (nl *NilLiteral) expressionNode()             {}
func (nl *NilLiteral) TokenLiteral() string        { return nl.Token.Literal }
func (nl *NilLiteral) String() string              { return "nil" }
func (nl *NilLiteral) Pos() lexer.Position         { return nl.Token.Pos }
func (nl *NilLiteral) GetType() *TypeAnnotation    { return nl.Type }
func (nl *NilLiteral) SetType(typ *TypeAnnotation) { nl.Type = typ }

// BlockStatement represents a block of statements (begin...end).
type BlockStatement struct {
	Token      lexer.Token // The 'begin' token
	Statements []Statement // The statements in the block
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) Pos() lexer.Position  { return bs.Token.Pos }
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
