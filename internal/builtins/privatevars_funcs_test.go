package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

type privateVarContext struct {
	*mockContext
	unit string
}

func (c *privateVarContext) CurrentUnit() string { return c.unit }

func TestPrivateVarBindings_MainModule(t *testing.T) {
	name := &runtime.StringValue{Value: "test"}
	for _, tc := range []struct {
		name string
		fn   BuiltinFunc
		args []Value
	}{
		{"ReadPrivateVar", ReadPrivateVar, []Value{name}},
		{"WritePrivateVar", WritePrivateVar, []Value{name, name}},
		{"PrivateVarsNames", PrivateVarsNames, []Value{name}},
		{"CleanupPrivateVars", CleanupPrivateVars, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &privateVarContext{mockContext: newMockContext()}
			got := tc.fn(ctx, tc.args)
			if got.Type() != "ERROR" || ctx.lastError != "Private variables cannot be referred from main module" {
				t.Fatalf("got %v (%q)", got, ctx.lastError)
			}
		})
	}
}

func TestPrivateVarBindings_WriteRead(t *testing.T) {
	ctx := &privateVarContext{mockContext: newMockContext(), unit: t.Name()}
	CleanupPrivateVars(ctx, nil)
	t.Cleanup(func() { CleanupPrivateVars(ctx, nil) })
	name := &runtime.StringValue{Value: "test"}
	value := &runtime.StringValue{Value: "hello"}
	if got := WritePrivateVar(ctx, []Value{name, value}); got.String() != "True" {
		t.Fatalf("new write = %v", got)
	}
	if got := WritePrivateVar(ctx, []Value{name, value}); got.String() != "False" {
		t.Fatalf("replacement write = %v", got)
	}
	if got := ReadPrivateVar(ctx, []Value{name}); got.String() != "hello" {
		t.Fatalf("read = %v", got)
	}
	missing := &runtime.StringValue{Value: "missing"}
	if got := ReadPrivateVar(ctx, []Value{missing, value}); got.String() != "hello" {
		t.Fatalf("default = %v", got)
	}
	if got := ReadPrivateVar(ctx, []Value{missing}); runtime.KindOf(got) != runtime.KindUnassigned {
		t.Fatalf("missing read = %v", got)
	}
}
