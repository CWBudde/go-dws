package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Phase 3.5.4 - Phase 2B: Type system access adapter methods
// These methods implement the InterpreterAdapter interface for type system access.

// ===== Class Registry =====

// LookupClass finds a class by name in the class registry.
// Task 3.5.46: Delegates to TypeSystem instead of using legacy map.
func (i *Interpreter) LookupClass(name string) (any, bool) {
	class := i.typeSystem.LookupClass(name)
	if class == nil {
		return nil, false
	}
	return class, true
}

// LookupRecord finds a record type by name in the record registry.
// Task 3.5.46: Delegates to TypeSystem instead of using legacy map.
func (i *Interpreter) LookupRecord(name string) (any, bool) {
	normalizedName := ident.Normalize(name)
	record, ok := i.records[normalizedName]
	if !ok {
		return nil, false
	}
	return record, true
}

// LookupInterface finds an interface by name in the interface registry.
// Task 3.5.184: Delegates to TypeSystem instead of using legacy map.
func (i *Interpreter) LookupInterface(name string) (any, bool) {
	iface := i.typeSystem.LookupInterface(name)
	if iface == nil {
		return nil, false
	}
	return iface, true
}

// lookupInterfaceInfo finds an interface by name and returns the typed *InterfaceInfo.
// Task 3.5.184a: Type-safe helper for internal interpreter use.
// Returns nil if the interface is not found or if the type assertion fails.
func (i *Interpreter) lookupInterfaceInfo(name string) *InterfaceInfo {
	return i.LookupInterfaceInfo(name)
}

// LookupInterfaceInfo finds an interface by name and returns the typed *InterfaceInfo.
// Task 3.5.184c: Public API for tests and external code to look up registered interfaces.
// Returns nil if the interface is not found or if the type assertion fails.
func (i *Interpreter) LookupInterfaceInfo(name string) *InterfaceInfo {
	iface := i.typeSystem.LookupInterface(name)
	if iface == nil {
		return nil
	}
	// Type assert from any to *InterfaceInfo
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo
	}
	return nil
}

// ===== Task 3.5.9: Interface Declaration Adapter Methods =====

// NewInterfaceInfoAdapter creates a new InterfaceInfo instance via NewInterfaceInfo function.
// Task 3.5.9.1: Allows evaluator to create InterfaceInfo without direct access to interp package.
func (i *Interpreter) NewInterfaceInfoAdapter(name string) interface{} {
	return NewInterfaceInfo(name)
}

// CastToInterfaceInfo performs type assertion from any to *InterfaceInfo.
// Task 3.5.9.2: Safe type casting for evaluator.
func (i *Interpreter) CastToInterfaceInfo(iface interface{}) (interface{}, bool) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo, true
	}
	return nil, false
}

// SetInterfaceParent sets the parent interface for inheritance.
// Task 3.5.9.2: Allows evaluator to set up interface hierarchy.
func (i *Interpreter) SetInterfaceParent(iface interface{}, parent interface{}) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		if parentInfo, ok := parent.(*InterfaceInfo); ok {
			ifaceInfo.Parent = parentInfo
		}
	}
}

// GetInterfaceName returns the name of an interface.
// Task 3.5.9: Allows evaluator to access interface name without direct InterfaceInfo access.
func (i *Interpreter) GetInterfaceName(iface interface{}) string {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo.Name
	}
	return ""
}

// GetInterfaceParent returns the parent interface.
// Task 3.5.9: Allows evaluator to traverse interface hierarchy for circular inheritance check.
func (i *Interpreter) GetInterfaceParent(iface interface{}) interface{} {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		return ifaceInfo.Parent
	}
	return nil
}

// AddInterfaceMethod adds a method to an interface.
// Task 3.5.9.1: Allows evaluator to register interface methods.
func (i *Interpreter) AddInterfaceMethod(iface interface{}, normalizedName string, method *ast.FunctionDecl) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		ifaceInfo.Methods[normalizedName] = method
	}
}

// AddInterfaceProperty adds a property to an interface.
// Task 3.5.9.4: Allows evaluator to register interface properties.
func (i *Interpreter) AddInterfaceProperty(iface interface{}, normalizedName string, propInfo any) {
	if ifaceInfo, ok := iface.(*InterfaceInfo); ok {
		if prop, ok := propInfo.(*types.PropertyInfo); ok {
			ifaceInfo.Properties[normalizedName] = prop
		}
	}
}

// ResolveClassInfoByName resolves a class by name for property type resolution.
// Task 3.5.9.4: Allows evaluator to resolve class types in property declarations.
func (i *Interpreter) ResolveClassInfoByName(name string) interface{} {
	return i.resolveClassInfoByName(name)
}

// GetClassNameFromInfo returns the name from a raw ClassInfo interface{}.
// Task 3.5.9.4: Extracts class name for type construction.
func (i *Interpreter) GetClassNameFromInfo(classInfo interface{}) string {
	if ci, ok := classInfo.(*ClassInfo); ok {
		return ci.Name
	}
	return ""
}

// LookupHelpers returns helpers for a given type name.
func (i *Interpreter) LookupHelpers(typeName string) []any {
	normalizedName := ident.Normalize(typeName)
	helpers, ok := i.helpers[normalizedName]
	if !ok {
		return nil
	}
	// Convert []*HelperInfo to []any
	result := make([]any, len(helpers))
	for idx, helper := range helpers {
		result[idx] = helper
	}
	return result
}

// ===== Task 3.5.12: Helper Declaration Adapter Methods =====

// CreateHelperInfo creates a new HelperInfo instance.
func (i *Interpreter) CreateHelperInfo(name string, targetType any, isRecordHelper bool) interface{} {
	if tt, ok := targetType.(types.Type); ok {
		return NewHelperInfo(name, tt, isRecordHelper)
	}
	return nil
}

// SetHelperParent sets the parent helper for inheritance chain.
func (i *Interpreter) SetHelperParent(helper interface{}, parent interface{}) {
	if h, ok := helper.(*HelperInfo); ok {
		if p, ok := parent.(*HelperInfo); ok {
			h.ParentHelper = p
		}
	}
}

// VerifyHelperTargetTypeMatch checks if parent helper's target type matches the given type.
func (i *Interpreter) VerifyHelperTargetTypeMatch(parent interface{}, targetType any) bool {
	if p, ok := parent.(*HelperInfo); ok {
		if tt, ok := targetType.(types.Type); ok {
			return p.TargetType.Equals(tt)
		}
	}
	return false
}

// GetHelperName returns the name of a helper (for parent lookup by name).
func (i *Interpreter) GetHelperName(helper interface{}) string {
	if h, ok := helper.(*HelperInfo); ok {
		return h.Name
	}
	return ""
}

// AddHelperMethod registers a method in the helper.
func (i *Interpreter) AddHelperMethod(helper interface{}, normalizedName string, method *ast.FunctionDecl) {
	if h, ok := helper.(*HelperInfo); ok {
		h.Methods[normalizedName] = method
	}
}

// AddHelperProperty registers a property in the helper.
func (i *Interpreter) AddHelperProperty(helper interface{}, prop *ast.PropertyDecl, propType any) {
	if h, ok := helper.(*HelperInfo); ok {
		pt, _ := propType.(types.Type)
		propInfo := &types.PropertyInfo{
			Name: prop.Name.Value,
			Type: pt,
		}
		// Set up property access
		if prop.ReadSpec != nil {
			if identExpr, ok := prop.ReadSpec.(*ast.Identifier); ok {
				propInfo.ReadKind = types.PropAccessMethod
				propInfo.ReadSpec = identExpr.Value
			}
		}
		if prop.WriteSpec != nil {
			if identExpr, ok := prop.WriteSpec.(*ast.Identifier); ok {
				propInfo.WriteKind = types.PropAccessMethod
				propInfo.WriteSpec = identExpr.Value
			}
		}
		h.Properties[prop.Name.Value] = propInfo
	}
}

// AddHelperClassVar adds a class variable to the helper.
func (i *Interpreter) AddHelperClassVar(helper interface{}, name string, value evaluator.Value) {
	if h, ok := helper.(*HelperInfo); ok {
		h.ClassVars[name] = value.(Value)
	}
}

// AddHelperClassConst adds a class constant to the helper.
func (i *Interpreter) AddHelperClassConst(helper interface{}, name string, value evaluator.Value) {
	if h, ok := helper.(*HelperInfo); ok {
		h.ClassConsts[name] = value.(Value)
	}
}

// RegisterHelperLegacy registers the helper in the legacy i.helpers map.
func (i *Interpreter) RegisterHelperLegacy(typeName string, helper interface{}) {
	if i.helpers == nil {
		i.helpers = make(map[string][]*HelperInfo)
	}
	if h, ok := helper.(*HelperInfo); ok {
		i.helpers[typeName] = append(i.helpers[typeName], h)
	}
}

// GetOperatorRegistry returns the operator registry for custom operator lookups.
func (i *Interpreter) GetOperatorRegistry() any {
	return i.globalOperators
}

// GetEnumTypeID returns the type ID for a named enum type.
func (i *Interpreter) GetEnumTypeID(enumName string) int {
	normalizedName := ident.Normalize(enumName)
	typeID, ok := i.enumTypeIDRegistry[normalizedName]
	if !ok {
		return 0
	}
	return typeID
}

// ===== Task 3.5.5: Type System Adapter Method Implementations =====

// Task 3.5.141: GetType removed - evaluator uses resolveTypeName() directly

// Task 3.5.139h: ParseInlineArrayType removed - evaluator uses parseInlineArrayType() directly

// Task 3.5.138: LookupSubrangeType removed - evaluator now uses ctx.Env().Get() directly

// TryImplicitConversion attempts an implicit type conversion.
func (i *Interpreter) TryImplicitConversion(value evaluator.Value, targetTypeName string) (evaluator.Value, bool) {
	converted, ok := i.tryImplicitConversion(value.(Value), targetTypeName)
	if ok {
		return converted, true
	}
	return value, false
}

// WrapInSubrange wraps an integer value in a subrange type with validation.
// Task 3.5.182: Updated to use TypeSystem instead of environment lookup.
func (i *Interpreter) WrapInSubrange(value evaluator.Value, subrangeTypeName string, node ast.Node) (evaluator.Value, error) {
	// Task 3.5.182: Use TypeSystem for subrange type lookup
	subrangeType := i.typeSystem.LookupSubrangeType(subrangeTypeName)
	if subrangeType == nil {
		return nil, fmt.Errorf("subrange type '%s' not found", subrangeTypeName)
	}

	// Extract integer value
	var intValue int
	if intVal, ok := value.(*IntegerValue); ok {
		intValue = int(intVal.Value)
	} else if srcSubrange, ok := value.(*SubrangeValue); ok {
		intValue = srcSubrange.Value
	} else {
		return nil, fmt.Errorf("cannot convert %s to subrange type %s", value.Type(), subrangeTypeName)
	}

	// Create subrange value and validate
	subrangeVal := &SubrangeValue{
		Value:        0, // Will be set by ValidateAndSet
		SubrangeType: subrangeType,
	}
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return nil, err
	}
	return subrangeVal, nil
}

// WrapInInterface wraps an object value in an interface instance.
// Task 3.5.184: Use TypeSystem lookup instead of i.interfaces map.
func (i *Interpreter) WrapInInterface(value evaluator.Value, interfaceName string, node ast.Node) (evaluator.Value, error) {
	ifaceInfo := i.lookupInterfaceInfo(interfaceName)
	if ifaceInfo == nil {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Check if the value is already an InterfaceInstance
	if _, alreadyInterface := value.(*InterfaceInstance); alreadyInterface {
		return value, nil
	}

	// Check if the value is an ObjectInstance
	objInst, isObj := value.(*ObjectInstance)
	if !isObj {
		return nil, fmt.Errorf("cannot wrap %s in interface %s", value.Type(), interfaceName)
	}

	// Validate that the object's class implements the interface
	if !classImplementsInterface(objInst.Class, ifaceInfo) {
		return nil, fmt.Errorf("class '%s' does not implement interface '%s'",
			objInst.Class.Name, ifaceInfo.Name)
	}

	// Wrap the object in an InterfaceInstance
	return NewInterfaceInstance(ifaceInfo, objInst), nil
}

// Task 3.5.140: EvalArrayLiteralWithExpectedType removed - evaluator uses evalArrayLiteralWithExpectedType() directly

// CallIndexedPropertyGetter calls an indexed property getter method on an object.
// Task 3.5.99c: Implements InterpreterAdapter.CallIndexedPropertyGetter for object default property access.
// DEPRECATED: Use ObjectValue.ReadIndexedProperty with ExecuteIndexedPropertyRead callback instead.
func (i *Interpreter) CallIndexedPropertyGetter(obj evaluator.Value, propImpl any, indices []evaluator.Value, node any) evaluator.Value {
	// Convert obj to ObjectInstance
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return &ErrorValue{Message: "CallIndexedPropertyGetter expects ObjectInstance"}
	}

	// Convert propImpl to *types.PropertyInfo
	propInfo, ok := propImpl.(*types.PropertyInfo)
	if !ok {
		return &ErrorValue{Message: "CallIndexedPropertyGetter expects *types.PropertyInfo"}
	}

	// Convert node to ast.Node
	astNode, ok := node.(ast.Node)
	if !ok {
		return &ErrorValue{Message: "CallIndexedPropertyGetter expects ast.Node"}
	}

	// Convert []evaluator.Value to []Value (they're the same underlying interface)
	// evaluator.Value is an alias for the local Value interface in the interp package
	convertedIndices := make([]Value, len(indices))
	for idx, indexVal := range indices {
		convertedIndices[idx] = indexVal
	}

	// Delegate to the existing evalIndexedPropertyRead method
	return i.evalIndexedPropertyRead(objInst, propInfo, convertedIndices, astNode)
}

// ExecuteIndexedPropertyRead executes an indexed property read with resolved PropertyInfo.
// Task 3.5.117: Low-level execution callback for ObjectValue.ReadIndexedProperty().
func (i *Interpreter) ExecuteIndexedPropertyRead(obj evaluator.Value, propInfo any, indices []evaluator.Value, node any) evaluator.Value {
	// Convert obj to ObjectInstance
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return &ErrorValue{Message: "ExecuteIndexedPropertyRead expects ObjectInstance"}
	}

	// Convert propInfo to *types.PropertyInfo
	pInfo, ok := propInfo.(*types.PropertyInfo)
	if !ok {
		return &ErrorValue{Message: "ExecuteIndexedPropertyRead expects *types.PropertyInfo"}
	}

	// Convert []evaluator.Value to []Value
	convertedIndices := make([]Value, len(indices))
	for idx, indexVal := range indices {
		convertedIndices[idx] = indexVal
	}

	// Convert node to ast.Node (optional - for error reporting)
	astNode, _ := node.(ast.Node)

	// Delegate to the existing evalIndexedPropertyRead method
	return i.evalIndexedPropertyRead(objInst, pInfo, convertedIndices, astNode)
}

// CallRecordPropertyGetter calls a record property getter method.
// Task 3.5.99e: Implements InterpreterAdapter.CallRecordPropertyGetter for record default property access.
func (i *Interpreter) CallRecordPropertyGetter(record evaluator.Value, propImpl any, indices []evaluator.Value, node any) evaluator.Value {
	// Convert record to RecordValue
	recordVal, ok := record.(*RecordValue)
	if !ok {
		return &ErrorValue{Message: "CallRecordPropertyGetter expects RecordValue"}
	}

	// Convert propImpl to *types.RecordPropertyInfo
	propInfo, ok := propImpl.(*types.RecordPropertyInfo)
	if !ok {
		return &ErrorValue{Message: "CallRecordPropertyGetter expects *types.RecordPropertyInfo"}
	}

	// Convert node to ast.Node (specifically *ast.IndexExpression for now)
	indexExpr, ok := node.(*ast.IndexExpression)
	if !ok {
		return &ErrorValue{Message: "CallRecordPropertyGetter expects *ast.IndexExpression"}
	}

	// Check if the property has a read accessor
	if propInfo.ReadField == "" {
		return i.newErrorWithLocation(indexExpr, "default property is write-only")
	}

	// Get the getter method
	// Task 3.5.128b: Use free function instead of method due to type alias
	getterMethod := GetRecordMethod(recordVal, propInfo.ReadField)
	if getterMethod == nil {
		return i.newErrorWithLocation(indexExpr, "default property read accessor '%s' is not a method", propInfo.ReadField)
	}

	// Convert []evaluator.Value to []Value
	convertedIndices := make([]Value, len(indices))
	for idx, val := range indices {
		convertedIndices[idx] = val
	}

	// Create a synthetic method call expression: record.GetterMethod(index)
	// We need to bind the index value(s) in the environment temporarily
	methodCall := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: indexExpr.Token},
		},
		Object: indexExpr.Left,
		Method: &ast.Identifier{
			Value: propInfo.ReadField,
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: indexExpr.Token},
			},
		},
		Arguments: make([]ast.Expression, len(indices)),
	}

	// Create temporary identifiers for each index argument
	for idx := range indices {
		tempVarName := fmt.Sprintf("__temp_default_index_%d__", idx)
		methodCall.Arguments[idx] = &ast.Identifier{
			Value: tempVarName,
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: indexExpr.Token},
			},
		}
		// Bind the index value in the environment
		i.env.Define(tempVarName, convertedIndices[idx])
	}

	// Call the getter method
	return i.evalMethodCall(methodCall)
}

// ExecuteRecordPropertyRead executes a record property getter method.
// Task 3.5.118: Low-level execution callback for RecordInstanceValue.ReadIndexedProperty().
// This delegates to the existing CallRecordPropertyGetter logic.
func (i *Interpreter) ExecuteRecordPropertyRead(record evaluator.Value, propInfo any, indices []evaluator.Value, node any) evaluator.Value {
	// Delegate to existing CallRecordPropertyGetter (reuse implementation)
	return i.CallRecordPropertyGetter(record, propInfo, indices, node)
}
