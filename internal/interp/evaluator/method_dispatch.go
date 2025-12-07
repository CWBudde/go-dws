// Package evaluator provides method dispatch infrastructure for the DWScript interpreter.
//
// # Method Dispatch Architecture
//
// This file documents and implements the consolidated method dispatch infrastructure.
// Method calls in DWScript are complex, supporting multiple dispatch modes based on
// the receiver type and method being called.
//
// ## 15 Distinct Method Call Modes
//
// Method calls are evaluated in the following order:
//
//  1. UNIT-QUALIFIED FUNCTION CALLS - UnitName.FunctionName()
//  2. STATIC CLASS METHOD CALLS - TClass.Method()
//  3. RECORD TYPE STATIC METHOD CALLS - TRecord.Method()
//  4. CLASSINFO VALUE METHOD CALLS - ClassInfoValue.Method()
//  5. METACLASS CONSTRUCTOR CALLS - ClassValue.Create()
//  6. SET VALUE BUILT-IN METHODS - SetValue.Include/Exclude()
//  7. RECORD INSTANCE METHOD CALLS - RecordValue.Method()
//  8. INTERFACE INSTANCE METHOD CALLS - InterfaceInstance.Method()
//  9. NIL OBJECT ERROR HANDLING - Always raises "Object not instantiated"
//  10. ENUM TYPE META METHODS - TypeMetaValue.Low/High/ByName()
//  11. HELPER METHOD CALLS - any_type.HelperMethod()
//  12. OBJECT INSTANCE METHOD CALLS - ObjectInstance.Method()
//  13. VIRTUAL CONSTRUCTOR DISPATCH - obj.Create()
//  14. CLASS METHOD EXECUTION - executeClassMethod
//  15. OVERLOAD RESOLUTION - resolveMethodOverload
//
// ## Dispatch Strategy
//
// The method dispatch uses a type-based routing strategy:
//
//	┌─────────────────────────────────────────────────────────────────────────┐
//	│                     VisitMethodCallExpression                           │
//	│                              │                                          │
//	│              ┌───────────────┼───────────────┐                          │
//	│              ▼               ▼               ▼                          │
//	│      Interface-based    Adapter-based    Helper-based                   │
//	│        Dispatch          Dispatch         Dispatch                      │
//	│              │               │               │                          │
//	│  ┌───────────┴───┐    ┌──────┴─────┐   ┌─────┴──────┐                   │
//	│  │ SET, TYPE_META│    │ OBJECT     │   │ STRING     │                   │
//	│  │ (direct)      │    │ INTERFACE  │   │ INTEGER    │                   │
//	│  └───────────────┘    │ CLASSINFO  │   │ FLOAT      │                   │
//	│                       │ CLASS      │   │ BOOLEAN    │                   │
//	│                       │ RECORD     │   │ ARRAY      │                   │
//	│                       └────────────┘   │ VARIANT    │                   │
//	│                                        │ ENUM       │                   │
//	│                                        └────────────┘                   │
//	└─────────────────────────────────────────────────────────────────────────┘
//
// ## Interface-Based Dispatch (Target Architecture)
//
// The following interfaces enable direct method dispatch without adapter:
//
//   - SetMethodDispatcher: Include(), Exclude() methods on set values
//   - EnumTypeMetaDispatcher: Low(), High(), ByName() methods on enum type meta
//
// These interfaces are implemented directly on value types, allowing the evaluator
// to dispatch methods without going through the adapter layer. This is the target
// architecture for all method dispatch - adapter calls should be eliminated.
//
// ## Adapter-Based Dispatch (Legacy - To Be Eliminated)
//
// Complex value types (OBJECT, INTERFACE, CLASSINFO, CLASS, RECORD) currently use
// adapter.CallMethod() for method dispatch. This is LEGACY code that should be
// migrated to interface-based dispatch. The adapter is needed because these types
// currently require interpreter-level operations:
//
//   - Environment setup (Self binding, parameter binding)
//   - Call stack management (recursion tracking)
//   - Method overload resolution
//   - Virtual method dispatch
//   - Constructor chains
//
// Phase 16 (Tasks 3.5.111-3.5.114) extended CallMethod in interpreter.go to handle
// all value types. The next step is to migrate this logic into the evaluator package
// to eliminate the adapter dependency entirely.
//
// ## Helper-Based Dispatch
//
// Primitive types (STRING, INTEGER, FLOAT, BOOLEAN, ARRAY, VARIANT, ENUM) use
// helper methods for type extension. These are dispatched via:
//
//   - FindHelperMethod(): Locates the helper method for a value type
//   - CallHelperMethod(): Executes the helper (builtin or AST-defined)
//
// ## Error Handling
//
// Method dispatch follows a consistent error handling strategy:
//
//   - Missing method: Returns error "method '%s' not found for type '%s'"
//   - Nil receiver: Returns error "Object not instantiated"
//   - Wrong argument count: Returns error with expected vs actual count
//   - Type mismatch: Returns error describing the type constraint

package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// MethodCallResult encapsulates the result of a method dispatch operation.
// It provides structured information about the dispatch outcome.
type MethodCallResult struct {
	// Value is the result of the method call, or an error value.
	Value Value
	// Handled indicates if the method call was successfully dispatched.
	// If false, the caller should try alternative dispatch mechanisms.
	Handled bool
	// MethodFound indicates if the method was found on the receiver type.
	// Used for error reporting when a method doesn't exist.
	MethodFound bool
}

// DispatchMethodCall routes method calls to the appropriate handler based on value type.
// This is the consolidated entry point for all method dispatch in the evaluator.
//
// Parameters:
//   - obj: The receiver value (object, record, set, etc.)
//   - methodName: The method name to call (case-insensitive)
//   - args: Evaluated argument values
//   - node: The AST node for error reporting
//   - ctx: The execution context
//
// Returns:
//   - Value: The method result or error value
//
// Dispatch order:
//  1. Interface-based dispatch (SET, TYPE_META)
//  2. Helper method dispatch (STRING, INTEGER, etc.)
//  3. Adapter-based dispatch (OBJECT, INTERFACE, CLASSINFO, CLASS, RECORD)
//  4. Error for unknown types
func (e *Evaluator) DispatchMethodCall(obj Value, methodName string, args []Value, node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	if obj == nil {
		return e.newError(node, "method call on nil value")
	}

	normalizedMethod := ident.Normalize(methodName)

	// Route based on object type
	switch obj.Type() {
	// ============================================================
	// Interface-based dispatch (direct, no adapter)
	// ============================================================

	case "SET":
		return e.dispatchSetMethod(obj, normalizedMethod, methodName, args, node)

	case "TYPE_META":
		return e.dispatchEnumTypeMetaMethod(obj, normalizedMethod, methodName, args, node)

	case "NIL":
		// Nil object error - consistent behavior
		return e.newError(node, "Object not instantiated")

	// ============================================================
	// Helper-based dispatch (builtin/AST helper methods)
	// ============================================================

	case "STRING", "INTEGER", "FLOAT", "BOOLEAN", "ARRAY", "VARIANT", "ENUM":
		return e.dispatchHelperMethod(obj, methodName, args, node, ctx)

	// ============================================================
	// Adapter-based dispatch (complex types requiring environment setup)
	// ============================================================

	case "OBJECT", "INTERFACE", "CLASSINFO", "CLASS", "RECORD":
		// Task 3.8.4: Check for helper methods first before delegating to adapter
		// This allows class helpers to extend object types
		helperResult := e.FindHelperMethod(obj, methodName)
		if helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}

		// No helper method - delegate to adapter for class/instance method handling
		// These types require full environment setup (Self binding, call stack)
		return e.adapter.CallMethod(obj, methodName, args, node)

	// ============================================================
	// Unknown type - try helper method or error
	// ============================================================

	default:
		// Try helper method lookup first (might be a custom type with helpers)
		helperResult := e.FindHelperMethod(obj, methodName)
		if helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}

		// No handler found - delegate to adapter as last resort
		return e.adapter.EvalNode(node)
	}
}

// dispatchSetMethod handles method calls on SET values.
// Implements SetMethodDispatcher interface dispatch.
//
// Supported methods:
//   - Include(value): Add an element to the set
//   - Exclude(value): Remove an element from the set
func (e *Evaluator) dispatchSetMethod(obj Value, normalizedMethod, methodName string, args []Value, node ast.Node) Value {
	setVal, ok := obj.(SetMethodDispatcher)
	if !ok {
		return e.newError(node, "internal error: SET value does not implement SetMethodDispatcher")
	}

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
}

// dispatchEnumTypeMetaMethod handles method calls on TYPE_META values (enum types).
// Implements EnumTypeMetaDispatcher interface dispatch.
//
// Supported methods:
//   - Low(): Returns lowest ordinal value
//   - High(): Returns highest ordinal value
//   - ByName(name): Returns ordinal value for enum name
func (e *Evaluator) dispatchEnumTypeMetaMethod(obj Value, normalizedMethod, methodName string, args []Value, node ast.Node) Value {
	enumMeta, ok := obj.(EnumTypeMetaDispatcher)
	if !ok {
		return e.newError(node, "internal error: TYPE_META value does not implement EnumTypeMetaDispatcher")
	}

	// Only enum types have these methods
	if !enumMeta.IsEnumTypeMeta() {
		return e.newError(node, "method '%s' not found for type '%s'", methodName, obj.String())
	}

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
}

// dispatchHelperMethod handles method calls via helper methods (type extensions).
// Helper methods extend built-in types with additional functionality.
//
// Examples:
//   - str.ToUpper() - String helper
//   - arr.Push(x) - Array helper
//   - num.ToString() - Integer helper
func (e *Evaluator) dispatchHelperMethod(obj Value, methodName string, args []Value, node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	helperResult := e.FindHelperMethod(obj, methodName)
	if helperResult == nil {
		return e.newError(node, "cannot call method '%s' on type '%s' (no helper found)", methodName, obj.Type())
	}

	return e.CallHelperMethod(helperResult, obj, args, node, ctx)
}
