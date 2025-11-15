// Package ast defines control flow AST node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// IfStatement represents an if-then-else conditional statement.
// Examples:
//
//	if x > 0 then PrintLn('positive');
//	if x > 0 then PrintLn('positive') else PrintLn('non-positive');
//	if condition then begin ... end;
type IfStatement struct {
	BaseNode
	Condition   Expression
	Consequence Statement
	Alternative Statement
}

func (is *IfStatement) statementNode() {}

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

// IfExpression represents an inline if-then-else conditional expression.
// This is similar to a ternary operator in other languages.
// Examples:
//
//	x := if b then 1 else 0;
//	PrintLn(if condition then 'yes' else 'no');
//	var o := if b then TObject.Create else nil;
//	var i := if b then 1;  // else clause optional, returns default value (0)
type IfExpression struct {
	TypedExpressionBase
	Condition   Expression
	Consequence Expression
	Alternative Expression // Optional, can be nil
}

func (ie *IfExpression) expressionNode() {}

// End returns the end position, preferring Alternative or Consequence if EndPos not set.
func (ie *IfExpression) End() token.Position {
	if ie.EndPos.Line != 0 {
		return ie.EndPos
	}
	// End position is at the end of Alternative (if present) or Consequence
	if ie.Alternative != nil {
		return ie.Alternative.End()
	}
	if ie.Consequence != nil {
		return ie.Consequence.End()
	}
	return ie.Token.Pos
}

func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" then ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(ie.Alternative.String())
	}

	out.WriteString(")")

	return out.String()
}

// WhileStatement represents a while loop.
// Examples:
//
//	while x < 10 do x := x + 1;
//	while condition do begin ... end;
type WhileStatement struct {
	BaseNode
	Condition Expression
	Body      Statement
}

func (ws *WhileStatement) statementNode() {}

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
	BaseNode
	Body      Statement
	Condition Expression
}

func (rs *RepeatStatement) statementNode() {}

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
//	for i := 1 to 10 step 2 do PrintLn(i);
type ForStatement struct {
	BaseNode
	Start     Expression
	EndValue  Expression
	Body      Statement
	Step      Expression
	Variable  *Identifier
	Direction ForDirection
	InlineVar bool
}

func (fs *ForStatement) statementNode() {}

// End returns the end position, preferring Body's end if EndPos not set.
func (fs *ForStatement) End() token.Position {
	if fs.EndPos.Line != 0 {
		return fs.EndPos
	}
	if fs.Body != nil {
		return fs.Body.End()
	}
	return fs.Token.Pos
}

func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	if fs.InlineVar {
		out.WriteString("var ")
	}
	out.WriteString(fs.Variable.String())
	out.WriteString(" := ")
	out.WriteString(fs.Start.String())
	out.WriteString(" ")
	out.WriteString(fs.Direction.String())
	out.WriteString(" ")
	out.WriteString(fs.EndValue.String())
	if fs.Step != nil {
		out.WriteString(" step ")
		out.WriteString(fs.Step.String())
	}
	out.WriteString(" do ")
	out.WriteString(fs.Body.String())

	return out.String()
}

// ForInStatement represents a for-in loop that iterates over a collection.
// Examples:
//
//	for e in mySet do PrintLn(e);
//	for var item in myArray do PrintLn(item);
//	for ch in "hello" do Print(ch);
type ForInStatement struct {
	BaseNode
	Variable   *Identifier // Loop variable
	Collection Expression  // Expression to iterate over (set, array, string, range)
	Body       Statement   // Loop body
	InlineVar  bool        // true if 'var' keyword used
}

func (fis *ForInStatement) statementNode() {}

func (fis *ForInStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	if fis.InlineVar {
		out.WriteString("var ")
	}
	out.WriteString(fis.Variable.String())
	out.WriteString(" in ")
	out.WriteString(fis.Collection.String())
	out.WriteString(" do ")
	out.WriteString(fis.Body.String())

	return out.String()
}

// CaseBranch represents a single branch in a case statement.
// Examples:
//
//	1: PrintLn('one');
//	2, 3, 4: PrintLn('two to four');
type CaseBranch struct {
	BaseNode
	Statement Statement
	Values    []Expression
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
	BaseNode
	Expression Expression
	Else       Statement
	Cases      []*CaseBranch
}

func (cs *CaseStatement) statementNode() {}

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

// BreakStatement represents a break statement that exits the innermost loop.
// Examples:
//
//	for i := 1 to 10 do begin
//	   if i > 5 then break;
//	end;
//
//	while condition do begin
//	   if shouldExit then break;
//	end;
type BreakStatement struct {
	BaseNode
}

func (bs *BreakStatement) statementNode() {}

func (bs *BreakStatement) String() string {
	return "break;"
}

// ContinueStatement represents a continue statement that skips to the next iteration of the innermost loop.
// Examples:
//
//	for i := 1 to 10 do begin
//	   if (i and 1) = 0 then continue;
//	   PrintLn(i);  // Only prints odd numbers
//	end;
//
//	while condition do begin
//	   if shouldSkip then begin
//	      counter := counter + 1;  // Update before continue!
//	      continue;
//	   end;
//	end;
type ContinueStatement struct {
	BaseNode
}

func (cs *ContinueStatement) statementNode() {}

func (cs *ContinueStatement) String() string {
	return "continue;"
}

// ExitStatement represents an exit statement that exits the current function or procedure.
// Examples:
//
//	procedure Test;
//	begin
//	   if condition then exit;
//	   PrintLn('still here');
//	end;
//
//	function MyFunc(i: Integer): Integer;
//	begin
//	   if i <= 0 then exit(-1);  // Exit with value
//	   Result := i * 2;
//	end;
type ExitStatement struct {
	BaseNode
	ReturnValue Expression // Optional expression returned from Exit
}

func (es *ExitStatement) statementNode() {}

func (es *ExitStatement) String() string {
	var out bytes.Buffer
	out.WriteString("Exit")
	if es.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(es.ReturnValue.String())
	}
	return out.String()
}
