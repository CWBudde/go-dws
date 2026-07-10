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

	// JSON namespace method call: JSON.Parse(s), JSON.Stringify(x), ...
	if e.isJSONNamespaceObject(node.Object, ctx) {
		return e.evalJSONNamespaceCall(node.Method.Value, node.Arguments, node, ctx)
	}
	if e.isDefaultNamespaceObject(node.Object, ctx) {
		builtinCall := &ast.CallExpression{
			TypedExpressionBase: node.TypedExpressionBase,
			Function:            node.Method,
			Arguments:           node.Arguments,
		}
		return e.VisitCallExpression(builtinCall, ctx)
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

	// Method call on a JSON value receiver: v.TypeName(), v.Add(x), ...
	if isJSONBoxed(obj) {
		args := make([]Value, len(node.Arguments))
		for i, arg := range node.Arguments {
			val := e.Eval(arg, ctx)
			if isError(val) {
				return val
			}
			args[i] = val
		}
		return e.evalJSONMethodCall(obj, methodName, args, node, ctx)
	}

	// Associative array built-in methods (Keys/Length/Count/Clear/Delete).
	if assoc, ok := obj.(*runtime.AssociativeArrayValue); ok {
		args := make([]Value, len(node.Arguments))
		for i, arg := range node.Arguments {
			val := e.Eval(arg, ctx)
			if isError(val) {
				return val
			}
			args[i] = val
		}
		if result, handled := e.evalAssociativeArrayMethod(assoc, methodName, args, node); handled {
			return result
		}
	}

	if recordVal, ok := obj.(RecordInstanceValue); ok {
		// Overload-aware: pick the best-matching overload for the provided
		// arguments instead of the first registered one.
		if rec, ok := obj.(*runtime.RecordValue); ok {
			if overloads := rec.GetRecordMethodOverloads(methodName); len(overloads) > 1 {
				argVals := make([]Value, len(node.Arguments))
				for i, arg := range node.Arguments {
					val := e.Eval(arg, ctx)
					if isError(val) {
						return val
					}
					argVals[i] = val
				}
				if selected, err := e.selectOverload(rec.GetRecordTypeName(), methodName, overloads, argVals); err == nil {
					return e.callRecordMethod(recordVal, selected, argVals, node, ctx)
				}
			}
		}
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
	decl := lookupUnambiguousMethodDecl(obj, methodName, len(node.Arguments))
	if decl == nil {
		// A nil receiver can still statically dispatch a non-virtual method
		// (see dispatchMethodOnNilObject); resolve the declaration from the
		// receiver's static type so var/lazy arguments keep their by-ref
		// binding on that path too.
		decl = e.lookupNilReceiverMethodDecl(obj, node, len(node.Arguments))
	}
	if decl != nil {
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

	// A field/class-var/readable-property of function-pointer type is directly
	// callable: o.FEvent(1). The analyzer annotates node.Method with the pointer
	// type when it resolved to such a member (never for a real method), so read
	// the stored pointer and dispatch it instead of looking up a method.
	if e.methodNameIsCallableMember(node.Method) {
		// The receiver was already evaluated into obj above; read the proc-typed
		// member from it directly so a side-effecting receiver
		// (NextObj().FEvent(1)) is not evaluated twice — which would also risk
		// reading the pointer from a different object than the one dispatched.
		memberVal, ok := e.readCallableMemberValue(obj, node, ctx)
		if !ok {
			// Non-object receiver (e.g. a record proc-typed field): fall back to
			// the member-access path.
			memberVal = e.Eval(&ast.MemberAccessExpression{
				TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: node.Token}},
				Object:              node.Object,
				Member:              node.Method,
			}, ctx)
		}
		if isError(memberVal) {
			return memberVal
		}
		return e.executeFunctionPointerDirect(memberVal, args, node, ctx)
	}

	// This provides unified error handling and consistent routing for all value types.
	// See method_dispatch.go for full documentation of the dispatch architecture.
	return e.DispatchMethodCall(obj, methodName, args, node, ctx)
}

// methodNameIsCallableMember reports whether the analyzer annotated this method
// identifier as a function/method pointer type — i.e. the "method" is really a
// proc-typed field, class var, or property to be read and called.
func (e *Evaluator) methodNameIsCallableMember(method *ast.Identifier) bool {
	if method == nil || e.SemanticInfo() == nil {
		return false
	}
	annot := e.SemanticInfo().GetType(method)
	if annot == nil {
		return false
	}
	resolved, err := e.ResolveTypeFromAnnotation(annot)
	if err != nil || resolved == nil {
		return false
	}
	kind := resolved.TypeKind()
	return kind == "FUNCTION_POINTER" || kind == "METHOD_POINTER"
}

// readCallableMemberValue reads a proc-typed member (readable property, field,
// or class var) from an already-evaluated object receiver, so a proc-typed
// member call (o.FEvent(1)) does not re-evaluate the receiver expression. The
// member kinds and precedence mirror VisitMemberAccessExpression's object case.
// It returns (value, true) for an object receiver and (nil, false) for any other
// receiver kind, so the caller can fall back to member-access evaluation.
func (e *Evaluator) readCallableMemberValue(obj Value, node *ast.MethodCallExpression, ctx *ExecutionContext) (Value, bool) {
	objVal, ok := obj.(ObjectValue)
	if !ok {
		return nil, false
	}
	memberName := node.Method.Value

	// Property (respecting getter/setter recursion guards), then field
	// (static-class aware for shadowed fields), then class var.
	propCtx := ctx.PropContext()
	if propCtx == nil || (!propCtx.InPropertyGetter && !propCtx.InPropertySetter) {
		if objVal.HasProperty(memberName) {
			propValue := objVal.ReadProperty(memberName, func(propInfo any) Value {
				return e.executePropertyRead(obj, propInfo, node, ctx)
			})
			if propValue != nil {
				return propValue, true
			}
		}
	}
	if fieldValue := getFieldWithStaticClass(objVal, memberName, e.staticClassNameOf(node.Object, ctx)); fieldValue != nil {
		return fieldValue, true
	}
	// The static-class-scoped read misses when the receiver has no static-type
	// annotation (e.g. a call/index result); fall back to a plain dynamic read.
	if fieldValue := objVal.GetField(memberName); fieldValue != nil {
		return fieldValue, true
	}
	if classVarValue, found := objVal.GetClassVar(memberName); found {
		return classVarValue, true
	}
	return nil, false
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

// lookupNilReceiverMethodDecl resolves the declaration for a method call on a
// nil receiver via the receiver's static type. It mirrors the constraints of
// lookupUnambiguousMethodDecl (not overloaded, matching arity, has var/lazy
// parameters) and only returns declarations that the nil-dispatch path can
// actually execute (non-virtual instance methods).
func (e *Evaluator) lookupNilReceiverMethodDecl(obj Value, node *ast.MethodCallExpression, argCount int) *ast.FunctionDecl {
	if obj == nil || obj.Type() != "NIL" {
		return nil
	}
	classInfo := e.staticClassInfoForNilReceiver(obj, node.Object)
	if classInfo == nil {
		return nil
	}
	decl := classInfo.LookupMethod(node.Method.Value)
	if decl == nil || !isNonVirtualInstanceMethod(classInfo, decl) {
		return nil
	}
	if decl.IsOverload || len(decl.Parameters) != argCount {
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
