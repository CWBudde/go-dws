package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// reportPointerAssignmentMismatch preserves both routine signatures and anchors
// the mismatch on the supplied expression. Child diagnostics (such as missing
// arguments on an incompatible reference) must precede this enclosing error.
func (a *Analyzer) reportPointerAssignmentMismatch(value ast.Expression, expected, actual types.Type) bool {
	if !types.IsPointerType(expected) || !types.IsPointerType(actual) {
		return false
	}
	err := NewIncompatibleTypesPairError(value.Pos(), semanticTypeNameForDiagnostic(expected), semanticTypeNameForDiagnostic(actual))
	err.AfterChildren = true
	a.addStructuredError(err)
	return true
}
