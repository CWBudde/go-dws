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
// is initialized with, and the resulting *types.RecordType is unnamed. Because
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

	recordValue := runtime.NewRecordValueWithInitializer(recordType, nil,
		func(fieldName string, fieldType types.Type) runtime.Value {
			if val, ok := fieldValues[ident.Normalize(fieldName)]; ok {
				return val
			}
			return e.getZeroValueForType(fieldType, ctx)
		})

	return recordValue
}
