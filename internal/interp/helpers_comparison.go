package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Helper Lookup and Comparison Functions
// ============================================================================

// findPropertyCaseInsensitive searches for a property by name using case-insensitive comparison.
func findPropertyCaseInsensitive(props map[string]*types.PropertyInfo, name string) *types.PropertyInfo {
	for key, prop := range props {
		if strings.EqualFold(key, name) {
			return prop
		}
	}
	return nil
}

// findMethodCaseInsensitive searches for a method by name using case-insensitive comparison.
func findMethodCaseInsensitive(methods map[string]*ast.FunctionDecl, name string) *ast.FunctionDecl {
	for key, method := range methods {
		if strings.EqualFold(key, name) {
			return method
		}
	}
	return nil
}

// findBuiltinMethodCaseInsensitive searches for a builtin method spec by name using case-insensitive comparison.
func findBuiltinMethodCaseInsensitive(builtinMethods map[string]string, name string) (string, bool) {
	for key, spec := range builtinMethods {
		if strings.EqualFold(key, name) {
			return spec, true
		}
	}
	return "", false
}

// ============================================================================
// HelperInfo Lookup Methods
// ============================================================================

// GetMethod looks up a method by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the method, the helper that owns it, and whether it was found.
// Task 9.1: Helper inheritance support
func (h *HelperInfo) GetMethod(name string) (*ast.FunctionDecl, *HelperInfo, bool) {
	// Look in this helper first (case-insensitive)
	if method := findMethodCaseInsensitive(h.Methods, name); method != nil {
		return method, h, true
	}

	// If not found and we have a parent, look there
	if h.ParentHelper != nil {
		return h.ParentHelper.GetMethod(name)
	}

	return nil, nil, false
}

// GetBuiltinMethod looks up a builtin method spec by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the builtin spec, the helper that owns it, and whether it was found.
// Task 9.1: Helper inheritance support
func (h *HelperInfo) GetBuiltinMethod(name string) (string, *HelperInfo, bool) {
	// Look in this helper first (case-insensitive)
	if spec, ok := findBuiltinMethodCaseInsensitive(h.BuiltinMethods, name); ok {
		return spec, h, true
	}

	// If not found and we have a parent, look there
	if h.ParentHelper != nil {
		return h.ParentHelper.GetBuiltinMethod(name)
	}

	return "", nil, false
}

// GetProperty looks up a property by name in this helper.
// If not found in this helper, searches in parent helper (if any).
// Returns the property, the helper that owns it, and whether it was found.
// Task 9.1: Helper inheritance support
func (h *HelperInfo) GetProperty(name string) (*types.PropertyInfo, *HelperInfo, bool) {
	// Look in this helper first (case-insensitive)
	if prop := findPropertyCaseInsensitive(h.Properties, name); prop != nil {
		return prop, h, true
	}

	// If not found and we have a parent, look there
	if h.ParentHelper != nil {
		return h.ParentHelper.GetProperty(name)
	}

	return nil, nil, false
}

// ============================================================================
// Helper Discovery and Lookup
// ============================================================================

// getHelpersForValue returns all helpers that apply to the given value's type
func (i *Interpreter) getHelpersForValue(val Value) []*HelperInfo {
	if i.helpers == nil {
		return nil
	}

	// Get the type name from the value
	var typeName string
	switch v := val.(type) {
	case *IntegerValue:
		typeName = "Integer"
	case *FloatValue:
		typeName = "Float"
	case *StringValue:
		typeName = "String"
	case *BooleanValue:
		typeName = "Boolean"
	case *ObjectInstance:
		typeName = v.Class.Name
	case *RecordValue:
		typeName = v.RecordType.Name
	case *ArrayValue:
		// First try specific array type (e.g., "array of String"), then generic ARRAY helpers
		specific := v.ArrayType.String()
		var combined []*HelperInfo
		if h, ok := i.helpers[specific]; ok {
			combined = append(combined, h...)
		}
		if h, ok := i.helpers["ARRAY"]; ok {
			combined = append(combined, h...)
		}
		return combined
	case *EnumValue:
		// First try specific enum type (e.g., "TColor"), then generic ENUM helpers
		specific := v.TypeName
		var combined []*HelperInfo
		if h, ok := i.helpers[specific]; ok {
			combined = append(combined, h...)
		}
		if h, ok := i.helpers["ENUM"]; ok {
			combined = append(combined, h...)
		}
		return combined
	default:
		// For other types, try to extract type name from Type() method
		typeName = v.Type()
	}

	// Look up helpers for this type
	return i.helpers[typeName]
}

// findHelperMethod searches all applicable helpers for a method with the given name
// and returns the helper that owns the method, method declaration (if any), and builtin specification identifier.
// Task 9.1: Updated to support helper inheritance
func (i *Interpreter) findHelperMethod(val Value, methodName string) (*HelperInfo, *ast.FunctionDecl, string) {
	helpers := i.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil, ""
	}

	// Search helpers in reverse order so later (user-defined) helpers override earlier ones.
	// For each helper, search the inheritance chain using GetMethod
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]

		// Use GetMethod which searches the inheritance chain and returns the owner helper
		if method, ownerHelper, ok := helper.GetMethod(methodName); ok {
			// Check if there's a builtin spec as well (search from the owner helper)
			if spec, _, ok := ownerHelper.GetBuiltinMethod(methodName); ok {
				return ownerHelper, method, spec
			}
			return ownerHelper, method, ""
		}
	}

	// If no declared method, check for builtin-only entries
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if spec, ownerHelper, ok := helper.GetBuiltinMethod(methodName); ok {
			return ownerHelper, nil, spec
		}
	}

	return nil, nil, ""
}

// findHelperProperty searches all applicable helpers for a property with the given name
// and returns the helper that owns the property and the property info.
// Task 9.1: Updated to support helper inheritance
func (i *Interpreter) findHelperProperty(val Value, propName string) (*HelperInfo, *types.PropertyInfo) {
	helpers := i.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil
	}

	// Search helpers in reverse order so later helpers override earlier ones
	// For each helper, search the inheritance chain using GetProperty
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if prop, ownerHelper, ok := helper.GetProperty(propName); ok {
			return ownerHelper, prop
		}
	}

	return nil, nil
}

// isBuiltinMethodParameterless returns true if the builtin method spec requires no parameters.
// This is used for auto-invoke logic in member access expressions.
func (i *Interpreter) isBuiltinMethodParameterless(builtinSpec string) bool {
	// Map of builtin method specs to their parameter counts
	// This must be kept in sync with the actual builtin method implementations
	parameterlessBuiltins := map[string]bool{
		"__array_pop":           true,  // Pop() - no parameters
		"__array_push":          false, // Push(value) - 1 parameter
		"__array_swap":          false, // Swap(i, j) - 2 parameters
		"__array_add":           false, // Add(value) - 1 parameter
		"__array_delete":        false, // Delete(index) - 1 parameter
		"__array_indexof":       false, // IndexOf(value) - 1 parameter
		"__array_setlength":     false, // SetLength(n) - 1 parameter
		"__integer_tostring":    true,  // ToString() - no parameters
		"__integer_tohexstring": true,  // ToHexString() - no parameters
		"__float_tostring_prec": false, // ToString(precision) - 1 parameter
		"__boolean_tostring":    true,  // ToString() - no parameters
		"__string_toupper":      true,  // ToUpper() - no parameters
		"__string_tolower":      true,  // ToLower() - no parameters
		"__string_array_join":   false, // Join(separator) - 1 parameter
	}

	if isParameterless, exists := parameterlessBuiltins[builtinSpec]; exists {
		return isParameterless
	}

	// For any builtin method not in our map, assume it has parameters (safer default)
	// This prevents incorrect auto-invocation
	return false
}
