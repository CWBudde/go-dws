package ast

import "github.com/cwbudde/go-dws/pkg/token"

// DebugBreakExpression represents the reserved, valueless DebugBreak intrinsic.
// Unlike an ordinary procedure, it cannot be referenced or take arguments.
type DebugBreakExpression struct {
	BaseNode
	// NextTokenPos anchors value-context errors after the intrinsic, including
	// whitespace/comments and the last real token at EOF.
	NextTokenPos token.Position
	// Incomplete preserves casing hints when an optional '(' lacks its ')'.
	// Such a node is syntax only and must never reach execution.
	Incomplete bool
}

func (e *DebugBreakExpression) expressionNode() {}
func (e *DebugBreakExpression) String() string  { return "DebugBreak" }
