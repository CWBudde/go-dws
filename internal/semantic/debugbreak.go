package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (a *Analyzer) analyzeDebugBreak(node *ast.DebugBreakExpression, statement bool) types.Type {
	a.addCaseMismatchHint(node.Token.Literal, "DebugBreak", node.Pos())
	if node.Incomplete {
		// The parser stop already owns this fragment. A recovery value avoids
		// initializer inference errors anchored before that stop.
		return types.VARIANT
	}
	if statement {
		a.semanticInfo.SetResolvedType(node, types.VOID)
		return types.VOID
	}
	pos := node.NextTokenPos
	if !pos.IsValid() {
		pos = node.End()
	}
	a.addStructuredError(NewGenericError(pos, "Expression expected"))
	// ReadTerm substitutes a bogus Variant value and continues upstream.
	return types.VARIANT
}
