package semantic

import (
	"unicode/utf8"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// interruptFlow mirrors DWScript's TInterruptFlowType. A loop consumes break
// and continue, whereas an exit or raise can interrupt the enclosing routine.
type interruptFlow uint8

const (
	flowNone interruptFlow = iota
	flowLoop
	flowProcedure
)

// statementInterruptsFlow follows upstream's unoptimized InterruptsFlow
// methods. In particular an if without else is not assumed to execute its
// branch, even when its condition is True. Calls and exception handlers do not
// propagate flow information from their bodies.
func statementInterruptsFlow(stmt ast.Statement) interruptFlow {
	switch s := stmt.(type) {
	case *ast.BreakStatement, *ast.ContinueStatement:
		return flowLoop
	case *ast.ExitStatement, *ast.ReturnStatement, *ast.RaiseStatement:
		return flowProcedure
	case *ast.BlockStatement:
		for _, child := range s.Statements {
			if flow := statementInterruptsFlow(child); flow != flowNone {
				return flow
			}
		}
	case *ast.IfStatement:
		return min(statementInterruptsFlow(s.Consequence), statementInterruptsFlow(s.Alternative))
	case *ast.CaseStatement:
		flow := statementInterruptsFlow(s.Else)
		for _, branch := range s.Cases {
			flow = min(flow, statementInterruptsFlow(branch.Statement))
		}
		return flow
	case *ast.RepeatStatement:
		if statementInterruptsFlow(s.Body) == flowProcedure {
			return flowProcedure
		}
	case *ast.WhileStatement:
		if condition, ok := s.Condition.(*ast.BooleanLiteral); ok && condition.Value && statementInterruptsFlow(s.Body) == flowProcedure {
			return flowProcedure
		}
	}
	return flowNone
}

// reachabilityState warns once per statement list, before the unreachable
// statement is analyzed. Invalid jumps still interrupt flow upstream, and the
// remaining statements must still be analyzed to report their own diagnostics.
type reachabilityState struct {
	interrupted bool
	warned      bool
}

func (r *reachabilityState) before(a *Analyzer, stmt ast.Statement, root bool) {
	if stmt == nil || !r.interrupted || r.warned {
		return
	}
	r.warned = true
	pos := unreachableStatementPos(stmt, root)
	a.addWarning("Unreachable code [line: %d, column: %d]", pos.Line, pos.Column)
}

func (r *reachabilityState) after(stmt ast.Statement) {
	if !r.interrupted && statementInterruptsFlow(stmt) != flowNone {
		r.interrupted = true
	}
}

// unreachableStatementPos mirrors ReadRootBlock's HotPos and ReadBlocks'
// CurrentPos. The latter is the scanner position just after the first token.
// Assignments carry the operator token, so their first token comes from their
// target instead. Expression statements already retain their starting token.
func unreachableStatementPos(stmt ast.Statement, root bool) token.Position {
	var first ast.Node = stmt
	if assignment, ok := stmt.(*ast.AssignmentStatement); ok && assignment.Target != nil {
		first = assignment.Target
		if name := leadingIdentifier(assignment.Target); name != nil {
			first = name
		}
	}
	pos := first.Pos()
	if !root {
		literal := first.TokenLiteral()
		pos.Column += utf8.RuneCountInString(literal)
		pos.Offset += len(literal)
	}
	return pos
}
