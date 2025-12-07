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
// **Task 3.5.23**: User Function Calls with Special Parameter Handling
// **Task 3.5.24**: Special Calls (Type Casts, Constructors, Implicit Self, Unit-Qualified)
//
// This implementation handles the following call types via delegation to adapter:
//
// **1. Function Pointer Calls** (lines 419-432, Task 3.5.23):
//   - Detects function pointer and lambda calls
//   - Delegates to adapter which handles:
//   - Lazy parameter creation (CreateLazyThunk for IsLazy params)
//   - Var parameter creation (CreateReferenceValue for ByRef params)
//   - Regular parameter evaluation
//   - Closure environment capture
//
// **2. Member Access Calls** (lines 434-456, Task 3.5.24):
//   - **Record/Interface/Object method calls**: obj.Method(args)
//   - Detects by evaluating object and checking type
//   - Delegates to adapter for method dispatch
//   - **Unit-qualified function calls**: UnitName.FunctionName(args)
//   - Detects by checking unitRegistry for unit name
//   - Delegates to adapter for qualified function resolution
//   - **Class constructor calls**: TClass.Create(args)
//   - Detects by checking if identifier is a class name
//   - Delegates to adapter for constructor dispatch and object instantiation
//
// **3. User Function Calls** (lines 465-479, Task 3.5.23):
//   - Detects user-defined function calls (with overloading support)
//   - Delegates to adapter which handles:
//   - Overload resolution based on argument types
//   - Lazy parameter creation (Jensen's Device pattern)
//   - Var parameter creation (pass-by-reference)
//   - Regular parameter evaluation (with caching to prevent double-eval)
//
// **4. Implicit Self Method Calls** (lines 481-490, Task 3.5.24):
//   - Pattern: MethodName(args) where Self is in environment
//   - Detects by checking for Self in environment
//   - Delegates to adapter which converts to Self.MethodName(args)
//
// **5. Record Static Method Calls** (lines 492-501, Task 3.5.24):
//   - Pattern: MethodName(args) in record method context
//   - Detects by checking for __CurrentRecord__ in environment
//   - Delegates to adapter for static method dispatch
//
// **6. Built-in Functions with Var Parameters** (lines 503-516, Task 3.5.24):
//   - Functions: Inc, Dec, Insert, Delete, SetLength, etc.
//   - Delegates to adapter for var parameter handling
//
// **7. Default() Function** (lines 524-529, Task 3.5.24):
//   - Pattern: Default(TypeName)
//   - Expects unevaluated type identifier
//   - Delegates to adapter for zero value creation
//
// **8. Type Casts** (lines 531-547, Task 3.5.24):
//   - Pattern: TypeName(expression) for single-argument calls
//   - Supported types: Integer, Float, String, Boolean, Variant, Enum, Class
//   - Delegates to adapter which calls evalTypeCast
//   - Falls through to built-in functions if not a type cast
//
// **9. Built-in Functions** (lines 549-562):
//   - Standard library functions (PrintLn, Length, Abs, etc.)
//   - Evaluates all arguments first, then delegates to adapter
//
// The adapter has access to CreateLazyThunk and CreateReferenceValue methods (Task 3.5.23)
// which enable proper handling of lazy and var parameters in all call contexts.
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	if node.Function == nil {
		return e.newError(node, "call expression missing function")
	}

	// Check for function pointer calls
	// Task 3.5.23: Function pointer calls with closure handling, lazy params, and var params
	// Task 3.5.70: Use direct environment access instead of adapter
	// Task 3.5.95: Migrated parameter preparation logic from Interpreter.evalCallExpression
	// Task 3.5.121: Migrated to use FunctionPointerCallable interface + ExecuteFunctionPointerCall
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if valRaw, exists := ctx.Env().Get(funcIdent.Value); exists {
			val := valRaw.(Value)
			if val.Type() == "FUNCTION_POINTER" || val.Type() == "LAMBDA" || val.Type() == "METHOD_POINTER" {
				// Task 3.5.121: Use FunctionPointerCallable interface for type-safe access
				funcPtr, ok := val.(FunctionPointerCallable)
				if !ok {
					// Fallback to adapter for types not implementing the interface
					// (should not happen for standard FunctionPointerValue)
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
						// Task 3.5.131d: Direct construction using runtime.NewLazyThunk with callback
						capturedArg := arg
						var evalCallback runtime.EvalCallback = func() runtime.Value {
							// Callback captures interpreter's eval via adapter.EvalNode
							return e.adapter.EvalNode(capturedArg)
						}
						args[idx] = runtime.NewLazyThunk(capturedArg, evalCallback)
					} else if isByRef {
						// For var parameters, create a ReferenceValue with callback-based get/set
						// Var parameters must be lvalues (variables)
						// Task 3.5.131d: Direct construction using runtime.NewReferenceValue with callbacks
						if argIdent, ok := arg.(*ast.Identifier); ok {
							varName := argIdent.Value
							capturedEnv := ctx.Env()

							var getter runtime.GetterCallback = func() (runtime.Value, error) {
								val, found := capturedEnv.Get(varName)
								if !found {
									return nil, fmt.Errorf("referenced variable %s not found", varName)
								}
								// Environment.Get returns interface{}, which will be a Value at runtime
								if runtimeVal, ok := val.(runtime.Value); ok {
									return runtimeVal, nil
								}
								return nil, fmt.Errorf("environment value is not a runtime.Value")
							}

							var setter runtime.SetterCallback = func(val runtime.Value) error {
								// Environment.Set expects interface{}
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

				// Task 3.5.121: Build metadata from interface getters and call via ExecuteFunctionPointerCall
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

	// Task 3.5.24: Member access calls (obj.Method(), UnitName.Func(), TClass.Create())
	// Task 3.5.96: Migrated to use adapter methods instead of EvalNode
	// Handles record methods, interface methods, object methods, unit-qualified functions, and constructor calls
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Task 3.5.96: Member method calls (record, interface, object)
		// Task 3.5.147: Migrated to use DispatchMethodCall directly
		// Examples: myRecord.GetValue(), myInterface.Process(), myObj.DoSomething()
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

		// Task 3.5.96: Unit-qualified function calls and class constructor calls
		// Task 3.5.147: Split handling for cleaner dispatch
		// Examples: Math.Sin(x), TMyClass.Create(args)
		if identNode, ok := memberAccess.Object.(*ast.Identifier); ok {
			// Task 3.5.147: Check for class constructor first (via TypeRegistry)
			// This creates a MethodCallExpression and dispatches via VisitMethodCallExpression
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

			// Task 3.5.147: Unit-qualified function calls (via unitRegistry)
			// Delegate to adapter for unit resolution and function dispatch
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
	// Task 3.5.23: Handle lazy and var parameters in user function calls
	// Task 3.5.67: Use direct FunctionRegistry access instead of adapter
	// Task 3.5.97a: Migrated to use CallUserFunctionWithOverloads adapter method
	funcNameLower := ident.Normalize(funcName.Value)
	if overloads := e.FunctionRegistry().Lookup(funcNameLower); len(overloads) > 0 {
		// Delegate to adapter for overload resolution and parameter preparation
		// The adapter handles:
		// 1. Overload resolution based on argument types
		// 2. Lazy parameter wrapping (LazyThunk)
		// 3. Var parameter wrapping (ReferenceValue)
		// 4. Calling the resolved user function
		return e.adapter.CallUserFunctionWithOverloads(node, funcName)
	}

	// Task 3.5.24: Record static method calls
	// Task 3.5.146: Use RecordTypeMetaValue interface for static method lookup.
	// When inside a record static method context, allows calling other static methods
	// Example: Inside record static method, calling Count() calls TRecord.Count()
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				// Task 3.5.146: Use RecordTypeMetaValue interface for direct static method lookup
				if rtmv, ok := recordVal.(RecordTypeMetaValue); ok {
					if rtmv.HasStaticMethod(funcName.Value) {
						// Static method exists - dispatch via simpler adapter method
						return e.adapter.DispatchRecordStaticMethod(rtmv.GetRecordTypeName(), node, funcName)
					}
					// Static method not found - fall through to other resolution attempts
				} else {
					// Fallback to deprecated adapter method for non-implementing types
					return e.adapter.CallRecordStaticMethod(node, funcName)
				}
			}
		}
	}

	// Task 3.5.24: Built-in functions with var parameter handling (modify arguments in place)
	// These functions require references to variables, not their values
	// Examples: Inc(x), Dec(y), Swap(a, b), SetLength(arr, 10)
	// Task 3.5.93: Inc, Dec, SetLength, Insert, Delete migrated to Evaluator
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
		// Only the 3-parameter form needs var parameter handling
		// Delete(str, pos, count) modifies str in place
		if len(node.Arguments) == 3 {
			return e.builtinDeleteString(node.Arguments, ctx)
		}
	}

	// Task 3.5.24: External (Go) functions that may need var parameter handling
	// External functions can declare var parameters in their signatures
	// Task 3.8.6.3: Only delegate if the function is actually an external function,
	// not just if the map is non-nil. Delegating all calls breaks environment chain
	// for helper methods where Self is defined in the evaluator's context.
	if e.externalFunctions != nil && e.externalFunctions.Has(funcName.Value) {
		return e.adapter.EvalNode(node)
	}

	// Task 3.5.94: Default(TypeName) function - expects unevaluated type identifier
	// Example: Default(Integer) returns 0, Default(String) returns ""
	// The type name is NOT evaluated as an expression
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.builtinDefault(node.Arguments, ctx)
	}

	// Task 3.5.94: Type casts - TypeName(expression) for single-argument calls
	// Examples: Integer(3.14), String(42), Boolean(1), TMyClass(someObject)
	// Supported types: Integer, Float, String, Boolean, Variant, Enum types, Class types
	// Falls through to built-in functions if not a type cast
	if len(node.Arguments) == 1 {
		result := e.evalTypeCast(funcName.Value, node.Arguments[0], ctx)
		// If type cast succeeded (not nil), return it
		// nil means it's not a type cast, so continue to built-in functions
		if result != nil {
			return result
		}
	}

	// Standard built-in functions - evaluate all arguments first, then call
	// Examples: PrintLn("hello"), Length(arr), Abs(-5), Sin(x)
	// All arguments are evaluated before calling the function (no lazy/var parameters)
	args := make([]Value, len(node.Arguments))
	for idx, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Call built-in function directly from registry
	// Task 3.5.143x: Use direct registry lookup instead of adapter delegation
	if fn, ok := builtins.DefaultRegistry.Lookup(funcName.Value); ok {
		return fn(e, args) // Evaluator implements builtins.Context
	}

	// Task 3.5.24: Implicit Self method calls (MethodName() is shorthand for Self.MethodName())
	// Task 3.5.97b: Migrated to use CallImplicitSelfMethod adapter method
	// When inside an instance method, calling MethodName() calls Self.MethodName()
	// Example: Inside method Foo(), calling Bar() means Self.Bar()
	// NOTE: This check is AFTER built-in functions to avoid shadowing built-ins like IntToStr
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok {
			if selfVal.Type() == "OBJECT" || selfVal.Type() == "CLASS" {
				return e.adapter.CallImplicitSelfMethod(node, funcName)
			}
		}
	}

	// Not found in any registry or context
	return e.newError(node, "function '%s' not found", funcName.Value)
}

// PrepareUserFunctionArgs prepares arguments for user function invocation.
// Task 3.5.144: Handles lazy/var/regular parameter wrapping with callback pattern.
//
// Parameters:
//   - fn: The resolved function declaration from overload resolution
//   - argExprs: The original argument AST expressions from the call site
//   - cachedArgs: Pre-evaluated argument values from overload resolution
//     (regular params use cached values, lazy params are nil)
//   - ctx: The execution context for environment access
//   - node: The call expression node for error reporting
//
// Returns:
//   - []Value: Final argument values with proper wrapping
//   - error: Error if var parameter is not an lvalue
func (e *Evaluator) PrepareUserFunctionArgs(
	fn *ast.FunctionDecl,
	argExprs []ast.Expression,
	cachedArgs []Value,
	ctx *ExecutionContext,
	node ast.Node,
) ([]Value, error) {
	args := make([]Value, len(argExprs))

	for idx, arg := range argExprs {
		// Get parameter metadata (if available)
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

		if isLazy {
			// Task 3.5.131d pattern: Lazy with callback
			// Capture the argument expression for deferred evaluation
			capturedArg := arg
			var evalCallback runtime.EvalCallback = func() runtime.Value {
				// Evaluate in the current context when forced
				return e.Eval(capturedArg, ctx)
			}
			args[idx] = runtime.NewLazyThunk(capturedArg, evalCallback)

		} else if isByRef {
			// Task 3.5.131d pattern: Var with callbacks
			// Var parameters must be lvalues (identifiers only)
			argIdent, ok := arg.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("var parameter requires a variable, got %T", arg)
			}

			varName := argIdent.Value
			capturedEnv := ctx.Env()

			// Getter callback: Read variable value from environment
			var getter runtime.GetterCallback = func() (runtime.Value, error) {
				val, found := capturedEnv.Get(varName)
				if !found {
					return nil, fmt.Errorf("variable %s not found", varName)
				}
				// Environment.Get returns interface{}, cast to runtime.Value
				if runtimeVal, ok := val.(runtime.Value); ok {
					return runtimeVal, nil
				}
				return nil, fmt.Errorf("environment value is not a runtime.Value")
			}

			// Setter callback: Write variable value to environment
			var setter runtime.SetterCallback = func(val runtime.Value) error {
				if !capturedEnv.Set(varName, val) {
					return fmt.Errorf("failed to set variable %s", varName)
				}
				return nil
			}

			args[idx] = runtime.NewReferenceValue(varName, getter, setter)

		} else {
			// Regular parameter: use cached value from overload resolution
			// This prevents double-evaluation of argument expressions
			args[idx] = cachedArgs[idx]
		}
	}

	return args, nil
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
//
// **COMPLEXITY**: High (~250 lines in original implementation)
// **STATUS**: Partial migration with argument evaluation in evaluator and object creation delegated to adapter
//
// **INSTANTIATION MODES** (evaluated in this order):
//
// **1. CLASS LOOKUP** (case-insensitive)
//   - Pattern: `new TMyClass` or `TMyClass.Create(...)`
//   - Searches class registry with case-insensitive comparison
//   - All class names in DWScript are case-insensitive by language spec
//   - Implementation: ~10 lines in original
//
// **2. RECORD TYPE DELEGATION**
//   - Pattern: `TMyRecord.Create(...)` where TMyRecord is a record type
//   - Detection: If class not found, check for record type in environment
//     with key `__record_type_<lowercase_name>`
//   - Action: Converts NewExpression to MethodCallExpression and delegates
//     to evalMethodCall for record static method handling
//   - This allows records to have static factory methods like classes
//   - Implementation: ~30 lines in original
//
// **3. ABSTRACT CLASS CHECK**
//   - Pattern: `new TAbstractClass` where class has `abstract` modifier
//   - Error: "Trying to create an instance of an abstract class"
//   - Prevents instantiation of classes meant only as base classes
//   - Implementation: ~4 lines in original
//
// **4. EXTERNAL CLASS CHECK**
//   - Pattern: `new TExternalClass` where class has `external` modifier
//   - Error: "cannot instantiate external class 'X' - external classes are not supported"
//   - External classes are for FFI integration (not yet supported)
//   - Implementation: ~6 lines in original
//
// **5. OBJECT CREATION**
//   - Action: Creates new ObjectInstance with reference to ClassInfo
//   - ObjectInstance contains field map, class reference, and VMT
//   - Implementation: ~2 lines in original
//
// **6. FIELD INITIALIZATION** (two-phase)
//   - **Phase A: Create temporary environment**
//   - Creates enclosed environment with class constants for field initializers
//   - Class constants are accessible during field initialization
//   - **Phase B: Initialize each field**
//   - If field has initializer expression: evaluate and assign
//   - Otherwise: use getZeroValueForType for appropriate default value
//   - Field types are used to determine correct zero values
//   - Error handling: Returns immediately if any initializer fails
//   - Implementation: ~30 lines in original
//
// **7. EXCEPTION CLASS HANDLING** (special cases)
//   - **EHost.Create(className, message)**:
//   - Pattern: `new EHost('SomeException', 'Error message')`
//   - Requires exactly 2 arguments
//   - Sets ExceptionClass and Message fields directly
//   - Returns immediately (no constructor body execution)
//   - **Other Exception.Create(message)**:
//   - Pattern: `new ESomeException('Error message')`
//   - Accepts single message argument
//   - Sets Message field directly
//   - Returns immediately (no constructor body execution)
//   - Detection via isExceptionClass() and InheritsFrom("EHost")
//   - Implementation: ~50 lines in original
//
// **8. CONSTRUCTOR RESOLUTION**
//   - **Step A: Get default constructor name**
//   - Checks class hierarchy for constructor marked as `default`
//   - Falls back to "Create" if no default constructor specified
//   - **Step B: Find constructor overloads**
//   - getMethodOverloadsInHierarchy() finds all constructors in hierarchy
//   - Case-insensitive lookup (DWScript standard)
//   - Includes inherited virtual constructors
//   - **Step C: Implicit parameterless constructor**
//   - If 0 arguments and no parameterless constructor exists,
//     return object with just field initialization (no constructor body)
//   - This allows classes without explicit Create() to be instantiated
//   - **Step D: Resolve overload**
//   - resolveMethodOverload() matches arguments to parameters
//   - Uses type compatibility and implicit conversions
//   - Error: Overload resolution failure messages
//   - Implementation: ~40 lines in original
//
// **9. CONSTRUCTOR EXECUTION**
//   - **Environment setup**:
//   - Creates enclosed method environment
//   - Binds `Self` to the new object instance
//   - Binds constructor parameters to evaluated arguments
//   - For constructors with return type: initializes `Result` variable
//   - Binds `__CurrentClass__` for ClassName access in constructor
//   - **Argument evaluation**:
//   - Evaluates each constructor argument in current environment
//   - Error propagation on evaluation failure
//   - **Argument count validation**:
//   - Error: "wrong number of arguments for constructor 'X': expected N, got M"
//   - **Body execution**:
//   - Executes constructor body via Eval()
//   - Error propagation on body failure
//   - **Environment restoration**:
//   - Restores previous environment after constructor completes
//   - Implementation: ~55 lines in original
//
// **10. RETURN VALUE**
//   - Returns the newly created ObjectInstance
//   - Object has all fields initialized and constructor executed
//
// **SPECIAL BEHAVIORS**:
//
// **Case-insensitive class lookup**:
//   - DWScript is case-insensitive by language spec
//   - Class names are matched without regard to case
//
// **Default constructor pattern**:
//   - Classes can mark a constructor as `default` for `new TClass` syntax
//   - Falls back to "Create" if no default specified
//   - Allows DSL-style APIs with custom constructor names
//
// **Implicit parameterless constructor**:
//   - Classes without explicit Create() can still be instantiated
//   - Fields are initialized but no constructor body runs
//   - Enables simple data classes without boilerplate
//
// **Record type delegation**:
//   - Parser creates NewExpression for `TRecord.Create(...)` syntax
//   - Evaluator converts to MethodCallExpression for proper handling
//   - Enables uniform syntax for class and record instantiation
//
// **Exception handling shortcuts**:
//   - Built-in exception constructors have special handling
//   - Bypasses normal constructor resolution for efficiency
//   - Sets Message field directly without constructor body
//
// **Class constants in field initializers**:
//   - Field initializers can reference class constants
//   - Temporary environment created with constants defined
//   - Enables `FMyField: Integer := MY_CONST + 1` patterns
//
// **DEPENDENCIES** (blockers for full migration):
//   - ClassInfo: Class metadata including fields, methods, constructors, parent
//   - ObjectInstance: Runtime object with fields, class reference, VMT
//   - RecordTypeValue: For record type detection and delegation
//   - ExceptionValue: For exception class detection
//   - Environment: Scope management for field initializers and constructor
//   - resolveMethodOverload(): Constructor overload resolution
//   - getMethodOverloadsInHierarchy(): Constructor lookup in class hierarchy
//   - getZeroValueForType(): Default value generation for field types
//   - ClassInfoValue: For __CurrentClass__ binding
//   - isExceptionClass(): Exception class detection helper
//   - InheritsFrom(): Class hierarchy traversal
//
// **MIGRATION STRATEGY**:
//   - Phase 1 (this task): Comprehensive documentation of all modes
//   - Phase 2 (future): Migrate simple class instantiation after ObjectInstance migration
//   - Phase 3 (future): Migrate field initialization after type system migration
//   - Phase 4 (future): Migrate constructor dispatch after method call migration
//   - Phase 5 (future): Migrate exception handling after exception system migration
//   - Phase 6 (future): Migrate record delegation after record type migration
//
// **ERROR CONDITIONS**:
//   - "class 'X' not found" - Class not in registry and not a record type
//   - "Trying to create an instance of an abstract class" - Abstract class instantiation
//   - "cannot instantiate external class 'X'" - External class instantiation
//   - "EHost.Create requires class name and message arguments" - Wrong EHost args
//   - Overload resolution errors - No matching constructor for arguments
//   - "wrong number of arguments for constructor 'X'" - Argument count mismatch
//   - Field initializer errors - Propagated from initializer evaluation
//   - Constructor body errors - Propagated from constructor execution
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

	// Task 3.5.22k: Direct object creation using IClassInfo interface.
	// Look up class via TypeSystem (returns any, which is *interp.ClassInfo)
	classInfoAny := e.typeSystem.LookupClass(className)
	if classInfoAny == nil {
		return e.newError(node, "class '%s' not found", className)
	}

	// Cast to IClassInfo interface for method access
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
	// Get field type mapping and field declarations
	fieldTypes := classInfo.GetFieldTypesMap()
	fieldDecls := classInfo.GetFieldsMap()

	for fieldName, fieldTypeAny := range fieldTypes {
		var fieldValue Value
		if fieldDecl, hasDecl := fieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			// Evaluate field initializer
			fieldValue = e.Eval(fieldDecl.InitValue, ctx)
			if isError(fieldValue) {
				return e.newError(node, "failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			// Use default value for type
			if fieldType, ok := fieldTypeAny.(types.Type); ok {
				fieldValue = e.getZeroValueForType(fieldType)
			} else {
				fieldValue = &runtime.NilValue{}
			}
		}
		obj.SetField(fieldName, fieldValue)
	}

	// Call constructor if it exists
	// Task 3.5.22k: Use IClassInfo.GetConstructor() for constructor lookup
	constructor := classInfo.GetConstructor("Create")
	if constructor != nil {
		// Execute constructor via adapter callback (constructor execution still requires interpreter)
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
// Task 3.5.124: Direct lambda creation without adapter.
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	return &runtime.FunctionPointerValue{
		Lambda:  node,
		Closure: ctx.Env(),
	}
}
