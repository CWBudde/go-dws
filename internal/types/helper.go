package types

import (
	"fmt"
	"strings"
)

// HelperRegistry manages the mapping of types to their helpers.
// In DWScript, helpers extend existing types with additional methods and properties.
// Multiple helpers can be defined for the same type, with the most recently defined
// helper taking precedence for method/property lookup.
//
// Key behaviors:
//   - Multiple helpers can extend the same type
//   - Later-defined helpers have higher priority (shadow earlier ones)
//   - Helper methods are looked up after instance methods
//   - Helpers can extend built-in types, records, classes, and other user types
type HelperRegistry struct {
	// Maps type name (case-insensitive) to list of helpers for that type
	// Helpers are stored in definition order - later helpers have higher priority
	helpersByType map[string][]*HelperType

	// Maps helper name (case-insensitive) to helper type
	// Used for helper lookup by name
	helpersByName map[string]*HelperType
}

// NewHelperRegistry creates a new empty helper registry
func NewHelperRegistry() *HelperRegistry {
	return &HelperRegistry{
		helpersByType: make(map[string][]*HelperType),
		helpersByName: make(map[string]*HelperType),
	}
}

// RegisterHelper registers a helper for a type.
// The helper is added to the end of the list for that type, giving it
// higher priority than previously defined helpers.
//
// Parameters:
//   - helper: The helper type to register
//
// Returns:
//   - error if the helper name is already registered
func (hr *HelperRegistry) RegisterHelper(helper *HelperType) error {
	if helper == nil {
		return fmt.Errorf("cannot register nil helper")
	}

	// Normalize helper name to lowercase for case-insensitive lookup
	helperNameLower := strings.ToLower(helper.Name)

	// Check if helper name is already registered
	if _, exists := hr.helpersByName[helperNameLower]; exists {
		return fmt.Errorf("helper '%s' is already defined", helper.Name)
	}

	// Get the target type name
	targetTypeName := helper.TargetType.String()
	targetTypeNameLower := strings.ToLower(targetTypeName)

	// Add helper to name map
	hr.helpersByName[helperNameLower] = helper

	// Add helper to type map (append to end for higher priority)
	hr.helpersByType[targetTypeNameLower] = append(
		hr.helpersByType[targetTypeNameLower],
		helper,
	)

	return nil
}

// GetHelpersForType returns all helpers that extend the given type.
// Helpers are returned in definition order, with later helpers having higher priority.
//
// For method/property lookup, search from the end of the list backwards
// to give priority to more recently defined helpers.
//
// Parameters:
//   - targetType: The type to get helpers for
//
// Returns:
//   - slice of helpers for that type (may be empty if no helpers defined)
func (hr *HelperRegistry) GetHelpersForType(targetType Type) []*HelperType {
	if targetType == nil {
		return nil
	}

	targetTypeName := targetType.String()
	targetTypeNameLower := strings.ToLower(targetTypeName)

	return hr.helpersByType[targetTypeNameLower]
}

// GetHelperByName looks up a helper by its name (case-insensitive).
//
// Parameters:
//   - name: The helper name to look up
//
// Returns:
//   - helper type if found, nil otherwise
//   - boolean indicating if helper was found
func (hr *HelperRegistry) GetHelperByName(name string) (*HelperType, bool) {
	helper, ok := hr.helpersByName[strings.ToLower(name)]
	return helper, ok
}

// FindMethod searches for a method in all helpers for the given type.
// Searches helpers in reverse order (most recent first) to respect priority.
//
// Parameters:
//   - targetType: The type to search helpers for
//   - methodName: The method name to find (case-insensitive)
//
// Returns:
//   - The method's function type if found, nil otherwise
//   - The helper that defines the method (if found)
//   - Boolean indicating if method was found
func (hr *HelperRegistry) FindMethod(targetType Type, methodName string) (*FunctionType, *HelperType, bool) {
	helpers := hr.GetHelpersForType(targetType)
	if len(helpers) == 0 {
		return nil, nil, false
	}

	methodNameLower := strings.ToLower(methodName)

	// Search helpers in reverse order (most recent first)
	for i := len(helpers) - 1; i >= 0; i-- {
		helper := helpers[i]
		if method, ok := helper.GetMethod(methodNameLower); ok {
			return method, helper, true
		}
	}

	return nil, nil, false
}

// FindProperty searches for a property in all helpers for the given type.
// Searches helpers in reverse order (most recent first) to respect priority.
//
// Parameters:
//   - targetType: The type to search helpers for
//   - propertyName: The property name to find (case-insensitive)
//
// Returns:
//   - The property info if found, nil otherwise
//   - The helper that defines the property (if found)
//   - Boolean indicating if property was found
func (hr *HelperRegistry) FindProperty(targetType Type, propertyName string) (*PropertyInfo, *HelperType, bool) {
	helpers := hr.GetHelpersForType(targetType)
	if len(helpers) == 0 {
		return nil, nil, false
	}

	propertyNameLower := strings.ToLower(propertyName)

	// Search helpers in reverse order (most recent first)
	for i := len(helpers) - 1; i >= 0; i-- {
		helper := helpers[i]
		if property, ok := helper.GetProperty(propertyNameLower); ok {
			return property, helper, true
		}
	}

	return nil, nil, false
}

// FindClassVar searches for a class variable in all helpers for the given type.
// Searches helpers in reverse order (most recent first) to respect priority.
//
// Parameters:
//   - targetType: The type to search helpers for
//   - varName: The class variable name to find (case-insensitive)
//
// Returns:
//   - The variable's type if found, nil otherwise
//   - The helper that defines the variable (if found)
//   - Boolean indicating if variable was found
func (hr *HelperRegistry) FindClassVar(targetType Type, varName string) (Type, *HelperType, bool) {
	helpers := hr.GetHelpersForType(targetType)
	if len(helpers) == 0 {
		return nil, nil, false
	}

	varNameLower := strings.ToLower(varName)

	// Search helpers in reverse order (most recent first)
	for i := len(helpers) - 1; i >= 0; i-- {
		helper := helpers[i]
		if varType, ok := helper.GetClassVar(varNameLower); ok {
			return varType, helper, true
		}
	}

	return nil, nil, false
}

// FindClassConst searches for a class constant in all helpers for the given type.
// Searches helpers in reverse order (most recent first) to respect priority.
//
// Parameters:
//   - targetType: The type to search helpers for
//   - constName: The class constant name to find (case-insensitive)
//
// Returns:
//   - The constant's value if found, nil otherwise
//   - The helper that defines the constant (if found)
//   - Boolean indicating if constant was found
func (hr *HelperRegistry) FindClassConst(targetType Type, constName string) (interface{}, *HelperType, bool) {
	helpers := hr.GetHelpersForType(targetType)
	if len(helpers) == 0 {
		return nil, nil, false
	}

	constNameLower := strings.ToLower(constName)

	// Search helpers in reverse order (most recent first)
	for i := len(helpers) - 1; i >= 0; i-- {
		helper := helpers[i]
		if constVal, ok := helper.GetClassConst(constNameLower); ok {
			return constVal, helper, true
		}
	}

	return nil, nil, false
}

// Clear removes all registered helpers from the registry.
// Useful for testing or when starting fresh.
func (hr *HelperRegistry) Clear() {
	hr.helpersByType = make(map[string][]*HelperType)
	hr.helpersByName = make(map[string]*HelperType)
}

// HelperCount returns the total number of registered helpers.
func (hr *HelperRegistry) HelperCount() int {
	return len(hr.helpersByName)
}

// TypeCount returns the number of types that have at least one helper.
func (hr *HelperRegistry) TypeCount() int {
	return len(hr.helpersByType)
}
