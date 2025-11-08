package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Resolution
// ============================================================================

// resolveTypeExpression resolves a TypeExpression directly from the AST to a Type.
// This is preferred over resolveType(getTypeExpressionName()) because it avoids
// string conversion issues with complex expressions like negative array bounds.
func (a *Analyzer) resolveTypeExpression(typeExpr ast.TypeExpression) (types.Type, error) {
	if typeExpr == nil {
		return nil, fmt.Errorf("nil type expression")
	}

	// Handle ArrayTypeNode directly to avoid string conversion issues
	if arrayNode, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		return a.resolveArrayTypeNode(arrayNode)
	}

	// Handle ClassOfTypeNode directly
	// Task 9.71: Support metaclass type resolution
	if classOfNode, ok := typeExpr.(*ast.ClassOfTypeNode); ok {
		return a.resolveClassOfTypeNode(classOfNode)
	}

	// For other type expressions, fall back to string-based resolution
	typeName := getTypeExpressionName(typeExpr)
	return a.resolveType(typeName)
}

// resolveType resolves a type name to a Type
// Handles basic types, class types, enum types, inline function pointer types, and inline array types
// DWScript is case-insensitive, so type names are normalized to lowercase for lookups
func (a *Analyzer) resolveType(typeName string) (types.Type, error) {
	// Handle "const" pseudo-type (used in "array of const")
	// "const" is a special DWScript keyword that means "any type" (Variant)
	if strings.ToLower(typeName) == "const" {
		return types.VARIANT, nil
	}

	// Check for inline function pointer types first
	// These are synthetic TypeAnnotations created by the parser with full signatures
	// Examples: "function(x: Integer): Integer", "procedure(msg: String)", "function(): Boolean of object"
	if strings.HasPrefix(typeName, "function(") || strings.HasPrefix(typeName, "procedure(") {
		return a.resolveInlineFunctionPointerType(typeName)
	}

	// Check for inline array types
	// These are synthetic TypeAnnotations created by the parser with full signatures
	// Examples: "array of Integer", "array[1..10] of String", "array of array of Integer"
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		return a.resolveInlineArrayType(typeName)
	}

	// Normalize type name for case-insensitive lookup
	// DWScript is case-insensitive, so "integer", "Integer", and "INTEGER" should all work
	normalizedName := strings.ToLower(typeName)

	// Try basic types first (TypeFromString now handles case-insensitivity)
	basicType, err := types.TypeFromString(typeName)
	if err == nil {
		return basicType, nil
	}

	// Try class types
	if classType, found := a.classes[normalizedName]; found {
		return classType, nil
	}

	// Try interface types
	if interfaceType, found := a.interfaces[normalizedName]; found {
		return interfaceType, nil
	}

	// Try enum types
	if enumType, found := a.enums[normalizedName]; found {
		return enumType, nil
	}

	// Try record types
	if recordType, found := a.records[normalizedName]; found {
		return recordType, nil
	}

	// Try set types
	if setType, found := a.sets[normalizedName]; found {
		return setType, nil
	}

	// Try array types
	if arrayType, found := a.arrays[normalizedName]; found {
		return arrayType, nil
	}

	// Try type aliases
	if typeAlias, found := a.typeAliases[normalizedName]; found {
		return typeAlias, nil
	}

	// Try subrange types
	if subrangeType, found := a.subranges[normalizedName]; found {
		return subrangeType, nil
	}

	return nil, fmt.Errorf("unknown type: %s", typeName)
}

// resolveInlineFunctionPointerType parses an inline function pointer type signature.
//
// Examples:
//   - "function(x: Integer): Integer" -> FunctionPointerType with 1 param, Integer return
//   - "procedure(msg: String)" -> FunctionPointerType with 1 param, nil return
//   - "function(): Boolean" -> FunctionPointerType with 0 params, Boolean return
//   - "procedure(Sender: TObject) of object" -> MethodPointerType
//
// The signature format is created by the parser in parseTypeExpression():
//   - FunctionPointerTypeNode.String() returns the full signature
//
// This function extracts parameter types and return type by parsing the string representation.
func (a *Analyzer) resolveInlineFunctionPointerType(signature string) (types.Type, error) {
	// Check if this is a method pointer ("of object")
	ofObject := strings.HasSuffix(signature, " of object")
	if ofObject {
		signature = strings.TrimSuffix(signature, " of object")
		signature = strings.TrimSpace(signature)
	}

	// Determine if it's a function or procedure
	isFunction := strings.HasPrefix(signature, "function(")

	// Extract the part between ( and )
	openParen := strings.Index(signature, "(")
	closeParen := strings.LastIndex(signature, ")")
	if openParen == -1 || closeParen == -1 || closeParen < openParen {
		return nil, fmt.Errorf("invalid function pointer signature: %s", signature)
	}

	// Extract parameters string
	paramsStr := signature[openParen+1 : closeParen]

	// Parse parameters
	paramTypes, err := a.parseInlineParameters(paramsStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing parameters in '%s': %w", signature, err)
	}

	// Extract return type (if function)
	var returnType types.Type
	if isFunction {
		// Look for ": ReturnType" after the closing )
		remainder := strings.TrimSpace(signature[closeParen+1:])
		if strings.HasPrefix(remainder, ":") {
			returnTypeName := strings.TrimSpace(remainder[1:])
			if returnTypeName != "" {
				returnType, err = a.resolveType(returnTypeName)
				if err != nil {
					return nil, fmt.Errorf("unknown return type '%s' in function pointer", returnTypeName)
				}
			}
		}
	}

	// Create function pointer type
	if ofObject {
		return types.NewMethodPointerType(paramTypes, returnType), nil
	}
	return types.NewFunctionPointerType(paramTypes, returnType), nil
}

// parseInlineParameters parses the parameter list from an inline function pointer signature.
//
// Format: "param1: Type1; param2, param3: Type2; ..."
// Returns a slice of parameter types in order.
func (a *Analyzer) parseInlineParameters(paramsStr string) ([]types.Type, error) {
	paramsStr = strings.TrimSpace(paramsStr)
	if paramsStr == "" {
		return []types.Type{}, nil
	}

	// Detect format by checking for colon
	// Full syntax: "name: Type" or "name1, name2: Type"
	// Shorthand syntax: "Type" or "Type1, Type2" or "Type1; Type2"
	hasColon := strings.Contains(paramsStr, ":")

	if !hasColon {
		// Shorthand format: just types, no names
		return a.parseShorthandParameters(paramsStr)
	}

	// Full format with names (existing logic)
	paramTypes := []types.Type{}

	// Split by semicolon to get parameter groups
	// Each group is "name1, name2, ...: TypeName"
	groups := strings.Split(paramsStr, ";")

	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		// Split by colon to separate names from type
		parts := strings.Split(group, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter group: %s", group)
		}

		// Get the type name
		typeName := strings.TrimSpace(parts[1])

		// Resolve the type
		paramType, err := a.resolveType(typeName)
		if err != nil {
			return nil, fmt.Errorf("unknown parameter type '%s'", typeName)
		}

		// Count how many parameters have this type (by counting commas + 1)
		namesStr := strings.TrimSpace(parts[0])
		// Remove modifiers (const, var, lazy) from count
		namesStr = strings.TrimPrefix(namesStr, "const ")
		namesStr = strings.TrimPrefix(namesStr, "var ")
		namesStr = strings.TrimPrefix(namesStr, "lazy ")
		namesStr = strings.TrimSpace(namesStr)

		paramCount := strings.Count(namesStr, ",") + 1

		// Add the type for each parameter name
		for i := 0; i < paramCount; i++ {
			paramTypes = append(paramTypes, paramType)
		}
	}

	return paramTypes, nil
}

// parseShorthandParameters parses shorthand parameter syntax (types only, no names).
// Format: "Type1, Type2, ..." or "Type1; Type2; ..."
// Both comma and semicolon are treated as separators.
func (a *Analyzer) parseShorthandParameters(paramsStr string) ([]types.Type, error) {
	paramTypes := []types.Type{}

	// Split by both comma and semicolon
	// Replace semicolons with commas for uniform splitting
	paramsStr = strings.ReplaceAll(paramsStr, ";", ",")

	typeNames := strings.Split(paramsStr, ",")

	for _, typeName := range typeNames {
		typeName = strings.TrimSpace(typeName)
		if typeName == "" {
			continue
		}

		// Remove modifiers if present
		typeName = strings.TrimPrefix(typeName, "const ")
		typeName = strings.TrimPrefix(typeName, "var ")
		typeName = strings.TrimPrefix(typeName, "lazy ")
		typeName = strings.TrimSpace(typeName)

		// Resolve the type
		paramType, err := a.resolveType(typeName)
		if err != nil {
			return nil, fmt.Errorf("unknown parameter type '%s'", typeName)
		}

		paramTypes = append(paramTypes, paramType)
	}

	return paramTypes, nil
}

// resolveInlineArrayType parses an inline array type signature.
//
// Examples:
//   - "array of Integer" -> DynamicArrayType of Integer
//   - "array[1..10] of String" -> StaticArrayType with bounds 1..10
//   - "array of array of Integer" -> Nested dynamic arrays
//   - "array[1..5] of array[1..10] of Integer" -> Nested static arrays
//
// The signature format is created by the parser in ArrayTypeNode.String():
//   - Dynamic: "array of ElementType"
//   - Static: "array[low..high] of ElementType"
func (a *Analyzer) resolveInlineArrayType(signature string) (types.Type, error) {
	var lowBound, highBound *int
	var ofPart string

	// Check if this is a static array with bounds
	if strings.HasPrefix(signature, "array[") {
		// Extract bounds: array[low..high] of Type
		endBracket := strings.Index(signature, "]")
		if endBracket == -1 {
			return nil, fmt.Errorf("invalid array type signature: %s", signature)
		}

		boundsStr := signature[6:endBracket] // Skip "array["
		parts := strings.Split(boundsStr, "..")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid array bounds in signature: %s", signature)
		}

		// Parse low bound
		low := 0
		if _, err := fmt.Sscanf(parts[0], "%d", &low); err != nil {
			return nil, fmt.Errorf("invalid low bound in signature: %s", signature)
		}
		lowBound = &low

		// Parse high bound
		high := 0
		if _, err := fmt.Sscanf(parts[1], "%d", &high); err != nil {
			return nil, fmt.Errorf("invalid high bound in signature: %s", signature)
		}
		highBound = &high

		// Extract the part after ']' which should be " of ElementType"
		ofPart = signature[endBracket+1:]
	} else if strings.HasPrefix(signature, "array of ") {
		// Dynamic array: "array of ElementType"
		// Skip "array" to get " of ElementType"
		ofPart = signature[5:] // Skip "array"
	} else {
		return nil, fmt.Errorf("invalid array type signature: %s", signature)
	}

	// Now ofPart should be " of ElementType"
	if !strings.HasPrefix(ofPart, " of ") {
		return nil, fmt.Errorf("expected ' of ' in array type signature: %s", signature)
	}

	// Extract element type name
	elementTypeName := strings.TrimSpace(ofPart[4:]) // Skip " of "

	// Recursively resolve element type (handles nested arrays)
	elementType, err := a.resolveType(elementTypeName)
	if err != nil {
		return nil, fmt.Errorf("unknown element type '%s' in array type", elementTypeName)
	}

	// Create array type
	if lowBound != nil && highBound != nil {
		return types.NewStaticArrayType(elementType, *lowBound, *highBound), nil
	}
	return types.NewDynamicArrayType(elementType), nil
}

// resolveArrayTypeNode resolves an ArrayTypeNode directly from the AST.
// This avoids string conversion issues with parentheses in bound expressions.
// Task: Fix negative array bounds like array[-5..5]
func (a *Analyzer) resolveArrayTypeNode(arrayNode *ast.ArrayTypeNode) (types.Type, error) {
	if arrayNode == nil {
		return nil, fmt.Errorf("nil array type node")
	}

	// Resolve element type first
	var elementType types.Type
	var err error

	// Check if element type is also an array (nested arrays)
	if nestedArray, ok := arrayNode.ElementType.(*ast.ArrayTypeNode); ok {
		elementType, err = a.resolveArrayTypeNode(nestedArray)
		if err != nil {
			return nil, err
		}
	} else {
		// Get element type name
		elementTypeName := getTypeExpressionName(arrayNode.ElementType)
		elementType, err = a.resolveType(elementTypeName)
		if err != nil {
			return nil, fmt.Errorf("unknown element type '%s': %w", elementTypeName, err)
		}
	}

	// Check if dynamic or static array
	if arrayNode.IsDynamic() {
		return types.NewDynamicArrayType(elementType), nil
	}

	// Static array - evaluate bounds using evaluateConstantInt
	lowBound, err := a.evaluateConstantInt(arrayNode.LowBound)
	if err != nil {
		return nil, fmt.Errorf("array lower bound must be a compile-time constant: %w", err)
	}

	highBound, err := a.evaluateConstantInt(arrayNode.HighBound)
	if err != nil {
		return nil, fmt.Errorf("array upper bound must be a compile-time constant: %w", err)
	}

	// Validate bounds
	if lowBound > highBound {
		return nil, fmt.Errorf("array lower bound (%d) cannot be greater than upper bound (%d)", lowBound, highBound)
	}

	return types.NewStaticArrayType(elementType, lowBound, highBound), nil
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

// findMatchingOverloadInParent finds a method overload in the parent class hierarchy
// that matches the given signature (Task 9.61)
func (a *Analyzer) findMatchingOverloadInParent(methodName string, signature *types.FunctionType, parent *types.ClassType) *types.MethodInfo {
	if parent == nil {
		return nil
	}

	// Check overloads in parent
	overloads := parent.GetMethodOverloads(methodName)
	for _, overload := range overloads {
		if a.methodSignaturesMatch(signature, overload.Signature) {
			return overload
		}
	}

	// Recursively search in grandparent
	return a.findMatchingOverloadInParent(methodName, signature, parent.Parent)
}

// hasMethodWithName checks if a method with the given name exists in parent hierarchy (Task 9.61)
// Returns true if ANY overload exists, regardless of signature
func (a *Analyzer) hasMethodWithName(methodName string, parent *types.ClassType) bool {
	if parent == nil {
		return false
	}

	// Check if any overloads exist
	if len(parent.GetMethodOverloads(methodName)) > 0 {
		return true
	}

	// Recursively check grandparent
	return a.hasMethodWithName(methodName, parent.Parent)
}

// getMethodOverloadsInHierarchy collects all method overloads from the class hierarchy (Task 9.61)
// Returns all overload variants for the given method name, searching up the inheritance chain
// Task 9.68: Also includes constructor overloads when the method name is a constructor
func (a *Analyzer) getMethodOverloadsInHierarchy(methodName string, classType *types.ClassType) []*types.MethodInfo {
	if classType == nil {
		return nil
	}

	var result []*types.MethodInfo

	// Task 9.68: Check if this is a constructor call (class static method call)
	// Constructors are stored separately in ConstructorOverloads
	// Task 9.19: Perform case-insensitive constructor lookup
	var constructorOverloads []*types.MethodInfo
	for ctorName, overloads := range classType.ConstructorOverloads {
		if strings.EqualFold(ctorName, methodName) {
			constructorOverloads = append(constructorOverloads, overloads...)
		}
	}
	if len(constructorOverloads) > 0 {
		// This is a constructor - include constructor overloads
		result = append(result, constructorOverloads...)

		// Task 9.68: Add implicit parameterless constructor if not already present
		// DWScript allows calling constructors with no arguments even if only
		// parameterized constructors are declared
		hasParameterlessConstructor := false
		for _, ctor := range constructorOverloads {
			if len(ctor.Signature.Parameters) == 0 {
				hasParameterlessConstructor = true
				break
			}
		}
		if !hasParameterlessConstructor {
			// Add implicit parameterless constructor that returns the class type
			implicitConstructor := &types.MethodInfo{
				Signature: types.NewFunctionType([]types.Type{}, classType),
			}
			result = append(result, implicitConstructor)
		}
		return result
	}

	// Collect overloads from current class (regular methods)
	overloads := classType.GetMethodOverloads(methodName)
	result = append(result, overloads...)

	// Recursively collect from parent (only if not hidden/overridden in current class)
	// Task 9.21.6: Fix overload resolution - child methods hide parent methods with same signature
	// In DWScript, when a child class defines a method with the same signature as a parent method,
	// it hides/shadows the parent method (regardless of override keyword).
	// We only want to include parent methods with DIFFERENT signatures (true overloads).
	if classType.Parent != nil {
		parentOverloads := a.getMethodOverloadsInHierarchy(methodName, classType.Parent)
		for _, parentOverload := range parentOverloads {
			// Check if this parent method is hidden by a child method with the same signature
			hidden := false
			for _, currentOverload := range overloads {
				// A child method hides a parent method if they have the same signature
				// (regardless of whether override keyword is used)
				if a.methodSignaturesMatch(currentOverload.Signature, parentOverload.Signature) {
					hidden = true
					break
				}
			}
			// Only add parent method if it has a different signature (true overload)
			if !hidden {
				result = append(result, parentOverload)
			}
		}
	}

	return result
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

// parametersMatch checks if two function signatures have the same parameters (ignoring return type)
// Task 9.62: Used to detect ambiguous overloads where parameters match but return types differ
func (a *Analyzer) parametersMatch(sig1, sig2 *types.FunctionType) bool {
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

// addParentClassVarsToScope recursively adds parent class variables to current scope
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

// resolveClassOfTypeNode resolves a ClassOfTypeNode to a ClassOfType.
//
// A metaclass type "class of TMyClass" is a type that holds a reference to
// a class type itself, not an instance.
//
// Task 9.71: Metaclass type resolution
func (a *Analyzer) resolveClassOfTypeNode(classOfNode *ast.ClassOfTypeNode) (types.Type, error) {
	if classOfNode == nil {
		return nil, fmt.Errorf("nil class of type node")
	}

	// Resolve the underlying class type
	classTypeName := getTypeExpressionName(classOfNode.ClassType)
	classType, err := a.resolveType(classTypeName)
	if err != nil {
		return nil, fmt.Errorf("unknown class type '%s' in metaclass declaration: %w", classTypeName, err)
	}

	// Verify that it's actually a class type
	concreteClassType, ok := classType.(*types.ClassType)
	if !ok {
		return nil, fmt.Errorf("'%s' is not a class type, cannot create metaclass 'class of %s'", classTypeName, classTypeName)
	}

	// Create and return the metaclass type
	return &types.ClassOfType{
		ClassType: concreteClassType,
	}, nil
}
