package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for function call and object creation AST nodes.
// These include function calls, constructor calls (new expressions), and lambda expressions.

// VisitCallExpression evaluates a function call expression.
//
// This implementation handles the following call types:
//
// **1. Function Pointer Calls**:
//   - Detects function pointer and lambda calls
//   - Handles lazy parameters (CreateLazyThunk for IsLazy params)
//   - Handles var parameters (CreateReferenceValue for ByRef params)
//   - Handles regular parameter evaluation
//   - Captures closure environment
//
// **2. Member Access Calls**:
//   - **Record/Interface/Object method calls**: obj.Method(args)
//   - **Unit-qualified function calls**: UnitName.FunctionName(args)
//   - **Class constructor calls**: TClass.Create(args)
//
// **3. User Function Calls**:
//   - User-defined function calls with overloading support
//   - Overload resolution based on argument types
//   - Lazy parameter creation (Jensen's Device pattern)
//   - Var parameter creation (pass-by-reference)
//
// **4. Implicit Self Method Calls**:
//   - Pattern: MethodName(args) where Self is in environment
//   - Converts to Self.MethodName(args)
//
// **5. Record Static Method Calls**:
//   - Pattern: MethodName(args) in record method context
//   - Uses __CurrentRecord__ from environment
//
// **6. Built-in Functions with Var Parameters**:
//   - Functions: Inc, Dec, Insert, Delete, SetLength, etc.
//
// **7. Default() Function**:
//   - Pattern: Default(TypeName)
//   - Returns zero value for type
//
// **8. Type Casts**:
//   - Pattern: TypeName(expression) for single-argument calls
//   - Supported types: Integer, Float, String, Boolean, Variant, Enum, Class
//
// **9. Built-in Functions**:
//   - Standard library functions (PrintLn, Length, Abs, etc.)
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	if node.Function == nil {
		return e.newError(node, "call expression missing function")
	}

	// Check for function pointer calls
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if valRaw, exists := ctx.Env().Get(funcIdent.Value); exists {
			val := valRaw.(Value)
			if val.Type() == "FUNCTION_POINTER" || val.Type() == "LAMBDA" || val.Type() == "METHOD_POINTER" {
				funcPtr, ok := val.(FunctionPointerCallable)
				if !ok {
					// Fallback to adapter for types not implementing the interface
					fallbackArgs := make([]Value, len(node.Arguments))
					for i, arg := range node.Arguments {
						fallbackArgs[i] = e.Eval(arg, ctx)
						if isError(fallbackArgs[i]) {
							return fallbackArgs[i]
						}
					}
					return e.adapter.CallFunctionPointer(val, fallbackArgs, node)
				}

				// Get the function AST for parameter metadata
				funcDecl, _ := funcPtr.GetFunctionDecl().(*ast.FunctionDecl)

				// Prepare arguments - handle lazy, var, and regular parameters
				args := make([]Value, len(node.Arguments))
				for idx, arg := range node.Arguments {
					// Check parameter flags (only for regular function pointers, not lambdas)
					isLazy := false
					isByRef := false
					if funcDecl != nil && idx < len(funcDecl.Parameters) {
						isLazy = funcDecl.Parameters[idx].IsLazy
						isByRef = funcDecl.Parameters[idx].ByRef
					}

					if isLazy {
						// For lazy parameters, create a LazyThunk with callback-based evaluation
						capturedArg := arg
						var evalCallback runtime.EvalCallback = func() runtime.Value {
							return e.adapter.EvalNode(capturedArg)
						}
						args[idx] = runtime.NewLazyThunk(capturedArg, evalCallback)
					} else if isByRef {
						// For var parameters, create a ReferenceValue with callback-based get/set
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
								if !capturedEnv.Set(varName, val) {
									return fmt.Errorf("failed to set variable %s", varName)
								}
								return nil
							}

							args[idx] = runtime.NewReferenceValue(varName, getter, setter)
						} else {
							return e.newError(arg, "var parameter requires a variable, got %T", arg)
						}
					} else {
						// For regular parameters, evaluate immediately
						argVal := e.Eval(arg, ctx)
						if isError(argVal) {
							return argVal
						}
						args[idx] = argVal
					}
				}

				// Build metadata and call via ExecuteFunctionPointerCall
				metadata := FunctionPointerMetadata{
					IsLambda:   funcPtr.IsLambda(),
					Lambda:     funcPtr.GetLambdaExpr(),
					Function:   funcPtr.GetFunctionDecl(),
					Closure:    funcPtr.GetClosure(),
					SelfObject: funcPtr.GetSelfObject(),
				}
				return e.adapter.ExecuteFunctionPointerCall(metadata, args, node)
			}
		}
	}

	// Member access calls: obj.Method(), UnitName.Func(), TClass.Create()
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Member method calls (record, interface, object)
		if objVal.Type() == "RECORD" || objVal.Type() == "INTERFACE" || objVal.Type() == "OBJECT" {
			// Evaluate arguments
			args := make([]Value, len(node.Arguments))
			for i, arg := range node.Arguments {
				val := e.Eval(arg, ctx)
				if isError(val) {
					return val
				}
				args[i] = val
			}

			// Create a synthetic MethodCallExpression for error reporting
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

			// Dispatch method call directly using existing infrastructure
			return e.DispatchMethodCall(objVal, memberAccess.Member.Value, args, mc, ctx)
		}

		// Unit-qualified function calls and class constructor calls
		if identNode, ok := memberAccess.Object.(*ast.Identifier); ok {
			// Check for class constructor first
			if e.typeSystem.HasClass(identNode.Value) {
				// Convert to MethodCallExpression for constructor/static method dispatch
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

			// Unit-qualified function calls
			if e.unitRegistry != nil {
				return e.adapter.CallQualifiedOrConstructor(node, memberAccess)
			}
		}

		return e.newError(node, "cannot call member expression that is not a method or unit-qualified function")
	}

	// Remaining call types require a simple identifier
	funcName, ok := node.Function.(*ast.Identifier)
	if !ok {
		return e.newError(node, "function call requires identifier or qualified name, got %T", node.Function)
	}

	// Check for user-defined functions (with potential overloading)
	funcNameLower := ident.Normalize(funcName.Value)
	if overloads := e.FunctionRegistry().Lookup(funcNameLower); len(overloads) > 0 {
		return e.adapter.CallUserFunctionWithOverloads(node, funcName)
	}

	// Implicit Self method calls (MethodName() is shorthand for Self.MethodName())
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok {
			if selfVal.Type() == "OBJECT" || selfVal.Type() == "CLASS" {
				return e.adapter.CallImplicitSelfMethod(node, funcName)
			}
		}
	}

	// Record static method calls
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				if rtmv, ok := recordVal.(RecordTypeMetaValue); ok {
					if rtmv.HasStaticMethod(funcName.Value) {
						return e.adapter.DispatchRecordStaticMethod(rtmv.GetRecordTypeName(), node, funcName)
					}
				} else {
					return e.adapter.CallRecordStaticMethod(node, funcName)
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
		// Delete(str, pos, count) modifies str in place
		if len(node.Arguments) == 3 {
			return e.builtinDeleteString(node.Arguments, ctx)
		}
	}

	// External (Go) functions that may need var parameter handling
	if e.externalFunctions != nil {
		return e.adapter.EvalNode(node)
	}

	// Default(TypeName) function - expects unevaluated type identifier
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.builtinDefault(node.Arguments, ctx)
	}

	// Type casts - TypeName(expression) for single-argument calls
	if len(node.Arguments) == 1 {
		result := e.evalTypeCast(funcName.Value, node.Arguments[0], ctx)
		if result != nil {
			return result
		}
	}

	// Standard built-in functions - evaluate all arguments first, then call
	args := make([]Value, len(node.Arguments))
	for idx, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Call built-in function directly from registry
	if fn, ok := builtins.DefaultRegistry.Lookup(funcName.Value); ok {
		return fn(e, args)
	}
	return e.newError(node, "builtin function '%s' not found in registry", funcName.Value)
}

// PrepareUserFunctionArgs prepares arguments for user function invocation.
// Handles lazy/var/regular parameter wrapping with callback pattern.
func (e *Evaluator) PrepareUserFunctionArgs(
	fn *ast.FunctionDecl,
	argExprs []ast.Expression,
	cachedArgs []Value,
	ctx *ExecutionContext,
	node ast.Node,
) ([]Value, error) {
	args := make([]Value, len(argExprs))

	for idx, arg := range argExprs {
		// Get parameter metadata
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

		if isLazy {
			// Lazy parameter - capture expression for deferred evaluation
			capturedArg := arg
			var evalCallback runtime.EvalCallback = func() runtime.Value {
				return e.Eval(capturedArg, ctx)
			}
			args[idx] = runtime.NewLazyThunk(capturedArg, evalCallback)

		} else if isByRef {
			// Var parameter - must be an lvalue (identifier)
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
				if !capturedEnv.Set(varName, val) {
					return fmt.Errorf("failed to set variable %s", varName)
				}
				return nil
			}

			args[idx] = runtime.NewReferenceValue(varName, getter, setter)

		} else {
			// Regular parameter - use cached value from overload resolution
			args[idx] = cachedArgs[idx]
		}
	}

	return args, nil
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
//
// **INSTANTIATION MODES**:
//
// **1. CLASS LOOKUP** (case-insensitive)
//   - Pattern: `new TMyClass` or `TMyClass.Create(...)`
//   - Searches class registry with case-insensitive comparison
//
// **2. RECORD TYPE DELEGATION**
//   - Pattern: `TMyRecord.Create(...)` where TMyRecord is a record type
//   - Converts NewExpression to MethodCallExpression for record static method handling
//
// **3. ABSTRACT CLASS CHECK**
//   - Error: "Trying to create an instance of an abstract class"
//
// **4. EXTERNAL CLASS CHECK**
//   - Error: "cannot instantiate external class 'X' - external classes are not supported"
//
// **5. OBJECT CREATION**
//   - Creates new ObjectInstance with reference to ClassInfo
//
// **6. FIELD INITIALIZATION**
//   - Phase A: Create temporary environment with class constants
//   - Phase B: Initialize each field (evaluate initializer or use zero value)
//
// **7. EXCEPTION CLASS HANDLING**
//   - EHost.Create(className, message) - Sets ExceptionClass and Message fields
//   - Other Exception.Create(message) - Sets Message field directly
//
// **8. CONSTRUCTOR RESOLUTION**
//   - Step A: Get default constructor name (falls back to "Create")
//   - Step B: Find constructor overloads in class hierarchy
//   - Step C: Implicit parameterless constructor if needed
//   - Step D: Resolve overload based on argument types
//
// **9. CONSTRUCTOR EXECUTION**
//   - Environment setup (Self, parameters, Result, __CurrentClass__)
//   - Argument evaluation and validation
//   - Body execution
//
// **SPECIAL BEHAVIORS**:
// - Case-insensitive class lookup
// - Default constructor pattern
// - Implicit parameterless constructor
// - Record type delegation
// - Exception handling shortcuts
// - Class constants in field initializers
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	// Get the class name
	className := node.ClassName.Value

	// Evaluate all constructor arguments
	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[i] = val
	}

	// Look up class via TypeSystem
	classInfoAny := e.typeSystem.LookupClass(className)
	if classInfoAny == nil {
		return e.newError(node, "class '%s' not found", className)
	}

	classInfo, ok := classInfoAny.(runtime.IClassInfo)
	if !ok {
		return e.newError(node, "class '%s' has invalid type", className)
	}

	// Check if trying to instantiate an abstract class
	if classInfo.IsAbstract() {
		return e.newError(node, "Trying to create an instance of an abstract class")
	}

	// Check if trying to instantiate an external class
	if classInfo.IsExternal() {
		return e.newError(node, "cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := runtime.NewObjectInstance(classInfo)

	// Initialize fields with default values or field initializers
	fieldTypes := classInfo.GetFieldTypesMap()
	fieldDecls := classInfo.GetFieldsMap()

	for fieldName, fieldTypeAny := range fieldTypes {
		var fieldValue Value
		if fieldDecl, hasDecl := fieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = e.Eval(fieldDecl.InitValue, ctx)
			if isError(fieldValue) {
				return e.newError(node, "failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			if fieldType, ok := fieldTypeAny.(types.Type); ok {
				fieldValue = e.getZeroValueForType(fieldType)
			} else {
				fieldValue = &runtime.NilValue{}
			}
		}
		obj.SetField(fieldName, fieldValue)
	}

	// Call constructor if it exists
	constructor := classInfo.GetConstructor("Create")
	if constructor != nil {
		err := e.adapter.ExecuteConstructor(obj, "Create", args)
		if err != nil {
			return e.newError(node, "constructor failed: %v", err)
		}
	} else if len(args) > 0 {
		return e.newError(node, "no constructor found for class '%s' with %d arguments", className, len(args))
	}

	return obj
}

// VisitNewArrayExpression evaluates a new array expression.
// Handles dimension evaluation and multi-dimensional array construction with default values.
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil new array expression")
	}

	if node.ElementTypeName == nil {
		return e.newError(node, "new array expression missing element type")
	}

	// Resolve the element type
	elementTypeName := node.ElementTypeName.Value
	elementType, typeErr := e.ResolveTypeWithContext(elementTypeName, ctx)
	if typeErr != nil {
		return e.newError(node, "unknown element type '%s': %s", elementTypeName, typeErr)
	}

	// Evaluate and validate dimensions
	dimensions, evalErr := e.evaluateDimensions(node.Dimensions, ctx, node)
	if evalErr != nil {
		return evalErr
	}

	// Create the multi-dimensional array directly
	return e.CreateMultiDimArray(elementType, dimensions)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
// Creates a lambda that captures the current scope.
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	return &runtime.FunctionPointerValue{
		Lambda:  node,
		Closure: ctx.Env(),
	}
}
