package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// GetZeroValueForType returns the zero/default value for a given type.
func (i *Interpreter) GetZeroValueForType(typeInfo any) evaluator.Value {
	t, ok := typeInfo.(types.Type)
	if !ok {
		return &NilValue{}
	}

	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.Env().Get(key); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return rtv.Methods
			}
		}
		return nil
	}

	return getZeroValueForType(t, methodsLookup)
}

// ============================================================================
// Exception Handling and Type Conversion Adapter Methods
// ============================================================================
//
// These methods are implemented in builtins_context.go as part of the
// builtins.Context interface. They take builtins.Value (runtime.Value) and
// are used by both the Interpreter and via the adapter bridge to the Evaluator.
