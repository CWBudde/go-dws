package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for expression AST nodes.
// Visitor pattern implementation for expressions.
//
// Expressions evaluate to values and can be nested (e.g., binary expressions
// contain left and right sub-expressions).

// ErrorValue represents a runtime error.
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting.
func (e *Evaluator) newError(_ ast.Node, format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// VisitIdentifier evaluates an identifier (variable reference).
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Self keyword refers to current object instance
	if node.Value == "Self" {
		val, ok := ctx.Env().Get("Self")
		if !ok {
			return e.newError(node, "Self used outside method context")
		}
		// Environment stores interface{}, cast to Value
		if selfVal, ok := val.(Value); ok {
			return selfVal
		}
		return e.newError(node, "Self has invalid type")
	}

	// Try to find identifier in current environment (variables, parameters, constants)
	// Task 3.5.70: Use direct environment access instead of adapter
	if valRaw, ok := ctx.Env().Get(node.Value); ok {
		val := valRaw.(Value)
		// Check if this is an external variable (not yet supported)
		// Task 3.5.73: Use type assertion instead of adapter
		if extVar, ok := val.(ExternalVarAccessor); ok {
			return e.newError(node, "Unsupported external variable access: %s", extVar.ExternalVarName())
		}

		// Check if this is a lazy parameter (LazyThunk)
		// If so, force evaluation - each access re-evaluates the expression
		// Task 3.5.73: Use type assertion instead of adapter
		if thunk, ok := val.(LazyEvaluator); ok {
			return thunk.Evaluate()
		}

		// Check if this is a var parameter (ReferenceValue)
		// If so, dereference it to get the actual value
		// Task 3.5.71: Use Type() check instead of adapter
		if val.Type() == "REFERENCE" {
			actualVal, err := e.adapter.DereferenceValue(val)
			if err != nil {
				return e.newError(node, "%s", err.Error())
			}
			return actualVal
		}

		// Variable found - return the value directly
		// All value types (primitives, arrays, objects, records) can be returned as-is
		return val
	}

	// Check if we're in an instance method context (Self is bound)
	// When Self is bound, identifiers can refer to instance fields, class variables,
	// properties, methods (auto-invoked if zero params), or ClassName/ClassType
	// Task 3.5.71: Use Type() check instead of adapter for IsObjectInstance
	if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
		if selfVal, ok := selfRaw.(Value); ok && selfVal.Type() == "OBJECT" {
			// Check for instance field
			if fieldValue, found := e.adapter.GetObjectFieldValue(selfVal, node.Value); found {
				return fieldValue
			}

			// Check for class variable
			if classVarValue, found := e.adapter.GetClassVariableValue(selfVal, node.Value); found {
				return classVarValue
			}

			// Check for property - but skip if we're in a property getter/setter to prevent recursion
			propCtx := ctx.PropContext()
			if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
				// Task 3.5.72: Use ObjectValue interface for direct property check
				if objVal, ok := selfVal.(ObjectValue); ok && objVal.HasProperty(node.Value) {
					propValue, err := e.adapter.ReadPropertyValue(selfVal, node.Value, node)
					if err != nil {
						return e.newError(node, "%s", err.Error())
					}
					return propValue
				}
			}

			// Check for method - auto-invoke if parameterless, or create method pointer
			// Task 3.5.72: Use ObjectValue interface for direct method check
			if objVal, ok := selfVal.(ObjectValue); ok && objVal.HasMethod(node.Value) {
				if e.adapter.IsMethodParameterless(selfVal, node.Value) {
					// Auto-invoke parameterless method
					return e.adapter.CreateMethodCall(selfVal, node.Value, node)
				}
				// Method with parameters - create method pointer
				methodPtr, err := e.adapter.CreateMethodPointerFromObject(selfVal, node.Value)
				if err != nil {
					return e.newError(node, "%s", err.Error())
				}
				return methodPtr
			}

			// Check for ClassName special identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassName") {
				className := e.adapter.GetClassName(selfVal)
				return &runtime.StringValue{Value: className}
			}

			// Check for ClassType special identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassType") {
				return e.adapter.GetClassType(selfVal)
			}
		}
	}

	// Check if we're in a class method context (__CurrentClass__ is bound)
	// Identifiers can refer to ClassName, ClassType, or class variables
	// Task 3.5.71: Use Type() check instead of adapter for IsClassInfoValue
	if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
		if classInfoVal, ok := currentClassRaw.(Value); ok && classInfoVal.Type() == "CLASSINFO" {
			// Check for ClassName identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassName") {
				className := e.adapter.GetClassNameFromClassInfo(classInfoVal)
				return &runtime.StringValue{Value: className}
			}

			// Check for ClassType identifier (case-insensitive)
			if ident.Equal(node.Value, "ClassType") {
				return e.adapter.GetClassTypeFromClassInfo(classInfoVal)
			}

			// Check for class variable
			if classVarValue, found := e.adapter.GetClassVariableFromClassInfo(classInfoVal, node.Value); found {
				return classVarValue
			}
		}
	}

	// Check if this identifier is a user-defined function name
	// Functions are auto-invoked if they have zero parameters, or converted to function pointers if they have parameters
	// Task 3.5.67: Use direct FunctionRegistry access instead of adapter
	// Task 3.5.85: Direct evaluation without adapter EvalNode call
	funcNameLower := ident.Normalize(node.Value)
	if overloads := e.FunctionRegistry().Lookup(funcNameLower); len(overloads) > 0 {
		// Find the appropriate overload
		var fn *ast.FunctionDecl
		if len(overloads) == 1 {
			fn = overloads[0]
		} else {
			// Multiple overloads - try to find the one with zero parameters
			for _, candidate := range overloads {
				if len(candidate.Parameters) == 0 {
					fn = candidate
					break
				}
			}
			// If no zero-param overload, default to first one (for function pointer use)
			if fn == nil {
				fn = overloads[0]
			}
		}

		// Check if function has zero parameters - auto-invoke
		if len(fn.Parameters) == 0 {
			return e.adapter.CallUserFunction(fn, []Value{})
		}

		// Function has parameters - create function pointer
		return e.adapter.CreateFunctionPointer(fn, ctx.Env())
	}

	// Check if this identifier is a class name (metaclass reference)
	// Task 3.5.64: Use direct TypeRegistry access instead of adapter
	// Task 3.5.85: Direct ClassValue creation without adapter EvalNode call
	if e.typeSystem.HasClass(node.Value) {
		classVal, err := e.adapter.CreateClassValue(node.Value)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return classVal
	}

	// Final check: check for built-in functions or return undefined error
	// Task 3.5.85: Direct built-in invocation without adapter EvalNode call
	if e.FunctionRegistry().IsBuiltin(node.Value) {
		// Parameterless built-in functions are auto-invoked
		return e.adapter.CallBuiltinFunction(node.Value, []Value{})
	}

	// Still not found - return error
	return e.newError(node, "undefined variable '%s'", node.Value)
}

// VisitBinaryExpression evaluates a binary expression (e.g., a + b, x == y).
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Handle short-circuit operators first (special evaluation order)
	switch node.Operator {
	case "??":
		return e.evalCoalesceOp(node, ctx)
	case "and":
		return e.evalAndOp(node, ctx)
	case "or":
		return e.evalOrOp(node, ctx)
	}

	// Evaluate both operands for non-short-circuit operators
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	right := e.Eval(node.Right, ctx)
	if isError(right) {
		return right
	}
	if right == nil {
		return e.newError(node.Right, "right operand evaluated to nil")
	}

	// Try operator overloading first (custom operators for objects)
	if result, ok := e.tryBinaryOperator(node.Operator, left, right, node); ok {
		return result
	}

	// Handle 'in' operator (membership testing)
	if node.Operator == "in" {
		return e.evalInOperator(left, right, node)
	}

	// Handle operations based on operand types
	// Check for Variant FIRST (Variant operations take precedence)
	if left.Type() == "VARIANT" || right.Type() == "VARIANT" {
		return e.evalVariantBinaryOp(node.Operator, left, right, node)
	}

	// Type-specific binary operations
	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return e.evalIntegerBinaryOp(node.Operator, left, right, node)

	case left.Type() == "FLOAT" || right.Type() == "FLOAT":
		return e.evalFloatBinaryOp(node.Operator, left, right, node)

	case left.Type() == "STRING" && right.Type() == "STRING":
		return e.evalStringBinaryOp(node.Operator, left, right, node)

	// Allow string concatenation with RTTI_TYPEINFO
	case (left.Type() == "STRING" && right.Type() == "RTTI_TYPEINFO") ||
		(left.Type() == "RTTI_TYPEINFO" && right.Type() == "STRING"):
		if node.Operator == "+" {
			return &runtime.StringValue{Value: left.String() + right.String()}
		}
		return e.newError(node, "type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())

	case left.Type() == "BOOLEAN" && right.Type() == "BOOLEAN":
		return e.evalBooleanBinaryOp(node.Operator, left, right, node)

	// Enum comparisons
	case left.Type() == "ENUM" && right.Type() == "ENUM":
		return e.evalEnumBinaryOp(node.Operator, left, right, node)

	// Object, interface, class, and nil comparisons (= and <>)
	case node.Operator == "=" || node.Operator == "<>":
		return e.evalEqualityComparison(node.Operator, left, right, node)

	default:
		return e.newError(node, "type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}

// VisitUnaryExpression evaluates a unary expression (e.g., -x, not b).
func (e *Evaluator) VisitUnaryExpression(node *ast.UnaryExpression, ctx *ExecutionContext) Value {
	// Evaluate the operand
	operand := e.Eval(node.Right, ctx)
	if isError(operand) {
		return operand
	}

	// Try operator overloading first (custom operators for objects)
	if result, ok := e.tryUnaryOperator(node.Operator, operand, node); ok {
		return result
	}

	// Handle standard unary operators
	switch node.Operator {
	case "-":
		return e.evalMinusUnaryOp(operand, node)
	case "+":
		return e.evalPlusUnaryOp(operand, node)
	case "not":
		return e.evalNotUnaryOp(operand, node)
	default:
		return e.newError(node, "unknown operator: %s%s", node.Operator, operand.Type())
	}
}

// VisitAddressOfExpression evaluates an address-of expression (@funcName or @obj.method).
// Creates function/method pointers that can be called later or assigned to variables.
//
// Task 3.5.37: Full migration of address-of operator evaluation.
//
// **SYNTAX FORMS**:
//   - `@FunctionName` - Creates a function pointer to a standalone function
//   - `@object.MethodName` - Creates a method pointer bound to an object instance
//
// **FUNCTION POINTERS**:
//   - Regular function/procedure references
//   - Resolved via function registry (case-insensitive lookup)
//   - For overloaded functions, the first overload is used
//   - The function pointer captures the closure environment
//
// **METHOD POINTERS** (procedure/function of object):
//   - Method references bound to a specific object instance
//   - The object is evaluated and stored in the pointer's SelfObject field
//   - When called, Self is bound to the captured object
//   - Enables callback patterns: `var callback: TNotifyEvent := @myObj.OnClick;`
//
// **TYPE INFORMATION**:
//   - Function pointers include type information (parameter types, return type)
//   - Method pointers are marked as "of object" type
//   - Used for type checking during assignment and calls
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	// The operator should be an identifier (function/procedure name) or member access (for methods)
	switch operand := node.Operator.(type) {
	case *ast.Identifier:
		// Regular function/procedure pointer: @FunctionName
		funcPtr, err := e.adapter.CreateFunctionPointerFromName(operand.Value, ctx.Env())
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return funcPtr

	case *ast.MemberAccessExpression:
		// Method pointer: @object.MethodName
		// First evaluate the object
		objectVal := e.Eval(operand.Object, ctx)
		if isError(objectVal) {
			return objectVal
		}

		// Get the method name
		methodName := operand.Member.Value

		// Create method pointer with the object as Self
		methodPtr, err := e.adapter.CreateMethodPointer(objectVal, methodName, ctx.Env())
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return methodPtr

	default:
		return e.newError(node, "address-of operator requires function or method name, got %T", operand)
	}
}

// VisitGroupedExpression evaluates a grouped expression (parenthesized).
func (e *Evaluator) VisitGroupedExpression(node *ast.GroupedExpression, ctx *ExecutionContext) Value {
	// Grouped expressions just evaluate their inner expression
	// Parentheses are only for precedence, they don't change the value
	return e.Eval(node.Expression, ctx)
}

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
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if valRaw, exists := ctx.Env().Get(funcIdent.Value); exists {
			val := valRaw.(Value)
			if val.Type() == "FUNCTION_POINTER" || val.Type() == "LAMBDA" {
				// Delegate to adapter which handles:
				// - Closure environment restoration
				// - Lazy parameter creation (CreateLazyThunk)
				// - Var parameter creation (CreateReferenceValue)
				// - Regular parameter evaluation
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Task 3.5.24: Member access calls (obj.Method(), UnitName.Func(), TClass.Create())
	// Handles record methods, interface methods, object methods, unit-qualified functions, and constructor calls
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Delegate record, interface, and object method calls to adapter
		// Examples: myRecord.GetValue(), myInterface.Process(), myObj.DoSomething()
		if objVal.Type() == "RECORD" || objVal.Type() == "INTERFACE" || objVal.Type() == "OBJECT" {
			return e.adapter.EvalNode(node)
		}

		// Task 3.5.24: Unit-qualified function calls and class constructor calls
		// Examples: Math.Sin(x), TMyClass.Create(args)
		// Task 3.5.64: Use direct TypeRegistry access instead of adapter
		if ident, ok := memberAccess.Object.(*ast.Identifier); ok {
			if e.unitRegistry != nil || e.typeSystem.HasClass(ident.Value) {
				return e.adapter.EvalNode(node)
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
	funcNameLower := ident.Normalize(funcName.Value)
	if overloads := e.FunctionRegistry().Lookup(funcNameLower); len(overloads) > 0 {
		// For now, delegate to adapter for overload resolution
		// In the future, this can be migrated to the evaluator
		// But we need to prepare arguments properly for lazy and var parameters

		// We can't determine which overload without evaluating arguments for type checking
		// So we delegate the entire call to the adapter, which will:
		// 1. Resolve the overload
		// 2. Prepare arguments with lazy thunks and references
		// 3. Call the user function
		return e.adapter.EvalNode(node)
	}

	// Task 3.5.24: Implicit Self method calls (MethodName() is shorthand for Self.MethodName())
	// When inside an instance method, calling MethodName() calls Self.MethodName()
	// Example: Inside method Foo(), calling Bar() means Self.Bar()
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok {
			if selfVal.Type() == "OBJECT" || selfVal.Type() == "CLASS" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Task 3.5.24: Record static method calls
	// When inside a record static method context, allows calling other static methods
	// Example: Inside record static method, calling Count() calls TRecord.Count()
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				return e.adapter.EvalNode(node)
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
	case "decodedate", "decodetime", "swap", "divmod", "trystrtoint", "trystrtofloat":
		// These still use the adapter (not yet migrated)
		return e.adapter.EvalNode(node)
	case "delete":
		// Only the 3-parameter form needs var parameter handling
		// Delete(str, pos, count) modifies str in place
		if len(node.Arguments) == 3 {
			return e.builtinDeleteString(node.Arguments, ctx)
		}
	}

	// Task 3.5.24: External (Go) functions that may need var parameter handling
	// External functions can declare var parameters in their signatures
	if e.externalFunctions != nil {
		return e.adapter.EvalNode(node)
	}

	// Task 3.5.24: Default(TypeName) function - expects unevaluated type identifier
	// Example: Default(Integer) returns 0, Default(String) returns ""
	// The type name is NOT evaluated as an expression
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.adapter.EvalNode(node)
	}

	// Task 3.5.24: Type casts - TypeName(expression) for single-argument calls
	// Examples: Integer(3.14), String(42), Boolean(1), TMyClass(someObject)
	// Supported types: Integer, Float, String, Boolean, Variant, Enum types, Class types
	// Falls through to built-in functions if not a type cast
	if len(node.Arguments) == 1 {
		result := e.adapter.EvalNode(node)
		// If type cast succeeded or there's a real error (not "unknown function"), return it
		if result != nil && !isError(result) {
			return result
		}
		if isError(result) {
			if !strings.Contains(result.String(), "unknown function") &&
				!strings.Contains(result.String(), "undefined identifier") {
				return result
			}
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

	// Call built-in function via adapter
	return e.adapter.CallBuiltinFunction(funcName.Value, args)
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
//   - Phase 1 (this task): Comprehensive documentation of all modes ✓
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

	// Create the object using the adapter
	// The adapter handles: class lookup, field initialization, and constructor execution
	// NOTE: Record type delegation, abstract/external checks, and exception handling
	// are still handled by the original evalNewExpression in objects_instantiation.go
	obj, err := e.adapter.CreateObject(className, args)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	return obj
}

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
//
// **COMPLEXITY**: Very High (700+ lines in original implementation)
// **STATUS**: Documentation-only migration with full adapter delegation
// **Task 3.5.25**: Simple Modes (Unit-qualified, Enum, Static class, Record fields)
// **Task 3.5.26**: Complex Modes (Object instance, Interface, Metaclass, Type casts)
//
// **11 DISTINCT ACCESS MODES** (evaluated in this order):
//
// === SIMPLE MODES (Task 3.5.25) ===
//
// **1. UNIT-QUALIFIED ACCESS** (UnitName.Symbol) [Task 3.5.25]
//   - Pattern: `Math.PI`, `System.Print`
//   - Evaluation order:
//     a. Check if left side is a registered unit name (via unitRegistry)
//     b. Try to resolve as qualified variable/constant (ResolveQualifiedVariable)
//     c. If not a variable, it might be a function reference (handled in VisitCallExpression)
//   - Error: "qualified name 'Unit.Symbol' cannot be used as a value (functions must be called)"
//   - Implementation: ~14 lines in original
//
// **2. STATIC CLASS ACCESS** (TClass.Member) [Task 3.5.25]
//   - Pattern: `TMyClass.ClassVar`, `TMyClass.Create`, `TMyClass.ClassName`
//   - Lookup order (case-insensitive):
//     a. Built-in properties: `ClassName` (string), `ClassType` (metaclass reference)
//     b. Class variables (lookupClassVar) - inherited from parent classes
//     c. Class constants (getClassConstant) - lazy evaluation with caching
//     d. Class properties (lookupProperty) - if IsClassProperty or uses class-level read specs
//     e. Constructors (HasConstructor) - auto-invoke with 0 arguments if no parentheses
//   - Supports constructor overloading and inheritance
//   - Falls back to implicit parameterless constructor if needed
//     f. Class methods (lookupClassMethodInHierarchy) - static methods
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: return as FunctionPointerValue
//   - Error: "member 'X' not found in class 'Y'"
//   - Implementation: ~100 lines in original
//
// **3. ENUM TYPE ACCESS** (TColor.Red, TColor.Low, TColor.High) [Task 3.5.25]
//   - Pattern: `TColor.Red`, `TMyEnum.Low`, `TMyEnum.High`
//   - Lookup in environment: `__enum_type_` + lowercase(enumTypeName)
//   - For scoped enums:
//     a. Look up enum value in EnumType.Values (takes precedence over properties)
//     b. Check for special properties: `Low` (lowest ordinal), `High` (highest ordinal)
//   - For unscoped enums: also check environment for value name
//   - Returns: EnumValue or IntegerValue (for Low/High)
//   - Error: "enum value 'X' not found in enum 'Y'"
//   - Implementation: ~45 lines in original
//
// **4. RECORD TYPE STATIC ACCESS** (TPoint.cOrigin, TPoint.Count) [Task 3.5.25]
//   - Pattern: `TPoint.cOrigin`, `TRecord.ClassMethod()`
//   - Lookup in environment: `__record_type_` + lowercase(recordTypeName)
//   - Lookup order (case-insensitive):
//     a. Constants (RecordTypeValue.Constants)
//     b. Class variables (RecordTypeValue.ClassVars)
//     c. Class methods (RecordTypeValue.ClassMethods)
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: error (requires parentheses to call)
//   - Error: "member 'X' not found in record type 'Y'"
//   - Implementation: ~40 lines in original
//
// **5. RECORD INSTANCE ACCESS** (record.Field, record.Method) [Task 3.5.25]
//   - Pattern: `point.X`, `point.GetLength()`, `point.Prop`
//   - Object type: RecordValue
//   - Lookup order (case-insensitive):
//     a. Direct field access (RecordValue.Fields)
//     b. Properties (RecordType.Properties):
//   - ReadField: field name → direct access, method name → call getter
//   - Write-only: error "property 'X' is write-only"
//     c. Instance methods (RecordValue.Methods):
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: error "method 'X' requires N parameter(s); use parentheses"
//     d. Class methods (from RecordTypeValue, accessible via instance)
//     e. Constants (from RecordTypeValue, accessible via instance)
//     f. Class variables (from RecordTypeValue, accessible via instance)
//     g. Helper properties (findHelperProperty → evalHelperPropertyRead)
//   - Error: "field 'X' not found in record 'Y'"
//   - Implementation: ~115 lines in original
//
// **10. ENUM VALUE PROPERTIES** (enumVal.Value) [Task 3.5.25]
//   - Pattern: `TColor.Red.Value` (returns ordinal as integer)
//   - Object type: EnumValue
//   - Supported properties:
//     a. `.Value` (case-insensitive): returns OrdinalValue as IntegerValue
//     b. `.ToString`: handled by helpers (if available)
//   - Fallback: Check helpers for additional properties
//   - Implementation: ~10 lines in original
//   - NOTE: This is listed here (out of precedence order) for Task 3.5.25 grouping
//   - In actual implementation, this is checked at position 10
//
// === COMPLEX MODES (Task 3.5.26) ===
//
// **6. CLASS/METACLASS ACCESS** (ClassInfoValue/ClassValue.Member) [Task 3.5.26]
//   - Pattern: When a class name is evaluated to ClassInfoValue or ClassValue
//   - Example: `var c := TMyClass; c.Create()`
//   - Lookup order (same as static class access #2):
//     a. Built-in properties: `ClassName`, `ClassType`
//     b. Class variables, constants, properties, constructors, class methods
//   - Returns: String/ClassValue/field value/method pointer
//   - Implementation: ~95 lines in original
//
// **7. INTERFACE INSTANCE ACCESS** (interface.Method, interface.Property) [Task 3.5.26]
//   - Pattern: `intfVar.Hello`, `intfVar.SomeMethod`
//   - Object type: InterfaceInstance
//   - Validation: Verify member exists in interface definition (HasMethod)
//   - For methods:
//   - Look up implementation in underlying object's class (getMethodOverloadsInHierarchy)
//   - Return FunctionPointerValue bound to underlying object (NO auto-invoke)
//   - Enables method delegate assignment: `var h : procedure := i.Hello;`
//   - For properties/fields: delegate to underlying object (without validation currently)
//   - Unwrap interface to underlying object and continue evaluation
//   - Error: "Interface is nil" or "method 'X' declared in interface 'Y' but not implemented"
//   - Implementation: ~50 lines in original
//
// **8. TYPE CAST VALUE HANDLING** (TBase(child).ClassVar) [Task 3.5.26]
//   - Pattern: Accessing members through a type cast expression
//   - Object type: TypeCastValue
//   - Extracts: StaticType (for class variable lookup), Object (actual instance)
//   - Purpose: Class variables use static type, not runtime type
//   - Unwraps to actual object and continues evaluation with static type context
//   - Implementation: ~5 lines in original
//
// **9. NIL OBJECT HANDLING** (nil.ClassVar) [Task 3.5.26]
//   - Pattern: `var o: TMyClass := nil; o.ClassVar`
//   - Object type: NilValue (with ClassType field) or nil evaluation result
//   - Special case: Accessing class variables on nil instances is allowed
//   - Lookup:
//     a. If staticClassType from cast (TBase(nil).ClassVar): use static type
//     b. If NilValue.ClassType set: look up class and check for class variable
//   - Success: Return class variable value
//   - Failure: Error "Object not instantiated" (for instance members)
//   - Implementation: ~35 lines in original
//
// **11. OBJECT INSTANCE ACCESS** (obj.Field, obj.Method, obj.Property) [Task 3.5.26]
//   - Pattern: `myObj.Name`, `myObj.GetValue()`, `myObj.Count`
//   - Object type: ObjectInstance
//   - Built-in properties (inherited from TObject, case-insensitive):
//     a. `ClassName`: returns obj.Class.Name (runtime type)
//     b. `ClassType`: returns ClassValue (metaclass for runtime type)
//   - Lookup order (case-insensitive):
//     a. Properties (Class.lookupProperty) - takes precedence over fields
//   - Call evalPropertyRead for read accessor (field, method, or expression)
//     b. Direct field access (obj.GetField) - instance fields
//     c. Class variables (lookupClassVar) - accessible from instance
//   - Uses static type from cast if available (e.g., TBase(child).ClassVar)
//     d. Class constants (getClassConstant) - accessible from instance
//     e. Instance methods (getMethodOverloadsInHierarchy):
//   - Check all overloads for parameterless variants
//   - Parameterless: auto-invoke via VisitMethodCallExpression
//   - With parameters: return first overload as FunctionPointerValue
//     f. Class methods (getMethodOverloadsInHierarchy with classMethod=true)
//   - Same logic as instance methods (auto-invoke or function pointer)
//     g. Helper properties (findHelperProperty → evalHelperPropertyRead)
//   - Error: "field 'X' not found in class 'Y'"
//   - Implementation: ~115 lines in original
//
// **SPECIAL BEHAVIORS**:
// - **Auto-invocation**: Parameterless methods/properties auto-invoke when accessed without ()
// - **Case-insensitive**: All name lookups are case-insensitive (DWScript spec)
// - **Inheritance**: Class variables, constants, properties, methods searched up hierarchy
// - **Helper support**: Type helpers can add properties/methods to any type
// - **Function pointers**: Methods with parameters return FunctionPointerValue
// - **Lazy evaluation**: Class constants evaluated once and cached on first access
// - **Type safety**: Static types respected for class variable access through casts
//
// **DEPENDENCIES** (blockers for full migration):
// - RecordValue, RecordTypeValue - in internal/interp (needs migration to runtime)
// - ObjectInstance, ClassInfo, ClassValue - in internal/interp (needs migration to runtime)
// - InterfaceInstance - in internal/interp (needs migration to runtime)
// - EnumValue, EnumTypeValue - in internal/interp (needs migration to runtime)
// - FunctionPointerValue - in internal/interp (needs migration to runtime)
// - TypeCastValue - in internal/interp (needs migration to runtime)
// - ClassInfoValue - in internal/interp (needs migration to runtime)
// - Helper infrastructure - findHelperProperty, findHelperMethod (needs adapter methods)
// - Method call infrastructure - evalMethodCall (delegated to VisitMethodCallExpression)
// - Property read infrastructure - evalPropertyRead, evalHelperPropertyRead (needs adapter)
// - Class hierarchy - lookupClassVar, lookupProperty, getMethodOverloadsInHierarchy (needs adapter)
// - Unit registry - ResolveQualifiedVariable (already in Evaluator via unitRegistry)
//
// **TESTING**:
// - Unit-qualified access (Math.PI, System.WriteLine)
// - Static class access (TMyClass.ClassVar, TMyClass.Create, TMyClass.ClassName)
// - Enum type access (TColor.Red, TColor.Low, TColor.High)
// - Record type static access (TPoint.cOrigin, TPoint.Count)
// - Record instance access (point.X, point.GetLength())
// - Object instance access (obj.Name, obj.GetValue(), obj.Count)
// - Interface method access (intf.Hello)
// - Type cast access (TBase(child).ClassVar)
// - Nil object access (nil.ClassVar, nil.Name → error)
// - Enum value properties (TColor.Red.Value)
// - Helper properties/methods (arr.Length, str.ToUpper)
// - Auto-invocation (obj.Method without parentheses for parameterless)
// - Function pointers (obj.Method with parameters returns pointer)
//
// **IMPLEMENTATION SUMMARY**:
// - Original implementation: 706 lines (objects_hierarchy.go:13-719)
// - Handles 11 distinct access modes with complex precedence rules
// - Supports case-insensitive lookups, inheritance, helpers, auto-invocation
// - Requires extensive value type infrastructure not yet in runtime package
// - Full migration deferred - will be broken into category-specific sub-tasks
//
// **MIGRATION STRATEGY**:
// - Phase 1 (this task): Comprehensive documentation of all access modes
// - Phase 2 (future): Migrate simple cases (built-in properties, direct field access)
// - Phase 3 (future): Migrate class/record static access after type system migration
// - Phase 4 (future): Migrate helper infrastructure after helper system migration
// - Phase 5 (future): Migrate method/property dispatch after OOP infrastructure migration
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	if node.Object == nil {
		return e.newError(node, "member access missing object")
	}
	if node.Member == nil {
		return e.newError(node, "member access missing member")
	}

	// Task 3.5.25 & 3.5.26: Implement member access with routing based on object type

	// Evaluate the object first
	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}

	memberName := node.Member.Value

	// Route based on object type
	switch obj.Type() {
	case "OBJECT":
		// Task 3.5.26: Object instance access (Mode 11)
		// Pattern: obj.Field, obj.Property, obj.Method
		// Lookup order per spec (lines 894-900): Properties → Fields → Class Variables

		// Task 3.5.86: Use ObjectValue interface for direct member access
		objVal, ok := obj.(ObjectValue)
		if !ok {
			return e.adapter.EvalNode(node)
		}

		// Try property access first (with recursion protection)
		propCtx := ctx.PropContext()
		if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
			// Task 3.5.72: Use ObjectValue interface for direct property check
			if objVal.HasProperty(memberName) {
				// Property reading still uses adapter due to complex method invocation logic
				propValue, err := e.adapter.ReadPropertyValue(obj, memberName, node)
				if err == nil {
					return propValue
				}
			}
		}

		// Task 3.5.86: Direct field access via ObjectValue interface
		if fieldValue := objVal.GetField(memberName); fieldValue != nil {
			return fieldValue
		}

		// Task 3.5.86: Direct class variable access via ObjectValue interface
		if classVarValue, found := objVal.GetClassVar(memberName); found {
			return classVarValue
		}

		// Try method or other member access via adapter
		return e.adapter.EvalNode(node)

	case "INTERFACE":
		// Task 3.5.26: Interface instance access (Mode 7)
		// Pattern: intf.Method, intf.Property

		// Task 3.5.87: Use InterfaceInstanceValue interface for direct member verification
		ifaceVal, ok := obj.(InterfaceInstanceValue)
		if !ok {
			return e.adapter.EvalNode(node)
		}

		// Get underlying object - nil check is critical for interface access
		underlying := ifaceVal.GetUnderlyingObjectValue()
		if underlying == nil {
			// Nil interface access - let adapter handle the error
			return e.adapter.EvalNode(node)
		}

		// Verify the member is part of the interface contract before delegating
		if ifaceVal.HasInterfaceProperty(memberName) {
			// Property access: delegate to underlying object's property reading
			// Uses adapter due to complex method invocation logic for property getters
			propValue, err := e.adapter.ReadPropertyValue(underlying, memberName, node)
			if err == nil {
				return propValue
			}
			// Property read failed - fallback to adapter for error handling
			return e.adapter.EvalNode(node)
		}

		// Method access or unknown member - delegate to adapter
		// Methods require complex dispatch logic (virtual method tables, etc.)
		return e.adapter.EvalNode(node)

	case "CLASS", "CLASS_INFO":
		// Task 3.5.88: Metaclass access (Mode 6)
		// Pattern: ClassValue.Member, ClassInfoValue.Member
		// Handle built-in properties and class variables/constants directly

		// Try to use ClassMetaValue interface for direct access
		classMetaVal, ok := obj.(ClassMetaValue)
		if !ok {
			// Fallback to adapter if interface not implemented
			return e.adapter.EvalNode(node)
		}

		// Handle built-in properties first
		if ident.Equal(memberName, "ClassName") {
			// ClassName property returns the class name as a string
			return &runtime.StringValue{Value: classMetaVal.GetClassName()}
		}
		if ident.Equal(memberName, "ClassType") {
			// ClassType property returns the class itself (for CLASS type, return self)
			// For CLASS_INFO, we return the same value wrapped as CLASS
			return obj
		}

		// Try class variables - direct lookup without adapter
		if val, found := classMetaVal.GetClassVar(memberName); found {
			return val
		}

		// Try class constants - direct lookup without adapter
		if val, found := classMetaVal.GetClassConstant(memberName); found {
			return val
		}

		// Complex cases (constructors, class methods, properties) need adapter
		// because they require method invocation logic
		return e.adapter.EvalNode(node)

	case "TYPE_CAST":
		// Task 3.5.89: Type cast value handling (Mode 8)
		// Pattern: TBase(child).ClassVar
		// Uses static type from cast for class variable lookup, not runtime type.

		// Try TypeCastAccessor interface for direct access
		typeCastVal, ok := obj.(TypeCastAccessor)
		if !ok {
			// Fallback to adapter if interface not implemented
			return e.adapter.EvalNode(node)
		}

		// First, try class variable lookup using the STATIC type
		// This is the key behavior: TBase(child).ClassVar accesses TBase's class var
		if classVarValue, found := typeCastVal.GetStaticClassVar(memberName); found {
			return classVarValue
		}

		// Get the wrapped value for further processing
		wrappedValue := typeCastVal.GetWrappedValue()

		// If wrapped value is an object, try field access and property reading
		if objVal, ok := wrappedValue.(ObjectValue); ok {
			// Try direct field access on the wrapped object
			if fieldValue := objVal.GetField(memberName); fieldValue != nil {
				return fieldValue
			}

			// Try property access on the wrapped object (uses adapter for getter logic)
			if objVal.HasProperty(memberName) {
				propValue, err := e.adapter.ReadPropertyValue(wrappedValue, memberName, node)
				if err == nil {
					return propValue
				}
			}
		}

		// For method calls and complex cases, delegate to adapter
		return e.adapter.EvalNode(node)

	case "NIL":
		// Task 3.5.90: Nil object handling (Mode 9)
		// Typed nil values can access class variables, but not instance members.

		// Try NilAccessor interface to get typed class name
		nilVal, ok := obj.(NilAccessor)
		if !ok {
			// Untyped nil - always error
			return e.newError(node, "Object not instantiated")
		}

		typedClassName := nilVal.GetTypedClassName()
		if typedClassName == "" {
			// Untyped nil - always error
			return e.newError(node, "Object not instantiated")
		}

		// Typed nil: Try to look up class variable via adapter
		// The adapter can access the class registry and lookup class variables
		// Delegate to adapter for class variable lookup with proper static type
		return e.adapter.EvalNode(node)

	case "RECORD":
		// Task 3.5.91: Record instance access (Mode 5)
		// Pattern: record.Field, record.Method
		// Use RecordInstanceValue interface for direct field access
		if recVal, ok := obj.(RecordInstanceValue); ok {
			// Direct field access - most common case
			if fieldVal, found := recVal.GetRecordField(memberName); found {
				return fieldVal
			}

			// Method reference - still uses adapter for method invocation
			if recVal.HasRecordMethod(memberName) {
				return e.adapter.EvalNode(node)
			}

			// Property check (records don't have properties, but check for consistency)
			if recVal.HasRecordProperty(memberName) {
				return e.adapter.EvalNode(node)
			}
		}
		// Fallback to adapter for other record member access patterns
		return e.adapter.EvalNode(node)

	case "ENUM":
		// Task 3.5.89: Enum value properties (Mode 10)
		// Pattern: enumVal.Value
		// Handle .Value property directly via EnumAccessor interface
		if enumVal, ok := obj.(EnumAccessor); ok {
			if ident.Equal(memberName, "Value") {
				return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}
			}
		}
		// Other enum properties (.Name, .ToString, etc.) handled by helpers via adapter
		return e.adapter.EvalNode(node)

	default:
		// For other types (identifiers that might be unit names, class names, enum types, etc.)
		// delegate to adapter for full handling
		return e.adapter.EvalNode(node)
	}
}

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
//
// **COMPLEXITY**: Very High (1,116 lines in original implementation)
// **STATUS**: Documentation-only migration with full adapter delegation
//
// **15 DISTINCT METHOD CALL MODES** (evaluated in this order):
//
// **1. UNIT-QUALIFIED FUNCTION CALLS** (UnitName.FunctionName())
//   - Pattern: `Math.Sin(x)`, `System.WriteLine(s)`
//   - Detection: Object is Identifier referring to registered unit name
//   - Process:
//     a. Check if identifier is registered unit (via unitRegistry.GetUnit)
//     b. Resolve qualified function (ResolveQualifiedFunction)
//     c. Evaluate all arguments
//     d. Call user function (callUserFunction)
//   - Error: "function 'X' not found in unit 'Y'"
//   - Implementation: ~20 lines in original
//
// **2. STATIC CLASS METHOD CALLS** (TClass.Method())
//   - Pattern: `TMyClass.ClassMethod()`, `TMyClass.Create()`
//   - Detection: Object is Identifier referring to registered class name
//   - Lookup order:
//     a. Collect class method overloads (getMethodOverloadsInHierarchy with isClassMethod=true)
//     b. Collect instance method overloads including constructors (isClassMethod=false)
//     c. Special: If constructor with 0 args and no parameterless constructor exists,
//     create object with implicit parameterless constructor (just initialize fields)
//     d. Resolve overload based on argument types (resolveMethodOverload)
//     e. If class method: execute with Self bound to ClassInfoValue (executeClassMethod)
//     f. If instance method/constructor: create new object, initialize fields,
//     execute method with Self bound to new instance
//   - Overload resolution: Uses semantic.ResolveOverload with type matching
//   - Virtual dispatch: NOT used for static calls (static binding)
//   - Field initialization: Field initializers evaluated, then default values for remaining fields
//   - Result: For constructors, always return object (not Result variable)
//   - Error: "wrong number of arguments", "There is no overloaded version that can be called with these arguments"
//   - Implementation: ~245 lines in original
//
// **3. RECORD TYPE STATIC METHOD CALLS** (TRecord.Method())
//   - Pattern: `TPoint.Create()`, `TRecord.ClassMethod()`
//   - Detection: Object is Identifier with `__record_type_` + lowercase(name) in environment
//   - Lookup: Check RecordTypeValue.ClassMethodOverloads (case-insensitive)
//   - Overload resolution: Same as class methods (resolveMethodOverload)
//   - Execution: Call callRecordStaticMethod WITHOUT Self binding
//   - Error: "static method 'X' not found in record type 'Y'"
//   - Implementation: ~30 lines in original
//
// **4. CLASSINFO VALUE METHOD CALLS** (ClassInfoValue.Method())
//   - Pattern: `Self.ClassMethod()` where Self is ClassInfoValue in class method
//   - Detection: Object evaluates to ClassInfoValue
//   - Lookup: Only class methods (getMethodOverloadsInHierarchy with isClassMethod=true)
//   - Execution: executeClassMethod with Self bound to ClassInfoValue
//   - Error: "class method 'X' not found in class 'Y'"
//   - Implementation: ~15 lines in original
//
// **5. METACLASS CONSTRUCTOR CALLS** (ClassValue.Create())
//   - Pattern: `var cls: class of TParent; cls := TChild; obj := cls.Create()`
//   - Detection: Object evaluates to ClassValue
//   - Purpose: Virtual constructor dispatch via metaclass variables
//   - Process:
//     a. Extract runtime class from ClassValue.ClassInfo
//     b. Look up constructor overloads in runtime class (getMethodOverloadsInHierarchy)
//     c. Resolve constructor overload based on arguments
//     d. Create new instance of runtime class (virtual dispatch - uses actual class, not declared type)
//     e. Initialize all fields with default values
//     f. Execute constructor with Self bound to new instance
//   - Key feature: Creates instance of RUNTIME type, not static type
//   - Error: "constructor 'X' not found in class 'Y'"
//   - Implementation: ~95 lines in original
//
// **6. SET VALUE BUILT-IN METHODS** (SetValue.Include/Exclude())
//   - Pattern: `mySet.Include(x)`, `mySet.Exclude(y)`
//   - Detection: Object evaluates to SetValue
//   - Supported methods (case-insensitive):
//     a. Include(element): Add element to set
//     b. Exclude(element): Remove element from set
//   - Error: "method 'X' not found for set type"
//   - Implementation: ~30 lines in original
//
// **7. RECORD INSTANCE METHOD CALLS** (RecordValue.Method())
//   - Pattern: `point.GetLength()`, `record.DoSomething(x)`
//   - Detection: Object evaluates to RecordValue
//   - Process: Convert to member access, delegate to evalRecordMethodCall
//   - Supports: Instance methods, properties, class methods (via instance)
//   - Implementation: ~10 lines in original (delegates to evalRecordMethodCall)
//
// **8. INTERFACE INSTANCE METHOD CALLS** (InterfaceInstance.Method())
//   - Pattern: `intfVar.Hello()`, `intf.DoSomething(x)`
//   - Detection: Object evaluates to InterfaceInstance
//   - Validation: Verify method exists in interface definition (Interface.HasMethod)
//   - Process: Unwrap to underlying object, continue with object method dispatch
//   - Error: "Interface is nil", "method 'X' not found in interface 'Y'"
//   - Implementation: ~15 lines in original
//
// **9. NIL OBJECT ERROR HANDLING**
//   - Pattern: `var o: TClass := nil; o.Method()`
//   - Detection: Object evaluates to NilValue
//   - Result: Always raise "Object not instantiated"
//   - Note: Class methods can only be called on class name, not nil instance
//   - Implementation: ~5 lines in original
//
// **10. ENUM TYPE META METHODS** (TypeMetaValue.Low/High/ByName())
//   - Pattern: `TColor.Low()`, `TColor.High()`, `TColor.ByName('Red')`
//   - Detection: Object evaluates to TypeMetaValue with EnumType
//   - Supported methods (case-insensitive):
//     a. Low(): Returns lowest ordinal value as Integer
//     b. High(): Returns highest ordinal value as Integer
//     c. ByName(name: string): Returns ordinal for enum value name (0 if not found)
//   - Supports qualified names (TypeName.ValueName)
//   - Case-insensitive lookup
//   - Returns 0 for empty string or not found (DWScript behavior)
//   - Implementation: ~50 lines in original
//
// **11. HELPER METHOD CALLS** (any_type.HelperMethod())
//   - Pattern: `"hello".ToUpper()`, `arr.Push(x)`, `123.ToString()`
//   - Detection: Object is not an object/record, but helpers provide this method
//   - Process:
//     a. Find helper method (findHelperMethod) - returns AST method or builtin spec
//     b. Evaluate all arguments
//     c. Call helper method (callHelperMethod)
//   - Supports: String, Array, Integer, Float helpers, etc.
//   - Error: "cannot call method 'X' on type 'Y' (no helper found)"
//   - Implementation: ~20 lines in original
//
// **12. OBJECT INSTANCE METHOD CALLS** (ObjectInstance.Method())
//   - Pattern: `obj.DoSomething()`, `obj.GetValue(x)`
//   - Detection: Object evaluates to ObjectInstance
//   - Built-in methods: `ClassName()` returns obj.Class.Name
//   - Process:
//     a. Collect instance method overloads (getMethodOverloadsInHierarchy with isClassMethod=false)
//     b. Collect class method overloads (isClassMethod=true) - can be called on instances
//     c. Resolve overload based on argument types (resolveMethodOverload)
//     d. Apply VIRTUAL DISPATCH for virtual/override methods (use VirtualMethodTable)
//   - Only for methods with IsVirtual or IsOverride (NOT reintroduce)
//   - Look up method signature in obj.Class.VirtualMethodTable
//   - Use most derived implementation if found
//     e. Check recursion depth (WillOverflow) before execution
//     f. Push method name to call stack for stack traces
//     g. Bind Self to object (or ClassInfoValue for class methods)
//     h. Add class constants to method scope
//     i. Bind parameters to arguments with implicit type conversion
//     j. Initialize Result variable (or method name alias)
//     k. Execute method body
//     l. Extract return value (Result or method name variable)
//     m. Apply implicit return type conversion
//   - Virtual constructor handling: If calling constructor on instance (o.Create),
//     create NEW instance of runtime type with virtual dispatch
//   - Error: "method 'X' not found in class 'Y'"
//   - Implementation: ~290 lines in original
//
// **13. VIRTUAL CONSTRUCTOR DISPATCH** (obj.Create())
//   - Pattern: `var o: TParent; o := TChild.Create; newObj := o.Create()`
//   - Detection: Resolved method is constructor (method.IsConstructor)
//   - Purpose: Create new instance of object's RUNTIME type (virtual dispatch)
//   - Process:
//     a. Find constructor in object's runtime class hierarchy (start from obj.Class)
//     b. Create NEW instance of runtime class (not existing object)
//     c. Initialize all fields with default values
//     d. Execute constructor with Self bound to new instance
//   - Key feature: Always creates NEW object, doesn't modify existing object
//   - Returns: New ObjectInstance (not the one the method was called on)
//   - Implementation: ~85 lines in original
//
// **14. CLASS METHOD EXECUTION** (executeClassMethod)
//   - Pattern: All class methods (called on class or instance)
//   - Self binding: Bound to ClassInfoValue (not instance)
//   - Environment: New environment with class constants, parameters, Result
//   - Call stack: Tracks method name for recursion detection and stack traces
//   - Recursion checking: Validates against max recursion limit
//   - Result handling: Checks Result and method name variables, implicit conversion
//   - Implementation: ~105 lines in original
//
// **15. OVERLOAD RESOLUTION** (resolveMethodOverload)
//   - Purpose: Select correct method from multiple overloads based on argument types
//   - Process:
//     a. Fast path: Single overload → return immediately
//     b. Evaluate all arguments to get runtime types (getValueType)
//     c. Convert method declarations to semantic.Symbol
//     d. Use semantic.ResolveOverload with type matching
//     e. Find method declaration corresponding to selected symbol
//   - Signature matching: Compares parameter types AND return type
//   - Inheritance: Child methods with same signature hide parent methods
//   - Error: "There is no overloaded version of 'X.Y' that can be called with these arguments"
//   - Implementation: ~50 lines in original + ~70 lines for getMethodOverloadsInHierarchy
//
// **SPECIAL BEHAVIORS**:
// - **Virtual dispatch**: Methods marked virtual/override use VirtualMethodTable for polymorphism
// - **Overload resolution**: Multiple methods with same name resolved by argument types
// - **Case-insensitive**: All name lookups are case-insensitive (DWScript spec)
// - **Inheritance**: Methods searched up class hierarchy, child signatures hide parent
// - **Recursion tracking**: Call stack monitored, max depth enforced
// - **Self binding**: Varies by context (object, ClassInfoValue, or nil)
// - **Result variable**: Functions initialize Result, can also use method name as alias
// - **Implicit conversion**: Parameters and return values converted to match declared types
// - **Helper support**: Types without native methods can have helper methods
// - **Constructor semantics**: Always return new object (not Result), initialize all fields
// - **Virtual constructors**: Constructor calls on instances use runtime type (virtual dispatch)
// - **Field initialization**: Field initializers evaluated in temporary environment with class constants
// - **Class constants**: Added to method scope for direct access without qualification
//
// **DEPENDENCIES** (blockers for full migration):
// - ObjectInstance, ClassInfo, ClassValue, ClassInfoValue - in internal/interp (needs migration)
// - RecordValue, RecordTypeValue - in internal/interp (needs migration)
// - InterfaceInstance, Interface - in internal/interp (needs migration)
// - SetValue - in internal/interp (needs migration)
// - TypeMetaValue, EnumType - in internal/interp (needs migration)
// - ReferenceValue - in internal/interp (needs migration)
// - VirtualMethodTable infrastructure - in internal/interp (needs migration)
// - Helper infrastructure - findHelperMethod, callHelperMethod (needs adapter)
// - Overload resolution - resolveMethodOverload, getMethodOverloadsInHierarchy (needs adapter)
// - User function calls - callUserFunction, callRecordStaticMethod (needs adapter)
// - Type system - getValueType, extractFunctionType, resolveTypeFromAnnotation (needs adapter)
// - Unit registry - ResolveQualifiedFunction (already in Evaluator)
// - Call stack - pushCallStack, popCallStack, WillOverflow (needs adapter)
// - Environment management - NewEnclosedEnvironment, bindClassConstantsToEnv (needs adapter)
//
// **TESTING**:
// - Unit-qualified function calls (Math.Sin, System.WriteLine)
// - Static class method calls (TClass.ClassMethod, TClass.Create)
// - Implicit parameterless constructor (TClass.Create with no constructor defined)
// - Constructor overloading (TClass.Create(), TClass.Create(x))
// - Record static method calls (TRecord.Create, TRecord.Count)
// - ClassInfoValue method calls (Self.ClassMethod in class context)
// - Metaclass constructor calls (cls.Create where cls is class of T)
// - Set methods (mySet.Include(x), mySet.Exclude(y))
// - Record instance methods (point.GetLength, record.DoSomething)
// - Interface method calls (intf.Hello, intf.Process)
// - Nil object error (nil.Method → error)
// - Enum meta methods (TColor.Low(), TColor.ByName('Red'))
// - Helper methods (str.ToUpper, arr.Push, num.ToString)
// - Object instance methods (obj.GetValue, obj.DoSomething)
// - Virtual dispatch (parent ref to child object calls child's override)
// - Virtual constructor dispatch (obj.Create creates new instance of runtime type)
// - Method overloading (multiple methods with same name, different parameters)
// - Overload inheritance (child methods hide parent methods with same signature)
// - Recursion limit enforcement (stack overflow detection)
// - Self binding variations (object, ClassInfoValue, record)
// - Result variable handling (Result, method name alias, implicit conversion)
// - Class constant access in methods (direct access without qualification)
// - Field initialization (initializers, default values, constructor execution)
//
// **IMPLEMENTATION SUMMARY**:
// - Original implementation: 1,116 lines (objects_methods.go:12-1116)
// - Handles 15 distinct method call modes with complex dispatch logic
// - Supports virtual dispatch, overload resolution, recursion tracking
// - Requires extensive OOP infrastructure not yet in runtime package
// - Full migration deferred - will be broken into category-specific sub-tasks
//
// **MIGRATION STRATEGY**:
// - Phase 1 (this task): Comprehensive documentation of all call modes
// - Phase 2 (future): Migrate simple cases (built-in methods, direct calls)
// - Phase 3 (future): Migrate overload resolution after type system migration
// - Phase 4 (future): Migrate virtual dispatch after VMT migration
// - Phase 5 (future): Migrate constructor dispatch after object creation migration
// - Phase 6 (future): Migrate helper methods after helper system migration
func (e *Evaluator) VisitMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	if node.Object == nil {
		return e.newError(node, "method call missing object")
	}
	if node.Method == nil {
		return e.newError(node, "method call missing method")
	}

	// Task 3.5.27 & 3.5.28: Implement method call routing based on object type

	// Evaluate the object first
	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}

	methodName := node.Method.Value

	// Evaluate all arguments
	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[i] = val
	}

	// Route based on object type
	switch obj.Type() {
	case "OBJECT":
		// Task 3.5.28: Object instance method calls (Mode 12)
		// Pattern: obj.Method(args)
		// Delegate to adapter for method lookup, virtual dispatch, and execution
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "INTERFACE":
		// Task 3.5.27: Interface instance method calls (Mode 8)
		// Pattern: intf.Method(args)
		// Delegate to adapter for interface unwrapping and method dispatch
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "CLASS", "CLASS_INFO":
		// Task 3.5.27: Static class method calls (Mode 2) or ClassInfoValue method calls (Mode 4)
		// Pattern: TClass.Method(args) or classInfoVal.Method(args)
		// Delegate to adapter for class method execution
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "CLASS_VALUE":
		// Task 3.5.28: Metaclass constructor calls (Mode 5)
		// Pattern: classVar.Create(args) where classVar is a metaclass variable
		// Delegate to adapter for virtual constructor dispatch
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "RECORD":
		// Task 3.5.27: Record instance method calls (Mode 7)
		// Pattern: record.Method(args)
		// Delegate to adapter for record method execution
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "SET":
		// Task 3.5.27: Set built-in methods (Mode 6)
		// Pattern: mySet.Include(x), mySet.Exclude(y)
		// Delegate to adapter for set method execution
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "NIL":
		// Task 3.5.28: Nil object error handling (Mode 9)
		// Always raise "Object not instantiated" error
		return e.newError(node, "Object not instantiated")

	case "TYPE_META":
		// Task 3.5.27: Enum type meta methods (Mode 10)
		// Pattern: TColor.Low(), TColor.High(), TColor.ByName('Red')
		// Delegate to adapter for enum meta method execution
		return e.adapter.CallMethod(obj, methodName, args, node)

	// Task 3.5.98a: Explicit cases for helper-enabled types
	// These types use helper methods (type extensions) for their method calls.
	// Examples: str.ToUpper(), arr.Push(), num.ToString()
	// Currently delegated to adapter which handles helper lookup and invocation.
	case "STRING":
		// String helpers: ToUpper, ToLower, Split, Trim, Contains, etc.
		return e.adapter.EvalNode(node)

	case "INTEGER":
		// Integer helpers: ToString, ToHexString, etc.
		return e.adapter.EvalNode(node)

	case "FLOAT":
		// Float helpers: ToString, etc.
		return e.adapter.EvalNode(node)

	case "BOOLEAN":
		// Boolean helpers: ToString, etc.
		return e.adapter.EvalNode(node)

	case "ARRAY":
		// Array helpers: Push, Pop, Add, Delete, IndexOf, SetLength, Swap, etc.
		return e.adapter.EvalNode(node)

	case "VARIANT":
		// Variant helpers (if any)
		return e.adapter.EvalNode(node)

	case "ENUM":
		// Enum value helpers (distinct from TYPE_META which is enum type methods)
		return e.adapter.EvalNode(node)

	default:
		// For other types (identifiers that might be unit names, record types, etc.)
		// Delegate to adapter for full handling
		return e.adapter.EvalNode(node)
	}
}

// VisitInheritedExpression evaluates an 'inherited' expression.
//
// **COMPLEXITY**: High (~176 lines in original implementation)
// **STATUS**: Partial migration with context validation, method name resolution, and argument evaluation in evaluator; inherited method execution delegated to adapter
//
// **SYNTAX FORMS**:
//   - `inherited MethodName(args)` - Explicit method call with arguments
//   - `inherited MethodName` - Explicit method/property/field access without args
//   - `inherited` - Bare inherited (calls same method in parent class)
//
// **EXECUTION PHASES** (evaluated in this order):
//
// **1. CONTEXT VALIDATION** (~10 lines)
//   - **Self check**: Must be in method context with Self defined
//   - Error: "inherited can only be used inside a method"
//   - **ObjectInstance check**: Self must be an ObjectInstance
//   - Error: "inherited requires Self to be an object instance"
//   - **Parent class check**: Current class must have a parent
//   - Error: "class 'X' has no parent class"
//   - Implementation: lines 794-811 in original
//
// **2. METHOD NAME RESOLUTION** (~17 lines)
//   - **Explicit method name**: `inherited MethodName(...)`
//   - Uses ie.Method.Value directly
//   - **Bare inherited**: `inherited` (no method name specified)
//   - Looks up `__CurrentMethod__` from environment
//   - Must be StringValue containing current method name
//   - Enables nested inherited calls through inheritance chain
//   - Error: "bare 'inherited' requires method context"
//   - Error: "invalid method context"
//   - Implementation: lines 813-829 in original
//
// **3. MEMBER LOOKUP IN PARENT CLASS** (case-insensitive)
//
//	Searches parent class members in priority order:
//
//	**3a. METHODS** (~87 lines)
//	  - Iterates parentClass.Methods map with case-insensitive comparison
//	  - If found, executes full method call (see Phase 4)
//	  - Implementation: lines 831-927 in original
//
//	**3b. PROPERTIES** (~17 lines)
//	  - Iterates parentClass.Properties map with case-insensitive comparison
//	  - Cannot be called with arguments or as method
//	    - Error: "cannot call property 'X' as a method"
//	  - Reads property via evalPropertyRead()
//	  - Implementation: lines 929-946 in original
//
//	**3c. FIELDS** (~13 lines)
//	  - Iterates parentClass.Fields map with case-insensitive comparison
//	  - Cannot be called with arguments or as method
//	    - Error: "cannot call field 'X' as a method"
//	  - Returns field value directly via obj.GetField()
//	  - Returns NilValue if field is nil
//	  - Implementation: lines 948-961 in original
//
//	**3d. NOT FOUND**
//	  - Error: "method, property, or field 'X' not found in parent class 'Y'"
//
// **4. METHOD EXECUTION** (when method found)
//   - **Argument evaluation**: Evaluates all arguments in current environment
//   - **Argument count validation**:
//   - Error: "wrong number of arguments for method 'X': expected N, got M"
//   - **Method environment setup**:
//     a. Creates enclosed environment via NewEnclosedEnvironment
//     b. Binds `Self` to **current object** (preserves instance identity!)
//     c. Binds `__CurrentClass__` to parent ClassInfoValue
//     d. Adds parent class constants via bindClassConstantsToEnv()
//     e. Binds `__CurrentMethod__` to method name (enables nested inherited)
//   - **Parameter binding**:
//   - Binds each parameter to corresponding argument
//   - Applies implicit type conversion if parameter has type annotation
//   - Uses tryImplicitConversion() helper
//   - **Result variable initialization** (for functions):
//   - Resolves return type via resolveTypeFromAnnotation()
//   - Gets default value via getDefaultValue()
//   - Binds `Result` to default value
//   - Binds method name as ReferenceValue alias to Result (DWScript style)
//   - **Body execution**: Executes parentMethod.Body via Eval()
//   - **Return value extraction**:
//   - For functions: checks Result, then method name alias, then NilValue
//   - For procedures: returns NilValue
//   - **Environment restoration**: Restores saved environment
//   - Implementation: lines 841-927 in original
//
// **SPECIAL BEHAVIORS**:
//
// **Self Preservation**:
//   - Critical feature of inherited calls
//   - Parent method executes with **current instance** as Self
//   - NOT a new parent instance - same object through inheritance chain
//   - Enables: Child.Method() → inherited → Parent.Method() on same object
//   - Example: overridden method in Child calls inherited, parent code
//     operates on the Child instance's fields
//
// **Bare Inherited Support**:
//   - `inherited` without method name calls same method in parent
//   - Uses `__CurrentMethod__` environment variable set during method entry
//   - Enables clean inheritance patterns without repeating method name
//   - Supports nested inheritance: GrandChild → Child → Parent all using bare inherited
//
// **Case-Insensitive Lookup**:
//   - DWScript standard: all identifiers are case-insensitive
//   - Method, property, and field lookups use strings.EqualFold()
//
// **Method Name as Return Alias**:
//   - DWScript allows `MethodName := value` as alternative to `Result := value`
//   - Implemented via ReferenceValue pointing to Result variable
//   - Both forms work interchangeably in inherited method context
//
// **Class Constants in Scope**:
//   - Parent class constants are bound to method scope
//   - Allows inherited method to access parent's constants directly
//   - Uses bindClassConstantsToEnv() helper
//
// **Implicit Type Conversion**:
//   - Arguments are converted to parameter types if possible
//   - Uses tryImplicitConversion() for Integer↔Float, etc.
//   - Applied during parameter binding, not argument evaluation
//
// **DEPENDENCIES** (blockers for full migration):
//   - ObjectInstance: Current object reference with Class pointer
//   - ClassInfo: Class metadata with Parent, Methods, Properties, Fields
//   - ClassInfoValue: Wrapper for __CurrentClass__ binding
//   - StringValue: For __CurrentMethod__ storage
//   - ReferenceValue: For method name alias to Result
//   - NilValue: For procedure returns and nil field values
//   - PropertyInfo: For property lookup and evalPropertyRead
//   - Environment: Scope management for method execution
//   - NewEnclosedEnvironment(): Scope creation
//   - bindClassConstantsToEnv(): Constant binding helper
//   - tryImplicitConversion(): Type conversion helper
//   - resolveTypeFromAnnotation(): Return type resolution
//   - getDefaultValue(): Default value for return type
//   - evalPropertyRead(): Property read evaluation
//
// **MIGRATION STRATEGY**:
//   - Phase 1 (this task): Comprehensive documentation of all modes ✓
//   - Phase 2 (future): Migrate context validation after ObjectInstance migration
//   - Phase 3 (future): Migrate method lookup after ClassInfo migration
//   - Phase 4 (future): Migrate method execution after method call migration
//   - Phase 5 (future): Migrate property/field access after property system migration
//
// **ERROR CONDITIONS**:
//   - "inherited can only be used inside a method" - No Self in environment
//   - "inherited requires Self to be an object instance" - Self not ObjectInstance
//   - "class 'X' has no parent class" - Root class with no parent
//   - "bare 'inherited' requires method context" - No __CurrentMethod__ for bare inherited
//   - "invalid method context" - __CurrentMethod__ not a StringValue
//   - "wrong number of arguments for method 'X'" - Argument/parameter count mismatch
//   - "cannot call property 'X' as a method" - Property access with args/call syntax
//   - "cannot call field 'X' as a method" - Field access with args/call syntax
//   - "method, property, or field 'X' not found in parent class 'Y'" - Member not found
func (e *Evaluator) VisitInheritedExpression(node *ast.InheritedExpression, ctx *ExecutionContext) Value {
	// Get Self from environment - must be in a method context
	selfVal, exists := ctx.Env().Get("Self")
	if !exists {
		return e.newError(node, "inherited can only be used inside a method")
	}

	// Convert to Value type
	self, ok := selfVal.(Value)
	if !ok {
		return e.newError(node, "inherited requires Self to be an object instance")
	}

	// Determine the method name
	var methodName string
	if node.Method != nil {
		// Explicit method name: inherited MethodName(args)
		methodName = node.Method.Value
	} else {
		// Bare inherited: get current method name from environment
		currentMethodVal, exists := ctx.Env().Get("__CurrentMethod__")
		if !exists {
			return e.newError(node, "bare 'inherited' requires method context")
		}

		// Extract method name string - check for runtime.StringValue
		// Note: internal/interp.StringValue is a type alias for runtime.StringValue,
		// so this check handles both cases.
		if strVal, ok := currentMethodVal.(*runtime.StringValue); ok {
			methodName = strVal.Value
		} else {
			return e.newError(node, "invalid method context")
		}
	}

	// Evaluate all arguments
	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[i] = val
	}

	// Call the inherited method using the adapter
	// The adapter handles: parent class lookup, method resolution, environment setup,
	// Self binding, parameter binding, and method execution
	return e.adapter.CallInheritedMethod(self, methodName, args)
}

// VisitSelfExpression evaluates a 'Self' expression.
// Self refers to the current instance (in instance methods) or the current class (in class methods).
func (e *Evaluator) VisitSelfExpression(node *ast.SelfExpression, ctx *ExecutionContext) Value {
	selfVal, exists := ctx.Env().Get("Self")
	if !exists {
		return e.newError(node, "Self used outside method context")
	}

	val, ok := selfVal.(Value)
	if !ok {
		return e.newError(node, "Self has invalid type")
	}

	return val
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	if value, ok := val.(Value); ok {
		return value
	}

	return e.newError(node, "enum value '%s' has invalid type", valueName)
}

// VisitRecordLiteralExpression evaluates a record literal expression.
// Handles typed and anonymous record literals with field initialization and default values.
// Task 3.5.40: Full migration from Interpreter.evalRecordLiteral()
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	// Task 3.5.40: Record literal evaluation with field initialization
	//
	// Record literal syntax:
	// - Typed: TMyRecord(Field1: 1, Field2: 'hello')
	// - Anonymous: (Field1: 1, Field2: 'hello') - requires type context
	//
	// Type determination:
	// - Explicit type name in literal → use that type
	// - Anonymous literal → requires type context (set by variable/const declaration)
	//
	// Field initialization:
	// - Evaluate field value expressions
	// - Validate fields exist in record type
	// - Initialize missing fields with field initializers or zero values
	// - Handle nested records recursively
	// - Handle interface-typed fields specially
	//
	// Positional field initialization:
	// - Not yet supported (requires field.Name == nil)
	// - Returns error if attempted
	//
	// Field validation:
	// - All specified fields must exist in record type
	// - Field names are case-insensitive
	// - Duplicate field names not allowed (parser enforces)
	//
	// Default value initialization:
	// - Missing fields get field initializers if defined
	// - Otherwise get zero values for their types
	// - Nested records initialized recursively
	// - Interface fields get nil InterfaceInstance

	if node == nil {
		return e.newError(node, "nil record literal")
	}

	// Determine record type
	var recordTypeName string
	switch {
	case node.TypeName != nil:
		recordTypeName = node.TypeName.Value
	case ctx.RecordTypeContext() != "":
		// Anonymous literal with type context from caller (e.g., var/const declaration)
		recordTypeName = ctx.RecordTypeContext()
	default:
		// Anonymous literal requires type context (should have been set by caller)
		return e.newError(node, "record literal requires explicit type name or type context")
	}

	// Look up record type
	// Task 3.5.46: Use TypeSystem directly instead of adapter
	if !e.typeSystem.HasRecord(recordTypeName) {
		return e.newError(node, "unknown record type '%s'", recordTypeName)
	}

	// Evaluate field values
	fieldValues := make(map[string]Value)
	for _, field := range node.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			return e.newError(node, "positional record field initialization not yet supported")
		}

		fieldName := field.Name.Value

		// Evaluate the field value expression
		fieldValue := e.Eval(field.Value, ctx)
		if isError(fieldValue) {
			return fieldValue
		}

		// Store the field value (case-insensitive)
		fieldValues[fieldName] = fieldValue
	}

	// Create record value with field initialization
	recordValue, err := e.adapter.CreateRecordValue(recordTypeName, fieldValues)
	if err != nil {
		return e.newError(node, "%v", err)
	}

	return recordValue
}

// VisitSetLiteral evaluates a set literal [value1, value2, ...].
// Handles simple elements, ranges, and mixed sets with proper type inference.
// Task 3.5.80: Direct evaluation without adapter EvalNode call.
func (e *Evaluator) VisitSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	return e.evalSetLiteralDirect(node, ctx)
}

// VisitArrayLiteralExpression evaluates an array literal [1, 2, 3].
// Task 3.5.83: Direct evaluation without adapter EvalNode call.
// Handles type inference, element coercion, and bounds validation for static and dynamic arrays.
func (e *Evaluator) VisitArrayLiteralExpression(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	return e.evalArrayLiteralDirect(node, ctx)
}

// VisitIndexExpression evaluates an index expression array[index].
// Handles array, string, property, and JSON indexing with bounds checking.
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil index expression")
	}

	if node.Left == nil {
		return e.newError(node, "index expression missing base")
	}

	// Check if this might be a multi-index property access (obj.Property[x, y])
	// We only flatten indices if the base is a MemberAccessExpression (property access)
	// For regular array access like arr[i][j], we process each level separately
	base, _ := CollectIndices(node)

	// Check if this is indexed property access: obj.Property[index1, index2, ...]
	// Only flatten indices for property access, not for regular arrays
	if _, ok := base.(*ast.MemberAccessExpression); ok {
		// Delegate property-based indexing to adapter (complex logic with class/interface lookups)
		return e.adapter.EvalNode(node)
	}

	// Not a property access - this is regular array/string indexing
	// Process ONLY the outermost index, not all nested indices
	// This allows FData[x][y] to work as: (FData[x])[y]
	leftVal := e.Eval(node.Left, ctx)
	if isError(leftVal) {
		return leftVal
	}

	if node.Index == nil {
		return e.newError(node, "index expression missing index")
	}

	// Evaluate the index for this level only
	indexVal := e.Eval(node.Index, ctx)
	if isError(indexVal) {
		return indexVal
	}

	// Unwrap variants for indexing
	leftVal = unwrapVariant(leftVal)

	// Handle interface/object default property access - delegate to adapter
	// These require class hierarchy lookup and property getter execution
	switch leftVal.Type() {
	case "INTERFACE", "OBJECT", "RECORD":
		// Default property access requires adapter (class lookup, method calls)
		return e.adapter.EvalNode(node)
	case "JSON":
		// JSON indexing requires VariantValue which is in interp package
		return e.adapter.EvalNode(node)
	}

	// Index must be an integer or enum for arrays and strings
	index, ok := ExtractIntegerIndex(indexVal)
	if !ok {
		return e.newError(node, "index must be an integer or enum, got %s", indexVal.Type())
	}

	// Check if left side is an array
	if arrayVal, ok := leftVal.(*runtime.ArrayValue); ok {
		return e.IndexArray(arrayVal, index, node)
	}

	// Check if left side is a string
	if strVal, ok := leftVal.(*runtime.StringValue); ok {
		return e.IndexString(strVal, index, node)
	}

	return e.newError(node, "cannot index type %s", leftVal.Type())
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
	elementType, typeErr := e.ResolveType(elementTypeName)
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
	return e.adapter.CreateLambda(node, ctx.Env())
}

// VisitIsExpression evaluates an 'is' type checking expression.
// Task 3.5.34: Runtime type checking with class hierarchy and interface support.
//
// The 'is' operator has two modes:
// 1. Type checking: `obj is TMyClass` or `obj is IMyInterface`
//   - Returns true if obj is an instance of the class (or derived class)
//   - Returns true if obj's class implements the interface
//   - Returns false for nil objects
//
// 2. Boolean value comparison: `boolExpr is True` or `boolExpr is False`
//   - This variant uses Right expression instead of TargetType
//   - Compares two values as booleans using variant-to-bool coercion
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	// Check if this is a boolean value comparison (expr.Right is set)
	// or a type check (expr.TargetType is set)
	if node.Right != nil {
		// Boolean value comparison: left is right
		// This is essentially checking if left == right for boolean values
		left := e.Eval(node.Left, ctx)
		if isError(left) {
			return left
		}

		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}

		// Convert both to boolean values using VariantToBool
		leftBool := VariantToBool(left)
		rightBool := VariantToBool(right)

		return &runtime.BooleanValue{Value: leftBool == rightBool}
	}

	// Type checking mode
	// Evaluate the left expression (the object to check)
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}

	// Handle nil - nil is not an instance of any type
	if left == nil || left.Type() == "NIL" {
		return &runtime.BooleanValue{Value: false}
	}

	// Get the target type name from the type expression
	targetTypeName := ""
	if typeAnnotation, ok := node.TargetType.(*ast.TypeAnnotation); ok {
		targetTypeName = typeAnnotation.Name
	} else {
		return e.newError(node, "cannot determine target type")
	}

	// Use the adapter's CheckType method which handles:
	// - Class hierarchy traversal (obj is TMyClass checks class and parent classes)
	// - Interface implementation checking (obj is IMyInterface checks if class implements it)
	result := e.adapter.CheckType(left, targetTypeName)
	return &runtime.BooleanValue{Value: result}
}

// VisitAsExpression evaluates an 'as' type casting expression.
// Task 3.5.35: Runtime type casting with interface wrapping/unwrapping.
//
// The 'as' operator performs type casting with the following behaviors:
// 1. nil → any type: returns nil (nil can be cast to any class or interface)
// 2. interface → class: extracts underlying object (validates class hierarchy)
// 3. interface → interface: creates new interface wrapper (validates implementation)
// 4. object → class: validates class hierarchy (returns same object if valid)
// 5. object → interface: creates interface wrapper (validates implementation)
//
// If the cast is invalid, returns an error that should be raised as an exception.
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	// Evaluate the left expression (the object to cast)
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}

	// Get the target type name from the type expression
	targetTypeName := ""
	if typeAnnotation, ok := node.TargetType.(*ast.TypeAnnotation); ok {
		targetTypeName = typeAnnotation.Name
	} else {
		return e.newError(node, "cannot determine target type")
	}

	// Use the adapter's CastType method which handles:
	// - nil casting
	// - interface-to-class casting (unwrapping)
	// - interface-to-interface casting (re-wrapping)
	// - object-to-class casting (hierarchy validation)
	// - object-to-interface casting (wrapping)
	result, err := e.adapter.CastType(left, targetTypeName)
	if err != nil {
		// Cast failed - return error that should be raised as exception
		// The error message from CastType includes the specific failure reason
		return e.newError(node, "%s", err.Error())
	}

	return result
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
// Task 3.5.36: Interface implementation verification.
//
// The 'implements' operator checks if a class/object implements an interface:
// - obj implements IInterface -> Boolean (object instance)
// - TClass implements IInterface -> Boolean (class type reference)
// - ClassVar implements IInterface -> Boolean (metaclass variable)
//
// Returns true if the class implements the interface, false otherwise.
// Unlike 'is' which checks class hierarchy too, 'implements' only checks interfaces.
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	// Evaluate the left expression (the object or class to check)
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}

	// Get the target interface name from the type expression
	targetInterfaceName := ""
	if typeAnnotation, ok := node.TargetType.(*ast.TypeAnnotation); ok {
		targetInterfaceName = typeAnnotation.Name
	} else {
		return e.newError(node, "cannot determine target interface type")
	}

	// Use the adapter's CheckImplements method which handles:
	// - nil objects (return false)
	// - ObjectInstance (extract class)
	// - ClassValue (metaclass variable)
	// - ClassInfoValue (class type identifier)
	result, err := e.adapter.CheckImplements(left, targetInterfaceName)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	return &runtime.BooleanValue{Value: result}
}

// VisitIfExpression evaluates an inline if-then-else expression.
func (e *Evaluator) VisitIfExpression(node *ast.IfExpression, ctx *ExecutionContext) Value {
	condition := e.Eval(node.Condition, ctx)
	if isError(condition) {
		return condition
	}

	// Evaluate consequence if condition is truthy
	if IsTruthy(condition) {
		result := e.Eval(node.Consequence, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// Evaluate alternative if present
	if node.Alternative != nil {
		result := e.Eval(node.Alternative, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// No else clause - return default value for the consequence type
	var typeAnnot *ast.TypeAnnotation
	if e.semanticInfo != nil {
		typeAnnot = e.semanticInfo.GetType(node)
	}
	if typeAnnot == nil {
		return e.newError(node, "if expression missing type annotation")
	}

	// Return default value based on type name
	typeName := ident.Normalize(typeAnnot.Name)
	switch typeName {
	case "integer", "int64":
		return &runtime.IntegerValue{Value: 0}
	case "float", "float64", "double", "real":
		return &runtime.FloatValue{Value: 0.0}
	case "string":
		return &runtime.StringValue{Value: ""}
	case "boolean", "bool":
		return &runtime.BooleanValue{Value: false}
	default:
		return &runtime.NilValue{}
	}
}

// VisitOldExpression evaluates an 'old' expression (used in postconditions).
func (e *Evaluator) VisitOldExpression(node *ast.OldExpression, ctx *ExecutionContext) Value {
	identName := node.Identifier.Value
	oldValue, found := ctx.GetOldValue(identName)
	if !found {
		return e.newError(node, "old value for '%s' not captured (internal error)", identName)
	}
	return oldValue.(Value)
}

// VisitRangeExpression handles range expressions (start..end).
// Range expressions are only valid in specific contexts:
// - Case statement branches (case x of 1..10: ...) - handled in VisitCaseStatement
// - Set literals ([1..10]) - handled in set.go
// Direct evaluation of a standalone range expression is not valid in DWScript.
func (e *Evaluator) VisitRangeExpression(node *ast.RangeExpression, ctx *ExecutionContext) Value {
	if node.Start == nil || node.RangeEnd == nil {
		return e.newError(node, "range expression missing start or end")
	}

	// Range expressions are structural - they don't evaluate to a value on their own.
	// Direct evaluation is an error (contexts that use ranges handle them specially).
	return e.newError(node, "range expression cannot be evaluated directly; only valid in case statements or set literals")
}
