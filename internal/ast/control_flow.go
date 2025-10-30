// Package ast defines control flow AST node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
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
	Token       lexer.Token
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) Pos() lexer.Position  { return is.Token.Pos }
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
	Token     lexer.Token
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) Pos() lexer.Position  { return ws.Token.Pos }
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
	Token     lexer.Token
}

func (rs *RepeatStatement) statementNode()       {}
func (rs *RepeatStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RepeatStatement) Pos() lexer.Position  { return rs.Token.Pos }
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
	Start     Expression
	End       Expression
	Body      Statement
	Variable  *Identifier
	Token     lexer.Token
	Direction ForDirection
	InlineVar bool
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) Pos() lexer.Position  { return fs.Token.Pos }
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
	out.WriteString(fs.End.String())
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
	Token      lexer.Token // The 'for' token
	InlineVar  bool        // true if 'var' keyword used
}

func (fis *ForInStatement) statementNode()       {}
func (fis *ForInStatement) TokenLiteral() string { return fis.Token.Literal }
func (fis *ForInStatement) Pos() lexer.Position  { return fis.Token.Pos }
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
	Token     lexer.Token
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
	Token      lexer.Token
}

func (cs *CaseStatement) statementNode()       {}
func (cs *CaseStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CaseStatement) Pos() lexer.Position  { return cs.Token.Pos }
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
	Token lexer.Token // The 'break' token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) Pos() lexer.Position  { return bs.Token.Pos }
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
	Token lexer.Token // The 'continue' token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) Pos() lexer.Position  { return cs.Token.Pos }
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
	ReturnValue Expression // Optional expression returned from Exit (Task 9.199)
	Token       lexer.Token
}

func (es *ExitStatement) statementNode()       {}
func (es *ExitStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExitStatement) Pos() lexer.Position  { return es.Token.Pos }
func (es *ExitStatement) String() string {
	var out bytes.Buffer
	out.WriteString("Exit")
	if es.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(es.ReturnValue.String())
	}
	return out.String()
}
