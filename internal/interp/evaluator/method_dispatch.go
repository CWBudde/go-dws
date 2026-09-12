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
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
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
		// Preserve the cast's static type on nil receivers so nil-receiver
		// dispatch (e.g. TObj(nil).Method) can resolve the method statically.
		if nilVal, isNil := obj.(*runtime.NilValue); isNil && nilVal.GetTypedClassName() == "" {
			obj = &runtime.NilValue{ClassType: castVal.GetStaticTypeName()}
		}
	} else if runtime.KindOf(obj) == runtime.KindTypeCast {
		return e.newError(node, "internal error: TYPE_CAST value does not implement TypeCastAccessor interface")
	}

	normalizedMethod := ident.Normalize(methodName)

	if recordType, ok := obj.(*RecordTypeValue); ok {
		return e.callRecordStaticMethod(recordType, methodName, args, node, ctx)
	}

	if recordVal, ok := obj.(RecordInstanceValue); ok {
		// Overload-aware record instance method dispatch
		if rec, ok := obj.(*runtime.RecordValue); ok {
			if overloads := rec.GetRecordMethodOverloads(methodName); len(overloads) > 1 {
				if selected, err := e.selectOverload(rec.GetRecordTypeName(), methodName, overloads, args, ctx); err == nil {
					return e.callRecordMethod(recordVal, selected, args, node, ctx)
				}
			}
		}
		if methodDecl, found := recordVal.GetRecordMethod(methodName); found {
			return e.callRecordMethod(recordVal, methodDecl, args, node, ctx)
		}
		if rec, ok := obj.(*runtime.RecordValue); ok && rec.RecordType != nil {
			if recordTypeRaw := e.typeSystem.LookupRecord(rec.RecordType.Name); recordTypeRaw != nil {
				if recordTypeRaw.HasStaticMethod(methodName) {
					return e.callRecordStaticMethod(recordTypeRaw, methodName, args, node, ctx)
				}
			}
		}
		if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		return e.newError(node, "method '%s' not found for type '%s'", methodName, obj.Type())
	}

	// Route based on object type
	switch runtime.KindOf(obj) {
	// ============================================================
	// Interface-based dispatch (direct, no adapter)
	// ============================================================

	case runtime.KindSet:
		return e.dispatchSetMethod(obj, normalizedMethod, methodName, args, node)

	case runtime.KindByteBuffer:
		return e.DispatchByteBufferMethod(obj, methodName, args, node, ctx)

	case runtime.KindTypeMeta:
		if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		if normalizedMethod == "classname" && len(args) == 0 {
			if meta, ok := obj.(*runtime.TypeMetaValue); ok {
				if meta.TypeName != "" {
					return &runtime.StringValue{Value: meta.TypeName}
				}
				if meta.TypeInfo != nil {
					return &runtime.StringValue{Value: meta.TypeInfo.String()}
				}
			}
		}
		return e.dispatchEnumTypeMetaMethod(obj, normalizedMethod, methodName, args, node)

	case runtime.KindNil:
		if normalizedMethod == "free" {
			return obj
		}
		return e.dispatchMethodOnNilObject(obj, methodName, args, node, ctx)

	// ============================================================
	// Helper-based dispatch (builtin/AST helper methods)
	// ============================================================

	case runtime.KindString, runtime.KindInteger, runtime.KindFloat, runtime.KindBoolean, runtime.KindArray, runtime.KindVariant, runtime.KindEnum:
		return e.dispatchHelperMethod(obj, methodName, args, node, ctx)

	// ============================================================
	// Evaluator-owned dispatch for OOP types
	// ============================================================

	case runtime.KindObject:
		return e.dispatchObjectMethod(obj, methodName, args, node, ctx)

	case runtime.KindInterface:
		intfInst, ok := obj.(*runtime.InterfaceInstance)
		if !ok {
			return e.newError(node, "internal error: INTERFACE value is not *runtime.InterfaceInstance")
		}
		result := e.dispatchInterfaceMethodDirect(intfInst, methodName, args, node, ctx)
		if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil && shouldFallbackToHelper(result) {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		return result

	case runtime.KindClass, runtime.KindClassInfo:
		classMeta, ok := obj.(ClassMetaValue)
		if !ok {
			return e.newError(node, "internal error: %s value does not implement ClassMetaValue", obj.Type())
		}
		if helperResult := e.FindHelperMethod(obj, methodName); helperResult != nil {
			return e.CallHelperMethod(helperResult, obj, args, node, ctx)
		}
		// Handle overloaded class methods via evaluator-owned dispatch
		if classInfo := classMeta.GetClassInfo(); classInfo != nil {
			classOverloads := classInfo.GetClassMethodOverloads(methodName)
			if classMeta.HasConstructor(methodName) && len(classOverloads) > 0 {
				// A constructor name shared with class methods: resolve across the
				// merged overload set and route on what was selected.
				merged := append(classInfo.GetConstructorOverloads(methodName), classOverloads...)
				selected, err := e.selectCallableOverload(classInfo.GetName(), methodName, merged, args, ctx)
				if err != nil {
					return e.newError(node, "%s", err.Error())
				}
				if !selected.IsConstructor {
					return e.executeClassMethodDirect(classMeta, selected, args, node, ctx)
				}
				return e.callClassConstructor(classMeta, methodName, args, node, ctx)
			}
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

// dispatchMethodOnNilObject implements DWScript's nil-receiver call semantics:
//   - Non-virtual instance methods resolved from the receiver's static type are
//     executed with Self = nil (the error only surfaces if Self is dereferenced
//     inside the body).
//   - Virtual methods, class methods, constructors and destructors require the
//     instance's dynamic class and raise "Object not instantiated" at the method
//     name's position.
func (e *Evaluator) dispatchMethodOnNilObject(obj Value, methodName string, args []Value, node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	var objectExpr ast.Expression
	if node != nil {
		objectExpr = node.Object
	}
	if classInfo := e.staticClassInfoForNilReceiver(obj, objectExpr); classInfo != nil {
		method := classInfo.LookupMethod(methodName)
		// With overloads, pick the best match for the argument types rather than
		// whatever LookupMethod happens to return first.
		if overloads := classInfo.GetMethodOverloads(methodName); len(overloads) > 1 {
			if selected, err := e.selectCallableOverload(classInfo.GetName(), methodName, overloads, args, ctx); err == nil {
				method = selected
			}
		}
		if method != nil && isNonVirtualInstanceMethod(classInfo, method) {
			return e.executeMethodWithClassInfo(obj, classInfo, method, args, ctx)
		}

		// Helper class methods (class function ... static) resolve statically
		// and can be invoked on a nil receiver.
		if method == nil {
			for info := classInfo; info != nil; info = info.GetParent() {
				if helpersAny := e.typeSystem.LookupHelpers(info.GetName()); helpersAny != nil {
					for _, helper := range orderedHelpersForLookup(helpersAny) {
						if result := e.findHelperMethodInHelper(helper, methodName); result != nil {
							return e.CallHelperMethod(result, obj, args, node, ctx)
						}
					}
				}
			}
		}
	}

	// Virtual dispatch (or unknown static type) needs an instance.
	return e.newError(methodNameErrorNode(node), "%s", e.nilReceiverMessage(obj, objectExpr, ctx))
}

// isNonVirtualInstanceMethod reports whether a method can be invoked on a nil
// receiver: DWScript statically dispatches non-virtual instance methods, so
// only virtual methods, class methods, constructors and destructors require an
// instantiated object at the call site. The class's virtual method table is
// consulted as well because method lookup may return the implementation
// declaration, which does not carry the virtual/override flags.
func isNonVirtualInstanceMethod(classInfo runtime.IClassInfo, method *runtime.MethodMetadata) bool {
	if method.IsVirtual || method.IsOverride || method.IsAbstract ||
		method.IsClassMethod || method.IsConstructor || method.IsDestructor {
		return false
	}
	if vmt := classInfo.GetVirtualMethodTable(); vmt != nil {
		sig := ident.Normalize(method.Name) + "_" + strconv.Itoa(len(method.Parameters))
		if _, isVirtual := vmt[sig]; isVirtual {
			return false
		}
	}
	return true
}

// staticClassInfoForNilReceiver resolves the static class of a nil receiver,
// using the typed nil's class when available and falling back to the semantic
// analyzer's type annotation for the receiver expression.
func (e *Evaluator) staticClassInfoForNilReceiver(obj Value, objectExpr ast.Expression) runtime.IClassInfo {
	className := ""
	if nilVal, ok := obj.(NilAccessor); ok {
		className = nilVal.GetTypedClassName()
	}
	if className == "" && objectExpr != nil && e.SemanticInfo() != nil {
		if typeAnnot := e.SemanticInfo().GetType(objectExpr); typeAnnot != nil {
			className = typeAnnot.Name
		}
	}
	if className == "" {
		return nil
	}
	if classInfo, ok := e.typeSystem.LookupClass(className).(runtime.IClassInfo); ok {
		return classInfo
	}
	return nil
}

// nilReceiverMessage picks the message DWScript reports for a call or member
// access on a nil receiver. A nil *metaclass* — a `class of X` variable that
// was never assigned — is a different mistake from a nil object reference, and
// upstream names it differently, so the receiver's static type decides.
func (e *Evaluator) nilReceiverMessage(obj Value, objectExpr ast.Expression, ctx *ExecutionContext) string {
	if nilVal, ok := obj.(*runtime.NilValue); ok && nilVal.IsMetaclass {
		return "ClassType is nil"
	}
	if e.isMetaclassExpression(objectExpr, ctx) {
		return "ClassType is nil"
	}
	return "Object not instantiated"
}

// isMetaclassExpression reports whether an expression's static type is a
// metaclass ("class of X"), resolving through type aliases.
func (e *Evaluator) isMetaclassExpression(objectExpr ast.Expression, ctx *ExecutionContext) bool {
	if objectExpr == nil || e.SemanticInfo() == nil {
		return false
	}
	annotation := e.SemanticInfo().GetType(objectExpr)
	if annotation == nil || annotation.Name == "" {
		return false
	}
	resolved, err := e.resolveTypeName(annotation.Name, ctx)
	if err != nil || resolved == nil {
		return false
	}
	_, isClassOf := types.GetUnderlyingType(resolved).(*types.ClassOfType)
	return isClassOf
}

// staticReceiverClassInfo resolves the class a receiver expression is *declared*
// as, which is what non-virtual dispatch resolves against. Returns nil when the
// call site has no usable static type (an unannotated expression, or a type that
// is not a class), in which case the caller keeps its dynamic resolution.
func (e *Evaluator) staticReceiverClassInfo(node ast.Node, ctx *ExecutionContext) runtime.IClassInfo {
	call, ok := node.(*ast.MethodCallExpression)
	if !ok || call.Object == nil || e.SemanticInfo() == nil {
		return nil
	}
	annotation := e.SemanticInfo().GetType(call.Object)
	if annotation == nil || annotation.Name == "" {
		return nil
	}
	className := annotation.Name
	if resolved := e.resolveClassAliasName(className, ctx); resolved != "" {
		className = resolved
	}
	return e.typeSystem.LookupClass(className)
}

// staticallyDispatchedMethod returns the method a call must run when the
// receiver's declared type hides a same-named method further down the
// hierarchy. DWScript, like Delphi, binds a non-virtual method at compile time
// against the declared type, so `var a : TA := TB.Create; a.P;` runs TA.P when
// TB merely redeclares P, and TB.P only when P is virtual and overridden.
//
// Returns nil whenever dynamic resolution is already correct: no static type at
// the call site, the declared class does not know the name, or the method is
// virtual/abstract/a constructor or destructor.
func (e *Evaluator) staticallyDispatchedMethod(
	dynamicClass runtime.IClassInfo,
	methodName string,
	argCount int,
	node ast.Node,
	ctx *ExecutionContext,
) (runtime.IClassInfo, *runtime.MethodMetadata) {
	staticClass := e.staticReceiverClassInfo(node, ctx)
	if staticClass == nil || dynamicClass == nil {
		return nil, nil
	}
	// Same class means dynamic lookup already answers with the declared one.
	if ident.Equal(staticClass.GetName(), dynamicClass.GetName()) {
		return nil, nil
	}
	// Only narrow to the declared type when the runtime class actually derives
	// from it; an unrelated annotation must not redirect the call.
	if !classDerivesFrom(dynamicClass, staticClass.GetName()) {
		return nil, nil
	}

	method := staticMethodForArity(staticClass, methodName, argCount)
	if method == nil {
		return nil, nil
	}
	if method.IsConstructor || method.IsDestructor {
		return nil, nil
	}

	sig := ident.Normalize(method.Name) + "_" + strconv.Itoa(len(method.Parameters))
	var staticEntry *runtime.VirtualMethodEntry
	if vmt := staticClass.GetVirtualMethodTable(); vmt != nil {
		staticEntry = vmt[sig]
	}
	staticIsVirtual := staticEntry != nil ||
		method.IsVirtual || method.IsOverride || method.IsAbstract

	if staticIsVirtual {
		return resolveVirtualChain(staticClass, dynamicClass, staticEntry, sig)
	}

	return staticClass, method
}

// resolveVirtualChain resolves a virtual call against the runtime class's
// virtual method table rather than by name.
//
// The distinction matters for `reintroduce`: a reintroduced method does not
// take over its ancestor's slot, so a call typed at the ancestor must reach
// neither it nor anything declared below it. Name lookup would return the
// most-derived declaration and ignore the broken chain.
//
// `reintroduce; virtual` is the harder case: it starts a *new* chain that
// shares a signature with the one it hides, and the table is keyed by signature
// alone, so the new chain overwrites the old slot. The class that first
// declared each chain tells them apart — when the runtime class's chain is not
// the one the declared type can see, the call stays on the declared type's.
func resolveVirtualChain(
	staticClass, dynamicClass runtime.IClassInfo,
	staticEntry *runtime.VirtualMethodEntry,
	sig string,
) (runtime.IClassInfo, *runtime.MethodMetadata) {
	vmt := dynamicClass.GetVirtualMethodTable()
	if vmt == nil {
		return nil, nil
	}
	entry := vmt[sig]
	if entry == nil || entry.Method == nil {
		return nil, nil
	}

	if isDifferentVirtualChain(entry, staticEntry) {
		owner := methodDeclaringClass(staticClass, staticEntry.Method)
		if owner == nil {
			owner = staticEntry.OwningClass
		}
		return owner, staticEntry.Method
	}

	owner := entry.OwningClass
	if resolved := methodDeclaringClass(dynamicClass, entry.Method); resolved != nil {
		owner = resolved
	}
	if owner == nil {
		return nil, nil
	}
	return owner, entry.Method
}

// isDifferentVirtualChain reports whether the runtime class's table entry
// belongs to a chain the declared type cannot see, which happens when a
// descendant restarted the chain with `reintroduce; virtual`.
func isDifferentVirtualChain(entry, staticEntry *runtime.VirtualMethodEntry) bool {
	if staticEntry == nil || staticEntry.Method == nil {
		return false
	}
	if entry.OwningClass == nil || staticEntry.OwningClass == nil {
		return false
	}
	return !ident.Equal(entry.OwningClass.GetName(), staticEntry.OwningClass.GetName())
}

// staticMethodForArity picks the declared type's method for a call of the given
// arity. It returns nil when the name is genuinely overloaded at that arity,
// because choosing between real overloads is the overload resolver's job and
// requires the argument types this path does not see.
func staticMethodForArity(staticClass runtime.IClassInfo, methodName string, argCount int) *runtime.MethodMetadata {
	candidates := append(
		append([]*runtime.MethodMetadata(nil), staticClass.GetMethodOverloads(methodName)...),
		staticClass.GetClassMethodOverloads(methodName)...,
	)
	if len(candidates) == 0 {
		method := staticClass.LookupMethod(methodName)
		if method == nil {
			method = staticClass.LookupClassMethod(methodName)
		}
		return method
	}

	var match *runtime.MethodMetadata
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		if !methodAcceptsArgCount(candidate, argCount) {
			continue
		}
		if match != nil {
			// Two declarations of the same name accept this call: a real
			// overload set. Leave it to the overload resolver.
			return nil
		}
		match = candidate
	}
	return match
}

// methodAcceptsArgCount reports whether a method can be called with argCount
// arguments, accounting for parameters that have defaults.
func methodAcceptsArgCount(method *runtime.MethodMetadata, argCount int) bool {
	if argCount > len(method.Parameters) {
		return false
	}
	required := 0
	for _, param := range method.Parameters {
		if param.DefaultValue == nil {
			required++
		}
	}
	return argCount >= required
}

// methodDeclaringClass finds the class in a hierarchy that owns a method, so a
// statically bound call executes with the right class context.
func methodDeclaringClass(from runtime.IClassInfo, method *runtime.MethodMetadata) runtime.IClassInfo {
	if method == nil {
		return nil
	}
	for current := from; current != nil; current = current.GetParent() {
		if current.LookupMethod(method.Name) == method || current.LookupClassMethod(method.Name) == method {
			return current
		}
	}
	return nil
}

// executeStaticallyDispatchedMethod runs a method bound to the receiver's
// declared class. It cannot go through executeObjectMethodDirect, which
// re-resolves the method name against the *runtime* class and would undo the
// static binding.
func (e *Evaluator) executeStaticallyDispatchedMethod(
	self Value,
	staticClass runtime.IClassInfo,
	method *runtime.MethodMetadata,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if method.IsClassMethod {
		classVal, err := e.typeSystem.CreateClassValue(staticClass.GetName())
		if err != nil || classVal == nil {
			return e.newError(node, "class method execution requires runtime class value")
		}
		classMeta, ok := classVal.(ClassMetaValue)
		if !ok {
			return e.newError(node, "class method execution requires runtime class value")
		}
		return e.executeClassMethodDirect(classMeta, method, args, node, ctx)
	}
	return e.executeMethodWithClassInfo(self, staticClass, method, args, ctx)
}

// classDerivesFrom reports whether classInfo is ancestorName or descends from it.
func classDerivesFrom(classInfo runtime.IClassInfo, ancestorName string) bool {
	for current := classInfo; current != nil; current = current.GetParent() {
		if ident.Equal(current.GetName(), ancestorName) {
			return true
		}
	}
	return false
}

// isExceptionClassInfo reports whether classInfo is the Exception base class or
// one of its descendants.
func isExceptionClassInfo(classInfo runtime.IClassInfo) bool {
	for current := classInfo; current != nil; current = current.GetParent() {
		if ident.Equal(current.GetName(), "Exception") {
			return true
		}
	}
	return false
}

// methodNameErrorNode picks the AST node whose position DWScript reports for
// receiver errors (nil or destroyed object): the method/member name identifier
// when available, otherwise the whole expression.
func methodNameErrorNode(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.MethodCallExpression:
		if n.Method != nil {
			return n.Method
		}
	case *ast.MemberAccessExpression:
		if n.Member != nil {
			return n.Member
		}
	}
	return node
}

func (e *Evaluator) callClassConstructor(classMeta ClassMetaValue, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	classInfo := classMeta.GetClassInfo()
	if classInfo == nil {
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
		if result, invoked := classMeta.InvokeParameterlessClassMethod(methodName, func(methodDecl *runtime.MethodMetadata) Value {
			return e.executeClassMethodDirect(classMeta, methodDecl, nil, node, ctx)
		}); invoked {
			return result
		}
	}

	if result, ok := classMeta.CreateClassMethodPointer(methodName, func(methodDecl *runtime.MethodMetadata) Value {
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
	// The analyzer records the receiver's static type when a helper resolves
	// against it; prefer helpers registered for that exact (alias) type so
	// strict helpers dispatch on the declared type rather than the dynamic one.
	if node != nil && node.Method != nil && e.SemanticInfo() != nil {
		if annot := e.SemanticInfo().GetType(node.Method); annot != nil && strings.HasPrefix(annot.Name, "__helper_receiver:") {
			target := strings.TrimPrefix(annot.Name, "__helper_receiver:")
			if helpersAny := e.typeSystem.LookupHelpers(ident.Normalize(target)); helpersAny != nil {
				for _, helper := range orderedHelpersForLookup(helpersAny) {
					if result := e.findHelperMethodInHelper(helper, methodName); result != nil {
						return e.CallHelperMethod(result, obj, args, node, ctx)
					}
				}
			}
		}
	}

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

	// Only an explicit Free/Destroy makes later method calls an error; objects
	// reclaimed eagerly by refcount cleanup stay callable (go-dws can release
	// them earlier than DWScript would), consistent with the member-access path.
	if objInst.ExplicitlyFreed {
		return e.newError(methodNameErrorNode(node), "Object already destroyed")
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

	// A non-virtual method hidden by a same-named one in a descendant binds to
	// the receiver's declared type, not its runtime class. This precedes the
	// overload path: a redeclaration in a descendant looks like an overload set
	// to the registry, but hiding is not overloading.
	if staticClass, staticMethod := e.staticallyDispatchedMethod(classInfo, methodName, len(args), node, ctx); staticMethod != nil {
		return e.executeStaticallyDispatchedMethod(obj, staticClass, staticMethod, args, node, ctx)
	}

	// Dispatch to evaluator-owned overload resolver when the method has
	// overloads. Instance and class (static) methods sharing a name form one
	// overload set for instance receivers.
	if classInfo.HasMethodOverloads(methodName) ||
		(len(classInfo.GetMethodOverloads(methodName))+len(classInfo.GetClassMethodOverloads(methodName)) > 1) {
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
func (e *Evaluator) runObjectDestructor(obj *runtime.ObjectInstance, destructor *runtime.MethodMetadata, node ast.Node, ctx *ExecutionContext) Value {
	if obj == nil {
		return e.nilValue()
	}
	if obj.Destroyed {
		return e.nilValue()
	}

	if destructor == nil {
		obj.Destroyed = true
		obj.ExplicitlyFreed = true
		obj.RefCount = 0
		return e.nilValue()
	}

	obj.DestroyCallDepth++
	defer func() {
		obj.DestroyCallDepth--
		if obj.DestroyCallDepth == 0 {
			obj.Destroyed = true
			obj.ExplicitlyFreed = true
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
