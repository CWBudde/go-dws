package evaluator

import (
	"reflect"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Helper Method Resolution
// ============================================================================
// Task 3.5.98b: Infrastructure for helper method lookup and resolution.
//
// Helper methods are type extensions that add methods to types that don't
// natively have them (e.g., str.ToUpper(), arr.Push(), num.ToString()).
//
// The helper system involves:
// - Looking up helpers for the value's type (getHelpersForValue)
// - Searching helper inheritance chains
// - Handling both AST-declared methods and builtin spec methods
//
// This file provides the infrastructure for finding helper methods directly
// in the evaluator, without delegating to the adapter.

// HelperInfo represents a helper type declaration at runtime.
// This is a temporary interface to avoid circular imports.
// The actual implementation is *interp.HelperInfo.
//
// Note: The interface uses wrapper method names (GetMethodAny, GetBuiltinMethodAny,
// GetPropertyAny, GetParentHelperAny) that return `any` types. This allows
// *interp.HelperInfo to satisfy the interface without modifying its original methods.
// Go's type system doesn't support covariant return types, so we need these wrappers.
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
	// Task 3.5.102g: Enables direct access to helper class variables.
	GetClassVars() map[string]Value

	// GetClassConsts returns the class constants defined in this helper.
	// Task 3.5.102g: Enables direct access to helper class constants.
	GetClassConsts() map[string]Value

	// GetParentHelperAny returns the parent helper for inheritance chain traversal.
	// Task 3.5.102g: Enables walking the helper inheritance chain.
	// Returns nil if this is a root helper.
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
// This method maps runtime values to their corresponding helper type names,
// then uses TypeSystem.LookupHelpers() to retrieve the helper registry entries.
//
// Task 3.5.98b: Migrated from Interpreter.getHelpersForValue().
func (e *Evaluator) getHelpersForValue(val Value) []HelperInfo {
	if e.typeSystem == nil {
		return nil
	}

	// Get the type name from the value
	var typeName string
	switch v := val.(type) {
	case ArrayAccessor:
		// For arrays, try multiple helper registries:
		// 1. Specific array type (e.g., "array of String")
		// 2. Dynamic array fallback (e.g., "array of <elem>" for static arrays)
		// 3. Generic "array" helpers
		var combined []HelperInfo

		// Get array type string for specific lookup
		arrayTypeStr := v.ArrayTypeString()
		specific := ident.Normalize(arrayTypeStr)
		if helpers := e.typeSystem.LookupHelpers(specific); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}

		// If it's a static array, also try the dynamic equivalent
		// Note: We'd need to check ArrayAccessor.IsStatic() and ElementType()
		// This is kept simple for now, matching the original logic structure

		// Fallback to generic "array" helpers
		if helpers := e.typeSystem.LookupHelpers("array"); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		return combined

	case EnumAccessor:
		// For enums, try both specific enum type and generic "enum" helpers
		var combined []HelperInfo

		// Get enum type name - we need to check if it has an EnumTypeName method
		// For now, we'll try to use the value's type string
		// TODO: Add EnumTypeName() method to EnumAccessor interface
		enumTypeName := val.Type() // This will return "ENUM" for now

		// Try to get the actual enum type name if the value implements the right interface
		// We'll use a type assertion to access the TypeName field if available
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

		// Fallback to generic "enum" helpers
		if helpers := e.typeSystem.LookupHelpers("enum"); helpers != nil {
			combined = append(combined, convertToHelperInfoSlice(helpers)...)
		}
		return combined

	case ObjectValue:
		// For objects, use the class name
		typeName = v.ClassName()
	case RecordInstanceValue:
		// For records, use the record type name
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
		// For other types, try to extract type name from Type() method
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
// Returns the helper that owns the method, the method declaration (if any),
// and the builtin specification identifier.
//
// The search order is:
// 1. User-defined methods (latest helper overrides earlier ones)
// 2. Builtin-only methods (no AST declaration)
//
// Task 3.5.98b: Migrated from Interpreter.findHelperMethod().
func (e *Evaluator) FindHelperMethod(val Value, methodName string) *HelperMethodResult {
	helpers := e.getHelpersForValue(val)
	if helpers == nil {
		return nil
	}

	// Search helpers in reverse order so later (user-defined) helpers override earlier ones.
	// For each helper, search the inheritance chain using GetMethodAny
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]

		// Use GetMethodAny which searches the inheritance chain and returns the owner helper
		if method, ownerHelperAny, ok := helper.GetMethodAny(methodName); ok {
			// The ownerHelperAny should satisfy the HelperInfo interface
			// Try to cast it to HelperInfo interface
			var ownerHelper HelperInfo
			if ownerHelperAny != nil {
				// First try direct interface assertion
				if oh, ok := ownerHelperAny.(HelperInfo); ok {
					ownerHelper = oh
				} else {
					// If that fails, use the current helper (which does implement HelperInfo)
					ownerHelper = helper
				}
			} else {
				ownerHelper = helper
			}

			// Check if there's a builtin spec as well (search from the owner helper)
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

// ============================================================================
// Helper Interfaces for Value Types
// ============================================================================
// These interfaces extend Value types to provide helper-specific information.
// Most of these are already defined in evaluator.go - we reference them here
// for documentation purposes.

// Note: RecordInstanceValue is defined in evaluator.go with GetRecordTypeName()
// Note: EnumAccessor is defined in evaluator.go with GetOrdinal()
// Note: ObjectValue is defined in evaluator.go with ClassName()

// ArrayAccessor is an optional interface for array values.
// Task 3.5.98b: Enables helper method resolution for arrays.
type ArrayAccessor interface {
	Value
	// ArrayTypeString returns the array type as a string (e.g., "array of String").
	ArrayTypeString() string
}

// IntegerValue, FloatValue, StringValue, BooleanValue interfaces
// Task 3.5.98b: These are marker interfaces to enable type switches in getHelpersForValue.
// The actual implementations are in the interp package.

// IntegerValue represents an integer runtime value.
type IntegerValue interface {
	Value
}

// FloatValue represents a float runtime value.
type FloatValue interface {
	Value
}

// StringValue represents a string runtime value.
type StringValue interface {
	Value
}

// BooleanValue represents a boolean runtime value.
type BooleanValue interface {
	Value
}

// ============================================================================
// Helper Method Execution
// ============================================================================
// Task 3.5.98c: Execute helper methods directly in the evaluator.

// CallHelperMethod executes a helper method (builtin or AST) on a value.
// This replaces the adapter's callHelperMethod for cases that can be handled
// directly in the evaluator.
//
// Task 3.5.102g: Added ctx parameter to enable direct AST helper execution.
//
// Returns:
// - The method result value
// - An error value if something went wrong
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
	// Task 3.5.102g: Migrated AST helper method execution from Interpreter
	if result.Method != nil {
		return e.CallASTHelperMethod(result.OwnerHelper, result.Method, selfValue, args, node, ctx)
	}

	return e.newError(node, "helper method has no implementation")
}

// CallBuiltinHelperMethod executes a builtin helper method directly in the evaluator.
// Task 3.5.98c: Migrates builtin helper method execution from Interpreter.
// Task 3.5.102a: String helper methods now handled directly in evaluator.
// Task 3.5.102b: Integer helper methods now handled directly in evaluator.
// Task 3.5.102c: Float helper methods now handled directly in evaluator.
// Task 3.5.102d: Boolean helper methods now handled directly in evaluator.
// Task 3.5.102e: Array helper methods now handled directly in evaluator.
// Task 3.5.102f: Enum helper methods now handled directly in evaluator.
// Task 3.5.102g: Added ctx parameter for consistency with CallHelperMethod.
//
// Specific helper implementations are migrated incrementally:
// - String helpers: ToUpper, ToLower, Length, ToString (Task 3.5.102a)
// - Integer helpers: ToString, ToHexString (Task 3.5.102b)
// - Float helpers: ToString, ToString(precision) (Task 3.5.102c)
// - Boolean helpers: ToString (Task 3.5.102d)
// - Array helpers: Length, Count, High, Low, Add, Push, Pop, Swap, Delete, Join (Task 3.5.102e)
// - Enum helpers: Value, Name, QualifiedName (Task 3.5.102f)
//
// Unhandled helpers fall through to the adapter.
func (e *Evaluator) CallBuiltinHelperMethod(spec string, selfValue Value, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	// Task 3.5.102a: Try string helpers first
	if result := e.evalStringHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Task 3.5.102b: Try integer helpers
	if result := e.evalIntegerHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Task 3.5.102c: Try float helpers
	if result := e.evalFloatHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Task 3.5.102d: Try boolean helpers
	if result := e.evalBooleanHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Task 3.5.102e: Try array helpers
	if result := e.evalArrayHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Task 3.5.102f: Try enum helpers
	if result := e.evalEnumHelper(spec, selfValue, args, node); result != nil {
		return result
	}

	// Fall through to adapter for unhandled helpers
	return e.adapter.EvalNode(node)
}

// CallASTHelperMethod executes a user-defined helper method (with AST body).
// Task 3.5.102g: Migrates AST helper method execution from Interpreter.callHelperMethod.
//
// This handles:
// - Creating a new method environment
// - Binding Self to the target value (the value being extended)
// - Binding helper class vars and consts from inheritance chain
// - Binding method parameters
// - Initializing Result variable
// - Executing method body
// - Extracting return value
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

	// Task 3.8.4: Safety check - helper can be nil if OwnerHelper lookup failed
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
// Walks from root parent to the current helper so child helpers override parents.
// Task 3.5.102g: Helper for CallASTHelperMethod.
func (e *Evaluator) bindHelperChainVarsConsts(helper HelperInfo, ctx *ExecutionContext) {
	// Handle nil helper gracefully
	if helper == nil {
		return
	}

	// Task 3.8.6.3: Check if helper interface wraps a nil pointer
	v := reflect.ValueOf(helper)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}

	// Build the helper chain from root to current
	var helperChain []HelperInfo
	for h := helper; h != nil; {
		// Task 3.8.6.3: Defensive nil pointer check
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
		// Task 3.8.6.3: Check if parent interface wraps a nil pointer
		pv := reflect.ValueOf(parent)
		if (pv.Kind() == reflect.Ptr || pv.Kind() == reflect.Interface) && pv.IsNil() {
			break
		}
		h = parent
	}

	// Bind vars and consts in order (root first, so children override)
	for _, h := range helperChain {
		// Task 3.8.6.3: Defensive nil check for helper chain iteration
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
// Checks Result first, then method name alias, following Pascal conventions.
// Task 3.5.102g: Helper for CallASTHelperMethod.
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

// convertToHelperInfoSlice converts a slice of any (from TypeSystem.LookupHelpers)
// to a slice of HelperInfo interfaces.
//
// TypeSystem.LookupHelpers() returns []any to avoid circular imports, but we know
// each element is a *interp.HelperInfo that satisfies our HelperInfo interface.
func convertToHelperInfoSlice(helpers []any) []HelperInfo {
	if helpers == nil {
		return nil
	}

	result := make([]HelperInfo, 0, len(helpers))
	for _, h := range helpers {
		// Skip nil entries
		if h == nil {
			continue
		}
		// Each helper should be a *interp.HelperInfo which implements our HelperInfo interface
		if helperInfo, ok := h.(HelperInfo); ok {
			// Check if the interface wraps a nil pointer (defensive check for corrupted registry)
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
// Task 3.5.37: Infrastructure for helper property lookup and reading.
//
// Helper properties are type extensions that add properties to types that don't
// natively have them (e.g., arr.Length, enum.Name, str.IsASCII).

// FindHelperProperty searches all applicable helpers for a property with the given name.
// Returns the helper that owns the property and the property info.
//
// Task 3.5.37: Migrated from Interpreter.findHelperProperty().
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
// Task 3.5.37: Migrated from Interpreter.evalHelperPropertyRead().
//
// This method handles four property access kinds:
// - PropAccessField: Direct field access or getter method call
// - PropAccessMethod: Getter method call
// - PropAccessBuiltin: Built-in property (e.g., array.Length)
// - PropAccessNone: Write-only property (error)
func (e *Evaluator) executeHelperPropertyRead(
	helper HelperInfo,
	propInfo *types.PropertyInfo,
	selfValue Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	switch propInfo.ReadKind {
	case types.PropAccessField:
		// For helpers on records, try to access the field from the record
		if recVal, ok := selfValue.(RecordInstanceValue); ok {
			if fieldValue, exists := recVal.GetRecordField(propInfo.ReadSpec); exists {
				return fieldValue
			}
		}

		// Otherwise, try as a method (getter)
		// Method names are case-insensitive
		normalizedReadSpec := ident.Normalize(propInfo.ReadSpec)

		// Search for the getter method in the owner helper's inheritance chain
		if method, methodOwnerAny, ok := helper.GetMethodAny(normalizedReadSpec); ok {
			methodOwner, _ := methodOwnerAny.(HelperInfo)
			// Get builtin spec if any
			var builtinSpec string
			if methodOwner != nil {
				if spec, _, ok := methodOwner.GetBuiltinMethodAny(normalizedReadSpec); ok {
					builtinSpec = spec
				}
			}
			// Call the getter method with no arguments
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
		// Call getter method
		// Method names are case-insensitive
		normalizedReadSpec := ident.Normalize(propInfo.ReadSpec)

		// Search for the getter method in the owner helper's inheritance chain
		if method, methodOwnerAny, ok := helper.GetMethodAny(normalizedReadSpec); ok {
			methodOwner, _ := methodOwnerAny.(HelperInfo)
			// Get builtin spec if any
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
		// Built-in helper properties (e.g., array.Length, enum.Name)
		return e.evalBuiltinHelperProperty(propInfo.ReadSpec, selfValue, node)

	case types.PropAccessNone:
		return e.newError(node, "property '%s' is write-only", propInfo.Name)

	default:
		return e.newError(node, "property '%s' has no read access", propInfo.Name)
	}
}

// evalBuiltinHelperProperty evaluates a built-in helper property.
// Task 3.5.37: Migrated from Interpreter.evalBuiltinHelperProperty().
//
// Implements built-in properties like:
// - Array: .Length, .Count, .High, .Low
// - Enum: .Value, .Name, .QualifiedName
// - String: .Length, .IsASCII, .Trim, etc.
// - Integer/Float/Boolean: .ToString
func (e *Evaluator) evalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value {
	switch propSpec {
	// Array properties - delegate to adapter (requires ArrayValue internals)
	case "__array_length", "__array_count", "__array_high", "__array_low":
		if _, ok := selfValue.(ArrayAccessor); !ok {
			return e.newError(node, "built-in property '%s' can only be used on arrays", propSpec)
		}
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)

	// Enum properties
	case "__enum_value":
		enumVal, ok := selfValue.(EnumAccessor)
		if !ok {
			return e.newError(node, "Enum.Value property requires enum receiver")
		}
		return &runtime.IntegerValue{Value: int64(enumVal.GetOrdinal())}

	case "__enum_name", "__enum_qualifiedname":
		// Need to get enum value/type names - delegate to adapter
		// as EnumAccessor doesn't expose ValueName/TypeName
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)

	// String properties - delegate to adapter (requires StringValue internals)
	case "__string_length":
		if _, ok := selfValue.(StringValue); !ok {
			return e.newError(node, "String.Length property requires string receiver")
		}
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)

	// For other built-in properties, delegate to adapter
	// This includes: __integer_tostring, __float_tostring_default, __boolean_tostring,
	// __string_isascii, __string_trim, __string_trimleft, __string_trimright, StripAccents
	default:
		return e.adapter.EvalBuiltinHelperProperty(propSpec, selfValue, node)
	}
}
