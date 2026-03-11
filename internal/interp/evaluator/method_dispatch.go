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
// Future improvement: Migrate CallMethod logic into the evaluator package
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
	"fmt"
	"strings"

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

	// Unwrap type casts so method dispatch uses the underlying value.
	// This preserves static type information for class variables while allowing
	// method calls like TMyClass(o).PrintMyName to dispatch on the actual object.
	if castVal, ok := obj.(TypeCastAccessor); ok {
		obj = castVal.GetWrappedValue()
		if obj == nil {
			return e.newError(node, "method call on nil value")
		}
	} else if obj.Type() == "TYPE_CAST" {
		return e.newError(node, "internal error: TYPE_CAST value does not implement TypeCastAccessor interface")
	}

	normalizedMethod := ident.Normalize(methodName)

	if recordType, ok := obj.(*RecordTypeValue); ok {
		return e.callRecordStaticMethod(recordType, methodName, args, node, ctx)
	}

	if recordVal, ok := obj.(RecordInstanceValue); ok {
		if methodDecl, found := recordVal.GetRecordMethod(methodName); found {
			return e.callRecordMethod(recordVal, methodDecl, args, node, ctx)
		}
		if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		return e.newError(node, "method '%s' not found for type '%s'", methodName, obj.Type())
	}

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
		if normalizedMethod == "free" {
			return obj
		}
		// Nil object error - consistent behavior
		return e.newError(node, "Object not instantiated")

	// ============================================================
	// Helper-based dispatch (builtin/AST helper methods)
	// ============================================================

	case "STRING", "INTEGER", "FLOAT", "BOOLEAN", "ARRAY", "VARIANT", "ENUM":
		return e.dispatchHelperMethod(obj, methodName, args, node, ctx)

	// ============================================================
	// Evaluator-owned dispatch for OOP types
	// ============================================================

	case "OBJECT":
		return e.dispatchObjectMethod(obj, methodName, args, node, ctx)

	case "INTERFACE":
		intfInst, ok := obj.(*runtime.InterfaceInstance)
		if !ok {
			return e.newError(node, "internal error: INTERFACE value is not *runtime.InterfaceInstance")
		}
		result := e.dispatchInterfaceMethodDirect(intfInst, methodName, args, node, ctx)
		if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil && shouldFallbackToHelper(result) {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		return result

	case "CLASS", "CLASSINFO":
		classMeta, ok := obj.(ClassMetaValue)
		if !ok {
			return e.newError(node, "internal error: %s value does not implement ClassMetaValue", obj.Type())
		}
		// Handle overloaded class methods via evaluator-owned dispatch
		if classInfo, ok := classMeta.GetClassInfo().(runtime.IClassInfo); ok {
			if !classMeta.HasConstructor(methodName) && classInfo.HasClassMethodOverloads(methodName) {
				return e.dispatchClassMethodOverloaded(classMeta, classInfo, methodName, args, node, ctx)
			}
		}
		if classMeta.HasConstructor(methodName) {
			return e.callClassConstructor(classMeta, methodName, args, node, ctx)
		}
		return e.callClassMethod(classMeta, methodName, args, node, ctx)

	// ============================================================
	// Unknown type - try helper method or error
	// ============================================================

	default:
		// Try helper method lookup first (might be a custom type with helpers)
		helperResult := e.FindHelperMethod(obj, methodName)
		if helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		return e.newError(node, "method '%s' not found for type '%s'", methodName, obj.Type())
	}
}

func (e *Evaluator) callClassConstructor(classMeta ClassMetaValue, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	classInfoAny := classMeta.GetClassInfo()
	classInfo, ok := classInfoAny.(runtime.IClassInfo)
	if !ok || classInfo == nil {
		return e.newError(node, "invalid class reference")
	}

	obj := runtime.NewObjectInstance(classInfo)
	if initErr := e.initializeObjectFields(classInfo, obj, node, ctx); initErr != nil {
		return initErr
	}

	if err := e.executeConstructorForObject(obj, methodName, args, node, ctx); err != nil {
		return e.newError(node, "constructor failed: %v", err)
	}

	return obj
}

func (e *Evaluator) callClassMethod(classMeta ClassMetaValue, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	if len(args) == 0 {
		if result, invoked := classMeta.InvokeParameterlessClassMethod(methodName, func(methodDecl any) Value {
			return e.executeClassMethodDirect(classMeta, methodDecl, nil, node, ctx)
		}); invoked {
			return result
		}
	}

	if result, ok := classMeta.CreateClassMethodPointer(methodName, func(methodDecl any) Value {
		return e.executeClassMethodDirect(classMeta, methodDecl, args, node, ctx)
	}); ok {
		return result
	}

	return e.newError(node, "class method '%s' not found", methodName)
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
		ordinal, err := runtime.GetOrdinalValue(args[0])
		if err != nil {
			return e.newError(node, "Include requires ordinal value: %s", err.Error())
		}
		setVal.AddElement(ordinal)
		return e.nilValue()

	case "exclude":
		if len(args) != 1 {
			return e.newError(node, "Exclude expects 1 argument, got %d", len(args))
		}
		ordinal, err := runtime.GetOrdinalValue(args[0])
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

// dispatchObjectMethod handles method calls on OBJECT instance values.
//
// Dispatch order:
//  1. Destroyed-object guard (raises Exception)
//  2. Free/destructor alias
//  3. Method lookup via class hierarchy (virtual dispatch via most-derived-first search)
//  4. Explicit destructor call (IsDestructor flag)
//  5. Overload check — delegates to evaluator-owned overload resolver
//  6. Class method (static) lookup
//  7. Helper method fallback
func (e *Evaluator) dispatchObjectMethod(obj Value, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	objInst, ok := obj.(*runtime.ObjectInstance)
	if !ok {
		return e.newError(node, "internal error: OBJECT value is not *runtime.ObjectInstance")
	}

	if objInst.Destroyed {
		pos := node.Pos()
		message := fmt.Sprintf("Object already destroyed [line: %d, column: %d]", pos.Line, pos.Column+1)
		exc := e.createException("Exception", message, &pos, ctx)
		ctx.SetException(exc)
		return e.nilValue()
	}

	classInfo := objInst.Class
	if classInfo == nil {
		return e.newError(node, "object has no class information")
	}

	normalizedName := ident.Normalize(methodName)

	// Free is a universal TObject method that delegates to Destroy
	if normalizedName == "free" {
		if len(args) != 0 {
			return e.newError(node, "Free takes no arguments")
		}
		return e.runObjectDestructor(objInst, classInfo.LookupMethod("Destroy"), node, ctx)
	}

	// Dispatch to evaluator-owned overload resolver when the method has overloads.
	if classInfo.HasMethodOverloads(methodName) {
		return e.dispatchObjectMethodOverloaded(objInst, methodName, args, node, ctx)
	}

	// Single-method dispatch: look up via class hierarchy (most-derived first —
	// this is virtual dispatch without overload ambiguity).
	method := classInfo.LookupMethod(methodName)
	if method != nil {
		if method.IsDestructor {
			return e.runObjectDestructor(objInst, method, node, ctx)
		}
		return e.executeObjectMethodDirect(obj, method, args, node, ctx)
	}

	// Try class (static) method
	if classMethod := classInfo.LookupClassMethod(methodName); classMethod != nil {
		return e.executeObjectMethodDirect(obj, classMethod, args, node, ctx)
	}

	// Helper method fallback (type helpers extend built-in and user-defined types)
	if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil {
		return e.CallHelperMethod(helperResult, obj, args, node, ctx)
	}

	return e.newError(node, "method '%s' not found in class '%s'", methodName, classInfo.GetName())
}

// runObjectDestructor executes an object's destructor and marks the object as destroyed.
func (e *Evaluator) runObjectDestructor(obj *runtime.ObjectInstance, destructor *ast.FunctionDecl, node ast.Node, ctx *ExecutionContext) Value {
	if obj == nil {
		return e.nilValue()
	}
	if obj.Destroyed {
		return e.nilValue()
	}

	if destructor == nil {
		obj.Destroyed = true
		obj.RefCount = 0
		return e.nilValue()
	}

	obj.DestroyCallDepth++
	defer func() {
		obj.DestroyCallDepth--
		if obj.DestroyCallDepth == 0 {
			obj.Destroyed = true
			obj.RefCount = 0
		}
	}()

	return e.executeObjectMethodDirect(obj, destructor, nil, node, ctx)
}

func shouldFallbackToHelper(result Value) bool {
	if !isError(result) {
		return false
	}
	msg := strings.ToLower(result.String())
	return strings.Contains(msg, "method") && strings.Contains(msg, "not found")
}
