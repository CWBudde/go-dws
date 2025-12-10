package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// executePropertyWrite handles property setter execution for ObjectValue.WriteProperty callback.
//
// This method handles two property write kinds:
// - PropAccessField: Direct field assignment
// - PropAccessMethod: Setter method call
//
// For method-based setters, we reuse the adapter's ExecuteMethodWithSelf callback
// to avoid duplicating method execution logic.
//
// Parameters:
//   - obj: The object to write the property to (implements ObjectValue)
//   - propInfo: The property metadata (*types.PropertyInfo passed as any)
//   - value: The value to assign to the property
//   - node: AST node for error reporting
//   - ctx: Execution context for environment and call stack
func (e *Evaluator) executePropertyWrite(obj Value, propInfo any, value Value, node ast.Node, ctx *ExecutionContext) Value {
	// Type assert propInfo to *types.PropertyInfo
	pInfo, ok := propInfo.(*types.PropertyInfo)
	if !ok {
		return e.newError(node, "invalid property info type")
	}

	// Get ObjectValue interface for class lookups
	objVal, ok := obj.(ObjectValue)
	if !ok {
		return e.newError(node, "cannot write property to non-object value")
	}

	// Check if property has write access
	if pInfo.WriteKind == types.PropAccessNone {
		return e.newError(node, "property '%s' is read-only", pInfo.Name)
	}

	// Get property context for circular reference tracking
	propCtx := ctx.PropContext()

	// Check for circular property references
	for _, prop := range propCtx.PropertyChain {
		if prop == pInfo.Name {
			return e.newError(node, "circular property reference detected: %s", pInfo.Name)
		}
	}

	// Push property onto chain
	propCtx.PropertyChain = append(propCtx.PropertyChain, pInfo.Name)
	defer func() {
		// Pop property from chain when done
		if len(propCtx.PropertyChain) > 0 {
			propCtx.PropertyChain = propCtx.PropertyChain[:len(propCtx.PropertyChain)-1]
		}
	}()

	switch pInfo.WriteKind {
	case types.PropAccessField:
		return e.executeFieldBackedPropertyWrite(obj, objVal, pInfo, value, node, ctx)

	case types.PropAccessMethod:
		return e.executeMethodBackedPropertyWrite(obj, objVal, pInfo, value, node, ctx)

	default:
		return e.newError(node, "property '%s' has unsupported write access kind", pInfo.Name)
	}
}

// executeFieldBackedPropertyWrite handles PropAccessField property writes.
// Writes directly to the field specified by WriteSpec.
func (e *Evaluator) executeFieldBackedPropertyWrite(obj Value, _ ObjectValue, pInfo *types.PropertyInfo, value Value, node ast.Node, _ *ExecutionContext) Value {
	fieldName := pInfo.WriteSpec

	// Field-backed properties with index directive are not allowed for write
	if pInfo.HasIndexValue {
		return e.newError(node, "property '%s' uses an index directive and cannot be field-backed for write", pInfo.Name)
	}

	// Try to set the field via ObjectFieldSetter interface
	if setter, ok := obj.(ObjectFieldSetter); ok {
		setter.SetField(fieldName, value)
		return value
	}

	return e.newError(node, "object does not support field assignment")
}

// executeMethodBackedPropertyWrite handles PropAccessMethod property writes.
func (e *Evaluator) executeMethodBackedPropertyWrite(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, value Value, node ast.Node, ctx *ExecutionContext) Value {
	// Indexed properties must be accessed with index syntax
	if pInfo.IsIndexed {
		return e.newError(node, "indexed property '%s' requires index arguments for write", pInfo.Name)
	}

	return e.executePropertySetterMethod(obj, objVal, pInfo, value, node, ctx)
}

// executePropertySetterMethod executes a property setter method.
// Used by PropAccessMethod for property writes.
//
// This reuses the adapter's ExecuteMethodWithSelf to avoid duplicating method execution logic.
func (e *Evaluator) executePropertySetterMethod(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, value Value, node ast.Node, ctx *ExecutionContext) Value {
	// Look up the setter method via ObjectValue interface
	methodName := pInfo.WriteSpec

	// Check if method exists
	if !objVal.HasMethod(methodName) {
		return e.newError(node, "property '%s' write specifier '%s' not found as method", pInfo.Name, pInfo.WriteSpec)
	}

	// Get the setter method declaration
	methodDecl := objVal.GetMethodDecl(methodName)
	if methodDecl == nil {
		return e.newError(node, "property '%s' setter method '%s' not found", pInfo.Name, methodName)
	}

	// Type-assert to get parameter info
	method, ok := methodDecl.(*ast.FunctionDecl)
	if !ok {
		return e.newError(node, "property '%s' setter is not a valid method", pInfo.Name)
	}

	// Build arguments: index directive args (if any) + value param
	indexArgs, err := e.buildIndexDirectiveArgs(pInfo)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	// Combine index args and value
	allArgs := make([]Value, 0, len(indexArgs)+1)
	allArgs = append(allArgs, indexArgs...)
	allArgs = append(allArgs, value)

	// Verify parameter count matches
	if len(method.Parameters) != len(allArgs) {
		return e.newError(node, "property '%s' setter method '%s' expects %d parameter(s), got %d",
			pInfo.Name, pInfo.WriteSpec, len(method.Parameters), len(allArgs))
	}

	// Set flag to indicate we're inside a property setter
	propCtx := ctx.PropContext()
	savedInSetter := propCtx.InPropertySetter
	propCtx.InPropertySetter = true
	defer func() {
		propCtx.InPropertySetter = savedInSetter
	}()

	// Execute the method with Self bound and arguments via adapter
	// The adapter's ExecuteMethodWithSelf handles environment setup, Self binding, etc.
	e.adapter.ExecuteMethodWithSelf(obj, methodDecl, allArgs)

	// Return the assigned value
	return value
}

// executeRecordPropertyWrite handles property setter execution for record property assignment.
//
// This method handles two property write kinds:
// - Field-backed: Direct field assignment (WriteField points to a field name)
// - Method-backed: Setter method call (WriteField points to a method name)
//
// Parameters:
//   - record: The record value to write the property to
//   - propInfo: The record property metadata (*types.RecordPropertyInfo)
//   - value: The value to assign to the property
//   - node: AST node for error reporting
//   - ctx: Execution context for environment and call stack
func (e *Evaluator) executeRecordPropertyWrite(record Value, propInfo *types.RecordPropertyInfo, value Value, node ast.Node, ctx *ExecutionContext) Value {
	// Check if property has write access
	if propInfo.WriteField == "" {
		return e.newError(node, "property '%s' is read-only", propInfo.Name)
	}

	// Get RecordValue interface
	recVal, ok := record.(*runtime.RecordValue)
	if !ok {
		return e.newError(node, "cannot write property to non-record value")
	}

	// Check if WriteField is a method or a field
	// Try as method first
	if recVal.HasRecordMethod(propInfo.WriteField) {
		return e.executeRecordPropertySetterMethod(record, recVal, propInfo, value, node, ctx)
	}

	// Field-backed property - direct field assignment
	return e.executeRecordFieldBackedPropertyWrite(recVal, propInfo, value, node)
}

// executeRecordFieldBackedPropertyWrite handles field-backed record property writes.
// Writes directly to the field specified by WriteField.
func (e *Evaluator) executeRecordFieldBackedPropertyWrite(recVal *runtime.RecordValue, propInfo *types.RecordPropertyInfo, value Value, node ast.Node) Value {
	fieldName := propInfo.WriteField

	// Use RecordFieldSetter interface to set the field
	if !recVal.SetRecordField(fieldName, value) {
		return e.newError(node, "field '%s' not found in record '%s'", fieldName, recVal.GetRecordTypeName())
	}

	return value
}

// executeRecordPropertySetterMethod handles method-backed record property writes.
// Executes the setter method specified by WriteField.
func (e *Evaluator) executeRecordPropertySetterMethod(record Value, recVal *runtime.RecordValue, propInfo *types.RecordPropertyInfo, value Value, node ast.Node, ctx *ExecutionContext) Value {
	methodName := propInfo.WriteField

	// Get the setter method declaration
	methodDecl, exists := recVal.GetRecordMethod(methodName)
	if !exists {
		return e.newError(node, "property '%s' setter method '%s' not found", propInfo.Name, methodName)
	}

	// Build arguments: just the value parameter for simple properties
	args := []Value{value}

	// Verify parameter count matches
	if len(methodDecl.Parameters) != len(args) {
		return e.newError(node, "property '%s' setter method '%s' expects %d parameter(s), got %d",
			propInfo.Name, methodName, len(methodDecl.Parameters), len(args))
	}

	// Execute the method with Self bound via adapter
	// The adapter's ExecuteMethodWithSelf handles environment setup, Self binding, etc.
	e.adapter.ExecuteMethodWithSelf(record, methodDecl, args)

	// Return the assigned value
	return value
}

// NOTE: Indexed property write support (executeIndexedPropertyWrite, executeIndexedPropertySetterMethod)
// will be added when indexed property assignment (obj.Items[i] := value) is implemented.
// See executeIndexedPropertyRead in property_read.go for the read counterpart pattern.
