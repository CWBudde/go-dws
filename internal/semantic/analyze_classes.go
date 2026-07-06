package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
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
		a.addStructuredError(NewUnknownNameError(expr.Token.Pos, className))
		return nil
	}
	a.warnDeprecatedClassUsage(classType, expr.Token.Pos)

	// Case-mismatch hints for the class name and (for the "TClass.Create(...)"
	// sugar) the constructor name against their declarations.
	if expr.ClassName != nil {
		if classType.Name != expr.ClassName.Value && ident.Equal(classType.Name, expr.ClassName.Value) {
			a.addCaseMismatchHint(expr.ClassName.Value, classType.Name, expr.ClassName.Token.Pos)
		}
		if !ident.Equal(expr.Token.Literal, "new") {
			// The parser folds "TClass.Create(args)" into a NewExpression only
			// when the member is spelled exactly "Create"; recover its position
			// from the class name token (same line, right after the dot).
			if declared := a.declaredMethodName(classType, "Create"); declared != "" && declared != "Create" {
				pos := expr.ClassName.Token.Pos
				pos.Column += len(expr.ClassName.Value) + 1
				a.addCaseMismatchHint("Create", declared, pos)
			}
		}
	}

	if classType.IsStatic {
		a.addStructuredError(NewStaticClassInstantiationError(expr.Pos(), classType.Name))
		return classType
	}

	// Check if trying to instantiate an abstract class
	if classType.IsAbstract {
		a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
		return nil
	}

	// Check for unimplemented abstract methods (inherited but not overridden)
	unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
	if len(unimplementedMethods) > 0 {
		a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
		return nil
	}

	// Get all constructor overloads. The hierarchy lookup also returns same-named
	// class methods (a constructor name can be shared with overloaded class
	// methods); "new" only ever invokes actual constructors, so filter those out.
	constructorName := a.getDefaultConstructorName(classType)
	allOverloads := a.getMethodOverloadsInHierarchy(constructorName, classType)
	constructorOverloads := make([]*types.MethodInfo, 0, len(allOverloads))
	for _, overload := range allOverloads {
		if overload.IsConstructor {
			constructorOverloads = append(constructorOverloads, overload)
		}
	}
	// The "TClass.Create(args)" sugar can also resolve to a same-named class
	// method; the literal "new" syntax only ever invokes constructors.
	if !ident.Equal(expr.Token.Literal, "new") {
		constructorOverloads = allOverloads
	}

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

	// Find constructors with matching argument count (params with default
	// values are optional).
	matchingCountConstructors := make([]*types.MethodInfo, 0)
	for _, ctor := range validConstructors {
		if len(expr.Arguments) > len(ctor.Signature.Parameters) {
			continue
		}
		if len(expr.Arguments) >= requiredParamCount(ctor.Signature) {
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
			argType := a.analyzeOverloadArgument(arg)
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

		var ok bool
		selectedSignature, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("internal error: expected function type for selected constructor, but got %T", selected.Type)
			return classType
		}
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

	// "TClass.Create(args)" resolved to a same-named class method, not a
	// constructor: the expression's type is the method's return type.
	if selectedConstructor != nil && !selectedConstructor.IsConstructor {
		return selectedSignature.ReturnType
	}

	return classType
}

func constructorCallPosition(expr *ast.CallExpression) token.Position {
	if expr == nil {
		return token.Position{}
	}
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok && memberAccess.Member != nil {
		return memberAccess.Member.Token.Pos
	}
	return expr.Pos()
}

// analyzeMemberAccessExpression analyzes member access on classes, records, interfaces, and helpers.
func (a *Analyzer) analyzeMemberAccessExpression(expr *ast.MemberAccessExpression) types.Type {
	if identExpr, ok := expr.Object.(*ast.Identifier); ok {
		switch ident.Normalize(identExpr.Value) {
		case "system", "internal":
			if sym, err := a.ResolveQualifiedSymbol(identExpr.Value, expr.Member.Value); err == nil && sym != nil {
				return sym.Type
			}
			a.addStructuredError(NewUnknownNameError(expr.Member.Token.Pos, identExpr.Value+"."+expr.Member.Value))
			return nil
		case "json":
			// Bare JSON namespace access (JSON.NewObject / JSON.NewArray invoked
			// without parentheses), unless shadowed by a user symbol.
			if a.isJSONNamespace(expr.Object) {
				return jsonNamespaceMemberType(expr.Member.Value)
			}
		}
	}

	objectType := a.analyzeExpression(expr.Object)
	if objectType == nil {
		return nil
	}
	if implicitType := a.getImplicitCallType(expr.Object); implicitType != nil {
		objectType = implicitType
	} else if implicitType := implicitCallReturnTypeFromType(objectType); implicitType != nil {
		objectType = implicitType
	}
	memberName := ident.Normalize(expr.Member.Value)

	// Resolve type aliases to get the underlying type
	objectTypeResolved := types.GetUnderlyingType(objectType)

	// JSONVariant is dynamically browsable: any member access yields another
	// JSONVariant (v.foo, v.foo.bar, ...). Method calls are handled at the call
	// site (analyzeJSONMethodCall).
	if types.IsJSONVariant(objectTypeResolved) {
		return types.JSON_VARIANT
	}

	// Handle record type (static methods or instance fields/methods)
	if recordType, ok := objectTypeResolved.(*types.RecordType); ok {
		if recordType.HasClassMethod(memberName) {
			classMethod := recordType.GetClassMethod(memberName)
			if classMethod != nil {
				return classMethod
			}
		}
		return a.analyzeRecordFieldAccess(expr.Object, expr.Member)
	}

	// Handle interface type method access
	if ifaceType, ok := objectTypeResolved.(*types.InterfaceType); ok {
		allMethods := types.GetAllInterfaceMethods(ifaceType)
		if methodType, hasMethod := allMethods[memberName]; hasMethod {
			return methodType
		}

		// Interface properties resolve to their declared type.
		if propInfo := ifaceType.GetProperty(memberName); propInfo != nil {
			return propInfo.Type
		}

		// Interface helpers can add both instance methods and helper class members.
		if helperMethod := a.hasHelperMethod(objectType, memberName); helperMethod != nil {
			if len(helperMethod.Parameters) == 0 {
				return helperMethod.ReturnType
			}
			return helperMethod
		}
		if helperProp := a.hasHelperProperty(objectType, memberName); helperProp != nil {
			return helperProp.Type
		}
		if _, helperClassVar := a.hasHelperClassVar(objectType, memberName); helperClassVar != nil {
			return helperClassVar
		}
		if _, helperConst := a.hasHelperClassConst(objectType, memberName); helperConst != nil {
			return objectType
		}

		a.addStructuredError(NewAccessibleMemberError(expr.Member.Token.Pos, expr.Member.Value, objectType.String()))
		return nil
	}

	isMetaclass := false
	// Handle metaclass type (class of T) - convert to base ClassType for constructor/class member access
	if metaclassType, ok := objectTypeResolved.(*types.ClassOfType); ok {
		isMetaclass = true
		baseClass := metaclassType.ClassType
		if baseClass != nil {
			objectTypeResolved = baseClass
		}
	}
	// If not a class type, check for helpers (String, Integer, Enum types, etc.)
	classType, ok := objectTypeResolved.(*types.ClassType)
	if !ok {
		if assoc, isAssoc := types.GetUnderlyingType(objectTypeResolved).(*types.AssociativeArrayType); isAssoc {
			if result := a.analyzeAssociativeArrayMemberAccess(expr, assoc); result != nil {
				return result
			}
		}
		if arrayType, isArray := objectTypeResolved.(*types.ArrayType); isArray {
			if result := a.analyzeArrayMemberAccess(expr, arrayType); result != nil {
				return result
			}
		}

		// Handle .Value property on enum types
		if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
			if memberName == "value" {
				return types.INTEGER
			}
		}

		// Check helpers (prefer properties before methods for property-style access)
		helperProp := a.hasHelperProperty(objectType, memberName)
		if helperProp != nil {
			if enumType, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
				a.maybeAddUnnamedEnumElementHint(expr.Object, expr.Member.Token.Pos, enumType, memberName)
			}
			return helperProp.Type
		}

		helperMethod := a.hasHelperMethod(objectType, memberName)
		if helperMethod != nil {
			// Record the receiver's static type so runtime helper dispatch
			// honors alias-specific (strict) helpers over the underlying
			// type's helpers.
			if a.semanticInfo != nil && expr.Member != nil {
				a.semanticInfo.SetType(expr.Member, &ast.TypeAnnotation{
					Token: expr.Member.Token,
					Name:  "__helper_receiver:" + objectType.String(),
				})
			}
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

		if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
			pos := expr.Member.Token.Pos
			a.addStructuredError(NewAccessibleMemberError(pos, expr.Member.Value, objectType.String()))
			return nil
		}

		a.addStructuredError(NewAccessibleMemberError(expr.Member.Token.Pos, expr.Member.Value, objectType.String()))
		return nil
	}

	// Helpers can override built-in TObject members and can also provide class
	// members when accessed through a metaclass value (TType.HelperMember).
	helperLookupType := objectType
	if isMetaclass {
		helperLookupType = objectTypeResolved
	}
	if helperMethod := a.hasHelperMethod(helperLookupType, memberName); helperMethod != nil {
		// Record the receiver's static class so runtime helper dispatch binds
		// helpers by the declared type (strict helper semantics).
		if a.semanticInfo != nil && expr.Member != nil {
			a.semanticInfo.SetType(expr.Member, &ast.TypeAnnotation{
				Token: expr.Member.Token,
				Name:  "__helper_receiver:" + helperLookupType.String(),
			})
		}
		if len(helperMethod.Parameters) == 0 {
			return helperMethod.ReturnType
		}
		return helperMethod
	}
	if helperProp := a.hasHelperProperty(helperLookupType, memberName); helperProp != nil {
		return helperProp.Type
	}
	if _, helperClassVar := a.hasHelperClassVar(helperLookupType, memberName); helperClassVar != nil {
		return helperClassVar
	}
	if _, helperConst := a.hasHelperClassConst(helperLookupType, memberName); helperConst != nil {
		return helperLookupType
	}

	// Handle built-in properties on all objects (inherited from TObject).
	// A user-declared method with the same name hides the builtin; when every
	// user overload needs at least one argument, the parameterless access
	// still resolves to the builtin (see fixture classname_hide_with_default:
	// the user's defaulted overload wins).
	if memberName == "classname" {
		userOverloads := a.getMethodOverloadsInHierarchy(memberName, classType)
		for _, overload := range userOverloads {
			if requiredParamCount(overload.Signature) == 0 {
				if declared := a.declaredMethodName(classType, memberName); declared != "" && expr.Member.Value != declared {
					pos := expr.Token.Pos
					pos.Column++
					a.addCaseMismatchHint(expr.Member.Value, declared, pos)
				}
				return overload.Signature.ReturnType
			}
		}
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
	if memberName == "classparent" {
		// ClassParent yields a metaclass reference to the parent class. Statically we
		// only know the receiver's declared class, so type the result as a metaclass of
		// that parent (or the class itself for a root class), which stays assignable to
		// a `class of` variable used to walk the hierarchy at runtime.
		if classType.Parent != nil {
			return types.NewClassOfType(classType.Parent)
		}
		return types.NewClassOfType(classType)
	}

	// Look up field (including inherited fields)
	fieldType, found := classType.GetField(memberName)
	if found {
		if isMetaclass {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Member.Token.Pos))
			return nil
		}
		fieldOwner := a.getFieldOwner(classType, memberName)
		if fieldOwner != nil {
			visibility, hasVisibility := fieldOwner.FieldVisibility[memberName]
			if hasVisibility && !a.checkVisibility(fieldOwner, visibility, memberName, "field") {
				a.addStructuredError(NewVisibilityScopeError(expr.Member.Token.Pos, expr.Member.Value))
				return nil
			}
			a.recordClassFieldUsage(fieldOwner, memberName)
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
				a.addStructuredError(NewVisibilityScopeError(expr.Member.Token.Pos, expr.Member.Value))
				return nil
			}
		}
		return classVarType
	}

	// Look up property (including inherited properties)
	propInfo, propFound := classType.GetProperty(memberName)
	if propFound {
		if propInfo.ReadKind == types.PropAccessNone {
			a.addStructuredError(NewWriteOnlyPropertyError(expr.Member.Token.Pos, expr.Member.Value))
			return nil
		}
		if isMetaclass && !propInfo.IsClassProperty {
			switch propInfo.ReadKind {
			case types.PropAccessField, types.PropAccessExpression, types.PropAccessBuiltin:
				a.addStructuredError(NewObjectReferenceNeededError(expr.Member.Token.Pos))
				return nil
			case types.PropAccessMethod:
				if propInfo.ReadSpec != "" && (classType.ClassMethodFlags == nil || !classType.ClassMethodFlags[ident.Normalize(propInfo.ReadSpec)]) {
					a.addStructuredError(NewPropertyReadShouldBeStaticMethodError(expr.Member.Token.Pos))
					a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Member.Token.Pos))
					return nil
				}
				a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Member.Token.Pos))
				return nil
			}
			a.addStructuredError(NewObjectReferenceNeededError(expr.Member.Token.Pos))
			return nil
		}
		return propInfo.Type
	}

	// Look up constructor (constructors are stored separately)
	constructorOverloads := classType.GetConstructorOverloads(memberName)
	if len(constructorOverloads) == 0 {
		if ctorType, found := classType.GetConstructor(memberName); found {
			constructorOverloads = []*types.MethodInfo{{Signature: ctorType}}
		}
	}
	if len(constructorOverloads) > 0 {
		declaredCtor := a.declaredMethodName(classType, memberName)
		if declaredCtor == "" && memberName == "create" {
			declaredCtor = "Create" // built-in TObject constructor
		}
		if declaredCtor != "" && expr.Member.Value != declaredCtor {
			pos := expr.Token.Pos
			pos.Column++
			a.addCaseMismatchHint(expr.Member.Value, declaredCtor, pos)
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
			a.recordClassMethodUsage(classType, memberName)
			if classType.IsStatic {
				a.addStructuredError(NewStaticClassInstantiationError(expr.Member.Token.Pos, classType.Name))
				return classType
			}
			if classType.IsAbstract || len(a.getUnimplementedAbstractMethods(classType)) > 0 {
				a.addStructuredError(NewAbstractInstantiationError(expr.Member.Token.Pos))
				return classType
			}
			return classType
		}
		// Constructor has parameters - return method pointer for deferred invocation
		if len(constructorOverloads) == 1 {
			return types.NewMethodPointerType(constructorOverloads[0].Signature.Parameters, classType)
		}
		return types.NewMethodPointerType([]types.Type{}, classType)
	}

	if isMetaclass && ident.Equal(memberName, "create") {
		if classType.IsAbstract || len(a.getUnimplementedAbstractMethods(classType)) > 0 {
			a.addStructuredError(NewAbstractInstantiationError(expr.Member.Token.Pos))
			return classType
		}
		return classType
	}

	// Look up method (including inherited methods)
	methodType, found := classType.GetMethod(memberName)
	if found {
		// A method accessed through a metaclass value must be a class method. Resolve the
		// class-method flag across the hierarchy so inherited class methods (absent from
		// this class's own ClassMethodFlags map) are accepted.
		if isMetaclass && !a.isClassMethodInHierarchy(classType, memberName) {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Member.Token.Pos))
			return nil
		}
		if memberName == "free" && expr.Member.Value != "Free" {
			pos := expr.Token.Pos
			pos.Column++
			a.addCaseMismatchHint(expr.Member.Value, "Free", pos)
		}
		methodOwner := a.getMethodOwner(classType, memberName)
		if methodOwner != nil {
			visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(memberName)]
			if hasVisibility && !a.checkVisibility(methodOwner, visibility, memberName, "method") {
				a.addStructuredError(NewVisibilityScopeError(expr.Member.Token.Pos, expr.Member.Value))
				return nil
			}
			a.recordClassMethodUsage(methodOwner, memberName)
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
	helperMethod := a.hasHelperMethod(objectType, memberName)
	if helperMethod != nil {
		if len(helperMethod.Parameters) == 0 {
			return helperMethod.ReturnType
		}
		return helperMethod
	}

	// Check helpers for properties
	helperProp := a.hasHelperProperty(objectType, memberName)
	if helperProp != nil {
		if enumType, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
			a.maybeAddUnnamedEnumElementHint(expr.Object, expr.Member.Token.Pos, enumType, memberName)
		}
		return helperProp.Type
	}

	// Check for class constants (including inherited constants)
	if constType := a.findClassConstantWithVisibility(classType, memberName, expr.Token.Pos.String()); constType != nil {
		return constType
	}

	a.addStructuredError(NewAccessibleMemberError(expr.Member.Token.Pos, expr.Member.Value, objectType.String()))
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
		argType := a.analyzeOverloadArgument(arg)
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
		a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, methodName))
		return nil
	}

	funcType, ok := selected.Type.(*types.FunctionType)
	if !ok {
		a.addError("internal error: expected function type for selected record method, but got %T", selected.Type)
		return nil
	}
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

func (a *Analyzer) maybeAddUnnamedEnumElementHint(expr ast.Expression, pos token.Position, enumType *types.EnumType, memberName string) {
	if expr == nil || enumType == nil {
		return
	}
	if memberName != "name" && memberName != "qualifiedname" {
		return
	}

	value, err := a.evaluateConstant(expr)
	if err != nil {
		return
	}

	ordinal, ok := value.(int)
	if !ok {
		return
	}
	if enumType.GetEnumName(ordinal) != "" {
		return
	}

	a.addHint("Enumeration element is unnamed or out of range [line: %d, column: %d]", pos.Line, pos.Column)
}
