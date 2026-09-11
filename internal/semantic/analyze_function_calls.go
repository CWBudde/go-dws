package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

func (a *Analyzer) argumentMatchesParameter(argType, paramType types.Type, strict bool) bool {
	if strict {
		return types.IsIdentical(argType, paramType)
	}
	return a.canAssign(argType, paramType)
}

func (a *Analyzer) analyzeArgumentForParameter(arg ast.Expression, paramType types.Type, strict bool) types.Type {
	if strict {
		return a.analyzeExpression(arg)
	}
	return a.analyzeExpressionWithExpectedType(arg, paramType)
}

func (a *Analyzer) analyzeCallExpression(expr *ast.CallExpression) types.Type {
	// Handle member access expressions (method calls like obj.Method())
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		if name, ok := memberAccess.Object.(*ast.Identifier); ok {
			if symbols, imported := a.unitSymbols[ident.Normalize(name.Value)]; imported {
				// Reuse regular call checking, including overloads and defaults, in the unit namespace.
				oldSymbols := a.symbols
				a.symbols = NewEnclosedSymbolTable(oldSymbols)
				if symbol, found := symbols.symbols.Get(memberAccess.Member.Value); found {
					a.symbols.symbols.Set(memberAccess.Member.Value, symbol)
				} else {
					a.symbols = oldSymbols
					a.addStructuredError(NewUnknownNameError(memberAccess.Member.Token.Pos, name.Value+"."+memberAccess.Member.Value))
					return nil
				}
				defer func() { a.symbols = oldSymbols }()
				call := *expr
				call.Function = memberAccess.Member
				return a.analyzeCallExpression(&call)
			}
		}
		// JSON namespace calls (JSON.Parse/Stringify/...) must be recognized before
		// the `JSON` identifier is analyzed as an ordinary (undefined) symbol.
		if a.isJSONNamespace(memberAccess.Object) {
			return a.analyzeJSONNamespaceResult(memberAccess.Member.Value, expr.Arguments)
		}
		if a.isDefaultNamespace(memberAccess.Object) {
			builtinCall := *expr
			builtinCall.Function = memberAccess.Member
			if resultType, isBuiltin := a.analyzeBuiltinFunction(memberAccess.Member.Value, expr.Arguments, &builtinCall); isBuiltin {
				return resultType
			}
			a.addStructuredError(NewUnknownNameError(memberAccess.Member.Token.Pos, "Default."+memberAccess.Member.Value))
			return nil
		}

		objectType := a.analyzeExpression(memberAccess.Object)
		if objectType == nil {
			return nil
		}

		// Method call on a JSONVariant receiver (v.TypeName(), v.Add(x), ...).
		if types.IsJSONVariant(objectType) {
			return a.analyzeJSONMethodResult(memberAccess.Member.Value, expr.Arguments)
		}
		// Method call on a ByteBuffer receiver spelled as a call expression.
		if types.IsByteBuffer(objectType) {
			return a.analyzeByteBufferMethodResult(memberAccess.Member.Value, expr.Arguments)
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
			argType := a.analyzeArgumentForParameter(arg, paramType, i < len(funcType.StrictParams) && funcType.StrictParams[i])
			if argType != nil && !a.argumentMatchesParameter(argType, paramType, i < len(funcType.StrictParams) && funcType.StrictParams[i]) {
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
		a.addError("Not a method at %s", expr.Token.Pos.String())
		return nil
	}

	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		a.addError("Not a method at %s", expr.Token.Pos.String())
		return nil
	}

	sym, ok := a.symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check built-in functions. The callee's case-mismatch hint is emitted
		// before the arguments are analyzed so hints appear in source order.
		if a.isBuiltinFunction(funcIdent.Value) {
			if declName := a.builtinDeclarationName(funcIdent.Value); declName != "" && declName != funcIdent.Value {
				a.addCaseMismatchHint(funcIdent.Value, declName, funcIdent.Token.Pos)
			}
		}
		if resultType, isBuiltin := a.analyzeBuiltinFunction(funcIdent.Value, expr.Arguments, expr); isBuiltin {
			return resultType
		}

		if resultType, handled := a.analyzeImplicitHelperCall(funcIdent.Value, expr.Arguments, expr.Token.Pos); handled {
			return resultType
		}

		// Check implicit Self method call within a class
		if a.currentClass != nil {
			methodNameLower := ident.Normalize(funcIdent.Value)
			overloads := a.getMethodOverloadsInHierarchy(methodNameLower, a.currentClass)
			if len(overloads) > 0 {
				var selectedMethod *types.MethodInfo
				if len(overloads) == 1 {
					selectedMethod = overloads[0]
				} else {
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
						candidates[i] = &Symbol{Type: overload.Signature}
					}
					selected, err := ResolveOverload(candidates, argTypes)
					if err != nil {
						a.addStructuredError(NewNoOverloadMatchError(funcIdent.Token.Pos, funcIdent.Value))
						return nil
					}
					for i, candidate := range candidates {
						if candidate == selected {
							selectedMethod = overloads[i]
							break
						}
					}
				}
				if selectedMethod == nil {
					return nil
				}
				if a.inClassMethod && !selectedMethod.IsClassMethod {
					a.addStructuredError(NewClassMethodOrConstructorExpectedError(funcIdent.Token.Pos))
					return nil
				}
				methodType := selectedMethod.Signature

				// Check visibility (the selected overload's own visibility governs)
				methodOwner := a.getMethodOwner(a.currentClass, methodNameLower)
				if methodOwner != nil {
					// selectedMethod is guaranteed non-nil above, so its own
					// visibility governs the check.
					visibility, hasVisibility := selectedMethod.Visibility, true
					if hasVisibility && !a.checkVisibility(methodOwner, visibility, funcIdent.Value, "method") {
						a.addStructuredError(NewVisibilityScopeError(funcIdent.Token.Pos, funcIdent.Value))
						return nil
					}
					a.recordClassMethodUsage(methodOwner, funcIdent.Value)
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
					argType := a.analyzeArgumentForParameter(arg, paramType, i < len(methodType.StrictParams) && methodType.StrictParams[i])
					if argType != nil && !a.argumentMatchesParameter(argType, paramType, i < len(methodType.StrictParams) && methodType.StrictParams[i]) {
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
					argType := a.analyzeOverloadArgument(arg)
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
					a.addStructuredError(NewNoOverloadMatchError(funcIdent.Token.Pos, funcIdent.Value))
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
					argType := a.analyzeOverloadArgument(arg)
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
					a.addStructuredError(NewNoOverloadMatchError(funcIdent.Token.Pos, funcIdent.Value))
					return nil
				}

				methodType, ok := selected.Type.(*types.FunctionType)
				if !ok {
					a.addError("internal error: expected function type for selected record class method, but got %T", selected.Type)
					return nil
				}
				for i, arg := range expr.Arguments {
					if i >= len(methodType.Parameters) {
						break
					}
					paramType := methodType.Parameters[i]
					argType := a.analyzeArgumentForParameter(arg, paramType, i < len(methodType.StrictParams) && methodType.StrictParams[i])
					if argType != nil && !a.argumentMatchesParameter(argType, paramType, i < len(methodType.StrictParams) && methodType.StrictParams[i]) {
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
				if methodOwner := a.getMethodOwner(a.currentClass, funcIdent.Value); methodOwner != nil {
					visibility, hasVisibility := methodOwner.MethodVisibility[ident.Normalize(funcIdent.Value)]
					if hasVisibility && !a.checkVisibility(methodOwner, visibility, funcIdent.Value, "method") {
						a.addStructuredError(NewVisibilityScopeError(funcIdent.Token.Pos, funcIdent.Value))
						return nil
					}
					a.recordClassMethodUsage(methodOwner, funcIdent.Value)
				}
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
		if castType, handled := a.analyzeTypeCast(funcIdent.Value, expr.Arguments, expr); handled {
			return castType
		}

		a.addStructuredError(NewUnknownNameError(expr.Token.Pos, funcIdent.Value))
		return nil
	}

	// The call references the resolved symbol (marks nested/local functions as used).
	a.recordSymbolUsage(sym.Name, funcIdent.Token.Pos)

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
			argType := a.analyzeOverloadArgument(arg)
			if argType == nil {
				return nil
			}
			argTypes[i] = argType
		}

		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, funcIdent.Value))
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
						argType := a.analyzeOverloadArgument(arg)
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
				if castType, handled := a.analyzeTypeCast(funcIdent.Value, expr.Arguments, expr); handled {
					return castType
				}
				a.addError("'%s' is not a function at %s", funcIdent.Value, expr.Token.Pos.String())
				return nil
			}
		}
	}

	overloadSet := a.symbols.GetOverloadSet(funcIdent.Value)
	hasOverloads := sym.IsOverloadSet || len(overloadSet) > 1

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
				pos := arg.Pos()
				a.addError("%s", errors.FormatArgumentError(i, semanticFunctionParamTypeName(funcType, i, expectedType), argType.String(), pos.Line, pos.Column))
			}
		} else {
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if isArrayOfConstType(expectedType) {
				if _, isLiteral := arg.(*ast.ArrayLiteralExpression); !isLiteral {
					// Any array value can be passed to an open "array of const"
					// parameter; only reject non-array arguments.
					if argType != nil {
						if _, isArr := types.GetUnderlyingType(argType).(*types.ArrayType); !isArr {
							pos := expr.Token.Pos
							a.addError("%s", errors.FormatArgumentError(i, semanticFunctionParamTypeName(funcType, i, expectedType), argType.String(), pos.Line, pos.Column))
						}
					}
					continue
				}
			}
			// Allow compatible array types for var parameters
			if isVar && argType != nil && !a.canAssign(argType, expectedType) {
				if a.areArrayTypesCompatibleForVarParam(argType, expectedType) {
					continue
				}
			}
			if argType != nil {
				if hasOverloads {
					if !a.canAssign(argType, expectedType) {
						pos := arg.Pos()
						a.addError("%s", errors.FormatArgumentError(i, semanticFunctionParamTypeName(funcType, i, expectedType), argType.String(), pos.Line, pos.Column))
					}
					continue
				}
				if i < len(funcType.StrictParams) && funcType.StrictParams[i] {
					if !types.IsIdentical(argType, expectedType) {
						pos := arg.Pos()
						a.addError("%s", errors.FormatArgumentError(i, semanticFunctionParamTypeName(funcType, i, expectedType), argType.String(), pos.Line, pos.Column))
					}
					continue
				}
				if !a.canAssign(argType, expectedType) {
					pos := arg.Pos()
					a.addError("%s", errors.FormatArgumentError(i, semanticFunctionParamTypeName(funcType, i, expectedType), argType.String(), pos.Line, pos.Column))
				}
			}
		}
	}

	return funcType.ReturnType
}

func (a *Analyzer) currentImplicitSelfType() types.Type {
	if a.currentSelfType != nil {
		return a.currentSelfType
	}
	if a.currentClass != nil {
		return a.currentClass
	}
	if a.currentRecord != nil {
		return a.currentRecord
	}
	return nil
}

func (a *Analyzer) analyzeImplicitHelperCall(methodName string, args []ast.Expression, pos token.Position) (types.Type, bool) {
	selfType := a.currentImplicitSelfType()
	if selfType == nil {
		return nil, false
	}

	methodType := a.resolveHelperMethodForCall(selfType, methodName, args)
	if methodType == nil {
		return nil, false
	}

	if len(args) != len(methodType.Parameters) {
		a.addError("method '%s' expects %d argument(s), got %d at %s",
			methodName, len(methodType.Parameters), len(args), pos.String())
		return methodType.ReturnType, true
	}

	for i, arg := range args {
		paramType := methodType.Parameters[i]
		argType := a.analyzeArgumentForParameter(arg, paramType, i < len(methodType.StrictParams) && methodType.StrictParams[i])
		if argType != nil && !a.argumentMatchesParameter(argType, paramType, i < len(methodType.StrictParams) && methodType.StrictParams[i]) {
			a.addError("argument %d to method '%s' has type %s, expected %s at %s",
				i+1, methodName, argType.String(), paramType.String(), pos.String())
		}
	}

	return methodType.ReturnType, true
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

// applyImplicitCallType unwraps a parameterless function reference into the
// type of its result, so that `Test.Member` and `Test[Index]` operate on the
// value the call yields rather than on the function itself. expr is the
// expression that produced typ; typ is returned unchanged when no implicit
// call applies.
func (a *Analyzer) applyImplicitCallType(expr ast.Expression, typ types.Type) types.Type {
	if implicitType := a.getImplicitCallType(expr); implicitType != nil {
		return implicitType
	}
	if implicitType := implicitCallReturnTypeFromType(typ); implicitType != nil {
		return implicitType
	}
	// Overload-set symbols deliberately carry no type of their own, so the
	// caller sees a nil type. The implicit call still has a result type when
	// the set holds exactly one parameterless overload.
	if typ == nil {
		if funcType := a.parameterlessOverloadType(expr); funcType != nil {
			if funcType.IsProcedure() {
				return types.VOID
			}
			return funcType.ReturnType
		}
	}
	return typ
}

// parameterlessOverloadType returns the signature of the single parameterless
// overload of expr's overload set, or nil when expr is not an overload set or
// the set has no unique parameterless overload.
func (a *Analyzer) parameterlessOverloadType(expr ast.Expression) *types.FunctionType {
	identExpr, ok := expr.(*ast.Identifier)
	if !ok {
		return nil
	}
	sym, symOk := a.symbols.Resolve(identExpr.Value)
	if !symOk || !sym.IsOverloadSet {
		return nil
	}

	var found *types.FunctionType
	for _, overload := range a.symbols.GetOverloadSet(identExpr.Value) {
		funcType, isFuncType := overload.Type.(*types.FunctionType)
		if !isFuncType || len(funcType.Parameters) > 0 {
			continue
		}
		if found != nil {
			// Ambiguous: leave the call to regular overload resolution.
			return nil
		}
		found = funcType
	}
	return found
}

func implicitCallReturnTypeFromType(typ types.Type) types.Type {
	funcType, ok := types.GetUnderlyingType(typ).(*types.FunctionType)
	if !ok || len(funcType.Parameters) > 0 {
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
	if classType.IsStatic {
		a.addStructuredError(NewStaticClassInstantiationError(constructorCallPosition(expr), classType.Name))
		return classType
	}

	constructorOverloads := a.getMethodOverloadsInHierarchy(constructorName, classType)
	if len(constructorOverloads) == 0 {
		if len(expr.Arguments) > 0 {
			a.addError("class '%s' has no constructor named '%s' at %s",
				classType.Name, constructorName, expr.Token.Pos.String())
			return classType
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

	// Resolve overload
	var selectedConstructor *types.MethodInfo
	var selectedSignature *types.FunctionType

	if len(constructorOverloads) == 1 {
		selectedConstructor = constructorOverloads[0]
		selectedSignature = selectedConstructor.Signature
	} else {
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeOverloadArgument(arg)
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

		var ok bool
		selectedSignature, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("internal error: expected function type for selected constructor, but got %T", selected.Type)
			return classType
		}
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
		a.recordClassMethodUsage(ownerClass, constructorName)
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
		argType := a.analyzeArgumentForParameter(arg, paramType, i < len(selectedSignature.StrictParams) && selectedSignature.StrictParams[i])
		if argType != nil && !a.argumentMatchesParameter(argType, paramType, i < len(selectedSignature.StrictParams) && selectedSignature.StrictParams[i]) {
			a.addError("argument %d to constructor '%s' has type %s, expected %s at %s",
				i+1, constructorName, argType.String(), paramType.String(),
				expr.Token.Pos.String())
		}
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
// Returns the target type and whether this was handled as a cast attempt.
func (a *Analyzer) analyzeTypeCast(typeName string, args []ast.Expression, expr *ast.CallExpression) (types.Type, bool) {
	if len(args) != 1 {
		return nil, false
	}

	targetType, err := a.resolveType(typeName)
	if err != nil {
		return nil, false
	}

	if typeAlias, ok := targetType.(*types.TypeAlias); ok {
		targetType = typeAlias.AliasedType
	}

	argType := a.analyzeExpression(args[0])
	if argType == nil {
		return nil, true
	}

	if !a.isValidCast(argType, targetType, args[0].Pos()) {
		return nil, true
	}
	return targetType, true
}

// maxSetIntegerCastOrdinal is the highest element ordinal a set may hold and
// still have an integer representation.
//
// The integer form of a set is a bitmask in which an element's *absolute*
// ordinal is the bit index — `castToSet` and the `Integer(set)` cast both index
// by ordinal, not by an offset from the base type's low bound. So the rule is
// about the highest reachable ordinal, not about how many elements the base
// type spans: `set of (h33 = 33, h64 = 64)` spans only 32 values but needs bit
// 64, which no mask holds.
//
// The bound itself is pinned by the corpus rather than by upstream source
// (`reference/dwscript-original/` is empty in this checkout):
// `SetOfPass/set_to_integer` and `set_to_integer2` cast base types with ordinals
// 0..1 and must succeed, while `SetOfFail/integer_vs_set` casts
// `(one = 1, fifty = 50)` and must fail. That places the limit somewhere in
// 1..49, so the mask is a 32-bit one and the highest usable bit is 31.
const maxSetIntegerCastOrdinal = 31

// checkSetIntegerCastWidth validates a set <-> Integer cast against the ordinal
// range of the set's base type, reporting DWScript's diagnostic when the set has
// no integer representation.
func (a *Analyzer) checkSetIntegerCastWidth(setType *types.SetType, pos token.Position) bool {
	if setType == nil || setType.ElementType == nil {
		return true
	}

	// An unbounded ordinal base (`set of Integer`) has no bitmask form at all:
	// the set itself works, backed by map storage, but a cast would silently
	// drop every element outside the mask.
	low, high, ok := types.OrdinalBounds(setType.ElementType)
	if !ok || low < 0 || high > maxSetIntegerCastOrdinal {
		a.addError("Set has too many elements for cast to integer at %s", pos.String())
		return false
	}

	return true
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

	// JSONVariant conversions are valid only to/from base types and Variant
	// (validated at runtime). Casts to unrelated types (e.g. class/interface)
	// are not permitted and fall through to the checks below.
	if types.IsJSONVariant(sourceType) || types.IsJSONVariant(targetType) {
		other := sourceType
		if types.IsJSONVariant(sourceType) {
			other = targetType
		}
		switch other.TypeKind() {
		case "INTEGER", "FLOAT", "STRING", "BOOLEAN", "VARIANT", "JSON_VARIANT":
			return true
		}
	}

	// ByteBuffer(s) builds a buffer from a data string; ByteBuffer to ByteBuffer
	// is an identity cast. Nothing else converts.
	if types.IsByteBuffer(targetType) {
		return sourceType.TypeKind() == "STRING"
	}
	if types.IsByteBuffer(sourceType) {
		return false
	}

	// nil can be cast to any reference type: class, interface, metaclass
	if sourceType.TypeKind() == "NIL" {
		switch targetType.TypeKind() {
		case "CLASS", "INTERFACE", "CLASSOF":
			return true
		}
	}

	// Class casts (must be related by inheritance)
	sourceClass, sourceIsClass := sourceType.(*types.ClassType)
	targetClass, targetIsClass := targetType.(*types.ClassType)
	if sourceIsClass && targetIsClass {
		if types.IsSubclassOf(sourceClass, targetClass) ||
			types.IsSubclassOf(targetClass, sourceClass) {
			return true
		}
		a.addStructuredError(NewIncompatibleTypesPairError(pos, sourceType.String(), targetType.String()))
		return false
	}

	// Interface casts (checked at runtime)
	_, sourceIsInterface := sourceType.(*types.InterfaceType)
	_, targetIsInterface := targetType.(*types.InterfaceType)
	if sourceIsInterface || targetIsInterface {
		return true
	}

	// Enum <-> Integer casts. A Float source is accepted on the way in and
	// truncated to its ordinal, so `TEnum(IntPower(10, i))` compiles.
	_, sourceIsEnum := sourceType.(*types.EnumType)
	_, targetIsEnum := targetType.(*types.EnumType)
	if (sourceIsEnum && targetType == types.INTEGER) ||
		((sourceType == types.INTEGER || sourceType == types.FLOAT) && targetIsEnum) {
		return true
	}

	// Set <-> Integer casts (a set's integer form is its ordinal bitmask). The
	// bitmask only exists for a set narrow enough to fit one, so a wider base
	// type is rejected the way DWScript rejects it.
	sourceSet, sourceIsSet := sourceType.(*types.SetType)
	targetSet, targetIsSet := targetType.(*types.SetType)
	if sourceIsSet && targetType == types.INTEGER {
		return a.checkSetIntegerCastWidth(sourceSet, pos)
	}
	if sourceType == types.INTEGER && targetIsSet {
		return a.checkSetIntegerCastWidth(targetSet, pos)
	}

	if _, isEnum := targetType.(*types.EnumType); isEnum {
		a.addError("Syntax Error: Cannot cast this type to \"Integer\" [line: %d, column: %d]",
			pos.Line, pos.Column)
		return false
	}

	a.addStructuredError(NewIncompatibleTypesPairError(pos, sourceType.String(), targetType.String()))
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
		argType := a.analyzeOverloadArgument(arg)
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
		a.addStructuredError(NewNoOverloadMatchError(expr.Token.Pos, methodName))
		return nil
	}

	funcType, ok := selected.Type.(*types.FunctionType)
	if !ok {
		a.addError("internal error: expected function type for selected record static method, but got %T", selected.Type)
		return nil
	}

	// Validate argument types
	for i, arg := range expr.Arguments {
		if i >= len(funcType.Parameters) {
			break
		}
		paramType := funcType.Parameters[i]
		argType := a.analyzeArgumentForParameter(arg, paramType, i < len(funcType.StrictParams) && funcType.StrictParams[i])
		if argType != nil && !a.argumentMatchesParameter(argType, paramType, i < len(funcType.StrictParams) && funcType.StrictParams[i]) {
			a.addError("argument %d to '%s.%s' has type %s, expected %s at %s",
				i+1, recordType.Name, methodName, argType.String(), paramType.String(),
				expr.Token.Pos.String())
		}
	}

	return funcType.ReturnType
}
