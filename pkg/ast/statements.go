// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// VarDeclStatement represents a variable declaration statement.
// Examples:
//
//	var x: Integer;
//	var x: Integer := 42;
//	var x := 5;
//	var x, y, z: Integer;                 // Multi-identifier declaration
//	var y: String; external;
//	var z: Integer; external 'externalZ';
type VarDeclStatement struct {
	BaseNode
	Value        Expression
	Names        []*Identifier // Changed from Name to support multi-identifier declarations
	Type         TypeExpression // Can be TypeAnnotation, ArrayTypeNode, FunctionPointerTypeNode, etc.
	ExternalName string
	IsExternal   bool
	Inferred     bool // true when the type is inferred from the initializer
}

func (vds *VarDeclStatement) statementNode() {}

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
	BaseNode
	Target   Expression
	Value    Expression
	Operator token.TokenType // ASSIGN, PLUS_ASSIGN, MINUS_ASSIGN, TIMES_ASSIGN, DIVIDE_ASSIGN
}

func (as *AssignmentStatement) statementNode() {}

func (as *AssignmentStatement) String() string {
	var out bytes.Buffer

	// Handle different target types
	if as.Target != nil {
		out.WriteString(as.Target.String())
	}

	// Use compound operator if set, otherwise use :=
	switch as.Operator {
	case token.PLUS_ASSIGN:
		out.WriteString(" += ")
	case token.MINUS_ASSIGN:
		out.WriteString(" -= ")
	case token.TIMES_ASSIGN:
		out.WriteString(" *= ")
	case token.DIVIDE_ASSIGN:
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
	TypedExpressionBase
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

// Pos returns the start position from the Function expression.
func (ce *CallExpression) Pos() token.Position { return ce.Function.Pos() }

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

// Condition represents a single contract condition (precondition or postcondition).
// A condition consists of a test expression (must be boolean) and an optional message.
// Examples:
//
//	x > 0
//	x > 0 : 'x must be positive'
//	a.Length = b.Length : 'arrays must have same length'
type Condition struct {
	BaseNode
	Test    Expression // Must evaluate to boolean
	Message Expression // Optional string message (if nil, use source code as message)
}

func (c *Condition) statementNode() {}

func (c *Condition) String() string {
	var out bytes.Buffer

	if c.Test != nil {
		out.WriteString(c.Test.String())
	}

	if c.Message != nil {
		out.WriteString(" : ")
		out.WriteString(c.Message.String())
	}

	return out.String()
}

// PreConditions represents a collection of preconditions for a function/method.
// Preconditions are checked before the function body executes.
// Example:
//
//	require
//	   x > 0;
//	   y <> 0 : 'y cannot be zero';
type PreConditions struct {
	BaseNode
	Conditions []*Condition
}

func (pc *PreConditions) statementNode() {}

func (pc *PreConditions) String() string {
	var out bytes.Buffer

	out.WriteString("require\n")
	for i, cond := range pc.Conditions {
		out.WriteString("   ")
		out.WriteString(cond.String())
		if i < len(pc.Conditions)-1 {
			out.WriteString(";")
		}
		out.WriteString("\n")
	}

	return out.String()
}

// PostConditions represents a collection of postconditions for a function/method.
// Postconditions are checked after the function body executes, before returning.
// They can reference 'old' values captured before execution.
// Example:
//
//	ensure
//	   Result > 0;
//	   Result = old x + 1 : 'result must be one more than original x';
type PostConditions struct {
	BaseNode
	Conditions []*Condition
}

func (pc *PostConditions) statementNode() {}

func (pc *PostConditions) String() string {
	var out bytes.Buffer

	out.WriteString("ensure\n")
	for i, cond := range pc.Conditions {
		out.WriteString("   ")
		out.WriteString(cond.String())
		if i < len(pc.Conditions)-1 {
			out.WriteString(";")
		}
		out.WriteString("\n")
	}

	return out.String()
}

// OldExpression represents a reference to a pre-execution value in a postcondition.
// The 'old' keyword captures the value of a variable/parameter before function execution.
// Syntax: old identifier (no parentheses)
// Examples:
//
//	old x
//	old val
//	Result = old count + 1
type OldExpression struct {
	TypedExpressionBase
	Identifier *Identifier
}

func (oe *OldExpression) expressionNode() {}

func (oe *OldExpression) String() string {
	return "old " + oe.Identifier.String()
}
