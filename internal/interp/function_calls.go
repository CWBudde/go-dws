package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// CallExternalFunction calls an external (Go) function with var parameter support.
func (i *Interpreter) CallExternalFunction(funcName string, argExprs []ast.Expression, node ast.Node) Value {
	// Check if this is an external function with var parameters
	if i.externalFunctions() == nil {
		return i.newErrorWithLocation(node, "external function registry not initialized")
	}

	registry := i.externalFunctions()
	extFunc, ok := registry.Get(funcName)
	if !ok {
		return i.newErrorWithLocation(node, "external function '%s' not found", funcName)
	}

	varParams := extFunc.Wrapper.GetVarParams()
	paramTypes := extFunc.Wrapper.GetParamTypes()

	// Prepare arguments - create ReferenceValues for var parameters
	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		isVarParam := idx < len(varParams) && varParams[idx]

		if isVarParam {
			// For var parameters, create a reference
			if argIdent, ok := arg.(*ast.Identifier); ok {
				if val, exists := i.Env().Get(argIdent.Value); exists {
					if refVal, isRef := val.(*ReferenceValue); isRef {
						args[idx] = refVal // Pass through existing reference
					} else {
						args[idx] = &ReferenceValue{Env: i.Env(), VarName: argIdent.Value}
					}
				} else {
					args[idx] = &ReferenceValue{Env: i.Env(), VarName: argIdent.Value}
				}
			} else {
				return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
			}
		} else {
			// For regular parameters, evaluate with type context if available
			var val Value
			if idx < len(paramTypes) {
				// Parse the parameter type string and provide context for type inference
				expectedType, _ := i.parseTypeString(paramTypes[idx])
				val = i.EvalWithExpectedType(arg, expectedType)
			} else {
				val = i.Eval(arg)
			}
			if isError(val) {
				return val
			}
			args[idx] = val
		}
	}

	return i.callExternalFunction(extFunc, args)
}

// parseTypeString parses a type string (e.g. "Integer", "array of String") into a types.Type.
func (i *Interpreter) parseTypeString(typeStr string) (types.Type, error) {
	return i.resolveType(typeStr)
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
