package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// CallExternalFunction invokes an external Go function with evaluator-prepared arguments.
func (i *Interpreter) CallExternalFunction(funcName string, args []Value) Value {
	if i.externalFunctions() == nil {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "external function registry not initialized")
	}

	registry := i.externalFunctions()
	extFunc, ok := registry.Get(funcName)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "external function '%s' not found", funcName)
	}

	return i.callExternalFunction(extFunc, args)
}

// ============================================================================
// Function Declaration Methods
// ============================================================================

// EvalMethodImplementation handles method implementation registration for classes/records.
// Delegated from Evaluator.VisitFunctionDecl because it requires ClassInfo internals.
func (i *Interpreter) EvalMethodImplementation(fn *ast.FunctionDecl) Value {
	if fn == nil || fn.ClassName == nil {
		return i.newErrorWithLocation(fn, "EvalMethodImplementation requires a method declaration with ClassName")
	}

	typeName := fn.ClassName.Value

	// Check if class first (case-insensitive lookup)
	classInfo := i.lookupRegisteredClassInfo(typeName)
	if classInfo != nil {
		i.evalClassMethodImplementation(fn, classInfo)
		return &NilValue{}
	}

	if i.typeSystem != nil {
		if recordInfoAny := i.typeSystem.LookupRecord(typeName); recordInfoAny != nil {
			if recordInfo, ok := recordInfoAny.(*runtime.RecordTypeValue); ok {
				i.evalRecordMethodImplementation(fn, recordInfo)
				return &NilValue{}
			}
		}
	}

	return i.newErrorWithLocation(fn, "type '%s' not found for method '%s'", typeName, fn.Name.Value)
}
