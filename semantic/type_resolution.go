package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Type Resolution
// ============================================================================

// resolveType resolves a type name to a Type
// Handles basic types, class types, and enum types
func (a *Analyzer) resolveType(typeName string) (types.Type, error) {
	// Try basic types first
	basicType, err := types.TypeFromString(typeName)
	if err == nil {
		return basicType, nil
	}

	// Try class types
	if classType, found := a.classes[typeName]; found {
		return classType, nil
	}

	// Try enum types (Task 8.43)
	if enumType, found := a.enums[typeName]; found {
		return enumType, nil
	}

	// Try record types (Task 8.68)
	if recordType, found := a.records[typeName]; found {
		return recordType, nil
	}

	return nil, fmt.Errorf("unknown type: %s", typeName)
}

// resolveOperatorType resolves type annotations used in operator declarations.
func (a *Analyzer) resolveOperatorType(typeName string) (types.Type, error) {
	name := strings.TrimSpace(typeName)
	if name == "" {
		return types.VOID, nil
	}

	if t, err := a.resolveType(name); err == nil {
		return t, nil
	}

	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "array of ") {
		elemName := strings.TrimSpace(name[len("array of "):])
		elemType, err := a.resolveOperatorType(elemName)
		if err != nil {
			return nil, err
		}
		return types.NewDynamicArrayType(elemType), nil
	}

	return nil, fmt.Errorf("unknown type: %s", name)
}

// ============================================================================
// Type Hierarchy Helpers
// ============================================================================

// hasCircularInheritance checks if a class has circular inheritance
func (a *Analyzer) hasCircularInheritance(class *types.ClassType) bool {
	seen := make(map[string]bool)
	current := class

	for current != nil {
		if seen[current.Name] {
			return true
		}
		seen[current.Name] = true
		current = current.Parent
	}

	return false
}

// isDescendantOf checks if a class is a descendant of another class
func (a *Analyzer) isDescendantOf(class, ancestor *types.ClassType) bool {
	if class == nil || ancestor == nil {
		return false
	}

	// Walk up the inheritance chain
	current := class.Parent
	for current != nil {
		if current.Name == ancestor.Name {
			return true
		}
		current = current.Parent
	}

	return false
}

// ============================================================================
// Method Resolution Helpers
// ============================================================================

// findMethodInParent searches for a method in the parent class hierarchy
func (a *Analyzer) findMethodInParent(methodName string, parent *types.ClassType) *types.FunctionType {
	if parent == nil {
		return nil
	}

	// Check if method exists in parent
	if methodType, exists := parent.Methods[methodName]; exists {
		return methodType
	}

	// Recursively search in grandparent
	return a.findMethodInParent(methodName, parent.Parent)
}

// isMethodVirtualOrOverride checks if a method is marked virtual or override in class hierarchy
func (a *Analyzer) isMethodVirtualOrOverride(methodName string, classType *types.ClassType) bool {
	if classType == nil {
		return false
	}

	// Check if method exists in this class
	if _, exists := classType.Methods[methodName]; exists {
		// Check if method is virtual or override
		isVirtual := classType.VirtualMethods[methodName]
		isOverride := classType.OverrideMethods[methodName]
		return isVirtual || isOverride
	}

	// Recursively check parent
	return a.isMethodVirtualOrOverride(methodName, classType.Parent)
}

// methodSignaturesMatch compares two function signatures
func (a *Analyzer) methodSignaturesMatch(sig1, sig2 *types.FunctionType) bool {
	// Check parameter count
	if len(sig1.Parameters) != len(sig2.Parameters) {
		return false
	}

	// Check parameter types
	for i := range sig1.Parameters {
		if !sig1.Parameters[i].Equals(sig2.Parameters[i]) {
			return false
		}
	}

	// Check return type
	if !sig1.ReturnType.Equals(sig2.ReturnType) {
		return false
	}

	return true
}

// ============================================================================
// Abstract Method Helpers
// ============================================================================

// getUnimplementedAbstractMethods returns a list of abstract methods from the inheritance chain
// that are not implemented in the given class or its ancestors.
func (a *Analyzer) getUnimplementedAbstractMethods(classType *types.ClassType) []string {
	unimplemented := []string{}

	// Collect all abstract methods from parent chain
	abstractMethods := a.collectAbstractMethods(classType.Parent)

	// Check which ones are not implemented in this class
	for methodName := range abstractMethods {
		// Check if this class implements the method (non-abstract)
		if classType.AbstractMethods[methodName] {
			// Still abstract in this class - not implemented
			unimplemented = append(unimplemented, methodName)
		} else if _, hasMethod := classType.Methods[methodName]; !hasMethod {
			// Method not defined in this class at all - not implemented
			unimplemented = append(unimplemented, methodName)
		}
		// Otherwise, method is implemented (exists and is not abstract)
	}

	return unimplemented
}

// collectAbstractMethods recursively collects all abstract methods from the parent chain
func (a *Analyzer) collectAbstractMethods(parent *types.ClassType) map[string]bool {
	abstractMethods := make(map[string]bool)

	if parent == nil {
		return abstractMethods
	}

	// Add parent's abstract methods
	for methodName, isAbstract := range parent.AbstractMethods {
		if isAbstract {
			abstractMethods[methodName] = true
		}
	}

	// Recursively collect from grandparent
	grandparentAbstract := a.collectAbstractMethods(parent.Parent)
	for methodName := range grandparentAbstract {
		// Only add if not already implemented (non-abstract) in parent
		if !parent.AbstractMethods[methodName] {
			if _, hasMethod := parent.Methods[methodName]; hasMethod {
				// Parent implemented it, don't add to abstract list
				continue
			}
		}
		abstractMethods[methodName] = true
	}

	return abstractMethods
}

// ============================================================================
// Scope Helpers
// ============================================================================

// addParentFieldsToScope recursively adds parent class fields to current scope
func (a *Analyzer) addParentFieldsToScope(parent *types.ClassType) {
	if parent == nil {
		return
	}

	// Add parent's fields
	for fieldName, fieldType := range parent.Fields {
		// Don't override if already defined (shadowing)
		if !a.symbols.IsDeclaredInCurrentScope(fieldName) {
			if visibility, ok := parent.FieldVisibility[fieldName]; ok && visibility == int(ast.VisibilityPrivate) {
				continue
			}
			a.symbols.Define(fieldName, fieldType)
		}
	}

	// Recursively add grandparent fields
	if parent.Parent != nil {
		a.addParentFieldsToScope(parent.Parent)
	}
}

// addParentClassVarsToScope recursively adds parent class variables to current scope (Task 7.62)
func (a *Analyzer) addParentClassVarsToScope(parent *types.ClassType) {
	if parent == nil {
		return
	}

	// Add parent's class variables
	for classVarName, classVarType := range parent.ClassVars {
		// Don't override if already defined (shadowing)
		if !a.symbols.IsDeclaredInCurrentScope(classVarName) {
			a.symbols.Define(classVarName, classVarType)
		}
	}

	// Recursively add grandparent class variables
	if parent.Parent != nil {
		a.addParentClassVarsToScope(parent.Parent)
	}
}

// ============================================================================
// Field/Method Owner Lookups
// ============================================================================

// getFieldOwner returns the class that declares a field, walking up the inheritance chain
func (a *Analyzer) getFieldOwner(class *types.ClassType, fieldName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the field
	if _, found := class.Fields[fieldName]; found {
		return class
	}

	// Check parent classes
	return a.getFieldOwner(class.Parent, fieldName)
}

// getMethodOwner returns the class that declares a method, walking up the inheritance chain
func (a *Analyzer) getMethodOwner(class *types.ClassType, methodName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the method
	if _, found := class.Methods[methodName]; found {
		return class
	}

	// Check parent classes
	return a.getMethodOwner(class.Parent, methodName)
}
