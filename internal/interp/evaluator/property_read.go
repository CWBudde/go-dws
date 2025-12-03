package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// executePropertyRead handles property getter execution for ObjectValue.ReadProperty callback.
//
// This method handles three property access kinds:
// - PropAccessField: Direct field access, class variable, constant, or method call
// - PropAccessMethod: Getter method call
// - PropAccessExpression: Expression evaluation with Self bound
//
// For method-based getters, we reuse the adapter's ExecuteMethodWithSelf callback
// to avoid duplicating method execution logic.
//
// Parameters:
//   - obj: The object to read the property from (implements ObjectValue)
//   - propInfo: The property metadata (*types.PropertyInfo passed as any)
//   - node: AST node for error reporting
//   - ctx: Execution context for environment and call stack
func (e *Evaluator) executePropertyRead(obj Value, propInfo any, node ast.Node, ctx *ExecutionContext) Value {
	// Type assert propInfo to *types.PropertyInfo
	pInfo, ok := propInfo.(*types.PropertyInfo)
	if !ok {
		return e.newError(node, "invalid property info type")
	}

	// Get ObjectValue interface for class lookups
	objVal, ok := obj.(ObjectValue)
	if !ok {
		return e.newError(node, "cannot read property from non-object value")
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

	switch pInfo.ReadKind {
	case types.PropAccessField:
		return e.executeFieldBackedPropertyRead(obj, objVal, pInfo, node, ctx)

	case types.PropAccessMethod:
		return e.executeMethodBackedPropertyRead(obj, objVal, pInfo, node, ctx)

	case types.PropAccessExpression:
		return e.executeExpressionBackedPropertyRead(obj, objVal, pInfo, node, ctx)

	default:
		return e.newError(node, "property '%s' has no read access", pInfo.Name)
	}
}

// executeFieldBackedPropertyRead handles PropAccessField property reads.
// At runtime, checks in order: class vars → constants → fields → methods
func (e *Evaluator) executeFieldBackedPropertyRead(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, node ast.Node, ctx *ExecutionContext) Value {
	// 1. Try as a class variable (case-insensitive)
	if classVarValue, found := objVal.GetClassVar(pInfo.ReadSpec); found {
		return classVarValue
	}

	// 2. Try as a constant via class metadata provider
	// Note: ClassMetaProvider is implemented by ObjectInstance to access lazy-evaluated constants
	if classMetaVal, ok := obj.(ClassMetaProvider); ok {
		if constValue, found := classMetaVal.GetClassConstantBySpec(pInfo.ReadSpec); found {
			return constValue
		}
	}

	// 3. Try as an instance field
	if fieldValue := objVal.GetField(pInfo.ReadSpec); fieldValue != nil {
		// Field-backed properties with index directive are not allowed
		if pInfo.HasIndexValue {
			return e.newError(node, "property '%s' uses an index directive and cannot be field-backed", pInfo.Name)
		}
		return fieldValue
	}

	// 4. Not a field, class var, or constant - try as a getter method
	return e.executePropertyGetterMethod(obj, objVal, pInfo, node, ctx)
}

// executeMethodBackedPropertyRead handles PropAccessMethod property reads.
func (e *Evaluator) executeMethodBackedPropertyRead(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, node ast.Node, ctx *ExecutionContext) Value {
	// Indexed properties must be accessed with index syntax
	if pInfo.IsIndexed {
		return e.newError(node, "indexed property '%s' requires index arguments (e.g., obj.%s[index])", pInfo.Name, pInfo.Name)
	}

	return e.executePropertyGetterMethod(obj, objVal, pInfo, node, ctx)
}

// executePropertyGetterMethod executes a property getter method.
// Used by both PropAccessField (when falling through to method) and PropAccessMethod.
//
// This reuses the adapter's ExecuteMethodWithSelf to avoid duplicating method execution logic.
func (e *Evaluator) executePropertyGetterMethod(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, node ast.Node, ctx *ExecutionContext) Value {
	// Indexed properties must be accessed with index syntax
	if pInfo.IsIndexed {
		return e.newError(node, "indexed property '%s' requires index arguments (e.g., obj.%s[index])", pInfo.Name, pInfo.Name)
	}

	// Look up the getter method via ObjectValue interface
	methodName := pInfo.ReadSpec

	// Check if method exists
	if !objVal.HasMethod(methodName) {
		return e.newError(node, "property '%s' read specifier '%s' not found as field, constant, class var, or method", pInfo.Name, pInfo.ReadSpec)
	}

	// Build implicit index directive arguments (if any)
	indexArgs, err := e.buildIndexDirectiveArgs(pInfo)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	// Set flag to indicate we're inside a property getter
	propCtx := ctx.PropContext()
	savedInGetter := propCtx.InPropertyGetter
	propCtx.InPropertyGetter = true
	defer func() {
		propCtx.InPropertyGetter = savedInGetter
	}()

	// Use ObjectValue callback to get method and execute it via adapter
	// The adapter's ExecuteMethodWithSelf handles environment setup, Self binding, etc.
	result, invoked := objVal.InvokeParameterlessMethod(methodName, func(methodDecl any) Value {
		method := methodDecl.(*ast.FunctionDecl)

		// Verify parameter count matches index directive arguments
		if len(method.Parameters) != len(indexArgs) {
			return e.newError(node, "property '%s' getter method '%s' expects %d parameter(s), but index directive supplies %d",
				pInfo.Name, pInfo.ReadSpec, len(method.Parameters), len(indexArgs))
		}

		// Execute the method with Self bound and index args via adapter
		return e.adapter.ExecuteMethodWithSelf(obj, methodDecl, indexArgs)
	})

	if invoked {
		return result
	}

	// Method has parameters but no index directive - error
	return e.newError(node, "property '%s' getter method '%s' has parameters but no index directive", pInfo.Name, pInfo.ReadSpec)
}

// executeExpressionBackedPropertyRead handles PropAccessExpression property reads.
// Evaluates the property expression with Self bound to the object and fields accessible.
func (e *Evaluator) executeExpressionBackedPropertyRead(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, node ast.Node, ctx *ExecutionContext) Value {
	// Retrieve the AST expression from PropertyInfo
	if pInfo.ReadExpr == nil {
		return e.newError(node, "property '%s' has expression-based getter but no expression stored", pInfo.Name)
	}

	// Type-assert to ast.Expression
	exprNode, ok := pInfo.ReadExpr.(ast.Expression)
	if !ok {
		return e.newError(node, "property '%s' has invalid expression type", pInfo.Name)
	}

	// Unwrap GroupedExpression if present (parser wraps expressions in parentheses)
	if groupedExpr, ok := exprNode.(*ast.GroupedExpression); ok {
		exprNode = groupedExpr.Expression
	}

	// Create new environment with Self bound to object
	ctx.PushEnv()
	defer ctx.PopEnv()

	// Bind Self to the object instance
	e.DefineVar(ctx, "Self", obj)

	// Bind all object fields to environment so they can be accessed directly
	// This allows expressions like (FWidth * FHeight) to work
	if fieldBinder, ok := obj.(FieldBinder); ok {
		fieldBinder.BindFieldsToEnvironment(func(name string, value Value) {
			e.DefineVar(ctx, name, value)
		})
	}

	// Evaluate the expression
	return e.Eval(exprNode, ctx)
}

// buildIndexDirectiveArgs converts a property's index directive into runtime arguments.
// Helper for property getter methods with index directives.
func (e *Evaluator) buildIndexDirectiveArgs(propInfo *types.PropertyInfo) ([]Value, error) {
	if propInfo == nil || !propInfo.HasIndexValue {
		return nil, nil
	}

	if propInfo.IndexValueType != nil && propInfo.IndexValueType.Equals(types.INTEGER) {
		if intVal, ok := propInfo.IndexValue.(int64); ok {
			return []Value{&runtime.IntegerValue{Value: intVal}}, nil
		}
	}

	return nil, fmt.Errorf("property '%s' has unsupported index directive", propInfo.Name)
}

// ClassMetaProvider is an optional interface for objects that can provide class constant lookup.
// This is used by field-backed property reads to check for class constants.
// ObjectInstance implements this interface to provide access to lazy-evaluated constants.
type ClassMetaProvider interface {
	Value
	// GetClassConstantBySpec looks up a class constant by name (the ReadSpec from PropertyInfo).
	// Returns the constant value and true if found, nil and false otherwise.
	// The implementation handles lazy evaluation of constant expressions.
	GetClassConstantBySpec(name string) (Value, bool)
}

// FieldBinder is an optional interface for objects that can bind their fields to an environment.
// This is used by expression-backed property reads to make fields directly accessible.
// ObjectInstance implements this interface to expose its fields map.
type FieldBinder interface {
	Value
	// BindFieldsToEnvironment calls the binder function for each field in the object.
	// This allows expression-backed properties to access fields like FWidth, FHeight directly.
	BindFieldsToEnvironment(binder func(name string, value Value))
}

// executeIndexedPropertyRead handles indexed property getter execution.
//
// This method handles indexed property access: obj.Property[index1, index2, ...]
// - Validates property has a method-backed getter (not field-backed)
// - Looks up the getter method via ObjectValue interface
// - Verifies parameter count matches provided indices
// - Executes the method with Self bound and index parameters via adapter
//
// Parameters:
//   - obj: The object to read the property from (implements ObjectValue)
//   - propInfo: The property metadata (*types.PropertyInfo passed as any)
//   - indices: The index values to pass to the getter method
//   - node: AST node for error reporting
//   - ctx: Execution context for environment and call stack
func (e *Evaluator) executeIndexedPropertyRead(obj Value, propInfo any, indices []Value, node ast.Node, ctx *ExecutionContext) Value {
	// Type assert propInfo to *types.PropertyInfo
	pInfo, ok := propInfo.(*types.PropertyInfo)
	if !ok {
		return e.newError(node, "invalid property info type")
	}

	// Get ObjectValue interface for method lookup
	objVal, ok := obj.(ObjectValue)
	if !ok {
		return e.newError(node, "cannot read indexed property from non-object value")
	}

	// Handle based on property access kind
	switch pInfo.ReadKind {
	case types.PropAccessField, types.PropAccessMethod:
		return e.executeIndexedPropertyGetterMethod(obj, objVal, pInfo, indices, node, ctx)

	case types.PropAccessExpression:
		// Expression-based indexed properties not supported yet
		return e.newError(node, "expression-based indexed property getters not yet supported")

	default:
		return e.newError(node, "indexed property '%s' has no read access", pInfo.Name)
	}
}

// executeIndexedPropertyGetterMethod executes an indexed property getter method.
//
// Indexed properties require getter methods (not fields), because the method
// receives the index values as parameters.
func (e *Evaluator) executeIndexedPropertyGetterMethod(obj Value, objVal ObjectValue, pInfo *types.PropertyInfo, indices []Value, node ast.Node, ctx *ExecutionContext) Value {
	// Get the getter method name
	methodName := pInfo.ReadSpec

	// Look up the getter method via ObjectValue interface
	methodDecl := objVal.GetMethodDecl(methodName)
	if methodDecl == nil {
		return e.newError(node, "indexed property '%s' getter method '%s' not found", pInfo.Name, methodName)
	}

	// Type-assert to get parameter count
	method, ok := methodDecl.(*ast.FunctionDecl)
	if !ok {
		return e.newError(node, "indexed property '%s' getter is not a valid method", pInfo.Name)
	}

	// Verify method has correct number of parameters (index params, no value param)
	expectedParamCount := len(indices)
	if len(method.Parameters) != expectedParamCount {
		return e.newError(node, "indexed property '%s' getter method '%s' expects %d parameter(s), got %d index argument(s)",
			pInfo.Name, methodName, len(method.Parameters), len(indices))
	}

	// Set flag to indicate we're inside a property getter
	propCtx := ctx.PropContext()
	savedInGetter := propCtx.InPropertyGetter
	propCtx.InPropertyGetter = true
	defer func() {
		propCtx.InPropertyGetter = savedInGetter
	}()

	// Execute the method with Self bound and index arguments via adapter
	// The adapter's ExecuteMethodWithSelf handles environment setup, Self binding, etc.
	return e.adapter.ExecuteMethodWithSelf(obj, methodDecl, indices)
}
