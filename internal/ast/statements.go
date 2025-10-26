// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/lexer"
)

// VarDeclStatement represents a variable declaration statement.
// Examples:
//
//	var x: Integer;
//	var x: Integer := 42;
//	var x := 5;
//	var y: String; external;              // Task 7.143
//	var z: Integer; external 'externalZ'; // Task 7.143
type VarDeclStatement struct {
	Token        lexer.Token     // The 'var' token
	Name         *Identifier     // The variable name
	Type         *TypeAnnotation // The type annotation (nil if not specified)
	Value        Expression      // The initialization value (nil if not initialized)
	IsExternal   bool            // True if this is an external variable (Task 7.143)
	ExternalName string          // External name for FFI binding (optional) - Task 7.143
}

func (vds *VarDeclStatement) statementNode()       {}
func (vds *VarDeclStatement) TokenLiteral() string { return vds.Token.Literal }
func (vds *VarDeclStatement) Pos() lexer.Position  { return vds.Token.Pos }
func (vds *VarDeclStatement) String() string {
	var out bytes.Buffer

	out.WriteString("var ")
	out.WriteString(vds.Name.String())

	if vds.Type != nil {
		out.WriteString(": ")
		out.WriteString(vds.Type.String())
	}

	if vds.Value != nil {
		out.WriteString(" := ")
		out.WriteString(vds.Value.String())
	}

	return out.String()
}

// AssignmentStatement represents an assignment statement.
// Examples:
//
//	x := 10;             // simple variable assignment
//	x := x + 1;          // assignment with expression
//	arr[i] := 42;        // array element assignment
//	obj.field := value;  // member assignment
//	matrix[i][j] := 99;  // nested array assignment
type AssignmentStatement struct {
	Token  lexer.Token // The ':=' token
	Target Expression  // The assignment target (Identifier, IndexExpression, or MemberAccessExpression)
	Value  Expression  // The value to assign
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
	out.WriteString(" := ")
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
	Token     lexer.Token     // The '(' token
	Function  Expression      // The function being called (usually an Identifier)
	Arguments []Expression    // The arguments to the function
	Type      *TypeAnnotation // The return type (determined by semantic analyzer)
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
