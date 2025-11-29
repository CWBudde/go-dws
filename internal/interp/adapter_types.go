package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
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
func (i *Interpreter) LookupInterface(name string) (any, bool) {
	normalizedName := ident.Normalize(name)
	iface, ok := i.interfaces[normalizedName]
	if !ok {
		return nil, false
	}
	return iface, true
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

// GetType resolves a type by name using the type system.
func (i *Interpreter) GetType(name string) (any, error) {
	typ, err := i.resolveType(name)
	if err != nil {
		return nil, err
	}
	return typ, nil
}

// ParseInlineArrayType parses inline array type signatures like "array of Integer".
func (i *Interpreter) ParseInlineArrayType(typeName string) (any, error) {
	arrType := i.parseInlineArrayType(typeName)
	if arrType == nil {
		return nil, fmt.Errorf("invalid inline array type: %s", typeName)
	}
	return arrType, nil
}

// LookupSubrangeType finds a subrange type by name.
func (i *Interpreter) LookupSubrangeType(name string) (any, bool) {
	normalizedName := ident.Normalize(name)
	subrangeTypeKey := "__subrange_type_" + normalizedName
	typeVal, ok := i.env.Get(subrangeTypeKey)
	return typeVal, ok
}

// BoxVariant wraps a value in a Variant container.
func (i *Interpreter) BoxVariant(value evaluator.Value) evaluator.Value {
	return boxVariant(value.(Value))
}

// TryImplicitConversion attempts an implicit type conversion.
func (i *Interpreter) TryImplicitConversion(value evaluator.Value, targetTypeName string) (evaluator.Value, bool) {
	converted, ok := i.tryImplicitConversion(value.(Value), targetTypeName)
	if ok {
		return converted, true
	}
	return value, false
}

// WrapInSubrange wraps an integer value in a subrange type with validation.
func (i *Interpreter) WrapInSubrange(value evaluator.Value, subrangeTypeName string, node ast.Node) (evaluator.Value, error) {
	normalizedName := ident.Normalize(subrangeTypeName)
	subrangeTypeKey := "__subrange_type_" + normalizedName
	typeVal, ok := i.env.Get(subrangeTypeKey)
	if !ok {
		return nil, fmt.Errorf("subrange type '%s' not found", subrangeTypeName)
	}

	stv, ok := typeVal.(*SubrangeTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a subrange type", subrangeTypeName)
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
		SubrangeType: stv.SubrangeType,
	}
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return nil, err
	}
	return subrangeVal, nil
}

// WrapInInterface wraps an object value in an interface instance.
func (i *Interpreter) WrapInInterface(value evaluator.Value, interfaceName string, node ast.Node) (evaluator.Value, error) {
	ifaceInfo, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
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

// EvalArrayLiteralWithExpectedType evaluates an array literal with expected type context.
func (i *Interpreter) EvalArrayLiteralWithExpectedType(lit ast.Node, expectedTypeName string) evaluator.Value {
	arrayLit, ok := lit.(*ast.ArrayLiteralExpression)
	if !ok {
		return i.newErrorWithLocation(lit, "expected array literal expression")
	}

	// Resolve expected type
	arrType, errVal := i.arrayTypeByName(expectedTypeName, lit)
	if errVal != nil {
		return errVal
	}

	return i.evalArrayLiteralWithExpected(arrayLit, arrType)
}

// WrapJSONValueInVariant wraps a jsonvalue.Value in a VariantValue containing a JSONValue.
// Task 3.5.99b: Implements InterpreterAdapter.WrapJSONValueInVariant for JSON indexing support.
func (i *Interpreter) WrapJSONValueInVariant(jv any) evaluator.Value {
	// Convert any to *jsonvalue.Value
	jsonVal, ok := jv.(*jsonvalue.Value)
	if !ok && jv != nil {
		// Invalid type passed - return error
		return &ErrorValue{Message: "invalid type passed to WrapJSONValueInVariant: expected *jsonvalue.Value"}
	}

	// Use the existing jsonValueToVariant helper
	return jsonValueToVariant(jsonVal)
}

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
