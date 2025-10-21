package semantic

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Class Analysis (Tasks 7.54-7.59)
// ============================================================================

// analyzeClassDecl analyzes a class declaration
func (a *Analyzer) analyzeClassDecl(decl *ast.ClassDecl) {
	className := decl.Name.Value

	// Check if class is already declared (Task 7.55)
	if _, exists := a.classes[className]; exists {
		a.addError("class '%s' already declared at %s", className, decl.Token.Pos.String())
		return
	}

	// Resolve parent class if specified (Task 7.55)
	var parentClass *types.ClassType
	if decl.Parent != nil {
		parentName := decl.Parent.Value
		var found bool
		parentClass, found = a.classes[parentName]
		if !found {
			a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
			return
		}
	}

	// Create new class type
	classType := types.NewClassType(className, parentClass)

	// Set abstract flag (Task 7.65e)
	classType.IsAbstract = decl.IsAbstract

	// Set external flags (Task 7.138)
	classType.IsExternal = decl.IsExternal
	classType.ExternalName = decl.ExternalName

	// Validate external class inheritance (Task 7.139)
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

	// Check for circular inheritance (Task 7.55)
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

			fieldType, err := a.resolveType(field.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for class variable '%s' in class '%s' at %s",
					field.Type.Name, fieldName, className, field.Token.Pos.String())
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

			fieldType, err := a.resolveType(field.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for field '%s' in class '%s' at %s",
					field.Type.Name, fieldName, className, field.Token.Pos.String())
				continue
			}

			// Add instance field to class
			classType.Fields[fieldName] = fieldType

			// Store field visibility (Task 7.63f)
			classType.FieldVisibility[fieldName] = int(field.Visibility)
		}
	}

	// Register class before analyzing methods (so methods can reference the class)
	a.classes[className] = classType

	// Analyze methods (Task 7.56)
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	for _, method := range decl.Methods {
		a.analyzeMethodDecl(method, classType)
	}

	// Analyze constructor if present (Task 7.56)
	if decl.Constructor != nil {
		a.analyzeMethodDecl(decl.Constructor, classType)
	}

	// Register class operators (Stage 8)
	a.registerClassOperators(classType, decl)

	// Check method overriding (Task 7.59)
	if parentClass != nil {
		a.checkMethodOverriding(classType, parentClass)
	}

	// Validate interface implementation (Task 7.100)
	if len(decl.Interfaces) > 0 {
		a.validateInterfaceImplementation(classType, decl)
	}

	// Validate abstract class rules (Task 7.65)
	a.validateAbstractClass(classType, decl)
}

// analyzeMethodImplementation analyzes a method implementation outside a class (Task 7.63v-z)
// This handles code like: function TExample.GetValue: Integer; begin ... end;
func (a *Analyzer) analyzeMethodImplementation(decl *ast.FunctionDecl) {
	className := decl.ClassName.Value

	// Look up the class
	classType, exists := a.classes[className]
	if !exists {
		a.addError("unknown type '%s' at %s", className, decl.Token.Pos.String())
		return
	}

	// Set the current class context
	previousClass := a.currentClass
	a.currentClass = classType
	defer func() { a.currentClass = previousClass }()

	// Use analyzeMethodDecl to analyze the method body with proper scope
	// This will set up Self, fields, and all method scope correctly
	a.analyzeMethodDecl(decl, classType)
}

// analyzeMethodDecl analyzes a method declaration within a class (Task 7.56, 7.61)
func (a *Analyzer) analyzeMethodDecl(method *ast.FunctionDecl, classType *types.ClassType) {
	// Convert parameter types
	paramTypes := make([]types.Type, 0, len(method.Parameters))
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

	// Create function type and add to class methods
	funcType := types.NewFunctionType(paramTypes, returnType)
	classType.Methods[method.Name.Value] = funcType
	if method.IsConstructor {
		classType.Constructors[method.Name.Value] = funcType
	}
	classType.ClassMethodFlags[method.Name.Value] = method.IsClassMethod

	// Store method visibility (Task 7.63f)
	// Only set visibility if this is the first time we're seeing this method (declaration in class body)
	// Method implementations outside the class shouldn't overwrite the visibility
	if _, exists := classType.MethodVisibility[method.Name.Value]; !exists {
		classType.MethodVisibility[method.Name.Value] = int(method.Visibility)
	}

	// Store virtual/override/abstract flags (Task 7.64, 7.65)
	classType.VirtualMethods[method.Name.Value] = method.IsVirtual
	classType.OverrideMethods[method.Name.Value] = method.IsOverride
	classType.AbstractMethods[method.Name.Value] = method.IsAbstract

	// Analyze method body in new scope
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()

	// Task 7.61: Check if this is a class method (static method)
	if method.IsClassMethod {
		// Class methods (static methods) do NOT have access to Self or instance fields
		// They can only access class variables (static fields)
		// Do NOT add Self to scope
		// Do NOT add instance fields to scope

		// Add class variables to scope (Task 7.62)
		for classVarName, classVarType := range classType.ClassVars {
			a.symbols.Define(classVarName, classVarType)
		}

		// If class has parent, add parent class variables too
		if classType.Parent != nil {
			a.addParentClassVarsToScope(classType.Parent)
		}
	} else {
		// Instance method - add Self reference to method scope (Task 7.56)
		a.symbols.Define("Self", classType)

		// Add class fields to method scope (Task 7.56)
		for fieldName, fieldType := range classType.Fields {
			a.symbols.Define(fieldName, fieldType)
		}

		// Add class variables to method scope (Task 7.62)
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

	// Task 7.64e-h: Validate virtual/override usage
	a.validateVirtualOverride(method, classType, funcType)

	// Analyze method body
	if method.Body != nil {
		a.analyzeBlock(method.Body)
	}
}

// validateVirtualOverride validates virtual/override method declarations (Task 7.64e-h)
func (a *Analyzer) validateVirtualOverride(method *ast.FunctionDecl, classType *types.ClassType, methodType *types.FunctionType) {
	methodName := method.Name.Value

	// Task 7.64f: If method is marked override, validate parent has virtual method
	if method.IsOverride {
		if classType.Parent == nil {
			a.addError("method '%s' marked as override, but class has no parent", methodName)
			return
		}

		// Find method in parent class hierarchy
		parentMethod := a.findMethodInParent(methodName, classType.Parent)
		if parentMethod == nil {
			a.addError("method '%s' marked as override, but no such method exists in parent class", methodName)
			return
		}

		// Task 7.64g: Check that parent method is virtual or override
		if !a.isMethodVirtualOrOverride(methodName, classType.Parent) {
			a.addError("method '%s' marked as override, but parent method is not virtual", methodName)
			return
		}

		// Task 7.64f: Ensure signatures match
		if !a.methodSignaturesMatch(methodType, parentMethod) {
			a.addError("method '%s' override signature does not match parent method signature", methodName)
			return
		}
	}

	// Task 7.64h: Warn if redefining virtual method without override keyword
	if !method.IsOverride && !method.IsVirtual && classType.Parent != nil {
		parentMethod := a.findMethodInParent(methodName, classType.Parent)
		if parentMethod != nil && a.isMethodVirtualOrOverride(methodName, classType.Parent) {
			a.addError("method '%s' hides virtual parent method; use 'override' keyword", methodName)
		}
	}
}

// checkMethodOverriding checks if overridden methods have compatible signatures (Task 7.59)
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
func (a *Analyzer) checkVisibility(memberClass *types.ClassType, visibility int, memberName, memberType string) bool {
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
	classType, found := a.classes[className]
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

// analyzeMemberAccessExpression analyzes member access (Task 7.58)
func (a *Analyzer) analyzeMemberAccessExpression(expr *ast.MemberAccessExpression) types.Type {
	// Analyze the object expression
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		// Error already reported
		return nil
	}

	// Check if object is a class or record type
	memberName := expr.Member.Value

	// Handle record type field access (Task 8.71)
	if _, ok := objectType.(*types.RecordType); ok {
		return a.analyzeRecordFieldAccess(expr.Object, memberName)
	}

	// Handle class type
	classType, ok := objectType.(*types.ClassType)
	if !ok {
		a.addError("member access requires class or record type, got %s at %s",
			objectType.String(), expr.Token.Pos.String())
		return nil
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
		return methodType
	}

	// Member not found
	a.addError("class '%s' has no member '%s' at %s",
		classType.Name, memberName, expr.Token.Pos.String())
	return nil
}

// analyzeMethodCallExpression analyzes a method call on an object
func (a *Analyzer) analyzeMethodCallExpression(expr *ast.MethodCallExpression) types.Type {
	// Analyze the object expression
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		// Error already reported
		return nil
	}

	// Check if object is a class type
	classType, ok := objectType.(*types.ClassType)
	if !ok {
		a.addError("method call requires class type, got %s at %s",
			objectType.String(), expr.Token.Pos.String())
		return nil
	}

	methodName := expr.Method.Value

	// Look up method in class (including inherited methods)
	methodType, found := classType.GetMethod(methodName)
	if !found {
		a.addError("class '%s' has no method '%s' at %s",
			classType.Name, methodName, expr.Token.Pos.String())
		return nil
	}

	// Check method visibility (Task 7.63k)
	methodOwner := a.getMethodOwner(classType, methodName)
	if methodOwner != nil {
		visibility, hasVisibility := methodOwner.MethodVisibility[methodName]
		if hasVisibility && !a.checkVisibility(methodOwner, visibility, methodName, "method") {
			visibilityStr := ast.Visibility(visibility).String()
			if methodOwner.HasConstructor(methodName) {
				a.addError("cannot access %s constructor '%s' of class '%s' at %s",
					visibilityStr, methodName, methodOwner.Name, expr.Token.Pos.String())
				return classType
			}
			a.addError("cannot call %s method '%s' of class '%s' at %s",
				visibilityStr, methodName, methodOwner.Name, expr.Token.Pos.String())
			return methodType.ReturnType
		}
	}

	// Check argument count
	if len(expr.Arguments) != len(methodType.Parameters) {
		a.addError("method '%s' of class '%s' expects %d arguments, got %d at %s",
			methodName, classType.Name, len(methodType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return methodType.ReturnType
	}

	// Check argument types
	for i, arg := range expr.Arguments {
		argType := a.analyzeExpression(arg)
		expectedType := methodType.Parameters[i]
		if argType != nil && !a.canAssign(argType, expectedType) {
			a.addError("argument %d to method '%s' of class '%s' has type %s, expected %s at %s",
				i+1, methodName, classType.Name, argType.String(), expectedType.String(),
				expr.Token.Pos.String())
		}
	}

	if classType.HasConstructor(methodName) {
		return classType
	}
	return methodType.ReturnType
}

// ============================================================================
// Abstract Class/Method Validation (Task 7.65)
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
