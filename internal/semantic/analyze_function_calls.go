package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// analyzeCallExpressionWithContext analyzes a call expression with optional expected type.
// TODO: Use expectedType for overload resolution when implemented.
func (a *Analyzer) analyzeCallExpressionWithContext(expr *ast.CallExpression, expectedType types.Type) types.Type {
	_ = expectedType
	return a.analyzeCallExpression(expr)
}

func (a *Analyzer) analyzeCallExpression(expr *ast.CallExpression) types.Type {
	// Handle member access expressions (method calls like obj.Method())
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		objectType := a.analyzeExpression(memberAccess.Object)
		if objectType == nil {
			return nil
		}

		// Constructor call: TClass.Create(args)
		if classType, isClassType := objectType.(*types.ClassType); isClassType {
			return a.analyzeConstructorCall(expr, classType, memberAccess.Member.Value)
		}

		// Static method call on record: TRecord.Method(args)
		if recordType, isRecordType := objectType.(*types.RecordType); isRecordType {
			return a.analyzeRecordStaticMethodCall(expr, recordType, memberAccess.Member.Value)
		}

		// Constructor via metaclass variable: cls.Create(args)
		if metaclassType, isMetaclassType := objectType.(*types.ClassOfType); isMetaclassType {
			return a.analyzeConstructorCall(expr, metaclassType.ClassType, memberAccess.Member.Value)
		}

		// Normal method call
		methodType := a.analyzeMemberAccessExpression(memberAccess)
		if methodType == nil {
			return nil
		}

		funcType, ok := methodType.(*types.FunctionType)
		if !ok {
			a.addError("cannot call non-function type %s at %s",
				methodType.String(), expr.Token.Pos.String())
			return nil
		}

		if len(expr.Arguments) != len(funcType.Parameters) {
			a.addError("method call expects %d argument(s), got %d at %s",
				len(funcType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
		}

		// Validate argument types
		for i, arg := range expr.Arguments {
			if i >= len(funcType.Parameters) {
				break
			}

			isVar := len(funcType.VarParams) > i && funcType.VarParams[i]
			if isVar && !a.isLValue(arg) {
				a.addError("var parameter %d requires a variable (identifier, array element, or field), got %s at %s",
					i+1, arg.String(), arg.Pos().String())
			}

			paramType := funcType.Parameters[i]
			argType := a.analyzeExpressionWithExpectedType(arg, paramType)
			if argType != nil && !a.canAssign(argType, paramType) {
				a.addError("argument %d has type %s, expected %s at %s",
					i+1, argType.String(), paramType.String(),
					expr.Token.Pos.String())
			}
		}

		return funcType.ReturnType
	}

	// Handle regular function calls (identifier-based)
	if _, isIdent := expr.Function.(*ast.Identifier); !isIdent {
		// Check if callee is a function pointer
		calleeType := a.analyzeExpression(expr.Function)
		if calleeType == nil {
			return nil
		}
		if funcPtrType := a.analyzeFunctionPointerCall(expr, calleeType); funcPtrType != nil {
			return funcPtrType
		}
		a.addError("function call must use identifier or member access at %s", expr.Token.Pos.String())
		return nil
	}

	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		a.addError("function call must use identifier or member access at %s", expr.Token.Pos.String())
		return nil
	}

	sym, ok := a.symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check built-in functions
		if resultType, isBuiltin := a.analyzeBuiltinFunction(funcIdent.Value, expr.Arguments, expr); isBuiltin {
			if declName := a.builtinDeclarationName(funcIdent.Value); declName != "" && declName != funcIdent.Value {
				a.addCaseMismatchHint(funcIdent.Value, declName, funcIdent.Token.Pos)
			}
			return resultType
		}

		// Check implicit Self method call within a class
		if a.currentClass != nil {
			methodNameLower := ident.Normalize(funcIdent.Value)
			methodType, found := a.currentClass.GetMethod(funcIdent.Value)
			if found {
				// Check visibility
				methodOwner := a.getMethodOwner(a.currentClass, methodNameLower)
				if methodOwner != nil {
					visibility, hasVisibility := methodOwner.MethodVisibility[methodNameLower]
					if hasVisibility && !a.checkVisibility(methodOwner, visibility, funcIdent.Value, "method") {
						visibilityStr := ast.Visibility(visibility).String()
						a.addError("cannot call %s method '%s' of class '%s' at %s",
							visibilityStr, funcIdent.Value, methodOwner.Name, expr.Token.Pos.String())
						return nil
					}
				}

				if len(expr.Arguments) != len(methodType.Parameters) {
					a.addError("method '%s' expects %d argument(s), got %d at %s",
						funcIdent.Value, len(methodType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
				}

				for i, arg := range expr.Arguments {
					if i >= len(methodType.Parameters) {
						break
					}

					isVar := len(methodType.VarParams) > i && methodType.VarParams[i]
					if isVar && !a.isLValue(arg) {
						a.addError("var parameter %d requires a variable (identifier, array element, or field), got %s at %s",
							i+1, arg.String(), arg.Pos().String())
					}

					paramType := methodType.Parameters[i]
					argType := a.analyzeExpressionWithExpectedType(arg, paramType)
					if argType != nil && !a.canAssign(argType, paramType) {
						a.addError("argument %d has type %s, expected %s at %s",
							i+1, argType.String(), paramType.String(),
							expr.Token.Pos.String())
					}
				}

				return methodType.ReturnType
			}
		}

		// Check implicit Self record method call
		if a.currentRecord != nil {
			methodNameLower := ident.Normalize(funcIdent.Value)
			overloads := a.currentRecord.GetMethodOverloads(methodNameLower)
			if len(overloads) > 0 {
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
					candidates[i] = &Symbol{Type: overload.Signature}
				}

				selected, err := ResolveOverload(candidates, argTypes)
				if err != nil {
					a.addError("%s", errors.FormatNoOverloadError(funcIdent.Value, expr.Token.Pos.Line, expr.Token.Pos.Column))
					return nil
				}

				if selected == nil || selected.Type == nil {
					return nil
				}
				funcType, ok := selected.Type.(*types.FunctionType)
				if !ok {
					return nil
				}
				return funcType.ReturnType
			}
		}

		// Check implicit Self record class method call
		if a.currentRecord != nil {
			methodNameLower := ident.Normalize(funcIdent.Value)
			overloads := a.currentRecord.GetClassMethodOverloads(methodNameLower)
			if len(overloads) > 0 {
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
					candidates[i] = &Symbol{Type: overload.Signature}
				}

				selected, err := ResolveOverload(candidates, argTypes)
				if err != nil {
					a.addError("%s", errors.FormatNoOverloadError(funcIdent.Value, expr.Token.Pos.Line, expr.Token.Pos.Column))
					return nil
				}

				methodType := selected.Type.(*types.FunctionType)
				for i, arg := range expr.Arguments {
					if i >= len(methodType.Parameters) {
						break
					}
					paramType := methodType.Parameters[i]
					argType := a.analyzeExpressionWithExpectedType(arg, paramType)
					if argType != nil && !a.canAssign(argType, paramType) {
						a.addError("argument %d to class method '%s' has type %s, expected %s at %s",
							i+1, funcIdent.Value, argType.String(), paramType.String(),
							expr.Token.Pos.String())
					}
				}
				return methodType.ReturnType
			}
		}

		// Assert: Boolean condition with optional String message
		if ident.Equal(funcIdent.Value, "Assert") {
			if len(expr.Arguments) < 1 || len(expr.Arguments) > 2 {
				a.addError("function 'Assert' expects 1-2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			condType := a.analyzeExpression(expr.Arguments[0])
			if condType != nil && condType != types.BOOLEAN {
				a.addError("function 'Assert' first argument must be Boolean, got %s at %s",
					condType.String(), expr.Token.Pos.String())
			}
			if len(expr.Arguments) == 2 {
				msgType := a.analyzeExpression(expr.Arguments[1])
				if msgType != nil && msgType != types.STRING {
					a.addError("function 'Assert' second argument must be String, got %s at %s",
						msgType.String(), expr.Token.Pos.String())
				}
			}
			return types.VOID
		}

		// Insert: Insert(source, targetVar, position)
		if ident.Equal(funcIdent.Value, "Insert") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'Insert' expects 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			sourceType := a.analyzeExpression(expr.Arguments[0])
			if sourceType != nil && sourceType != types.STRING {
				a.addError("function 'Insert' first argument must be String, got %s at %s",
					sourceType.String(), expr.Token.Pos.String())
			}
			if _, ok := expr.Arguments[1].(*ast.Identifier); !ok {
				a.addError("function 'Insert' second argument must be a variable at %s",
					expr.Token.Pos.String())
			} else {
				targetType := a.analyzeExpression(expr.Arguments[1])
				if targetType != nil && targetType != types.STRING {
					a.addError("function 'Insert' second argument must be String, got %s at %s",
						targetType.String(), expr.Token.Pos.String())
				}
			}
			posType := a.analyzeExpression(expr.Arguments[2])
			if posType != nil && posType != types.INTEGER {
				a.addError("function 'Insert' third argument must be Integer, got %s at %s",
					posType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// Higher-order functions: Map, Filter, Reduce, ForEach, Every, Some, Find, FindIndex, Slice
		if ident.Equal(funcIdent.Value, "Map") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'Map' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType
			}
			return types.VOID
		}

		if ident.Equal(funcIdent.Value, "Filter") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'Filter' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType
			}
			return types.VOID
		}

		if ident.Equal(funcIdent.Value, "Reduce") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'Reduce' expects 3 arguments (array, lambda, initial), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			return a.analyzeExpression(expr.Arguments[2])
		}

		if ident.Equal(funcIdent.Value, "ForEach") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'ForEach' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			return types.VOID
		}

		if ident.Equal(funcIdent.Value, "Every") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'Every' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			return types.BOOLEAN
		}

		if ident.Equal(funcIdent.Value, "Some") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'Some' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			return types.BOOLEAN
		}

		if ident.Equal(funcIdent.Value, "Find") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'Find' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType.ElementType
			}
			return types.VARIANT
		}

		if ident.Equal(funcIdent.Value, "FindIndex") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'FindIndex' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			return types.INTEGER
		}

		if ident.Equal(funcIdent.Value, "Slice") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'Slice' expects 3 arguments (array, start, end), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			a.analyzeExpression(expr.Arguments[2])
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType
			}
			return types.VOID
		}

		// Implicit class method call (fallback check)
		if a.currentClass != nil {
			if methodType, found := a.currentClass.GetMethod(funcIdent.Value); found {
				if len(expr.Arguments) != len(methodType.Parameters) {
					a.addError("method '%s' expects %d arguments, got %d at %s",
						funcIdent.Value, len(methodType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
					return methodType.ReturnType
				}
				for i, arg := range expr.Arguments {
					argType := a.analyzeExpression(arg)
					expectedType := methodType.Parameters[i]
					if argType != nil && !a.canAssign(argType, expectedType) {
						a.addError("argument %d to method '%s' has type %s, expected %s at %s",
							i+1, funcIdent.Value, argType.String(), expectedType.String(), expr.Token.Pos.String())
					}
				}
				return methodType.ReturnType
			}
		}

		// GetStackTrace() returns String
		if ident.Equal(funcIdent.Value, "GetStackTrace") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetStackTrace' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// GetCallStack() returns array of stack frame records
		if ident.Equal(funcIdent.Value, "GetCallStack") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetCallStack' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			return types.NewDynamicArrayType(types.VARIANT)
		}

		// Type cast: TypeName(expression)
		if castType := a.analyzeTypeCast(funcIdent.Value, expr.Arguments, expr); castType != nil {
			return castType
		}

		a.addError("%s", errors.FormatUnknownName(funcIdent.Value, expr.Token.Pos.Line, expr.Token.Pos.Column))
		return nil
	}

	// Resolve overloaded functions
	var funcType *types.FunctionType
	if sym.IsOverloadSet {
		candidates := a.symbols.GetOverloadSet(funcIdent.Value)
		if len(candidates) == 0 {
			a.addError("no overload candidates found for '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}

		// Lambda type inference not yet supported for overloaded functions
		hasLambdas, lambdaIndices := detectOverloadedCallWithLambdas(expr.Arguments)
		if hasLambdas {
			a.addError("lambda type inference not yet supported for overloaded function '%s' - please provide explicit parameter types for lambda at argument position(s) %v at %s",
				funcIdent.Value, lambdaIndices, expr.Token.Pos.String())
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

		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			a.addError("%s", errors.FormatNoOverloadError(funcIdent.Value, expr.Token.Pos.Line, expr.Token.Pos.Column))
			return nil
		}

		var ok bool
		funcType, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("selected overload for '%s' is not a function type at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}
	} else {
		// Check function pointer first
		if funcPtrType := a.analyzeFunctionPointerCall(expr, sym.Type); funcPtrType != nil {
			return funcPtrType
		}

		var ok bool
		funcType, ok = sym.Type.(*types.FunctionType)
		if !ok {
			// Check record method overloads (handles shadowed symbols like Result alias)
			if a.currentRecord != nil {
				resolveRecordOverloads := func(overloads []*types.MethodInfo) *types.FunctionType {
					if len(overloads) == 0 {
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
						candidates[i] = &Symbol{Type: overload.Signature}
					}
					selected, err := ResolveOverload(candidates, argTypes)
					if err != nil || selected == nil || selected.Type == nil {
						return nil
					}
					if ft, ok := selected.Type.(*types.FunctionType); ok {
						return ft
					}
					return nil
				}

				methodNameLower := ident.Normalize(funcIdent.Value)
				if ft := resolveRecordOverloads(a.currentRecord.GetMethodOverloads(methodNameLower)); ft != nil {
					funcType = ft
					ok = true
				} else if ft := resolveRecordOverloads(a.currentRecord.GetClassMethodOverloads(methodNameLower)); ft != nil {
					funcType = ft
					ok = true
				}
			}

			if !ok {
				// Check type cast before reporting error
				if castType := a.analyzeTypeCast(funcIdent.Value, expr.Arguments, expr); castType != nil {
					return castType
				}
				a.addError("'%s' is not a function at %s", funcIdent.Value, expr.Token.Pos.String())
				return nil
			}
		}
	}

	// Check argument count (handles optional parameters)
	requiredParams := 0
	for _, defaultVal := range funcType.DefaultValues {
		if defaultVal == nil {
			requiredParams++
		}
	}

	if len(expr.Arguments) < requiredParams {
		if requiredParams == len(funcType.Parameters) {
			a.addError("function '%s' expects %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		} else {
			a.addError("function '%s' expects at least %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		}
		return nil
	}
	if len(expr.Arguments) > len(funcType.Parameters) {
		a.addError("function '%s' expects at most %d arguments, got %d at %s",
			funcIdent.Value, len(funcType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return nil
	}

	// Check argument types (handles lazy and var parameters)
	for i, arg := range expr.Arguments {
		expectedType := funcType.Parameters[i]
		isLazy := len(funcType.LazyParams) > i && funcType.LazyParams[i]
		isVar := len(funcType.VarParams) > i && funcType.VarParams[i]

		// Var parameters must be lvalues
		if isVar && !a.isLValue(arg) {
			a.addError("var parameter %d to function '%s' requires a variable (identifier, array element, or field), got %s at %s",
				i+1, funcIdent.Value, arg.String(), arg.Pos().String())
		}

		if isLazy {
			// Lazy: check type without evaluating
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			// Handle parameterless functions as implicit calls
			if implicitType := a.getImplicitCallType(arg); implicitType != nil {
				argType = implicitType
			}
			if argType != nil && !a.canAssign(argType, expectedType) {
				pos := expr.Token.Pos
				a.addError("%s", errors.FormatArgumentError(i, expectedType.String(), argType.String(), pos.Line, pos.Column))
			}
		} else {
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			// Allow compatible array types for var parameters
			if isVar && argType != nil && !a.canAssign(argType, expectedType) {
				if a.areArrayTypesCompatibleForVarParam(argType, expectedType) {
					continue
				}
			}
			if argType != nil && !a.canAssign(argType, expectedType) {
				pos := expr.Token.Pos
				a.addError("%s", errors.FormatArgumentError(i, expectedType.String(), argType.String(), pos.Line, pos.Column))
			}
		}
	}

	return funcType.ReturnType
}

// getImplicitCallType returns the return type when a parameterless function
// identifier is used as a lazy argument (implicit call).
func (a *Analyzer) getImplicitCallType(arg ast.Expression) types.Type {
	ident, ok := arg.(*ast.Identifier)
	if !ok {
		return nil
	}

	sym, symOk := a.symbols.Resolve(ident.Value)
	if !symOk {
		return nil
	}

	funcType, isFuncType := sym.Type.(*types.FunctionType)
	if !isFuncType || len(funcType.Parameters) > 0 {
		return nil
	}

	if funcType.IsProcedure() {
		return types.VOID
	}
	return funcType.ReturnType
}

// analyzeConstructorCall analyzes constructor calls like TClass.Create(args).
// Handles overload resolution, visibility, and argument validation.
func (a *Analyzer) analyzeConstructorCall(expr *ast.CallExpression, classType *types.ClassType, constructorName string) types.Type {
	constructorOverloads := a.getMethodOverloadsInHierarchy(constructorName, classType)
	if len(constructorOverloads) == 0 {
		a.addError("class '%s' has no constructor named '%s' at %s",
			classType.Name, constructorName, expr.Token.Pos.String())
		return classType
	}

	// Resolve overload
	var selectedConstructor *types.MethodInfo
	var selectedSignature *types.FunctionType

	if len(constructorOverloads) == 1 {
		selectedConstructor = constructorOverloads[0]
		selectedSignature = selectedConstructor.Signature
	} else {
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return classType
			}
			argTypes[i] = argType
		}

		candidates := make([]*Symbol, len(constructorOverloads))
		for i, overload := range constructorOverloads {
			candidates[i] = &Symbol{Type: overload.Signature}
		}

		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			a.addError("there is no overloaded constructor '%s' that can be called with these arguments at %s",
				constructorName, expr.Token.Pos.String())
			return classType
		}

		selectedSignature = selected.Type.(*types.FunctionType)
		for _, overload := range constructorOverloads {
			if overload.Signature == selectedSignature {
				selectedConstructor = overload
				break
			}
		}
	}

	// Check visibility
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

	// Validate argument count
	if len(expr.Arguments) != len(selectedSignature.Parameters) {
		a.addError("constructor '%s' expects %d arguments, got %d at %s",
			constructorName, len(selectedSignature.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return classType
	}

	// Validate argument types
	for i, arg := range expr.Arguments {
		if i >= len(selectedSignature.Parameters) {
			break
		}
		paramType := selectedSignature.Parameters[i]
		argType := a.analyzeExpressionWithExpectedType(arg, paramType)
		if argType != nil && !a.canAssign(argType, paramType) {
			a.addError("argument %d to constructor '%s' has type %s, expected %s at %s",
				i+1, constructorName, argType.String(), paramType.String(),
				expr.Token.Pos.String())
		}
	}

	return classType
}

// analyzeOldExpression analyzes an 'old' expression in postconditions.
func (a *Analyzer) analyzeOldExpression(expr *ast.OldExpression) types.Type {
	if expr.Identifier == nil {
		return nil
	}
	sym, ok := a.symbols.Resolve(expr.Identifier.Value)
	if !ok {
		return nil
	}
	return sym.Type
}

// isLambdaNeedingInference checks if a lambda has untyped parameters or return type.
func isLambdaNeedingInference(expr ast.Expression) bool {
	lambda, ok := expr.(*ast.LambdaExpression)
	if !ok {
		return false
	}
	for _, param := range lambda.Parameters {
		if param.Type == nil {
			return true
		}
	}
	return lambda.ReturnType == nil
}

// detectOverloadedCallWithLambdas returns indices of lambdas needing type inference.
func detectOverloadedCallWithLambdas(args []ast.Expression) (bool, []int) {
	lambdaIndices := []int{}
	for i, arg := range args {
		if isLambdaNeedingInference(arg) {
			lambdaIndices = append(lambdaIndices, i)
		}
	}
	return len(lambdaIndices) > 0, lambdaIndices
}

// analyzeTypeCast analyzes type cast expressions like Integer(x) or TMyClass(obj).
// Returns target type if valid, nil otherwise.
func (a *Analyzer) analyzeTypeCast(typeName string, args []ast.Expression, expr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		return nil
	}

	targetType, err := a.resolveType(typeName)
	if err != nil {
		return nil
	}

	if typeAlias, ok := targetType.(*types.TypeAlias); ok {
		targetType = typeAlias.AliasedType
	}

	argType := a.analyzeExpression(args[0])
	if argType == nil {
		return nil
	}

	if !a.isValidCast(argType, targetType, expr.Token.Pos) {
		return nil
	}
	return targetType
}

// isValidCast checks if casting from sourceType to targetType is valid.
func (a *Analyzer) isValidCast(sourceType, targetType types.Type, pos token.Position) bool {
	sourceType = types.GetUnderlyingType(sourceType)
	targetType = types.GetUnderlyingType(targetType)

	// Same type always valid
	if sourceType.Equals(targetType) {
		return true
	}

	// Numeric conversions (Integer <-> Float)
	if a.isNumericType(sourceType) && a.isNumericType(targetType) {
		return true
	}

	// Boolean conversions
	if targetType == types.BOOLEAN {
		if sourceType == types.INTEGER || sourceType == types.FLOAT || sourceType == types.STRING {
			return true
		}
	}

	// String conversions (most types can convert to string)
	if targetType == types.STRING {
		return true
	}

	// Variant conversions
	if sourceType == types.VARIANT || targetType == types.VARIANT {
		return true
	}

	// Class casts (must be related by inheritance)
	sourceClass, sourceIsClass := sourceType.(*types.ClassType)
	targetClass, targetIsClass := targetType.(*types.ClassType)
	if sourceIsClass && targetIsClass {
		if types.IsSubclassOf(sourceClass, targetClass) ||
			types.IsSubclassOf(targetClass, sourceClass) {
			return true
		}
		a.addError("cannot cast %s to %s: types are not related by inheritance at %s",
			sourceType.String(), targetType.String(), pos.String())
		return false
	}

	// Interface casts (checked at runtime)
	_, sourceIsInterface := sourceType.(*types.InterfaceType)
	_, targetIsInterface := targetType.(*types.InterfaceType)
	if sourceIsInterface || targetIsInterface {
		return true
	}

	// Enum <-> Integer casts
	_, sourceIsEnum := sourceType.(*types.EnumType)
	_, targetIsEnum := targetType.(*types.EnumType)
	if (sourceIsEnum && targetType == types.INTEGER) ||
		(sourceType == types.INTEGER && targetIsEnum) {
		return true
	}

	a.addError("cannot cast %s to %s at %s",
		sourceType.String(), targetType.String(), pos.String())
	return false
}

// isNumericType checks if a type is Integer or Float.
func (a *Analyzer) isNumericType(t types.Type) bool {
	t = types.GetUnderlyingType(t)
	return t == types.INTEGER || t == types.FLOAT
}

// analyzeRecordStaticMethodCall analyzes static method calls like TRecord.Method(args).
func (a *Analyzer) analyzeRecordStaticMethodCall(expr *ast.CallExpression, recordType *types.RecordType, methodName string) types.Type {
	lowerMethodName := ident.Normalize(methodName)
	overloads := recordType.GetClassMethodOverloads(lowerMethodName)
	if len(overloads) == 0 {
		a.addError("record type '%s' has no class method '%s' at %s",
			recordType.Name, methodName, expr.Token.Pos.String())
		return nil
	}

	// Resolve overload
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
		candidates[i] = &Symbol{Type: overload.Signature}
	}

	selected, err := ResolveOverload(candidates, argTypes)
	if err != nil {
		a.addError("%s", errors.FormatNoOverloadError(methodName, expr.Token.Pos.Line, expr.Token.Pos.Column))
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
