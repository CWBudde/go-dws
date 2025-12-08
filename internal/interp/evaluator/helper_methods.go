package evaluator

import (
	"reflect"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Helper methods are type extensions that add methods to types that don't
// natively have them (e.g., str.ToUpper(), arr.Push(), num.ToString()).

// HelperInfo represents a helper type declaration at runtime.
// Uses wrapper methods returning `any` to avoid circular imports with *interp.HelperInfo.
type HelperInfo interface {
	// GetMethodAny looks up a method by name in this helper's inheritance chain.
	// Returns the method declaration, the helper that owns it (as any), and whether it was found.
	GetMethodAny(name string) (*ast.FunctionDecl, any, bool)

	// GetBuiltinMethodAny looks up a builtin method spec by name in this helper's inheritance chain.
	// Returns the builtin spec, the helper that owns it (as any), and whether it was found.
	GetBuiltinMethodAny(name string) (string, any, bool)

	// GetPropertyAny looks up a property by name in this helper's inheritance chain.
	// Returns the property info (as any), the helper that owns it (as any), and whether it was found.
	GetPropertyAny(name string) (any, any, bool)

	// GetClassVars returns the class variables defined in this helper.
	GetClassVars() map[string]Value

	// GetClassConsts returns the class constants defined in this helper.
	GetClassConsts() map[string]Value

	// GetParentHelperAny returns the parent helper (nil for root helpers).
	GetParentHelperAny() any
}

// HelperMethodResult represents the result of a helper method lookup.
type HelperMethodResult struct {
	// OwnerHelper is the helper that owns the method (after searching inheritance chain)
	OwnerHelper HelperInfo
	// Method is the AST function declaration (nil for builtin-only methods)
	Method *ast.FunctionDecl
	// BuiltinSpec is the builtin method identifier (empty string for AST-only methods)
	BuiltinSpec string
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
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		// TODO: For static arrays, also try dynamic equivalent
		if helpers := e.typeSystem.LookupHelpers("array"); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		return combined

	case EnumAccessor:
		// Try specific enum type first, then generic "enum" helpers
		var combined []HelperInfo
		enumTypeName := val.Type()
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
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		if helpers := e.typeSystem.LookupHelpers("enum"); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		return combined

	case ObjectValue:
		typeName = v.ClassName()
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
	default:
		typeName = v.Type()
	}

	// Look up helpers for this type
	helpers := e.typeSystem.LookupHelpers(typeName)
	if helpers == nil {
		return nil
	}
	return convertToHelperInfoSlice(helpers)
}

// FindHelperMethod searches all applicable helpers for a method with the given name.
// Returns the owning helper, method declaration (if any), and builtin spec identifier.
// Later helpers override earlier ones; AST methods are checked before builtin-only.
func (e *Evaluator) FindHelperMethod(val Value, methodName string) *HelperMethodResult {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil
	}

	// Search in reverse so later helpers override earlier ones
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]

		if method, ownerHelperAny, ok := helper.GetMethodAny(methodName); ok {
			// Resolve owner helper from the returned any type
			var ownerHelper HelperInfo
			if ownerHelperAny != nil {
				if oh, ok := ownerHelperAny.(HelperInfo); ok {
					ownerHelper = oh
				} else {
					ownerHelper = helper
				}
			} else {
				ownerHelper = helper
			}

			// Also check for builtin spec
			if spec, _, ok := ownerHelper.GetBuiltinMethodAny(methodName); ok {
				return &HelperMethodResult{
					OwnerHelper: ownerHelper,
					Method:      method,
					BuiltinSpec: spec,
				}
			}
			return &HelperMethodResult{
				OwnerHelper: ownerHelper,
				Method:      method,
				BuiltinSpec: "",
			}
		}
	}

	// If no declared method, check for builtin-only entries
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if spec, ownerHelperAny, ok := helper.GetBuiltinMethodAny(methodName); ok {
			ownerHelper, _ := ownerHelperAny.(HelperInfo)
			return &HelperMethodResult{
				OwnerHelper: ownerHelper,
				Method:      nil,
				BuiltinSpec: spec,
			}
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
		return e.CallASTHelperMethod(result.OwnerHelper, result.Method, selfValue, args, node, ctx)
	}

	return e.newError(node, "helper method has no implementation")
}

// CallBuiltinHelperMethod executes a builtin helper method.
// Tries type-specific helpers in order; unhandled specs fall through to adapter.
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
	if result := e.evalArrayHelper(spec, selfValue, args, node); result != nil {
		return result
	}
	if result := e.evalEnumHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Unhandled - delegate to adapter
	return e.adapter.EvalNode(node)
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
	v := reflect.ValueOf(helper)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && v.IsNil() {
		return e.newError(node, "helper method not found (nil owner)")
	}

	// Check argument count
	if len(args) != len(method.Parameters) {
		return e.newError(node, "wrong number of arguments for helper method '%s': expected %d, got %d",
			method.Name.Value, len(method.Parameters), len(args))
	}

	// Create method environment (enclosed scope)
	ctx.PushEnv()
	defer ctx.PopEnv()

	// Bind Self to the target value (the value being extended)
	ctx.Env().Define("Self", selfValue)

	// Bind helper class vars and consts from entire inheritance chain.
	// Walk from root parent to current helper so child helpers override parents.
	e.bindHelperChainVarsConsts(helper, ctx)

	// Bind method parameters
	for idx, param := range method.Parameters {
		ctx.Env().Define(param.Name.Value, args[idx])
	}

	// For functions, initialize the Result variable
	if method.ReturnType != nil {
		returnType, err := e.ResolveTypeFromAnnotation(method.ReturnType)
		if err != nil {
			return e.newError(node, "failed to resolve return type: %v", err)
		}
		defaultVal := e.GetDefaultValue(returnType)
		ctx.Env().Define("Result", defaultVal)
		// Also define method name as alias for Result (Pascal convention)
		ctx.Env().Define(method.Name.Value, defaultVal)
	}

	// Execute method body
	result := e.Eval(method.Body, ctx)
	if isError(result) {
		return result
	}

	// Extract return value
	if method.ReturnType != nil {
		return e.extractReturnValue(method.Name.Value, ctx)
	}

	// For procedures, return nil
	return e.nilValue()
}

// bindHelperChainVarsConsts binds class vars and consts from the helper inheritance chain.
// Walks from root to current helper so child definitions override parents.
func (e *Evaluator) bindHelperChainVarsConsts(helper HelperInfo, ctx *ExecutionContext) {
	if helper == nil {
		return
	}

	// Check if helper interface wraps a nil pointer
	v := reflect.ValueOf(helper)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}

	// Build helper chain from root to current
	var helperChain []HelperInfo
	for h := helper; h != nil; {
		hv := reflect.ValueOf(h)
		if hv.Kind() == reflect.Ptr && hv.IsNil() {
			break
		}
		helperChain = append([]HelperInfo{h}, helperChain...)
		parentAny := h.GetParentHelperAny()
		if parentAny == nil {
			break
		}
		parent, ok := parentAny.(HelperInfo)
		if !ok || parent == nil {
			break
		}
		pv := reflect.ValueOf(parent)
		if (pv.Kind() == reflect.Ptr || pv.Kind() == reflect.Interface) && pv.IsNil() {
			break
		}
		h = parent
	}

	// Bind vars and consts (root first, so children override)
	for _, h := range helperChain {
		if h == nil {
			continue
		}
		for name, value := range h.GetClassVars() {
			ctx.Env().Define(name, value)
		}
		for name, value := range h.GetClassConsts() {
			ctx.Env().Define(name, value)
		}
	}
}

// extractReturnValue extracts the return value from a function's environment.
// Checks Result first, then method name alias (Pascal convention).
func (e *Evaluator) extractReturnValue(methodName string, ctx *ExecutionContext) Value {
	// Check Result variable first
	if resultVal, ok := ctx.Env().Get("Result"); ok {
		if val, ok := resultVal.(Value); ok && val.Type() != "NIL" {
			return val
		}
	}

	// Check method name alias
	if methodNameVal, ok := ctx.Env().Get(methodName); ok {
		if val, ok := methodNameVal.(Value); ok && val.Type() != "NIL" {
			return val
		}
	}

	// Fallback to Result even if NIL
	if resultVal, ok := ctx.Env().Get("Result"); ok {
		if val, ok := resultVal.(Value); ok {
			return val
		}
	}

	// Final fallback
	return e.nilValue()
}

// ============================================================================
// Helper Utilities
// ============================================================================

// convertToHelperInfoSlice converts []any from TypeSystem.LookupHelpers to []HelperInfo.
func convertToHelperInfoSlice(helpers []any) []HelperInfo {
	if helpers == nil {
		return nil
	}

	result := make([]HelperInfo, 0, len(helpers))
	for _, h := range helpers {
		if h == nil {
			continue
		}
		if helperInfo, ok := h.(HelperInfo); ok {
			if helperInfo != nil && !reflect.ValueOf(helperInfo).IsNil() {
				result = append(result, helperInfo)
			}
		}
	}
	return result
}

// ============================================================================
// Helper Property Resolution
// ============================================================================

// FindHelperProperty searches all applicable helpers for a property with the given name.
func (e *Evaluator) FindHelperProperty(val Value, propName string) (HelperInfo, *types.PropertyInfo) {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil, nil
	}

	// Search helpers in reverse order so later (user-defined) helpers override earlier ones
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]

		// Use GetPropertyAny which searches the inheritance chain and returns the owner helper
		if propInfo, ownerHelperAny, found := helper.GetPropertyAny(propName); found && propInfo != nil {
			pInfo, ok := propInfo.(*types.PropertyInfo)
			if ok {
				ownerHelper, _ := ownerHelperAny.(HelperInfo)
				return ownerHelper, pInfo
			}
		}
	}

	return nil, nil
}

// executeHelperPropertyRead evaluates a helper property read access.
// Handles PropAccessField, PropAccessMethod, PropAccessBuiltin, and PropAccessNone.
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
		if method, methodOwnerAny, ok := helper.GetMethodAny(normalizedReadSpec); ok {
			methodOwner, _ := methodOwnerAny.(HelperInfo)
			var builtinSpec string
			if methodOwner != nil {
				if spec, _, ok := methodOwner.GetBuiltinMethodAny(normalizedReadSpec); ok {
					builtinSpec = spec
				}
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
		if method, methodOwnerAny, ok := helper.GetMethodAny(normalizedReadSpec); ok {
			methodOwner, _ := methodOwnerAny.(HelperInfo)
			var builtinSpec string
			if methodOwner != nil {
				if spec, _, ok := methodOwner.GetBuiltinMethodAny(normalizedReadSpec); ok {
					builtinSpec = spec
				}
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

	case types.PropAccessBuiltin:
		return e.evalBuiltinHelperProperty(propInfo.ReadSpec, selfValue, node)

	case types.PropAccessNone:
		return e.newError(node, "property '%s' is write-only", propInfo.Name)

	default:
		return e.newError(node, "property '%s' has no read access", propInfo.Name)
	}
}

// evalBuiltinHelperProperty evaluates a built-in helper property (array, enum, string, etc.).
func (e *Evaluator) evalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value {
	switch propSpec {
	case "__array_length", "__array_count", "__array_high", "__array_low":
		if _, ok := selfValue.(ArrayAccessor); !ok {
			return e.newError(node, "built-in property '%s' can only be used on arrays", propSpec)
		}
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)

	case "__enum_value":
		enumVal, ok := selfValue.(EnumAccessor)
		if !ok {
			return e.newError(node, "Enum.Value property requires enum receiver")
		}
		return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}

	case "__enum_name", "__enum_qualifiedname":
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)

	case "__string_length":
		if _, ok := selfValue.(StringValue); !ok {
			return e.newError(node, "String.Length property requires string receiver")
		}
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)

	default:
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)
	}
}
