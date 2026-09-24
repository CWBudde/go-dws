package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// analyzeRangeBound applies the value reading of a routine name: a range
// endpoint is a call result, never a function reference.
func (a *Analyzer) analyzeRangeBound(expr ast.Expression) types.Type {
	if id, ok := expr.(*ast.Identifier); ok {
		if _, declared := a.symbols.Resolve(id.Value); !declared && a.isBuiltinFunction(id.Value) {
			call := &ast.CallExpression{BaseNode: id.BaseNode, Function: id}
			if result, handled := a.analyzeBuiltinFunction(id.Value, nil, call); handled {
				return result
			}
		}
	}
	result := a.analyzeExpression(expr)
	if result != nil {
		if implicit := implicitValueContextType(result); implicit != nil {
			return implicit
		}
	}
	return result
}

// compatibleRangeTypes rejects void even against itself. Case conditions may
// mix numeric types or defer Variant bounds to runtime; array constructors
// and sets require ordinal bounds.
func compatibleRangeTypes(start, end types.Type, allowNumeric bool) bool {
	start, end = types.GetUnderlyingType(start), types.GetUnderlyingType(end)
	if start.Equals(types.VOID) || end.Equals(types.VOID) {
		return false
	}
	if allowNumeric && ((types.IsNumericType(start) && types.IsNumericType(end)) ||
		start.Equals(types.VARIANT) || end.Equals(types.VARIANT)) {
		return true
	}
	return types.IsOrdinalType(start) && types.IsOrdinalType(end) && start.Equals(end)
}

func (a *Analyzer) hintReversedCaseRange(expr *ast.RangeExpression) {
	// Preserve the existing enum-range diagnostics: descending enum bounds
	// simply never match. The reversed-bound hint applies to primitive bounds.
	if boundType := a.semanticInfo.GetResolvedType(expr.Start); boundType != nil {
		if _, isEnum := types.GetUnderlyingType(boundType).(*types.EnumType); isEnum {
			return
		}
	}
	lower, err := a.evaluateConstant(expr.Start)
	if err != nil {
		return
	}
	upper, err := a.evaluateConstant(expr.RangeEnd)
	if err != nil {
		return
	}
	if reversedCaseBounds(lower, upper) {
		pos := expr.Start.Pos()
		a.addHintAt(pos, "Case range condition lower bound is greater than higher bound [line: %d, column: %d]", pos.Line, pos.Column)
	}
}

// reversedCaseBounds compares folded primitive bounds without coercing strings
// or losing precision when both values are integers.
func reversedCaseBounds(lower, upper any) bool {
	reversed := false
	switch low := lower.(type) {
	case int:
		switch high := upper.(type) {
		case int:
			reversed = low > high
		case float64:
			reversed = float64(low) > high
		}
	case float64:
		switch high := upper.(type) {
		case int:
			reversed = low > float64(high)
		case float64:
			reversed = low > high
		}
	case string:
		if high, ok := upper.(string); ok {
			reversed = low > high
		}
	case bool:
		if high, ok := upper.(bool); ok {
			reversed = low && !high
		}
	}
	return reversed
}
