package evaluator

import (
	"cmp"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func isComparisonOperator(op string) bool {
	return op == "=" || op == "<>" || op == "<" || op == "<=" || op == ">" || op == ">="
}

func isUndefinedJSON(value Value) bool {
	j, ok := value.(*runtime.JSONValue)
	return ok && j.Value.Kind() == jsonvalue.KindUndefined
}

// jsonComparisonScalar mirrors the connector's ToVariant conversion. It is
// deliberately comparison-only: arithmetic and JSON ownership keep their values.
func jsonComparisonScalar(value Value) Value {
	j, ok := value.(*runtime.JSONValue)
	if !ok {
		return value
	}
	switch j.Value.Kind() {
	case jsonvalue.KindInt64:
		return &runtime.IntegerValue{Value: j.Value.Int64Value()}
	case jsonvalue.KindNumber:
		return &runtime.FloatValue{Value: j.Value.NumberValue()}
	case jsonvalue.KindBoolean:
		return &runtime.BooleanValue{Value: j.Value.BoolValue()}
	case jsonvalue.KindNull:
		return &runtime.NullValue{}
	case jsonvalue.KindUndefined:
		return &runtime.UnassignedValue{}
	default:
		return &runtime.StringValue{Value: j.String()}
	}
}

// compareVariantScalars returns an unordered result for a number compared with
// a nonnumeric string, as VarCompareSafe does. Two strings remain lexical.
func compareVariantScalars(left, right Value) (order int, ordered, handled bool) {
	_, leftString := left.(*runtime.StringValue)
	_, rightString := right.(*runtime.StringValue)
	if leftString && rightString {
		return strings.Compare(left.String(), right.String()), true, true
	}
	leftNumeric := isNumericKind(runtime.KindOf(left))
	rightNumeric := isNumericKind(runtime.KindOf(right))
	numericPair := (leftNumeric && (rightNumeric || rightString)) || (rightNumeric && leftString)
	if !numericPair {
		return 0, false, false
	}
	l, lok := parseVariantComparisonNumber(left)
	r, rok := parseVariantComparisonNumber(right)
	if !lok || !rok {
		return 0, false, true
	}
	li, lint := l.(*runtime.IntegerValue)
	ri, rint := r.(*runtime.IntegerValue)
	if lint && rint {
		return cmp.Compare(li.Value, ri.Value), true, true
	}
	lf, rf := variantComparisonFloat(l), variantComparisonFloat(r)
	if math.IsNaN(lf) || math.IsNaN(rf) {
		return 0, false, true
	}
	return cmp.Compare(lf, rf), true, true
}

func comparisonResult(op string, order int, ordered bool) Value {
	var result bool
	switch op {
	case "=":
		result = ordered && order == 0
	case "<>":
		result = !ordered || order != 0
	case "<":
		result = ordered && order < 0
	case "<=":
		result = ordered && order <= 0
	case ">":
		result = ordered && order > 0
	case ">=":
		result = ordered && order >= 0
	}
	return &runtime.BooleanValue{Value: result}
}

func isVariantComparisonValue(value Value) bool {
	kind := runtime.KindOf(value)
	return kind == runtime.KindVariant || kind == runtime.KindJSON
}

func (e *Evaluator) membershipEqual(left, right Value, node ast.Node) Value {
	if isVariantComparisonValue(left) || isVariantComparisonValue(right) {
		return e.evalVariantBinaryOp("=", left, right, node)
	}
	return &runtime.BooleanValue{Value: ValuesEqual(left, right)}
}

func (e *Evaluator) membershipRangeBound(op string, left, right Value, node ast.Node) Value {
	l, lerr := runtime.GetOrdinalValue(jsonComparisonScalar(unwrapVariant(left)))
	r, rerr := runtime.GetOrdinalValue(jsonComparisonScalar(unwrapVariant(right)))
	if lerr == nil && rerr == nil {
		return comparisonResult(op, cmp.Compare(l, r), true)
	}
	return e.evalVariantBinaryOp(op, left, right, node)
}

func parseVariantComparisonNumber(v Value) (Value, bool) {
	s, ok := v.(*runtime.StringValue)
	if !ok {
		return v, true
	}
	text := strings.TrimSpace(s.Value)
	if i, err := strconv.ParseInt(text, 10, 64); err == nil {
		return &runtime.IntegerValue{Value: i}, true
	}
	f, err := strconv.ParseFloat(text, 64)
	return &runtime.FloatValue{Value: f}, err == nil
}

func variantComparisonFloat(value Value) float64 {
	switch v := value.(type) {
	case *runtime.IntegerValue:
		return float64(v.Value)
	case *runtime.FloatValue:
		return v.Value
	default:
		return math.NaN()
	}
}

func (e *Evaluator) membershipRange(left Value, interval *ast.RangeExpression, ctx *ExecutionContext) Value {
	low := e.Eval(interval.Start, ctx)
	if isError(low) {
		return low
	}
	high := e.Eval(interval.RangeEnd, ctx)
	if isError(high) {
		return high
	}
	ge := e.membershipRangeBound(">=", left, low, interval)
	if isError(ge) {
		return ge
	}
	le := e.membershipRangeBound("<=", left, high, interval)
	if isError(le) {
		return le
	}
	return &runtime.BooleanValue{Value: VariantToBool(ge) && VariantToBool(le)}
}
