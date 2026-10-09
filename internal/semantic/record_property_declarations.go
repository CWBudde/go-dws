package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// validateRecordPropertyDeclarations uses the record's own declaration view:
// method AST parameters already exclude Self, including instance methods.
func (a *Analyzer) validateRecordPropertyDeclarations(record *types.RecordType, props []ast.RecordPropertyDecl, methods []*ast.FunctionDecl) {
	for i := range props {
		prop := &props[i]
		info := record.GetProperty(prop.Name.Value)
		if info == nil || !info.IsIndexed || len(info.IndexParamTypes) != len(prop.IndexParams) {
			continue
		}
		for _, write := range []bool{false, true} {
			method, target := recordPropertyAccessor(props, methods, prop, write)
			if method == nil {
				if !write && target != nil {
					a.validateRecordPropertyField(record, prop, info, target)
				}
				continue
			}
			a.validateRecordPropertyMethod(prop, info, method, write)
		}
	}
}

func recordPropertyAccessor(props []ast.RecordPropertyDecl, methods []*ast.FunctionDecl, prop *ast.RecordPropertyDecl, write bool) (*ast.FunctionDecl, *ast.RecordPropertyDecl) {
	seen := make(map[string]bool)
	for prop != nil {
		if synthetic := recordPropertyExpressionAccessor(prop, write); synthetic != nil {
			return synthetic, prop
		}
		name := prop.ReadField
		if write {
			name = prop.WriteField
		}
		key := ident.Normalize(name)
		if name == "" || seen[key] {
			return nil, prop
		}
		seen[key] = true
		for _, method := range methods {
			if method.Name != nil && ident.Equal(method.Name.Value, name) {
				return method, prop
			}
		}
		next := (*ast.RecordPropertyDecl)(nil)
		for i := range props {
			if ident.Equal(props[i].Name.Value, name) {
				next = &props[i]
				break
			}
		}
		if next == nil {
			return nil, prop
		}
		prop = next
	}
	return nil, prop
}

func (a *Analyzer) validateRecordPropertyMethod(prop *ast.RecordPropertyDecl, info *types.RecordPropertyInfo, method *ast.FunctionDecl, write bool) {
	pos := recordPropertyAccessorPos(prop, write)
	var result types.Type
	if method.ReturnType != nil {
		var err error
		result, err = a.resolveTypeExpression(method.ReturnType)
		if err != nil {
			result = nil
		}
	}
	if write && method.ReturnType != nil {
		a.addStructuredError(NewPropertyDeclarationTypeMismatchError(pos, "Procedure expected"))
		return
	}
	if !write && (result == nil || !recordPropertyResultCompatible(result, info.Type)) {
		a.addStructuredError(NewPropertyDeclarationTypeMismatchError(pos, fmt.Sprintf(`Field/method "%s" has an incompatible type`, method.Name.Value)))
		return
	}
	if !a.checkRecordPropertyParameters(prop, info, method, write, pos) {
		a.addStructuredError(NewPropertyDeclarationArgumentCountError(pos, fmt.Sprintf(`Method "%s" has incompatible parameters`, method.Name.Value)))
	}
}

func (a *Analyzer) checkRecordPropertyParameters(prop *ast.RecordPropertyDecl, info *types.RecordPropertyInfo, method *ast.FunctionDecl, write bool, pos token.Position) bool {
	count := len(prop.IndexParams)
	if write {
		count++
	}
	if len(method.Parameters) != count {
		return false
	}
	if write {
		value, err := a.resolveTypeExpression(method.Parameters[count-1].Type)
		if err != nil || !types.OperatorTypesEqual(value, info.Type) {
			return false
		}
	}
	compatible := true
	for i, param := range prop.IndexParams {
		actual := method.Parameters[i]
		actualType := a.semanticInfo.GetResolvedType(actual.Type)
		if actualType == nil {
			var err error
			actualType, err = a.resolveTypeExpression(actual.Type)
			if err != nil {
				return false
			}
		}
		message := recordPropertyIndexMismatch(i, param, info.IndexParamTypes[i], actual, actualType)
		if message != "" {
			compatible = false
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(pos, message))
		}
	}
	return compatible
}

func recordPropertyAccessorPos(prop *ast.RecordPropertyDecl, write bool) token.Position {
	pos := prop.ReadAccessorPos
	if write {
		pos = prop.WriteAccessorPos
	}
	if pos.Line == 0 {
		return prop.Pos()
	}
	return pos
}

func recordPropertyIndexMismatch(i int, param *ast.Parameter, expected types.Type, actual *ast.Parameter, actualType types.Type) string {
	if !propertyIndexTypesCompatible(expected, actualType) {
		return fmt.Sprintf(`Parameter %d - Type "%s" expected (instead of "%s")`, i, semanticTypeNameForDiagnostic(expected), semanticTypeNameForDiagnostic(actualType))
	}
	if mode := propertyParameterModeMismatch(param, actual.ByRef, actual.IsConst); mode != "" {
		return fmt.Sprintf("Parameter %d (%s) - %s-parameter expected", i, param.Name.Value, mode)
	}
	if actual.DefaultValue != nil {
		return fmt.Sprintf("Parameter %d (%s) - default value at implementation does not match declaration or forward", i, param.Name.Value)
	}
	return ""
}

// recordPropertyExpressionAccessor mirrors anonymous declaration signatures only;
// it does not register an executable method or make index names runnable in bodies.
func recordPropertyExpressionAccessor(prop *ast.RecordPropertyDecl, write bool) *ast.FunctionDecl {
	if (!write && prop.ReadExpr == nil) || (write && prop.WriteStmt == nil) {
		return nil
	}
	method := &ast.FunctionDecl{Name: &ast.Identifier{}, Parameters: append([]*ast.Parameter(nil), prop.IndexParams...), IsClassMethod: prop.IsClassProperty}
	if write {
		method.Parameters = append(method.Parameters, &ast.Parameter{Name: &ast.Identifier{Value: "Value"}, Type: prop.Type, IsConst: true})
	} else {
		method.ReturnType = prop.Type
	}
	return method
}

func (a *Analyzer) validateRecordPropertyField(record *types.RecordType, prop *ast.RecordPropertyDecl, info *types.RecordPropertyInfo, target *ast.RecordPropertyDecl) {
	key := ident.Normalize(target.ReadField)
	field := record.Fields[key]
	if field == nil {
		return
	}
	pos := recordPropertyAccessorPos(prop, false)
	message := "Function expected"
	if !recordPropertyResultCompatible(field, info.Type) {
		message = fmt.Sprintf(`Field/method "%s" has an incompatible type`, record.FieldNames[key])
	}
	a.addStructuredError(NewPropertyDeclarationTypeMismatchError(pos, message))
}

// recordPropertyResultCompatible mirrors SameType/IsOfType without numeric or
// class-to-interface assignment coercions. Setter Value still requires SameType.
func recordPropertyResultCompatible(actual, expected types.Type) bool {
	if types.OperatorTypesEqual(actual, expected) {
		return true
	}
	switch result := types.GetUnderlyingType(actual).(type) {
	case *types.ClassType:
		target, ok := types.GetUnderlyingType(expected).(*types.ClassType)
		return ok && types.IsSubclassOf(result, target)
	case *types.InterfaceType:
		target, ok := types.GetUnderlyingType(expected).(*types.InterfaceType)
		return ok && types.IsSubinterfaceOf(result, target)
	}
	return false
}
