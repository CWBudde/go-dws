package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// recordReceiverType unwraps a record receiver for member ownership only.
// Assignment and overload compatibility must retain the metatype distinction.
func recordReceiverType(typ types.Type) (*types.RecordType, bool) {
	switch t := types.GetUnderlyingType(typ).(type) {
	case *types.RecordType:
		return t, false
	case *types.RecordMetaType:
		return t.RecordType, true
	default:
		return nil, false
	}
}

func recordTypeValueType(typ types.Type) types.Type {
	if meta := types.NewRecordMetaType(typ); meta != nil {
		return meta
	}
	return typ
}

func recordPropertyIsStatic(record *types.RecordType, prop *types.RecordPropertyInfo) bool {
	if prop.IsClassProperty {
		return true
	}
	for _, spec := range []string{prop.ReadField, prop.WriteField} {
		if spec == "" {
			continue
		}
		key := ident.Normalize(spec)
		if _, ok := record.ClassVars[key]; ok {
			continue
		}
		if _, ok := record.Constants[key]; ok {
			continue
		}
		if record.HasClassMethod(spec) {
			continue
		}
		return false
	}
	return prop.ReadField != "" || prop.WriteField != ""
}

// analyzeRecordMetaIndex validates one bracket list against its actual receiver.
// The final two results let ordinary single-index analysis reuse a receiver
// already analyzed here, without repeating its diagnostics.
func (a *Analyzer) analyzeRecordMetaIndex(expr *ast.IndexExpression) (types.Type, bool, types.Type, bool) {
	indices := []ast.Expression{expr.Index}
	first := expr
	for first.CommaPos.Line != 0 {
		inner, ok := first.Left.(*ast.IndexExpression)
		if !ok {
			break
		}
		indices = append([]ast.Expression{inner.Index}, indices...)
		first = inner
	}
	base := first.Left
	receiver := base
	name := ""
	if member, ok := base.(*ast.MemberAccessExpression); ok {
		receiver, name = member.Object, member.Member.Value
	}
	// Keep unrelated lexical receivers on their existing analysis paths.
	typ := a.semanticInfo.GetResolvedType(receiver)
	if typ == nil {
		typ = a.inferMemberObjectType(receiver)
	}
	if typ != nil {
		if _, meta := recordReceiverType(typ); !meta {
			return nil, false, nil, false
		}
	}
	typ = a.analyzeIndexBase(receiver)
	record, meta := recordReceiverType(typ)
	if !meta {
		return nil, false, typ, receiver == expr.Left
	}
	prop := recordMetaPropertyByName(record, name)
	if prop == nil || !prop.IsIndexed || !recordPropertyIsStatic(record, prop) {
		return nil, false, typ, receiver == expr.Left
	}
	a.warnDeprecatedRecordPropertyUsage(prop, base.Pos())
	expected := a.getIndexedRecordPropertyParamTypes(prop, record)
	if len(indices) != len(expected) {
		a.addArgumentCountError(expr.Token.Pos, len(indices), len(expected), len(expected))
		return prop.Type, true, nil, false
	}
	for i, index := range indices {
		got := a.analyzeExpressionWithExpectedType(index, expected[i])
		if got != nil && !a.canAssign(got, expected[i]) {
			a.addStructuredError(NewArrayIndexError(index.Pos(), expected[i].String(), got.String()))
		}
	}
	return prop.Type, true, nil, false
}

// recordMetaPropertyByName selects the default property when no name is supplied.
func recordMetaPropertyByName(record *types.RecordType, name string) *types.RecordPropertyInfo {
	if name != "" {
		return record.GetProperty(name)
	}
	for _, prop := range record.Properties {
		if prop.IsDefault {
			return prop
		}
	}
	return nil
}

// recordMetaHelperOrder retains the existing reverse-search convention while
// matching runtime lookup: alias helpers first, descendants before ancestors,
// and declaration order for unrelated helpers of the same target.
func recordMetaHelperOrder(helpers []*types.HelperType, source types.Type) []*types.HelperType {
	var prioritized []*types.HelperType
	for _, exact := range []bool{true, false} {
		for _, helper := range helpers {
			if ident.Equal(helper.TargetType.String(), source.String()) == exact {
				prioritized = append(prioritized, helper)
			}
		}
	}
	var ordered []*types.HelperType
	for _, helper := range prioritized {
		at := len(ordered)
		for i, placed := range ordered {
			for parent := helper.ParentHelper; parent != nil; parent = parent.ParentHelper {
				if parent == placed && i < at {
					at = i
				}
			}
		}
		ordered = append(ordered, nil)
		copy(ordered[at+1:], ordered[at:])
		ordered[at] = helper
	}
	for i, j := 0, len(ordered)-1; i < j; i, j = i+1, j-1 {
		ordered[i], ordered[j] = ordered[j], ordered[i]
	}
	return ordered
}

// helperBodySelfType changes only the record class-helper receiver category;
// other helpers retain their established body scope rules.
func helperBodySelfType(target types.Type, method *ast.FunctionDecl) types.Type {
	if method.IsClassMethod {
		if meta := types.NewRecordMetaType(target); meta != nil {
			return meta
		}
	}
	return target
}
