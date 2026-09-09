package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// A representation tag remains authoritative even when an embedding value uses
// a custom display name. Builtins must not interpret its Type string as a tag.
type namedBuiltinValue struct {
	label string
	kind  runtime.ValueKind
}

func (v namedBuiltinValue) Type() string                 { return v.label }
func (v namedBuiltinValue) String() string               { return v.label }
func (v namedBuiltinValue) ValueKind() runtime.ValueKind { return v.kind }

func TestVariantIntrospection_UsesRepresentationKind(t *testing.T) {
	for _, test := range []struct {
		label         string
		code          int64
		kind          runtime.ValueKind
		numeric, text bool
	}{
		{"STRING", varInteger, runtime.KindInteger, true, false},
		{"INTEGER", varString, runtime.KindString, false, true},
		{"custom collection", varArray, runtime.KindArray, false, false},
	} {
		t.Run(test.label, func(t *testing.T) {
			ctx := &mockContext{}
			args := []Value{namedBuiltinValue{kind: test.kind, label: test.label}}
			if got := VarType(ctx, args).(*runtime.IntegerValue).Value; got != test.code {
				t.Errorf("VarType = %d; want %d", got, test.code)
			}
			if got := VarIsNumeric(ctx, args).(*runtime.BooleanValue).Value; got != test.numeric {
				t.Errorf("VarIsNumeric = %t; want %t", got, test.numeric)
			}
			if got := VarIsStr(ctx, args).(*runtime.BooleanValue).Value; got != test.text {
				t.Errorf("VarIsStr = %t; want %t", got, test.text)
			}
		})
	}
}
