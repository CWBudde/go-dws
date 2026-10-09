package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// analyzeMethodCallExpression analyzes a method call on an object
func (a *Analyzer) analyzeMethodCallExpression(expr *ast.MethodCallExpression) types.Type {
	if expr.Truncated {
		if name, ok := expr.Object.(*ast.Identifier); ok && !a.specialFunctionHasShadow(name.Value) {
			return a.analyzeTruncatedCall(expr.Arguments)
		}
	}
	if name, ok := expr.Object.(*ast.Identifier); ok {
		if _, imported := a.importedUnitNamespace(name.Value); imported {
			if expr.Truncated {
				return a.analyzeTruncatedCall(expr.Arguments)
			}
			return a.analyzeCallExpression(&ast.CallExpression{
				BaseNode:  expr.BaseNode,
				Function:  &ast.MemberAccessExpression{BaseNode: expr.BaseNode, Object: expr.Object, Member: expr.Method},
				Arguments: expr.Arguments,
			})
		}
	}
	// JSON namespace method call: JSON.Parse(s), JSON.Stringify(x). Recognized
	// before `JSON` is analyzed as an ordinary (undefined) identifier.
	if a.isJSONNamespace(expr.Object) {
		if expr.Truncated {
			return a.analyzeTruncatedCall(expr.Arguments)
		}
		return a.analyzeJSONNamespaceResult(expr.Method.Value, expr.Arguments)
	}
	if a.isDefaultNamespace(expr.Object) {
		if expr.Truncated {
			return a.analyzeTruncatedCall(expr.Arguments)
		}
		builtinCall := &ast.CallExpression{
			BaseNode:  ast.BaseNode{Token: expr.Token},
			Function:  expr.Method,
			Arguments: expr.Arguments,
		}
		return a.analyzeDefaultNamespaceCall(expr.Object, expr.Method, builtinCall)
	}
	if a.stopUnitQualifiedSpecialName(expr.Object, expr.Method) {
		return nil
	}

	// Analyze the object expression
	objectType := a.analyzeProbedReceiver(expr.Object)
	if helper, ok := objectType.(*types.HelperType); ok {
		if expr.Truncated {
			return a.analyzeTruncatedCall(expr.Arguments)
		}
		if result, handled := a.analyzeExplicitHelperCall(helper, expr.Method, expr.Arguments); handled {
			return result
		}
		a.addCompilerStop(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, helper.Name))
		return nil
	}
	if objectType == nil {
		// An overload set deliberately carries no type of its own, so a
		// routine name used as a receiver reads as nil here. When the set
		// holds a unique parameterless overload returning a set, the mutator
		// checks below still apply to that call's temporary result, so
		// `Make.Exclude(e)` is rejected just like `Test.Exclude(e)`.
		implicit := a.implicitCallTypePreview(expr.Object, nil)
		if _, isSet := types.GetUnderlyingType(implicit).(*types.SetType); !isSet {
			// Error already reported
			return nil
		}
		objectType = a.applyImplicitCallType(expr.Object, nil)
	}

	// Probe ordinary implicit receivers and aliases before classifying property
	// punctuation. The equivalent member read records any implicit invocation.
	propertyReceiver := a.implicitCallTypePreview(expr.Object, objectType)
	if propertyReceiver == nil {
		propertyReceiver = objectType
	}
	var propertyClass *types.ClassType
	if class, ok := types.GetUnderlyingType(propertyReceiver).(*types.ClassType); ok {
		propertyClass = class
	}
	if meta, ok := types.GetUnderlyingType(propertyReceiver).(*types.ClassOfType); ok {
		propertyClass = meta.ClassType
	}
	if propertyClass != nil && a.hasHelperMethod(propertyReceiver, expr.Method.Value) == nil {
		if result, handled := a.analyzePropertyCompatibilityRead(expr, propertyClass); handled {
			return result
		}
	}
	if expr.Truncated {
		return a.analyzeTruncatedCall(expr.Arguments)
	}

	// Method call on a JSONVariant receiver: v.TypeName(), v.Add(x), ...
	if types.IsJSONVariant(objectType) {
		return a.analyzeJSONMethodResult(expr.Method.Value, expr.Arguments)
	}

	// Method call on a ByteBuffer receiver: b.SetLength(4), b.GetByte(0), ...
	if types.IsByteBuffer(objectType) {
		return a.analyzeByteBufferMethodResult(expr.Method.Value, expr.Arguments)
	}

	methodName := expr.Method.Value
	methodNameLower := ident.Normalize(methodName)
	isMetaclass := false

	// Handle metaclass type (class of T) for constructor calls.
	// When we have TExample.CreateWith(...), unwrap ClassOfType to ClassType for constructor lookup.
	// An alias of a class reference (`type TMeta = class of TObject`) is one too,
	// which is how analyzeMemberAccessExpression reads it as well.
	if metaclassType, ok := types.GetUnderlyingType(objectType).(*types.ClassOfType); ok {
		isMetaclass = true
		if metaclassType.ClassType != nil {
			objectType = metaclassType.ClassType
		}
	}

	// Check if object is an interface type
	if interfaceType, ok := objectType.(*types.InterfaceType); ok {
		// Look up method in interface (including inherited methods from parent interfaces)
		methodType, found := interfaceType.GetMethod(methodNameLower)

		// Check parent interfaces
		if !found && interfaceType.Parent != nil {
			allMethods := types.GetAllInterfaceMethods(interfaceType)
			methodType, found = allMethods[methodNameLower]
		}

		if !found {
			helperMethod, found := a.resolveHelperMethodForCall(objectType, expr.Method, expr.Arguments)
			if !found {
				a.addCompilerStop(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
				return nil
			}
			if helperMethod == nil {
				return nil
			}
			methodType = helperMethod
			a.addIdentifierCaseHint(expr.Method, a.declaredHelperCallName(objectType, methodName))
		} else {
			a.addIdentifierCaseHint(expr.Method, a.declaredInterfaceMethodName(interfaceType, methodName))
			// Native interface methods keep Self outside the written arguments,
			// like class methods. Read children and check supplied types before
			// reporting a count error.
			a.analyzeMemberCallArguments(methodType, expr.Arguments, expr.Method.Token.Pos, false)
			return methodType.ReturnType
		}

		a.analyzeHelperCallArguments(objectType, methodName, methodType, expr.Arguments, expr.Method.Token.Pos)

		return methodType.ReturnType
	}

	// Check if object is a class type
	classType, ok := objectType.(*types.ClassType)
	if !ok {
		if assoc, isAssoc := types.GetUnderlyingType(objectType).(*types.AssociativeArrayType); isAssoc {
			if result := a.analyzeAssociativeArrayMethodCall(expr, assoc); result != nil {
				return result
			}
		}
		if arrayType, isArray := types.GetUnderlyingType(objectType).(*types.ArrayType); isArray {
			if result := a.analyzeArrayMethodCall(expr, arrayType); result != nil {
				return result
			}
		}

		// Check if object is a record type with methods
		if recordType, isMeta := recordReceiverType(objectType); recordType != nil {
			a.addIdentifierCaseHint(expr.Method, a.declaredRecordMethodName(recordType, methodName))

			// Class-side methods can be called through either a record meta value
			// or an instance; instance methods only join the latter's candidate set.
			overloads := recordType.GetClassMethodOverloads(methodNameLower)
			if !isMeta {
				overloads = append(append([]*types.MethodInfo{}, overloads...), recordType.GetMethodOverloads(methodNameLower)...)
			}
			if len(overloads) > 0 {
				return a.analyzeRecordCall(overloads, expr.Arguments, methodName, expr.Method.Token.Pos)
			}
			var method *types.FunctionType
			if method == nil {
				// A proc-typed record field is directly callable: rec.Proc1(x).
				if fieldType := recordType.GetFieldType(methodNameLower); !isMeta && fieldType != nil && isFunctionPointerType(fieldType) {
					if a.semanticInfo != nil && expr.Method != nil {
						a.semanticInfo.SetType(expr.Method, &ast.TypeAnnotation{
							Token: expr.Method.Token,
							Name:  types.GetUnderlyingType(fieldType).String(),
						})
						a.semanticInfo.SetResolvedType(expr.Method, fieldType)
					}
					return a.analyzeFunctionPointerCallArgs(expr.Arguments, fieldType, expr.Token.Pos)
				}
				// Method not found in record, check if a helper provides it
				helperMethod, found := a.resolveHelperMethodForCall(objectType, expr.Method, expr.Arguments)
				if !found {
					a.addCompilerStop(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
					return nil
				}
				if helperMethod == nil {
					return nil
				}
				// Use the helper method
				method = helperMethod
				a.addIdentifierCaseHint(expr.Method, a.declaredHelperCallName(objectType, methodName))
			}

			a.analyzeHelperCallArguments(objectType, methodName, method, expr.Arguments, expr.Method.Token.Pos)

			return method.ReturnType
		}

		// A parameterless routine named as the receiver is called first, so
		// `Test.Exclude(e)` reaches the set the routine returns — as a
		// temporary, which the mutator check below then rejects
		// (SetOfFail/test_non_variable).
		setReceiverType := objectType
		if _, isSet := types.GetUnderlyingType(setReceiverType).(*types.SetType); !isSet {
			// Probe without side effects first: applyImplicitCallType records
			// an identifier-case hint, which must not be left behind for a
			// receiver whose implicit result is not a set at all.
			implicit := a.implicitCallTypePreview(expr.Object, objectType)
			if _, isSet := types.GetUnderlyingType(implicit).(*types.SetType); isSet {
				setReceiverType = a.applyImplicitCallType(expr.Object, objectType)
			}
		}

		// Handle set types with built-in methods (Include/Exclude) without helpers
		if setType, isSet := types.GetUnderlyingType(setReceiverType).(*types.SetType); isSet {
			switch methodNameLower {
			case "include", "exclude":
				// Both mutate the set in place, so the receiver must be a
				// writable variable — not a constant, and not a call's result.
				if !a.isMutableSetReceiver(expr.Object, expr.Method.Token.Pos) {
					return types.VOID
				}

				if len(expr.Arguments) != 1 {
					a.addError("set method '%s' expects 1 argument, got %d at %s",
						methodName, len(expr.Arguments), expr.Token.Pos.String())
					return types.VOID
				}

				expectedElemType := setType.ElementType
				a.analyzeCallArgument(0, expr.Arguments[0], expectedElemType)
				return types.VOID
			default:
				a.addCompilerStop(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value,
					a.setTypeDiagnosticName(setReceiverType)))
				return nil
			}
		}

		// Receiver eligibility is established before overload selection upstream.
		a.addIdentifierCaseHint(expr.Method, a.declaredHelperCallName(objectType, methodName))
		invalidHelperReceiver := a.isTypeMetaValueExpression(expr.Object) &&
			a.helperCallOwner(objectType, methodName) != nil && !a.isHelperCallClassMethod(objectType, methodName)
		if invalidHelperReceiver {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
		}

		// Check if helpers provide this method for non-class, non-record types
		helperMethod, found := a.resolveHelperMethodForCall(objectType, expr.Method, expr.Arguments)
		if !found {
			a.addCompilerStop(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
			return nil
		}

		if helperMethod == nil {
			return nil
		}

		if invalidHelperReceiver {
			return nil
		}

		a.addIdentifierCaseHint(expr.Method, a.declaredHelperCallName(objectType, methodName))

		// Record the receiver's static type so runtime helper dispatch honors
		// alias-specific (strict) helpers over the underlying type's helpers.
		if a.semanticInfo != nil && expr.Method != nil && !a.hasHelperCallBinding(expr.Method) {
			a.semanticInfo.SetType(expr.Method, &ast.TypeAnnotation{
				Token: expr.Method.Token,
				Name:  "__helper_receiver:" + objectType.String(),
			})
		}

		a.analyzeHelperCallArguments(objectType, methodName, helperMethod, expr.Arguments, expr.Method.Token.Pos)

		return helperMethod.ReturnType
	}

	a.addIdentifierCaseHint(expr.Method, a.declaredClassMemberName(classType, methodName))

	// Handle built-in methods available on all objects (inherited from TObject)
	if methodName == "ClassName" {
		// ClassName() returns String
		return types.STRING
	}

	// Constructors are stored separately from methods and can be inherited.
	// The hierarchy lookup merges constructors with same-named class methods,
	// so the resolved overload decides whether this call is a construction.
	if constructorOverloads := a.getMethodOverloadsInHierarchy(methodName, classType); len(constructorOverloads) > 0 && classType.HasConstructor(methodName) {
		selectedInfo := a.selectMemberCallOverload(constructorOverloads, expr.Arguments, methodName, expr.Method.Token.Pos)
		if selectedInfo == nil {
			return classType
		}
		methodType := selectedInfo.Signature
		a.analyzeClassCallArguments(methodType, expr.Arguments, expr.Method.Token.Pos)

		// Resolved to a same-named class method rather than a constructor.
		if !selectedInfo.IsConstructor {
			return methodType.ReturnType
		}

		if classType.IsAbstract {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		if unimplementedMethods := a.getUnimplementedAbstractMethods(classType); len(unimplementedMethods) > 0 {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		return classType
	}

	// Check if method is overloaded
	var methodType *types.FunctionType
	var selectedOverload *types.MethodInfo
	methodOwner := a.getMethodOwner(classType, methodName)
	overloads := a.getMethodOverloadsInHierarchy(methodName, classType)

	switch {
	case len(overloads) > 1 || (len(overloads) == 1 && overloads[0].HasOverloadDirective):
		// Method is overloaded - resolve based on argument types
		// Analyze argument types first
		argTypes := make([]types.Type, len(expr.Arguments))
		valueTypes := make([]types.Type, len(expr.Arguments))
		hasImplicitCallable := false
		failed := false
		for i, arg := range expr.Arguments {
			argType := a.analyzeOverloadArgument(arg)
			if argType == nil {
				failed = true
			}
			argTypes[i] = argType
			valueTypes[i] = argType
			if identifier, ok := arg.(*ast.Identifier); ok {
				if sym, found := a.symbols.Resolve(identifier.Value); found {
					if fn, ok := sym.Type.(*types.FunctionType); ok && len(fn.Parameters) == 0 && fn.ReturnType != nil {
						valueTypes[i] = fn.ReturnType
						hasImplicitCallable = true
					}
				}
			}
		}

		if failed {
			return nil
		}

		// Convert MethodInfo to Symbol for ResolveOverload
		candidates := make([]*Symbol, len(overloads))
		for i, overload := range overloads {
			candidates[i] = &Symbol{
				Type: overload.Signature,
			}
		}

		// Resolve overload based on argument types
		var selected *Symbol
		var err error
		if hasImplicitCallable {
			selected, err = ResolveOverload(candidates, valueTypes)
		}
		if !hasImplicitCallable || err != nil {
			selected, err = ResolveOverload(candidates, argTypes)
		}
		if err != nil {
			diagnostic := NewNoOverloadMatchError(expr.Method.Token.Pos, methodName)
			diagnostic.AfterChildren = true
			a.addStructuredError(diagnostic)
			return nil
		}

		var ok bool
		methodType, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("internal error: expected function type for selected overloaded method, but got %T", selected.Type)
			return nil
		}
		for i := range candidates {
			if candidates[i] == selected {
				selectedOverload = overloads[i]
				break
			}
		}
		if isMetaclass && selectedOverload != nil && !selectedOverload.IsClassMethod {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
			return nil
		}
	case len(overloads) == 1:
		// Single method (not overloaded). A method reached through a metaclass value must
		// be a class method or constructor. Use the resolved overload's flag so inherited
		// class methods (not present in this class's own ClassMethodFlags map) are accepted.
		if isMetaclass && !overloads[0].IsClassMethod {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
			return nil
		}
		methodType = overloads[0].Signature
	default:
		// Method not found - check helpers
		a.addIdentifierCaseHint(expr.Method, a.declaredHelperCallName(objectType, methodName))
		invalidHelperReceiver := isMetaclass && a.helperCallOwner(objectType, methodName) != nil &&
			!a.isHelperCallClassMethod(objectType, methodName)
		if invalidHelperReceiver {
			a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
		}
		helperMethod, found := a.resolveHelperMethodForCall(objectType, expr.Method, expr.Arguments)
		if invalidHelperReceiver || (found && helperMethod == nil) {
			return nil
		}
		if isMetaclass {
			// Metaclass target availability remains a separate lookup audit.
			if !found || !a.isHelperCallClassMethod(objectType, methodName) {
				a.addStructuredError(NewClassMethodOrConstructorExpectedError(expr.Method.Token.Pos))
				return nil
			}
			methodType = helperMethod
		} else if found {
			methodType = helperMethod
		} else if callableType := a.classCallableMemberType(classType, methodName); callableType != nil {
			// A private/protected proc-typed field or class var must not become
			// callable from outside its scope just by adding parentheses.
			if !a.classCallableMemberVisible(classType, methodName, expr.Method) {
				return nil
			}
			// A field, class var, or readable property of function-pointer type is
			// directly callable: o.FEvent(1). Annotate so the evaluator reads the
			// stored pointer instead of dispatching a same-named method.
			if a.semanticInfo != nil && expr.Method != nil {
				a.semanticInfo.SetType(expr.Method, &ast.TypeAnnotation{
					Token: expr.Method.Token,
					Name:  types.GetUnderlyingType(callableType).String(),
				})
				a.semanticInfo.SetResolvedType(expr.Method, callableType)
			}
			return a.analyzeFunctionPointerCallArgs(expr.Arguments, callableType, expr.Token.Pos)
		} else {
			a.addCompilerStop(NewAccessibleMemberError(expr.Method.Token.Pos, expr.Method.Value, objectType.String()))
			return nil
		}
	}

	// Check method visibility. Overloads carry their own visibility, so the
	// SELECTED overload's visibility governs (overloads of one name can mix
	// private and public sections).
	if methodOwner != nil && len(overloads) > 0 {
		// Use lowercase key for case-insensitive lookup
		visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(methodName)]
		if selectedOverload != nil {
			visibility, hasVisibility = selectedOverload.Visibility, true
		}
		if hasVisibility && !a.checkVisibility(methodOwner, visibility, methodName, "method") {
			a.addStructuredError(NewVisibilityScopeError(expr.Method.Token.Pos, expr.Method.Value))
			if methodOwner.HasConstructor(methodName) {
				return classType
			}
			return methodType.ReturnType
		}
	}

	if len(overloads) == 0 {
		a.analyzeHelperCallArguments(objectType, methodName, methodType, expr.Arguments, expr.Method.Token.Pos)
	} else {
		a.analyzeClassCallArguments(methodType, expr.Arguments, expr.Method.Token.Pos)
	}

	if classType.HasConstructor(methodName) {
		// Check if trying to instantiate an abstract class via constructor call
		if classType.IsAbstract {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		// Check if class has unimplemented abstract methods
		unimplementedMethods := a.getUnimplementedAbstractMethods(classType)
		if len(unimplementedMethods) > 0 {
			a.addStructuredError(NewAbstractInstantiationError(expr.Token.Pos))
			return classType
		}

		return classType
	}
	return methodType.ReturnType
}
