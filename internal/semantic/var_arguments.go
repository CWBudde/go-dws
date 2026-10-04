package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// checkVarArgument preserves the parameter's declared name and the written
// argument's position. Constants and read-only bindings are data expressions,
// but cannot provide writable storage to a var parameter.
func (a *Analyzer) checkVarArgument(index int, name string, arg ast.Expression, argType types.Type, pos token.Position) bool {
	if a.isWritableVarArgument(arg, argType) {
		a.markVarArgumentWritten(arg)
		return true
	}
	a.addStructuredError(NewGenericError(pos,
		fmt.Sprintf("Argument %d (%s) cannot be passed as Var-parameter", index, name)))
	return false
}

func (a *Analyzer) isWritableVarArgument(arg ast.Expression, argType types.Type) bool {
	switch expr := arg.(type) {
	case *ast.GroupedExpression:
		return a.isWritableVarArgument(expr.Expression, argType)
	case *ast.SelfExpression:
		// Record instance methods receive Self by reference; object methods
		// receive an immutable object reference.
		_, record := types.GetUnderlyingType(argType).(*types.RecordType)
		return record && !a.inClassMethod
	case *ast.Identifier:
		sym, found := a.symbols.Resolve(expr.Value)
		return found && isWritableVarSymbol(sym)
	case *ast.IndexExpression:
		baseType := types.GetUnderlyingType(a.semanticInfo.GetResolvedType(expr.Left))
		if array, ok := baseType.(*types.ArrayType); ok && array.IsStatic() {
			return a.isWritableVarArgument(expr.Left, baseType)
		}
		return baseType != types.STRING
	case *ast.MemberAccessExpression:
		return a.isWritableVarMember(expr)
	default:
		// Calls and explicit routine references have no caller-owned slot.
		// Routine-reference temporaries require separate runtime support.
		return false
	}
}

// isWritableVarSymbol reports whether a resolved variable binding provides
// writable storage; constants, read-only bindings and routines do not.
func isWritableVarSymbol(sym *Symbol) bool {
	if sym.ReadOnly || sym.IsConst {
		return false
	}
	_, routine := sym.Type.(*types.FunctionType)
	return !routine
}

func (a *Analyzer) isWritableVarMember(expr *ast.MemberAccessExpression) bool {
	// Unit-qualified members resolve from the unit's symbol table and carry no
	// receiver type, so classify the qualified symbol itself.
	if unitName, ok := expr.Object.(*ast.Identifier); ok && !a.hasLexicalValueReceiver(expr.Object) {
		if _, imported := a.importedUnitNamespace(unitName.Value); imported {
			sym, err := a.ResolveQualifiedSymbol(unitName.Value, expr.Member.Value)
			return err == nil && sym != nil && isWritableVarSymbol(sym)
		}
	}
	objectType := a.semanticInfo.GetResolvedType(expr.Object)
	baseType := types.GetUnderlyingType(objectType)
	if types.IsJSONVariant(baseType) {
		return true
	}
	if meta, ok := baseType.(*types.ClassOfType); ok {
		baseType = meta.ClassType
	}
	name := ident.Normalize(expr.Member.Value)
	if record, _ := recordReceiverType(baseType); record != nil {
		if writable, found := a.isWritableRecordVarMember(expr, record, name); found {
			return writable
		}
	}
	if class, ok := baseType.(*types.ClassType); ok {
		if writable, found := writableClassVarMember(class, name); found {
			return writable
		}
	}
	if _, classVar := a.hasHelperClassVar(objectType, name); classVar != nil {
		return true
	}
	if property := a.hasHelperProperty(objectType, name); property != nil {
		// Helper accessor metadata retains named readers as method-shaped.
		// A reader naming a helper class variable still provides storage.
		_, classVar := a.hasHelperClassVar(objectType, property.ReadSpec)
		return classVar != nil
	}
	return false
}

// isWritableRecordVarMember distinguishes record instance storage from static
// backing fields. Constants remain immutable through a field-backed property.
func (a *Analyzer) isWritableRecordVarMember(expr *ast.MemberAccessExpression, record *types.RecordType, name string) (bool, bool) {
	if _, constant := record.Constants[name]; constant {
		return false, true
	}
	if _, classVar := record.ClassVars[name]; classVar {
		return true, true
	}
	if property := record.GetProperty(name); property != nil {
		if property.ReadKind != types.PropAccessField {
			return false, true
		}
		backingName := ident.Normalize(property.ReadField)
		if _, constant := record.Constants[backingName]; constant {
			return false, true
		}
		if _, classVar := record.ClassVars[backingName]; classVar {
			return true, true
		}
		return a.isWritableVarArgument(expr.Object, record), true
	}
	if record.GetFieldType(name) != nil {
		return a.isWritableVarArgument(expr.Object, record), true
	}
	return false, false
}

func writableClassVarMember(class *types.ClassType, name string) (bool, bool) {
	if _, constant := class.GetConstant(name); constant {
		return false, true
	}
	if property, found := class.GetProperty(name); found {
		_, constant := class.GetConstant(ident.Normalize(property.ReadSpec))
		return property.ReadKind == types.PropAccessField && !constant, true
	}
	_, field := class.GetField(name)
	_, classVar := class.GetClassVar(name)
	return field || classVar, field || classVar
}

func functionParameterName(signature *types.FunctionType, index int) string {
	if index < len(signature.ParamNames) {
		return signature.ParamNames[index]
	}
	return ""
}
