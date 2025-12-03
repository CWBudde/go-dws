package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// createUserFunctionCallbacks creates the callback struct for ExecuteUserFunction.
//
// This method creates callbacks that allow the evaluator's ExecuteUserFunction to:
//  1. Apply implicit type conversion to parameters
//  2. Get default values for return types (records, arrays, interfaces)
//  3. Create function name aliases (ReferenceValue pointing to Result)
//  4. Apply implicit conversion to return values
//  5. Increment interface reference counts when returning
//  6. Clean up interface/object references when function exits
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
//
// This handles complex default value initialization for:
//   - Basic types: Integer→0, Float→0.0, String→"", Boolean→false
//   - Record types: Creates empty record with nested initialization
//   - Array types: Creates empty array
//   - Interface types: Creates InterfaceInstance with nil object
//
// Note: This callback receives the return type as a string (from fn.ReturnType.String()).
// We need to look up the type and create appropriate default values.
func (i *Interpreter) createDefaultValueGetterCallback() evaluator.DefaultValueFunc {
	return func(returnTypeName string) evaluator.Value {
		// Get basic default value based on type name
		lowerReturnType := ident.Normalize(returnTypeName)
		var defaultValue evaluator.Value

		// Basic type defaults (matches getDefaultValue behavior)
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

		// Check if return type is a record (overrides default)
		lowerReturnType = ident.Normalize(returnTypeName)
		if recordTypeValueAny := i.typeSystem.LookupRecord(returnTypeName); recordTypeValueAny != nil {
			// TypeSystem.LookupRecord returns any to avoid import cycles
			// We know it's actually *RecordTypeValue at runtime
			if rtv, ok := recordTypeValueAny.(*RecordTypeValue); ok {
				// Use createRecordValue for proper nested record initialization
				return i.createRecordValue(rtv.RecordType)
			}
		}

		// Check if return type is an array (overrides default)
		// Array return types should be initialized to empty arrays, not NIL
		if arrayType := i.typeSystem.LookupArrayType(returnTypeName); arrayType != nil {
			return NewArrayValue(arrayType)
		} else if strings.HasPrefix(lowerReturnType, "array") {
			// Handle inline array return types (dynamic or static)
			if arrayType := i.parseInlineArrayType(lowerReturnType); arrayType != nil {
				return NewArrayValue(arrayType)
			}
		}

		// Check if return type is an interface (overrides default)
		// Interface return types should be initialized to InterfaceInstance with nil object
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
//
// In DWScript, assigning to either Result or the function name sets the return value.
// This creates a ReferenceValue that points to "Result" in the function's environment.
func (i *Interpreter) createFunctionNameAliasCallback() evaluator.FunctionNameAliasFunc {
	return func(funcName string, funcEnv evaluator.Environment) evaluator.Value {
		// Convert evaluator.Environment to *Environment for ReferenceValue
		// The funcEnv is actually an *Environment wrapped in EnvironmentAdapter
		if adapter, ok := funcEnv.(*evaluator.EnvironmentAdapter); ok {
			if env, ok := adapter.Underlying().(*Environment); ok {
				return &ReferenceValue{Env: env, VarName: "Result"}
			}
		}
		// If we can't get the environment, return nil (will cause issues but provides debug info)
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
//
// When returning an interface value from a function, the ref count needs to be incremented
// for the caller's reference. This will be balanced by cleanup releasing Result after return.
func (i *Interpreter) createInterfaceRefCounterCallback() evaluator.IncrementInterfaceRefCountFunc {
	return func(returnValue evaluator.Value) {
		// If returning an interface, increment RefCount for the caller's reference
		if intfInst, isIntf := returnValue.(*InterfaceInstance); isIntf {
			if intfInst.Object != nil {
				intfInst.Object.RefCount++
			}
		}
	}
}

// createInterfaceCleanupCallback creates the interface/object cleanup callback.
//
// This cleans up interface/object references when the function scope ends,
// decrementing ref counts and calling destructors as needed.
//
// Note: The evaluator passes the function environment, but we need to call
// cleanupInterfaceReferences which expects *Environment. We save/restore i.env
// to make it work with the existing cleanupInterfaceReferences implementation.
func (i *Interpreter) createInterfaceCleanupCallback() evaluator.CleanupInterfaceReferencesFunc {
	return func(env evaluator.Environment) {
		// Save current interpreter environment
		savedEnv := i.env

		// Temporarily set i.env to the function environment for cleanup
		// The environment passed is actually an *Environment at runtime
		// We use a type assertion with a direct field access pattern
		type envImpl interface {
			evaluator.Environment
			getStore() interface{} // We'll access the environment directly
		}

		// Since we can't type-assert directly due to interface mismatch,
		// we'll manually iterate through the environment's variables
		// by using the Get/Set methods available on the interface.
		// This is a limitation we'll work around by saving i.env and using
		// the cleanupInterfaceReferences method which accesses i.env directly.

		// Actually, let's use a simpler approach: access the environment store directly
		// by calling a helper method that does the cleanup with the environment.
		// For now, we'll just call cleanupInterfaceReferences with i.env
		// since the function environment should already be set when this is called.

		// The ExecuteUserFunction creates a funcEnv and passes it to this callback.
		// We need to clean up that environment, but cleanupInterfaceReferences
		// expects *Environment. Let's create a new helper that works with the interface.
		i.cleanupInterfaceReferencesForEnv(env)

		// Restore original environment
		i.env = savedEnv
	}
}

// cleanupInterfaceReferencesForEnv is a helper that cleans up interface references
// using the evaluator.Environment interface instead of *Environment.
func (i *Interpreter) cleanupInterfaceReferencesForEnv(env evaluator.Environment) {
	// We need to iterate through all variables in the environment
	// and release any interface references and object references.
	// However, the evaluator.Environment interface doesn't expose iteration.
	//
	// Solution: Extract the concrete *Environment and call the existing cleanup method.
	// The env passed from ExecuteUserFunction is either:
	// 1. An *EnvironmentAdapter wrapping *Environment (via NewEnclosedEnvironment)
	// 2. Directly an *Environment (shouldn't happen but handle it)

	savedEnv := i.env

	// Try to extract the concrete *Environment
	var concreteEnv *Environment

	// First, check if it's an *EnvironmentAdapter
	if adapter, ok := env.(*evaluator.EnvironmentAdapter); ok {
		// Extract the underlying environment
		if underlying, ok := adapter.Underlying().(*Environment); ok {
			concreteEnv = underlying
		}
	} else if directEnv, ok := interface{}(env).(*Environment); ok {
		// Direct *Environment (fallback)
		concreteEnv = directEnv
	}

	// Perform cleanup if we got a concrete environment
	if concreteEnv != nil {
		i.env = concreteEnv
		i.cleanupInterfaceReferences(i.env)
		i.env = savedEnv
	}
}

// createEnvSyncerCallback creates the environment synchronization callback.
//
// Why this is needed:
//   - ExecuteUserFunction creates a new funcEnv for the function scope
//   - The function body is executed with funcCtx.Env() = funcEnv
//   - However, when EvalNode is called back to the interpreter (e.g., for
//     function pointer assignments like "Result := @obj.Method"), it uses i.env
//   - Without syncing, i.env points to the caller's environment, not funcEnv
//   - This causes function pointer assignments to Result to go to the wrong env
//
// The callback:
// 1. Saves the current i.env
// 2. Sets i.env to the concrete *Environment from funcEnv
// 3. Returns a restore function that resets i.env to its original value
func (i *Interpreter) createEnvSyncerCallback() evaluator.EnvSyncerFunc {
	return func(funcEnv evaluator.Environment) func() {
		// Save current interpreter environment
		savedEnv := i.env

		// Extract concrete *Environment from funcEnv
		// The funcEnv is actually an *EnvironmentAdapter wrapping *Environment
		synced := false
		if adapter, ok := funcEnv.(*evaluator.EnvironmentAdapter); ok {
			if concreteEnv, ok := adapter.Underlying().(*Environment); ok {
				i.env = concreteEnv
				synced = true
			}
		} else if concreteEnv, ok := interface{}(funcEnv).(*Environment); ok {
			// Direct *Environment (shouldn't happen but handle it)
			i.env = concreteEnv
			synced = true
		}

		if !synced {
			// Couldn't sync - this is a problem
			// For debugging, print type info
			_ = funcEnv // reference to prevent unused warning
		}

		// Return restore function
		return func() {
			i.env = savedEnv
		}
	}
}
