package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// VisitAnonymousRecordExpression evaluates DWScript's anonymous record
// constructor expression:
//
//	record a := 1; b := 'x'; end
//
// The record is structurally typed: each field's type is taken from the value it
// is initialized with, and the resulting *types.RecordType is unnamed. Properties
// and methods share metadata between copies. Because
// the expression carries its own type, this path deliberately does not consult
// or set ExecutionContext's record type context, unlike
// VisitRecordLiteralExpression.
func (e *Evaluator) VisitAnonymousRecordExpression(node *ast.AnonymousRecordExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil record expression")
	}

	recordType := types.NewRecordType("", nil)
	fieldValues := make(map[string]Value, len(node.Fields))

	for _, field := range node.Fields {
		if field == nil || field.Name == nil {
			return e.newError(node, "record expression field requires a name")
		}

		fieldName := field.Name.Value
		if recordType.HasField(fieldName) {
			return e.newError(node, "duplicate field '%s' in record expression", fieldName)
		}

		fieldValue := e.Eval(field.Value, ctx)
		if isError(fieldValue) {
			return fieldValue
		}

		recordType.AddField(fieldName, GetValueType(fieldValue), true)
		fieldValues[ident.Normalize(fieldName)] = fieldValue
	}
	if resolved, ok := e.resolvedSemanticType(node).(*types.RecordType); ok {
		recordType = resolved
	} else {
		// Direct evaluator callers may not run semantic analysis first.
		for _, prop := range node.Properties {
			propType, err := e.ResolveTypeFromAnnotation(prop.Type, ctx)
			if err != nil || propType == nil {
				return e.newError(node, "unknown type for property '%s' in record expression", prop.Name.Value)
			}
			if prop.IsAutoProperty && !recordType.HasField(prop.ReadField) {
				recordType.AddField(prop.ReadField, propType, false)
			}
			info := &types.RecordPropertyInfo{Name: prop.Name.Value, Type: propType,
				ReadField: prop.ReadField, WriteField: prop.WriteField}
			if prop.ReadField != "" {
				info.ReadKind = types.PropAccessField
			} else if prop.ReadExpr != nil {
				info.ReadKind = types.PropAccessExpression
				info.ReadExpr = prop.ReadExpr
			}
			if prop.WriteField != "" {
				info.WriteKind = types.PropAccessField
			} else if prop.WriteStmt != nil {
				info.WriteKind = types.PropAccessExpression
				info.WriteExpr = prop.WriteStmt
			}
			recordType.Properties[ident.Normalize(prop.Name.Value)] = info
		}
	}

	var metadata *runtime.RecordMetadata
	if len(node.Methods) != 0 || len(node.Properties) != 0 {
		methods := make(map[string]*ast.FunctionDecl)
		overloads := make(map[string][]*ast.FunctionDecl)
		for _, method := range node.Methods {
			key := ident.Normalize(method.Name.Value)
			if _, exists := methods[key]; !exists {
				methods[key] = method
			}
			overloads[key] = append(overloads[key], method)
		}
		metadata = e.buildRecordMetadata("", recordType, methods, nil, overloads, nil, nil, nil, ctx)
	}

	recordValue := runtime.NewRecordValueWithInitializer(recordType, metadata,
		func(fieldName string, fieldType types.Type) runtime.Value {
			if val, ok := fieldValues[ident.Normalize(fieldName)]; ok {
				return val
			}
			return e.getZeroValueForType(fieldType, ctx)
		})

	return recordValue
}
