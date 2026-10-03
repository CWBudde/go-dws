package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func isStaticArraySizeMismatch(expected, got types.Type) bool {
	expectedArray, ok := types.GetUnderlyingType(expected).(*types.ArrayType)
	if !ok || !expectedArray.IsStatic() {
		return false
	}
	gotArray, ok := types.GetUnderlyingType(got).(*types.ArrayType)
	if !ok || !gotArray.IsStatic() {
		return false
	}
	return !expectedArray.Equals(gotArray)
}

func allocationMismatchPos(value ast.Expression, fallback lexer.Position) lexer.Position {
	if allocation, ok := value.(*ast.NewArrayExpression); ok && allocation.LBracketPos.Line > 0 {
		return allocation.LBracketPos
	}
	return fallback
}

func isEmptyBracketLiteral(value ast.Expression) bool {
	switch literal := value.(type) {
	case *ast.ArrayLiteralExpression:
		return len(literal.Elements) == 0
	case *ast.SetLiteral:
		return len(literal.Elements) == 0
	default:
		return false
	}
}

// isEnumMismatch reports an assignment between two enumeration types, which
// DWScript anchors at the assignment operator rather than the value.
func isEnumMismatch(expected, got types.Type) bool {
	_, expectedEnum := types.GetUnderlyingType(expected).(*types.EnumType)
	_, gotEnum := types.GetUnderlyingType(got).(*types.EnumType)
	return expectedEnum && gotEnum
}

func assignmentMismatchPos(value ast.Expression, fallback lexer.Position, expected, got types.Type) lexer.Position {
	if pos, ok := castAssignmentPos(value); ok {
		return pos
	}
	if _, expectedMeta := types.GetUnderlyingType(expected).(*types.RecordMetaType); expectedMeta {
		return assignmentSupplierPos(value, fallback)
	}
	if _, gotMeta := types.GetUnderlyingType(got).(*types.RecordMetaType); gotMeta {
		return assignmentSupplierPos(value, fallback)
	}
	if pos := allocationMismatchPos(value, fallback); pos != fallback {
		return pos
	}
	if value != nil {
		if _, ok := types.GetUnderlyingType(expected).(*types.ArrayType); ok {
			if _, ok := types.GetUnderlyingType(got).(*types.ArrayType); ok {
				return value.Pos()
			}
		}
	}
	if isStaticArraySizeMismatch(expected, got) && value != nil {
		return value.Pos()
	}
	if value != nil && !isEnumMismatch(expected, got) {
		_, expectedArray := types.GetUnderlyingType(expected).(*types.ArrayType)
		_, gotArray := types.GetUnderlyingType(got).(*types.ArrayType)
		if !expectedArray && !gotArray {
			return value.Pos()
		}
	}
	return fallback
}

// reportAssignmentTypeMismatch keeps conversion-specific diagnostics at the
// assignment operator while ordinary type mismatches use the expression's anchor.
func (a *Analyzer) reportAssignmentTypeMismatch(valuePos, assignmentPos lexer.Position, from, to types.Type) {
	if a.reportClassInterfaceAssignmentMismatch(assignmentPos, from, to) {
		return
	}
	if isJSONReferenceBoxingMismatch(from, to) {
		valuePos = assignmentPos
	}
	a.addError("%s", errors.FormatCannotAssign(semanticTypeNameForDiagnostic(from), semanticTypeNameForDiagnostic(to), valuePos.Line, valuePos.Column))
}

// Routine and class references cannot be boxed as JSON values. Preserve other
// conversion categories' own anchors, including record metatypes and arrays.
func isJSONReferenceBoxingMismatch(from, to types.Type) bool {
	if !types.IsJSONVariant(to) {
		return false
	}
	switch types.GetUnderlyingType(from).(type) {
	case *types.FunctionType, *types.FunctionPointerType, *types.MethodPointerType, *types.ClassOfType:
		return true
	default:
		return false
	}
}

func (a *Analyzer) reportClassInterfaceAssignmentMismatch(pos lexer.Position, from, to types.Type) bool {
	class, fromClass := types.GetUnderlyingType(from).(*types.ClassType)
	iface, toInterface := types.GetUnderlyingType(to).(*types.InterfaceType)
	if fromClass && toInterface {
		a.addStructuredError(NewClassDoesNotImplementInterfaceError(pos, class.Name, iface.Name))
		return true
	}
	return false
}

func (a *Analyzer) analyzeAssignmentValue(stmt *ast.AssignmentStatement, expected types.Type) types.Type {
	valueType := a.analyzeAssignmentExpression(stmt.Value, expected, true)
	if a.reportAssignmentValueRecovery(stmt, expected, valueType) {
		return nil
	}
	return valueType
}

func (a *Analyzer) arrayAssignmentMismatchPos(value ast.Expression, fallback lexer.Position, expected, got types.Type) lexer.Position {
	if identifier, ok := value.(*ast.Identifier); ok {
		if symbol, found := a.symbols.Resolve(identifier.Value); found && symbol.IsConst {
			if _, isArray := types.GetUnderlyingType(got).(*types.ArrayType); isArray {
				return fallback
			}
		}
	}
	if array, ok := types.GetUnderlyingType(expected).(*types.ArrayType); ok && array.IsDynamic() && isBracketLiteral(value) {
		return fallback
	}
	return assignmentMismatchPos(value, fallback, expected, got)
}

// assignmentTargetMismatchPos applies the array anchoring rules of
// arrayAssignmentMismatchPos to array-to-array mismatches and otherwise keeps
// the path's own anchor for other mismatches.
func (a *Analyzer) assignmentTargetMismatchPos(value ast.Expression, assignmentPos, otherPos lexer.Position, expected, got types.Type) lexer.Position {
	if pos, ok := castAssignmentPos(value); ok {
		return pos
	}
	_, expectedArray := types.GetUnderlyingType(expected).(*types.ArrayType)
	_, gotArray := types.GetUnderlyingType(got).(*types.ArrayType)
	if value != nil && expectedArray && gotArray {
		return a.arrayAssignmentMismatchPos(value, assignmentPos, expected, got)
	}
	return allocationMismatchPos(value, otherPos)
}

// rejectBareTypeValue reports DWScript's `"(" expected` when a plain assignment's value
// is a bare non-class type name (`v := TEnum`, `obj.F := TEnum`, `a[0] := TEnum`):
// upstream reads the name as the start of a cast and stops right after it, whatever the
// target is. It reports whether the error was recorded.
func (a *Analyzer) rejectBareTypeValue(stmt *ast.AssignmentStatement, isCompound bool) bool {
	if isCompound || !a.isBareTypeValue(stmt.Value) {
		return false
	}
	pos := stmt.Value.End()
	a.addError("Syntax Error: \"(\" expected [line: %d, column: %d]", pos.Line, pos.Column)
	return true
}

// isReadOnlyArrayIndexTarget reports whether an indexed assignment target such
// as `arr[0] := x` writes into a static array bound to a read-only symbol — a
// `const` parameter or a declared constant.
//
// Two cases are deliberately excluded, matching DWScript:
//   - Open and dynamic arrays. A `const` open-array parameter only pins the
//     reference, so element assignment is legal (FailureScripts/array_of_const
//     expects no error for `procedure Test1(const AInts: array of Integer)`).
//   - Chains rooted at a member access (`obj.Items[0]`), which mutate the
//     referenced object rather than the const binding itself.
func (a *Analyzer) isReadOnlyArrayIndexTarget(target *ast.IndexExpression, baseType types.Type) bool {
	arrayType, ok := types.GetUnderlyingType(baseType).(*types.ArrayType)
	if !ok || !arrayType.IsStatic() {
		return false
	}

	root := ast.Expression(target)
	for {
		idx, ok := root.(*ast.IndexExpression)
		if !ok {
			break
		}
		root = idx.Left
	}

	rootIdent, ok := root.(*ast.Identifier)
	if !ok {
		return false
	}

	sym, ok := a.symbols.Resolve(rootIdent.Value)
	if !ok {
		return false
	}

	return sym.ReadOnly || sym.IsConst
}

// castAssignmentPos preserves the AS diagnostic anchor through grouping.
func castAssignmentPos(value ast.Expression) (lexer.Position, bool) {
	for {
		switch expr := value.(type) {
		case *ast.GroupedExpression:
			value = expr.Expression
		case *ast.AsExpression:
			return expr.Token.Pos, true
		default:
			return lexer.Position{}, false
		}
	}
}

// assignmentSupplierPos identifies the member producing a value, rather than
// its receiver, while retaining the expression anchor for other suppliers.
func assignmentSupplierPos(value ast.Expression, fallback lexer.Position) lexer.Position {
	switch expr := value.(type) {
	case *ast.GroupedExpression:
		return assignmentSupplierPos(expr.Expression, fallback)
	case *ast.MethodCallExpression:
		return expr.Method.Pos()
	case *ast.MemberAccessExpression:
		return expr.Member.Pos()
	case *ast.CallExpression:
		return assignmentSupplierPos(expr.Function, fallback)
	}
	if value != nil {
		return value.Pos()
	}
	return fallback
}
