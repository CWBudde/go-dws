package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/contracts"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// createUserFunctionCallbacks creates the callback struct for ExecuteUserFunction.
func (i *Interpreter) createUserFunctionCallbacks() *contracts.UserFunctionCallbacks {
	return &contracts.UserFunctionCallbacks{
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
func (i *Interpreter) createImplicitConversionCallback() contracts.ImplicitConversionFunc {
	return func(value Value, targetTypeName string) (Value, bool) {
		return i.tryImplicitConversion(value, targetTypeName)
	}
}

// createDefaultValueGetterCallback creates the return type default value callback.
func (i *Interpreter) createDefaultValueGetterCallback() contracts.DefaultValueFunc {
	return func(returnTypeName string) Value {
		lowerReturnType := ident.Normalize(returnTypeName)
		var defaultValue Value

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
func (i *Interpreter) createFunctionNameAliasCallback() contracts.FunctionNameAliasFunc {
	return func(funcName string, funcEnv *runtime.Environment) Value {
		return &ReferenceValue{Env: funcEnv, VarName: "Result"}
	}
}

// createReturnValueConverterCallback creates the return value conversion callback.
func (i *Interpreter) createReturnValueConverterCallback() contracts.TryImplicitConversionReturnFunc {
	return func(returnValue Value, expectedReturnType string) (Value, bool) {
		return i.tryImplicitConversion(returnValue, expectedReturnType)
	}
}

// createInterfaceRefCounterCallback creates the interface ref count increment callback.
// Increments ref count when returning an interface for the caller's reference.
func (i *Interpreter) createInterfaceRefCounterCallback() contracts.IncrementInterfaceRefCountFunc {
	return func(returnValue Value) {
		if intfInst, isIntf := returnValue.(*InterfaceInstance); isIntf {
			if intfInst.Object != nil {
				i.evaluatorInstance.RefCountManager().IncrementRef(intfInst.Object)
			}
		}
	}
}

// createInterfaceCleanupCallback creates the interface/object cleanup callback.
// Cleans up interface/object references when function scope ends.
func (i *Interpreter) createInterfaceCleanupCallback() contracts.CleanupInterfaceReferencesFunc {
	return func(env *runtime.Environment) {
		savedEnv := i.Env()
		i.SetEnvironment(env)
		i.cleanupInterfaceReferences(i.Env())
		i.RestoreEnvironment(savedEnv)
	}
}

// createEnvSyncerCallback creates the environment synchronization callback.
// Syncs i.Env() with funcEnv so interpreter callbacks use the correct scope.
func (i *Interpreter) createEnvSyncerCallback() contracts.EnvSyncerFunc {
	return func(funcEnv *runtime.Environment) func() {
		savedEnv := i.Env()
		i.SetEnvironment(funcEnv)

		return func() {
			i.RestoreEnvironment(savedEnv)
		}
	}
}
