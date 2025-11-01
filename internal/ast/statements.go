// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// VarDeclStatement represents a variable declaration statement.
// Examples:
//
//	var x: Integer;
//	var x: Integer := 42;
//	var x := 5;
//	var x, y, z: Integer;                 // Multi-identifier declaration
//	var y: String; external;              // Task 7.143
//	var z: Integer; external 'externalZ'; // Task 7.143
type VarDeclStatement struct {
	Value        Expression
	Names        []*Identifier // Changed from Name to support multi-identifier declarations
	Type         *TypeAnnotation
	ExternalName string
	Token        lexer.Token
	IsExternal   bool
	Inferred     bool // true when the type is inferred from the initializer
}

func (vds *VarDeclStatement) statementNode()       {}
func (vds *VarDeclStatement) TokenLiteral() string { return vds.Token.Literal }
func (vds *VarDeclStatement) Pos() lexer.Position  { return vds.Token.Pos }
func (vds *VarDeclStatement) String() string {
	var out bytes.Buffer

	out.WriteString("var ")

	// Join multiple names with ", "
	names := []string{}
	for _, name := range vds.Names {
		names = append(names, name.String())
	}
	out.WriteString(strings.Join(names, ", "))

	if vds.Type != nil {
		out.WriteString(": ")
		out.WriteString(vds.Type.String())
	}

	if vds.Value != nil {
		separator := " := "
		if vds.Inferred && vds.Type == nil {
			separator = " = "
		}
		out.WriteString(separator)
		out.WriteString(vds.Value.String())
	}

	return out.String()
}

// AssignmentStatement represents an assignment statement.
// Examples:
//
//	x := 10;             // simple variable assignment
//	x := x + 1;          // assignment with expression
//	x += 5;              // compound assignment
//	arr[i] := 42;        // array element assignment
//	obj.field := value;  // member assignment
//	matrix[i][j] := 99;  // nested array assignment
type AssignmentStatement struct {
	Target   Expression
	Value    Expression
	Token    lexer.Token
	Operator lexer.TokenType // ASSIGN, PLUS_ASSIGN, MINUS_ASSIGN, TIMES_ASSIGN, DIVIDE_ASSIGN
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) Pos() lexer.Position  { return as.Token.Pos }
func (as *AssignmentStatement) String() string {
	var out bytes.Buffer

	// Handle different target types
	if as.Target != nil {
		out.WriteString(as.Target.String())
	}

	// Use compound operator if set, otherwise use :=
	switch as.Operator {
	case lexer.PLUS_ASSIGN:
		out.WriteString(" += ")
	case lexer.MINUS_ASSIGN:
		out.WriteString(" -= ")
	case lexer.TIMES_ASSIGN:
		out.WriteString(" *= ")
	case lexer.DIVIDE_ASSIGN:
		out.WriteString(" /= ")
	default:
		out.WriteString(" := ")
	}

	if as.Value != nil {
		out.WriteString(as.Value.String())
	}

	return out.String()
}

// CallExpression represents a function call expression.
// Examples:
//
//	PrintLn('hello')
//	Add(3, 5)
//	Foo()
type CallExpression struct {
	Function  Expression
	Type      *TypeAnnotation
	Arguments []Expression
	Token     lexer.Token
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) Pos() lexer.Position  { return ce.Function.Pos() }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ce.Function.String())
	out.WriteString("(")

	args := []string{}
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))

	out.WriteString(")")

	return out.String()
}
func (ce *CallExpression) GetType() *TypeAnnotation    { return ce.Type }
func (ce *CallExpression) SetType(typ *TypeAnnotation) { ce.Type = typ }
