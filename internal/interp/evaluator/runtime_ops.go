package evaluator

import (
	"fmt"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
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
	ifaceInfo, ok := ifaceInfoAny.(runtime.IInterfaceInfo)
	if !ok {
		return nil, fmt.Errorf("interface '%s' has invalid type", ifaceName)
	}
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
func (t *interp_TypeCastValue) String() string  { return t.Object.String() }

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

	var constructor *ast.FunctionDecl
	if len(overloads) == 1 {
		constructor = overloads[0]
	} else if len(overloads) > 1 {
		// Select by arg count first
		for _, candidate := range overloads {
			if len(candidate.Parameters) == len(args) {
				constructor = candidate
				break
			}
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

	// Try instance method overloads first
	overloads := classInfo.GetMethodOverloads(methodName)
	if len(overloads) > 0 {
		method, err := e.selectOverload(classInfo.GetName(), methodName, overloads, args)
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return e.executeObjectMethodDirect(obj, method, args, node, ctx)
	}

	// Fall back to class method overloads
	classOverloads := classInfo.GetClassMethodOverloads(methodName)
	if len(classOverloads) > 0 {
		method, err := e.selectOverload(classInfo.GetName(), methodName, classOverloads, args)
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
		method, err := e.selectOverload(classInfo.GetName(), methodName, overloads, args)
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
		method, err := e.selectOverload(classInfo.GetName(), methodName, classOverloads, args)
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

// selectOverload picks the best matching overload for the given arguments.
// Uses arg-count matching (fast path). Returns an error if no match found.
func (e *Evaluator) selectOverload(className, methodName string, overloads []*ast.FunctionDecl, args []Value) (*ast.FunctionDecl, error) {
	if len(overloads) == 1 {
		return overloads[0], nil
	}
	// Try exact arg-count match first
	for _, candidate := range overloads {
		if len(candidate.Parameters) == len(args) {
			return candidate, nil
		}
	}
	// Check if any overload accepts default parameters that cover the given count
	for _, candidate := range overloads {
		if len(args) <= len(candidate.Parameters) {
			// Count required params
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
	// Fall back to first overload
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
	method, err := e.selectOverload(classInfo.GetName(), methodName, overloads, args)
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
func (e *Evaluator) evalTryBinaryOperator(operator string, left, right Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	operandTypes := []string{left.Type(), right.Type()}
	operands := []Value{left, right}

	// Check left operand's class operators
	if obj, ok := left.(*runtime.ObjectInstance); ok {
		if entry, found := obj.Class.LookupOperator(operator, operandTypes); found {
			return e.invokeRuntimeOperatorEntry(entry, operands, node, ctx), true
		}
	}
	// Check right operand's class operators
	if obj, ok := right.(*runtime.ObjectInstance); ok {
		if entry, found := obj.Class.LookupOperator(operator, operandTypes); found {
			return e.invokeRuntimeOperatorEntry(entry, operands, node, ctx), true
		}
	}
	// Check global operator registry
	if entry, found := e.typeSystem.Operators().Lookup(operator, operandTypes); found {
		return e.invokeGlobalOperatorEntry(entry, operands, node, ctx), true
	}
	return nil, false
}

// evalTryUnaryOperator attempts to find and invoke a unary operator overload.
// Self-contained: replaces e.oopEngine.TryUnaryOperator.
func (e *Evaluator) evalTryUnaryOperator(operator string, operand Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	operandTypes := []string{operand.Type()}
	operands := []Value{operand}

	if obj, ok := operand.(*runtime.ObjectInstance); ok {
		if entry, found := obj.Class.LookupOperator(operator, operandTypes); found {
			return e.invokeRuntimeOperatorEntry(entry, operands, node, ctx), true
		}
	}
	if entry, found := e.typeSystem.Operators().Lookup(operator, operandTypes); found {
		return e.invokeGlobalOperatorEntry(entry, operands, node, ctx), true
	}
	return nil, false
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
		return e.executeObjectMethodDirect(obj, method, args, node, ctx)
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

// callExternalFunctionViaEngineState dispatches an external function via the
// ExternalFunctionCaller callback wired in EngineState.
// Self-contained: replaces e.oopEngine.CallExternalFunction.
func (e *Evaluator) callExternalFunctionViaEngineState(funcName string, argExprs []ast.Expression, node ast.Node) Value {
	if e.engineState == nil || e.engineState.ExternalFunctionCaller == nil {
		return e.newError(node, "external function '%s' not available (no caller registered)", funcName)
	}
	return e.engineState.ExternalFunctionCaller(funcName, argExprs, node)
}

// ============================================================================
// Class Identifier Lookup Fallback
// ============================================================================

// lookupClassByNameFallback tries to find a class via Self's class type (for nested class access).
// Self-contained: replaces e.oopEngine.LookupClassByName.
func (e *Evaluator) lookupClassByNameFallback(name string, ctx *ExecutionContext) ClassMetaValue {
	if ctx == nil {
		return nil
	}
	// Try Self's class for nested class access
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if objVal, ok := selfRaw.(ObjectValue); ok {
			if classMeta, ok := objVal.GetClassType().(ClassMetaValue); ok {
				if nested := classMeta.GetNestedClass(name); nested != nil {
					if nestedMeta, ok := nested.(ClassMetaValue); ok {
						return nestedMeta
					}
				}
			}
		}
	}
	// Also check the __CurrentClass__ context variable
	if ccRaw, ok := ctx.Env().Get("__CurrentClass__"); ok {
		if cm, ok := ccRaw.(ClassMetaValue); ok {
			if nested := cm.GetNestedClass(name); nested != nil {
				if nestedMeta, ok := nested.(ClassMetaValue); ok {
					return nestedMeta
				}
			}
		}
	}
	return nil
}
