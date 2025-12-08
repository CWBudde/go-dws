package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

func (i *Interpreter) evalMethodCall(mc *ast.MethodCallExpression) Value {
	methodNameLower := pkgident.Normalize(mc.Method.Value)

	// Check if the left side is an identifier (unit, class, or instance variable)
	if ident, ok := mc.Object.(*ast.Identifier); ok {
		// Try unit-qualified function call
		if result := i.tryUnitQualifiedCall(ident, mc); result != nil {
			return result
		}

		// Try class name method call
		if result := i.tryClassNameMethodCall(ident, mc); result != nil {
			return result
		}

		// Try record type static method call
		if result := i.tryRecordTypeMethodCall(ident, mc); result != nil {
			return result
		}
	}

	// Evaluate object expression for instance method call
	objVal := i.Eval(mc.Object)
	if isError(objVal) {
		return objVal
	}

	// Unwrap type cast wrappers so method dispatch uses the underlying value.
	if castVal, ok := objVal.(*TypeCastValue); ok {
		objVal = castVal.Object
	}

	// Try various value type method calls
	if result := i.tryClassInfoValueMethodCall(objVal, mc); result != nil {
		return result
	}
	if result := i.tryClassValueConstructorCall(objVal, mc); result != nil {
		return result
	}
	if result := i.trySetValueMethodCall(objVal, mc); result != nil {
		return result
	}
	if result := i.tryRecordValueMethodCall(objVal, mc); result != nil {
		return result
	}

	// Check for interface instance - delegate to underlying object
	objVal = i.unwrapInterfaceInstance(objVal, mc)
	if isError(objVal) {
		return objVal
	}

	// Initialize typed nil values when possible
	objVal = i.initializeTypedNilValue(objVal, mc)

	// Check if object is nil (TObject.Free is nil-safe)
	if result := i.handleNilObjectMethodCall(objVal, mc); result != nil {
		return result
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return i.handleNonObjectMethodCall(objVal, mc)
	}

	return i.executeObjectInstanceMethod(obj, methodNameLower, mc)
}

// executeClassMethod executes a class method with Self bound to ClassInfo.
func (i *Interpreter) executeClassMethod(
	classInfo *ClassInfo,
	classMethod *ast.FunctionDecl,
	mc *ast.MethodCallExpression,
) Value {
	// Evaluate arguments
	args, evalErr := i.evalMethodArguments(mc.Arguments)
	if evalErr != nil {
		return evalErr
	}

	// Check argument count
	if len(args) != len(classMethod.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for class method '%s': expected %d, got %d",
			mc.Method.Value, len(classMethod.Parameters), len(args))
	}

	// Create method environment
	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)

	// Check recursion depth before pushing to call stack
	if i.ctx.GetCallStack().WillOverflow() {
		i.RestoreEnvironment(savedEnv)
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces
	fullMethodName := classInfo.Name + "." + mc.Method.Value
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind Self and __CurrentClass__ to ClassInfo for class methods
	methodEnv.Define("Self", &ClassInfoValue{ClassInfo: classInfo})
	methodEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})
	i.bindClassConstantsToEnv(classInfo)
	i.bindMethodParameters(methodEnv, classMethod.Parameters, args)
	i.initializeResultVariable(methodEnv, classMethod)

	// Execute method body
	result := i.Eval(classMethod.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	returnValue := i.extractMethodReturnValue(classMethod)
	i.RestoreEnvironment(savedEnv)
	return returnValue
}

// initializeResultVariable initializes the Result variable for methods with return types.
func (i *Interpreter) initializeResultVariable(env *Environment, method *ast.FunctionDecl) {
	if method.ReturnType == nil {
		return
	}
	returnType := i.resolveTypeFromAnnotation(method.ReturnType)
	defaultVal := i.getDefaultValue(returnType)
	env.Define("Result", defaultVal)
	env.Define(method.Name.Value, &ReferenceValue{Env: env, VarName: "Result"})
}

// resolveMethodOverload resolves method overload based on argument types.
func (i *Interpreter) resolveMethodOverload(className, methodName string, overloads []*ast.FunctionDecl, argExprs []ast.Expression) (*ast.FunctionDecl, error) {
	// If only one overload, use it (fast path)
	if len(overloads) == 1 {
		return overloads[0], nil
	}

	// Evaluate arguments to get their types
	argTypes := make([]types.Type, len(argExprs))
	for idx, argExpr := range argExprs {
		val := i.Eval(argExpr)
		if isError(val) {
			return nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
		}
		argTypes[idx] = i.getValueType(val)
	}

	// Convert method declarations to semantic symbols for resolution
	candidates := make([]*semantic.Symbol, len(overloads))
	for idx, method := range overloads {
		methodType := i.extractFunctionType(method)
		if methodType == nil {
			return nil, fmt.Errorf("unable to extract method type for overload %d of '%s.%s'", idx+1, className, methodName)
		}

		candidates[idx] = &semantic.Symbol{
			Name:                 method.Name.Value,
			Type:                 methodType,
			HasOverloadDirective: method.IsOverload,
		}
	}

	// Use semantic analyzer's overload resolution
	selected, err := semantic.ResolveOverload(candidates, argTypes)
	if err != nil {
		return nil, fmt.Errorf("there is no overloaded version of \"%s.%s\" that can be called with these arguments", className, methodName)
	}

	// Find which method declaration corresponds to the selected symbol
	selectedType := selected.Type.(*types.FunctionType)
	for _, method := range overloads {
		methodType := i.extractFunctionType(method)
		if methodType != nil && semantic.SignaturesEqual(methodType, selectedType) &&
			methodType.ReturnType.Equals(selectedType.ReturnType) {
			return method, nil
		}
	}

	return nil, fmt.Errorf("internal error: resolved overload not found in candidate list")
}

// getMethodOverloadsInHierarchy collects all overloads of a method from the class hierarchy.
func (i *Interpreter) getMethodOverloadsInHierarchy(classInfo *ClassInfo, methodName string, isClassMethod bool) []*ast.FunctionDecl {
	// Check for constructors first (only when isClassMethod = false)
	if !isClassMethod {
		if ctorOverloads := i.findConstructorOverloads(classInfo, methodName); len(ctorOverloads) > 0 {
			return ctorOverloads
		}
	}

	// Walk up the class hierarchy for regular methods
	var result []*ast.FunctionDecl
	for classInfo != nil {
		overloads := i.getMethodOverloadsForClass(classInfo, methodName, isClassMethod)
		result = i.addNonHiddenOverloads(result, overloads)
		classInfo = classInfo.Parent
	}

	return result
}

// findConstructorOverloads finds constructor overloads matching the given name.
func (i *Interpreter) findConstructorOverloads(classInfo *ClassInfo, methodName string) []*ast.FunctionDecl {
	for ctorName, constructorOverloads := range classInfo.ConstructorOverloads {
		if pkgident.Equal(ctorName, methodName) && len(constructorOverloads) > 0 {
			return constructorOverloads
		}
	}
	return nil
}

// getMethodOverloadsForClass gets method overloads for a single class level.
func (i *Interpreter) getMethodOverloadsForClass(classInfo *ClassInfo, methodName string, isClassMethod bool) []*ast.FunctionDecl {
	var overloadMap map[string][]*ast.FunctionDecl
	if isClassMethod {
		overloadMap = classInfo.ClassMethodOverloads
	} else {
		overloadMap = classInfo.MethodOverloads
	}

	for name, methods := range overloadMap {
		if pkgident.Equal(name, methodName) {
			return methods
		}
	}
	return nil
}

// addNonHiddenOverloads adds overloads that aren't hidden by existing ones in the result.
func (i *Interpreter) addNonHiddenOverloads(result, overloads []*ast.FunctionDecl) []*ast.FunctionDecl {
	for _, candidate := range overloads {
		if !i.isMethodHidden(candidate, result) {
			result = append(result, candidate)
		}
	}
	return result
}

// isMethodHidden checks if a method is hidden by an existing method in the list.
func (i *Interpreter) isMethodHidden(candidate *ast.FunctionDecl, existing []*ast.FunctionDecl) bool {
	candidateType := i.extractFunctionType(candidate)
	if candidateType == nil {
		return false
	}

	for _, method := range existing {
		existingType := i.extractFunctionType(method)
		if existingType != nil && semantic.SignaturesEqual(candidateType, existingType) {
			return true
		}
	}
	return false
}

// executeObjectInstanceMethod executes a method call on an object instance.
func (i *Interpreter) executeObjectInstanceMethod(obj *ObjectInstance, methodNameLower string, mc *ast.MethodCallExpression) Value {
	// Prevent method calls on destroyed instances
	if obj.Destroyed {
		message := fmt.Sprintf("Object already destroyed [line: %d, column: %d]", mc.Token.Pos.Line, mc.Token.Pos.Column)
		i.raiseException("Exception", message, &mc.Token.Pos)
		return &NilValue{}
	}

	// TObject.Free delegates to Destroy and is available on all classes
	if methodNameLower == "free" {
		if len(mc.Arguments) != 0 {
			return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
				"Free", 0, len(mc.Arguments))
		}
		return i.runDestructor(obj, obj.Class.LookupMethod("Destroy"), mc)
	}

	// Handle built-in methods that are available on all objects (inherited from TObject)
	if methodNameLower == "classname" {
		return &StringValue{Value: obj.Class.GetName()}
	}

	concreteClass, ok := obj.Class.(*ClassInfo)
	if !ok {
		return i.newErrorWithLocation(mc, "object has invalid class type")
	}

	// Resolve method
	method, isClassMethod, err := i.resolveObjectMethod(obj, concreteClass, mc)
	if err != nil {
		return i.newErrorWithLocation(mc, "%s", err.Error())
	}

	// If no method found, try helper
	if method == nil {
		return i.tryObjectHelperMethod(obj, mc)
	}

	// Evaluate method arguments
	args, evalErr := i.evalMethodArguments(mc.Arguments)
	if evalErr != nil {
		return evalErr
	}

	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
			mc.Method.Value, len(method.Parameters), len(args))
	}

	if method.IsDestructor {
		return i.runDestructor(obj, method, mc)
	}

	if method.IsConstructor {
		return i.executeVirtualConstructor(obj, concreteClass, method, args, mc)
	}

	return i.executeResolvedMethod(obj, concreteClass, method, isClassMethod, args, mc)
}

// resolveObjectMethod resolves the method to call on an object instance.
func (i *Interpreter) resolveObjectMethod(
	obj *ObjectInstance,
	concreteClass *ClassInfo,
	mc *ast.MethodCallExpression,
) (*ast.FunctionDecl, bool, error) {
	methodOverloads := i.getMethodOverloadsInHierarchy(concreteClass, mc.Method.Value, false)
	classMethodOverloads := i.getMethodOverloadsInHierarchy(concreteClass, mc.Method.Value, true)

	var method *ast.FunctionDecl
	var err error
	var isClassMethod bool

	// Try instance methods first
	if len(methodOverloads) > 0 {
		method, err = i.resolveMethodOverload(obj.Class.GetName(), mc.Method.Value, methodOverloads, mc.Arguments)
		if err != nil {
			return nil, false, err
		}
		method = i.resolveVirtualMethod(method, concreteClass)
	}

	// If no instance method found, try class methods
	if method == nil && len(classMethodOverloads) > 0 {
		method, err = i.resolveMethodOverload(obj.Class.GetName(), mc.Method.Value, classMethodOverloads, mc.Arguments)
		if err != nil {
			return nil, false, err
		}
		isClassMethod = true
		method = i.resolveStaticClassMethod(method, concreteClass, mc.Method.Value, mc.Arguments)
		method = i.resolveVirtualMethod(method, concreteClass)
	}

	return method, isClassMethod, nil
}

// resolveVirtualMethod resolves virtual/override methods through the virtual method table.
func (i *Interpreter) resolveVirtualMethod(method *ast.FunctionDecl, concreteClass *ClassInfo) *ast.FunctionDecl {
	if method == nil || (!method.IsVirtual && !method.IsOverride) || concreteClass.VirtualMethodTable == nil {
		return method
	}

	sig := methodSignature(method)
	if entry, exists := concreteClass.VirtualMethodTable[sig]; exists && entry != nil && entry.IsVirtual {
		if entry.Implementation != nil {
			return entry.Implementation
		}
	}
	return method
}

// resolveStaticClassMethod finds the top-most declaration for non-virtual class methods.
func (i *Interpreter) resolveStaticClassMethod(
	method *ast.FunctionDecl,
	concreteClass *ClassInfo,
	methodName string,
	arguments []ast.Expression,
) *ast.FunctionDecl {
	if method == nil || method.IsVirtual || method.IsOverride {
		return method
	}

	topMostMethod := method
	for currentClass := concreteClass.Parent; currentClass != nil; currentClass = currentClass.Parent {
		var parentClassMethodOverloads []*ast.FunctionDecl
		for name, methods := range currentClass.ClassMethodOverloads {
			if pkgident.Equal(name, methodName) {
				parentClassMethodOverloads = methods
				break
			}
		}

		if len(parentClassMethodOverloads) > 0 {
			parentMethod, parentErr := i.resolveMethodOverload(currentClass.Name, methodName, parentClassMethodOverloads, arguments)
			if parentErr == nil && parentMethod != nil {
				topMostMethod = parentMethod
			}
		}
	}
	return topMostMethod
}

// tryObjectHelperMethod tries to call a helper method on an object.
func (i *Interpreter) tryObjectHelperMethod(obj *ObjectInstance, mc *ast.MethodCallExpression) Value {
	helper, helperMethod, builtinSpec := i.findHelperMethod(obj, mc.Method.Value)
	if helperMethod == nil && builtinSpec == "" {
		return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, obj.Class.GetName())
	}

	args, evalErr := i.evalMethodArguments(mc.Arguments)
	if evalErr != nil {
		return evalErr
	}

	return i.callHelperMethod(helper, helperMethod, builtinSpec, obj, args, mc)
}

// evalMethodArguments evaluates method call arguments.
func (i *Interpreter) evalMethodArguments(arguments []ast.Expression) ([]Value, Value) {
	args := make([]Value, len(arguments))
	for idx, arg := range arguments {
		val := i.Eval(arg)
		if isError(val) {
			return nil, val
		}
		args[idx] = val
	}
	return args, nil
}

// executeVirtualConstructor executes a constructor on an object instance.
func (i *Interpreter) executeVirtualConstructor(
	obj *ObjectInstance,
	concreteClass *ClassInfo,
	method *ast.FunctionDecl,
	args []Value,
	mc *ast.MethodCallExpression,
) Value {
	actualConstructor := i.findActualConstructor(concreteClass, mc.Method.Value, method)

	newObj := NewObjectInstance(obj.Class)
	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)
	methodEnv.Define("Self", newObj)
	i.bindClassConstantsToEnv(concreteClass)

	i.bindMethodParameters(methodEnv, actualConstructor.Parameters, args)

	result := i.Eval(actualConstructor.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	i.RestoreEnvironment(savedEnv)
	return newObj
}

// findActualConstructor finds the actual constructor to call in the class hierarchy.
func (i *Interpreter) findActualConstructor(concreteClass *ClassInfo, constructorName string, defaultMethod *ast.FunctionDecl) *ast.FunctionDecl {
	for class := concreteClass; class != nil; class = class.Parent {
		if ctor, exists := class.Constructors[constructorName]; exists {
			return ctor
		}
		for name, ctor := range class.Constructors {
			if pkgident.Equal(name, constructorName) {
				return ctor
			}
		}
	}
	return defaultMethod
}

// bindMethodParameters binds method parameters to arguments with implicit conversion.
func (i *Interpreter) bindMethodParameters(env *Environment, params []*ast.Parameter, args []Value) {
	for idx, param := range params {
		arg := args[idx]
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}
		env.Define(param.Name.Value, arg)
	}
}

// executeResolvedMethod executes a resolved method on an object.
func (i *Interpreter) executeResolvedMethod(
	obj *ObjectInstance,
	concreteClass *ClassInfo,
	method *ast.FunctionDecl,
	isClassMethod bool,
	args []Value,
	mc *ast.MethodCallExpression,
) Value {
	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)

	if i.ctx.GetCallStack().WillOverflow() {
		i.RestoreEnvironment(savedEnv)
		return i.raiseMaxRecursionExceeded()
	}

	fullMethodName := obj.Class.GetName() + "." + mc.Method.Value
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	if isClassMethod {
		methodEnv.Define("Self", &ClassInfoValue{ClassInfo: concreteClass})
	} else {
		methodEnv.Define("Self", obj)
	}

	methodOwner := i.findMethodOwner(concreteClass, method, isClassMethod)
	methodEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: methodOwner})
	i.bindClassConstantsToEnv(concreteClass)

	i.bindMethodParameters(methodEnv, method.Parameters, args)

	if method.ReturnType != nil {
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		defaultVal := i.getDefaultValue(returnType)
		methodEnv.Define("Result", defaultVal)
		methodEnv.Define(method.Name.Value, &ReferenceValue{Env: methodEnv, VarName: "Result"})
	}

	result := i.Eval(method.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	returnValue := i.extractMethodReturnValue(method)
	i.RestoreEnvironment(savedEnv)
	return returnValue
}

// extractMethodReturnValue extracts the return value from a method execution.
func (i *Interpreter) extractMethodReturnValue(method *ast.FunctionDecl) Value {
	if method.ReturnType == nil {
		return &NilValue{}
	}

	resultVal, resultOk := i.env.Get("Result")
	methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

	if resultOk {
		if refVal, isRef := resultVal.(*ReferenceValue); isRef {
			if derefVal, err := refVal.Dereference(); err == nil {
				resultVal = derefVal
			}
		}
	}
	if methodNameOk {
		if refVal, isRef := methodNameVal.(*ReferenceValue); isRef {
			if derefVal, err := refVal.Dereference(); err == nil {
				methodNameVal = derefVal
			}
		}
	}

	var returnValue Value
	switch {
	case resultOk && resultVal.Type() != "NIL":
		returnValue = resultVal
	case methodNameOk && methodNameVal.Type() != "NIL":
		returnValue = methodNameVal
	case resultOk:
		returnValue = resultVal
	case methodNameOk:
		returnValue = methodNameVal
	default:
		returnValue = &NilValue{}
	}

	if returnValue.Type() != "NIL" {
		expectedReturnType := method.ReturnType.String()
		if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
			returnValue = converted
		}
	}

	return returnValue
}

// tryClassInfoValueMethodCall handles method calls on ClassInfoValue (Self in a class method).
// Returns nil if objVal is not a ClassInfoValue.
func (i *Interpreter) tryClassInfoValueMethodCall(objVal Value, mc *ast.MethodCallExpression) Value {
	classInfoVal, ok := objVal.(*ClassInfoValue)
	if !ok {
		return nil
	}

	classInfo := classInfoVal.ClassInfo
	classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, true)

	if len(classMethodOverloads) == 0 {
		return i.newErrorWithLocation(mc, "class method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
	}

	classMethod, err := i.resolveMethodOverload(classInfo.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
	if err != nil {
		return i.newErrorWithLocation(mc, "%s", err.Error())
	}

	return i.executeClassMethod(classInfo, classMethod, mc)
}

// tryClassValueConstructorCall handles constructor calls on ClassValue (metaclass).
// Returns nil if objVal is not a ClassValue.
func (i *Interpreter) tryClassValueConstructorCall(objVal Value, mc *ast.MethodCallExpression) Value {
	classVal, ok := objVal.(*ClassValue)
	if !ok {
		return nil
	}

	methodName := mc.Method.Value
	runtimeClass := classVal.ClassInfo
	if runtimeClass == nil {
		return i.newErrorWithLocation(mc, "invalid class reference")
	}

	constructorOverloads := i.getMethodOverloadsInHierarchy(runtimeClass, methodName, false)

	if len(constructorOverloads) == 0 {
		return i.newErrorWithLocation(mc, "constructor '%s' not found in class '%s'", methodName, runtimeClass.Name)
	}

	constructor, err := i.resolveMethodOverload(runtimeClass.Name, methodName, constructorOverloads, mc.Arguments)
	if err != nil {
		return i.newErrorWithLocation(mc, "%s", err.Error())
	}

	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	if len(args) != len(constructor.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for constructor '%s': expected %d, got %d",
			methodName, len(constructor.Parameters), len(args))
	}

	newInstance := NewObjectInstance(runtimeClass)

	for fieldName, fieldType := range runtimeClass.Fields {
		var defaultValue Value
		switch fieldType {
		case types.INTEGER:
			defaultValue = &IntegerValue{Value: 0}
		case types.FLOAT:
			defaultValue = &FloatValue{Value: 0.0}
		case types.STRING:
			defaultValue = &StringValue{Value: ""}
		case types.BOOLEAN:
			defaultValue = &BooleanValue{Value: false}
		default:
			defaultValue = &NilValue{}
		}
		newInstance.SetField(fieldName, defaultValue)
	}

	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)
	methodEnv.Define("Self", newInstance)
	i.bindClassConstantsToEnv(runtimeClass)

	for idx, param := range constructor.Parameters {
		arg := args[idx]
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}
		methodEnv.Define(param.Name.Value, arg)
	}

	methodEnv.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: runtimeClass})

	result := i.Eval(constructor.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	i.RestoreEnvironment(savedEnv)
	return newInstance
}

// trySetValueMethodCall handles method calls on SetValue.
// Returns nil if objVal is not a SetValue.
func (i *Interpreter) trySetValueMethodCall(objVal Value, mc *ast.MethodCallExpression) Value {
	setVal, ok := objVal.(*SetValue)
	if !ok {
		return nil
	}

	methodName := pkgident.Normalize(mc.Method.Value)

	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	switch methodName {
	case "include":
		if len(args) != 1 {
			return i.newErrorWithLocation(mc, "Include expects 1 argument, got %d", len(args))
		}
		return i.evalSetInclude(setVal, args[0])

	case "exclude":
		if len(args) != 1 {
			return i.newErrorWithLocation(mc, "Exclude expects 1 argument, got %d", len(args))
		}
		return i.evalSetExclude(setVal, args[0])

	default:
		return i.newErrorWithLocation(mc, "method '%s' not found for set type", methodName)
	}
}

// tryRecordValueMethodCall handles method calls on RecordValue.
// Returns nil if objVal is not a RecordValue.
func (i *Interpreter) tryRecordValueMethodCall(objVal Value, mc *ast.MethodCallExpression) Value {
	recVal, ok := objVal.(*RecordValue)
	if !ok {
		return nil
	}

	memberAccess := &ast.MemberAccessExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: mc.Token,
			},
		},
		Object: mc.Object,
		Member: mc.Method,
	}
	return i.evalRecordMethodCall(recVal, memberAccess, mc.Arguments, mc.Object)
}

// unwrapInterfaceInstance unwraps an InterfaceInstance to its underlying object.
// Returns the original value if not an InterfaceInstance, or an error if invalid.
func (i *Interpreter) unwrapInterfaceInstance(objVal Value, mc *ast.MethodCallExpression) Value {
	intfInst, ok := objVal.(*InterfaceInstance)
	if !ok {
		return objVal
	}

	if intfInst.Object == nil {
		return i.newErrorWithLocation(mc, "Interface is nil")
	}

	if !intfInst.Interface.HasMethod(mc.Method.Value) {
		return i.newErrorWithLocation(mc, "method '%s' not found in interface '%s'",
			mc.Method.Value, intfInst.Interface.GetName())
	}

	return intfInst.Object
}

// initializeTypedNilValue initializes typed nil values when possible.
func (i *Interpreter) initializeTypedNilValue(objVal Value, mc *ast.MethodCallExpression) Value {
	if objVal == nil || objVal.Type() != "NIL" || i.semanticInfo == nil {
		return objVal
	}

	objType := i.semanticInfo.GetType(mc.Object)
	if objType == nil {
		return objVal
	}

	typeName := objType.String()
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		if arrayType := i.parseInlineArrayType(typeName); arrayType != nil {
			return NewArrayValue(arrayType)
		}
	}

	return objVal
}

// handleNilObjectMethodCall handles method calls on nil objects.
// Returns nil if the object is not nil, otherwise returns the appropriate result.
func (i *Interpreter) handleNilObjectMethodCall(objVal Value, mc *ast.MethodCallExpression) Value {
	if objVal != nil && objVal.Type() != "NIL" {
		return nil
	}

	// TObject.Free is nil-safe
	if strings.EqualFold(strings.TrimSpace(mc.Method.Value), "Free") {
		return &NilValue{}
	}

	message := fmt.Sprintf("Object not instantiated [line: %d, column: %d]", mc.Token.Pos.Line, mc.Token.Pos.Column+1)
	i.raiseException("Exception", message, &mc.Token.Pos)
	return &NilValue{}
}

// handleNonObjectMethodCall handles method calls on non-object values (enums, helpers).
func (i *Interpreter) handleNonObjectMethodCall(objVal Value, mc *ast.MethodCallExpression) Value {
	// Special handling for enum type methods: Low(), High(), and ByName()
	if result := i.tryEnumTypeMethodCall(objVal, mc); result != nil {
		return result
	}

	// Check if helpers provide this method
	helper, helperMethod, builtinSpec := i.findHelperMethod(objVal, mc.Method.Value)
	if helperMethod == nil && builtinSpec == "" {
		return i.newErrorWithLocation(mc, "cannot call method '%s' on type '%s' (no helper found)",
			mc.Method.Value, objVal.Type())
	}

	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	return i.callHelperMethod(helper, helperMethod, builtinSpec, objVal, args, mc)
}

// tryEnumTypeMethodCall handles method calls on enum types.
// Returns nil if objVal is not an enum type meta value.
func (i *Interpreter) tryEnumTypeMethodCall(objVal Value, mc *ast.MethodCallExpression) Value {
	tmv, isTypeMeta := objVal.(*TypeMetaValue)
	if !isTypeMeta {
		return nil
	}

	enumType, isEnum := tmv.TypeInfo.(*types.EnumType)
	if !isEnum {
		return nil
	}

	methodName := pkgident.Normalize(mc.Method.Value)
	switch methodName {
	case "low":
		return &IntegerValue{Value: int64(enumType.Low())}
	case "high":
		return &IntegerValue{Value: int64(enumType.High())}
	case "byname":
		if len(mc.Arguments) != 1 {
			return i.newErrorWithLocation(mc, "ByName expects 1 argument, got %d", len(mc.Arguments))
		}
		nameArg := i.Eval(mc.Arguments[0])
		if isError(nameArg) {
			return nameArg
		}
		nameStr, ok := nameArg.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(mc, "ByName expects string argument, got %s", nameArg.Type())
		}

		searchName := nameStr.Value
		if searchName == "" {
			return &IntegerValue{Value: 0}
		}

		parts := strings.Split(searchName, ".")
		if len(parts) == 2 {
			searchName = parts[1]
		}

		for valueName, ordinalValue := range enumType.Values {
			if pkgident.Equal(valueName, searchName) {
				return &IntegerValue{Value: int64(ordinalValue)}
			}
		}

		return &IntegerValue{Value: 0}
	}

	return nil
}

// tryUnitQualifiedCall attempts to call a unit-qualified function.
// Returns nil if the identifier is not a unit name.
func (i *Interpreter) tryUnitQualifiedCall(ident *ast.Identifier, mc *ast.MethodCallExpression) Value {
	if i.unitRegistry == nil {
		return nil
	}
	if _, exists := i.unitRegistry.GetUnit(ident.Value); !exists {
		return nil
	}

	fn, err := i.ResolveQualifiedFunction(ident.Value, mc.Method.Value)
	if err != nil {
		return i.newErrorWithLocation(mc, "function '%s' not found in unit '%s'", mc.Method.Value, ident.Value)
	}

	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}
	return i.executeUserFunctionViaEvaluator(fn, args)
}

// tryClassNameMethodCall attempts to call a method on a class name (static or constructor).
// Returns nil if the identifier is not a class name.
func (i *Interpreter) tryClassNameMethodCall(ident *ast.Identifier, mc *ast.MethodCallExpression) Value {
	classInfo := i.resolveClassInfoByName(ident.Value)
	if classInfo == nil {
		return nil
	}

	classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, true)
	instanceMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, mc.Method.Value, false)

	// Handle implicit parameterless constructor
	if result := i.tryImplicitConstructor(classInfo, instanceMethodOverloads, mc); result != nil {
		return result
	}

	// Resolve method overloads
	classMethod, instanceMethod, err := i.resolveClassAndInstanceMethods(
		classInfo, mc.Method.Value, classMethodOverloads, instanceMethodOverloads, mc.Arguments,
	)
	if err != nil {
		return i.newErrorWithLocation(mc, "%s", err.Error())
	}

	if classMethod != nil {
		return i.executeClassMethod(classInfo, classMethod, mc)
	}
	if instanceMethod != nil {
		return i.executeInstanceMethodOnClassName(classInfo, instanceMethod, mc)
	}

	return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
}

// tryImplicitConstructor handles calls like TClass.Create with no arguments when
// a constructor exists but none are parameterless.
func (i *Interpreter) tryImplicitConstructor(
	classInfo *ClassInfo,
	instanceMethodOverloads []*ast.FunctionDecl,
	mc *ast.MethodCallExpression,
) Value {
	if len(mc.Arguments) != 0 {
		return nil
	}

	hasConstructor := false
	hasParameterlessConstructor := false
	for _, method := range instanceMethodOverloads {
		if method.IsConstructor {
			hasConstructor = true
			if len(method.Parameters) == 0 {
				hasParameterlessConstructor = true
				break
			}
		}
	}

	if !hasConstructor || hasParameterlessConstructor {
		return nil
	}

	obj := NewObjectInstance(classInfo)
	savedEnv := i.env
	tempEnv := i.PushEnvironment(i.env)
	for constName, constValue := range classInfo.ConstantValues {
		tempEnv.Define(constName, constValue)
	}

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.RestoreEnvironment(savedEnv)
				return fieldValue
			}
		} else {
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	i.RestoreEnvironment(savedEnv)
	return obj
}

// resolveClassAndInstanceMethods resolves both class and instance method overloads.
func (i *Interpreter) resolveClassAndInstanceMethods(
	classInfo *ClassInfo,
	methodName string,
	classMethodOverloads, instanceMethodOverloads []*ast.FunctionDecl,
	arguments []ast.Expression,
) (classMethod, instanceMethod *ast.FunctionDecl, err error) {
	if len(classMethodOverloads) > 0 {
		classMethod, err = i.resolveMethodOverload(classInfo.Name, methodName, classMethodOverloads, arguments)
		if err != nil && len(instanceMethodOverloads) == 0 {
			return nil, nil, err
		}
	}

	if len(instanceMethodOverloads) > 0 {
		instanceMethod, err = i.resolveMethodOverload(classInfo.Name, methodName, instanceMethodOverloads, arguments)
		if err != nil && classMethod == nil {
			return nil, nil, err
		}
	}

	return classMethod, instanceMethod, nil
}

// executeInstanceMethodOnClassName executes an instance method called on a class name
// (e.g., TClass.Create) by creating a new instance first.
func (i *Interpreter) executeInstanceMethodOnClassName(
	classInfo *ClassInfo,
	instanceMethod *ast.FunctionDecl,
	mc *ast.MethodCallExpression,
) Value {
	obj := NewObjectInstance(classInfo)

	// Initialize fields
	fieldInitEnv := i.env
	fieldTempEnv := i.PushEnvironment(i.env)
	for constName, constValue := range classInfo.ConstantValues {
		fieldTempEnv.Define(constName, constValue)
	}

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.RestoreEnvironment(fieldInitEnv)
				return fieldValue
			}
		} else {
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}
	i.RestoreEnvironment(fieldInitEnv)

	// Evaluate method arguments
	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	if len(args) != len(instanceMethod.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
			mc.Method.Value, len(instanceMethod.Parameters), len(args))
	}

	// Create method environment
	savedEnv := i.env
	methodEnv := i.PushEnvironment(i.env)
	methodEnv.Define("Self", obj)
	i.bindClassConstantsToEnv(classInfo)

	for idx, param := range instanceMethod.Parameters {
		arg := args[idx]
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}
		methodEnv.Define(param.Name.Value, arg)
	}

	if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
		var defaultVal Value
		if instanceMethod.IsConstructor {
			defaultVal = &NilValue{}
		} else {
			returnType := i.resolveTypeFromAnnotation(instanceMethod.ReturnType)
			defaultVal = i.getDefaultValue(returnType)
		}
		methodEnv.Define("Result", defaultVal)
		methodEnv.Define(instanceMethod.Name.Value, &ReferenceValue{Env: methodEnv, VarName: "Result"})
	}

	result := i.Eval(instanceMethod.Body)
	if isError(result) {
		i.RestoreEnvironment(savedEnv)
		return result
	}

	returnValue := i.extractInstanceMethodReturnValue(obj, instanceMethod)
	i.RestoreEnvironment(savedEnv)
	return returnValue
}

// extractInstanceMethodReturnValue extracts the return value from an instance method.
func (i *Interpreter) extractInstanceMethodReturnValue(obj *ObjectInstance, method *ast.FunctionDecl) Value {
	if method.ReturnType == nil && !method.IsConstructor {
		return &NilValue{}
	}

	if method.IsConstructor && method.ReturnType == nil {
		return obj
	}

	if method.IsConstructor && method.ReturnType != nil {
		resultVal, resultOk := i.env.Get("Result")
		if resultOk && resultVal.Type() != "NIL" {
			return resultVal
		}
		return obj
	}

	// Non-constructor with return type
	resultVal, resultOk := i.env.Get("Result")
	methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

	var returnValue Value
	switch {
	case resultOk && resultVal.Type() != "NIL":
		returnValue = resultVal
	case methodNameOk && methodNameVal.Type() != "NIL":
		returnValue = methodNameVal
	case resultOk:
		returnValue = resultVal
	case methodNameOk:
		returnValue = methodNameVal
	default:
		returnValue = &NilValue{}
	}

	if method.ReturnType != nil && returnValue.Type() != "NIL" {
		expectedReturnType := method.ReturnType.String()
		if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
			returnValue = converted
		}
	}

	return returnValue
}

// tryRecordTypeMethodCall attempts to call a static method on a record type.
// Returns nil if the identifier is not a record type.
func (i *Interpreter) tryRecordTypeMethodCall(ident *ast.Identifier, mc *ast.MethodCallExpression) Value {
	recordTypeKey := "__record_type_" + pkgident.Normalize(ident.Value)
	typeVal, ok := i.env.Get(recordTypeKey)
	if !ok {
		return nil
	}

	rtv, ok := typeVal.(*RecordTypeValue)
	if !ok {
		return nil
	}

	methodNameLower := pkgident.Normalize(mc.Method.Value)
	classMethodOverloads, hasOverloads := rtv.ClassMethodOverloads[methodNameLower]

	if !hasOverloads || len(classMethodOverloads) == 0 {
		return i.newErrorWithLocation(mc, "static method '%s' not found in record type '%s'", mc.Method.Value, ident.Value)
	}

	var staticMethod *ast.FunctionDecl
	var err error

	if len(classMethodOverloads) > 1 {
		staticMethod, err = i.resolveMethodOverload(rtv.RecordType.Name, mc.Method.Value, classMethodOverloads, mc.Arguments)
		if err != nil {
			return i.newErrorWithLocation(mc, "%s", err.Error())
		}
	} else {
		staticMethod = classMethodOverloads[0]
	}

	return i.callRecordStaticMethod(rtv, staticMethod, mc.Arguments, mc)
}

// findMethodOwner returns the class in the hierarchy that declares the given method.
// Falls back to the runtime class if not found.
func (i *Interpreter) findMethodOwner(classInfo *ClassInfo, method *ast.FunctionDecl, isClassMethod bool) *ClassInfo {
	if classInfo == nil || method == nil {
		return classInfo
	}

	for c := classInfo; c != nil; c = c.Parent {
		var overloads map[string][]*ast.FunctionDecl
		if isClassMethod {
			overloads = c.ClassMethodOverloads
		} else {
			overloads = c.MethodOverloads
		}

		for _, methods := range overloads {
			for _, m := range methods {
				if m == method {
					return c
				}
			}
		}
	}

	return classInfo
}
