package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
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

	// 7. Check for early return/exit
	// Control flow is managed by ExecutionContext.ControlFlow()
	if ctx.ControlFlow().IsExit() {
		// Exit propagates up the call stack
		return e.nilValue()
	}

	// 8. Extract Result variable (if method has return type)
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

	for _, candidate := range overloads {
		if len(candidate.Parameters) != len(args) {
			continue
		}
		return candidate, nil
	}

	return nil, e.newError(node, "wrong number of arguments for static method '%s.%s'",
		recordType.GetRecordTypeName(), methodName)
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

	for fieldName, fieldValue := range recVal.Fields {
		scope.defineExposed(ctx, fieldName, fieldValue)
	}

	if recVal.RecordType != nil && recVal.RecordType.Properties != nil {
		for propName, propInfo := range recVal.RecordType.Properties {
			if propInfo.ReadField == "" {
				continue
			}
			if fieldValue, exists := recVal.Fields[ident.Normalize(propInfo.ReadField)]; exists {
				scope.defineExposed(ctx, propName, fieldValue)
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
