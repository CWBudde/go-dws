// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"
)

// TryStatement represents a try...except...finally...end statement.
// Supports three forms:
//   - try...except...end
//   - try...finally...end
//   - try...except...finally...end
//
// Examples:
//
//	try
//	  raise Exception.Create('error');
//	except
//	  on E: Exception do
//	    PrintLn(E.Message);
//	end;
//
//	try
//	  DoSomething();
//	finally
//	  Cleanup();
//	end;
type TryStatement struct {
	BaseNode
	TryBlock      *BlockStatement
	ExceptClause  *ExceptClause
	FinallyClause *FinallyClause
}

func (ts *TryStatement) statementNode() {}
func (ts *TryStatement) String() string {
	var out bytes.Buffer

	out.WriteString("try")
	if ts.TryBlock != nil {
		out.WriteString("\n  ")
		out.WriteString(strings.ReplaceAll(ts.TryBlock.String(), "\n", "\n  "))
	}

	if ts.ExceptClause != nil {
		out.WriteString("\n")
		out.WriteString(ts.ExceptClause.String())
	}

	if ts.FinallyClause != nil {
		out.WriteString("\n")
		out.WriteString(ts.FinallyClause.String())
	}

	out.WriteString("\nend")

	return out.String()
}

// ExceptClause represents the except part of a try statement.
// Can contain:
//   - Specific exception handlers (on E: Type do ...)
//   - Bare except (no handlers, catches all)
//   - Optional else block (executes if no exception)
//
// Examples:
//
//	except
//	  on E: Exception do
//	    PrintLn(E.Message);
//	end
//
//	except
//	  on E: EMyException do
//	    HandleMyException(E);
//	  on E: Exception do
//	    HandleGeneric(E);
//	end
type ExceptClause struct {
	BaseNode
	ElseBlock *BlockStatement
	Handlers  []*ExceptionHandler
}

func (ec *ExceptClause) String() string {
	var out bytes.Buffer

	out.WriteString("except")

	if len(ec.Handlers) > 0 {
		for _, handler := range ec.Handlers {
			out.WriteString("\n  ")
			out.WriteString(strings.ReplaceAll(handler.String(), "\n", "\n  "))
		}
	}

	if ec.ElseBlock != nil {
		out.WriteString("\nelse")
		out.WriteString("\n  ")
		out.WriteString(strings.ReplaceAll(ec.ElseBlock.String(), "\n", "\n  "))
	}

	return out.String()
}

// ExceptionHandler represents a specific exception handler in an except clause.
// Syntax: on <variable>: <type> do <statement>
//
// Examples:
//
//	on E: Exception do
//	  PrintLn(E.Message);
//
//	on E: EMyException do begin
//	  HandleMyException(E);
//	  raise;
//	end;
type ExceptionHandler struct {
	BaseNode
	Statement     Statement
	Variable      *Identifier
	ExceptionType *TypeAnnotation
}

func (eh *ExceptionHandler) String() string {
	var out bytes.Buffer

	out.WriteString("on ")
	if eh.Variable != nil {
		out.WriteString(eh.Variable.String())
	}
	out.WriteString(": ")
	if eh.ExceptionType != nil {
		out.WriteString(eh.ExceptionType.String())
	}
	out.WriteString(" do ")
	if eh.Statement != nil {
		out.WriteString(eh.Statement.String())
	}

	return out.String()
}

// FinallyClause represents the finally part of a try statement.
// Syntax: finally <statements> end
//
// The finally block always executes, even if:
//   - No exception occurs
//   - An exception occurs and is caught
//   - An exception occurs and is not caught
//   - Exit, Break, or Continue is executed in the try or except block
//
// Example:
//
//	try
//	  DoSomething();
//	finally
//	  Cleanup();
//	end;
type FinallyClause struct {
	BaseNode
	Block *BlockStatement
}

func (fc *FinallyClause) String() string {
	var out bytes.Buffer

	out.WriteString("finally")
	if fc.Block != nil {
		out.WriteString("\n  ")
		out.WriteString(strings.ReplaceAll(fc.Block.String(), "\n", "\n  "))
	}

	return out.String()
}

// RaiseStatement represents a raise statement.
// Can be:
//   - raise <expression>; (raise new exception)
//   - raise; (re-raise current exception, only valid in except block)
//
// Examples:
//
//	raise Exception.Create('error');
//	raise new EMyException('custom error');
//	raise; // re-raise
type RaiseStatement struct {
	BaseNode
	Exception Expression
}

func (rs *RaiseStatement) statementNode() {}
func (rs *RaiseStatement) String() string {
	var out bytes.Buffer

	out.WriteString("raise")
	if rs.Exception != nil {
		out.WriteString(" ")
		out.WriteString(rs.Exception.String())
	}

	return out.String()
}
