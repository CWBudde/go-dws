package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// checkDefaultNamespaceMember follows the standard result unit's local table.
// Special pseudo-symbols and other units' globals are not members of Default.
func (a *Analyzer) checkDefaultNamespaceMember(object ast.Expression, member *ast.Identifier) bool {
	a.semanticInfo.SetDefaultNamespace(object)
	if qualifier, ok := object.(*ast.Identifier); ok {
		a.addIdentifierCaseHint(qualifier, "Default")
	}
	if ident.Equal(member.Value, "Print") || ident.Equal(member.Value, "PrintLn") {
		return true
	}
	a.addPunctuationStop(member.Pos(), fmt.Sprintf(`Unknown name "Default.%s"`, member.Value))
	return false
}

func (a *Analyzer) analyzeDefaultNamespaceCall(object ast.Expression, member *ast.Identifier, call *ast.CallExpression) types.Type {
	if !a.checkDefaultNamespaceMember(object, member) {
		return nil
	}
	result, _ := a.analyzeBuiltinFunction(member.Value, call.Arguments, call)
	return result
}

func (a *Analyzer) analyzeDefaultNamespaceMember(member *ast.MemberAccessExpression, expected types.Type, address bool) types.Type {
	if !a.checkDefaultNamespaceMember(member.Object, member.Member) {
		return nil
	}
	pointer := a.getBuiltinFunctionPointerType(member.Member.Value)
	if address || expected != nil && types.IsPointerType(expected) && a.canAssign(pointer, expected) {
		a.semanticInfo.SetResolvedType(member.Member, pointer)
		return pointer
	}
	// A bare procedure is a call unless a compatible callback was expected.
	a.semanticInfo.SetImplicitCall(member)
	return types.VOID
}
