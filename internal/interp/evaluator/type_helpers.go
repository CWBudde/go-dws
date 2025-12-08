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

// getArrayElementTypeFromValue extracts the element type from an ArrayValue.
// Uses a type-safe interface check to avoid circular imports.
func getArrayElementTypeFromValue(val Value) types.Type {
	// Define a local interface that ArrayValue implements
	// This avoids importing the ArrayValue type directly
	type ArrayLike interface {
		// ArrayValue has an ElementType field, but we can't access it directly
		// For now, we'll just return nil and let type inference handle it
		Type() string
	}

	// For now, we can't extract the element type without importing the ArrayValue type
	// For type inference, we'll work with the array values directly
	return nil
}

// getRecordTypeFromValue extracts the RecordType from a RecordValue.
// For now, returns nil to avoid circular imports.
func getRecordTypeFromValue(val Value) types.Type {
	// TODO: Extract record type without circular import
	// For now, return nil - this will be improved in future tasks
	return nil
}

// getEnumTypeFromValue extracts the EnumType from an EnumValue.
// For now, returns nil to avoid circular imports.
func getEnumTypeFromValue(val Value) types.Type {
	// TODO: Extract enum type without circular import
	// For now, return nil - this will be improved in future tasks
	return nil
}

// getSetTypeFromValue extracts the SetType from a SetValue.
// For now, returns nil to avoid circular imports.
func getSetTypeFromValue(val Value) types.Type {
	// TODO: Extract set type without circular import
	// For now, return nil - this will be improved in future tasks
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
