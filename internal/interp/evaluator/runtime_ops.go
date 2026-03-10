package evaluator

import (
	"fmt"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
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
		if v.Class != nil && e.typeSystem != nil {
			// Try to find the class type in the type system
			if classInfoAny := e.typeSystem.LookupClass(v.Class.GetName()); classInfoAny != nil {
				if ct, ok := classInfoAny.(types.Type); ok {
					return ct
				}
			}
			// Return a simple class type by name
			return types.NewClassType(v.Class.GetName(), nil)
		}
		return types.NIL
	case *runtime.RecordValue:
		if v.RecordType != nil {
			return v.RecordType
		}
		return types.NIL
	default:
		return types.NIL
	}
}

// extractMethodType extracts a types.FunctionType from an *ast.FunctionDecl.
// Returns nil if the type cannot be determined.
func (e *Evaluator) extractMethodType(method *ast.FunctionDecl) *types.FunctionType {
	ctx := e.currentContext
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
		pt, err := e.resolveTypeName(param.Type.String(), ctx)
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
		rt, err := e.resolveTypeName(method.ReturnType.String(), ctx)
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
// Uses semantic overload resolution with type-based matching, falling back to arg-count.
func (e *Evaluator) selectOverload(className, methodName string, overloads []*ast.FunctionDecl, args []Value) (*ast.FunctionDecl, error) {
	if len(overloads) == 1 {
		return overloads[0], nil
	}

	// Build argument types
	argTypes := make([]types.Type, len(args))
	for i, arg := range args {
		argTypes[i] = e.runtimeValueType(arg)
	}

	// Build candidates for semantic resolution
	candidates := make([]*semantic.Symbol, 0, len(overloads))
	candidateDecls := make([]*ast.FunctionDecl, 0, len(overloads))
	for _, method := range overloads {
		methodType := e.extractMethodType(method)
		if methodType == nil {
			continue
		}
		candidates = append(candidates, &semantic.Symbol{
			Name:                 method.Name.Value,
			Type:                 methodType,
			HasOverloadDirective: method.IsOverload,
		})
		candidateDecls = append(candidateDecls, method)
	}

	if len(candidates) > 0 {
		selected, err := semantic.ResolveOverload(candidates, argTypes)
		if err == nil {
			selectedType, ok := selected.Type.(*types.FunctionType)
			if ok {
				for i, candidate := range candidates {
					if ct, ok := candidate.Type.(*types.FunctionType); ok {
						if semantic.SignaturesEqual(ct, selectedType) &&
							ct.ReturnType.Equals(selectedType.ReturnType) {
							return candidateDecls[i], nil
						}
					}
				}
			}
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
	method, err := e.selectOverload(classInfo.GetName(), methodName, overloads, args)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}
	return e.executeClassMethodDirect(classMeta, method, args, node, ctx)
}

// ============================================================================
// Operator Overloading
// ============================================================================

// operatorTypeKey converts a runtime value to the normalized type key used in
// the operator registry. Must match the key format produced by the interp package's
// valueTypeKey() function and NormalizeTypeAnnotation().
func operatorTypeKey(val Value) string {
	if val == nil {
		return "nil"
	}
	switch v := val.(type) {
	case *runtime.ObjectInstance:
		if v.Class != nil {
			return "class:" + ident.Normalize(v.Class.GetName())
		}
		return "class:"
	case *runtime.RecordValue:
		// Records use "class:<name>" format to match interp package's valueTypeKey()
		if v.RecordType != nil && v.RecordType.Name != "" {
			return "class:" + ident.Normalize(v.RecordType.Name)
		}
		return "record"
	case *runtime.ArrayValue:
		if v.ArrayType != nil && v.ArrayType.ElementType != nil {
			return "array of " + ident.Normalize(v.ArrayType.ElementType.String())
		}
		return "array"
	default:
		// Normalize: "STRING" -> "string", "INTEGER" -> "integer", etc.
		return ident.Normalize(val.Type())
	}
}

// evalTryBinaryOperator attempts to find and invoke a binary operator overload.
// Self-contained: replaces e.oopEngine.TryBinaryOperator.
func (e *Evaluator) evalTryBinaryOperator(operator string, left, right Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if e.typeSystem == nil {
		return nil, false
	}
	operands := []Value{left, right}

	// Check left operand's class operators (with inheritance fallback)
	if obj, ok := left.(*runtime.ObjectInstance); ok {
		if result, found := e.lookupClassOperator(operator, obj.Class, operands, node, ctx); found {
			return result, true
		}
	}
	// Check right operand's class operators (with inheritance fallback)
	if obj, ok := right.(*runtime.ObjectInstance); ok {
		if result, found := e.lookupClassOperator(operator, obj.Class, operands, node, ctx); found {
			return result, true
		}
	}
	// Check global operator registry (with inheritance-compatible type keys)
	if result, found := e.lookupGlobalOperator(operator, operands, node, ctx); found {
		return result, true
	}
	return nil, false
}

// lookupClassOperator looks up an operator in the class hierarchy, trying parent type keys.
func (e *Evaluator) lookupClassOperator(operator string, classInfo runtime.IClassInfo, operands []Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if classInfo == nil {
		return nil, false
	}
	// Try with actual types first, then walk up both operand class hierarchies
	// Build the type key combinations to try
	leftTypeKeys := classTypeKeyChain(operands[0])
	rightTypeKeys := classTypeKeyChain(operands[1])

	// Try all combinations of type keys (actual types first, then parent types)
	for _, leftKey := range leftTypeKeys {
		for _, rightKey := range rightTypeKeys {
			operandTypes := []string{leftKey, rightKey}
			// Check in the class and its ancestors
			for current := classInfo; current != nil; current = current.GetParent() {
				if entry, found := current.LookupOperator(operator, operandTypes); found {
					return e.invokeRuntimeOperatorEntry(entry, operands, node, ctx), true
				}
			}
		}
	}
	return nil, false
}

// lookupGlobalOperator looks up a global operator with inheritance-compatible type keys.
func (e *Evaluator) lookupGlobalOperator(operator string, operands []Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if ops := e.typeSystem.Operators(); ops != nil {
		leftTypeKeys := classTypeKeyChain(operands[0])
		rightTypeKeys := classTypeKeyChain(operands[1])

		for _, leftKey := range leftTypeKeys {
			for _, rightKey := range rightTypeKeys {
				operandTypes := []string{leftKey, rightKey}
				if entry, found := ops.Lookup(operator, operandTypes); found {
					return e.invokeGlobalOperatorEntry(entry, operands, node, ctx), true
				}
			}
		}
	}
	return nil, false
}

// classTypeKeyChain returns the normalized type key for a value, plus parent class keys.
// For objects: ["class:tchild", "class:tparent", ...up to root]
// For other types: just [normalizedKey]
func classTypeKeyChain(val Value) []string {
	if obj, ok := val.(*runtime.ObjectInstance); ok && obj.Class != nil {
		var keys []string
		for current := obj.Class; current != nil; current = current.GetParent() {
			keys = append(keys, "class:"+ident.Normalize(current.GetName()))
		}
		return keys
	}
	return []string{operatorTypeKey(val)}
}

// evalTryUnaryOperator attempts to find and invoke a unary operator overload.
// Self-contained: replaces e.oopEngine.TryUnaryOperator.
func (e *Evaluator) evalTryUnaryOperator(operator string, operand Value, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	if e.typeSystem == nil {
		return nil, false
	}

	// For object operands, try class-level operator with inheritance
	if obj, ok := operand.(*runtime.ObjectInstance); ok {
		typeKeys := classTypeKeyChain(operand)
		operands := []Value{operand}
		for _, typeKey := range typeKeys {
			operandTypes := []string{typeKey}
			for current := obj.Class; current != nil; current = current.GetParent() {
				if entry, found := current.LookupOperator(operator, operandTypes); found {
					return e.invokeRuntimeOperatorEntry(entry, operands, node, ctx), true
				}
			}
		}
	}

	// Check global operator registry
	operandTypes := []string{operatorTypeKey(operand)}
	operands := []Value{operand}
	if ops := e.typeSystem.Operators(); ops != nil {
		if entry, found := ops.Lookup(operator, operandTypes); found {
			return e.invokeGlobalOperatorEntry(entry, operands, node, ctx), true
		}
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
		result := e.executeObjectMethodDirect(obj, method, args, node, ctx)
		// For procedures (no return type), return self so compound assignment
		// like 't += x' doesn't overwrite t with nil.
		if method.ReturnType == nil {
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
