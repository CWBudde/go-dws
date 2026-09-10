package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Function Pointer Type Declaration Analysis
// ============================================================================

// analyzeFunctionPointerTypeDeclaration analyzes a function pointer type declaration.
// Examples: type TComparator = function(a, b: Integer): Integer;
func (a *Analyzer) analyzeFunctionPointerTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil || decl.FunctionPointerType == nil {
		return
	}

	fpType := decl.FunctionPointerType

	// Check for duplicate parameter names (skip shorthand syntax with nil names)
	paramNames := make(map[string]bool)
	for _, param := range fpType.Parameters {
		if param.Name == nil {
			continue
		}
		if paramNames[param.Name.Value] {
			a.addError("duplicate parameter name '%s' in function pointer type at %s",
				param.Name.Value, param.Name.Token.Pos.String())
			return
		}
		paramNames[param.Name.Value] = true
	}

	// Validate all parameter types exist
	paramTypes := make([]types.Type, 0, len(fpType.Parameters))
	for _, param := range fpType.Parameters {
		paramType, err := a.resolveType(getTypeExpressionName(param.Type))
		if err != nil {
			a.addError("unknown parameter type '%s' in function pointer type at %s",
				getTypeExpressionName(param.Type), param.Type.Pos().String())
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Validate return type (for functions)
	var returnType types.Type
	if fpType.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(getTypeExpressionName(fpType.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' in function pointer type at %s",
				getTypeExpressionName(fpType.ReturnType), fpType.ReturnType.Pos().String())
			return
		}
	}

	// Create the function pointer type
	var funcPtrType types.Type
	if fpType.OfObject {
		funcPtrType = types.NewMethodPointerType(paramTypes, returnType)
	} else {
		funcPtrType = types.NewFunctionPointerType(paramTypes, returnType)
	}

	// Register in the function pointers map
	if a.functionPointers == nil {
		a.functionPointers = make(map[string]*types.FunctionPointerType)
	}
	if methodPtr, ok := funcPtrType.(*types.MethodPointerType); ok {
		a.functionPointers[decl.Name.Value] = &methodPtr.FunctionPointerType
	} else if funcPtr, ok := funcPtrType.(*types.FunctionPointerType); ok {
		a.functionPointers[decl.Name.Value] = funcPtr
	}

	// Register as a type alias so resolveType can find it
	typeAlias := &types.TypeAlias{
		Name:        decl.Name.Value,
		AliasedType: funcPtrType,
	}
	a.registerTypeWithPos(decl.Name.Value, typeAlias, decl.Token.Pos)
}

// ============================================================================
// Address-of Expression Analysis
// ============================================================================

// analyzeAddressOfExpression analyzes an address-of expression (@FunctionName).
// Examples: @Ascending, @MyCallback, @TMyClass.MyMethod
func (a *Analyzer) analyzeAddressOfExpression(expr *ast.AddressOfExpression) types.Type {
	if expr == nil || expr.Operator == nil {
		return nil
	}

	// The operator should be an identifier or member access expression
	switch target := expr.Operator.(type) {
	case *ast.Identifier:
		// Simple function/procedure reference: @FunctionName
		return a.analyzeAddressOfFunction(target.Value, expr)

	case *ast.MemberAccessExpression:
		// Bound method reference: @instance.MethodName
		return a.analyzeAddressOfMethod(target, expr)

	default:
		a.addError("address-of operator (@) requires a function or procedure name at %s",
			expr.Token.Pos.String())
		return nil
	}
}

// analyzeAddressOfMethod resolves a bound method reference (@instance.Method)
// and produces a method pointer type describing the method's signature.
//
// Only instance receivers are supported: taking the address of a method through
// a class name (@TClass.Method) yields an unbound method reference, which is a
// separate feature and is rejected here with an explicit diagnostic.
func (a *Analyzer) analyzeAddressOfMethod(target *ast.MemberAccessExpression, expr *ast.AddressOfExpression) types.Type {
	if target.Member == nil {
		a.addError("address-of operator (@) requires a method name at %s", expr.Token.Pos.String())
		return nil
	}

	objectType := a.analyzeExpression(target.Object)
	if objectType == nil {
		// The receiver already produced a diagnostic; do not pile on.
		return nil
	}

	classType, isMetaclass, ok := addressOfReceiverClass(objectType)
	if !ok {
		a.addError("address-of operator (@) requires an object instance to bind a method, got %s at %s",
			objectType.String(), expr.Token.Pos.String())
		return nil
	}

	methodName := target.Member.Value

	// TObject's intrinsic class members (ClassName, ClassType) are not ordinary
	// methods; @TObject.ClassType captures one as a parameterless pointer.
	if ptrType, isIntrinsic := a.intrinsicMemberPointerType(classType, methodName, nil); isIntrinsic {
		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{Name: ptrType.String()})
		a.semanticInfo.SetResolvedType(expr, ptrType)
		return ptrType
	}

	if isMetaclass && !a.isClassMethodInHierarchy(classType, ident.Normalize(methodName)) {
		a.addError("unbound method pointers (@TClass.%s) are not supported at %s",
			methodName, expr.Token.Pos.String())
		return nil
	}

	method := a.firstBindableMethodOverload(methodName, classType, isMetaclass)
	if method == nil {
		a.addError("'%s' is not a method of class '%s' at %s",
			methodName, classType.Name, expr.Token.Pos.String())
		return nil
	}

	// Taking a method's address is a member access and obeys the same visibility
	// rules: @obj.PrivateMethod must be rejected wherever obj.PrivateMethod() is.
	methodOwner := a.getMethodOwner(classType, methodName)
	if methodOwner != nil {
		visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(methodName)]
		if hasVisibility && !a.checkVisibility(methodOwner, visibility, methodName, "method") {
			a.addStructuredError(NewVisibilityScopeError(target.Member.Token.Pos, target.Member.Value))
			return nil
		}
	}

	a.recordClassMethodUsage(classType, methodName)

	var returnType types.Type
	if method.Signature.ReturnType != nil && method.Signature.ReturnType != types.VOID {
		returnType = method.Signature.ReturnType
	}

	methodPtrType := types.NewMethodPointerType(method.Signature.Parameters, returnType)
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("method pointer to %s.%s", classType.Name, methodName),
	}
	a.semanticInfo.SetType(expr, typeAnnotation)

	return methodPtrType
}

// addressOfReceiverClass resolves the class an address-of receiver binds against.
// A metaclass receiver (@TClass.Method) yields its class with isMetaclass set,
// which restricts the reference to the class side.
func addressOfReceiverClass(objectType types.Type) (classType *types.ClassType, isMetaclass bool, ok bool) {
	underlying := types.GetUnderlyingType(objectType)
	if metaclass, isMeta := underlying.(*types.ClassOfType); isMeta {
		if metaclass.ClassType == nil {
			return nil, true, false
		}
		return metaclass.ClassType, true, true
	}
	classType, ok = underlying.(*types.ClassType)
	return classType, false, ok
}

// firstBindableMethodOverload returns the overload a method pointer binds to. A
// pointer cannot represent an overload set, so the runtime and the analyzer agree
// on the first declared non-constructor overload.
//
// classMethodsOnly restricts the candidates to the class side, which is what a
// metaclass receiver (@TClass.M) can bind and all the runtime's
// CreateClassMethodPointer looks at; without it a same-named instance overload
// declared first would be recorded here while the runtime binds a class one.
func (a *Analyzer) firstBindableMethodOverload(
	methodName string,
	classType *types.ClassType,
	classMethodsOnly bool,
) *types.MethodInfo {
	for _, candidate := range a.getMethodOverloadsInHierarchy(methodName, classType) {
		if candidate == nil || candidate.Signature == nil || candidate.IsConstructor {
			continue
		}
		if classMethodsOnly && !candidate.IsClassMethod {
			continue
		}
		return candidate
	}
	return nil
}

// analyzeAddressOfFunction resolves a function name and creates a function pointer type.
// Queries both symbol table and builtin registry.
func (a *Analyzer) analyzeAddressOfFunction(funcName string, expr *ast.AddressOfExpression) types.Type {
	sym, ok := a.symbols.Resolve(funcName)
	if !ok {
		// Query builtin registry for function signatures
		if sig, found := builtins.DefaultRegistry.GetSignature(funcName); found {
			if sig.IsVariadic {
				a.addError("cannot take address of variadic built-in function '%s' at %s",
					funcName, expr.Token.Pos.String())
				return nil
			}
			// Builtins with optional parameters allowed; validated at call time
			return a.buildFunctionPointerTypeFromBuiltin(funcName, sig, expr)
		}

		// Builtin without signature metadata - return generic function pointer type
		if builtins.DefaultRegistry.Has(funcName) {
			funcPtrType := types.NewFunctionPointerType(nil, types.VARIANT)
			typeAnnotation := &ast.TypeAnnotation{
				Name: fmt.Sprintf("function pointer to %s", funcName),
			}
			a.semanticInfo.SetType(expr, typeAnnotation)
			return funcPtrType
		}

		a.addStructuredError(NewUnknownNameError(expr.Token.Pos, funcName))
		return nil
	}

	// Taking the address of a variable that already holds a function pointer is
	// the identity: `@f` and `f` denote the same value. Reading it through `@`
	// counts as a use, or the variable would draw an "unused" hint.
	if funcPtrType, ok := types.GetUnderlyingType(sym.Type).(*types.FunctionPointerType); ok {
		a.recordSymbolUsage(funcName, expr.Token.Pos)
		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
			Name: fmt.Sprintf("function pointer to %s", funcName),
		})
		return funcPtrType
	}

	// The symbol must be a function type
	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("'%s' is not a function or procedure (got %s) at %s",
			funcName, sym.Type.String(), expr.Token.Pos.String())
		return nil
	}

	return a.buildFunctionPointerType(funcName, funcType, expr)
}

// buildFunctionPointerTypeFromBuiltin creates a FunctionPointerType from a builtin signature.
func (a *Analyzer) buildFunctionPointerTypeFromBuiltin(funcName string, sig *builtins.FunctionSignature, expr *ast.AddressOfExpression) types.Type {
	var returnType types.Type
	if sig.ReturnType != nil && sig.ReturnType != types.VOID {
		returnType = sig.ReturnType
	}

	funcPtrType := types.NewFunctionPointerType(sig.ParamTypes, returnType)
	// Builtins may declare optional trailing parameters. Record the required
	// count so a call through the pointer keeps every arity the builtin itself
	// accepts instead of demanding the fully expanded parameter list.
	funcPtrType.MinArgs = sig.MinArgs
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("function pointer to %s", funcName),
	}
	a.semanticInfo.SetType(expr, typeAnnotation)

	return funcPtrType
}

// buildFunctionPointerType creates a FunctionPointerType from a function signature.
func (a *Analyzer) buildFunctionPointerType(funcName string, funcType *types.FunctionType, expr *ast.AddressOfExpression) types.Type {
	var returnType types.Type
	if funcType.ReturnType != nil && funcType.ReturnType != types.VOID {
		returnType = funcType.ReturnType
	}

	funcPtrType := types.NewFunctionPointerType(funcType.Parameters, returnType)
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("function pointer to %s", funcName),
	}
	a.semanticInfo.SetType(expr, typeAnnotation)

	return funcPtrType
}

// ============================================================================
// Function Pointer Assignment Validation
// ============================================================================

// NOTE: Function pointer assignment validation is handled by the IsCompatibleWith
// method in the types package, which is called from analyzer.canAssign().
// This provides automatic compatibility checking for all assignment types.

// ============================================================================
// Function Pointer Call Validation
// ============================================================================

// analyzeFunctionPointerCall analyzes a call to a function pointer variable.
// Validates argument types and infers the return type from the function pointer.
func (a *Analyzer) analyzeFunctionPointerCall(callExpr *ast.CallExpression, calleeType types.Type) types.Type {
	return a.analyzeFunctionPointerCallArgs(callExpr.Arguments, calleeType, callExpr.Token.Pos)
}

// analyzeFunctionPointerCallArgs validates a call to a function/method pointer
// value given the raw argument list and a call position. It is shared by the
// bare-variable call path (analyzeFunctionPointerCall) and the member-call path
// (a proc-typed field/property invoked as o.FEvent(x)).
func (a *Analyzer) analyzeFunctionPointerCallArgs(args []ast.Expression, calleeType types.Type, pos token.Position) types.Type {
	underlyingType := types.GetUnderlyingType(calleeType)

	// Extract function pointer type (could be FunctionPointerType or MethodPointerType)
	var funcPtr *types.FunctionPointerType
	if fp, ok := underlyingType.(*types.FunctionPointerType); ok {
		funcPtr = fp
	} else if mp, ok := underlyingType.(*types.MethodPointerType); ok {
		funcPtr = &mp.FunctionPointerType
	} else {
		return nil
	}

	// Validate argument count. Pointers to builtins with optional parameters
	// accept any arity between the required count and the full parameter list.
	required := funcPtr.RequiredParamCount()
	if len(args) < required || len(args) > len(funcPtr.Parameters) {
		expected := fmt.Sprintf("%d", len(funcPtr.Parameters))
		if required != len(funcPtr.Parameters) {
			expected = fmt.Sprintf("%d to %d", required, len(funcPtr.Parameters))
		}
		a.addError("function pointer call argument count mismatch at %s: expected %s arguments, got %d",
			pos.String(), expected, len(args))
		return nil
	}

	// Validate each argument type
	for i, arg := range args {
		argType := a.analyzeExpressionWithExpectedType(arg, funcPtr.Parameters[i])
		if argType == nil {
			continue
		}
		expectedType := funcPtr.Parameters[i]
		if !a.canAssign(argType, expectedType) {
			a.addError("function pointer call argument %d type mismatch at %s: expected %s, got %s",
				i+1, pos.String(), expectedType.String(), argType.String())
		}
	}

	if funcPtr.ReturnType != nil {
		return funcPtr.ReturnType
	}
	return types.VOID
}

// classCallableMemberType returns the function/method pointer type of a field,
// class var, or readable property named memberName on classType, or nil if the
// member is absent or not of pointer type. Used so a proc-typed member can be
// invoked directly: o.FEvent(1).
func (a *Analyzer) classCallableMemberType(classType *types.ClassType, memberName string) types.Type {
	if fieldType, found := classType.GetField(memberName); found {
		if isFunctionPointerType(fieldType) {
			return fieldType
		}
	}
	if classVarType, found := classType.GetClassVar(memberName); found {
		if isFunctionPointerType(classVarType) {
			return classVarType
		}
	}
	if propInfo, found := classType.GetProperty(memberName); found && propInfo.Type != nil {
		if propInfo.ReadKind != types.PropAccessNone && isFunctionPointerType(propInfo.Type) {
			return propInfo.Type
		}
	}
	return nil
}

// classCallableMemberVisible reports whether the callable proc-typed member
// (field or class var) named memberName on classType — as resolved by
// classCallableMemberType — is accessible from the current scope. It mirrors the
// field/class-var visibility rules of ordinary member access so a private or
// protected proc-typed member cannot be invoked from outside its declaring scope
// merely by adding parentheses (o.FEvent(...)). A visibility diagnostic is
// emitted at member's position when access is denied. Properties are not
// visibility-checked here, matching ordinary property member access.
func (a *Analyzer) classCallableMemberVisible(classType *types.ClassType, memberName string, member *ast.Identifier) bool {
	normalized := ident.Normalize(memberName)
	if _, found := classType.GetField(memberName); found {
		if owner := a.getFieldOwner(classType, memberName); owner != nil {
			if visibility, ok := owner.FieldVisibility[normalized]; ok &&
				!a.checkVisibility(owner, visibility, memberName, "field") {
				a.addStructuredError(NewVisibilityScopeError(member.Token.Pos, member.Value))
				return false
			}
		}
		return true
	}
	if _, found := classType.GetClassVar(memberName); found {
		if owner := a.getClassVarOwner(classType, memberName); owner != nil {
			if visibility, ok := owner.ClassVarVisibility[normalized]; ok &&
				!a.checkVisibility(owner, visibility, memberName, "class variable") {
				a.addStructuredError(NewVisibilityScopeError(member.Token.Pos, member.Value))
				return false
			}
		}
		return true
	}
	return true
}

// isFunctionPointerType reports whether t (after alias resolution) is a function
// or method pointer type.
func isFunctionPointerType(t types.Type) bool {
	return types.IsPointerType(t)
}

// ============================================================================
// Intrinsic Class Member Pointers (ClassName / ClassType)
// ============================================================================

// intrinsicClassMemberNaturalType returns the result type of one of TObject's
// intrinsic, parameterless class members when captured on classType:
//
//	ClassName -> String
//	ClassType -> class of <classType>
//
// It returns nil for any other member name, or when the class (or an ancestor)
// declares a real method of that name, which then owns the reference.
func (a *Analyzer) intrinsicClassMemberNaturalType(classType *types.ClassType, memberName string) types.Type {
	if classType == nil {
		return nil
	}
	switch {
	case ident.Equal(memberName, "ClassName"):
		// TObject's ClassName is registered as a synthesized overload; a
		// user-declared override is a real method and takes precedence.
		if info := a.firstNonSynthesizedMethod(classType, memberName); info != nil {
			return nil
		}
		return types.STRING
	case ident.Equal(memberName, "ClassType"):
		if info := a.firstNonSynthesizedMethod(classType, memberName); info != nil {
			return nil
		}
		return types.NewClassOfType(classType)
	default:
		return nil
	}
}

// firstNonSynthesizedMethod returns the first user-declared (non-synthesized)
// overload of memberName in classType's hierarchy, or nil when the name is only
// backed by a synthesized built-in.
func (a *Analyzer) firstNonSynthesizedMethod(classType *types.ClassType, memberName string) *types.MethodInfo {
	for _, candidate := range a.getMethodOverloadsInHierarchy(memberName, classType) {
		if candidate != nil && !candidate.IsSynthesized {
			return candidate
		}
	}
	return nil
}

// intrinsicMemberPointerType forms the parameterless function pointer produced by
// capturing an intrinsic class member (ClassName, ClassType) in a context that
// expects a pointer.
//
// The declared target type is adopted for the result when the intrinsic's own
// result is assignable to it: `TClassA.ClassType` yields `class of TClassA`, and
// storing it in a `function : TClass` slot must keep the slot's signature so the
// exact signature match performed by pointer assignability succeeds.
//
// `expected` may be nil (an inferred `var p := @TObject.ClassType`), in which
// case the intrinsic's natural signature is used.
func (a *Analyzer) intrinsicMemberPointerType(classType *types.ClassType, memberName string, expected types.Type) (*types.FunctionPointerType, bool) {
	naturalType := a.intrinsicClassMemberNaturalType(classType, memberName)
	if naturalType == nil {
		return nil, false
	}

	expectedPtr := parameterlessPointerTarget(expected)
	if expected != nil && expectedPtr == nil {
		return nil, false
	}
	if expectedPtr == nil {
		return types.NewFunctionPointerType(nil, naturalType), true
	}
	if expectedPtr.ReturnType == nil || !a.canAssign(naturalType, expectedPtr.ReturnType) {
		return nil, false
	}
	return types.NewFunctionPointerType(nil, expectedPtr.ReturnType), true
}

// parameterlessPointerTarget returns the function pointer signature behind a
// function or method pointer type, but only when it takes no parameters.
// Returns nil for every other type.
func parameterlessPointerTarget(t types.Type) *types.FunctionPointerType {
	if t == nil {
		return nil
	}
	var ptr *types.FunctionPointerType
	switch underlying := types.GetUnderlyingType(t).(type) {
	case *types.MethodPointerType:
		ptr = &underlying.FunctionPointerType
	case *types.FunctionPointerType:
		ptr = underlying
	default:
		return nil
	}
	if len(ptr.Parameters) != 0 {
		return nil
	}
	return ptr
}
