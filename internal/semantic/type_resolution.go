package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
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

	// Handle SetTypeNode directly to validate element type without string round-tripping
	if setNode, ok := typeExpr.(*ast.SetTypeNode); ok {
		return a.resolveSetTypeNode(setNode)
	}

	// Handle ArrayTypeNode directly to avoid string conversion issues
	if arrayNode, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		return a.resolveArrayTypeNode(arrayNode)
	}

	// Handle ClassOfTypeNode directly (metaclass type resolution)
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
	if ident.Normalize(typeName) == "const" {
		return types.VARIANT, nil
	}

	// Check for inline function pointer types first
	// These are synthetic TypeAnnotations created by the parser with full signatures
	// Examples: "function(x: Integer): Integer", "procedure(msg: String)", "function(): Boolean of object"
	if strings.HasPrefix(typeName, "function(") || strings.HasPrefix(typeName, "procedure(") {
		return a.resolveInlineFunctionPointerType(typeName)
	}

	// Check for nested class types in the current class context
	if a.currentNestedTypes != nil {
		if qualified, ok := a.currentNestedTypes[ident.Normalize(typeName)]; ok {
			typeName = qualified
		}
	}
	if a.currentNestedTypes == nil && a.currentClass != nil {
		if aliases, ok := a.nestedTypeAliases[ident.Normalize(a.currentClass.Name)]; ok {
			if qualified, ok := aliases[ident.Normalize(typeName)]; ok {
				typeName = qualified
			}
		}
	}

	// Check for inline array types
	// These are synthetic TypeAnnotations created by the parser with full signatures
	// Examples: "array of Integer", "array[1..10] of String", "array of array of Integer"
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		return a.resolveInlineArrayType(typeName)
	}

	// Check for inline set types: "set of TEnum", "set of Integer", etc.
	lowerName := ident.Normalize(typeName)
	if strings.HasPrefix(lowerName, "set of ") {
		elemName := strings.TrimSpace(typeName[len("set of "):])
		if elemName == "" {
			return nil, fmt.Errorf("unknown type: %s", typeName)
		}
		elementType, err := a.resolveType(elemName)
		if err != nil {
			return nil, fmt.Errorf("unknown element type '%s' in set type", elemName)
		}
		if !types.IsOrdinalType(elementType) {
			return nil, fmt.Errorf("set element type must be ordinal, got %s", elementType.String())
		}

		// Cache inline set types by their normalized signature
		if cached := a.getSetType(lowerName); cached != nil {
			return cached, nil
		}

		setType := types.NewSetType(elementType)
		a.registerType(lowerName, setType)
		return setType, nil
	}

	// Normalize type name for case-insensitive lookup
	normalizedName := ident.Normalize(typeName)

	// Try basic types first (TypeFromString handles case-insensitivity)
	basicType, err := types.TypeFromString(typeName)
	if err == nil {
		return basicType, nil
	}

	// Try user-defined types (classes, interfaces, enums, records, sets, arrays, aliases)
	if userType, found := a.lookupType(typeName); found {
		return userType, nil
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
	var ordinalLow, ordinalHigh int
	var hasOrdinalBounds bool
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
		if len(parts) == 2 {
			// Numeric bounds: array[low..high]
			low := 0
			if _, err := fmt.Sscanf(parts[0], "%d", &low); err != nil {
				return nil, fmt.Errorf("invalid low bound in signature: %s", signature)
			}
			lowBound = &low

			high := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &high); err != nil {
				return nil, fmt.Errorf("invalid high bound in signature: %s", signature)
			}
			highBound = &high
		} else {
			// Ordinal index type: array[TEnum] or array[Boolean]
			indexTypeName := strings.TrimSpace(boundsStr)
			indexType, err := a.resolveType(indexTypeName)
			if err != nil {
				return nil, fmt.Errorf("unknown array index type '%s': %w", indexTypeName, err)
			}

			low, high, ok := types.OrdinalBounds(indexType)
			if !ok {
				return nil, fmt.Errorf("array index type '%s' must be a bounded ordinal type", indexTypeName)
			}

			ordinalLow, ordinalHigh = low, high
			hasOrdinalBounds = true
		}

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
	if hasOrdinalBounds {
		return types.NewStaticArrayType(elementType, ordinalLow, ordinalHigh), nil
	}
	return types.NewDynamicArrayType(elementType), nil
}

// resolveArrayTypeNode resolves an ArrayTypeNode directly from the AST
// (avoids string conversion issues with parentheses and negative bounds).
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

	// Handle enum/ordinal-indexed arrays (extended to all bounded ordinals)
	if arrayNode.IsEnumIndexed() {
		// Resolve the index type
		indexTypeName := getTypeExpressionName(arrayNode.IndexType)
		indexType, err := a.resolveType(indexTypeName)
		if err != nil {
			return nil, fmt.Errorf("unknown array index type '%s': %w", indexTypeName, err)
		}

		// Ensure the index type has finite ordinal bounds (Boolean, Enum, Subrange)
		lowBound, highBound, ok := types.OrdinalBounds(indexType)
		if !ok {
			return nil, fmt.Errorf("array index type '%s' must be a bounded ordinal type, got %s", indexTypeName, indexType.TypeKind())
		}

		return types.NewStaticArrayType(elementType, lowBound, highBound), nil
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

// resolveSetTypeNode resolves a SetTypeNode directly from the AST.
// Validates that the element type is ordinal (enum, subrange, integer/char/boolean).
func (a *Analyzer) resolveSetTypeNode(setNode *ast.SetTypeNode) (types.Type, error) {
	if setNode == nil {
		return nil, fmt.Errorf("nil set type node")
	}

	elementType, err := a.resolveTypeExpression(setNode.ElementType)
	if err != nil {
		return nil, fmt.Errorf("unknown set element type '%s': %w",
			getTypeExpressionName(setNode.ElementType), err)
	}

	if !types.IsOrdinalType(elementType) {
		return nil, fmt.Errorf("set element type must be ordinal, got %s", elementType.String())
	}

	setType := types.NewSetType(elementType)

	// Cache inline set types by their string representation for consistent reuse
	normalizedName := ident.Normalize(setNode.String())
	if !a.hasType(normalizedName) {
		a.registerType(normalizedName, setType)
	}

	return setType, nil
}

// resolveOperatorType resolves type annotations used in operator declarations.
func (a *Analyzer) resolveOperatorType(typeName string) (types.Type, error) {
	name := normalizeOperatorOperandTypeName(typeName)
	if name == "" {
		return types.VOID, nil
	}

	if t, err := a.resolveType(name); err == nil {
		return t, nil
	}

	lower := ident.Normalize(name)
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

// normalizeOperatorOperandTypeName sanitizes operator operand type strings so they can be
// resolved by the regular type resolver. Class operator declarations often include parameter
// names and modifiers (e.g. "const items: array of const"), so we strip those decorations and
// collapse whitespace before attempting to resolve the type name.
func normalizeOperatorOperandTypeName(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if idx := strings.Index(trimmed, ":"); idx != -1 {
		trimmed = trimmed[idx+1:]
	}

	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return ""
	}

	// Collapse repeated whitespace to simplify further checks.
	trimmed = strings.Join(strings.Fields(trimmed), " ")

	lower := ident.Normalize(trimmed)
	modifiers := []string{
		"constref ",
		"const ",
		"var ",
		"out ",
		"reference ",
		"lazy ",
	}

	for _, mod := range modifiers {
		if strings.HasPrefix(lower, mod) && len(trimmed) > len(mod) {
			trimmed = strings.TrimSpace(trimmed[len(mod):])
			lower = ident.Normalize(trimmed)
		}
	}

	return trimmed
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

// findMatchingOverloadInParent finds a method overload in the parent class hierarchy
// that matches the given signature
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

// hasMethodWithName checks if a method with the given name exists in parent hierarchy
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

// findMatchingConstructorInParent finds a constructor overload in the parent class hierarchy
// that matches the given signature
// Note: For constructors, we only compare parameters, not return types, because derived class
// constructors return the derived class type while parent constructors return the parent type
func (a *Analyzer) findMatchingConstructorInParent(constructorName string, signature *types.FunctionType, parent *types.ClassType) *types.MethodInfo {
	if parent == nil {
		return nil
	}

	// Check constructor overloads in parent
	overloads := parent.GetConstructorOverloads(constructorName)
	for _, overload := range overloads {
		// For constructors, only compare parameters (not return type)
		if a.parametersMatch(signature, overload.Signature) {
			return overload
		}
	}

	// Recursively search in grandparent
	return a.findMatchingConstructorInParent(constructorName, signature, parent.Parent)
}

// hasConstructorWithName checks if a constructor with the given name exists in parent hierarchy
// Returns true if ANY overload exists, regardless of signature
func (a *Analyzer) hasConstructorWithName(constructorName string, parent *types.ClassType) bool {
	if parent == nil {
		return false
	}

	// Check if any constructor overloads exist
	if len(parent.GetConstructorOverloads(constructorName)) > 0 {
		return true
	}

	// Recursively check grandparent
	return a.hasConstructorWithName(constructorName, parent.Parent)
}

// getMethodOverloadsInHierarchy collects all method overloads from the class hierarchy
// Returns all overload variants for the given method name, searching up the inheritance chain
// getMethodOverloadsInHierarchy also includes constructor overloads when the method name is a constructor.
func (a *Analyzer) getMethodOverloadsInHierarchy(methodName string, classType *types.ClassType) []*types.MethodInfo {
	if classType == nil {
		return nil
	}

	var result []*types.MethodInfo

	// Check if this is a constructor call (case-insensitive), stored separately in ConstructorOverloads
	var constructorOverloads []*types.MethodInfo
	for ctorName, overloads := range classType.ConstructorOverloads {
		if ident.Equal(ctorName, methodName) {
			constructorOverloads = append(constructorOverloads, overloads...)
		}
	}
	if len(constructorOverloads) > 0 {
		result = append(result, constructorOverloads...)

		// Add implicit parameterless constructor if not already present
		// (DWScript allows calling constructors with no arguments)
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

	// Recursively collect from parent (child methods hide parent methods with same signature)
	// In DWScript, child methods shadow parent methods regardless of override keyword.
	// Only include parent methods with different signatures (true overloads).
	if classType.Parent != nil {
		parentOverloads := a.getMethodOverloadsInHierarchy(methodName, classType.Parent)
		for _, parentOverload := range parentOverloads {
			hidden := false
			for _, currentOverload := range overloads {
				if a.methodSignaturesMatch(currentOverload.Signature, parentOverload.Signature) {
					hidden = true
					break
				}
			}
			if !hidden {
				result = append(result, parentOverload)
			}
		}
	}

	return result
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

// parametersMatch checks if two function signatures have the same parameters (ignoring return type).
// Used to detect ambiguous overloads where parameters match but return types differ.
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
		// Check if this class has its own implementation (methodName already lowercase from AbstractMethods map)
		lowerMethodName := ident.Normalize(methodName)
		hasOwnMethod := len(classType.MethodOverloads[lowerMethodName]) > 0

		if !hasOwnMethod {
			// Method not defined in this class at all - inherited but not implemented
			unimplemented = append(unimplemented, methodName)
		} else {
			// Method is defined in this class
			if isReintroduce, exists := classType.ReintroduceMethods[lowerMethodName]; exists && isReintroduce {
				// Reintroduced method doesn't implement the parent abstract method
				unimplemented = append(unimplemented, methodName)
			} else if isAbstract, exists := classType.AbstractMethods[lowerMethodName]; exists && isAbstract {
				// Still abstract in this class - not implemented
				unimplemented = append(unimplemented, methodName)
			}
		}
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
		// Use case-insensitive lookup since DWScript is case-insensitive
		if _, hasMethod := parent.GetMethod(methodName); hasMethod {
			// Check if parent still marks it as abstract
			lowerMethodName := ident.Normalize(methodName)
			if isAbstract, exists := parent.AbstractMethods[lowerMethodName]; exists && isAbstract {
				// Still abstract in parent
				abstractMethods[methodName] = true
			}
			// Otherwise parent implemented it, don't add to abstract list
		} else {
			// Parent doesn't have this method at all, still abstract
			abstractMethods[methodName] = true
		}
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
			// FieldVisibility uses normalized keys, but Fields uses original case
			normalizedFieldName := ident.Normalize(fieldName)
			visibility, ok := parent.FieldVisibility[normalizedFieldName]
			if ok && visibility == int(ast.VisibilityPrivate) {
				continue
			}
			// Use zero position for synthesized parent field bindings
			a.symbols.Define(fieldName, fieldType, token.Position{})
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
			// Use zero position for synthesized parent class variable bindings
			a.symbols.Define(classVarName, classVarType, token.Position{})
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

	// Check if this class declares the field (case-insensitive)
	// Fields map uses original case keys, so iterate and compare
	for storedName := range class.Fields {
		if ident.Equal(storedName, fieldName) {
			return class
		}
	}

	// Check parent classes
	return a.getFieldOwner(class.Parent, fieldName)
}

// getClassVarOwner returns the class that declares a class variable, walking up the inheritance chain
func (a *Analyzer) getClassVarOwner(class *types.ClassType, classVarName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the class variable (case-insensitive)
	lowerClassVarName := ident.Normalize(classVarName)
	if _, found := class.ClassVars[lowerClassVarName]; found {
		return class
	}

	// Check parent classes
	return a.getClassVarOwner(class.Parent, classVarName)
}

// getMethodOwner returns the class that declares a method, walking up the inheritance chain
func (a *Analyzer) getMethodOwner(class *types.ClassType, methodName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check method overloads (normalized for case-insensitivity)
	methodKey := ident.Normalize(methodName)
	if _, found := class.MethodOverloads[methodKey]; found {
		return class
	}

	// Check parent classes
	return a.getMethodOwner(class.Parent, methodName)
}

// getConstantOwner returns the class that declares a constant, walking up the inheritance chain
func (a *Analyzer) getConstantOwner(class *types.ClassType, constantName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the constant
	if _, found := class.Constants[constantName]; found {
		return class
	}

	// Check parent classes
	return a.getConstantOwner(class.Parent, constantName)
}

// findClassConstantWithVisibility searches for a class constant by name (case-insensitive)
// in the given class and its parent hierarchy, checking visibility permissions.
// Returns the constant's type if found and accessible, nil otherwise.
// If the constant is found but not accessible, an error is added to the analyzer.
func (a *Analyzer) findClassConstantWithVisibility(startClass *types.ClassType, name string, errorPos string) types.Type {
	if startClass == nil {
		return nil
	}

	// Check current class and all parent classes for constants
	for class := startClass; class != nil; class = class.Parent {
		for constName, constType := range class.ConstantTypes {
			if ident.Equal(constName, name) {
				// Check visibility - find which class owns this constant
				constantOwner := a.getConstantOwner(startClass, constName)
				if constantOwner != nil {
					visibility, hasVisibility := constantOwner.ConstantVisibility[constName]
					if hasVisibility && !a.checkVisibility(constantOwner, visibility, constName, "constant") {
						visibilityStr := ast.Visibility(visibility).String()
						a.addError("cannot access %s constant '%s' of class '%s' at %s",
							visibilityStr, constName, constantOwner.Name, errorPos)
						return nil
					}
				}
				// Return the constant's type
				return constType
			}
		}
	}

	return nil
}

// resolveClassOfTypeNode resolves a ClassOfTypeNode to a ClassOfType.
//
// A metaclass type "class of TMyClass" is a type that holds a reference to
// a class type itself, not an instance (metaclass type resolution).
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
