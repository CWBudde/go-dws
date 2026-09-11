package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

// jsonValue parses a JSON document and wraps it as a runtime JSON value.
func jsonValue(t *testing.T, src string) Value {
	t.Helper()
	parsed, err := jsonvalue.Parse(src)
	if err != nil {
		t.Fatalf("jsonvalue.Parse(%q): %v", src, err)
	}
	return runtime.NewJSONValue(parsed)
}

// Null, Unassigned and a nil reference are three distinct Variant states in
// DWScript. testdata/fixtures/FunctionsVariant records the observable answers.
func TestVariantEmptinessPredicates(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		value                    Value
		name                     string
		isNull, isEmpty, isClear bool
	}{
		{name: "unassigned", value: &runtime.UnassignedValue{}, isNull: false, isEmpty: true, isClear: true},
		{name: "null", value: &runtime.NullValue{}, isNull: true, isEmpty: false, isClear: false},
		{name: "nil reference", value: &runtime.NilValue{}, isNull: false, isEmpty: false, isClear: false},
		{name: "integer", value: &runtime.IntegerValue{Value: 1}, isNull: false, isEmpty: false, isClear: false},
		{name: "empty string", value: &runtime.StringValue{Value: ""}, isNull: false, isEmpty: false, isClear: false},
		{name: "json null", value: jsonValue(t, "null"), isNull: true, isEmpty: false, isClear: false},
		{name: "json number", value: jsonValue(t, "123"), isNull: false, isEmpty: false, isClear: false},
		{name: "json undefined", value: runtime.NewJSONValue(nil), isNull: false, isEmpty: true, isClear: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []Value{tt.value}
			assertBool(t, "VarIsNull", VarIsNull(ctx, args), tt.isNull)
			assertBool(t, "VarIsEmpty", VarIsEmpty(ctx, args), tt.isEmpty)
			assertBool(t, "VarIsClear", VarIsClear(ctx, args), tt.isClear)
		})
	}
}

// A JSON payload takes part in Variant kind introspection.
func TestVariantKindPredicates_JSON(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name                      string
		json                      string
		isArray, isStr, isNumeric bool
	}{
		{name: "array", json: `[123]`, isArray: true},
		{name: "empty array", json: `[]`, isArray: true},
		{name: "object", json: `{"a":1}`},
		{name: "string", json: `"hello"`, isStr: true},
		{name: "integer", json: `123`, isNumeric: true},
		{name: "float", json: `123.5`, isNumeric: true},
		{name: "boolean", json: `false`},
		{name: "null", json: `null`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []Value{jsonValue(t, tt.json)}
			assertBool(t, "VarIsArray", VarIsArray(ctx, args), tt.isArray)
			assertBool(t, "VarIsStr", VarIsStr(ctx, args), tt.isStr)
			assertBool(t, "VarIsNumeric", VarIsNumeric(ctx, args), tt.isNumeric)
		})
	}
}

// VarAsType accepts every code that aliases the same runtime representation.
func TestVarAsType_TypeCodeAliases(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		source   Value
		want     Value
		name     string
		typeCode int64
	}{
		{name: "varString", typeCode: varString, source: &runtime.IntegerValue{Value: 123}, want: &runtime.StringValue{Value: "123"}},
		{name: "varUString", typeCode: varUString, source: &runtime.IntegerValue{Value: 123}, want: &runtime.StringValue{Value: "123"}},
		{name: "varInteger", typeCode: varInteger, source: &runtime.StringValue{Value: "123"}, want: &runtime.IntegerValue{Value: 123}},
		{name: "varInt64", typeCode: varInt64, source: &runtime.StringValue{Value: "123"}, want: &runtime.IntegerValue{Value: 123}},
		{name: "varDouble", typeCode: varDouble, source: &runtime.IntegerValue{Value: 2}, want: &runtime.FloatValue{Value: 2}},
		{name: "varSingle", typeCode: varSingle, source: &runtime.IntegerValue{Value: 2}, want: &runtime.FloatValue{Value: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VarAsType(ctx, []Value{tt.source, &runtime.IntegerValue{Value: tt.typeCode}})
			if got.Type() != tt.want.Type() || got.String() != tt.want.String() {
				t.Errorf("VarAsType(%s, %d) = %s %q; want %s %q",
					tt.source.String(), tt.typeCode, got.Type(), got.String(), tt.want.Type(), tt.want.String())
			}
		})
	}
}

// Every script-visible constant has to agree with the code VarType reports.
func TestVarTypeConstants_MatchVarType(t *testing.T) {
	ctx := newMockContext()

	codes := map[string]int64{}
	for _, c := range VarTypeConstants() {
		if previous, seen := codes[c.Name]; seen {
			t.Fatalf("duplicate constant %s (%d and %d)", c.Name, previous, c.Value)
		}
		codes[c.Name] = c.Value
	}

	tests := []struct {
		value    Value
		constant string
	}{
		{value: &runtime.StringValue{Value: "str"}, constant: "varString"},
		{value: &runtime.IntegerValue{Value: 123}, constant: "varInt64"},
		{value: &runtime.FloatValue{Value: 12.3}, constant: "varDouble"},
		{value: &runtime.BooleanValue{Value: true}, constant: "varBoolean"},
		{value: &runtime.UnassignedValue{}, constant: "varEmpty"},
	}

	for _, tt := range tests {
		t.Run(tt.constant, func(t *testing.T) {
			want, ok := codes[tt.constant]
			if !ok {
				t.Fatalf("constant %s is not exposed to scripts", tt.constant)
			}
			got := VarType(ctx, []Value{tt.value}).(*runtime.IntegerValue).Value
			if got != want {
				t.Errorf("VarType(%s) = %d; want %s = %d", tt.value.Type(), got, tt.constant, want)
			}
		})
	}
}

func assertBool(t *testing.T, name string, got Value, want bool) {
	t.Helper()
	boolVal, ok := got.(*runtime.BooleanValue)
	if !ok {
		t.Fatalf("%s: expected BooleanValue, got %T", name, got)
	}
	if boolVal.Value != want {
		t.Errorf("%s = %t; want %t", name, boolVal.Value, want)
	}
}
