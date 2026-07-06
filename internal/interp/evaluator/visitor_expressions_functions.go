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

	// Nested (scoped) function declarations hide all same-named outer
	// functions and methods for the duration of the enclosing call.
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if set := e.lookupLocalFunctions(funcIdent.Value, ctx); set != nil {
			return e.callLocalFunctionSet(set, node.Arguments, node, ctx)
		}
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
					return e.executeFunctionPointerDirect(val, fallbackArgs, node, ctx)
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
						ref, err := e.prepareByRefArgument(arg, ctx)
						if err != nil {
							return e.newError(arg, "%s", err.Error())
						}
						args[idx] = ref
					} else {
						// Regular parameters: evaluate immediately
						argVal := e.Eval(arg, ctx)
						if isError(argVal) {
							return argVal
						}
						args[idx] = argVal
					}
				}

				return e.executeFunctionPointerDirect(val, args, node, ctx)
			}
		}
	}

	// Member access calls: obj.Method(), UnitName.Func(), TClass.Create()
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		// JSON namespace call: JSON.Parse(s), JSON.Stringify(x), ...
		if e.isJSONNamespaceObject(memberAccess.Object, ctx) {
			return e.evalJSONNamespaceCall(memberAccess.Member.Value, node.Arguments, node, ctx)
		}

		if identNode, ok := memberAccess.Object.(*ast.Identifier); ok {
			if _, exists := ctx.Env().Get(identNode.Value); !exists {
				// Unit-qualified function call
				if e.UnitRegistry() != nil {
					if _, exists := e.UnitRegistry().GetUnit(identNode.Value); exists {
						return e.executeQualifiedFunctionCall(identNode.Value, memberAccess.Member, node.Arguments, node, ctx)
					}
				}

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
			}
		}

		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Method call on a JSON value receiver: v.TypeName(), v.Add(x), ...
		if jv := jsonValueOf(objVal); jv != nil || isJSONBoxed(objVal) {
			args := make([]Value, len(node.Arguments))
			for i, arg := range node.Arguments {
				val := e.Eval(arg, ctx)
				if isError(val) {
					return val
				}
				args[i] = val
			}
			return e.evalJSONMethodCall(objVal, memberAccess.Member.Value, args, node, ctx)
		}

		if objVal.Type() == "RECORD_TYPE" {
			recordType, ok := objVal.(*RecordTypeValue)
			if !ok {
				return e.newError(node, "record type '%s' has invalid runtime metadata", memberAccess.Object.String())
			}
			args := make([]Value, len(node.Arguments))
			for i, arg := range node.Arguments {
				val := e.Eval(arg, ctx)
				if isError(val) {
					return val
				}
				args[i] = val
			}
			return e.callRecordStaticMethod(recordType, memberAccess.Member.Value, args, node, ctx)
		}

		// Record, interface, or object method calls
		_, isRecordInstance := objVal.(RecordInstanceValue)
		if isRecordInstance || objVal.Type() == "INTERFACE" || objVal.Type() == "OBJECT" {
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

			if recordVal, ok := objVal.(RecordInstanceValue); ok {
				methodDecl, found := recordVal.GetRecordMethod(memberAccess.Member.Value)
				// Overload-aware: pick the best-matching overload for the
				// provided arguments instead of the first registered one.
				if rec, ok := objVal.(*runtime.RecordValue); ok {
					if overloads := rec.GetRecordMethodOverloads(memberAccess.Member.Value); len(overloads) > 1 {
						argVals := make([]Value, len(node.Arguments))
						for i, arg := range node.Arguments {
							val := e.Eval(arg, ctx)
							if isError(val) {
								return val
							}
							argVals[i] = val
						}
						if selected, err := e.selectOverload(rec.GetRecordTypeName(), memberAccess.Member.Value, overloads, argVals); err == nil {
							return e.callRecordMethod(recordVal, selected, argVals, mc, ctx)
						}
					}
				}
				if found {
					args, err := e.prepareArgsForParameters(methodDecl.Parameters, node.Arguments, ctx)
					if err != nil {
						return e.newError(node, "%s", err.Error())
					}
					return e.callRecordMethod(recordVal, methodDecl, args, mc, ctx)
				}
			}

			args := make([]Value, len(node.Arguments))
			for i, arg := range node.Arguments {
				val := e.Eval(arg, ctx)
				if isError(val) {
					return val
				}
				args[i] = val
			}

			return e.DispatchMethodCall(objVal, memberAccess.Member.Value, args, mc, ctx)
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
		// A user overload set can extend a same-named builtin: the builtin
		// participates in resolution and wins on a strictly better match
		// (e.g. IntToStr(Integer) beats a user IntToStr(Variant) overload).
		if result, handled := e.maybeCallBuiltinOverload(funcNameLower, overloads, node, ctx); handled {
			return result
		}
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

		return e.ExecuteUserFunctionDirect(fn, args, ctx)
	}

	// Record static method calls (when inside record method context)
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				if rtmv, ok := recordVal.(RecordTypeMetaValue); ok {
					if rtmv.HasStaticMethod(funcName.Value) {
						recordType, ok := recordVal.(*RecordTypeValue)
						if !ok {
							return e.newError(node, "__CurrentRecord__ is not a record type value")
						}
						recordArgs := make([]Value, len(node.Arguments))
						for i, arg := range node.Arguments {
							val := e.Eval(arg, ctx)
							if isError(val) {
								return val
							}
							recordArgs[i] = val
						}
						return e.callRecordStaticMethod(recordType, funcName.Value, recordArgs, node, ctx)
					}
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
	case "include", "exclude":
		return e.builtinIncludeExclude(funcNameLower, node.Arguments, ctx)
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
	if e.ExternalFunctions() != nil && e.ExternalFunctions().Has(funcName.Value) {
		return e.callExternalFunction(funcName.Value, node.Arguments, node, ctx)
	}

	// Default(TypeName) - expects unevaluated type identifier
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.builtinDefault(node.Arguments)
	}

	// Type casts: TypeName(expression) for single-argument calls
	if len(node.Arguments) == 1 {
		result := e.evalTypeCast(funcName.Value, node.Arguments[0], ctx)
		if result != nil {
			return result
		}
		// Check if an exception was raised during type cast (e.g., invalid downcast)
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
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

	// If evaluating an argument raised an exception (e.g. a failed type cast
	// inside the argument list), the call must not run.
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Call built-in function from registry
	if fn, ok := builtins.DefaultRegistry.Lookup(funcName.Value); ok {
		// Variant-typed values reach builtins as their dynamic type; coerce
		// them to the declared parameter types (DWScript variant casts).
		if errVal := e.coerceBuiltinArgsToSignature(funcName, node.Arguments, args, ctx); errVal != nil {
			return errVal
		}
		if ctx.Exception() != nil {
			return &runtime.NilValue{}
		}
		return fn(e, args)
	}

	// Implicit Self method calls (checked after built-ins to avoid shadowing)
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if _, ok := selfRaw.(Value); ok {
			return e.executeImplicitSelfCall(node, funcName, ctx)
		}
	}

	// Not found in any registry or context
	return e.newError(node, "undefined function: %s", funcName.Value)
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
			ref, err := e.prepareByRefArgument(arg, ctx)
			if err != nil {
				return nil, err
			}
			args[idx] = ref

		} else {
			// Regular parameter: use cached value
			args[idx] = cachedArgs[idx]
		}
	}

	return args, nil
}

func (e *Evaluator) prepareByRefArgument(arg ast.Expression, ctx *ExecutionContext) (Value, error) {
	if argIdent, ok := arg.(*ast.Identifier); ok {
		varName := argIdent.Value
		capturedEnv := ctx.Env()

		if val, exists := capturedEnv.Get(varName); exists {
			if refVal, isRef := val.(ReferenceAccessor); isRef {
				return refVal.(Value), nil
			}
		}

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

		return runtime.NewReferenceValue(varName, getter, setter), nil
	}

	// Array element by reference: bind to (array, index) so later resizes of
	// the array are observed, with bounds re-checked on every access
	// (DWScript raises "Upper/Lower bound exceeded!" at access time).
	if idxExpr, ok := arg.(*ast.IndexExpression); ok {
		if ref, handled, err := e.prepareArrayElementReference(idxExpr, ctx); handled {
			return ref, err
		}
	}

	// Field of a value that is itself a var parameter (e.g. r.y where r is a
	// byref array element): chain through the base reference so staleness of
	// the base is observed on each access.
	if memberExpr, ok := arg.(*ast.MemberAccessExpression); ok {
		if ref, handled, err := e.prepareMemberFieldReference(memberExpr, ctx); handled {
			return ref, err
		}
	}

	current, assign, err := e.EvaluateLValue(arg, ctx)
	if err != nil {
		return nil, fmt.Errorf("var parameter requires a variable, got %T", arg)
	}

	if refVal, isRef := current.(ReferenceAccessor); isRef {
		return refVal.(Value), nil
	}

	currentValue := current
	var getter runtime.GetterCallback = func() (runtime.Value, error) {
		if currentValue == nil {
			return &runtime.NilValue{}, nil
		}
		return currentValue, nil
	}
	var setter runtime.SetterCallback = func(val runtime.Value) error {
		value, ok := val.(Value)
		if !ok {
			return fmt.Errorf("var parameter assignment requires runtime value, got %T", val)
		}
		if err := assign(value); err != nil {
			return err
		}
		currentValue = value
		return nil
	}

	return runtime.NewReferenceValue(arg.String(), getter, setter), nil
}

// raiseBoundExceededError converts a boundExceededError from a stale array
// element reference into a catchable script exception with the plain message
// (DWScript omits the source location for var-param bound violations).
func (e *Evaluator) raiseBoundExceededError(err error, ctx *ExecutionContext) (Value, bool) {
	be, ok := err.(boundExceededError)
	if !ok {
		return nil, false
	}
	if ctx == nil {
		ctx = e.currentContext
	}
	if ctx != nil {
		ctx.SetException(e.createException("Exception", be.msg, nil, ctx))
	}
	return e.nilValue(), true
}

// boundExceededError marks an array-element var-param access whose index is no
// longer within the (possibly resized) array. Consumers convert it into a
// catchable DWScript exception with the plain message (no source location).
type boundExceededError struct{ msg string }

func (b boundExceededError) Error() string { return b.msg }

// arrayElementBounds reports the logical bounds of an array value.
func arrayElementBounds(arr *runtime.ArrayValue) (low, high int) {
	if arr.ArrayType != nil && arr.ArrayType.IsStatic() {
		return *arr.ArrayType.LowBound, *arr.ArrayType.HighBound
	}
	return 0, len(arr.Elements) - 1
}

// arrayElementPhysicalIndex converts a logical index into a physical index,
// returning a boundExceededError when out of range.
func arrayElementPhysicalIndex(arr *runtime.ArrayValue, index int) (int, error) {
	low, high := arrayElementBounds(arr)
	if index < low {
		return 0, boundExceededError{msg: fmt.Sprintf("Lower bound exceeded! Index %d", index)}
	}
	if index > high || index-low >= len(arr.Elements) {
		return 0, boundExceededError{msg: fmt.Sprintf("Upper bound exceeded! Index %d", index)}
	}
	return index - low, nil
}

// prepareArrayElementReference binds a var parameter to an array element as a
// live (array, index) reference. Reads and writes re-check bounds at access
// time, so resizes of the underlying dynamic array are observed (DWScript
// raises "Upper/Lower bound exceeded!" on stale accesses). Returns
// handled=false when the argument is not a plain array element (e.g. string
// index or default property), letting the generic lvalue path take over.
func (e *Evaluator) prepareArrayElementReference(idxExpr *ast.IndexExpression, ctx *ExecutionContext) (Value, bool, error) {
	arrRaw := e.Eval(idxExpr.Left, ctx)
	if isError(arrRaw) {
		return nil, false, nil
	}
	if ref, isRef := arrRaw.(ReferenceAccessor); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return nil, false, nil
		}
		arrRaw = deref
	}
	arr, ok := arrRaw.(*runtime.ArrayValue)
	if !ok || arr.ArrayType == nil {
		return nil, false, nil
	}

	idxVal := e.Eval(idxExpr.Index, ctx)
	if isError(idxVal) {
		return nil, false, nil
	}
	index, ok := ExtractIntegerIndex(idxVal)
	if !ok {
		return nil, false, nil
	}

	// Bind-time bounds check raises a catchable exception at the call site.
	if _, err := arrayElementPhysicalIndex(arr, index); err != nil {
		low, _ := arrayElementBounds(arr)
		e.raiseIndexBoundExceededAt(idxExpr.End(), index, index >= low)
		return nil, true, err
	}

	getter := func() (runtime.Value, error) {
		phys, err := arrayElementPhysicalIndex(arr, index)
		if err != nil {
			return nil, err
		}
		el := arr.Elements[phys]
		if el == nil {
			return e.getZeroValueForType(arr.ArrayType.ElementType), nil
		}
		return el, nil
	}
	setter := func(val runtime.Value) error {
		phys, err := arrayElementPhysicalIndex(arr, index)
		if err != nil {
			return err
		}
		arr.Elements[phys] = val
		return nil
	}

	return runtime.NewReferenceValue(idxExpr.String(), getter, setter), true, nil
}

// prepareMemberFieldReference builds a live reference for obj.field byref
// arguments whose base is itself a reference (var parameter). Each access
// re-dereferences the base, so bound violations of a stale base propagate.
// Returns handled=false when the base is not a reference or the member is not
// a plain record/object field.
func (e *Evaluator) prepareMemberFieldReference(memberExpr *ast.MemberAccessExpression, ctx *ExecutionContext) (Value, bool, error) {
	baseIdent, ok := memberExpr.Object.(*ast.Identifier)
	if !ok || memberExpr.Member == nil {
		return nil, false, nil
	}
	baseRaw, exists := ctx.Env().Get(baseIdent.Value)
	if !exists {
		return nil, false, nil
	}
	baseRef, ok := baseRaw.(ReferenceAccessor)
	if !ok {
		return nil, false, nil
	}

	// Only handle plain record/object field access; anything else (properties,
	// helpers) goes through the generic lvalue path.
	fieldName := memberExpr.Member.Value
	if cur, err := baseRef.Dereference(); err == nil && !memberFieldExists(cur, fieldName) {
		return nil, false, nil
	}

	getter := makeMemberFieldGetter(baseRef, fieldName)
	setter := makeMemberFieldSetter(baseRef, fieldName)
	return runtime.NewReferenceValue(memberExpr.String(), getter, setter), true, nil
}

// memberFieldExists reports whether cur is a record/object with a plain field
// of the given name.
func memberFieldExists(cur Value, fieldName string) bool {
	switch base := cur.(type) {
	case RecordInstanceValue:
		_, found := base.GetRecordField(fieldName)
		return found
	case ObjectValue:
		return base.GetField(fieldName) != nil
	default:
		return false
	}
}

func makeMemberFieldGetter(baseRef ReferenceAccessor, fieldName string) func() (runtime.Value, error) {
	return func() (runtime.Value, error) {
		cur, err := baseRef.Dereference()
		if err != nil {
			return nil, err
		}
		switch base := cur.(type) {
		case RecordInstanceValue:
			if val, found := base.GetRecordField(fieldName); found {
				return val, nil
			}
		case ObjectValue:
			if val := base.GetField(fieldName); val != nil {
				return val, nil
			}
		}
		return nil, fmt.Errorf("field '%s' not found", fieldName)
	}
}

func makeMemberFieldSetter(baseRef ReferenceAccessor, fieldName string) func(runtime.Value) error {
	return func(val runtime.Value) error {
		cur, err := baseRef.Dereference()
		if err != nil {
			return err
		}
		if setterVal, ok := cur.(RecordFieldSetter); ok {
			if setterVal.SetRecordField(fieldName, val) {
				return nil
			}
		}
		if setterVal, ok := cur.(ObjectFieldSetter); ok {
			setterVal.SetField(fieldName, val)
			return nil
		}
		return fmt.Errorf("cannot assign field '%s'", fieldName)
	}
}

func (e *Evaluator) prepareArgsForParameters(
	parameters []*ast.Parameter,
	argExprs []ast.Expression,
	ctx *ExecutionContext,
) ([]Value, error) {
	if len(argExprs) != len(parameters) {
		return nil, fmt.Errorf("wrong number of arguments: expected %d, got %d", len(parameters), len(argExprs))
	}

	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		param := parameters[idx]
		if param.IsLazy {
			args[idx] = e.wrapLazyArg(arg, ctx, func(expr ast.Expression) Value {
				return e.Eval(expr, ctx)
			})
			continue
		}
		if param.ByRef {
			ref, err := e.prepareByRefArgument(arg, ctx)
			if err != nil {
				return nil, err
			}
			args[idx] = ref
			continue
		}
		val := e.Eval(arg, ctx)
		if isError(val) {
			return nil, fmt.Errorf("%s", val.String())
		}
		args[idx] = val
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

// callExternalFunction prepares runtime arguments for an external Go function,
// then delegates value-level invocation to the interpreter shell.
func (e *Evaluator) callExternalFunction(
	funcName string,
	argExprs []ast.Expression,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	registry := e.ExternalFunctions()
	if registry == nil {
		return e.newError(node, "external function registry not initialized")
	}

	signature, ok := registry.Signature(funcName)
	if !ok {
		return e.newError(node, "external function '%s' not found", funcName)
	}

	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		isVarParam := idx < len(signature.VarParams) && signature.VarParams[idx]
		if isVarParam {
			ref, err := e.prepareByRefArgument(arg, ctx)
			if err != nil {
				return e.newError(arg, "%s", err.Error())
			}
			args[idx] = ref
			continue
		}

		var val Value
		if idx < len(signature.ParamTypes) {
			expectedType, err := e.resolveTypeName(signature.ParamTypes[idx], ctx)
			if err == nil {
				val = e.evalWithExpectedType(arg, expectedType, ctx)
			} else {
				val = e.Eval(arg, ctx)
			}
		} else {
			val = e.Eval(arg, ctx)
		}
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	return e.callExternalFunctionViaEngineState(funcName, args, node)
}

func (e *Evaluator) evalWithExpectedType(node ast.Node, expectedType types.Type, ctx *ExecutionContext) Value {
	if expectedType == nil {
		return e.Eval(node, ctx)
	}

	switch typed := types.GetUnderlyingType(expectedType).(type) {
	case *types.ArrayType:
		prev := ctx.ArrayTypeContext()
		ctx.SetArrayTypeContext(typed)
		defer ctx.SetArrayTypeContext(prev)
	case *types.RecordType:
		prev := ctx.RecordTypeContext()
		ctx.SetRecordTypeContext(typed.Name)
		defer ctx.SetRecordTypeContext(prev)
	}

	return e.Eval(node, ctx)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
// Handles class lookup, field initialization, and constructor execution.
// Validates abstract/external classes and supports implicit parameterless constructors.
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	className := node.ClassName.Value

	// evalArgsByValue evaluates the constructor arguments left to right.
	// Deferred until the constructor declaration is known so that var/lazy
	// parameters can be wrapped instead (see below).
	argsEvaluated := false
	var args []Value
	evalArgsByValue := func() Value {
		if argsEvaluated {
			return nil
		}
		argsEvaluated = true
		args = make([]Value, len(node.Arguments))
		for i, arg := range node.Arguments {
			val := e.Eval(arg, ctx)
			if isError(val) {
				return val
			}
			args[i] = val
		}
		return nil
	}

	// Look up class in type system
	classInfoAny := e.typeSystem.LookupClass(className)
	if classInfoAny == nil {
		if recordTypeRaw := e.typeSystem.LookupRecord(className); recordTypeRaw != nil {
			if recordType, ok := recordTypeRaw.(*RecordTypeValue); ok && recordType.HasStaticMethod("Create") {
				if errVal := evalArgsByValue(); errVal != nil {
					return errVal
				}
				return e.callRecordStaticMethod(recordType, "Create", args, node, ctx)
			}
		}

		// Try nested class lookup from current context.
		if currentClassRaw, ok := ctx.Env().Get("__CurrentClass__"); ok {
			if classMeta, ok := currentClassRaw.(ClassMetaValue); ok {
				if nested := classMeta.GetNestedClass(className); nested != nil {
					if nestedMeta, ok := nested.(ClassMetaValue); ok {
						classInfoAny = nestedMeta.GetClassInfo()
					}
				}
			}
		}
		if classInfoAny == nil {
			if selfRaw, ok := ctx.Env().Get("Self"); ok {
				if objVal, ok := selfRaw.(ObjectValue); ok {
					if classMeta, ok := objVal.GetClassType().(ClassMetaValue); ok {
						if nested := classMeta.GetNestedClass(className); nested != nil {
							if nestedMeta, ok := nested.(ClassMetaValue); ok {
								classInfoAny = nestedMeta.GetClassInfo()
							}
						}
					}
				}
			}
		}
		if classInfoAny == nil {
			return e.newError(node, "class '%s' not found", className)
		}
	}

	classInfo, ok := classInfoAny.(runtime.IClassInfo)
	if !ok {
		return e.newError(node, "class '%s' has invalid type", className)
	}

	// The "TClass.Create(args)" sugar (node token is the class name, not "new")
	// can resolve to a class method named Create rather than a constructor.
	if !ident.Equal(node.Token.Literal, "new") {
		if classOverloads := classInfo.GetClassMethodOverloads("Create"); len(classOverloads) > 0 {
			if errVal := evalArgsByValue(); errVal != nil {
				return errVal
			}
			merged := append(classInfo.GetConstructorOverloads("Create"), classOverloads...)
			if selected, err := e.selectOverload(classInfo.GetName(), "Create", merged, args); err == nil &&
				selected.IsClassMethod && !selected.IsConstructor {
				classValAny, cvErr := e.typeSystem.CreateClassValue(classInfo.GetName())
				if cvErr != nil {
					return e.newError(node, "failed to get class value: %s", cvErr.Error())
				}
				if cm, ok := classValAny.(ClassMetaValue); ok {
					return e.executeClassMethodDirect(cm, selected, args, node, ctx)
				}
			}
		}
	}

	// Validate class can be instantiated
	if classInfo.IsAbstract() {
		return e.newError(node, "Trying to create an instance of an abstract class")
	}
	if classInfo.IsExternal() {
		return e.newError(node, "cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Execute constructor: when the declaration is unambiguous and declares
	// var/lazy parameters, wrap the arguments (by-ref references / lazy
	// thunks) so writes inside the constructor reach the caller's variable
	// (see fixture oop_field). Otherwise evaluate them by value.
	constructor := classInfo.GetConstructor("Create")
	if constructor != nil && !constructor.IsOverload && len(constructor.Parameters) == len(node.Arguments) && hasVarOrLazyParams(constructor) {
		preparedArgs, err := e.prepareArgsForParameters(constructor.Parameters, node.Arguments, ctx)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		args = preparedArgs
		argsEvaluated = true
	} else if errVal := evalArgsByValue(); errVal != nil {
		return errVal
	}

	// Create object instance
	obj := runtime.NewObjectInstance(classInfo)

	if initErr := e.initializeObjectFields(classInfo, obj, node, ctx); initErr != nil {
		return initErr
	}

	if constructor != nil {
		if err := e.executeConstructorForObject(obj, "Create", args, node, ctx); err != nil {
			return e.newError(node, "constructor failed: %v", err)
		}
	} else if len(args) == 1 && e.typeSystem.IsClassDescendantOf(className, "Exception") {
		if strVal, ok := args[0].(*runtime.StringValue); ok {
			obj.SetField("Message", &runtime.StringValue{Value: strVal.Value})
		} else {
			obj.SetField("Message", &runtime.StringValue{Value: args[0].String()})
		}
	} else if len(args) == 2 && e.typeSystem.IsClassDescendantOf(className, "EHost") {
		excClass := args[0].String()
		if strVal, ok := args[0].(*runtime.StringValue); ok {
			excClass = strVal.Value
		}
		message := args[1].String()
		if strVal, ok := args[1].(*runtime.StringValue); ok {
			message = strVal.Value
		}
		obj.SetField("ExceptionClass", &runtime.StringValue{Value: excClass})
		obj.SetField("Message", &runtime.StringValue{Value: message})
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
