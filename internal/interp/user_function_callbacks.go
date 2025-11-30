package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// createUserFunctionCallbacks creates the callback struct for ExecuteUserFunction.
// Task 3.5.144b.12: Consolidates all interpreter-dependent operations for user function execution.
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
	}
}

// createImplicitConversionCallback creates the parameter type conversion callback.
// Task 3.5.144b.2: Callback for BindFunctionParameters.
func (i *Interpreter) createImplicitConversionCallback() evaluator.ImplicitConversionFunc {
	return func(value evaluator.Value, targetTypeName string) (evaluator.Value, bool) {
		return i.tryImplicitConversion(value, targetTypeName)
	}
}

// createDefaultValueGetterCallback creates the return type default value callback.
// Task 3.5.144b.3: Callback for InitializeResultVariable.
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
		recordTypeKey := "__record_type_" + lowerReturnType
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
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
		if interfaceInfo, ok := i.interfaces[lowerReturnType]; ok {
			return &InterfaceInstance{
				Interface: interfaceInfo,
				Object:    nil,
			}
		}

		return defaultValue
	}
}

// createFunctionNameAliasCallback creates the function name alias callback.
// Task 3.5.144b.3: Callback for InitializeResultVariable.
//
// In DWScript, assigning to either Result or the function name sets the return value.
// This creates a ReferenceValue that points to "Result" in the function's environment.
//
// Note: The callback receives the function name but needs access to the function's environment.
// Since the callback is called from within InitializeResultVariable (which has the function context),
// we create a closure that captures the interpreter and creates the reference using i.env.
// The evaluator will call this with the function environment as i.env when the function executes.
func (i *Interpreter) createFunctionNameAliasCallback() evaluator.FunctionNameAliasFunc {
	return func(funcName string) evaluator.Value {
		// Create a ReferenceValue pointing to "Result" in the current environment
		// Note: i.env will be the function's environment when this is called
		// (the interpreter's environment is swapped during function execution)
		return &ReferenceValue{Env: i.env, VarName: "Result"}
	}
}

// createReturnValueConverterCallback creates the return value conversion callback.
// Task 3.5.144b.8: Callback for return value type conversion.
func (i *Interpreter) createReturnValueConverterCallback() evaluator.TryImplicitConversionReturnFunc {
	return func(returnValue evaluator.Value, expectedReturnType string) (evaluator.Value, bool) {
		return i.tryImplicitConversion(returnValue, expectedReturnType)
	}
}

// createInterfaceRefCounterCallback creates the interface ref count increment callback.
// Task 3.5.144b.8: Callback for interface reference counting.
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
// Task 3.5.144b.10: Callback for cleanupInterfaceReferences.
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
// Task 3.5.144b.10: Helper for interface cleanup callback.
func (i *Interpreter) cleanupInterfaceReferencesForEnv(env evaluator.Environment) {
	// We need to iterate through all variables in the environment
	// and release any interface references and object references.
	// However, the evaluator.Environment interface doesn't expose iteration.
	//
	// Solution: Save and temporarily swap i.env, then call the existing method.
	// The env passed from ExecuteUserFunction is the function's environment,
	// and we need to temporarily make it i.env so cleanupInterfaceReferences works.

	// This is safe because:
	// 1. The env is actually an *Environment at runtime (created by NewEnclosedEnvironment)
	// 2. We're just temporarily swapping the reference
	// 3. We restore it immediately after

	savedEnv := i.env

	// Unsafe type assertion - we know env is *Environment at runtime
	// even though the interface signatures don't match perfectly
	type environmentImpl interface {
		Define(string, interface{})
		Get(string) (interface{}, bool)
		Set(string, interface{}) bool
		NewEnclosedEnvironment() evaluator.Environment
	}

	// Use reflection-free approach: just swap i.env temporarily
	// Cast through interface{} to work around type system
	if concreteEnv, ok := interface{}(env).(*Environment); ok {
		i.env = concreteEnv
		i.cleanupInterfaceReferences(i.env)
		i.env = savedEnv
	}
}
