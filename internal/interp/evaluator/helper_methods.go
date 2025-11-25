package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Helper Method Resolution
// ============================================================================
// Task 3.5.98b: Infrastructure for helper method lookup and resolution.
//
// Helper methods are type extensions that add methods to types that don't
// natively have them (e.g., str.ToUpper(), arr.Push(), num.ToString()).
//
// The helper system involves:
// - Looking up helpers for the value's type (getHelpersForValue)
// - Searching helper inheritance chains
// - Handling both AST-declared methods and builtin spec methods
//
// This file provides the infrastructure for finding helper methods directly
// in the evaluator, without delegating to the adapter.

// HelperInfo represents a helper type declaration at runtime.
// This is a temporary interface to avoid circular imports.
// The actual implementation is *interp.HelperInfo.
type HelperInfo interface {
	// GetMethod looks up a method by name in this helper's inheritance chain.
	// Returns the method declaration, the helper that owns it, and whether it was found.
	GetMethod(name string) (*ast.FunctionDecl, HelperInfo, bool)

	// GetBuiltinMethod looks up a builtin method spec by name in this helper's inheritance chain.
	// Returns the builtin spec, the helper that owns it, and whether it was found.
	GetBuiltinMethod(name string) (string, HelperInfo, bool)

	// GetProperty looks up a property by name in this helper's inheritance chain.
	// Returns the property info, the helper that owns it, and whether it was found.
	// Note: Uses any to avoid importing internal/interp/types (circular dependency)
	GetProperty(name string) (any, HelperInfo, bool)
}

// HelperMethodResult represents the result of a helper method lookup.
type HelperMethodResult struct {
	// OwnerHelper is the helper that owns the method (after searching inheritance chain)
	OwnerHelper HelperInfo
	// Method is the AST function declaration (nil for builtin-only methods)
	Method *ast.FunctionDecl
	// BuiltinSpec is the builtin method identifier (empty string for AST-only methods)
	BuiltinSpec string
}

// getHelpersForValue returns all helpers that apply to the given value's type.
// This method maps runtime values to their corresponding helper type names,
// then uses TypeSystem.LookupHelpers() to retrieve the helper registry entries.
//
// Task 3.5.98b: Migrated from Interpreter.getHelpersForValue().
func (e *Evaluator) getHelpersForValue(val Value) []HelperInfo {
	if e.typeSystem == nil {
		return nil
	}

	// Get the type name from the value
	var typeName string
	switch v := val.(type) {
	case ArrayAccessor:
		// For arrays, try multiple helper registries:
		// 1. Specific array type (e.g., "array of String")
		// 2. Dynamic array fallback (e.g., "array of <elem>" for static arrays)
		// 3. Generic "array" helpers
		var combined []HelperInfo

		// Get array type string for specific lookup
		arrayTypeStr := v.ArrayTypeString()
		specific := ident.Normalize(arrayTypeStr)
		if helpers := e.typeSystem.LookupHelpers(specific); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}

		// If it's a static array, also try the dynamic equivalent
		// Note: We'd need to check ArrayAccessor.IsStatic() and ElementType()
		// This is kept simple for now, matching the original logic structure

		// Fallback to generic "array" helpers
		if helpers := e.typeSystem.LookupHelpers("array"); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		return combined

	case EnumAccessor:
		// For enums, try both specific enum type and generic "enum" helpers
		var combined []HelperInfo

		// Get enum type name - we need to check if it has an EnumTypeName method
		// For now, we'll try to use the value's type string
		// TODO: Add EnumTypeName() method to EnumAccessor interface
		enumTypeName := val.Type() // This will return "ENUM" for now

		// Try to get the actual enum type name if the value implements the right interface
		// We'll use a type assertion to access the TypeName field if available
		type enumWithTypeName interface {
			Value
			GetEnumTypeName() string
		}
		if ev, ok := val.(enumWithTypeName); ok {
			enumTypeName = ev.GetEnumTypeName()
		}

		specific := ident.Normalize(enumTypeName)
		if helpers := e.typeSystem.LookupHelpers(specific); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}

		// Fallback to generic "enum" helpers
		if helpers := e.typeSystem.LookupHelpers("enum"); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		return combined

	case ObjectValue:
		// For objects, use the class name
		typeName = v.ClassName()
	case RecordInstanceValue:
		// For records, use the record type name
		typeName = v.GetRecordTypeName()
	case IntegerValue:
		typeName = "Integer"
	case FloatValue:
		typeName = "Float"
	case StringValue:
		typeName = "String"
	case BooleanValue:
		typeName = "Boolean"
	default:
		// For other types, try to extract type name from Type() method
		typeName = v.Type()
	}

	// Look up helpers for this type
	helpers := e.typeSystem.LookupHelpers(typeName)
	if helpers == nil {
		return nil
	}
	return convertToHelperInfoSlice(helpers)
}

// FindHelperMethod searches all applicable helpers for a method with the given name.
// Returns the helper that owns the method, the method declaration (if any),
// and the builtin specification identifier.
//
// The search order is:
// 1. User-defined methods (latest helper overrides earlier ones)
// 2. Builtin-only methods (no AST declaration)
//
// Task 3.5.98b: Migrated from Interpreter.findHelperMethod().
func (e *Evaluator) FindHelperMethod(val Value, methodName string) *HelperMethodResult {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil
	}

	// Search helpers in reverse order so later (user-defined) helpers override earlier ones.
	// For each helper, search the inheritance chain using GetMethod
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]

		// Use GetMethod which searches the inheritance chain and returns the owner helper
		if method, ownerHelper, ok := helper.GetMethod(methodName); ok {
			// Check if there's a builtin spec as well (search from the owner helper)
			if spec, _, ok := ownerHelper.GetBuiltinMethod(methodName); ok {
				return &HelperMethodResult{
					OwnerHelper: ownerHelper,
					Method:      method,
					BuiltinSpec: spec,
				}
			}
			return &HelperMethodResult{
				OwnerHelper: ownerHelper,
				Method:      method,
				BuiltinSpec: "",
			}
		}
	}

	// If no declared method, check for builtin-only entries
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if spec, ownerHelper, ok := helper.GetBuiltinMethod(methodName); ok {
			return &HelperMethodResult{
				OwnerHelper: ownerHelper,
				Method:      nil,
				BuiltinSpec: spec,
			}
		}
	}

	return nil
}

// ============================================================================
// Helper Interfaces for Value Types
// ============================================================================
// These interfaces extend Value types to provide helper-specific information.
// Most of these are already defined in evaluator.go - we reference them here
// for documentation purposes.

// Note: RecordInstanceValue is defined in evaluator.go with GetRecordTypeName()
// Note: EnumAccessor is defined in evaluator.go with GetOrdinal()
// Note: ObjectValue is defined in evaluator.go with ClassName()

// ArrayAccessor is an optional interface for array values.
// Task 3.5.98b: Enables helper method resolution for arrays.
type ArrayAccessor interface {
	Value
	// ArrayTypeString returns the array type as a string (e.g., "array of String").
	ArrayTypeString() string
}

// IntegerValue, FloatValue, StringValue, BooleanValue interfaces
// Task 3.5.98b: These are marker interfaces to enable type switches in getHelpersForValue.
// The actual implementations are in the interp package.

// IntegerValue represents an integer runtime value.
type IntegerValue interface {
	Value
}

// FloatValue represents a float runtime value.
type FloatValue interface {
	Value
}

// StringValue represents a string runtime value.
type StringValue interface {
	Value
}

// BooleanValue represents a boolean runtime value.
type BooleanValue interface {
	Value
}

// ============================================================================
// Helper Utilities
// ============================================================================

// convertToHelperInfoSlice converts a slice of any (from TypeSystem.LookupHelpers)
// to a slice of HelperInfo interfaces.
//
// TypeSystem.LookupHelpers() returns []any to avoid circular imports, but we know
// each element is a *interp.HelperInfo that satisfies our HelperInfo interface.
func convertToHelperInfoSlice(helpers []any) []HelperInfo {
	if helpers == nil {
		return nil
	}

	result := make([]HelperInfo, 0, len(helpers))
	for _, h := range helpers {
		// Each helper should be a *interp.HelperInfo which implements our HelperInfo interface
		if helperInfo, ok := h.(HelperInfo); ok {
			result = append(result, helperInfo)
		}
	}
	return result
}

