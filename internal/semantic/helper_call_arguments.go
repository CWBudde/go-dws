package semantic

import (
	"slices"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// analyzeHelperCallArguments preserves the selected signature's receiver role.
// Instance helpers (including function helpers) and nonstatic class helpers for
// structured targets pass Self as argument zero upstream. Static methods and
// class helpers for primitive/interface targets have no receiver argument.
func (a *Analyzer) analyzeHelperCallArguments(receiver types.Type, name string, signature *types.FunctionType, args []ast.Expression, pos token.Position) {
	hasSelf := true
	key := ident.Normalize(name)
	for _, helper := range a.getHelpersForType(receiver) {
		for owner := helper; owner != nil; owner = owner.ParentHelper {
			if owner.StaticMethods[signature] {
				hasSelf = false
			} else if slices.Contains(owner.ClassMethodOverloads[key], signature) {
				switch types.GetUnderlyingType(owner.TargetType).(type) {
				case *types.ClassType, *types.ClassOfType, *types.RecordType, *types.RecordMetaType:
					hasSelf = true
				default:
					hasSelf = false
				}
			}
		}
	}
	a.analyzeMemberCallArguments(signature, args, pos, hasSelf)
}
