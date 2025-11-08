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
	Condition   Expression
	Consequence Statement
	Alternative Statement
	Token       token.Token
	EndPos      token.Position
}

func (i *IfStatement) End() token.Position {
	if i.EndPos.Line != 0 {
		return i.EndPos
	}
	return i.Token.Pos
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) Pos() token.Position  { return is.Token.Pos }
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
	Condition Expression
	Body      Statement
	Token     token.Token
	EndPos    token.Position
}

func (w *WhileStatement) End() token.Position {
	if w.EndPos.Line != 0 {
		return w.EndPos
	}
	return w.Token.Pos
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) Pos() token.Position  { return ws.Token.Pos }
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
	Body      Statement
	Condition Expression
	Token     token.Token
	EndPos    token.Position
}

func (r *RepeatStatement) End() token.Position {
	if r.EndPos.Line != 0 {
		return r.EndPos
	}
	return r.Token.Pos
}

func (rs *RepeatStatement) statementNode()       {}
func (rs *RepeatStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RepeatStatement) Pos() token.Position  { return rs.Token.Pos }
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
	Start     Expression
	EndValue  Expression
	Body      Statement
	Step      Expression
	Variable  *Identifier
	Token     token.Token
	EndPos    token.Position
	Direction ForDirection
	InlineVar bool
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) Pos() token.Position  { return fs.Token.Pos }
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
	Variable   *Identifier // Loop variable
	Collection Expression  // Expression to iterate over (set, array, string, range)
	Body       Statement   // Loop body
	Token      token.Token // The 'for' token
	InlineVar  bool        // true if 'var' keyword used
	EndPos     token.Position
}

func (f *ForInStatement) End() token.Position {
	if f.EndPos.Line != 0 {
		return f.EndPos
	}
	return f.Token.Pos
}

func (fis *ForInStatement) statementNode()       {}
func (fis *ForInStatement) TokenLiteral() string { return fis.Token.Literal }
func (fis *ForInStatement) Pos() token.Position  { return fis.Token.Pos }
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
	Statement Statement
	Values    []Expression
	Token     token.Token
	EndPos    token.Position
}

func (c *CaseBranch) End() token.Position {
	if c.EndPos.Line != 0 {
		return c.EndPos
	}
	return c.Token.Pos
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
	Expression Expression
	Else       Statement
	Cases      []*CaseBranch
	Token      token.Token
	EndPos     token.Position
}

func (c *CaseStatement) End() token.Position {
	if c.EndPos.Line != 0 {
		return c.EndPos
	}
	return c.Token.Pos
}

func (cs *CaseStatement) statementNode()       {}
func (cs *CaseStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CaseStatement) Pos() token.Position  { return cs.Token.Pos }
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
	Token  token.Token // The 'break' token
	EndPos token.Position
}

func (b *BreakStatement) End() token.Position {
	if b.EndPos.Line != 0 {
		return b.EndPos
	}
	return b.Token.Pos
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) Pos() token.Position  { return bs.Token.Pos }
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
	Token  token.Token // The 'continue' token
	EndPos token.Position
}

func (c *ContinueStatement) End() token.Position {
	if c.EndPos.Line != 0 {
		return c.EndPos
	}
	return c.Token.Pos
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) Pos() token.Position  { return cs.Token.Pos }
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
	ReturnValue Expression // Optional expression returned from Exit
	Token       token.Token
	EndPos      token.Position
}

func (e *ExitStatement) End() token.Position {
	if e.EndPos.Line != 0 {
		return e.EndPos
	}
	return e.Token.Pos
}

func (es *ExitStatement) statementNode()       {}
func (es *ExitStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExitStatement) Pos() token.Position  { return es.Token.Pos }
func (es *ExitStatement) String() string {
	var out bytes.Buffer
	out.WriteString("Exit")
	if es.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(es.ReturnValue.String())
	}
	return out.String()
}
