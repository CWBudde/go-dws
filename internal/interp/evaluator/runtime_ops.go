package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Subrange and Interface Wrapping
// ============================================================================

// wrapInSubrange creates a SubrangeValue from an integer value.
// Self-contained: replaces e.oopEngine.WrapInSubrange.
func (e *Evaluator) wrapInSubrange(value Value, typeName string, node ast.Node) (Value, error) {
	subrangeType := e.typeSystem.LookupSubrangeType(typeName)
	if subrangeType == nil {
		return nil, fmt.Errorf("subrange type '%s' not found", typeName)
	}
	var intValue int
	switch v := value.(type) {
	case *runtime.IntegerValue:
		intValue = int(v.Value)
	case *runtime.SubrangeValue:
		intValue = v.Value
	default:
		return nil, fmt.Errorf("cannot convert %s to subrange type %s", value.Type(), typeName)
	}
	subrangeVal := &runtime.SubrangeValue{SubrangeType: subrangeType}
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return nil, err
	}
	return subrangeVal, nil
}

// wrapInInterface wraps an object value in an interface instance.
// Self-contained: replaces e.oopEngine.WrapInInterface.
func (e *Evaluator) wrapInInterface(value Value, ifaceName string, node ast.Node) (Value, error) {
	ifaceInfoAny := e.typeSystem.LookupInterface(ifaceName)
	if ifaceInfoAny == nil {
		return nil, fmt.Errorf("interface '%s' not found", ifaceName)
	}
	ifaceInfo := ifaceInfoAny
	// Already wrapped: pass through
	if _, already := value.(*runtime.InterfaceInstance); already {
		return value, nil
	}
	objInst, ok := value.(*runtime.ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("cannot wrap %s in interface %s", value.Type(), ifaceName)
	}
	if e.engineState.RefCountManager == nil {
		// Fallback: create without ref counting
		return &runtime.InterfaceInstance{Interface: ifaceInfo, Object: objInst}, nil
	}
	return e.engineState.RefCountManager.WrapInInterface(ifaceInfo, objInst), nil
}

// ============================================================================
// TypeCast Wrapper
// ============================================================================

// createTypeCastValue creates a TypeCastValue wrapping obj with static type className.
// Self-contained: replaces e.oopEngine.CreateTypeCastWrapper.
// Returns nil if the class is not found.
func (e *Evaluator) createTypeCastValue(className string, obj Value) Value {
	classInfoAny := e.typeSystem.LookupClass(className)
	if classInfoAny == nil {
		return nil
	}
	classInfo, ok := classInfoAny.(runtime.IClassInfo)
	if !ok {
		return nil
	}
	return &interp_TypeCastValue{Object: obj, StaticType: classInfo}
}

// interp_TypeCastValue is an evaluator-local TypeCastValue backed by IClassInfo.
// This replaces the interp.TypeCastValue which uses *ClassInfo.
type interp_TypeCastValue struct {
	Object     Value
	StaticType runtime.IClassInfo
}

func (t *interp_TypeCastValue) Type() string   { return "TYPE_CAST" }
func (t *interp_TypeCastValue) String() string { return t.Object.String() }

// GetStaticTypeName returns the static type name.
func (t *interp_TypeCastValue) GetStaticTypeName() string {
	if t.StaticType == nil {
		return ""
	}
	return t.StaticType.GetName()
}

// GetWrappedValue returns the actual wrapped value.
func (t *interp_TypeCastValue) GetWrappedValue() Value {
	return t.Object
}

// GetStaticClassVar retrieves a class variable using the static type.
func (t *interp_TypeCastValue) GetStaticClassVar(name string) (Value, bool) {
	if t.StaticType == nil {
		return nil, false
	}
	value, owningClass := t.StaticType.LookupClassVar(name)
	return value, owningClass != nil
}

// ============================================================================
// Constructor Execution
// ============================================================================

// executeConstructorForObject runs the named constructor on an already-allocated object.
// Self-contained: replaces e.oopEngine.ExecuteConstructor.
func (e *Evaluator) executeConstructorForObject(obj *runtime.ObjectInstance, constructorName string, args []Value, node ast.Node, ctx *ExecutionContext) error {
	classInfo := obj.Class
	if classInfo == nil {
		return fmt.Errorf("object has no class information")
	}

	// Collect overloads from the class hierarchy
	overloads := classInfo.GetConstructorOverloads(constructorName)

	var constructor *runtime.MethodMetadata
	if len(overloads) == 1 {
		constructor = overloads[0]
	} else if len(overloads) > 1 {
		// Select the best match by argument types (falls back internally to
		// arg-count and default-parameter matching).
		if selected, err := e.selectCallableOverload(classInfo.GetName(), constructorName, overloads, args, ctx); err == nil {
			constructor = selected
		}
		if constructor == nil {
			constructor = overloads[0]
		}
	} else {
		// Fall back to single-constructor lookup
		constructor = classInfo.GetConstructor(constructorName)
	}

	if constructor == nil {
		if len(args) == 0 {
			return nil // Parameterless - implicit default constructor
		}
		return fmt.Errorf("no constructor '%s' found for class '%s' with %d arguments",
			constructorName, classInfo.GetName(), len(args))
	}

	result := e.executeObjectMethodDirect(obj, constructor, args, node, ctx)
	if isError(result) {
		return fmt.Errorf("%s", result.String())
	}
	return nil
}

// ============================================================================
// Method Overload Dispatch
// ============================================================================

// dispatchObjectMethodOverloaded handles instance method dispatch when overloads exist.
// Self-contained: replaces the e.oopEngine.CallMethod path for OBJECT overloads.
func (e *Evaluator) dispatchObjectMethodOverloaded(obj *runtime.ObjectInstance, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	classInfo := obj.Class

	// Instance and class (static) methods sharing a name form a single
	// overload set for instance receivers; route on what was selected.
	overloads := classInfo.GetMethodOverloads(methodName)
	merged := append(append([]*runtime.MethodMetadata{}, overloads...), classInfo.GetClassMethodOverloads(methodName)...)
	if len(merged) > 0 {
		method, err := e.selectCallableOverload(classInfo.GetName(), methodName, merged, args, ctx)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		if !method.IsClassMethod {
			return e.executeObjectMethodDirect(obj, method, args, node, ctx)
		}
		classValAny, err2 := e.typeSystem.CreateClassValue(classInfo.GetName())
		if err2 != nil {
			return e.newError(node, "failed to get class value: %s", err2.Error())
		}
		if cm, ok := classValAny.(ClassMetaValue); ok {
			return e.executeClassMethodDirect(cm, method, args, node, ctx)
		}
		return e.newError(node, "internal error: class meta value not available for '%s'", classInfo.GetName())
	}

	return e.newError(node, "method '%s' not found in class '%s'", methodName, classInfo.GetName())
}

// dispatchInterfaceMethodDirect handles interface method dispatch using evaluator-owned logic.
// Self-contained: replaces e.oopEngine.CallMethod for INTERFACE type.
func (e *Evaluator) dispatchInterfaceMethodDirect(intfInst *runtime.InterfaceInstance, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	if intfInst.Object == nil {
		return e.newError(node, "Interface is nil")
	}
	if !intfInst.HasInterfaceMethod(methodName) {
		return e.newError(node, "method '%s' not found in interface '%s'", methodName, intfInst.InterfaceName())
	}

	objVal := intfInst.Object
	classInfo := objVal.Class

	// Try instance method overloads
	overloads := classInfo.GetMethodOverloads(methodName)
	if len(overloads) > 0 {
		method, err := e.selectCallableOverload(classInfo.GetName(), methodName, overloads, args, ctx)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return e.executeObjectMethodDirect(objVal, method, args, node, ctx)
	}

	// Try single instance method
	if method := classInfo.LookupMethod(methodName); method != nil {
		return e.executeObjectMethodDirect(objVal, method, args, node, ctx)
	}

	// Try class method overloads
	classOverloads := classInfo.GetClassMethodOverloads(methodName)
	if len(classOverloads) > 0 {
		method, err := e.selectCallableOverload(classInfo.GetName(), methodName, classOverloads, args, ctx)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		classValAny, err2 := e.typeSystem.CreateClassValue(classInfo.GetName())
		if err2 != nil {
			return e.newError(node, "failed to get class value: %s", err2.Error())
		}
		if cm, ok := classValAny.(ClassMetaValue); ok {
			return e.executeClassMethodDirect(cm, method, args, node, ctx)
		}
		return e.newError(node, "internal error: class meta value not available")
	}

	// Try single class method
	if classMethod := classInfo.LookupClassMethod(methodName); classMethod != nil {
		classValAny, err2 := e.typeSystem.CreateClassValue(classInfo.GetName())
		if err2 != nil {
			return e.newError(node, "failed to get class value: %s", err2.Error())
		}
		if cm, ok := classValAny.(ClassMetaValue); ok {
			return e.executeClassMethodDirect(cm, classMethod, args, node, ctx)
		}
	}

	return e.newError(node, "method '%s' not found in class '%s'", methodName, classInfo.GetName())
}

// runtimeValueType converts a runtime Value to a types.Type for overload resolution.
func (e *Evaluator) runtimeValueType(val Value) types.Type {
	switch v := val.(type) {
	case *runtime.IntegerValue:
		return types.INTEGER
	case *runtime.FloatValue:
		return types.FLOAT
	case *runtime.StringValue:
		return types.STRING
	case *runtime.BooleanValue:
		return types.BOOLEAN
	case *runtime.NilValue:
		// Typed nil (e.g. a declared "var o: TObject") matches its static
		// class exactly, so F(o) prefers F(x: TObject) over array overloads.
		if v.ClassType != "" {
			return e.buildClassTypeWithHierarchy(v.ClassType)
		}
		return types.NIL
	case *runtime.VariantValue:
		return types.VARIANT
	case *runtime.EnumValue:
		return types.INTEGER // Enums are ordinal / integer-compatible
	case *runtime.ArrayValue:
		if v.ArrayType != nil {
			return v.ArrayType
		}
		return types.NIL
	case *runtime.ObjectInstance:
		if v.Class != nil {
			// Build the class type including its parent chain so overload
			// resolution can rank subclass -> base-class conversions.
			return e.buildClassTypeWithHierarchy(v.Class.GetName())
		}
		return types.NIL
	case *runtime.RecordValue:
		if v.RecordType != nil {
			return v.RecordType
		}
		return types.NIL
	default:
		// Metaclass references (TClass values) participate in overload
		// resolution as "class of <name>".
		if cm, ok := val.(ClassMetaValue); ok {
			classType := e.buildClassTypeWithHierarchy(cm.GetClassName())
			return types.NewClassOfType(classType)
		}
		return types.NIL
	}
}

// buildClassTypeWithHierarchy builds a types.ClassType including its parent
// chain from the runtime type system, so class-distance checks work.
func (e *Evaluator) buildClassTypeWithHierarchy(className string) *types.ClassType {
	if e.typeSystem == nil {
		return types.NewClassType(className, nil)
	}
	classInfo, _ := e.typeSystem.LookupClass(className).(runtime.IClassInfo)
	if classInfo == nil {
		return types.NewClassType(className, nil)
	}
	if canonical := classInfo.GetClassType(); canonical != nil {
		return canonical
	}
	var parent *types.ClassType
	if p := classInfo.GetParent(); p != nil {
		parent = e.buildClassTypeWithHierarchy(p.GetName())
	}
	return types.NewClassType(classInfo.GetName(), parent)
}

// extractMethodType extracts a types.FunctionType from an *ast.FunctionDecl.
// Returns nil if the type cannot be determined.
func (e *Evaluator) extractMethodType(method *ast.FunctionDecl, ctx *ExecutionContext) *types.FunctionType {
	paramTypes := make([]types.Type, len(method.Parameters))
	paramNames := make([]string, len(method.Parameters))
	lazyParams := make([]bool, len(method.Parameters))
	varParams := make([]bool, len(method.Parameters))
	constParams := make([]bool, len(method.Parameters))
	defaultValues := make([]interface{}, len(method.Parameters))

	for idx, param := range method.Parameters {
		if param.Type == nil {
			return nil
		}
		pt, err := e.ResolveTypeFromAnnotation(param.Type, ctx)
		if err != nil {
			return nil
		}
		paramTypes[idx] = pt
		paramNames[idx] = param.Name.Value
		lazyParams[idx] = param.IsLazy
		varParams[idx] = param.ByRef
		constParams[idx] = param.IsConst
		defaultValues[idx] = param.DefaultValue
	}

	var returnType types.Type = types.VOID
	if method.ReturnType != nil {
		rt, err := e.ResolveTypeFromAnnotation(method.ReturnType, ctx)
		if err == nil {
			returnType = rt
		}
	}

	return types.NewFunctionTypeWithMetadata(
		paramTypes, paramNames, defaultValues,
		lazyParams, varParams, constParams,
		returnType,
	)
}

// selectOverload picks the best matching overload for the given arguments.
// Uses shared overload resolution with type-based matching, falling back to arg-count.
func (e *Evaluator) selectOverload(className, methodName string, overloads []*ast.FunctionDecl, args []Value, ctx *ExecutionContext) (*ast.FunctionDecl, error) {
	if len(overloads) == 1 {
		return overloads[0], nil
	}

	// Build argument types
	argTypes := make([]types.Type, len(args))
	for i, arg := range args {
		argTypes[i] = e.runtimeValueType(arg)
	}

	// Build candidates for type-based resolution
	candidates := make([]types.Type, 0, len(overloads))
	candidateDecls := make([]*ast.FunctionDecl, 0, len(overloads))
	for _, method := range overloads {
		methodType := e.extractMethodType(method, ctx)
		if methodType == nil {
			continue
		}
		candidates = append(candidates, methodType)
		candidateDecls = append(candidateDecls, method)
	}

	if len(candidates) > 0 {
		selected, err := types.ResolveOverload(candidates, argTypes)
		if err == nil {
			return candidateDecls[selected], nil
		}
	}

	// Fall back to arg-count match
	for _, candidate := range overloads {
		if len(candidate.Parameters) == len(args) {
			return candidate, nil
		}
	}

	// Check with default parameters
	for _, candidate := range overloads {
		if len(args) <= len(candidate.Parameters) {
			required := 0
			for _, p := range candidate.Parameters {
				if p.DefaultValue == nil {
					required++
				}
			}
			if len(args) >= required {
				return candidate, nil
			}
		}
	}

	// Last resort: return first overload
	return overloads[0], nil
}

// ============================================================================
// Class Method Overload Dispatch (for CLASS type)
// ============================================================================

// dispatchClassMethodOverloaded dispatches an overloaded class method.
// Self-contained: replaces e.oopEngine.CallMethod for CLASS overloads.
func (e *Evaluator) dispatchClassMethodOverloaded(classMeta ClassMetaValue, classInfo runtime.IClassInfo, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	overloads := classInfo.GetClassMethodOverloads(methodName)
	if len(overloads) == 0 {
		return e.newError(node, "class method '%s' not found in '%s'", methodName, classInfo.GetName())
	}
	method, err := e.selectCallableOverload(classInfo.GetName(), methodName, overloads, args, ctx)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}
	return e.executeClassMethodDirect(classMeta, method, args, node, ctx)
}

// ============================================================================
// Operator Overloading
// ============================================================================

// evalTryBinaryOperator attempts to find and invoke a binary operator overload.
// Self-contained: replaces e.oopEngine.TryBinaryOperator.
func (e *Evaluator) evalTryBinaryOperator(operator string, left, right Value, leftExpr, rightExpr ast.Expression, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if e.typeSystem == nil {
		return nil, false
	}
	operands := []Value{left, right}
	operandExprs := []ast.Expression{leftExpr, rightExpr}

	// Check left operand's class operators (with inheritance fallback)
	if classInfo, ok := e.operandClassInfo(left, leftExpr); ok {
		if result, found := e.lookupClassOperator(operator, classInfo, operands, operandExprs, node, ctx); found {
			return result, true
		}
	}
	// Check right operand's class operators (with inheritance fallback)
	if classInfo, ok := e.operandClassInfo(right, rightExpr); ok {
		if result, found := e.lookupClassOperator(operator, classInfo, operands, operandExprs, node, ctx); found {
			return result, true
		}
	}
	// Check global operator registry (with inheritance-compatible type keys)
	if result, found := e.lookupGlobalOperator(operator, operands, operandExprs, node, ctx); found {
		return result, true
	}
	return nil, false
}

// operandOperatorType resolves the type key used for operator lookup. Runtime
// values normally carry their own language type, but an unassigned class or
// interface reference evaluates to nil and would otherwise erase the operand's
// identity. In that case the analyzer's resolved static type is used instead,
// so that `operator = (TMy, TMy)` still matches two nil TMy operands.
func (e *Evaluator) operandOperatorType(operand Value, expr ast.Expression) types.Type {
	languageType := runtime.LanguageType(operand)
	if languageType != nil && languageType != types.NIL {
		return languageType
	}
	if expr == nil {
		return languageType
	}
	if resolved := e.resolvedSemanticType(expr); resolved != nil {
		return resolved
	}
	return languageType
}

// operandClassInfo returns the runtime class whose operators should be searched
// for an operand, falling back to the operand's static class type when the
// runtime value is nil.
func (e *Evaluator) operandClassInfo(operand Value, expr ast.Expression) (runtime.IClassInfo, bool) {
	if obj, ok := operand.(*runtime.ObjectInstance); ok {
		return obj.Class, true
	}
	if expr == nil || e.typeSystem == nil {
		return nil, false
	}
	classType, ok := e.operandOperatorType(operand, expr).(*types.ClassType)
	if !ok || classType == nil {
		return nil, false
	}
	classInfo, ok := e.typeSystem.LookupClass(classType.Name).(runtime.IClassInfo)
	if !ok || classInfo == nil {
		return nil, false
	}
	return classInfo, true
}

// operatorOperandTypes builds the operand type keys for an operator lookup.
// operandExprs may be nil, or hold nil entries, when no source expression is
// available (compound assignment); those operands keep their runtime type.
func (e *Evaluator) operatorOperandTypes(operands []Value, operandExprs []ast.Expression) []types.Type {
	operandTypes := make([]types.Type, len(operands))
	for i, operand := range operands {
		var expr ast.Expression
		if i < len(operandExprs) {
			expr = operandExprs[i]
		}
		operandTypes[i] = e.operandOperatorType(operand, expr)
	}
	return operandTypes
}

// lookupClassOperator searches the class hierarchy using resolved operand types.
func (e *Evaluator) lookupClassOperator(operator string, classInfo runtime.IClassInfo, operands []Value, operandExprs []ast.Expression, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if classInfo == nil {
		return nil, false
	}
	if entry, found := classInfo.LookupOperator(operator, e.operatorOperandTypes(operands, operandExprs)); found {
		return e.invokeRuntimeOperatorEntry(entry, operands, node, ctx), true
	}
	return nil, false
}

// lookupGlobalOperator searches exact signatures before assignment-compatible signatures.
func (e *Evaluator) lookupGlobalOperator(operator string, operands []Value, operandExprs []ast.Expression, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if entry, found := e.typeSystem.Operators().Lookup(operator, e.operatorOperandTypes(operands, operandExprs)); found {
		return e.invokeGlobalOperatorEntry(entry, operands, node, ctx), true
	}
	return nil, false
}

// evalTryUnaryOperator invokes a matching class or global unary operator.
func (e *Evaluator) evalTryUnaryOperator(operator string, operand Value, operandExpr ast.Expression, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if e.typeSystem == nil {
		return nil, false
	}
	operands := []Value{operand}
	operandExprs := []ast.Expression{operandExpr}
	if classInfo, ok := e.operandClassInfo(operand, operandExpr); ok {
		if result, found := e.lookupClassOperator(operator, classInfo, operands, operandExprs, node, ctx); found {
			return result, true
		}
	}
	return e.lookupGlobalOperator(operator, operands, operandExprs, node, ctx)
}

// invokeRuntimeOperatorEntry invokes a runtime.OperatorEntry (from IClassInfo.LookupOperator).
func (e *Evaluator) invokeRuntimeOperatorEntry(entry *runtime.OperatorEntry, operands []Value, node ast.Node, ctx *ExecutionContext) Value {
	if entry == nil {
		return e.newError(node, "operator not registered")
	}

	if entry.Class != nil {
		if entry.IsClassMethod {
			classValAny, err := e.typeSystem.CreateClassValue(entry.Class.GetName())
			if err != nil {
				return e.newError(node, "operator class '%s' not found", entry.Class.GetName())
			}
			method := entry.Class.LookupClassMethod(entry.BindingName)
			if method == nil {
				return e.newError(node, "class operator method '%s' not found", entry.BindingName)
			}
			if cm, ok := classValAny.(ClassMetaValue); ok {
				return e.executeClassMethodDirect(cm, method, operands, node, ctx)
			}
			return e.newError(node, "internal error: class meta value unavailable")
		}

		if entry.SelfIndex < 0 || entry.SelfIndex >= len(operands) {
			return e.newError(node, "invalid operator configuration for '%s'", entry.Operator)
		}
		selfVal := operands[entry.SelfIndex]
		obj, ok := selfVal.(*runtime.ObjectInstance)
		if !ok {
			return e.newError(node, "operator '%s' requires object operand", entry.Operator)
		}

		args := make([]Value, 0, len(operands)-1)
		for i, v := range operands {
			if i != entry.SelfIndex {
				args = append(args, v)
			}
		}

		method := entry.Class.LookupMethod(entry.BindingName)
		if method == nil {
			return e.newError(node, "operator method '%s' not found", entry.BindingName)
		}
		result := e.executeObjectMethodDirect(obj, method, args, node, ctx)
		// For procedures (no return type), return self so compound assignment
		// like 't += x' doesn't overwrite t with nil.
		if method.IsProcedure() {
			return selfVal
		}
		return result
	}

	// Global operator — no class
	return e.invokeGlobalOperatorByBindingName(entry.BindingName, operands, node, ctx)
}

// invokeGlobalOperatorEntry invokes a global (non-class) operator from the TypeSystem registry.
func (e *Evaluator) invokeGlobalOperatorEntry(entry *interptypes.OperatorEntry, operands []Value, node ast.Node, ctx *ExecutionContext) Value {
	return e.invokeGlobalOperatorByBindingName(entry.BindingName, operands, node, ctx)
}

// invokeGlobalOperatorByBindingName invokes a global function by its binding name.
func (e *Evaluator) invokeGlobalOperatorByBindingName(bindingName string, operands []Value, node ast.Node, ctx *ExecutionContext) Value {
	if e.typeSystem == nil {
		return e.newError(node, "type system not initialized")
	}
	normalizedName := ident.Normalize(bindingName)
	overloads := e.typeSystem.LookupFunctions(normalizedName)
	if len(overloads) == 0 {
		return e.newError(node, "operator binding '%s' not found", bindingName)
	}
	return e.ExecuteUserFunctionDirect(overloads[0], operands, ctx)
}

// ============================================================================
// External Function Dispatch
// ============================================================================

// callExternalFunctionViaEngineState dispatches an external function through
// the value-level host-function callback wired in EngineState.
func (e *Evaluator) callExternalFunctionViaEngineState(funcName string, args []Value, node ast.Node) Value {
	if e.engineState == nil || e.engineState.ExternalFunctionCaller == nil {
		return e.newError(node, "external function '%s' not available (no caller registered)", funcName)
	}
	return e.engineState.ExternalFunctionCaller(funcName, args)
}
