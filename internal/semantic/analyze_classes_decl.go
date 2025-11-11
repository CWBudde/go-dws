package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Class Declaration Analysis Functions
// ============================================================================

// analyzeClassDecl analyzes a class declaration
func (a *Analyzer) analyzeClassDecl(decl *ast.ClassDecl) {
	className := decl.Name.Value

	// Task 9.11: Detect if this is a forward declaration
	// A forward declaration has no body - the slices are nil (not initialized)
	// An empty class has initialized but empty slices
	isForwardDecl := (decl.Fields == nil &&
		decl.Methods == nil &&
		decl.Properties == nil &&
		decl.Operators == nil &&
		decl.Constants == nil)

	// Check if class is already declared
	// Task 9.285: Use lowercase for case-insensitive lookup
	existingClass, exists := a.classes[strings.ToLower(className)]
	resolvingForwardDecl := false
	mergingPartialClass := false
	if exists {
		// Task 9.13: Handle partial class merging
		if existingClass.IsPartial && decl.IsPartial {
			// Both are partial - merge them
			mergingPartialClass = true

			// Validate that parent class matches if specified in both declarations
			if decl.Parent != nil && existingClass.Parent != nil {
				if !strings.EqualFold(decl.Parent.Value, existingClass.Parent.Name) {
					a.addError("partial class '%s' has conflicting parent classes at %s",
						className, decl.Token.Pos.String())
					return
				}
			}
		} else if existingClass.IsPartial && !decl.IsPartial && !isForwardDecl {
			// Previous was partial, this is non-partial - issue a hint and finalize
			a.addHint("Previous declaration of class was \"partial\" at %s", decl.Token.Pos.String())
			mergingPartialClass = true
		} else if !existingClass.IsPartial && decl.IsPartial {
			// Previous was non-partial, this is partial - error
			a.addError("class '%s' already declared as non-partial at %s", className, decl.Token.Pos.String())
			return
		} else if existingClass.IsForward && !isForwardDecl {
			// Task 9.11: Handle forward declaration resolution
			// This is the full implementation of a forward-declared class
			// Validate that parent class matches between forward declaration and full implementation
			var fullImplParent *types.ClassType
			if decl.Parent != nil {
				parentName := decl.Parent.Value
				var found bool
				fullImplParent, found = a.classes[strings.ToLower(parentName)]
				if !found {
					a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
					return
				}
			}

			// Compare parent classes
			// Rule: If forward declaration specified a parent, implementation must match it
			// If forward declaration had no parent, implementation can specify any parent (or none)
			if existingClass.Parent != nil {
				// Forward declaration specified a parent - implementation must match
				if fullImplParent == nil {
					a.addError("class '%s' forward declared with parent '%s', but implementation has no parent at %s",
						className, existingClass.Parent.Name, decl.Token.Pos.String())
					return
				} else if existingClass.Parent.Name != fullImplParent.Name {
					a.addError("class '%s' forward declared with parent '%s', but implementation specifies different parent '%s' at %s",
						className, existingClass.Parent.Name, fullImplParent.Name, decl.Token.Pos.String())
					return
				}
			}
			// If forward declaration had no parent, implementation can specify any parent - no validation needed
			// Parent classes are compatible - mark that we're resolving a forward declaration
			resolvingForwardDecl = true
		} else if existingClass.IsForward && isForwardDecl {
			// Duplicate forward declaration
			a.addError("class '%s' already forward declared at %s", className, decl.Token.Pos.String())
			return
		} else {
			// Class already fully declared
			a.addError("class '%s' already declared at %s", className, decl.Token.Pos.String())
			return
		}
	}

	// Task 9.11: If this is a forward declaration, create a minimal class type
	if isForwardDecl {
		// For forward declarations, we still need to resolve the parent if specified
		// so that later uses of the class can access parent members
		var parentClass *types.ClassType
		if decl.Parent != nil {
			parentName := decl.Parent.Value
			var found bool
			parentClass, found = a.classes[strings.ToLower(parentName)]
			if !found {
				a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
				return
			}
		}

		// Create minimal class type for forward declaration
		classType := types.NewClassType(className, parentClass)
		classType.IsForward = true
		classType.IsAbstract = decl.IsAbstract
		classType.IsExternal = decl.IsExternal
		classType.ExternalName = decl.ExternalName

		// Register the forward declaration
		a.classes[strings.ToLower(className)] = classType
		return
	}

	// Resolve parent class if specified (or reuse from forward declaration or partial)
	var parentClass *types.ClassType
	var classType *types.ClassType

	if resolvingForwardDecl || mergingPartialClass {
		// Reuse the existing class instance
		classType = existingClass
		parentClass = classType.Parent

		// Update parent if specified in this partial declaration and wasn't set before
		if decl.Parent != nil && parentClass == nil {
			parentName := decl.Parent.Value
			var found bool
			parentClass, found = a.classes[strings.ToLower(parentName)]
			if !found {
				a.addError("parent class '%s' not found at %s", parentName, decl.Token.Pos.String())
				return
			}
			classType.Parent = parentClass
		}

		// Handle implicit TObject parent if needed
		if parentClass == nil && !strings.EqualFold(className, "TObject") && !decl.IsExternal {
			parentClass = a.classes["tobject"]
			if parentClass == nil {
				a.addError("implicit parent class 'TObject' not found at %s", decl.Token.Pos.String())
				return
			}
			classType.Parent = parentClass
		}
	} else {
		// Not resolving a forward declaration or partial - resolve parent and create new class
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
		classType = types.NewClassType(className, parentClass)
	}

	// Update class flags
	classType.IsForward = false // No longer a forward declaration

	// Task 9.13: Update IsPartial flag
	// If this declaration is partial, keep IsPartial=true
	// If this declaration is non-partial, set IsPartial=false (finalize the class)
	if decl.IsPartial {
		classType.IsPartial = true
	} else if !isForwardDecl {
		// Non-partial, non-forward declaration finalizes any partial class
		classType.IsPartial = false
	}

	classType.IsAbstract = decl.IsAbstract || classType.IsAbstract // Preserve abstract if already set
	classType.IsExternal = decl.IsExternal || classType.IsExternal // Preserve external if already set
	if decl.ExternalName != "" {
		classType.ExternalName = decl.ExternalName
	}

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
		// Task 9.285: Normalize field names to lowercase for case-insensitive lookup
		fieldName := strings.ToLower(field.Name.Value)

		// Check if this is a class variable (static field)
		if field.IsClassVar {
			// Check for duplicate class variable names
			// When merging partial classes, check if already exists in ClassType
			_, existsInClass := classType.ClassVars[fieldName]
			if existsInClass {
				a.addError("duplicate class variable '%s' in class '%s' at %s",
					fieldName, className, field.Token.Pos.String())
				continue
			}
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
			// When merging partial classes, check if already exists in ClassType
			_, existsInClass := classType.Fields[fieldName]
			if existsInClass {
				a.addError("duplicate field '%s' in class '%s' at %s",
					fieldName, className, field.Token.Pos.String())
				continue
			}
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
		// When merging partial classes, check if already exists in ClassType
		_, existsInClass := classType.Constants[constantName]
		if existsInClass {
			a.addError("duplicate constant '%s' in class '%s' at %s",
				constantName, className, constant.Token.Pos.String())
			continue
		}
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

	// Task 9.19: If any constructor has the 'overload' directive, synthesize implicit parameterless constructor
	// In DWScript, when a constructor is marked with 'overload', the compiler implicitly provides
	// a parameterless constructor if one doesn't already exist
	a.synthesizeImplicitParameterlessConstructor(classType)

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

	// Track if this was originally marked as constructor by parser (before auto-detection)
	wasExplicitConstructor := method.IsConstructor

	// Auto-detect constructors: methods named "Create" that return the class type
	// This handles inline constructor declarations like: function Create(...): TClass;
	if !method.IsConstructor && strings.EqualFold(method.Name.Value, "Create") && method.ReturnType != nil {
		returnTypeName := method.ReturnType.Name
		if strings.EqualFold(returnTypeName, classType.Name) {
			method.IsConstructor = true
		}
	}

	// Task 9.17: Validate constructors don't have explicit return types
	if method.IsConstructor && method.ReturnType != nil {
		if wasExplicitConstructor {
			// Explicit constructors (using 'constructor' keyword) cannot have return types
			a.addError("constructor '%s' cannot have an explicit return type at %s",
				method.Name.Value, method.Token.Pos.String())
			return
		}
		// Auto-detected constructors (function Create: TClass) must have matching return type
		returnTypeName := method.ReturnType.Name
		if !strings.EqualFold(returnTypeName, classType.Name) {
			a.addError("constructor '%s' must return '%s', not '%s' at %s",
				method.Name.Value, classType.Name, returnTypeName, method.Token.Pos.String())
			return
		}
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

	// Track if this is an implementation for an existing forward declaration
	isImplementationOfForward := false

	for _, existing := range existingOverloads {
		// Task 9.63: Check if signatures are identical (duplicate) - use DWScript error format
		if a.methodSignaturesMatch(funcType, existing.Signature) {
			// Task 9.60: Check if this is a forward + implementation pair (like in symbol_table.go:211-222)
			if existing.IsForwarded && !isForward {
				// Implementation following forward declaration
				// Update the existing forward declaration instead of adding a new overload
				existing.IsForwarded = false
				existing.Signature = funcType

				// Task 9.6: Do NOT update virtual/override/abstract flags when matching implementation to declaration
				// The implementation doesn't have these keywords - they're only in the declaration
				// So preserve the declaration's flags instead of overwriting with implementation's false values

				// Mark that we found the forward declaration and updated it
				isImplementationOfForward = true
				break // Exit overload loop - method body will be analyzed below
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

	// Only add a new overload if this isn't an implementation of an existing forward declaration
	if !isImplementationOfForward {
		if method.IsConstructor {
			classType.AddConstructorOverload(method.Name.Value, methodInfo)
		} else {
			classType.AddMethodOverload(method.Name.Value, methodInfo)
		}

		// Store method metadata in legacy maps for backward compatibility
		// Only update metadata for new declarations, not implementations
		// (implementations don't have override/virtual keywords, those are only in declarations)
		classType.ClassMethodFlags[method.Name.Value] = method.IsClassMethod
		classType.VirtualMethods[method.Name.Value] = method.IsVirtual
		classType.OverrideMethods[method.Name.Value] = method.IsOverride
		classType.AbstractMethods[method.Name.Value] = method.IsAbstract
	}

	// Task 9.280: Mark method as forward if it has no body (declaration without implementation)
	// Methods declared in class body without implementation are implicitly forward
	if method.Body == nil {
		classType.ForwardedMethods[method.Name.Value] = true
	}

	// Store method visibility
	// Only set visibility if this is the first time we're seeing this method (declaration in class body)
	// Method implementations outside the class shouldn't overwrite the visibility
	// Task 9.16.1: Use lowercase key for case-insensitive lookups
	methodKey := strings.ToLower(method.Name.Value)
	if _, exists := classType.MethodVisibility[methodKey]; !exists {
		classType.MethodVisibility[methodKey] = int(method.Visibility)
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

// Suppress unused import error for fmt
var _ = fmt.Sprint
