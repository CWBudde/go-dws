package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Class Analysis (Tasks 7.54-7.59)
// ============================================================================

// analyzeClassDecl analyzes a class declaration
func (a *Analyzer) analyzeClassDecl(decl *ast.ClassDecl) {
	className := decl.Name.Value

	// Check if class is already declared
	// Task 9.285: Use lowercase for case-insensitive lookup
	if _, exists := a.classes[strings.ToLower(className)]; exists {
		a.addError("class '%s' already declared at %s", className, decl.Token.Pos.String())
		return
	}

	// Resolve parent class if specified
	var parentClass *types.ClassType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		var found bool
		// Task 9.285: Use lowercase for case-insensitive lookup
		parentClass, found = a.classes[strings.ToLower(parentName)]
		if !found {
			a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
			return
		}
	} else {
		// Task 9.51: If no explicit parent, implicitly inherit from TObject (unless this IS TObject or external)
		// External classes can have nil parent (inherit from Object)
		if !strings.EqualFold(className, "TObject") && !decl.IsExternal {
			parentClass = a.classes["tobject"]
			if parentClass == nil {
				a.addError("implicit parent class 'TObject' not found at %s", decl.Token.Pos.String())
				return
			}
		}
	}

	// Create new class type
	classType := types.NewClassType(className, parentClass)

	// Set abstract flag
	classType.IsAbstract = decl.IsAbstract

	// Set external flags
	classType.IsExternal = decl.IsExternal
	classType.ExternalName = decl.ExternalName

	// Validate external class inheritance
	if decl.IsExternal {
		// External class must inherit from nil (Object) or another external class
		if parentClass != nil && !parentClass.IsExternal {
			a.addError("external class '%s' cannot inherit from non-external class '%s' at %s",
				className, parentClass.Name, decl.Token.Pos.String())
			return
		}
	} else {
		// Non-external class cannot inherit from external class
		if parentClass != nil && parentClass.IsExternal {
			a.addError("non-external class '%s' cannot inherit from external class '%s' at %s",
				className, parentClass.Name, decl.Token.Pos.String())
			return
		}
	}

	// Check for circular inheritance
	if parentClass != nil && a.hasCircularInheritance(classType) {
		a.addError("circular inheritance detected in class '%s' at %s", className, decl.Token.Pos.String())
		return
	}

	// Analyze and add fields
	fieldNames := make(map[string]bool)
	classVarNames := make(map[string]bool)
	for _, field := range decl.Fields {
		fieldName := field.Name.Value

		// Check if this is a class variable (static field)
		if field.IsClassVar {
			// Check for duplicate class variable names
			if classVarNames[fieldName] {
				a.addError("duplicate class variable '%s' in class '%s' at %s",
					fieldName, className, field.Token.Pos.String())
				continue
			}
			classVarNames[fieldName] = true

			// Verify class variable type exists
			if field.Type == nil {
				a.addError("class variable '%s' missing type annotation in class '%s'",
					fieldName, className)
				continue
			}

			typeName := getTypeExpressionName(field.Type)
			fieldType, err := a.resolveType(typeName)
			if err != nil {
				a.addError("unknown type '%s' for class variable '%s' in class '%s' at %s",
					typeName, fieldName, className, field.Token.Pos.String())
				continue
			}

			// Store class variable type in ClassType
			classType.ClassVars[fieldName] = fieldType
		} else {
			// Instance field
			// Check for duplicate field names
			if fieldNames[fieldName] {
				a.addError("duplicate field '%s' in class '%s' at %s",
					fieldName, className, field.Token.Pos.String())
				continue
			}
			fieldNames[fieldName] = true

			// Verify field type exists
			if field.Type == nil {
				a.addError("field '%s' missing type annotation in class '%s'",
					fieldName, className)
				continue
			}

			typeName := getTypeExpressionName(field.Type)
			fieldType, err := a.resolveType(typeName)
			if err != nil {
				a.addError("unknown type '%s' for field '%s' in class '%s' at %s",
					typeName, fieldName, className, field.Token.Pos.String())
				continue
			}

			// Add instance field to class
			classType.Fields[fieldName] = fieldType

			// Store field visibility
			classType.FieldVisibility[fieldName] = int(field.Visibility)
		}
	}

	// Analyze and add constants
	constantNames := make(map[string]bool)
	for _, constant := range decl.Constants {
		constantName := constant.Name.Value

		// Check for duplicate constant names
		if constantNames[constantName] {
			a.addError("duplicate constant '%s' in class '%s' at %s",
				constantName, className, constant.Token.Pos.String())
			continue
		}
		constantNames[constantName] = true

		// Validate constant value is a compile-time constant expression
		// For now, we accept any expression and will evaluate it later
		// In a full implementation, we'd check if the expression is constant
		// (literals, other constants, or constant expressions)

		// Store constant visibility
		classType.ConstantVisibility[constantName] = int(constant.Visibility)

		// Note: We don't evaluate the constant value here during semantic analysis.
		// The constant values will be evaluated at runtime when accessed.
		// We just mark that this constant exists.
		classType.Constants[constantName] = nil // Placeholder; actual value evaluated at runtime
	}

	// Register class before analyzing methods (so methods can reference the class)
	// Task 9.285: Use lowercase for case-insensitive lookup
	a.classes[strings.ToLower(className)] = classType

	// Analyze methods
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	for _, method := range decl.Methods {
		a.analyzeMethodDecl(method, classType)
	}

	// Analyze constructor if present
	if decl.Constructor != nil {
		a.analyzeMethodDecl(decl.Constructor, classType)
	}

	// Task 9.1 & 9.2: Constructor inheritance and implicit default constructor
	// If child class has no constructors:
	// 1. Check if parent has constructors (Task 9.2)
	// 2. If yes, inherit accessible parent constructors
	// 3. If no, generate implicit default constructor (Task 9.1)
	if len(classType.Constructors) == 0 && len(classType.ConstructorOverloads) == 0 {
		if parentClass != nil && len(parentClass.Constructors) > 0 {
			// Task 9.2: Inherit parent constructors
			a.inheritParentConstructors(classType, parentClass)
		} else {
			// Task 9.1: Generate implicit default constructor
			a.synthesizeDefaultConstructor(classType)
		}
	}

	// Analyze properties
	// Properties are analyzed after methods so they can reference both fields and methods
	for _, property := range decl.Properties {
		a.analyzePropertyDecl(property, classType)
	}

	// Register class operators (Stage 8)
	a.registerClassOperators(classType, decl)

	// Check method overriding
	if parentClass != nil {
		a.checkMethodOverriding(classType, parentClass)
	}

	// Validate interface implementation
	if len(decl.Interfaces) > 0 {
		a.validateInterfaceImplementation(classType, decl)
	}

	// Validate abstract class rules
	a.validateAbstractClass(classType, decl)
}

// analyzeMethodImplementation analyzes a method implementation outside a class
// This handles code like: function TExample.GetValue: Integer; begin ... end;
func (a *Analyzer) analyzeMethodImplementation(decl *ast.FunctionDecl) {
	className := decl.ClassName.Value

	// Look up the class
	// Task 9.285: Use lowercase for case-insensitive lookup
	classType, exists := a.classes[strings.ToLower(className)]
	if !exists {
		a.addError("unknown type '%s' at %s", className, decl.Token.Pos.String())
		return
	}

	// Task 9.281: Look up the method in the class to ensure it was declared
	// Task 9.19: Handle overloaded methods and constructors
	methodName := decl.Name.Value
	var declaredMethod *types.FunctionType
	var methodExists bool

	if decl.IsConstructor {
		// For constructors, check all overloads to find matching signature
		overloads := classType.GetConstructorOverloads(methodName)
		if len(overloads) > 0 {
			// Find the overload that matches this implementation's signature
			declaredMethod, methodExists = a.findMatchingOverloadForImplementation(decl, overloads, className)
		}
	} else {
		// For regular methods, check all overloads
		overloads := classType.GetMethodOverloads(methodName)
		if len(overloads) > 0 {
			declaredMethod, methodExists = a.findMatchingOverloadForImplementation(decl, overloads, className)
		} else {
			// Fallback to simple lookup for non-overloaded methods
			declaredMethod, methodExists = classType.Methods[methodName]
		}
	}

	if !methodExists {
		a.addError("method '%s' not declared in class '%s' at %s",
			methodName, className, decl.Token.Pos.String())
		return
	}

	// Task 9.282: Validate signature matches the declaration (already done in findMatchingOverloadForImplementation for overloads)
	// For non-overloaded methods, still validate
	if len(classType.GetMethodOverloads(methodName)) <= 1 && len(classType.GetConstructorOverloads(methodName)) <= 1 {
		if err := a.validateMethodSignature(decl, declaredMethod, className); err != nil {
			a.addError("%s at %s", err.Error(), decl.Token.Pos.String())
			return
		}
	}

	// Task 9.283: Clear the forward flag since we now have an implementation
	delete(classType.ForwardedMethods, methodName)

	// Set the current class context
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	// Use analyzeMethodDecl to analyze the method body with proper scope
	// This will set up Self, fields, and all method scope correctly
	a.analyzeMethodDecl(decl, classType)
}

// findMatchingOverloadForImplementation finds the declared overload that matches the implementation signature
// Task 9.19: Support for overloaded constructor implementations
func (a *Analyzer) findMatchingOverloadForImplementation(implDecl *ast.FunctionDecl, overloads []*types.MethodInfo, className string) (*types.FunctionType, bool) {
	// Resolve implementation parameter count
	implParamCount := len(implDecl.Parameters)

	// Find overloads with matching parameter count
	matchingCount := make([]*types.MethodInfo, 0)
	for _, overload := range overloads {
		if len(overload.Signature.Parameters) == implParamCount {
			matchingCount = append(matchingCount, overload)
		}
	}

	if len(matchingCount) == 0 {
		return nil, false
	}

	if len(matchingCount) == 1 {
		// Only one overload with matching count - use it
		return matchingCount[0].Signature, true
	}

	// Multiple overloads with same count - match by parameter types
	for _, overload := range matchingCount {
		matches := true
		for i, param := range implDecl.Parameters {
			if param.Type == nil {
				continue // Allow omitting types in implementation
			}
			paramType, err := a.resolveType(param.Type.Name)
			if err != nil || !paramType.Equals(overload.Signature.Parameters[i]) {
				matches = false
				break
			}
		}
		if matches {
			return overload.Signature, true
		}
	}

	// No exact match found - return the first one with matching count
	// The validateMethodSignature will report the error
	return matchingCount[0].Signature, true
}

// validateMethodSignature validates that an out-of-line method implementation
// signature matches the forward declaration (Task 9.282)
func (a *Analyzer) validateMethodSignature(implDecl *ast.FunctionDecl, declaredType *types.FunctionType, className string) error {
	// Resolve parameter types from implementation
	implParamTypes := make([]types.Type, 0, len(implDecl.Parameters))
	for _, param := range implDecl.Parameters {
		if param.Type == nil {
			// DWScript allows omitting parameter types in implementation if they match declaration
			// We'll accept this and rely on the declared types
			continue
		}
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			return fmt.Errorf("unknown parameter type '%s' in method '%s.%s'",
				param.Type.Name, className, implDecl.Name.Value)
		}
		implParamTypes = append(implParamTypes, paramType)
	}

	// If implementation specifies parameter types, validate count matches
	if len(implParamTypes) > 0 && len(implParamTypes) != len(declaredType.Parameters) {
		return fmt.Errorf("method '%s.%s' implementation has %d parameters, but declaration has %d",
			className, implDecl.Name.Value, len(implParamTypes), len(declaredType.Parameters))
	}

	// Validate parameter types match (if implementation specifies them)
	for i, implType := range implParamTypes {
		if i >= len(declaredType.Parameters) {
			break
		}
		declType := declaredType.Parameters[i]
		if !implType.Equals(declType) {
			return fmt.Errorf("method '%s.%s' parameter %d has type %s in implementation, but %s in declaration",
				className, implDecl.Name.Value, i+1, implType.String(), declType.String())
		}
	}

	// Resolve return type from implementation (if specified)
	var implReturnType types.Type
	if implDecl.ReturnType != nil {
		var err error
		implReturnType, err = a.resolveType(implDecl.ReturnType.Name)
		if err != nil {
			return fmt.Errorf("unknown return type '%s' in method '%s.%s'",
				implDecl.ReturnType.Name, className, implDecl.Name.Value)
		}

		// Validate return type matches
		if !implReturnType.Equals(declaredType.ReturnType) {
			return fmt.Errorf("method '%s.%s' has return type %s in implementation, but %s in declaration",
				className, implDecl.Name.Value, implReturnType.String(), declaredType.ReturnType.String())
		}
	}

	return nil
}

// inheritParentConstructors copies accessible parent constructors to a child class (Task 9.2).
// In DWScript, child classes inherit parent constructors if the child doesn't declare any.
// Only public and protected constructors are inherited (private constructors are not).
func (a *Analyzer) inheritParentConstructors(childClass *types.ClassType, parentClass *types.ClassType) {
	// Iterate through all parent constructor overloads
	for ctorName, overloads := range parentClass.ConstructorOverloads {
		for _, parentCtor := range overloads {
			// Task 9.2: Private constructors are not inherited
			visibility := ast.Visibility(parentCtor.Visibility)
			if visibility == ast.VisibilityPrivate {
				continue
			}

			// Copy the constructor signature, updating return type to child class
			// Create new function type with same parameters but child class return type
			childCtorType := types.NewFunctionTypeWithMetadata(
				parentCtor.Signature.Parameters,
				parentCtor.Signature.ParamNames,
				parentCtor.Signature.DefaultValues,
				parentCtor.Signature.LazyParams,
				parentCtor.Signature.VarParams,
				parentCtor.Signature.ConstParams,
				childClass, // Returns instance of the child class, not parent
			)

			// Create method info for the inherited constructor
			childCtorInfo := &types.MethodInfo{
				Signature:            childCtorType,
				IsVirtual:            parentCtor.IsVirtual,
				IsOverride:           false, // Inherited constructors are not marked as override
				IsAbstract:           parentCtor.IsAbstract,
				IsForwarded:          false,
				IsClassMethod:        parentCtor.IsClassMethod,
				HasOverloadDirective: parentCtor.HasOverloadDirective,
				Visibility:           parentCtor.Visibility,
			}

			// Add inherited constructor to child class
			// Use lowercase for case-insensitive lookup
			lowerCtorName := strings.ToLower(ctorName)
			if _, exists := childClass.Constructors[lowerCtorName]; !exists {
				childClass.Constructors[lowerCtorName] = childCtorType
			}
			childClass.AddConstructorOverload(ctorName, childCtorInfo)
		}
	}

	// If no constructors were inherited (all were private), generate implicit default
	if len(childClass.Constructors) == 0 && len(childClass.ConstructorOverloads) == 0 {
		a.synthesizeDefaultConstructor(childClass)
	}
}

// synthesizeDefaultConstructor generates an implicit default constructor for a class (Task 9.1).
// DWScript automatically provides a parameterless `Create` constructor for classes
// that don't declare any explicit constructors.
func (a *Analyzer) synthesizeDefaultConstructor(classType *types.ClassType) {
	constructorName := "Create"

	// Create function type: no parameters, returns the class type
	funcType := types.NewFunctionTypeWithMetadata(
		[]types.Type{},  // No parameters
		[]string{},      // No parameter names
		[]interface{}{}, // No default values
		[]bool{},        // No lazy params
		[]bool{},        // No var params
		[]bool{},        // No const params
		classType,       // Returns instance of the class
	)

	// Create method info for the implicit constructor
	methodInfo := &types.MethodInfo{
		Signature:            funcType,
		IsVirtual:            false,
		IsOverride:           false,
		IsAbstract:           false,
		IsForwarded:          false,
		IsClassMethod:        false,
		HasOverloadDirective: false,
		Visibility:           int(ast.VisibilityPublic), // Public access
	}

	// Add to class constructor maps
	// Use lowercase for case-insensitive lookup
	classType.Constructors[strings.ToLower(constructorName)] = funcType
	classType.AddConstructorOverload(constructorName, methodInfo)
}

// analyzeMethodDecl analyzes a method declaration within a class
func (a *Analyzer) analyzeMethodDecl(method *ast.FunctionDecl, classType *types.ClassType) {
	// Convert parameter types and extract metadata
	// Task 9.21.4.3: Extract parameter metadata including variadic detection
	// Task 9.1: Extract default values for optional parameters
	paramTypes := make([]types.Type, 0, len(method.Parameters))
	paramNames := make([]string, 0, len(method.Parameters))
	defaultValues := make([]interface{}, 0, len(method.Parameters))
	lazyParams := make([]bool, 0, len(method.Parameters))
	varParams := make([]bool, 0, len(method.Parameters))
	constParams := make([]bool, 0, len(method.Parameters))

	for _, param := range method.Parameters {
		if param.Type == nil {
			a.addError("parameter '%s' missing type annotation in method '%s'",
				param.Name.Value, method.Name.Value)
			return
		}

		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in method '%s': %v",
				param.Type.Name, method.Name.Value, err)
			return
		}
		paramTypes = append(paramTypes, paramType)
		paramNames = append(paramNames, param.Name.Value)
		defaultValues = append(defaultValues, param.DefaultValue) // Store default value (may be nil)
		lazyParams = append(lazyParams, param.IsLazy)
		varParams = append(varParams, param.ByRef)
		constParams = append(constParams, param.IsConst)
	}

	// Task 9.17: Validate constructors don't have explicit return types
	if method.IsConstructor && method.ReturnType != nil {
		a.addError("constructor '%s' cannot have an explicit return type at %s",
			method.Name.Value, method.Token.Pos.String())
		return
	}

	// Determine return type
	var returnType types.Type
	if method.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(method.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in method '%s': %v",
				method.ReturnType.Name, method.Name.Value, err)
			return
		}
	} else if method.IsConstructor {
		// Task 9.17: Constructors implicitly return the class type
		returnType = classType
	} else {
		returnType = types.VOID
	}

	// Create function type with metadata and add to class methods
	// Task 9.21.4.3: Detect variadic parameters (last parameter is array type)
	// Task 9.1: Include default values in function type metadata
	var funcType *types.FunctionType
	if len(paramTypes) > 0 {
		// Check if last parameter is a dynamic array (variadic)
		lastParamType := paramTypes[len(paramTypes)-1]
		if arrayType, ok := lastParamType.(*types.ArrayType); ok && arrayType.IsDynamic() {
			// This is a variadic parameter
			variadicType := arrayType.ElementType
			funcType = types.NewVariadicFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams,
				variadicType, returnType,
			)
		} else {
			// Regular (non-variadic) function
			funcType = types.NewFunctionTypeWithMetadata(
				paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType,
			)
		}
	} else {
		// No parameters - create regular function type
		funcType = types.NewFunctionTypeWithMetadata(
			paramTypes, paramNames, defaultValues, lazyParams, varParams, constParams, returnType,
		)
	}

	// Task 9.61: Add method to overload set instead of overwriting
	methodInfo := &types.MethodInfo{
		Signature:            funcType,
		IsVirtual:            method.IsVirtual,
		IsOverride:           method.IsOverride,
		IsAbstract:           method.IsAbstract,
		IsForwarded:          method.Body == nil,
		IsClassMethod:        method.IsClassMethod,
		HasOverloadDirective: method.IsOverload,
		Visibility:           int(method.Visibility),
	}

	// Task 9.62: Check for duplicate/ambiguous signatures before adding
	existingOverloads := classType.GetMethodOverloads(method.Name.Value)
	if method.IsConstructor {
		existingOverloads = classType.GetConstructorOverloads(method.Name.Value)
	}

	// Check if this is an implementation for a forward declaration
	isForward := method.Body == nil

	for _, existing := range existingOverloads {
		// Task 9.63: Check if signatures are identical (duplicate) - use DWScript error format
		if a.methodSignaturesMatch(funcType, existing.Signature) {
			// Task 9.60: Check if this is a forward + implementation pair (like in symbol_table.go:211-222)
			if existing.IsForwarded && !isForward {
				// Implementation following forward declaration
				// Update the existing forward declaration instead of adding a new overload
				existing.IsForwarded = false
				existing.Signature = funcType

				// Update method metadata for the implementation
				classType.ClassMethodFlags[method.Name.Value] = method.IsClassMethod
				classType.VirtualMethods[method.Name.Value] = method.IsVirtual
				classType.OverrideMethods[method.Name.Value] = method.IsOverride
				classType.AbstractMethods[method.Name.Value] = method.IsAbstract
				return
			}

			// True duplicate (both forward or both implementation)
			a.addError("Syntax Error: There is already a method with name \"%s\" [line: %d, column: %d]",
				method.Name.Value, method.Token.Pos.Line, method.Token.Pos.Column)
			return
		}

		// Task 9.63: Check if parameters match but return types differ (ambiguous)
		if a.parametersMatch(funcType, existing.Signature) && !funcType.ReturnType.Equals(existing.Signature.ReturnType) {
			a.addError("Syntax Error: Overload of \"%s\" will be ambiguous with a previously declared version [line: %d, column: %d]",
				method.Name.Value, method.Token.Pos.Line, method.Token.Pos.Column)
			return
		}
	}

	if method.IsConstructor {
		classType.AddConstructorOverload(method.Name.Value, methodInfo)
	} else {
		classType.AddMethodOverload(method.Name.Value, methodInfo)
	}

	// Store method metadata in legacy maps for backward compatibility
	classType.ClassMethodFlags[method.Name.Value] = method.IsClassMethod
	classType.VirtualMethods[method.Name.Value] = method.IsVirtual
	classType.OverrideMethods[method.Name.Value] = method.IsOverride
	classType.AbstractMethods[method.Name.Value] = method.IsAbstract

	// Task 9.280: Mark method as forward if it has no body (declaration without implementation)
	// Methods declared in class body without implementation are implicitly forward
	if method.Body == nil {
		classType.ForwardedMethods[method.Name.Value] = true
	}

	// Store method visibility
	// Only set visibility if this is the first time we're seeing this method (declaration in class body)
	// Method implementations outside the class shouldn't overwrite the visibility
	if _, exists := classType.MethodVisibility[method.Name.Value]; !exists {
		classType.MethodVisibility[method.Name.Value] = int(method.Visibility)
	}

	// Analyze method body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Check if this is a class method (static method)
	if method.IsClassMethod {
		// Class methods (static methods) do NOT have access to Self or instance fields
		// They can only access class variables (static fields)
		// DO NOT add Self to scope
		// DO NOT add instance fields to scope

		// Add class variables to scope
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType)
		}

		// If class has parent, add parent class variables too
		if classType.Parent != nil {
			a.addParentClassVarsToScope(classType.Parent)
		}
	} else {
		// Instance method - add Self reference to method scope
		a.symbols.Define("Self", classType)

		// Add class fields to method scope
		for fieldName, fieldType := range classType.Fields {
			a.symbols.Define(fieldName, fieldType)
		}

		// Add class variables to method scope
		// Instance methods can also access class variables
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType)
		}

		// If class has parent, add parent fields and class variables too
		if classType.Parent != nil {
			a.addParentFieldsToScope(classType.Parent)
			a.addParentClassVarsToScope(classType.Parent)
		}
	}

	// Add parameters to method scope (both instance and class methods have parameters)
	for i, param := range method.Parameters {
		a.symbols.Define(param.Name.Value, paramTypes[i])
	}

	// For methods with return type, add Result variable
	if returnType != types.VOID {
		a.symbols.Define("Result", returnType)
		a.symbols.Define(method.Name.Value, returnType)
	}

	// Set current function for return statement checking
	previousFunc := a.currentFunction
	a.currentFunction = method
	defer func() { a.currentFunction = previousFunc }()

	// Set inClassMethod flag for class methods
	previousInClassMethod := a.inClassMethod
	a.inClassMethod = method.IsClassMethod
	defer func() { a.inClassMethod = previousInClassMethod }()

	// Validate virtual/override usage
	a.validateVirtualOverride(method, classType, funcType)

	// Analyze method body
	if method.Body != nil {
		a.analyzeBlock(method.Body)
	}
}

// validateVirtualOverride validates virtual/override method declarations (Task 9.61: Updated for overloading)
// Task 9.4.1: Updated to support virtual/override on constructors
func (a *Analyzer) validateVirtualOverride(method *ast.FunctionDecl, classType *types.ClassType, methodType *types.FunctionType) {
	methodName := method.Name.Value
	isConstructor := method.IsConstructor

	// If method is marked override, validate parent has virtual method with matching signature
	if method.IsOverride {
		if classType.Parent == nil {
			a.addError("method '%s' marked as override, but class has no parent", methodName)
			return
		}

		// Task 9.4.1: Find matching overload in parent class hierarchy (check both methods and constructors)
		var parentOverload *types.MethodInfo
		if isConstructor {
			parentOverload = a.findMatchingConstructorInParent(methodName, methodType, classType.Parent)
		} else {
			parentOverload = a.findMatchingOverloadInParent(methodName, methodType, classType.Parent)
		}

		if parentOverload == nil {
			// Check if method/constructor with this name exists at all
			var hasParentMember bool
			if isConstructor {
				hasParentMember = a.hasConstructorWithName(methodName, classType.Parent)
			} else {
				hasParentMember = a.hasMethodWithName(methodName, classType.Parent)
			}

			if hasParentMember {
				// Method/constructor name exists but signature doesn't match any parent overload
				a.addError("method '%s' marked as override, but no matching signature exists in parent class", methodName)
			} else {
				// Method/constructor name doesn't exist at all in parent
				a.addError("method '%s' marked as override, but no such method exists in parent class", methodName)
			}
			return
		}

		// Check that parent method/constructor is virtual, override, or abstract (Task 9.81)
		// Abstract methods are implicitly virtual and can be overridden
		if !parentOverload.IsVirtual && !parentOverload.IsOverride && !parentOverload.IsAbstract {
			a.addError("method '%s' marked as override, but parent method is not virtual", methodName)
			return
		}

		// Task 9.61.4: Add hint if override is part of an overload set but doesn't have overload directive
		// Check if there are other overloads of this method/constructor in the current class
		var currentOverloads []*types.MethodInfo
		if isConstructor {
			currentOverloads = classType.GetConstructorOverloads(methodName)
		} else {
			currentOverloads = classType.GetMethodOverloads(methodName)
		}
		if len(currentOverloads) > 1 && !method.IsOverload {
			a.addHint("Overloaded method \"%s\" should be marked with the \"overload\" directive at %s",
				methodName, method.Token.Pos.String())
		}
	}

	// Warn if redefining virtual method without override keyword
	// Note: Constructors can be marked as virtual, so this check applies to both methods and constructors
	if !method.IsOverride && !method.IsVirtual && classType.Parent != nil {
		// Task 9.4.1: Check if any parent overload with matching signature is virtual
		var parentOverload *types.MethodInfo
		if isConstructor {
			parentOverload = a.findMatchingConstructorInParent(methodName, methodType, classType.Parent)
		} else {
			parentOverload = a.findMatchingOverloadInParent(methodName, methodType, classType.Parent)
		}

		if parentOverload != nil && (parentOverload.IsVirtual || parentOverload.IsOverride) {
			a.addError("method '%s' hides virtual parent method; use 'override' keyword", methodName)
		}
	}
}

// checkMethodOverriding checks if overridden methods have compatible signatures
func (a *Analyzer) checkMethodOverriding(class, parent *types.ClassType) {
	for methodName, childMethodType := range class.Methods {
		// Check if method exists in parent
		parentMethodType, found := parent.GetMethod(methodName)
		if !found {
			// New method in child class - OK
			continue
		}

		// Method exists in parent - check signature compatibility
		if !childMethodType.Equals(parentMethodType) {
			a.addError("method '%s' signature mismatch in class '%s': expected %s, got %s",
				methodName, class.Name, parentMethodType.String(), childMethodType.String())
		}
	}
}

// checkVisibility checks if a member (field or method) is accessible from the current context
// Returns true if accessible, false otherwise.
//
// Visibility rules:
//   - Private: only accessible from the same class
//   - Protected: accessible from the same class and all descendants
//   - Public: accessible from anywhere
//
// Parameters:
//   - memberClass: the class that owns the member
//   - visibility: the visibility level of the member (ast.Visibility as int)
//   - memberName: the name of the member (for error messages)
//   - memberType: "field" or "method" (for error messages)
func (a *Analyzer) checkVisibility(memberClass *types.ClassType, visibility int, _, _ string) bool {
	// Public is always accessible
	if visibility == int(ast.VisibilityPublic) {
		return true
	}

	// If we're analyzing code outside any class context, only public members are accessible
	if a.currentClass == nil {
		return false
	}

	// Private members are only accessible from the same class
	if visibility == int(ast.VisibilityPrivate) {
		return a.currentClass.Name == memberClass.Name
	}

	// Protected members are accessible from the same class and descendants
	if visibility == int(ast.VisibilityProtected) {
		// Same class?
		if a.currentClass.Name == memberClass.Name {
			return true
		}

		// Check if current class inherits from member's class
		return a.isDescendantOf(a.currentClass, memberClass)
	}

	// Should not reach here, but default to false for safety
	return false
}

// analyzeNewExpression analyzes object creation
// Handles both:
//   - new TClass(args)
//   - TClass.Create(args)
//
// Task 9.18: NewExpression semantic validation with constructor overload resolution
func (a *Analyzer) analyzeNewExpression(expr *ast.NewExpression) types.Type {
	className := expr.ClassName.Value

	// Look up class in registry
	// Task 9.285: Use lowercase for case-insensitive lookup
	classType, found := a.classes[strings.ToLower(className)]
	if !found {
		a.addError("undefined class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Task 9.18: Check if trying to instantiate an abstract class
	if classType.IsAbstract {
		a.addError("cannot instantiate abstract class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Task 9.13-9.16: Get all constructor overloads (assuming "Create" as default constructor name)
	// In DWScript, constructors are typically named "Create" but can have other names
	// For NewExpression, we assume "Create" unless the AST specifies otherwise
	constructorName := "Create"
	constructorOverloads := a.getMethodOverloadsInHierarchy(constructorName, classType)

	if len(constructorOverloads) == 0 {
		// No explicit constructor - use implicit default constructor
		// Task 9.17: Validate that no arguments are provided for default constructor
		if len(expr.Arguments) > 0 {
			a.addError("class '%s' has no constructor, cannot pass arguments at %s",
				className, expr.Token.Pos.String())
		}
		return classType
	}

	// Task 9.15: Filter out implicit parameterless constructor
	// The implicit constructor has Visibility=0 and len(Parameters)=0 and is only added by getMethodOverloadsInHierarchy
	// We should ignore it if there are explicit constructors with parameters
	validConstructors := make([]*types.MethodInfo, 0, len(constructorOverloads))
	hasExplicitConstructor := false
	for _, ctor := range constructorOverloads {
		// Check if this is an explicit constructor (has visibility set OR has parameters)
		if ctor.Visibility != 0 || len(ctor.Signature.Parameters) > 0 {
			validConstructors = append(validConstructors, ctor)
			hasExplicitConstructor = true
		}
	}

	// If we only found implicit constructors but need explicit ones, use empty list
	if !hasExplicitConstructor {
		validConstructors = constructorOverloads
	}

	// Task 9.13: Select constructor based on argument count first
	var selectedConstructor *types.MethodInfo
	var selectedSignature *types.FunctionType

	// Find constructors with matching argument count
	matchingCountConstructors := make([]*types.MethodInfo, 0)
	for _, ctor := range validConstructors {
		if len(ctor.Signature.Parameters) == len(expr.Arguments) {
			matchingCountConstructors = append(matchingCountConstructors, ctor)
		}
	}

	if len(matchingCountConstructors) == 0 {
		// Task 9.15: No constructor with matching argument count
		if len(validConstructors) > 0 {
			// Report the expected count from the first constructor
			a.addError("constructor '%s' expects %d arguments, got %d at %s",
				constructorName, len(validConstructors[0].Signature.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
		} else {
			a.addError("class '%s' has no constructor that accepts %d arguments at %s",
				className, len(expr.Arguments), expr.Token.Pos.String())
		}
		return classType
	}

	// Now select the best match based on argument types
	if len(matchingCountConstructors) == 1 {
		selectedConstructor = matchingCountConstructors[0]
		selectedSignature = selectedConstructor.Signature
	} else {
		// Multiple constructors with same count - resolve by type
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return classType
			}
			argTypes[i] = argType
		}

		candidates := make([]*Symbol, len(matchingCountConstructors))
		for i, overload := range matchingCountConstructors {
			candidates[i] = &Symbol{
				Type: overload.Signature,
			}
		}

		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			a.addError("there is no constructor for class '%s' that matches these argument types at %s",
				className, expr.Token.Pos.String())
			return classType
		}

		selectedSignature = selected.Type.(*types.FunctionType)
		for _, overload := range matchingCountConstructors {
			if overload.Signature == selectedSignature {
				selectedConstructor = overload
				break
			}
		}
	}

	// Task 9.16: Check constructor visibility
	var ownerClass *types.ClassType
	for class := classType; class != nil; class = class.Parent {
		if class.HasConstructor(constructorName) {
			ownerClass = class
			break
		}
	}
	if ownerClass != nil && selectedConstructor != nil {
		visibility := selectedConstructor.Visibility
		// Note: Visibility 0 is private, so we must check all values including 0
		if !a.checkVisibility(ownerClass, visibility, constructorName, "constructor") {
			visibilityStr := ast.Visibility(visibility).String()
			a.addError("cannot access %s constructor '%s' of class '%s' at %s",
				visibilityStr, constructorName, ownerClass.Name, expr.Token.Pos.String())
			return classType
		}
	}

	// Task 9.14: Validate argument types (more detailed error messages)
	for i, arg := range expr.Arguments {
		if i >= len(selectedSignature.Parameters) {
			break
		}

		paramType := selectedSignature.Parameters[i]
		argType := a.analyzeExpressionWithExpectedType(arg, paramType)
		if argType != nil && !a.canAssign(argType, paramType) {
			a.addError("argument %d to constructor of '%s' has type %s, expected %s at %s",
				i+1, className, argType.String(), paramType.String(),
				expr.Token.Pos.String())
		}
	}

	return classType
}

// analyzeMemberAccessExpression analyzes member access
func (a *Analyzer) analyzeMemberAccessExpression(expr *ast.MemberAccessExpression) types.Type {
	// Analyze the object expression
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		// Error already reported
		return nil
	}

	// Check if object is a class or record type
	memberName := expr.Member.Value

	// Task 9.73.5: Resolve type aliases to get the underlying type
	// This allows member access on type alias variables like TBaseClass
	objectTypeResolved := types.GetUnderlyingType(objectType)

	// Handle record type field access
	if _, ok := objectTypeResolved.(*types.RecordType); ok {
		return a.analyzeRecordFieldAccess(expr.Object, memberName)
	}

	// Task 9.73.2: Handle metaclass type (class of T) - allows calling constructors through metaclass
	// Convert ClassOfType to the underlying ClassType so we can check for constructors and class members
	if metaclassType, ok := objectTypeResolved.(*types.ClassOfType); ok {
		baseClass := metaclassType.ClassType
		if baseClass != nil {
			// Continue with the base class type to check for constructors, class methods, and class variables
			// This allows expressions like TBase.Create, TBase.SomeClassMethod, or TBase.ClassVar to work
			objectTypeResolved = baseClass
		}
	}

	// Handle class type
	classType, ok := objectTypeResolved.(*types.ClassType)
	if !ok {
		// Task 9.83: For non-class/record types (like String, Integer), check helpers
		// Prefer helper properties before methods so that property-style access
		// (e.g., i.ToString) resolves correctly when parentheses are omitted.
		_, helperProp := a.hasHelperProperty(objectType, memberName)
		if helperProp != nil {
			return helperProp.Type
		}

		_, helperMethod := a.hasHelperMethod(objectType, memberName)
		if helperMethod != nil {
			return helperMethod
		}

		// Task 9.54: Check for helper class constants (for scoped enum access like TColor.Red)
		_, helperConst := a.hasHelperClassConst(objectType, memberName)
		if helperConst != nil {
			// For enum types, the constant is the enum value, so return the enum type itself
			if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
				return objectType
			}
			// For other types, we'd need to determine the constant's type
			// For now, return the object type (conservative approach)
			return objectType
		}

		a.addError("member access on type %s requires a helper, got no helper with member '%s' at %s",
			objectType.String(), memberName, expr.Token.Pos.String())
		return nil
	}

	// Handle built-in properties/methods available on all objects (inherited from TObject)
	if memberName == "ClassName" {
		// ClassName returns String
		return types.STRING
	}

	// Look up field in class (including inherited fields)
	fieldType, found := classType.GetField(memberName)
	if found {
		// Check field visibility
		fieldOwner := a.getFieldOwner(classType, memberName)
		if fieldOwner != nil {
			visibility, hasVisibility := fieldOwner.FieldVisibility[memberName]
			if hasVisibility && !a.checkVisibility(fieldOwner, visibility, memberName, "field") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot access %s field '%s' of class '%s' at %s",
					visibilityStr, memberName, fieldOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		return fieldType
	}

	// Task 9.3: Look up property in class (including inherited properties)
	propInfo, propFound := classType.GetProperty(memberName)
	if propFound {
		// Property access returns the property type
		return propInfo.Type
	}

	// Task 9.68: Check for constructors first (constructors are stored separately)
	constructorOverloads := classType.GetConstructorOverloads(memberName)
	if len(constructorOverloads) > 0 {
		// Task 9.21: Check if this is a parameterless constructor
		// Parameterless constructors can be called without parentheses (auto-invoked)
		hasParameterless := false
		for _, ctor := range constructorOverloads {
			if len(ctor.Signature.Parameters) == 0 {
				hasParameterless = true
				break
			}
		}

		// If there's a parameterless constructor, treat member access as auto-invocation
		// and return the class type directly (not a method pointer)
		if hasParameterless {
			return classType
		}

		// Constructor has parameters - return method pointer type for deferred invocation
		if len(constructorOverloads) == 1 {
			return types.NewMethodPointerType(constructorOverloads[0].Signature.Parameters, classType)
		}
		// Multiple constructor overloads - return a generic constructor pointer type
		// The actual overload will be resolved at call time
		return types.NewMethodPointerType([]types.Type{}, classType)
	}

	// Look up method in class (for method references)
	methodType, found := classType.GetMethod(memberName)
	if found {
		// Check method visibility
		methodOwner := a.getMethodOwner(classType, memberName)
		if methodOwner != nil {
			visibility, hasVisibility := methodOwner.MethodVisibility[memberName]
			if hasVisibility && !a.checkVisibility(methodOwner, visibility, memberName, "method") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot access %s method '%s' of class '%s' at %s",
					visibilityStr, memberName, methodOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		// When accessing a method as a value (not calling it),
		// return a method pointer type instead of just the function type
		return types.NewMethodPointerType(methodType.Parameters, methodType.ReturnType)
	}

	// Task 9.83: Check helpers for methods
	// If not found in class, check if any helpers extend this type
	_, helperMethod := a.hasHelperMethod(objectType, memberName)
	if helperMethod != nil {
		return helperMethod
	}

	// Task 9.83: Check helpers for properties
	_, helperProp := a.hasHelperProperty(objectType, memberName)
	if helperProp != nil {
		return helperProp.Type
	}

	// Task 9.22: Check for class constants (with inheritance support)
	if _, found := classType.GetConstant(memberName); found {
		// Check constant visibility
		constantOwner := a.getConstantOwner(classType, memberName)
		if constantOwner != nil {
			visibility, hasVisibility := constantOwner.ConstantVisibility[memberName]
			if hasVisibility && !a.checkVisibility(constantOwner, visibility, memberName, "constant") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot access %s constant '%s' of class '%s' at %s",
					visibilityStr, memberName, constantOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		// The constant exists and visibility is allowed
		// We don't know its exact type at compile time since it's evaluated at runtime
		// For now, return VARIANT to indicate it's valid but type is determined at runtime
		// In a full implementation, we'd analyze the constant expression to determine type
		return types.VARIANT // Accept any type for constants
	}

	// Member not found
	a.addError("class '%s' has no member '%s' at %s",
		classType.Name, memberName, expr.Token.Pos.String())
	return nil
}

// ============================================================================
// Abstract Class/Method Validation
// ============================================================================

// validateAbstractClass validates abstract class rules:
// 1. Abstract methods can only exist in abstract classes
// 2. Concrete classes must implement all inherited abstract methods
// 3. Abstract methods are implicitly virtual
func (a *Analyzer) validateAbstractClass(classType *types.ClassType, decl *ast.ClassDecl) {
	// Rule 1: Abstract methods can only exist in abstract classes
	for methodName, isAbstract := range classType.AbstractMethods {
		if isAbstract && !classType.IsAbstract {
			a.addError("abstract method '%s' can only be declared in an abstract class at %s",
				methodName, decl.Token.Pos.String())
		}

		// Abstract methods are implicitly virtual
		if isAbstract {
			classType.VirtualMethods[methodName] = true
		}
	}

	// Rule 2: Concrete classes must implement all inherited abstract methods
	if !classType.IsAbstract {
		unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
		if len(unimplementedMethods) > 0 {
			for _, methodName := range unimplementedMethods {
				a.addError("concrete class '%s' must implement abstract method '%s' at %s",
					classType.Name, methodName, decl.Token.Pos.String())
			}
		}
	}
}
