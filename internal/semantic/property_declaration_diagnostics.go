package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

func propertyAccessorDiagnosticPos(prop *ast.PropertyDecl, write bool) token.Position {
	accessor := prop.ReadSpec
	if write {
		accessor = prop.WriteSpec
	}
	if accessor != nil {
		return accessor.Pos()
	}
	return prop.Pos()
}

func (a *Analyzer) propertyAccessorDeclaredName(class *types.ClassType, name string) string {
	key := ident.Normalize(name)
	if owner := a.getMethodOwner(class, name); owner != nil {
		if declared := owner.MethodDeclNames[key]; declared != "" {
			return declared
		}
	}
	if owner := a.getFieldOwner(class, name); owner != nil {
		if declared := owner.FieldDeclNames[key]; declared != "" {
			return declared
		}
	}
	if owner := a.getClassVarOwner(class, name); owner != nil {
		if declared := owner.ClassVarDeclNames[key]; declared != "" {
			return declared
		}
	}
	for owner := class; owner != nil; owner = owner.Parent {
		for declared := range owner.ConstantTypes {
			if ident.Equal(declared, name) {
				return declared
			}
		}
	}
	return name
}

func (a *Analyzer) reportPropertyAccessorTypeMismatch(prop *ast.PropertyDecl, class *types.ClassType, name string, write bool) {
	caption := "Field/method"
	if write {
		caption = "Symbol"
	}
	a.addStructuredError(NewPropertyDeclarationTypeMismatchError(propertyAccessorDiagnosticPos(prop, write),
		fmt.Sprintf(`%s "%s" has an incompatible type`, caption, a.propertyAccessorDeclaredName(class, name))))
}

// checkPropertyAccessorParameters follows CheckPropertyFuncParams: count and
// value/index-directive types produce only the summary error; explicit index
// parameter mismatches produce their detail before the summary.
func (a *Analyzer) checkPropertyAccessorParameters(prop *ast.PropertyDecl, method *types.FunctionType, indices []types.Type, value types.Type, pos token.Position) bool {
	count := len(indices)
	if value != nil {
		count++
	}
	if len(method.Parameters) != count {
		return false
	}
	if value != nil && !types.OperatorTypesEqual(method.Parameters[count-1], value) {
		return false
	}
	offset := 0
	if prop.IndexValue != nil {
		if len(indices) == 0 || !types.OperatorTypesEqual(method.Parameters[0], indices[0]) {
			return false
		}
		offset = 1
	}
	compatible := true
	for i, expected := range indices[offset:] {
		actual := method.Parameters[i+offset]
		if !propertyIndexTypesCompatible(expected, actual) {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(pos,
				fmt.Sprintf(`Parameter %d - Type "%s" expected (instead of "%s")`, i, semanticTypeNameForDiagnostic(expected), semanticTypeNameForDiagnostic(actual))))
			compatible = false
			continue // A type mismatch suppresses this index's mode mismatch.
		}
		if i >= len(prop.IndexParams) {
			continue // A pure index directive has no declared index parameters.
		}
		param := prop.IndexParams[i]
		actualVar := i+offset < len(method.VarParams) && method.VarParams[i+offset]
		actualConst := i+offset < len(method.ConstParams) && method.ConstParams[i+offset]
		mode := propertyParameterModeMismatch(param, actualVar, actualConst)
		if mode != "" {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(pos,
				fmt.Sprintf("Parameter %d (%s) - %s-parameter expected", i, param.Name.Value, mode)))
			compatible = false
		}
	}
	return compatible
}

// propertyIndexTypesCompatible mirrors the declaration-side primitive relation
// in CheckParams. In particular Integer/Float do not coerce, and a Variant
// index can be backed by a scalar, enum, class, or interface parameter.
func propertyIndexTypesCompatible(expected, actual types.Type) bool {
	expected, actual = types.GetUnderlyingType(expected), types.GetUnderlyingType(actual)
	if expected == nil || actual == nil {
		return false
	}
	if expected.TypeKind() == "VARIANT" {
		switch actual.TypeKind() {
		case "INTEGER", "FLOAT", "STRING", "BOOLEAN", "VARIANT", "ENUM", "CLASS", "INTERFACE", "NIL":
			return true
		}
	}
	if (expected.TypeKind() == "INTEGER" || expected.TypeKind() == "ENUM") &&
		(actual.TypeKind() == "INTEGER" || actual.TypeKind() == "ENUM") {
		return true
	}
	return operatorBindingTypesCompatible(expected, actual)
}

// propertyParameterModeMismatch preserves CheckParams' var/value/const priority.
func propertyParameterModeMismatch(param *ast.Parameter, actualVar, actualConst bool) string {
	switch {
	case param.ByRef && !actualVar:
		return "Var"
	case !param.ByRef && actualVar:
		return "Value"
	case param.IsConst && !actualConst:
		return "Const"
	case !param.IsConst && actualConst:
		return "Value"
	}
	return ""
}
