package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (i *Interpreter) globalFunctionOverloads(name string) []*ast.FunctionDecl {
	if i.typeSystem == nil {
		return nil
	}
	return i.typeSystem.LookupFunctions(name)
}
