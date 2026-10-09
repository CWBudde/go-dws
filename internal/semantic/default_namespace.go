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

// stopUnitQualifiedSpecialName stops System.<special> and Internal.<special>.
// Upstream identifies special keywords only for an unqualified name, and a
// unit prefix looks the member up in that unit's own table, which holds none,
// so the lookup fails at the member with CPE_UnknownNameDotName.
func (a *Analyzer) stopUnitQualifiedSpecialName(object ast.Expression, member *ast.Identifier) bool {
	qualifier, ok := object.(*ast.Identifier)
	if !ok || member == nil || !isSpecialKeywordName(member.Value) {
		return false
	}
	var unit string
	switch ident.Normalize(qualifier.Value) {
	case "system":
		unit = "System"
	case "internal":
		unit = "Internal"
	default:
		return false
	}
	if _, resolved := a.symbols.Resolve(qualifier.Value); resolved || a.hasLexicalValueReceiver(object) {
		return false
	}
	if _, imported := a.importedUnitNamespace(qualifier.Value); imported {
		return false
	}
	a.addPunctuationStop(member.Pos(), fmt.Sprintf(`Unknown name "%s.%s"`, unit, member.Value))
	return true
}
