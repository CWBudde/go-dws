package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for function call and object creation AST nodes.

// VisitCallExpression evaluates a function call expression.
// Handles multiple call types: function pointers, member access, user functions,
// built-in functions, type casts, implicit self methods, and record static methods.
// Supports lazy parameters (Jensen's Device) and var parameters (pass-by-reference).
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	if node.Function == nil {
		return e.newError(node, "call expression missing function")
	}

	// Function pointer calls: lambdas, function pointers, method pointers
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if valRaw, exists := ctx.Env().Get(funcIdent.Value); exists {
			val := valRaw.(Value)
			if val.Type() == "FUNCTION_POINTER" || val.Type() == "LAMBDA" || val.Type() == "METHOD_POINTER" {
				funcPtr, ok := val.(FunctionPointerCallable)
				if !ok {
					// Fallback for non-standard function pointer types
					fallbackArgs := make([]Value, len(node.Arguments))
					for i, arg := range node.Arguments {
						fallbackArgs[i] = e.Eval(arg, ctx)
						if isError(fallbackArgs[i]) {
							return fallbackArgs[i]
						}
					}
					return e.oopEngine.CallFunctionPointer(val, fallbackArgs, node)
				}

				// Get function declaration for parameter metadata
				funcDecl, _ := funcPtr.GetFunctionDecl().(*ast.FunctionDecl)

				// Prepare arguments with proper wrapping
				args := make([]Value, len(node.Arguments))
				for idx, arg := range node.Arguments {
					// Extract parameter flags
					isLazy := false
					isByRef := false
					if funcDecl != nil && idx < len(funcDecl.Parameters) {
						isLazy = funcDecl.Parameters[idx].IsLazy
						isByRef = funcDecl.Parameters[idx].ByRef
					}

					if isLazy {
						// Wrap lazy parameters in thunk
						args[idx] = e.wrapLazyArg(arg, ctx, func(expr ast.Expression) Value {
							return e.Eval(expr, ctx)
						})
					} else if isByRef {
						// Wrap var parameters in reference with get/set callbacks
						if argIdent, ok := arg.(*ast.Identifier); ok {
							varName := argIdent.Value
							capturedEnv := ctx.Env()

							var getter runtime.GetterCallback = func() (runtime.Value, error) {
								val, found := capturedEnv.Get(varName)
								if !found {
									return nil, fmt.Errorf("referenced variable %s not found", varName)
								}
								if runtimeVal, ok := val.(runtime.Value); ok {
									return runtimeVal, nil
								}
								return nil, fmt.Errorf("environment value is not a runtime.Value")
							}

							var setter runtime.SetterCallback = func(val runtime.Value) error {
								if err := capturedEnv.Set(varName, val); err != nil {
									return fmt.Errorf("failed to set variable %s: %w", varName, err)
								}
								return nil
							}

							args[idx] = runtime.NewReferenceValue(varName, getter, setter)
						} else {
							return e.newError(arg, "var parameter requires a variable, got %T", arg)
						}
					} else {
						// Regular parameters: evaluate immediately
						argVal := e.Eval(arg, ctx)
						if isError(argVal) {
							return argVal
						}
						args[idx] = argVal
					}
				}

				// Build metadata and execute via adapter
				metadata := FunctionPointerMetadata{
					IsLambda:   funcPtr.IsLambda(),
					Lambda:     funcPtr.GetLambdaExpr(),
					Function:   funcPtr.GetFunctionDecl(),
					Closure:    funcPtr.GetClosure(),
					SelfObject: funcPtr.GetSelfObject(),
				}
				return e.oopEngine.ExecuteFunctionPointerCall(metadata, args, node)
			}
		}
	}

	// Member access calls: obj.Method(), UnitName.Func(), TClass.Create()
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Record, interface, or object method calls
		if objVal.Type() == "RECORD" || objVal.Type() == "INTERFACE" || objVal.Type() == "OBJECT" {
			args := make([]Value, len(node.Arguments))
			for i, arg := range node.Arguments {
				val := e.Eval(arg, ctx)
				if isError(val) {
					return val
				}
				args[i] = val
			}

			// Create synthetic MethodCallExpression for error reporting
			mc := &ast.MethodCallExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: node.Token,
					},
				},
				Object:    memberAccess.Object,
				Method:    memberAccess.Member,
				Arguments: node.Arguments,
			}

			return e.DispatchMethodCall(objVal, memberAccess.Member.Value, args, mc, ctx)
		}

		// Unit-qualified functions or class constructors
		if identNode, ok := memberAccess.Object.(*ast.Identifier); ok {
			// Class constructor or static method
			if e.typeSystem.HasClass(identNode.Value) {
				mc := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: node.Token,
						},
					},
					Object:    identNode,
					Method:    memberAccess.Member,
					Arguments: node.Arguments,
				}
				return e.VisitMethodCallExpression(mc, ctx)
			}

			// Unit-qualified function call
			if e.unitRegistry != nil {
				return e.oopEngine.CallQualifiedOrConstructor(node, memberAccess)
			}
		}

		return e.newError(node, "cannot call member expression that is not a method or unit-qualified function")
	}

	// Remaining call types require a simple identifier
	funcName, ok := node.Function.(*ast.Identifier)
	if !ok {
		return e.newError(node, "function call requires identifier or qualified name, got %T", node.Function)
	}

	// User-defined functions with overload resolution
	funcNameLower := ident.Normalize(funcName.Value)
	if overloads := e.FunctionRegistry().Lookup(funcNameLower); len(overloads) > 0 {
		// Resolve overload and prepare arguments
		var fn *ast.FunctionDecl
		var cachedArgs []Value
		var err error

		if len(overloads) == 1 {
			// Fast path: single overload, skip type checking
			fn = overloads[0]
			cachedArgs, err = e.ResolveOverloadFast(fn, node.Arguments, ctx)
		} else {
			// Multiple overloads: resolve based on argument types
			fn, cachedArgs, err = e.ResolveOverloadMultiple(funcNameLower, overloads, node.Arguments, ctx)
		}

		if err != nil {
			return e.newError(node, "%s", err.Error())
		}

		// Prepare all arguments (handles lazy and var parameters)
		args, err := e.PrepareUserFunctionArgs(fn, node.Arguments, cachedArgs, ctx, node)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}

		// Execute the function via adapter
		return e.oopEngine.CallUserFunction(fn, args)
	}

	// Record static method calls (when inside record method context)
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				if rtmv, ok := recordVal.(RecordTypeMetaValue); ok {
					if rtmv.HasStaticMethod(funcName.Value) {
						return e.oopEngine.DispatchRecordStaticMethod(rtmv.GetRecordTypeName(), node, funcName)
					}
				} else {
					return e.oopEngine.CallRecordStaticMethod(node, funcName)
				}
			}
		}
	}

	// Built-in functions with var parameter handling
	switch funcNameLower {
	case "inc":
		return e.builtinInc(node.Arguments, ctx)
	case "dec":
		return e.builtinDec(node.Arguments, ctx)
	case "setlength":
		return e.builtinSetLength(node.Arguments, ctx)
	case "insert":
		return e.builtinInsert(node.Arguments, ctx)
	case "swap":
		return e.builtinSwap(node.Arguments, ctx)
	case "divmod":
		return e.builtinDivMod(node.Arguments, ctx)
	case "trystrtoint":
		return e.builtinTryStrToInt(node.Arguments, ctx)
	case "trystrtofloat":
		return e.builtinTryStrToFloat(node.Arguments, ctx)
	case "decodedate":
		return e.builtinDecodeDate(node.Arguments, ctx)
	case "decodetime":
		return e.builtinDecodeTime(node.Arguments, ctx)
	case "delete":
		// 3-parameter form modifies string in place
		if len(node.Arguments) == 3 {
			return e.builtinDeleteString(node.Arguments, ctx)
		}
	}

	// External (Go) functions with potential var parameters
	if e.externalFunctions != nil && e.externalFunctions.Has(funcName.Value) {
		return e.callExternalFunction(funcName.Value, node.Arguments, ctx, node)
	}

	// Default(TypeName) - expects unevaluated type identifier
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.builtinDefault(node.Arguments, ctx)
	}

	// Type casts: TypeName(expression) for single-argument calls
	if len(node.Arguments) == 1 {
		result := e.evalTypeCast(funcName.Value, node.Arguments[0], ctx)
		if result != nil {
			return result
		}
	}

	// Standard built-in functions
	args := make([]Value, len(node.Arguments))
	for idx, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Call built-in function from registry
	if fn, ok := builtins.DefaultRegistry.Lookup(funcName.Value); ok {
		return fn(e, args)
	}

	// Implicit Self method calls (checked after built-ins to avoid shadowing)
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok {
			if selfVal.Type() == "OBJECT" || selfVal.Type() == "CLASS" {
				return e.oopEngine.CallImplicitSelfMethod(node, funcName)
			}
			if _, isRecord := selfVal.(*runtime.RecordValue); isRecord {
				return e.oopEngine.CallImplicitSelfMethod(node, funcName)
			}
		}
	}

	// Not found in any registry or context
	return e.newError(node, "function '%s' not found", funcName.Value)
}

// PrepareUserFunctionArgs prepares arguments for user function invocation.
// Handles lazy parameters (thunks), var parameters (references), and regular parameters.
// Uses cached argument values from overload resolution to prevent double-evaluation.
func (e *Evaluator) PrepareUserFunctionArgs(
	fn *ast.FunctionDecl,
	argExprs []ast.Expression,
	cachedArgs []Value,
	ctx *ExecutionContext,
	node ast.Node,
) ([]Value, error) {
	args := make([]Value, len(argExprs))

	for idx, arg := range argExprs {
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

		if isLazy {
			// Lazy parameter: wrap in thunk for delayed evaluation
			args[idx] = e.wrapLazyArg(arg, ctx, func(expr ast.Expression) Value {
				return e.Eval(expr, ctx)
			})

		} else if isByRef {
			// Var parameter: wrap in reference with get/set callbacks
			argIdent, ok := arg.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("var parameter requires a variable, got %T", arg)
			}

			varName := argIdent.Value
			capturedEnv := ctx.Env()

			var getter runtime.GetterCallback = func() (runtime.Value, error) {
				val, found := capturedEnv.Get(varName)
				if !found {
					return nil, fmt.Errorf("variable %s not found", varName)
				}
				if runtimeVal, ok := val.(runtime.Value); ok {
					return runtimeVal, nil
				}
				return nil, fmt.Errorf("environment value is not a runtime.Value")
			}

			var setter runtime.SetterCallback = func(val runtime.Value) error {
				if err := capturedEnv.Set(varName, val); err != nil {
					return fmt.Errorf("failed to set variable %s: %w", varName, err)
				}
				return nil
			}

			args[idx] = runtime.NewReferenceValue(varName, getter, setter)

		} else {
			// Regular parameter: use cached value
			args[idx] = cachedArgs[idx]
		}
	}

	return args, nil
}

// wrapLazyArg creates a thunk for lazy parameters.
// Reuses existing thunks to avoid self-recursive wrapping.
func (e *Evaluator) wrapLazyArg(arg ast.Expression, ctx *ExecutionContext, eval func(ast.Expression) Value) Value {
	capturedArg := arg
	return runtime.NewLazyThunk(capturedArg, func() runtime.Value {
		if identArg, ok := capturedArg.(*ast.Identifier); ok {
			if valRaw, ok := ctx.Env().Get(identArg.Value); ok {
				if lazyVal, ok := valRaw.(LazyEvaluator); ok {
					return lazyVal.Evaluate()
				}
			}
		}
		return eval(capturedArg)
	})
}

// callExternalFunction handles external (Go) function dispatch with var parameter support.
// For now, delegates to the adapter to handle external functions properly.
// This will be fully migrated in a future phase when we eliminate the adapter.
func (e *Evaluator) callExternalFunction(
	funcName string,
	argExprs []ast.Expression,
	ctx *ExecutionContext,
	node ast.Node,
) Value {
	// For now, delegate to adapter which has full access to external function infrastructure
	// TODO: Move external function handling completely into evaluator in Phase 3.2.9.1
	return e.oopEngine.CallExternalFunction(funcName, argExprs, node)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
// Handles class lookup, field initialization, and constructor execution.
// Validates abstract/external classes and supports implicit parameterless constructors.
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	className := node.ClassName.Value

	// Evaluate constructor arguments
	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[i] = val
	}

	// Look up class in type system
	classInfoAny := e.typeSystem.LookupClass(className)
	if classInfoAny == nil {
		return e.newError(node, "class '%s' not found", className)
	}

	classInfo, ok := classInfoAny.(runtime.IClassInfo)
	if !ok {
		return e.newError(node, "class '%s' has invalid type", className)
	}

	// Validate class can be instantiated
	if classInfo.IsAbstract() {
		return e.newError(node, "Trying to create an instance of an abstract class")
	}
	if classInfo.IsExternal() {
		return e.newError(node, "cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create object instance
	obj := runtime.NewObjectInstance(classInfo)

	// Initialize fields
	fieldTypes := classInfo.GetFieldTypesMap()
	fieldDecls := classInfo.GetFieldsMap()

	for fieldName, fieldTypeAny := range fieldTypes {
		var fieldValue Value
		if fieldDecl, hasDecl := fieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			// Use field initializer
			fieldValue = e.Eval(fieldDecl.InitValue, ctx)
			if isError(fieldValue) {
				return e.newError(node, "failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			// Use zero value for type
			if fieldType, ok := fieldTypeAny.(types.Type); ok {
				fieldValue = e.getZeroValueForType(fieldType)
			} else {
				fieldValue = &runtime.NilValue{}
			}
		}
		obj.SetField(fieldName, fieldValue)
	}

	// Execute constructor
	constructor := classInfo.GetConstructor("Create")
	if constructor != nil {
		err := e.oopEngine.ExecuteConstructor(obj, "Create", args)
		if err != nil {
			return e.newError(node, "constructor failed: %v", err)
		}
	} else if len(args) > 0 {
		return e.newError(node, "no constructor found for class '%s' with %d arguments", className, len(args))
	}

	return obj
}

// VisitNewArrayExpression evaluates a new array expression.
// Resolves element type, evaluates dimensions, and creates multi-dimensional array.
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil new array expression")
	}
	if node.ElementTypeName == nil {
		return e.newError(node, "new array expression missing element type")
	}

	// Resolve element type
	elementTypeName := node.ElementTypeName.Value
	elementType, typeErr := e.ResolveTypeWithContext(elementTypeName, ctx)
	if typeErr != nil {
		return e.newError(node, "unknown element type '%s': %s", elementTypeName, typeErr)
	}

	// Evaluate dimensions
	dimensions, evalErr := e.evaluateDimensions(node.Dimensions, ctx, node)
	if evalErr != nil {
		return evalErr
	}

	return e.CreateMultiDimArray(elementType, dimensions)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
// Captures the current environment in the lambda's closure.
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	return &runtime.FunctionPointerValue{
		Lambda:  node,
		Closure: ctx.Env(),
	}
}
