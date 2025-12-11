package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Value to Type Conversion
// ============================================================================

// GetValueType converts a runtime Value to its corresponding types.Type.
// Returns nil for values that don't have a corresponding semantic type (e.g., nil, unassigned).
//
// This is used for array literal type inference where we need to determine
// the element type from the runtime values.
//
// Mapping:
//   - IntegerValue → types.INTEGER
//   - FloatValue → types.FLOAT
//   - StringValue → types.STRING
//   - BooleanValue → types.BOOLEAN
//   - NilValue → nil (context-dependent)
//   - ArrayValue → types.ArrayType (with element type)
//   - RecordValue → types.RecordType
//   - ObjectInstance → types.Class (the object's class)
//   - VariantValue → unwrap to underlying type
//   - EnumValue → types.EnumType
//   - SetValue → types.SetType
//   - NullValue/UnassignedValue → nil
func GetValueType(val Value) types.Type {
	if val == nil {
		return nil
	}

	// Get the type string from the value
	typeStr := val.Type()

	// Handle value types based on their Type() string
	switch typeStr {
	case "INTEGER":
		return types.INTEGER

	case "FLOAT":
		return types.FLOAT

	case "STRING":
		return types.STRING

	case "BOOLEAN":
		return types.BOOLEAN

	case "NIL":
		// Nil is context-dependent - could be any reference type
		return nil

	case "NULL":
		// Variant NULL value - no specific type
		return nil

	case "UNASSIGNED":
		// Variant UNASSIGNED value - no specific type
		return nil

	case "VARIANT":
		// For Variant values, we need to unwrap to get the actual type
		// This requires access to the wrapped value, which we'll handle via a helper
		return unwrapVariantType(val)

	case "ARRAY":
		// Arrays have their element type stored
		// We'll use a helper to extract the element type
		return getArrayElementTypeFromValue(val)

	case "RECORD":
		// Record types - use helper to extract the type
		return getRecordTypeFromValue(val)

	case "ENUM":
		// Enum types - use helper to extract the type
		return getEnumTypeFromValue(val)

	case "SET":
		// Set types - use helper to extract the type
		return getSetTypeFromValue(val)

	default:
		// Check if it's an object (class instance)
		// Object types have the class name as their Type() string
		// For now, we'll just return nil for unknown types
		// In the future, we could look up the class type from the type system
		return nil
	}
}

// unwrapVariantType unwraps a Variant value to get its actual type.
// This uses a type-safe interface check for VariantWrapper.
func unwrapVariantType(val Value) types.Type {
	// Check if the value implements the VariantWrapper interface
	// This interface is defined in runtime package to avoid circular imports
	type VariantWrapper interface {
		UnwrapVariant() Value
	}

	if wrapper, ok := val.(VariantWrapper); ok {
		// Unwrap and recursively get the type
		unwrapped := wrapper.UnwrapVariant()
		return GetValueType(unwrapped)
	}

	// Variant with unknown content - fallback to Variant type
	return types.VARIANT
}

// getArrayElementTypeFromValue extracts the ArrayType from an ArrayValue.
func getArrayElementTypeFromValue(val Value) types.Type {
	if arrVal, ok := val.(*runtime.ArrayValue); ok && arrVal.ArrayType != nil {
		return arrVal.ArrayType
	}
	return nil
}

// getRecordTypeFromValue extracts the RecordType from a RecordValue.
func getRecordTypeFromValue(val Value) types.Type {
	if recVal, ok := val.(*runtime.RecordValue); ok && recVal.RecordType != nil {
		return recVal.RecordType
	}
	return nil
}

// getEnumTypeFromValue extracts type info from an EnumValue.
// Returns nil since EnumValue only stores the type name, not the full EnumType.
// Full type lookup would require TypeSystem access.
func getEnumTypeFromValue(val Value) types.Type {
	// EnumValue only has TypeName string, not *types.EnumType
	// We can't look up the full type without TypeSystem access
	// Return nil - callers should use TypeName for identification
	return nil
}

// getSetTypeFromValue extracts the SetType from a SetValue.
func getSetTypeFromValue(val Value) types.Type {
	if setVal, ok := val.(*runtime.SetValue); ok && setVal.SetType != nil {
		return setVal.SetType
	}
	return nil
}

// getTypeByName converts a type name to a types.Type.
func getTypeByName(name string) types.Type {
	switch name {
	case "Integer":
		return types.INTEGER
	case "Float":
		return types.FLOAT
	case "String":
		return types.STRING
	case "Boolean":
		return types.BOOLEAN
	default:
		// For custom types, return Integer as placeholder.
		// Full type resolution would require TypeSystem access.
		return types.INTEGER
	}
}

// ============================================================================
// Function Pointer Creation Helpers
// ============================================================================

// createFunctionPointerFromDecl creates a FunctionPointerValue from a function declaration.
// This is a simple wrapper that creates a function pointer without type information.
func createFunctionPointerFromDecl(fn *ast.FunctionDecl, closure any) Value {
	return &runtime.FunctionPointerValue{
		Function: fn,
		Closure:  closure,
	}
}

// buildFunctionPointerType builds a FunctionPointerType from a function declaration.
func buildFunctionPointerType(fn *ast.FunctionDecl) *types.FunctionPointerType {
	// Build parameter types from type annotations
	paramTypes := make([]types.Type, len(fn.Parameters))
	for idx, param := range fn.Parameters {
		if param.Type != nil {
			paramTypes[idx] = getTypeByName(param.Type.String())
		} else {
			paramTypes[idx] = types.INTEGER // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if fn.ReturnType != nil {
		returnType = getTypeByName(fn.ReturnType.String())
	}

	// Create the function pointer type
	if returnType != nil {
		return types.NewFunctionPointerType(paramTypes, returnType)
	}
	return types.NewProcedurePointerType(paramTypes)
}

// createFunctionPointerFromDecl creates a FunctionPointerValue from a method declaration.
// Task 3.2.10: Used for creating function pointers from class methods in member access.
// The methodDecl parameter is any to support both ClassMetaValue and ObjectValue callback signatures.
func (e *Evaluator) createFunctionPointerFromDecl(methodDecl any, selfObject Value, ctx *ExecutionContext) Value {
	fn, ok := methodDecl.(*ast.FunctionDecl)
	if !ok {
		return e.newError(nil, "internal error: expected FunctionDecl, got %T", methodDecl)
	}

	pointerType := buildFunctionPointerType(fn)
	return &runtime.FunctionPointerValue{
		Function:    fn,
		Closure:     ctx.Env(),
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}
