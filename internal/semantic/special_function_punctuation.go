package semantic

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// specialFunctionName identifies the intrinsics whose bare names require '('.
// This is the pinned DWScript special-keyword set, excluding DebugBreak's
// optional parentheses and Default's fallback to ordinary name lookup.
func specialFunctionName(name string) string {
	return requiredSpecialFunctionNames[ident.Normalize(name)]
}

var requiredSpecialFunctionNames = map[string]string{
	"assert":             "Assert",
	"assigned":           "Assigned",
	"high":               "High",
	"length":             "Length",
	"low":                "Low",
	"ord":                "Ord",
	"sizeof":             "SizeOf",
	"defined":            "Defined",
	"declared":           "Declared",
	"inc":                "Inc",
	"dec":                "Dec",
	"succ":               "Succ",
	"pred":               "Pred",
	"include":            "Include",
	"exclude":            "Exclude",
	"swap":               "Swap",
	"conditionaldefined": "ConditionalDefined",
}

// stopBareSpecialFunction is called only after ordinary shadow resolution.
func (a *Analyzer) stopBareSpecialFunction(identifier *ast.Identifier) bool {
	name := specialFunctionName(identifier.Value)
	if name == "" {
		return false
	}
	a.addIdentifierCaseHint(identifier, name)
	a.addPunctuationStop(identifierLookaheadPos(identifier), `"(" expected`)
	return true
}

// stopSpecialFunctionAddress preserves ordinary binding lookup while requiring
// the special child's '(' before ReadAt could reject or capture the result.
func (a *Analyzer) stopSpecialFunctionAddress(target ast.Expression) bool {
	identifier := specialFunctionAddressIdentifier(target)
	if identifier == nil || specialFunctionName(identifier.Value) == "" || a.specialFunctionHasShadow(identifier.Value) {
		return false
	}
	return a.stopBareSpecialFunction(identifier)
}

// specialFunctionAddressIdentifier finds the name read before a group, index or
// member suffix. Completed calls retain the ordinary address-operand path.
func specialFunctionAddressIdentifier(target ast.Expression) *ast.Identifier {
	switch target := target.(type) {
	case *ast.Identifier:
		return target
	case *ast.GroupedExpression:
		return specialFunctionAddressIdentifier(target.Expression)
	case *ast.IndexExpression:
		return specialFunctionAddressIdentifier(target.Left)
	case *ast.MemberAccessExpression:
		return specialFunctionAddressIdentifier(target.Object)
	default:
		return nil
	}
}

// specialFunctionHasShadow probes bindings without analyzing the identifier:
// analysis would emit usage/casing diagnostics and implicitly invoke members.
func (a *Analyzer) specialFunctionHasShadow(name string) bool {
	if _, found := a.symbols.Resolve(name); found {
		return true
	}
	if class := a.currentClass; class != nil {
		if _, found := class.GetMethod(name); found {
			return true
		}
		if _, found := class.GetField(name); found || class.HasConstructor(name) {
			return true
		}
		for ; class != nil; class = class.Parent {
			for property := range class.Properties {
				if ident.Equal(property, name) {
					return true
				}
			}
			for constant := range class.ConstantTypes {
				if ident.Equal(constant, name) {
					return true
				}
			}
		}
	}
	if selfType := a.currentImplicitSelfType(); selfType != nil {
		return a.hasHelperMethod(selfType, name) != nil
	}
	return false
}
