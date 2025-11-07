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
	}

	// Create new class type
	classType := types.NewClassType(className, parentClass)

	// Set abstract flag (Task 7.65e)
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

	// Analyze and add fields (Task 7.55, 7.62)
	fieldNames := make(map[string]bool)
	classVarNames := make(map[string]bool)
	for _, field := range decl.Fields {
		fieldName := field.Name.Value

		// Check if this is a class variable (static field) - Task 7.62
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

			// Store class variable type in ClassType - Task 7.62
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

			// Store field visibility (Task 7.63f)
			classType.FieldVisibility[fieldName] = int(field.Visibility)
		}
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

	// Analyze properties (Task 8.46-8.51)
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

// analyzeMethodImplementation analyzes a method implementation outside a class (Task 7.63v-z)
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
	methodName := decl.Name.Value
	declaredMethod, methodExists := classType.Methods[methodName]
	if !methodExists {
		// Check constructor map too
		if decl.IsConstructor {
			declaredMethod, methodExists = classType.Constructors[methodName]
		}
	}

	if !methodExists {
		a.addError("method '%s' not declared in class '%s' at %s",
			methodName, className, decl.Token.Pos.String())
		return
	}

	// Task 9.282: Validate signature matches the declaration
	if err := a.validateMethodSignature(decl, declaredMethod, className); err != nil {
		a.addError("%s at %s", err.Error(), decl.Token.Pos.String())
		return
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

// analyzeMethodDecl analyzes a method declaration within a class (Task 7.56, 7.61)
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

	// Store method visibility (Task 7.63f)
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
func (a *Analyzer) validateVirtualOverride(method *ast.FunctionDecl, classType *types.ClassType, methodType *types.FunctionType) {
	methodName := method.Name.Value

	// If method is marked override, validate parent has virtual method with matching signature
	if method.IsOverride {
		if classType.Parent == nil {
			a.addError("method '%s' marked as override, but class has no parent", methodName)
			return
		}

		// Task 9.61: Find matching overload in parent class hierarchy
		parentOverload := a.findMatchingOverloadInParent(methodName, methodType, classType.Parent)
		if parentOverload == nil {
			// Check if method with this name exists at all
			if a.hasMethodWithName(methodName, classType.Parent) {
				// Method name exists but signature doesn't match any parent overload
				a.addError("method '%s' marked as override, but no matching signature exists in parent class", methodName)
			} else {
				// Method name doesn't exist at all in parent
				a.addError("method '%s' marked as override, but no such method exists in parent class", methodName)
			}
			return
		}

		// Check that parent method is virtual or override (Task 9.61)
		if !parentOverload.IsVirtual && !parentOverload.IsOverride {
			a.addError("method '%s' marked as override, but parent method is not virtual", methodName)
			return
		}

		// Task 9.61.4: Add hint if override is part of an overload set but doesn't have overload directive
		// Check if there are other overloads of this method in the current class
		currentOverloads := classType.GetMethodOverloads(methodName)
		if len(currentOverloads) > 1 && !method.IsOverload {
			a.addHint("Overloaded method \"%s\" should be marked with the \"overload\" directive at %s",
				methodName, method.Token.Pos.String())
		}
	}

	// Warn if redefining virtual method without override keyword
	if !method.IsOverride && !method.IsVirtual && classType.Parent != nil {
		// Task 9.61: Check if any parent overload with matching signature is virtual
		parentOverload := a.findMatchingOverloadInParent(methodName, methodType, classType.Parent)
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

// checkVisibility checks if a member (field or method) is accessible from the current context (Task 7.63g-l).
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
	// Public is always accessible (Task 7.63i)
	if visibility == int(ast.VisibilityPublic) {
		return true
	}

	// If we're analyzing code outside any class context, only public members are accessible
	if a.currentClass == nil {
		return false
	}

	// Private members are only accessible from the same class (Task 7.63g, 7.63l)
	if visibility == int(ast.VisibilityPrivate) {
		return a.currentClass.Name == memberClass.Name
	}

	// Protected members are accessible from the same class and descendants (Task 7.63h)
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

// analyzeNewExpression analyzes object creation (Task 7.57, 7.65f)
func (a *Analyzer) analyzeNewExpression(expr *ast.NewExpression) types.Type {
	className := expr.ClassName.Value

	// Look up class in registry
	// Task 9.285: Use lowercase for case-insensitive lookup
	classType, found := a.classes[strings.ToLower(className)]
	if !found {
		a.addError("undefined class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Check if trying to instantiate an abstract class (Task 7.65f)
	if classType.IsAbstract {
		a.addError("cannot instantiate abstract class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Check if class has a constructor
	constructorType, hasConstructor := classType.GetMethod("Create")
	if hasConstructor {
		if owner := a.getMethodOwner(classType, "Create"); owner != nil {
			if visibility, ok := owner.MethodVisibility["Create"]; ok {
				if !a.checkVisibility(owner, visibility, "Create", "method") {
					visibilityStr := ast.Visibility(visibility).String()
					a.addError("cannot access %s constructor 'Create' of class '%s' at %s",
						visibilityStr, owner.Name, expr.Token.Pos.String())
					return classType
				}
			}
		}

		// Validate constructor arguments
		if len(expr.Arguments) != len(constructorType.Parameters) {
			a.addError("constructor for class '%s' expects %d arguments, got %d at %s",
				className, len(constructorType.Parameters), len(expr.Arguments),
				expr.Token.Pos.String())
			return classType
		}

		// Check argument types
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			expectedType := constructorType.Parameters[i]
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to constructor of '%s' has type %s, expected %s at %s",
					i+1, className, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}
	}
	// If no constructor but arguments provided, that's OK - default constructor

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

	// Handle record type field access
	if _, ok := objectType.(*types.RecordType); ok {
		return a.analyzeRecordFieldAccess(expr.Object, memberName)
	}

	// Handle class type
	classType, ok := objectType.(*types.ClassType)
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
		// Check field visibility (Task 7.63j)
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

	// Look up method in class (for method references)
	methodType, found := classType.GetMethod(memberName)
	if found {
		// Check method visibility (Task 7.63k)
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

	// Member not found
	a.addError("class '%s' has no member '%s' at %s",
		classType.Name, memberName, expr.Token.Pos.String())
	return nil
}

// ============================================================================
// Abstract Class/Method Validation
// ============================================================================

// validateAbstractClass validates abstract class rules:
// 1. Abstract methods can only exist in abstract classes (Task 7.65i)
// 2. Concrete classes must implement all inherited abstract methods (Task 7.65g)
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
