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
// **STATUS**: Consolidated to use DispatchMethodCall
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

	if identObj, ok := node.Object.(*ast.Identifier); ok {
		if _, exists := ctx.Env().Get(identObj.Value); !exists {
			unitExists := false
			if e.UnitRegistry() != nil {
				_, unitExists = e.UnitRegistry().GetUnit(identObj.Value)
			}
			if unitExists {
				return e.executeQualifiedFunctionCall(identObj.Value, node.Method, node.Arguments, node, ctx)
			}
		}
	}

	// Evaluate the object first
	obj := e.Eval(node.Object, ctx)
	if isError(obj) {
		return obj
	}

	methodName := node.Method.Value

	if recordVal, ok := obj.(RecordInstanceValue); ok {
		if methodDecl, found := recordVal.GetRecordMethod(methodName); found {
			args, err := e.prepareArgsForParameters(methodDecl.Parameters, node.Arguments, ctx)
			if err != nil {
				return e.newError(node, "%s", err.Error())
			}
			return e.callRecordMethod(recordVal, methodDecl, args, node, ctx)
		}
	}

	// When the target method is unambiguous and declares var/lazy parameters,
	// wrap the corresponding arguments (by-ref references / lazy thunks) so
	// writes inside the method reach the caller's variable. This covers
	// constructors and methods (e.g. TMyClass.Create(var x); see fixture
	// oop_field). Overloaded methods keep the plain path: their resolution
	// happens later during dispatch.
	if decl := lookupUnambiguousMethodDecl(obj, methodName, len(node.Arguments)); decl != nil {
		args, err := e.prepareArgsForParameters(decl.Parameters, node.Arguments, ctx)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return e.DispatchMethodCall(obj, methodName, args, node, ctx)
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

	// This provides unified error handling and consistent routing for all value types.
	// See method_dispatch.go for full documentation of the dispatch architecture.
	return e.DispatchMethodCall(obj, methodName, args, node, ctx)
}

// lookupUnambiguousMethodDecl resolves the declaration of a method call target
// when it can be determined statically: the method is not overloaded and its
// parameter count matches the call. Returns nil when the declaration cannot be
// (or does not need to be) resolved; callers then evaluate arguments by value.
// Only declarations that actually use var/lazy parameters are returned, so the
// common case keeps the existing evaluation path.
func lookupUnambiguousMethodDecl(obj Value, methodName string, argCount int) *ast.FunctionDecl {
	var decl *ast.FunctionDecl

	switch o := obj.(type) {
	case ObjectValue:
		if d, ok := o.GetMethodDecl(methodName).(*ast.FunctionDecl); ok && d != nil {
			decl = d
		} else if d, ok := o.GetClassMethodDecl(methodName).(*ast.FunctionDecl); ok && d != nil {
			decl = d
		}
	case ClassMetaValue:
		if classInfo := o.GetClassInfo(); classInfo != nil {
			decl = classInfo.GetConstructor(methodName)
			if decl == nil {
				decl = classInfo.LookupMethod(methodName)
			}
			if decl == nil {
				decl = classInfo.LookupClassMethod(methodName)
			}
		}
	}

	if decl == nil || decl.IsOverload || len(decl.Parameters) != argCount {
		return nil
	}
	if hasVarOrLazyParams(decl) {
		return decl
	}
	return nil
}

// hasVarOrLazyParams reports whether a declaration has any var (by-ref) or
// lazy parameter, i.e. whether argument preparation needs the declaration.
func hasVarOrLazyParams(decl *ast.FunctionDecl) bool {
	for _, param := range decl.Parameters {
		if param.ByRef || param.IsLazy {
			return true
		}
	}
	return false
}

// VisitInheritedExpression evaluates an 'inherited' expression.
//
// **COMPLEXITY**: High (~176 lines in original implementation)
// **STATUS**: Migrated to use ObjectValue.CallInheritedMethod interface
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

	return e.executeInheritedCallDirect(self, methodName, args, node, ctx)
}
