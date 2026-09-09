package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Helper methods are type extensions that add methods to types that don't
// natively have them (e.g., str.ToUpper(), arr.Push(), num.ToString()).

// HelperInfo is the runtime-owned helper declaration and its shared class state.
type HelperInfo = *runtime.MutableHelperInfo

// HelperMethodResult represents the result of a helper method lookup.
type HelperMethodResult struct {
	OwnerHelper HelperInfo
	Method      *ast.FunctionDecl
	BuiltinSpec string
	Overloads   []*ast.FunctionDecl
}

// zeroArgHelperOverload returns the overload callable with no arguments, or
// nil when none exists. Falls back to the primary method when there are no
// recorded overloads.
func zeroArgHelperOverload(result *HelperMethodResult) *ast.FunctionDecl {
	if result == nil {
		return nil
	}
	if len(result.Overloads) == 0 {
		if result.Method != nil && helperASTMethodEffectiveParamCount(result.Method) == 0 {
			return result.Method
		}
		return nil
	}
	for _, m := range result.Overloads {
		if m != nil && helperASTMethodEffectiveParamCount(m) == 0 {
			return m
		}
	}
	return nil
}

func helperASTMethodEffectiveParamCount(method *ast.FunctionDecl) int {
	if method == nil {
		return 0
	}
	count := len(method.Parameters)
	if method.IsHelper {
		count--
	}
	if count < 0 {
		return 0
	}
	return count
}

func helperResultAliasWouldShadowTarget(selfValue Value, name string) bool {
	if selfValue == nil {
		return false
	}

	if objVal, ok := selfValue.(ObjectValue); ok {
		if ident.Equal(name, "ClassName") || ident.Equal(name, "ClassType") {
			return true
		}
		if objVal.GetField(name) != nil || objVal.HasProperty(name) || objVal.HasMethod(name) {
			return true
		}
		if _, found := objVal.GetClassVar(name); found {
			return true
		}
	}

	if recVal, ok := selfValue.(RecordInstanceValue); ok {
		if rec, ok := recVal.(*runtime.RecordValue); ok {
			if _, found := rec.Fields[ident.Normalize(name)]; found {
				return true
			}
		}
		if recVal.HasRecordMethod(name) || recVal.HasRecordProperty(name) {
			return true
		}
	}

	if _, ok := selfValue.(*runtime.TypeMetaValue); ok {
		if ident.Equal(name, "ClassName") || ident.Equal(name, "ClassType") {
			return true
		}
	}

	return false
}

func isCurrentHelperMethod(ctx *ExecutionContext, name string) bool {
	if ctx == nil || ctx.Env() == nil {
		return false
	}
	raw, ok := ctx.Env().Get("__CurrentHelperMethod__")
	if !ok {
		return false
	}
	current, ok := raw.(*runtime.StringValue)
	return ok && ident.Equal(current.Value, name)
}

func helperIsBuiltin(helper HelperInfo) bool {
	if helper == nil {
		return false
	}
	name := helper.GetName()
	return ident.HasSuffix(name, "IntrinsicHelper") || ident.Equal(name, "TArrayHelper")
}

func orderedHelpersForLookup(helpers []HelperInfo) []HelperInfo {
	if len(helpers) <= 1 {
		return helpers
	}
	// User-declared helpers are searched in declaration order (the first
	// declared helper wins, matching DWScript; see HelpersPass/helper_precedence),
	// except that a helper inheriting from an earlier one overrides its
	// ancestor. Built-in helpers are fallback.
	var user []HelperInfo
	for _, helper := range helpers {
		if !helperIsBuiltin(helper) {
			user = append(user, helper)
		}
	}
	// Move descendants ahead of their ancestors, keeping declaration order
	// among unrelated helpers.
	ordered := make([]HelperInfo, 0, len(helpers))
	for _, helper := range user {
		insertAt := len(ordered)
		for idx, placed := range ordered {
			if helperDescendsFrom(helper, placed) {
				insertAt = idx
				break
			}
		}
		ordered = append(ordered, nil)
		copy(ordered[insertAt+1:], ordered[insertAt:])
		ordered[insertAt] = helper
	}
	for _, helper := range helpers {
		if helperIsBuiltin(helper) {
			ordered = append(ordered, helper)
		}
	}
	return ordered
}

// helperDescendsFrom reports whether helper inherits (directly or
// transitively) from ancestor.
func helperDescendsFrom(helper, ancestor HelperInfo) bool {
	for cur := helper; cur != nil; {
		parent := cur.GetParentHelper()
		if parent == nil {
			return false
		}
		if parent == ancestor {
			return true
		}
		cur = parent
	}
	return false
}

// getHelpersForValue returns all helpers that apply to the given value's type.
func (e *Evaluator) getHelpersForValue(val Value) []HelperInfo {
	if e.typeSystem == nil {
		return nil
	}

	// Get the type name from the value
	var typeName string
	switch v := val.(type) {
	case ArrayAccessor:
		// Try specific array type first, then generic "array" helpers
		var combined []HelperInfo
		arrayTypeStr := v.ArrayTypeString()
		specific := ident.Normalize(arrayTypeStr)
		if helpers := e.typeSystem.LookupHelpers(specific); helpers != nil {
			combined = append(combined, helpers...)
		}
		if arrayVal, ok := v.(*runtime.ArrayValue); ok {
			if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
				// For static arrays, also try the dynamic array equivalent
				dynArrayType := types.NewDynamicArrayType(arrayVal.ArrayType.ElementType)
				dynSpecific := ident.Normalize(dynArrayType.String())
				if helpers := e.typeSystem.LookupHelpers(dynSpecific); helpers != nil {
					combined = append(combined, helpers...)
				}
			}
		}

		if helpers := e.typeSystem.LookupHelpers("array"); helpers != nil {
			combined = append(combined, helpers...)
		}
		for _, helpersAny := range e.typeSystem.AllHelpers() {
			for _, helper := range helpersAny {
				target := types.GetUnderlyingType(helper.GetTargetType())
				if _, ok := target.(*types.ArrayType); !ok {
					continue
				}
				if containsHelperInfo(combined, helper) {
					continue
				}
				combined = append(combined, helper)
			}
		}
		return combined

	case EnumAccessor:
		// Try specific enum type first, then generic "enum" helpers
		var combined []HelperInfo
		enumTypeName := "enum"
		// Try to get actual enum type name if available
		type enumWithTypeName interface {
			Value
			GetEnumTypeName() string
		}
		if ev, ok := val.(enumWithTypeName); ok {
			enumTypeName = ev.GetEnumTypeName()
		}
		specific := ident.Normalize(enumTypeName)
		if helpers := e.typeSystem.LookupHelpers(specific); helpers != nil {
			combined = append(combined, helpers...)
		}
		if helpers := e.typeSystem.LookupHelpers("enum"); helpers != nil {
			combined = append(combined, helpers...)
		}
		return combined

	case ObjectValue:
		typeName = v.ClassName()
	case ClassMetaValue:
		typeName = v.GetClassName()
	case *runtime.InterfaceInstance:
		typeName = v.InterfaceName()
	case RecordInstanceValue:
		typeName = v.GetRecordTypeName()
	case *runtime.IntegerValue:
		typeName = "Integer"
	case *runtime.FloatValue:
		typeName = "Float"
	case *runtime.StringValue:
		typeName = "String"
	case *runtime.BooleanValue:
		typeName = "Boolean"
	case *runtime.TypeMetaValue:
		if v.TypeName != "" {
			typeName = v.TypeName
		} else if v.TypeInfo != nil {
			typeName = v.TypeInfo.String()
		} else {
			return nil
		}
	default:
		return nil
	}

	// Look up helpers for this type
	helpers := e.typeSystem.LookupHelpers(typeName)
	if helpers == nil {
		return nil
	}
	return helpers
}

// FindHelperMethod searches all applicable helpers for a method with the given name.
// Returns the owning helper, method declaration (if any), and builtin spec identifier.
// Later helpers override earlier ones; AST methods are checked before builtin-only.
func (e *Evaluator) FindHelperMethod(val Value, methodName string) *HelperMethodResult {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil
	}

	// User-declared helpers keep declaration order; built-in helpers are fallback.
	for _, helper := range orderedHelpersForLookup(helpers) {

		if overloads, ownerHelperAny, ok := helper.GetMethodOverloads(methodName); ok && len(overloads) > 0 {
			ownerHelper := ownerHelperAny
			if ownerHelper == nil {
				ownerHelper = helper
			}

			// Also check for builtin spec
			if spec, _, ok := ownerHelper.GetBuiltinMethod(methodName); ok {
				return &HelperMethodResult{
					OwnerHelper: ownerHelper,
					Method:      overloads[len(overloads)-1],
					Overloads:   overloads,
					BuiltinSpec: spec,
				}
			}
			return &HelperMethodResult{
				OwnerHelper: ownerHelper,
				Method:      overloads[len(overloads)-1],
				Overloads:   overloads,
				BuiltinSpec: "",
			}
		}
	}

	// If no declared method, check for builtin-only entries
	for _, helper := range orderedHelpersForLookup(helpers) {
		if spec, ownerHelperAny, ok := helper.GetBuiltinMethod(methodName); ok {
			ownerHelper := ownerHelperAny
			if ownerHelper == nil {
				// Should not happen if registered correctly
				return nil
			}
			return &HelperMethodResult{
				OwnerHelper: ownerHelper,
				Method:      nil,
				BuiltinSpec: spec,
			}
		}
	}

	return nil
}

func (e *Evaluator) findHelperMethodInHelper(helper HelperInfo, methodName string) *HelperMethodResult {
	if helper == nil {
		return nil
	}

	if overloads, ownerHelperAny, ok := helper.GetMethodOverloads(methodName); ok && len(overloads) > 0 {
		ownerHelper := ownerHelperAny
		if ownerHelper == nil {
			ownerHelper = helper
		}
		result := &HelperMethodResult{
			OwnerHelper: ownerHelper,
			Method:      overloads[len(overloads)-1],
			Overloads:   overloads,
		}
		if spec, _, ok := ownerHelper.GetBuiltinMethod(methodName); ok {
			result.BuiltinSpec = spec
		}
		return result
	}

	if spec, ownerHelperAny, ok := helper.GetBuiltinMethod(methodName); ok {
		ownerHelper := ownerHelperAny
		if ownerHelper == nil {
			ownerHelper = helper
		}
		return &HelperMethodResult{
			OwnerHelper: ownerHelper,
			BuiltinSpec: spec,
		}
	}

	return nil
}

// Helper interfaces for value types - enable helper method resolution.

// ArrayAccessor is an optional interface for array values.
type ArrayAccessor interface {
	Value
	// ArrayTypeString returns the array type as a string (e.g., "array of String").
	ArrayTypeString() string
}

// Marker interfaces for primitive types (actual implementations in interp package).
type IntegerValue interface{ Value }
type FloatValue interface{ Value }
type StringValue interface{ Value }
type BooleanValue interface{ Value }

// CallHelperMethod executes a helper method (builtin or AST) on a value.
func (e *Evaluator) CallHelperMethod(
	result *HelperMethodResult,
	selfValue Value,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if result == nil {
		return e.newError(node, "helper method not found")
	}

	// If it's a builtin method, handle it directly
	if result.BuiltinSpec != "" {
		return e.CallBuiltinHelperMethod(result.BuiltinSpec, selfValue, args, node, ctx)
	}

	// If it's an AST method, execute it with proper Self binding
	if result.Method != nil {
		if len(result.Overloads) > 1 {
			for _, candidate := range result.Overloads {
				expected := len(candidate.Parameters)
				if candidate.IsHelper {
					expected--
				}
				if expected == len(args) {
					return e.CallASTHelperMethod(result.OwnerHelper, candidate, selfValue, args, node, ctx)
				}
			}
		}
		return e.CallASTHelperMethod(result.OwnerHelper, result.Method, selfValue, args, node, ctx)
	}

	return e.newError(node, "helper method has no implementation")
}

// CallBuiltinHelperMethod executes a builtin helper method.
// Tries type-specific helpers in order; unhandled specs are treated as missing
// evaluator support rather than bouncing through the legacy path.
func (e *Evaluator) CallBuiltinHelperMethod(spec string, selfValue Value, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	// Try each helper type in order
	if result := e.evalStringHelper(spec, selfValue, args, node); result != nil {
		return result
	}
	if result := e.evalIntegerHelper(spec, selfValue, args, node); result != nil {
		return result
	}
	if result := e.evalFloatHelper(spec, selfValue, args, node); result != nil {
		return result
	}
	if result := e.evalBooleanHelper(spec, selfValue, args, node); result != nil {
		return result
	}
	if result := e.evalArrayHelper(spec, selfValue, args, node, ctx); result != nil {
		return result
	}
	if result := e.evalEnumHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Specs that are plain builtin names (PadLeft, StripAccents, ...) are
	// implemented in internal/builtins; the receiver is the builtin's first argument.
	// Evaluator-owned specs always carry the "__" prefix, so skip the registry
	// lookup for them rather than paying for a guaranteed miss on every call.
	if !strings.HasPrefix(spec, "__") {
		if fn, ok := builtins.DefaultRegistry.Lookup(spec); ok {
			return fn(e.builtinContext(ctx), append([]Value{selfValue}, args...))
		}
	}

	return e.newError(node, "unknown built-in helper method '%s'", spec)
}

// CallBuiltinHelperProperty executes a built-in helper property read.
func (e *Evaluator) CallBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node, ctx *ExecutionContext) Value {
	// Property specs that are plain builtin names (e.g. StripAccents) are
	// implemented in internal/builtins with the receiver as the only argument.
	// Evaluator-owned specs always carry the "__" prefix, so skip the registry
	// lookup for them to keep hot property reads (.Length, .High) cheap.
	if !strings.HasPrefix(propSpec, "__") {
		if fn, ok := builtins.DefaultRegistry.Lookup(propSpec); ok {
			return fn(e.builtinContext(ctx), []Value{selfValue})
		}
	}
	return e.evalBuiltinHelperProperty(propSpec, selfValue, node, ctx)
}

// CallASTHelperMethod executes a user-defined helper method (with AST body).
// Sets up environment with Self, class vars/consts, parameters, and Result variable.
func (e *Evaluator) CallASTHelperMethod(
	helper HelperInfo,
	method *ast.FunctionDecl,
	selfValue Value,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if method == nil {
		return e.newError(node, "helper method not implemented")
	}

	// Safety check - helper can be nil if OwnerHelper lookup failed
	if helper == nil {
		return e.newError(node, "helper method not found")
	}
	// Check if the interface wraps a nil pointer

	expectedArgCount := len(method.Parameters)
	if method.IsHelper {
		expectedArgCount--
	}
	if expectedArgCount < 0 {
		expectedArgCount = 0
	}
	if len(args) != expectedArgCount {
		return e.newError(node, "wrong number of arguments for helper method '%s': expected %d, got %d",
			method.Name.Value, expectedArgCount, len(args))
	}

	// Create method environment (enclosed scope)
	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())

	// Bind Self to the target value (the value being extended)
	scope.defineExposed(ctx, "Self", selfValue)
	scope.defineExposed(ctx, "__CurrentHelperMethod__", &runtime.StringValue{Value: method.Name.Value})
	scope.defineExposed(ctx, "__CurrentHelperName__", &runtime.StringValue{Value: helper.GetName()})
	if method.IsHelper && len(method.Parameters) > 0 {
		scope.defineExposed(ctx, method.Parameters[0].Name.Value, selfValue)
	}

	// Bind helper class vars and consts from entire inheritance chain.
	// Walk from root parent to current helper so child helpers override parents.
	e.bindHelperChainVarsConsts(helper, ctx, scope)

	// Bind method parameters
	for idx, param := range method.Parameters {
		if method.IsHelper && idx == 0 {
			continue
		}
		argIdx := idx
		if method.IsHelper {
			argIdx--
		}
		scope.defineOwned(e, ctx, param.Name.Value, args[argIdx])
	}

	// For functions, initialize the Result variable
	if method.ReturnType != nil {
		returnType, err := e.ResolveTypeFromAnnotation(method.ReturnType, ctx)
		if err != nil {
			return e.newError(node, "failed to resolve return type: %v", err)
		}
		defaultVal := e.GetDefaultValue(returnType, ctx)
		scope.defineOwned(e, ctx, "Result", defaultVal)
		// Also define method name as alias for Result (Pascal convention)
		if !helperResultAliasWouldShadowTarget(selfValue, method.Name.Value) {
			scope.defineExposed(ctx, method.Name.Value, defaultVal)
		}
	}

	// Execute method body
	result := e.Eval(method.Body, ctx)
	if isError(result) {
		return result
	}

	// Extract return value
	if method.ReturnType != nil {
		returnValue := e.extractReturnValue(method.Name.Value, ctx)
		return e.retainValueForBinding(returnValue, ctx)
	}

	// For procedures, return nil
	return e.nilValue()
}

// bindHelperChainVarsConsts binds class vars and consts from the helper inheritance chain.
// Walks from root to current helper so child definitions override parents.
func (e *Evaluator) bindHelperChainVarsConsts(helper HelperInfo, ctx *ExecutionContext, scope *bindingScope) {
	if helper == nil {
		return
	}

	// Check if helper interface wraps a nil pointer

	// Build helper chain from root to current
	var helperChain []HelperInfo
	for h := helper; h != nil; {
		helperChain = append([]HelperInfo{h}, helperChain...)
		parent := h.GetParentHelper()
		if parent == nil {
			break
		}
		h = parent
	}

	// Bind vars and consts (root first, so children override).
	// Class vars bind as references into the helper's storage so mutations
	// are shared across calls (including nested helper method calls).
	for _, h := range helperChain {
		if h == nil {
			continue
		}
		classVars := h.GetClassVars()
		for name := range classVars {
			varName := name
			getter := func() (runtime.Value, error) {
				return classVars[varName], nil
			}
			setter := func(v runtime.Value) error {
				classVars[varName] = v
				return nil
			}
			scope.defineExposed(ctx, name, runtime.NewReferenceValue(name, getter, setter))
		}
		for name, value := range h.GetClassConsts() {
			scope.defineExposed(ctx, name, value)
		}
	}
}

// extractReturnValue extracts the return value from a function's environment.
// Checks Result first, then method name alias (Pascal convention).
func (e *Evaluator) extractReturnValue(methodName string, ctx *ExecutionContext) Value {
	// Check Result variable first
	if resultVal, ok := ctx.Env().Get("Result"); ok {
		if runtime.KindOf(resultVal) != runtime.KindNil {
			return resultVal
		}
	}

	// Check method name alias
	if methodNameVal, ok := ctx.Env().Get(methodName); ok {
		if runtime.KindOf(methodNameVal) != runtime.KindNil {
			return methodNameVal
		}
	}

	// Fallback to Result even if NIL
	if resultVal, ok := ctx.Env().Get("Result"); ok {
		return resultVal
	}

	// Final fallback
	return e.nilValue()
}

// ============================================================================
// Helper Utilities
// ============================================================================

func containsHelperInfo(helpers []HelperInfo, target HelperInfo) bool {
	for _, helper := range helpers {
		if helper == target {
			return true
		}
	}
	return false
}

// ============================================================================
// Helper Property Resolution
// ============================================================================

// findHelperClassMember searches all applicable helpers for a class constant or
// class variable with the given name. It is used to resolve helper class members
// accessed through a type's metaclass value (e.g. `String.Hello`, `TMyArray.ByeBye`).
// Class constants take precedence over class variables.
func (e *Evaluator) findHelperClassMember(val Value, name string) (Value, bool) {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil, false
	}
	for _, helper := range orderedHelpersForLookup(helpers) {
		if helper == nil {
			continue
		}
		for constName, constVal := range helper.GetClassConsts() {
			if ident.Equal(constName, name) {
				return constVal, true
			}
		}
		for varName, varVal := range helper.GetClassVars() {
			if ident.Equal(varName, name) {
				return varVal, true
			}
		}
	}
	return nil, false
}

// FindHelperProperty searches all applicable helpers for a property with the given name.
func (e *Evaluator) FindHelperProperty(val Value, propName string) (HelperInfo, *types.PropertyInfo) {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil
	}

	for _, helper := range orderedHelpersForLookup(helpers) {

		// Use GetProperty which searches the inheritance chain and returns the owner helper
		if propInfo, ownerHelperAny, found := helper.GetProperty(propName); found && propInfo != nil {
			return ownerHelperAny, propInfo
		}
	}

	return nil, nil
}

// executeHelperPropertyRead evaluates a helper property read access.
// Handles PropAccessField, PropAccessMethod, PropAccessExpression, PropAccessBuiltin,
// and PropAccessNone.
func (e *Evaluator) executeHelperPropertyRead(
	helper HelperInfo,
	propInfo *types.PropertyInfo,
	selfValue Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	switch propInfo.ReadKind {
	case types.PropAccessField:
		// For records, try direct field access first
		if recVal, ok := selfValue.(RecordInstanceValue); ok {
			if fieldValue, exists := recVal.GetRecordField(propInfo.ReadSpec); exists {
				return fieldValue
			}
		}
		// Otherwise try as getter method
		normalizedReadSpec := ident.Normalize(propInfo.ReadSpec)
		if method, methodOwnerAny, ok := helper.GetMethod(normalizedReadSpec); ok {
			methodOwner := methodOwnerAny
			if methodOwner == nil {
				// Should not happen
				return e.newError(node, "invalid helper method owner")
			}
			var builtinSpec string
			if spec, _, ok := methodOwner.GetBuiltinMethod(normalizedReadSpec); ok {
				builtinSpec = spec
			}
			result := &HelperMethodResult{
				OwnerHelper: methodOwner,
				Method:      method,
				BuiltinSpec: builtinSpec,
			}
			return e.CallHelperMethod(result, selfValue, []Value{}, node, ctx)
		}
		return e.newError(node, "property '%s' read specifier '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessMethod:
		normalizedReadSpec := ident.Normalize(propInfo.ReadSpec)
		if method, methodOwnerAny, ok := helper.GetMethod(normalizedReadSpec); ok {
			methodOwner := methodOwnerAny
			if methodOwner == nil {
				// Should not happen
				return e.newError(node, "invalid helper method owner")
			}
			var builtinSpec string
			if spec, _, ok := methodOwner.GetBuiltinMethod(normalizedReadSpec); ok {
				builtinSpec = spec
			}
			result := &HelperMethodResult{
				OwnerHelper: methodOwner,
				Method:      method,
				BuiltinSpec: builtinSpec,
			}
			return e.CallHelperMethod(result, selfValue, []Value{}, node, ctx)
		}
		return e.newError(node, "property '%s' getter method '%s' not found",
			propInfo.Name, propInfo.ReadSpec)

	case types.PropAccessExpression:
		return e.executeHelperPropertyExpressionRead(propInfo, selfValue, node, ctx)

	case types.PropAccessBuiltin:
		return e.CallBuiltinHelperProperty(propInfo.ReadSpec, selfValue, node, ctx)

	case types.PropAccessNone:
		return e.newError(node, "property '%s' is write-only", propInfo.Name)

	default:
		return e.newError(node, "property '%s' has no read access", propInfo.Name)
	}
}

// executeHelperPropertyExpressionRead evaluates an expression-form helper property
// getter, e.g. `class property MultBy2 : Integer read (2*Field)` on a class helper.
// A class property binds the extended type's class variables and a metaclass Self;
// an instance property binds the receiver's own members. Both cases reuse the
// class/record/object accessor scopes rather than defining a helper-only one.
func (e *Evaluator) executeHelperPropertyExpressionRead(
	propInfo *types.PropertyInfo,
	selfValue Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if propInfo.IsClassProperty {
		if classInfo := e.helperReceiverClassInfo(selfValue); classInfo != nil {
			return e.evalClassPropertyExpressionRead(classInfo, propInfo, node, ctx)
		}
	}
	if recVal, ok := selfValue.(RecordInstanceValue); ok {
		return e.evalHelperRecordExpressionRead(recVal, propInfo, node, ctx)
	}
	return e.executeExpressionBackedPropertyRead(selfValue, propInfo, node, ctx)
}

// evalHelperRecordExpressionRead evaluates an expression-form helper property on a
// record receiver. Records carry class variables of their own, so the accessor needs
// the same fields-plus-class-state scope a record method body gets; the object path
// binds instance fields only and would leave a `class var` unresolved.
func (e *Evaluator) evalHelperRecordExpressionRead(
	recVal RecordInstanceValue,
	propInfo *types.PropertyInfo,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	exprNode, ok := propInfo.ReadExpr.(ast.Expression)
	if !ok {
		return e.newError(node, "property '%s' has invalid read expression type", propInfo.Name)
	}
	if groupedExpr, ok := exprNode.(*ast.GroupedExpression); ok {
		exprNode = groupedExpr.Expression
	}

	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())

	scope.defineExposed(ctx, "Self", recVal)
	if !propInfo.IsClassProperty {
		e.bindRecordMethodFields(recVal, ctx, scope)
	}
	e.bindRecordMethodClassState(recVal, ctx, scope)

	return e.Eval(exprNode, ctx)
}

// evalHelperRecordExpressionWrite mirrors evalHelperRecordExpressionRead for setters,
// binding the implicit `Value` and syncing class-variable writes back afterwards.
func (e *Evaluator) evalHelperRecordExpressionWrite(
	recVal RecordInstanceValue,
	propInfo *types.PropertyInfo,
	value Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	stmt, ok := propInfo.WriteExpr.(ast.Statement)
	if !ok {
		return e.newError(node, "property '%s' has invalid write statement type", propInfo.Name)
	}

	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())

	scope.defineExposed(ctx, "Self", recVal)
	if !propInfo.IsClassProperty {
		e.bindRecordMethodFields(recVal, ctx, scope)
	}
	e.bindRecordMethodClassState(recVal, ctx, scope)
	scope.defineOwned(e, ctx, "Value", value)

	if result := e.Eval(stmt, ctx); isError(result) {
		return result
	}
	e.syncRecordMethodClassState(recVal, ctx)
	return value
}

// executeHelperPropertyExpressionWrite executes an expression-form helper property
// setter. It mirrors executeHelperPropertyExpressionRead.
func (e *Evaluator) executeHelperPropertyExpressionWrite(
	propInfo *types.PropertyInfo,
	selfValue Value,
	value Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if propInfo.IsClassProperty {
		if classInfo := e.helperReceiverClassInfo(selfValue); classInfo != nil {
			return e.evalClassPropertyExpressionWrite(classInfo, propInfo, value, node, ctx)
		}
	}
	if recVal, ok := selfValue.(RecordInstanceValue); ok {
		return e.evalHelperRecordExpressionWrite(recVal, propInfo, value, node, ctx)
	}
	return e.executeExpressionBackedPropertyWrite(selfValue, propInfo, value, node, ctx)
}

// helperReceiverClassInfo resolves the class metadata behind a helper receiver, so
// a `class property` declared in a class helper can be evaluated in class context
// whether it was reached through an instance or through the class name. Returns nil
// for receivers that carry no class metadata (records, scalars), leaving the caller
// to fall back to the instance-shaped accessor scope.
func (e *Evaluator) helperReceiverClassInfo(selfValue Value) runtime.IClassInfo {
	if classMeta, ok := selfValue.(ClassMetaValue); ok {
		return classMeta.GetClassInfo()
	}
	if objVal, ok := selfValue.(ObjectValue); ok {
		if classValAny, err := e.typeSystem.CreateClassValue(objVal.ClassName()); err == nil {
			if classMeta, ok := classValAny.(ClassMetaValue); ok {
				return classMeta.GetClassInfo()
			}
		}
	}
	return nil
}

func (e *Evaluator) executeHelperPropertyWrite(
	helper HelperInfo,
	propInfo *types.PropertyInfo,
	selfValue Value,
	value Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if propInfo.WriteKind == types.PropAccessExpression {
		return e.executeHelperPropertyExpressionWrite(propInfo, selfValue, value, node, ctx)
	}
	if propInfo.WriteKind == types.PropAccessNone || propInfo.WriteSpec == "" {
		return e.newError(node, readOnlyPropertyWriteMessage)
	}

	normalizedWriteSpec := ident.Normalize(propInfo.WriteSpec)
	if method, methodOwnerAny, ok := helper.GetMethod(normalizedWriteSpec); ok {
		methodOwner := methodOwnerAny
		if methodOwner == nil {
			return e.newError(node, "invalid helper method owner")
		}
		result := &HelperMethodResult{
			OwnerHelper: methodOwner,
			Method:      method,
			Overloads:   []*ast.FunctionDecl{method},
		}
		return e.CallHelperMethod(result, selfValue, []Value{value}, node, ctx)
	}

	return e.newError(node, "property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
}

// evalBuiltinHelperProperty evaluates a built-in helper property (array, enum, string, etc.).
func (e *Evaluator) evalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node, ctx *ExecutionContext) Value {
	switch types.BuiltinHelperOperation(propSpec) {
	case types.HelperArrayLength, types.HelperArrayCount, types.HelperArrayHigh, types.HelperArrayLow:
		if _, ok := selfValue.(ArrayAccessor); !ok {
			return e.newError(node, "built-in property '%s' can only be used on arrays", propSpec)
		}
		return e.evalArrayHelper(propSpec, selfValue, nil, node, ctx)

	case types.HelperEnumValue:
		enumVal, ok := selfValue.(EnumAccessor)
		if !ok {
			return e.newError(node, "Enum.Value property requires enum receiver")
		}
		return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}

	case types.HelperEnumName, types.HelperEnumQualifiedName:
		return e.evalEnumHelper(propSpec, selfValue, nil, node)

	case types.HelperStringLength:
		if _, ok := selfValue.(StringValue); !ok {
			return e.newError(node, "String.Length property requires string receiver")
		}
		return e.evalStringHelper(propSpec, selfValue, nil, node)

	case types.HelperStringIsASCII, types.HelperStringTrim, types.HelperStringTrimLeft, types.HelperStringTrimRight:
		return e.evalStringHelper(propSpec, selfValue, nil, node)

	case types.HelperIntegerToString:
		return e.evalIntegerHelper(propSpec, selfValue, nil, node)

	case types.HelperFloatToStringDefault:
		return e.evalFloatHelper(propSpec, selfValue, nil, node)

	case types.HelperBooleanToString:
		return e.evalBooleanHelper(propSpec, selfValue, nil, node)

	default:
		return e.newError(node, "unknown built-in property '%s'", propSpec)
	}
}

func (e *Evaluator) registerFunctionHelper(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	if len(node.Parameters) == 0 || node.Parameters[0].Type == nil {
		return e.newError(node, "helper function '%s' must declare at least one typed parameter", node.Name.Value)
	}

	targetType, err := e.ResolveTypeFromAnnotation(node.Parameters[0].Type, ctx)
	if err != nil {
		return e.newError(node, "unknown target type '%s' for helper function '%s'",
			node.Parameters[0].Type.String(), node.Name.Value)
	}

	methodName := node.Name.Value
	if node.HelperName != nil {
		methodName = node.HelperName.Value
	}

	helperInfo := runtime.NewMutableHelperInfo("__"+methodName+"FunctionHelper", targetType, false)
	methodKey := ident.Normalize(methodName)
	helperInfo.Methods[methodKey] = node
	helperInfo.MethodOverloads[methodKey] = append(helperInfo.MethodOverloads[methodKey], node)

	typeName := ident.Normalize(targetType.String())
	e.typeSystem.RegisterHelper(typeName, helperInfo)

	simpleTypeName := ident.Normalize(extractSimpleTypeName(targetType.String()))
	if simpleTypeName != typeName {
		e.typeSystem.RegisterHelper(simpleTypeName, helperInfo)
	}

	return &runtime.NilValue{}
}

// Extracts simple type name from qualified string ("array of Integer" -> "array").
func extractSimpleTypeName(typeName string) string {
	if idx := strings.Index(typeName, " "); idx != -1 {
		return typeName[:idx]
	}
	return typeName
}
