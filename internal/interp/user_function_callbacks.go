package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
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
	return func(funcName string, funcEnv evaluator.Environment) evaluator.Value {
		if adapter, ok := funcEnv.(*evaluator.EnvironmentAdapter); ok {
			if env, ok := adapter.Underlying().(*Environment); ok {
				return &ReferenceValue{Env: env, VarName: "Result"}
			}
		}
		return &NilValue{}
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
	return func(env evaluator.Environment) {
		savedEnv := i.env
		i.cleanupInterfaceReferencesForEnv(env)
		i.RestoreEnvironment(savedEnv)
	}
}

// cleanupInterfaceReferencesForEnv extracts concrete *Environment and performs cleanup.
func (i *Interpreter) cleanupInterfaceReferencesForEnv(env evaluator.Environment) {
	savedEnv := i.env
	var concreteEnv *Environment

	// Extract concrete *Environment from adapter
	if adapter, ok := env.(*evaluator.EnvironmentAdapter); ok {
		if underlying, ok := adapter.Underlying().(*Environment); ok {
			concreteEnv = underlying
		}
	} else if directEnv, ok := interface{}(env).(*Environment); ok {
		concreteEnv = directEnv
	}

	// Perform cleanup
	if concreteEnv != nil {
		i.SetEnvironment(concreteEnv)
		i.cleanupInterfaceReferences(i.env)
		i.RestoreEnvironment(savedEnv)
	}
}

// createEnvSyncerCallback creates the environment synchronization callback.
// Syncs i.env with funcEnv so interpreter callbacks use the correct scope.
func (i *Interpreter) createEnvSyncerCallback() evaluator.EnvSyncerFunc {
	return func(funcEnv evaluator.Environment) func() {
		savedEnv := i.env

		// Extract concrete *Environment from adapter
		if adapter, ok := funcEnv.(*evaluator.EnvironmentAdapter); ok {
			if concreteEnv, ok := adapter.Underlying().(*Environment); ok {
				i.SetEnvironment(concreteEnv)
			}
		} else if concreteEnv, ok := interface{}(funcEnv).(*Environment); ok {
			i.SetEnvironment(concreteEnv)
		}

		return func() {
			i.RestoreEnvironment(savedEnv)
		}
	}
}
