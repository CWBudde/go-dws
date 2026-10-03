package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// bindRecordMetaMembers makes bare class members share the canonical storage.
func (e *Evaluator) bindRecordMetaMembers(record *runtime.RecordTypeValue, ctx *ExecutionContext, scope *bindingScope) {
	scope.defineExposed(ctx, "__CurrentRecord__", record)
	for name := range record.ClassVars {
		key := name
		scope.defineExposed(ctx, key, runtime.NewReferenceValue(key, func() (runtime.Value, error) {
			return record.ClassVars[key], nil
		}, func(value runtime.Value) error {
			record.ClassVars[key] = value
			if record.Metadata != nil {
				record.Metadata.ClassVars[key] = value
			}
			return nil
		}))
	}
	for name, value := range record.Constants {
		scope.defineExposed(ctx, name, value)
	}
}

func recordMetaProperty(record *runtime.RecordTypeValue, name string) *types.RecordPropertyInfo {
	if record == nil || record.RecordType == nil {
		return nil
	}
	if name != "" {
		return record.RecordType.Properties[ident.Normalize(name)]
	}
	for _, prop := range record.RecordType.Properties {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

// resolveRecordMetaIndexedProperty captures one bracket list. Commas belong to
// this property's arguments; earlier brackets belong to its receiver, which
// may itself select a metatype from an aggregate.
func (e *Evaluator) resolveRecordMetaIndexedProperty(node *ast.IndexExpression, ctx *ExecutionContext) (*runtime.RecordTypeValue, *types.RecordPropertyInfo, []Value, bool, Value) {
	indices := []ast.Expression{node.Index}
	first := node
	for first.CommaPos.Line != 0 {
		inner, ok := first.Left.(*ast.IndexExpression)
		if !ok {
			break
		}
		indices = append([]ast.Expression{inner.Index}, indices...)
		first = inner
	}
	receiver := first.Left
	name := ""
	if member, ok := receiver.(*ast.MemberAccessExpression); ok {
		receiver, name = member.Object, member.Member.Value
	}
	meta := e.recordMetaIndexReceiverType(receiver, ctx)
	if meta == nil {
		return nil, nil, nil, false, nil
	}
	prop := recordMetaProperty(&runtime.RecordTypeValue{RecordType: meta.RecordType}, name)
	if prop == nil || !prop.IsIndexed {
		return nil, nil, nil, false, nil
	}
	value := e.Eval(receiver, ctx)
	value = e.normalizeMemberReceiver(value, receiver, node, ctx)
	if isError(value) || ctx.Exception() != nil {
		return nil, nil, nil, true, value
	}
	record, ok := value.(*runtime.RecordTypeValue)
	if !ok {
		return nil, nil, nil, true, e.newError(node, "record type expected for indexed property '%s'", prop.Name)
	}
	values, err := e.recordMetaPropertyIndices(indices, ctx)
	return record, prop, values, true, err
}

// recordMetaIndexReceiverType finds ownership without evaluating the receiver.
// Compiled expressions carry their actual intermediate type; the environment
// fallback preserves direct variable access when semantic analysis is disabled.
func (e *Evaluator) recordMetaIndexReceiverType(expr ast.Expression, ctx *ExecutionContext) *types.RecordMetaType {
	var typ types.Type
	if e.SemanticInfo() != nil {
		typ = e.SemanticInfo().GetResolvedType(expr)
	} else if id, ok := expr.(*ast.Identifier); ok {
		if value, found := ctx.Env().Get(id.Value); found {
			typ = runtime.LanguageType(value)
		}
	}
	meta, ok := types.GetUnderlyingType(typ).(*types.RecordMetaType)
	if !ok {
		return nil
	}
	return meta
}

func (e *Evaluator) recordMetaPropertyRead(record *runtime.RecordTypeValue, prop *types.RecordPropertyInfo, indices []Value, node ast.Node, ctx *ExecutionContext) Value {
	if expr, ok := prop.ReadExpr.(ast.Expression); ok && prop.ReadKind == types.PropAccessExpression {
		return e.evalRecordMetaPropertyExpression(record, expr, nil, nil, ctx)
	}
	if record.HasStaticMethod(prop.ReadField) {
		return e.callRecordStaticMethod(record, prop.ReadField, indices, node, ctx)
	}
	if value, ok := readRecordTypePropertyValue(record, prop); ok {
		return value
	}
	return e.newError(node, "property '%s' has no readable record type accessor", prop.Name)
}

func (e *Evaluator) recordMetaPropertyWrite(record *runtime.RecordTypeValue, prop *types.RecordPropertyInfo, indices []Value, value Value, node ast.Node, ctx *ExecutionContext) Value {
	if stmt, ok := prop.WriteExpr.(ast.Statement); ok && prop.WriteKind == types.PropAccessExpression {
		return e.evalRecordMetaPropertyExpression(record, nil, stmt, value, ctx)
	}
	if record.HasStaticMethod(prop.WriteField) {
		args := append(append([]Value(nil), indices...), value)
		return e.callRecordStaticMethod(record, prop.WriteField, args, node, ctx)
	}
	key := ident.Normalize(prop.WriteField)
	if _, found := record.ClassVars[key]; found {
		record.ClassVars[key] = value
		if record.Metadata != nil {
			record.Metadata.ClassVars[key] = value
		}
		return value
	}
	return e.newError(node, "property '%s' has no writable record type accessor", prop.Name)
}

func (e *Evaluator) recordMetaPropertyIndices(expressions []ast.Expression, ctx *ExecutionContext) ([]Value, Value) {
	values := make([]Value, len(expressions))
	for i, expr := range expressions {
		values[i] = e.Eval(expr, ctx)
		if isError(values[i]) || ctx.Exception() != nil {
			return nil, values[i]
		}
	}
	return values, nil
}

func (e *Evaluator) compoundRecordMetaProperty(record *runtime.RecordTypeValue, prop *types.RecordPropertyInfo, indices []Value, index *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	current := e.recordMetaPropertyRead(record, prop, indices, index, ctx)
	if isError(current) || ctx.Exception() != nil {
		return current
	}
	right := e.Eval(stmt.Value, ctx)
	if isError(right) || ctx.Exception() != nil {
		return right
	}
	result := e.applyCompoundOperation(stmt.Operator, current, right, stmt, ctx)
	if isError(result) || ctx.Exception() != nil {
		return result
	}
	return e.recordMetaPropertyWrite(record, prop, indices, result, stmt, ctx)
}

// evalRecordMetaPropertyExpression binds class state by reference, so nested
// accessors and direct writes observe the same storage throughout evaluation.
func (e *Evaluator) evalRecordMetaPropertyExpression(record *runtime.RecordTypeValue, read ast.Expression, write ast.Statement, value Value, ctx *ExecutionContext) Value {
	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())
	scope.defineExposed(ctx, "Self", record)
	e.bindRecordMetaMembers(record, ctx, scope)
	if read != nil {
		return e.Eval(read, ctx)
	}
	scope.defineOwned(e, ctx, "Value", value)
	if result := e.Eval(write, ctx); isError(result) {
		return result
	}
	return value
}
