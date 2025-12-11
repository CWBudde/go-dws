package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Class Expression Analysis Functions
// ============================================================================

// analyzeNewExpression analyzes object creation (new TClass(args) or TClass.Create(args))
// with constructor overload resolution and visibility checking.
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
	classType := a.getClassType(className)
	if classType == nil {
		// Check if it's a record type with static method call (e.g., TRecord.Create())
		if recordType := a.getRecordType(className); recordType != nil {
			return a.analyzeRecordStaticMethodCallFromNew(expr, recordType)
		}
		a.addError("%s", errors.FormatUnknownName(className, expr.Token.Pos.Line, expr.Token.Pos.Column))
		return nil
	}

	// Check if trying to instantiate an abstract class
	if classType.IsAbstract {
		a.addError("%s", errors.FormatAbstractClassError(expr.Token.Pos.Line, expr.Token.Pos.Column))
		return nil
	}

	// Check for unimplemented abstract methods (inherited but not overridden)
	unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
	if len(unimplementedMethods) > 0 {
		a.addError("%s", errors.FormatAbstractClassError(expr.Token.Pos.Line, expr.Token.Pos.Column))
		return nil
	}

	// Get all constructor overloads
	constructorName := a.getDefaultConstructorName(classType)
	constructorOverloads := a.getMethodOverloadsInHierarchy(constructorName, classType)

	if len(constructorOverloads) == 0 {
		// No explicit constructor - use implicit default constructor (no arguments allowed)
		if len(expr.Arguments) > 0 {
			a.addError("class '%s' has no constructor, cannot pass arguments at %s",
				className, expr.Token.Pos.String())
		}
		return classType
	}

	// Filter out implicit parameterless constructor if there are explicit constructors
	validConstructors := make([]*types.MethodInfo, 0, len(constructorOverloads))
	hasExplicitConstructor := false
	for _, ctor := range constructorOverloads {
		if ctor.Visibility != 0 || len(ctor.Signature.Parameters) > 0 {
			validConstructors = append(validConstructors, ctor)
			hasExplicitConstructor = true
		}
	}
	if !hasExplicitConstructor {
		validConstructors = constructorOverloads
	}

	// Select constructor based on argument count first
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
		if len(validConstructors) > 0 {
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

	// Check constructor visibility
	var ownerClass *types.ClassType
	for class := classType; class != nil; class = class.Parent {
		if class.HasConstructor(constructorName) {
			ownerClass = class
			break
		}
	}
	if ownerClass != nil && selectedConstructor != nil {
		visibility := selectedConstructor.Visibility
		if !a.checkVisibility(ownerClass, visibility, constructorName, "constructor") {
			visibilityStr := ast.Visibility(visibility).String()
			a.addError("cannot access %s constructor '%s' of class '%s' at %s",
				visibilityStr, constructorName, ownerClass.Name, expr.Token.Pos.String())
			return classType
		}
	}

	// Validate argument types
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

// analyzeMemberAccessExpression analyzes member access on classes, records, interfaces, and helpers.
func (a *Analyzer) analyzeMemberAccessExpression(expr *ast.MemberAccessExpression) types.Type {
	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		return nil
	}
	memberName := ident.Normalize(expr.Member.Value)

	// Resolve type aliases to get the underlying type
	objectTypeResolved := types.GetUnderlyingType(objectType)

	// Handle record type (static methods or instance fields/methods)
	if recordType, ok := objectTypeResolved.(*types.RecordType); ok {
		if recordType.HasClassMethod(memberName) {
			classMethod := recordType.GetClassMethod(memberName)
			if classMethod != nil {
				return classMethod
			}
		}
		return a.analyzeRecordFieldAccess(expr.Object, memberName)
	}

	// Handle interface type method access
	if ifaceType, ok := objectTypeResolved.(*types.InterfaceType); ok {
		allMethods := types.GetAllInterfaceMethods(ifaceType)
		if methodType, hasMethod := allMethods[memberName]; hasMethod {
			return methodType
		}
		a.addError("interface '%s' has no method '%s' at %s",
			ifaceType.Name, expr.Member.Value, expr.Token.Pos.String())
		return nil
	}

	// Handle metaclass type (class of T) - convert to base ClassType for constructor/class member access
	if metaclassType, ok := objectTypeResolved.(*types.ClassOfType); ok {
		baseClass := metaclassType.ClassType
		if baseClass != nil {
			objectTypeResolved = baseClass
		}
	}
	// If not a class type, check for helpers (String, Integer, Enum types, etc.)
	classType, ok := objectTypeResolved.(*types.ClassType)
	if !ok {
		// Handle .Value property on enum types
		if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
			if memberName == "value" {
				return types.INTEGER
			}
		}

		// Check helpers (prefer properties before methods for property-style access)
		_, helperProp := a.hasHelperProperty(objectType, memberName)
		if helperProp != nil {
			return helperProp.Type
		}

		_, helperMethod := a.hasHelperMethod(objectType, memberName)
		if helperMethod != nil {
			if len(helperMethod.Parameters) == 0 {
				return helperMethod.ReturnType
			}
			return helperMethod
		}

		_, helperClassVar := a.hasHelperClassVar(objectType, memberName)
		if helperClassVar != nil {
			return helperClassVar
		}

		_, helperConst := a.hasHelperClassConst(objectType, memberName)
		if helperConst != nil {
			if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
				return objectType
			}
			return objectType
		}

		a.addError("member access on type %s requires a helper, got no helper with member '%s' at %s",
			objectType.String(), expr.Member.Value, expr.Token.Pos.String())
		return nil
	}

	// Handle built-in properties on all objects (inherited from TObject)
	if memberName == "classname" {
		if expr.Member.Value != "ClassName" {
			pos := expr.Token.Pos
			pos.Column++
			a.addCaseMismatchHint(expr.Member.Value, "ClassName", pos)
		}
		return types.STRING
	}
	if memberName == "classtype" {
		return types.NewClassOfType(classType)
	}

	// Look up field (including inherited fields)
	fieldType, found := classType.GetField(memberName)
	if found {
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

	// Look up class variable (including inherited class vars)
	classVarType, foundClassVar := classType.GetClassVar(memberName)
	if foundClassVar {
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

	// Look up property (including inherited properties)
	propInfo, propFound := classType.GetProperty(memberName)
	if propFound {
		return propInfo.Type
	}

	// Look up constructor (constructors are stored separately)
	constructorOverloads := classType.GetConstructorOverloads(memberName)
	if len(constructorOverloads) > 0 {
		if memberName == "create" && expr.Member.Value != "Create" {
			pos := expr.Token.Pos
			pos.Column++
			a.addCaseMismatchHint(expr.Member.Value, "Create", pos)
		}
		// Check if parameterless (auto-invoked when accessed without parentheses)
		hasParameterless := false
		for _, ctor := range constructorOverloads {
			if len(ctor.Signature.Parameters) == 0 {
				hasParameterless = true
				break
			}
		}
		if hasParameterless {
			return classType
		}
		// Constructor has parameters - return method pointer for deferred invocation
		if len(constructorOverloads) == 1 {
			return types.NewMethodPointerType(constructorOverloads[0].Signature.Parameters, classType)
		}
		return types.NewMethodPointerType([]types.Type{}, classType)
	}

	// Look up method (including inherited methods)
	methodType, found := classType.GetMethod(memberName)
	if found {
		if memberName == "free" && expr.Member.Value != "Free" {
			pos := expr.Token.Pos
			pos.Column++
			a.addCaseMismatchHint(expr.Member.Value, "Free", pos)
		}
		methodOwner := a.getMethodOwner(classType, memberName)
		if methodOwner != nil {
			visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(memberName)]
			if hasVisibility && !a.checkVisibility(methodOwner, visibility, memberName, "method") {
				visibilityStr := ast.Visibility(visibility).String()
				a.addError("cannot call %s method '%s' of class '%s' at %s",
					visibilityStr, expr.Member.Value, methodOwner.Name, expr.Token.Pos.String())
				return nil
			}
		}
		// Parameterless methods are auto-invoked when accessed without parentheses
		if len(methodType.Parameters) == 0 {
			if methodType.ReturnType == nil {
				return types.VOID
			}
			return methodType.ReturnType
		}
		// Methods with parameters return method pointer for deferred invocation
		return types.NewMethodPointerType(methodType.Parameters, methodType.ReturnType)
	}

	// Check helpers for methods
	_, helperMethod := a.hasHelperMethod(objectType, memberName)
	if helperMethod != nil {
		if len(helperMethod.Parameters) == 0 {
			return helperMethod.ReturnType
		}
		return helperMethod
	}

	// Check helpers for properties
	_, helperProp := a.hasHelperProperty(objectType, memberName)
	if helperProp != nil {
		return helperProp.Type
	}

	// Check for class constants (including inherited constants)
	if constType := a.findClassConstantWithVisibility(classType, memberName, expr.Token.Pos.String()); constType != nil {
		return constType
	}

	a.addError("class '%s' has no member '%s' at %s",
		classType.Name, expr.Member.Value, expr.Token.Pos.String())
	return nil
}

// analyzeRecordStaticMethodCallFromNew handles record static method calls that use NewExpression syntax.
// This occurs when the parser encounters TRecord.Create(...) which generates a NewExpression.
func (a *Analyzer) analyzeRecordStaticMethodCallFromNew(expr *ast.NewExpression, recordType *types.RecordType) types.Type {
	methodName := "Create"
	lowerMethodName := ident.Normalize(methodName)

	overloads := recordType.GetClassMethodOverloads(lowerMethodName)
	if len(overloads) == 0 {
		a.addError("record type '%s' has no class method '%s' at %s",
			recordType.Name, methodName, expr.Token.Pos.String())
		return nil
	}

	argTypes := make([]types.Type, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		argType := a.analyzeExpression(arg)
		if argType == nil {
			return nil
		}
		argTypes[i] = argType
	}
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
