package interp

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// callBuiltin dispatches built-in and external functions by name.
// It normalizes the function name for DWScript's case-insensitive matching,
// and routes to the appropriate built-in implementation or external Go function.
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
	// Check for external Go functions first
	if i.evaluatorInstance.ExternalFunctions() != nil {
		// Type-assert to concrete type to access Get method
		if registry, ok := i.evaluatorInstance.ExternalFunctions().(*ExternalFunctionRegistry); ok {
			if extFunc, ok := registry.Get(name); ok {
				return i.callExternalFunction(extFunc, args)
			}
		}
	}

	// Check the built-in function registry (case-insensitive lookup)
	if fn, ok := builtins.DefaultRegistry.Lookup(name); ok {
		return fn(i, args)
	}

	// Normalize function name for case-insensitive matching (DWScript is case-insensitive)
	name = normalizeBuiltinName(name)

	// Functions not yet migrated to the builtins registry
	// These are either:
	// - Not yet migrated (HexToInt, BinToInt, VarToIntDef, VarToFloatDef)
	// - Need interpreter access for callbacks (Map, Filter, Reduce, etc.)
	switch name {
	case "HexToInt":
		return i.builtinHexToInt(args)
	case "BinToInt":
		return i.builtinBinToInt(args)
	case "VarToIntDef":
		return i.builtinVarToIntDef(args)
	case "VarToFloatDef":
		return i.builtinVarToFloatDef(args)
	// Higher-order functions for working with arrays and lambdas
	// These need interpreter access for callback evaluation
	case "Map":
		return i.builtinMap(args)
	case "Filter":
		return i.builtinFilter(args)
	case "Reduce":
		return i.builtinReduce(args)
	case "ForEach":
		return i.builtinForEach(args)
	case "Every":
		return i.builtinEvery(args)
	case "Some":
		return i.builtinSome(args)
	case "Find":
		return i.builtinFind(args)
	case "FindIndex":
		return i.builtinFindIndex(args)
	case "Slice":
		return i.builtinSlice(args)
	default:
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "undefined function: %s", name)
	}
}

// callExternalFunction calls an external Go function registered via FFI
// It uses the existing FFI error handling infrastructure to safely call the Go function
// and convert any errors or panics to DWScript exceptions.
func (i *Interpreter) callExternalFunction(extFunc *ExternalFunctionValue, args []Value) Value {
	// Set interpreter reference for callback support
	// This allows the FFI wrapper to create Go callbacks that call back into DWScript
	extFunc.Wrapper.SetInterpreter(i)

	// Use the existing callExternalFunctionSafe wrapper which handles panics
	// and converts them to EHost exceptions (from ffi_errors.go)
	return i.callExternalFunctionSafe(func() (Value, error) {
		// Call the wrapped Go function
		return extFunc.Wrapper.Call(args)
	})
}

// callBuiltinWithVarParam calls a built-in function that requires var parameters.
// These functions need access to the AST nodes to modify variables in place.
// The implementations are in internal/builtins/var_param.go.
func (i *Interpreter) callBuiltinWithVarParam(name string, args []ast.Expression) Value {
	if fn, ok := builtins.VarParamFunctions[name]; ok {
		return fn(i, args)
	}
	return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "undefined var-param function: %s", name)
}
