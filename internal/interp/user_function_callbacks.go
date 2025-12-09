package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// createUserFunctionCallbacks creates the callback struct for ExecuteUserFunction.
func (i *Interpreter) createUserFunctionCallbacks() *evaluator.UserFunctionCallbacks {
	return &evaluator.UserFunctionCallbacks{
		ImplicitConversion:   i.createImplicitConversionCallback(),
		DefaultValueGetter:   i.createDefaultValueGetterCallback(),
		FunctionNameAlias:    i.createFunctionNameAliasCallback(),
		ReturnValueConverter: i.createReturnValueConverterCallback(),
		InterfaceRefCounter:  i.createInterfaceRefCounterCallback(),
		InterfaceCleanup:     i.createInterfaceCleanupCallback(),
		EnvSyncer:            i.createEnvSyncerCallback(),
	}
}

// createImplicitConversionCallback creates the parameter type conversion callback.
func (i *Interpreter) createImplicitConversionCallback() evaluator.ImplicitConversionFunc {
	return func(value evaluator.Value, targetTypeName string) (evaluator.Value, bool) {
		return i.tryImplicitConversion(value, targetTypeName)
	}
}

// createDefaultValueGetterCallback creates the return type default value callback.
func (i *Interpreter) createDefaultValueGetterCallback() evaluator.DefaultValueFunc {
	return func(returnTypeName string) evaluator.Value {
		lowerReturnType := ident.Normalize(returnTypeName)
		var defaultValue evaluator.Value

		// Basic type defaults
		switch lowerReturnType {
		case "integer":
			defaultValue = &IntegerValue{Value: 0}
		case "float":
			defaultValue = &FloatValue{Value: 0.0}
		case "string":
			defaultValue = &StringValue{Value: ""}
		case "boolean":
			defaultValue = &BooleanValue{Value: false}
		default:
			defaultValue = &NilValue{}
		}

		// Check for record type
		lowerReturnType = ident.Normalize(returnTypeName)
		if recordTypeValueAny := i.typeSystem.LookupRecord(returnTypeName); recordTypeValueAny != nil {
			if rtv, ok := recordTypeValueAny.(*RecordTypeValue); ok {
				return i.createRecordValue(rtv.RecordType)
			}
		}

		// Check for array type
		if arrayType := i.typeSystem.LookupArrayType(returnTypeName); arrayType != nil {
			return NewArrayValue(arrayType)
		} else if strings.HasPrefix(lowerReturnType, "array") {
			if arrayType := i.parseInlineArrayType(lowerReturnType); arrayType != nil {
				return NewArrayValue(arrayType)
			}
		}

		// Check for interface type
		if interfaceInfo := i.lookupInterfaceInfo(lowerReturnType); interfaceInfo != nil {
			return &InterfaceInstance{
				Interface: interfaceInfo,
				Object:    nil,
			}
		}

		return defaultValue
	}
}

// createFunctionNameAliasCallback creates the function name alias callback.
// In DWScript, assigning to the function name sets the return value (alias for Result).
func (i *Interpreter) createFunctionNameAliasCallback() evaluator.FunctionNameAliasFunc {
	return func(funcName string, funcEnv *runtime.Environment) evaluator.Value {
		// Phase 3.1.3: Direct runtime.Environment - no adapter unwrapping needed
		return &ReferenceValue{Env: funcEnv, VarName: "Result"}
	}
}

// createReturnValueConverterCallback creates the return value conversion callback.
func (i *Interpreter) createReturnValueConverterCallback() evaluator.TryImplicitConversionReturnFunc {
	return func(returnValue evaluator.Value, expectedReturnType string) (evaluator.Value, bool) {
		return i.tryImplicitConversion(returnValue, expectedReturnType)
	}
}

// createInterfaceRefCounterCallback creates the interface ref count increment callback.
// Increments ref count when returning an interface for the caller's reference.
func (i *Interpreter) createInterfaceRefCounterCallback() evaluator.IncrementInterfaceRefCountFunc {
	return func(returnValue evaluator.Value) {
		if intfInst, isIntf := returnValue.(*InterfaceInstance); isIntf {
			if intfInst.Object != nil {
				i.evaluatorInstance.RefCountManager().IncrementRef(intfInst.Object)
			}
		}
	}
}

// createInterfaceCleanupCallback creates the interface/object cleanup callback.
// Cleans up interface/object references when function scope ends.
func (i *Interpreter) createInterfaceCleanupCallback() evaluator.CleanupInterfaceReferencesFunc {
	return func(env *runtime.Environment) {
		// Phase 3.1.3: Direct runtime.Environment - no adapter unwrapping needed
		savedEnv := i.env
		i.SetEnvironment(env)
		i.cleanupInterfaceReferences(i.env)
		i.RestoreEnvironment(savedEnv)
	}
}

// createEnvSyncerCallback creates the environment synchronization callback.
// Syncs i.env with funcEnv so interpreter callbacks use the correct scope.
func (i *Interpreter) createEnvSyncerCallback() evaluator.EnvSyncerFunc {
	return func(funcEnv *runtime.Environment) func() {
		// Phase 3.1.3: Direct runtime.Environment - no adapter unwrapping needed
		savedEnv := i.env
		i.SetEnvironment(funcEnv)

		return func() {
			i.RestoreEnvironment(savedEnv)
		}
	}
}
