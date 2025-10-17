// Package ast defines control flow AST node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/lexer"
)

// IfStatement represents an if-then-else conditional statement.
// Examples:
//
//	if x > 0 then PrintLn('positive');
//	if x > 0 then PrintLn('positive') else PrintLn('non-positive');
//	if condition then begin ... end;
type IfStatement struct {
	Token       lexer.Token // The 'if' token
	Condition   Expression  // The condition expression
	Consequence Statement   // The 'then' branch
	Alternative Statement   // The 'else' branch (optional, can be nil)
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(is.Condition.String())
	out.WriteString(" then ")
	out.WriteString(is.Consequence.String())

	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}

	return out.String()
}

// WhileStatement represents a while loop.
// Examples:
//
//	while x < 10 do x := x + 1;
//	while condition do begin ... end;
type WhileStatement struct {
	Token     lexer.Token // The 'while' token
	Condition Expression  // The loop condition
	Body      Statement   // The loop body
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer

	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" do ")
	out.WriteString(ws.Body.String())

	return out.String()
}

// RepeatStatement represents a repeat-until loop.
// The body executes at least once, then repeats until the condition becomes true.
// Examples:
//
//	repeat x := x + 1; until x >= 10;
//	repeat begin ... end; until condition;
type RepeatStatement struct {
	Token     lexer.Token // The 'repeat' token
	Body      Statement   // The loop body (can be a block or single statement)
	Condition Expression  // The 'until' condition
}

func (rs *RepeatStatement) statementNode()       {}
func (rs *RepeatStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RepeatStatement) String() string {
	var out bytes.Buffer

	out.WriteString("repeat ")
	out.WriteString(rs.Body.String())
	out.WriteString(" until ")
	out.WriteString(rs.Condition.String())

	return out.String()
}

// ForDirection represents the direction of a for loop (to or downto).
type ForDirection int

const (
	ForTo ForDirection = iota
	ForDownto
)

func (fd ForDirection) String() string {
	switch fd {
	case ForTo:
		return "to"
	case ForDownto:
		return "downto"
	default:
		return "unknown"
	}
}

// ForStatement represents a for loop.
// Examples:
//
//	for i := 1 to 10 do PrintLn(i);
//	for i := 10 downto 1 do PrintLn(i);
//	for i := start to end do begin ... end;
type ForStatement struct {
	Token     lexer.Token  // The 'for' token
	Variable  *Identifier  // The loop variable
	Start     Expression   // The start value
	End       Expression   // The end value
	Direction ForDirection // The direction (to or downto)
	Body      Statement    // The loop body
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(fs.Variable.String())
	out.WriteString(" := ")
	out.WriteString(fs.Start.String())
	out.WriteString(" ")
	out.WriteString(fs.Direction.String())
	out.WriteString(" ")
	out.WriteString(fs.End.String())
	out.WriteString(" do ")
	out.WriteString(fs.Body.String())

	return out.String()
}

// CaseBranch represents a single branch in a case statement.
// Examples:
//
//	1: PrintLn('one');
//	2, 3, 4: PrintLn('two to four');
type CaseBranch struct {
	Token     lexer.Token  // The first value token
	Values    []Expression // The values that match this branch
	Statement Statement    // The statement to execute
}

func (cb *CaseBranch) String() string {
	var out bytes.Buffer

	values := []string{}
	for _, v := range cb.Values {
		values = append(values, v.String())
	}
	out.WriteString(strings.Join(values, ", "))
	out.WriteString(": ")
	out.WriteString(cb.Statement.String())

	return out.String()
}

// CaseStatement represents a case statement (switch/case).
// Examples:
//
//	case x of
//	  1: PrintLn('one');
//	  2, 3: PrintLn('two or three');
//	else
//	  PrintLn('other');
//	end;
type CaseStatement struct {
	Token      lexer.Token   // The 'case' token
	Expression Expression    // The expression to match against
	Cases      []*CaseBranch // The case branches
	Else       Statement     // The else branch (optional, can be nil)
}

func (cs *CaseStatement) statementNode()       {}
func (cs *CaseStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CaseStatement) String() string {
	var out bytes.Buffer

	out.WriteString("case ")
	out.WriteString(cs.Expression.String())
	out.WriteString(" of\n")

	for _, c := range cs.Cases {
		out.WriteString("  ")
		out.WriteString(c.String())
		out.WriteString("\n")
	}

	if cs.Else != nil {
		out.WriteString("else\n  ")
		out.WriteString(cs.Else.String())
		out.WriteString("\n")
	}

	out.WriteString("end")

	return out.String()
}
