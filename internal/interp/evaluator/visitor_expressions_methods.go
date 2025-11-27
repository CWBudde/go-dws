package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for method call and inherited expression AST nodes.
// These handle obj.Method(args) style calls and inherited keyword for parent class access.

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
//
// **COMPLEXITY**: Very High (1,116 lines in original implementation)
// **STATUS**: Documentation-only migration with full adapter delegation
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
// See comprehensive documentation in visitor_expressions.go for full mode details.
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
		// Task 3.5.111d: Object instance method calls (Mode 12)
		// Pattern: obj.Method(args)
		// Delegate to adapter.CallMethod which handles ObjectInstance directly
		// with method lookup and Self binding
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "INTERFACE":
		// Task 3.5.27: Interface instance method calls (Mode 8)
		// Pattern: intf.Method(args)
		// Delegate to adapter for interface unwrapping and method dispatch
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "CLASSINFO":
		// Task 3.5.111c: Static class method calls (Mode 2) or ClassInfoValue method calls (Mode 4)
		// Pattern: TClass.Method(args) or classInfoVal.Method(args)
		// Delegate to adapter.CallMethod which now handles ClassInfoValue directly
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "CLASS":
		// Task 3.5.111e: Metaclass (ClassValue) method calls (Mode 5)
		// Pattern: classVar.Create(args) where classVar is a "class of TBase" metaclass variable
		// Delegate to adapter.CallMethod which needs to handle ClassValue
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "RECORD":
		// Task 3.5.27: Record instance method calls (Mode 7)
		// Pattern: record.Method(args)
		// Delegate to adapter for record method execution
		return e.adapter.CallMethod(obj, methodName, args, node)

	case "SET":
		// Task 3.5.111a: Set built-in methods (Mode 6) - Direct dispatch
		// Pattern: mySet.Include(x), mySet.Exclude(y)
		// Directly dispatch to SetMethodDispatcher interface methods
		setVal, ok := obj.(SetMethodDispatcher)
		if !ok {
			return e.newError(node, "internal error: SET value does not implement SetMethodDispatcher")
		}

		normalizedMethod := ident.Normalize(methodName)
		switch normalizedMethod {
		case "include":
			if len(args) != 1 {
				return e.newError(node, "Include expects 1 argument, got %d", len(args))
			}
			ordinal, err := GetOrdinalValue(args[0])
			if err != nil {
				return e.newError(node, "Include requires ordinal value: %s", err.Error())
			}
			setVal.AddElement(ordinal)
			return e.nilValue()

		case "exclude":
			if len(args) != 1 {
				return e.newError(node, "Exclude expects 1 argument, got %d", len(args))
			}
			ordinal, err := GetOrdinalValue(args[0])
			if err != nil {
				return e.newError(node, "Exclude requires ordinal value: %s", err.Error())
			}
			setVal.RemoveElement(ordinal)
			return e.nilValue()

		default:
			return e.newError(node, "method '%s' not found for set type", methodName)
		}

	case "NIL":
		// Task 3.5.28: Nil object error handling (Mode 9)
		// Always raise "Object not instantiated" error
		return e.newError(node, "Object not instantiated")

	case "TYPE_META":
		// Task 3.5.111b: Enum type meta methods (Mode 10) - Direct dispatch
		// Pattern: TColor.Low(), TColor.High(), TColor.ByName('Red')
		// Directly dispatch to EnumTypeMetaDispatcher interface methods
		enumMeta, ok := obj.(EnumTypeMetaDispatcher)
		if !ok {
			return e.newError(node, "internal error: TYPE_META value does not implement EnumTypeMetaDispatcher")
		}

		// Only enum types have these methods
		if !enumMeta.IsEnumTypeMeta() {
			return e.newError(node, "method '%s' not found for type '%s'", methodName, obj.String())
		}

		normalizedMethod := ident.Normalize(methodName)
		switch normalizedMethod {
		case "low":
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumLow())}

		case "high":
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumHigh())}

		case "byname":
			if len(args) != 1 {
				return e.newError(node, "ByName expects 1 argument, got %d", len(args))
			}
			nameStr, ok := args[0].(*runtime.StringValue)
			if !ok {
				return e.newError(node, "ByName expects string argument, got %s", args[0].Type())
			}
			return &runtime.IntegerValue{Value: int64(enumMeta.EnumByName(nameStr.Value))}

		default:
			return e.newError(node, "method '%s' not found for enum type", methodName)
		}

	// Task 3.5.98a+c: Explicit cases for helper-enabled types
	// These types use helper methods (type extensions) for their method calls.
	// Examples: str.ToUpper(), arr.Push(), num.ToString()
	//
	// Task 3.5.98c: Now using direct helper method lookup and execution instead of adapter delegation.
	case "STRING", "INTEGER", "FLOAT", "BOOLEAN", "ARRAY", "VARIANT", "ENUM":
		// Find the helper method for this value type
		helperResult := e.FindHelperMethod(obj, methodName)
		if helperResult == nil {
			return e.newError(node, "cannot call method '%s' on type '%s' (no helper found)", methodName, obj.Type())
		}

		// Execute the helper method (builtin or AST)
		return e.CallHelperMethod(helperResult, obj, args, node, ctx)

	default:
		// For other types (identifiers that might be unit names, record types, etc.)
		// Try helper method lookup first
		helperResult := e.FindHelperMethod(obj, methodName)
		if helperResult != nil {
			// Found a helper method - execute it
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}

		// No helper found - delegate to adapter for full handling
		// (might be unit-qualified call, record method, or other complex case)
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

	// Call the inherited method using the adapter
	// The adapter handles: parent class lookup, method resolution, environment setup,
	// Self binding, parameter binding, and method execution
	return e.adapter.CallInheritedMethod(self, methodName, args)
}
