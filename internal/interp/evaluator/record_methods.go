package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// callRecordMethod executes a record method in the evaluator.
//
// Record methods are user-defined methods attached to record types.
// They execute with Self bound to the record instance.
//
// Execution semantics:
// - Create new environment (child of current)
// - Bind Self to record instance
// - Bind method parameters from args
// - Initialize Result variable (if method has return type)
// - Execute method body
// - Extract and return Result
func (e *Evaluator) callRecordMethod(
	record RecordInstanceValue,
	method *ast.FunctionDecl,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	// 1. Validate parameter count
	if len(args) != len(method.Parameters) {
		return e.newError(node,
			"wrong number of arguments for method '%s': expected %d, got %d",
			method.Name.Value, len(method.Parameters), len(args))
	}

	// 2. Create method environment (child of current context)
	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())

	// 3. Bind Self to record instance and expose fields directly.
	scope.defineExposed(ctx, "Self", record)
	e.bindRecordMethodFields(record, ctx, scope)
	e.bindRecordMethodClassState(record, ctx, scope)

	// 4. Bind method parameters
	for i, param := range method.Parameters {
		paramName := param.Name.Value
		scope.defineOwned(e, ctx, paramName, args[i])
	}

	// 5. Initialize Result variable
	// DWScript uses implicit Result variable for function return values
	if method.ReturnType != nil {
		returnType, err := e.ResolveTypeFromAnnotation(method.ReturnType)
		if err != nil {
			return e.newError(node, "failed to resolve return type: %v", err)
		}
		defaultVal := e.defaultReturnValue(returnType)
		scope.defineOwned(e, ctx, "Result", defaultVal)
		scope.defineExposed(ctx, method.Name.Value, e.newResultAlias(ctx.Env()))
	}

	// 6. Execute method body in new environment
	result := e.Eval(method.Body, ctx)
	if isError(result) {
		return result
	}

	e.syncRecordMethodFields(record, ctx)
	e.syncRecordMethodClassState(record, ctx)

	// Exit only leaves the record method body. Clear it before returning to the caller,
	// matching user-function and lambda call handling.
	if ctx.ControlFlow().IsExit() {
		ctx.ControlFlow().Clear()
	}

	// 7. Extract Result variable (if method has return type)
	if method.ReturnType != nil {
		returnValue := e.extractReturnValue(method.Name.Value, ctx)
		return e.retainValueForBinding(returnValue, ctx)
	}

	// Procedure (no return type) - return nil
	return e.nilValue()
}

func (e *Evaluator) callRecordStaticMethod(
	recordType *RecordTypeValue,
	methodName string,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if recordType == nil {
		return e.newError(node, "invalid record type")
	}

	method, errVal := e.resolveRecordStaticMethod(recordType, methodName, args, node)
	if errVal != nil {
		return errVal
	}

	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())

	scope.defineExposed(ctx, "__CurrentRecord__", recordType)
	if recordType.GetRecordTypeName() != "" {
		scope.defineExposed(ctx, recordType.GetRecordTypeName(), recordType)
	}
	for name, value := range recordType.Constants {
		scope.defineExposed(ctx, name, value)
	}
	for name, value := range recordType.ClassVars {
		scope.defineExposed(ctx, name, value)
	}

	for idx, param := range method.Parameters {
		scope.defineOwned(e, ctx, param.Name.Value, args[idx])
	}

	if method.ReturnType != nil {
		returnType, err := e.ResolveTypeFromAnnotation(method.ReturnType)
		if err != nil {
			return e.newError(node, "failed to resolve return type: %v", err)
		}
		defaultVal := e.defaultReturnValue(returnType)
		scope.defineOwned(e, ctx, "Result", defaultVal)
		scope.defineExposed(ctx, method.Name.Value, e.newResultAlias(ctx.Env()))
	}

	result := e.Eval(method.Body, ctx)
	if isError(result) {
		return result
	}

	for name := range recordType.ClassVars {
		if val, ok := ctx.Env().Get(name); ok {
			recordType.ClassVars[name] = val
		}
	}

	if ctx.ControlFlow().IsExit() {
		ctx.ControlFlow().Clear()
	}

	if method.ReturnType != nil {
		returnValue := e.extractReturnValue(method.Name.Value, ctx)
		return e.retainValueForBinding(returnValue, ctx)
	}

	return e.nilValue()
}

func (e *Evaluator) resolveRecordStaticMethod(
	recordType *RecordTypeValue,
	methodName string,
	args []Value,
	node ast.Node,
) (*ast.FunctionDecl, Value) {
	normalized := ident.Normalize(methodName)
	overloads := recordType.ClassMethodOverloads[normalized]
	if len(overloads) == 0 {
		if method, ok := recordType.ClassMethods[normalized]; ok {
			overloads = []*ast.FunctionDecl{method}
		}
	}
	if len(overloads) == 0 {
		return nil, e.newError(node, "static method '%s' not found in record type '%s'", methodName, recordType.GetRecordTypeName())
	}

	if len(overloads) == 1 {
		candidate := overloads[0]
		if len(candidate.Parameters) == len(args) {
			return candidate, nil
		}
	} else if candidate := e.resolveRecordMethodOverload(methodName, overloads, args); candidate != nil {
		return candidate, nil
	}

	return nil, e.newError(node, "wrong number of arguments for static method '%s.%s'",
		recordType.GetRecordTypeName(), methodName)
}

func (e *Evaluator) resolveRecordMethodOverload(methodName string, overloads []*ast.FunctionDecl, args []Value) *ast.FunctionDecl {
	argTypes := make([]types.Type, len(args))
	for idx, arg := range args {
		argTypes[idx] = e.getValueType(arg)
	}

	candidates := make([]*semantic.Symbol, 0, len(overloads))
	for _, overload := range overloads {
		if len(overload.Parameters) != len(args) {
			continue
		}
		funcType := e.extractFunctionType(overload, e.currentContext)
		if funcType == nil {
			continue
		}
		candidates = append(candidates, &semantic.Symbol{
			Name:                 overload.Name.Value,
			Type:                 funcType,
			HasOverloadDirective: overload.IsOverload,
		})
	}
	if len(candidates) == 0 {
		return nil
	}

	selected, err := semantic.ResolveOverload(candidates, argTypes)
	if err != nil {
		return nil
	}
	if selectedType, ok := selected.Type.(*types.FunctionType); ok {
		for _, overload := range overloads {
			funcType := e.extractFunctionType(overload, e.currentContext)
			if funcType != nil && semantic.SignaturesEqual(funcType, selectedType) &&
				funcType.ReturnType.Equals(selectedType.ReturnType) {
				return overload
			}
		}
	}

	return nil
}

func (e *Evaluator) defaultReturnValue(returnType types.Type) Value {
	if returnType != nil && returnType.TypeKind() == "RECORD" {
		return e.getZeroValueForType(returnType)
	}
	return e.GetDefaultValue(returnType)
}

func (e *Evaluator) bindRecordMethodFields(record RecordInstanceValue, ctx *ExecutionContext, scope *bindingScope) {
	recVal, ok := record.(*runtime.RecordValue)
	if !ok || recVal == nil {
		return
	}

	for fieldName := range recVal.Fields {
		normalizedFieldName := fieldName
		ref := runtime.NewReferenceValue(fieldName, func() (runtime.Value, error) {
			return recVal.Fields[normalizedFieldName], nil
		}, func(value runtime.Value) error {
			recVal.Fields[normalizedFieldName] = value
			return nil
		})
		scope.defineExposed(ctx, fieldName, ref)
	}

	if recVal.RecordType != nil && recVal.RecordType.Properties != nil {
		for propName, propInfo := range recVal.RecordType.Properties {
			if propInfo.ReadField == "" {
				continue
			}
			fieldKey := ident.Normalize(propInfo.ReadField)
			if _, exists := recVal.Fields[fieldKey]; exists {
				normalizedFieldName := fieldKey
				ref := runtime.NewReferenceValue(propName, func() (runtime.Value, error) {
					return recVal.Fields[normalizedFieldName], nil
				}, func(value runtime.Value) error {
					recVal.Fields[normalizedFieldName] = value
					return nil
				})
				scope.defineExposed(ctx, propName, ref)
			}
		}
	}
}

func (e *Evaluator) syncRecordMethodFields(record RecordInstanceValue, ctx *ExecutionContext) {
	recVal, ok := record.(*runtime.RecordValue)
	if !ok || recVal == nil {
		return
	}

	for fieldName := range recVal.Fields {
		if value, ok := ctx.Env().Get(fieldName); ok {
			if ref, isRef := value.(ReferenceAccessor); isRef {
				deref, err := ref.Dereference()
				if err == nil {
					recVal.Fields[fieldName] = deref
				}
				continue
			}
			recVal.Fields[fieldName] = value
		}
	}
}

func (e *Evaluator) bindRecordMethodClassState(record RecordInstanceValue, ctx *ExecutionContext, scope *bindingScope) {
	recVal, ok := record.(*runtime.RecordValue)
	if !ok || recVal == nil || recVal.RecordType == nil {
		return
	}

	recordTypeKey := "__record_type_" + ident.Normalize(recVal.RecordType.Name)
	recordTypeRaw, found := ctx.Env().Get(recordTypeKey)
	if !found {
		return
	}
	recordType, ok := recordTypeRaw.(*RecordTypeValue)
	if !ok || recordType == nil {
		return
	}

	scope.defineExposed(ctx, "__CurrentRecord__", recordType)
	for name, value := range recordType.Constants {
		scope.defineExposed(ctx, name, value)
	}
	for name, value := range recordType.ClassVars {
		scope.defineExposed(ctx, name, value)
	}
}

func (e *Evaluator) syncRecordMethodClassState(record RecordInstanceValue, ctx *ExecutionContext) {
	recVal, ok := record.(*runtime.RecordValue)
	if !ok || recVal == nil || recVal.RecordType == nil {
		return
	}

	recordTypeKey := "__record_type_" + ident.Normalize(recVal.RecordType.Name)
	recordTypeRaw, found := ctx.Env().Get(recordTypeKey)
	if !found {
		return
	}
	recordType, ok := recordTypeRaw.(*RecordTypeValue)
	if !ok || recordType == nil {
		return
	}

	for name := range recordType.ClassVars {
		if value, ok := ctx.Env().Get(name); ok {
			recordType.ClassVars[name] = value
		}
	}
}
