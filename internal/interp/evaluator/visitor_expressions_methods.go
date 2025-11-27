package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains visitor methods for method call and inherited expression AST nodes.
// These handle obj.Method(args) style calls and inherited keyword for parent class access.

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
//
// **COMPLEXITY**: Very High (1,116 lines in original implementation)
// **STATUS**: Task 3.5.115 - Consolidated to use DispatchMethodCall
//
// **15 DISTINCT METHOD CALL MODES** (evaluated in this order):
//
// **1. UNIT-QUALIFIED FUNCTION CALLS** (UnitName.FunctionName())
// **2. STATIC CLASS METHOD CALLS** (TClass.Method())
// **3. RECORD TYPE STATIC METHOD CALLS** (TRecord.Method())
// **4. CLASSINFO VALUE METHOD CALLS** (ClassInfoValue.Method())
// **5. METACLASS CONSTRUCTOR CALLS** (ClassValue.Create())
// **6. SET VALUE BUILT-IN METHODS** (SetValue.Include/Exclude())
// **7. RECORD INSTANCE METHOD CALLS** (RecordValue.Method())
// **8. INTERFACE INSTANCE METHOD CALLS** (InterfaceInstance.Method())
// **9. NIL OBJECT ERROR HANDLING**
// **10. ENUM TYPE META METHODS** (TypeMetaValue.Low/High/ByName())
// **11. HELPER METHOD CALLS** (any_type.HelperMethod())
// **12. OBJECT INSTANCE METHOD CALLS** (ObjectInstance.Method())
// **13. VIRTUAL CONSTRUCTOR DISPATCH** (obj.Create())
// **14. CLASS METHOD EXECUTION** (executeClassMethod)
// **15. OVERLOAD RESOLUTION** (resolveMethodOverload)
//
// See method_dispatch.go for comprehensive documentation of dispatch architecture.
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

	// Task 3.5.115: Consolidated method dispatch via DispatchMethodCall
	// This provides unified error handling and consistent routing for all value types.
	// See method_dispatch.go for full documentation of the dispatch architecture.
	return e.DispatchMethodCall(obj, methodName, args, node, ctx)
}

// VisitInheritedExpression evaluates an 'inherited' expression.
//
// **COMPLEXITY**: High (~176 lines in original implementation)
// **STATUS**: Task 3.5.114 - Migrated to use ObjectValue.CallInheritedMethod interface
//
// **SYNTAX FORMS**:
//   - `inherited MethodName(args)` - Explicit method call with arguments
//   - `inherited MethodName` - Explicit method/property/field access without args
//   - `inherited` - Bare inherited (calls same method in parent class)
//
// See comprehensive documentation in visitor_expressions.go for full details.
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

	// Task 3.5.114: Use ObjectValue interface for parent class method lookup
	// Then delegate method execution to adapter.ExecuteMethodWithSelf
	objVal, ok := self.(ObjectValue)
	if !ok {
		// Fallback to adapter for non-ObjectValue types
		return e.adapter.CallInheritedMethod(self, methodName, args)
	}

	// Create method executor callback that delegates to adapter
	methodExecutor := func(methodDecl any, methodArgs []Value) Value {
		return e.adapter.ExecuteMethodWithSelf(self, methodDecl, methodArgs)
	}

	// Call inherited method via ObjectValue interface
	// This performs parent class lookup directly on the object
	return objVal.CallInheritedMethod(methodName, args, methodExecutor)
}
