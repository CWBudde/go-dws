package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Class Expression Analysis Functions
// ============================================================================

// analyzeNewExpression analyzes object creation
// Handles both:
//   - new TClass(args)
//   - TClass.Create(args)
//
// Task 9.18: NewExpression semantic validation with constructor overload resolution
func (a *Analyzer) analyzeNewExpression(expr *ast.NewExpression) types.Type {
	className := expr.ClassName.Value

	// If we're inside a class with nested types, prefer the nested type name
	if a.currentNestedTypes != nil {
		if qualified, ok := a.currentNestedTypes[ident.Normalize(className)]; ok {
			className = qualified
		}
	} else if a.currentClass != nil {
		if aliases, ok := a.nestedTypeAliases[ident.Normalize(a.currentClass.Name)]; ok {
			if qualified, ok := aliases[ident.Normalize(className)]; ok {
				className = qualified
			}
		}
	}

	// Look up class in registry
	// Task 6.1.1.3: Use TypeRegistry for unified type lookup
	classType := a.getClassType(className)
	if classType == nil {
		// Check if it's a record type (records don't use 'new', but might use TRecord.Create() syntax)
		if recordType := a.getRecordType(className); recordType != nil {
			// This is actually a record static method call (e.g., TTest.Create(...))
			// Treat it as a static method call
			return a.analyzeRecordStaticMethodCallFromNew(expr, recordType)
		}
		a.addError("undefined class '%s' at %s", className, expr.Token.Pos.String())
		return nil
	}

	// Task 9.18: Check if trying to instantiate an abstract class
	if classType.IsAbstract {
		a.addError("Trying to create an instance of an abstract class at [line: %d, column: %d]",
			expr.Token.Pos.Line, expr.Token.Pos.Column)
		return nil
	}

	// Task 9.12.3: Check if class has unimplemented abstract methods
	// Even if the class itself is not marked abstract, it cannot be instantiated
	// if it inherits abstract methods that haven't been implemented
	unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
	if len(unimplementedMethods) > 0 {
		a.addError("Trying to create an instance of an abstract class at [line: %d, column: %d]",
			expr.Token.Pos.Line, expr.Token.Pos.Column)
		return nil
	}

	// Task 9.13-9.16: Get all constructor overloads
	// Task 9.3: Use default constructor if specified, otherwise fall back to "Create"
	constructorName := a.getDefaultConstructorName(classType)
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
	memberName := ident.Normalize(expr.Member.Value)

	// Resolve type aliases to get the underlying type
	// This allows member access on type alias variables like TBaseClass
	objectTypeResolved := types.GetUnderlyingType(objectType)

	// Handle record type - check for both type-level (static) and instance-level access
	if recordType, ok := objectTypeResolved.(*types.RecordType); ok {
		// Check if this is a class method (static method) access on the record TYPE itself
		// (e.g., TTest.Create or TTest.Sum)
		if recordType.HasClassMethod(memberName) {
			classMethod := recordType.GetClassMethod(memberName)
			if classMethod != nil {
				return classMethod
			}
		}

		// Otherwise, treat as instance field/method access
		return a.analyzeRecordFieldAccess(expr.Object, memberName)
	}

	// Task 9.16.2: Handle interface type method access
	if ifaceType, ok := objectTypeResolved.(*types.InterfaceType); ok {
		// Check if the interface has this method
		allMethods := types.GetAllInterfaceMethods(ifaceType)
		if methodType, hasMethod := allMethods[memberName]; hasMethod {
			return methodType
		}
		a.addError("interface '%s' has no method '%s' at %s",
			ifaceType.Name, expr.Member.Value, expr.Token.Pos.String())
		return nil
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
		// Task 9.15.10: Handle .Value property on enum types
		if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
			if memberName == "value" {
				// .Value returns the ordinal value as Integer
				return types.INTEGER
			}
			// Continue to check helpers for other properties/methods on enums
		}

		// Task 9.83: For non-class/record types (like String, Integer), check helpers
		// Prefer helper properties before methods so that property-style access
		// (e.g., i.ToString) resolves correctly when parentheses are omitted.
		_, helperProp := a.hasHelperProperty(objectType, memberName)
		if helperProp != nil {
			return helperProp.Type
		}

		_, helperMethod := a.hasHelperMethod(objectType, memberName)
		if helperMethod != nil {
			// Task 9.8.5: Auto-invoke parameterless helper methods when accessed without ()
			// This allows arr.Pop to work the same as arr.Pop()
			if len(helperMethod.Parameters) == 0 {
				// Parameterless method - auto-invoke and return the return type
				return helperMethod.ReturnType
			}
			// Method has parameters - return the method type for deferred invocation
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
			objectType.String(), expr.Member.Value, expr.Token.Pos.String())
		return nil
	}

	// Handle built-in properties/methods available on all objects (inherited from TObject)
	if memberName == "classname" {
		if expr.Member.Value != "ClassName" {
			a.addHint("\"%s\" does not match case of declaration (\"ClassName\") [line: %d, column: %d]",
				expr.Member.Value, expr.Token.Pos.Line, expr.Token.Pos.Column+1)
		}
		// ClassName returns String
		return types.STRING
	}
	if memberName == "classtype" {
		// ClassType returns the metaclass (class of T) for the object's runtime type
		return types.NewClassOfType(classType)
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
					visibilityStr, expr.Member.Value, fieldOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		return fieldType
	}

	// Task 9.5.3: Look up class variable in class (including inherited class vars)
	classVarType, foundClassVar := classType.GetClassVar(memberName)
	if foundClassVar {
		// Check class variable visibility
		classVarOwner := a.getClassVarOwner(classType, memberName)
		if classVarOwner != nil {
			visibility, hasVisibility := classVarOwner.ClassVarVisibility[memberName]
			if hasVisibility && !a.checkVisibility(classVarOwner, visibility, memberName, "class variable") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot access %s class variable '%s' of class '%s' at %s",
					visibilityStr, expr.Member.Value, classVarOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		return classVarType
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
		if memberName == "create" && expr.Member.Value != "Create" {
			a.addHint("\"%s\" does not match case of declaration (\"Create\") [line: %d, column: %d]",
				expr.Member.Value, expr.Token.Pos.Line, expr.Token.Pos.Column+1)
		}
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
		if memberName == "free" && expr.Member.Value != "Free" {
			a.addHint("\"%s\" does not match case of declaration (\"Free\") [line: %d, column: %d]",
				expr.Member.Value, expr.Token.Pos.Line, expr.Token.Pos.Column+1)
		}
		// Check method visibility - Task 9.16.1
		methodOwner := a.getMethodOwner(classType, memberName)
		if methodOwner != nil {
			// Use normalized key for case-insensitive lookup
			visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(memberName)]
			if hasVisibility && !a.checkVisibility(methodOwner, visibility, memberName, "method") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot call %s method '%s' of class '%s' at %s",
					visibilityStr, expr.Member.Value, methodOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		// In DWScript/Pascal, parameterless methods can be called without parentheses
		// When referenced via member access, they should be treated as implicit calls
		if len(methodType.Parameters) == 0 {
			// Implicit call - return the method's return type
			if methodType.ReturnType == nil {
				// Procedure (no return value)
				return types.VOID
			}
			return methodType.ReturnType
		}

		// Methods with parameters cannot be called without parentheses
		// Return a method pointer type for deferred invocation
		return types.NewMethodPointerType(methodType.Parameters, methodType.ReturnType)
	}

	// Task 9.83: Check helpers for methods
	// If not found in class, check if any helpers extend this type
	_, helperMethod := a.hasHelperMethod(objectType, memberName)
	if helperMethod != nil {
		// Task 9.8.5: Auto-invoke parameterless helper methods when accessed without ()
		// This allows arr.Pop to work the same as arr.Pop()
		if len(helperMethod.Parameters) == 0 {
			// Parameterless method - auto-invoke and return the return type
			return helperMethod.ReturnType
		}
		// Method has parameters - return the method type for deferred invocation
		return helperMethod
	}

	// Task 9.83: Check helpers for properties
	_, helperProp := a.hasHelperProperty(objectType, memberName)
	if helperProp != nil {
		return helperProp.Type
	}

	// Task 9.22: Check for class constants (with inheritance support)
	// Task 9.2: Use case-insensitive comparison for constant lookup
	if constType := a.findClassConstantWithVisibility(classType, memberName, expr.Token.Pos.String()); constType != nil {
		return constType
	}

	// Member not found
	a.addError("class '%s' has no member '%s' at %s",
		classType.Name, expr.Member.Value, expr.Token.Pos.String())
	return nil
}

// analyzeRecordStaticMethodCallFromNew handles NewExpression when it's actually a record static method call
// This happens when the parser generates a NewExpression for TRecord.Create(...) syntax
func (a *Analyzer) analyzeRecordStaticMethodCallFromNew(expr *ast.NewExpression, recordType *types.RecordType) types.Type {
	// Assume "Create" as the method name (standard DWScript pattern)
	methodName := "Create"
	lowerMethodName := ident.Normalize(methodName)

	// Look up class method overloads
	overloads := recordType.GetClassMethodOverloads(lowerMethodName)
	if len(overloads) == 0 {
		a.addError("record type '%s' has no class method '%s' at %s",
			recordType.Name, methodName, expr.Token.Pos.String())
		return nil
	}

	// Resolve overload based on arguments
	argTypes := make([]types.Type, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		argType := a.analyzeExpression(arg)
		if argType == nil {
			return nil
		}
		argTypes[i] = argType
	}

	// Find matching overload
	candidates := make([]*Symbol, len(overloads))
	for i, overload := range overloads {
		candidates[i] = &Symbol{
			Type: overload.Signature,
		}
	}

	selected, err := ResolveOverload(candidates, argTypes)
	if err != nil {
		a.addError("no matching overload for '%s.%s' with %d arguments at %s",
			recordType.Name, methodName, len(argTypes), expr.Token.Pos.String())
		return nil
	}

	funcType := selected.Type.(*types.FunctionType)

	// Validate argument types
	for i, arg := range expr.Arguments {
		if i >= len(funcType.Parameters) {
			break
		}

		paramType := funcType.Parameters[i]
		argType := a.analyzeExpressionWithExpectedType(arg, paramType)
		if argType != nil && !a.canAssign(argType, paramType) {
			a.addError("argument %d to '%s.%s' has type %s, expected %s at %s",
				i+1, recordType.Name, methodName, argType.String(), paramType.String(),
				expr.Token.Pos.String())
		}
	}

	return funcType.ReturnType
}

// getDefaultConstructorName returns the name of the default constructor for a class.
// It checks the class hierarchy for a constructor marked as 'default'.
// Falls back to "Create" if no default constructor is found.
// Task 9.3: Support for default constructors
func (a *Analyzer) getDefaultConstructorName(class *types.ClassType) string {
	// Check current class and parents for default constructor
	for current := class; current != nil; current = current.Parent {
		if current.DefaultConstructor != "" {
			return current.DefaultConstructor
		}
	}
	// No default constructor found, use "Create" as fallback
	return "Create"
}
