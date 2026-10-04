package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// analyzeRecordCall selects a native record method before checking its supplied
// arguments. Only marked or multiple declarations use overload diagnostics.
func (a *Analyzer) analyzeRecordCall(overloads []*types.MethodInfo, args []ast.Expression, name string, pos token.Position) types.Type {
	selected := a.selectMemberCallOverload(overloads, args, name, pos)
	if selected == nil {
		return nil
	}
	a.analyzeMemberCallArguments(selected.Signature, args, pos, !selected.IsClassMethod)
	return selected.Signature.ReturnType
}

// implicitRecordCallOverloads limits implicit calls in class methods to the
// class side, which has no Self instance.
func (a *Analyzer) implicitRecordCallOverloads(name string) []*types.MethodInfo {
	classOverloads := a.currentRecord.GetClassMethodOverloads(name)
	if a.inClassMethod {
		return classOverloads
	}
	return append(append([]*types.MethodInfo{}, classOverloads...), a.currentRecord.GetMethodOverloads(name)...)
}
