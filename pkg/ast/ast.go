// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// Node is the base interface for all AST nodes.
// Every node in the AST must be able to provide its token literal,
// position information, and a string representation for debugging.
type Node interface {
	// TokenLiteral returns the literal value of the token this node is associated with.
	TokenLiteral() string

	// String returns a string representation of the node for debugging and testing.
	String() string

	// Pos returns the start position of the node in the source code for error reporting.
	Pos() token.Position

	// End returns the end position of the node in the source code for error reporting.
	// This enables precise error underlining and source range tracking.
	End() token.Position
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
	EndPos     token.Position
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

func (p *Program) Pos() token.Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return token.Position{Line: 1, Column: 1, Offset: 0}
}

func (p *Program) End() token.Position {
	if p.EndPos.Line != 0 {
		return p.EndPos
	}
	if len(p.Statements) > 0 {
		return p.Statements[len(p.Statements)-1].End()
	}
	return token.Position{Line: 1, Column: 1, Offset: 0}
}

// Identifier represents an identifier (variable name, function name, etc.)
type Identifier struct {
	Type   *TypeAnnotation
	Value  string
	Token  token.Token
	EndPos token.Position
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) Pos() token.Position  { return i.Token.Pos }
func (i *Identifier) End() token.Position {
	if i.EndPos.Line != 0 {
		return i.EndPos
	}
	// Calculate end position from start + length of identifier
	pos := i.Token.Pos
	pos.Column += len(i.Value)
	pos.Offset += len(i.Value)
	return pos
}
func (i *Identifier) GetType() *TypeAnnotation    { return i.Type }
func (i *Identifier) SetType(typ *TypeAnnotation) { i.Type = typ }

// IntegerLiteral represents an integer literal value.
type IntegerLiteral struct {
	Type   *TypeAnnotation
	Token  token.Token
	Value  int64
	EndPos token.Position
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) Pos() token.Position  { return il.Token.Pos }
func (il *IntegerLiteral) End() token.Position {
	if il.EndPos.Line != 0 {
		return il.EndPos
	}
	pos := il.Token.Pos
	pos.Column += len(il.Token.Literal)
	pos.Offset += len(il.Token.Literal)
	return pos
}
func (il *IntegerLiteral) GetType() *TypeAnnotation    { return il.Type }
func (il *IntegerLiteral) SetType(typ *TypeAnnotation) { il.Type = typ }

// FloatLiteral represents a floating-point literal value.
type FloatLiteral struct {
	Type   *TypeAnnotation
	Token  token.Token
	Value  float64
	EndPos token.Position
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }
func (fl *FloatLiteral) Pos() token.Position  { return fl.Token.Pos }
func (fl *FloatLiteral) End() token.Position {
	if fl.EndPos.Line != 0 {
		return fl.EndPos
	}
	pos := fl.Token.Pos
	pos.Column += len(fl.Token.Literal)
	pos.Offset += len(fl.Token.Literal)
	return pos
}
func (fl *FloatLiteral) GetType() *TypeAnnotation    { return fl.Type }
func (fl *FloatLiteral) SetType(typ *TypeAnnotation) { fl.Type = typ }

// StringLiteral represents a string literal value.
type StringLiteral struct {
	Type   *TypeAnnotation
	Value  string
	Token  token.Token
	EndPos token.Position
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }
func (sl *StringLiteral) Pos() token.Position  { return sl.Token.Pos }
func (sl *StringLiteral) End() token.Position {
	if sl.EndPos.Line != 0 {
		return sl.EndPos
	}
	pos := sl.Token.Pos
	pos.Column += len(sl.Token.Literal)
	pos.Offset += len(sl.Token.Literal)
	return pos
}
func (sl *StringLiteral) GetType() *TypeAnnotation    { return sl.Type }
func (sl *StringLiteral) SetType(typ *TypeAnnotation) { sl.Type = typ }

// BooleanLiteral represents a boolean literal value (true or false).
type BooleanLiteral struct {
	Type   *TypeAnnotation
	Token  token.Token
	Value  bool
	EndPos token.Position
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }
func (bl *BooleanLiteral) Pos() token.Position  { return bl.Token.Pos }
func (bl *BooleanLiteral) End() token.Position {
	if bl.EndPos.Line != 0 {
		return bl.EndPos
	}
	pos := bl.Token.Pos
	pos.Column += len(bl.Token.Literal)
	pos.Offset += len(bl.Token.Literal)
	return pos
}
func (bl *BooleanLiteral) GetType() *TypeAnnotation    { return bl.Type }
func (bl *BooleanLiteral) SetType(typ *TypeAnnotation) { bl.Type = typ }

// CharLiteral represents a character literal value.
// Supports three forms: 'H' (single char string), #13 (decimal), #$41 (hex).
type CharLiteral struct {
	Type   *TypeAnnotation
	Token  token.Token
	Value  rune
	EndPos token.Position
}

func (cl *CharLiteral) expressionNode()      {}
func (cl *CharLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *CharLiteral) String() string       { return cl.Token.Literal }
func (cl *CharLiteral) Pos() token.Position  { return cl.Token.Pos }
func (cl *CharLiteral) End() token.Position {
	if cl.EndPos.Line != 0 {
		return cl.EndPos
	}
	pos := cl.Token.Pos
	pos.Column += len(cl.Token.Literal)
	pos.Offset += len(cl.Token.Literal)
	return pos
}
func (cl *CharLiteral) GetType() *TypeAnnotation    { return cl.Type }
func (cl *CharLiteral) SetType(typ *TypeAnnotation) { cl.Type = typ }

// BinaryExpression represents a binary operation (e.g., a + b, x < y).
type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Type     *TypeAnnotation
	Operator string
	Token    token.Token
	EndPos   token.Position
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) Pos() token.Position  { return be.Token.Pos }
func (be *BinaryExpression) End() token.Position {
	if be.EndPos.Line != 0 {
		return be.EndPos
	}
	if be.Right != nil {
		return be.Right.End()
	}
	return be.Token.Pos
}
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
	Right    Expression
	Type     *TypeAnnotation
	Operator string
	Token    token.Token
	EndPos   token.Position
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }
func (ue *UnaryExpression) Pos() token.Position  { return ue.Token.Pos }
func (ue *UnaryExpression) End() token.Position {
	if ue.EndPos.Line != 0 {
		return ue.EndPos
	}
	if ue.Right != nil {
		return ue.Right.End()
	}
	return ue.Token.Pos
}
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
	Expression Expression
	Type       *TypeAnnotation
	Token      token.Token
	EndPos     token.Position
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupedExpression) Pos() token.Position  { return ge.Token.Pos }
func (ge *GroupedExpression) End() token.Position {
	if ge.EndPos.Line != 0 {
		return ge.EndPos
	}
	// End position is after the closing parenthesis
	if ge.Expression != nil {
		pos := ge.Expression.End()
		pos.Column++ // For the ')'
		pos.Offset++
		return pos
	}
	return ge.Token.Pos
}
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
	Start    Expression
	RangeEnd Expression // Renamed from 'End' to avoid conflict with End() method
	Type     *TypeAnnotation
	Token    token.Token
	EndPos   token.Position
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) Pos() token.Position  { return re.Token.Pos }
func (re *RangeExpression) End() token.Position {
	if re.EndPos.Line != 0 {
		return re.EndPos
	}
	if re.RangeEnd != nil {
		return re.RangeEnd.End()
	}
	return re.Token.Pos
}
func (re *RangeExpression) String() string {
	var out bytes.Buffer

	out.WriteString(re.Start.String())
	out.WriteString("..")
	out.WriteString(re.RangeEnd.String())

	return out.String()
}
func (re *RangeExpression) GetType() *TypeAnnotation    { return re.Type }
func (re *RangeExpression) SetType(typ *TypeAnnotation) { re.Type = typ }

// ExpressionStatement represents a statement that consists of a single expression.
// This is used when an expression appears in a statement context.
type ExpressionStatement struct {
	Expression Expression
	Token      token.Token
	EndPos     token.Position
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) Pos() token.Position  { return es.Token.Pos }
func (es *ExpressionStatement) End() token.Position {
	if es.EndPos.Line != 0 {
		return es.EndPos
	}
	if es.Expression != nil {
		return es.Expression.End()
	}
	return es.Token.Pos
}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// NilLiteral represents a nil literal value.
type NilLiteral struct {
	Type   *TypeAnnotation
	Token  token.Token
	EndPos token.Position
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return "nil" }
func (nl *NilLiteral) Pos() token.Position  { return nl.Token.Pos }
func (nl *NilLiteral) End() token.Position {
	if nl.EndPos.Line != 0 {
		return nl.EndPos
	}
	pos := nl.Token.Pos
	pos.Column += len(nl.Token.Literal)
	pos.Offset += len(nl.Token.Literal)
	return pos
}
func (nl *NilLiteral) GetType() *TypeAnnotation    { return nl.Type }
func (nl *NilLiteral) SetType(typ *TypeAnnotation) { nl.Type = typ }

// AsExpression represents a type cast operation using the 'as' operator.
// Example: obj as IMyInterface
// This operator casts an object to an interface type, creating an InterfaceInstance
// wrapper at runtime. The semantic analyzer validates that the object's class
// implements the target interface.
type AsExpression struct {
	Left       Expression     // The object being cast
	TargetType TypeExpression // The target interface type
	Type       *TypeAnnotation // Resolved type (will be the interface type)
	Token      token.Token    // The 'as' token
	EndPos     token.Position
}

func (ae *AsExpression) expressionNode()      {}
func (ae *AsExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AsExpression) Pos() token.Position  { return ae.Left.Pos() }
func (ae *AsExpression) End() token.Position {
	if ae.EndPos.Line != 0 {
		return ae.EndPos
	}
	if ae.TargetType != nil {
		return ae.TargetType.End()
	}
	return ae.Token.Pos
}
func (ae *AsExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ae.Left.String())
	out.WriteString(" as ")
	out.WriteString(ae.TargetType.String())
	out.WriteString(")")
	return out.String()
}
func (ae *AsExpression) GetType() *TypeAnnotation    { return ae.Type }
func (ae *AsExpression) SetType(typ *TypeAnnotation) { ae.Type = typ }

// ImplementsExpression represents the 'implements' operator.
// Example: obj implements IMyInterface  -> Boolean
// This operator checks whether an object's class implements a given interface.
// Can be used at compile-time (TClass implements IInterface) or runtime
// (objInstance implements IInterface).
type ImplementsExpression struct {
	Left       Expression     // The object or class being checked
	TargetType TypeExpression // The interface type to check against
	Type       *TypeAnnotation // Always resolves to Boolean
	Token      token.Token    // The 'implements' token
	EndPos     token.Position
}

func (ie *ImplementsExpression) expressionNode()      {}
func (ie *ImplementsExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *ImplementsExpression) Pos() token.Position  { return ie.Left.Pos() }
func (ie *ImplementsExpression) End() token.Position {
	if ie.EndPos.Line != 0 {
		return ie.EndPos
	}
	if ie.TargetType != nil {
		return ie.TargetType.End()
	}
	return ie.Token.Pos
}
func (ie *ImplementsExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" implements ")
	out.WriteString(ie.TargetType.String())
	out.WriteString(")")
	return out.String()
}
func (ie *ImplementsExpression) GetType() *TypeAnnotation    { return ie.Type }
func (ie *ImplementsExpression) SetType(typ *TypeAnnotation) { ie.Type = typ }

// BlockStatement represents a block of statements (begin...end).
type BlockStatement struct {
	Statements []Statement
	Token      token.Token
	EndPos     token.Position
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) Pos() token.Position  { return bs.Token.Pos }
func (bs *BlockStatement) End() token.Position {
	if bs.EndPos.Line != 0 {
		return bs.EndPos
	}
	// End position should be after the 'end' keyword
	// For now, use the last statement's end if available
	if len(bs.Statements) > 0 {
		return bs.Statements[len(bs.Statements)-1].End()
	}
	return bs.Token.Pos
}
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
