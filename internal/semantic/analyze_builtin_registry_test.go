package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestBuiltinAnalysis_RegistrySignature(t *testing.T) {
	tests := []struct {
		name      string
		sig       *builtins.FunctionSignature
		args      []ast.Expression
		want      types.Type
		errorText string
	}{
		{name: "new builtin", sig: builtins.Sig([]types.Type{types.INTEGER}, types.STRING), args: []ast.Expression{&ast.IntegerLiteral{Value: 42}}, want: types.STRING},
		{name: "numeric conversion", sig: builtins.Sig([]types.Type{types.FLOAT}, types.FLOAT), args: []ast.Expression{&ast.IntegerLiteral{Value: 42}}, want: types.FLOAT},
		{name: "wrong type", sig: builtins.Sig([]types.Type{types.STRING}, types.INTEGER), args: []ast.Expression{&ast.IntegerLiteral{Value: 42}}, want: types.INTEGER, errorText: "expects String as argument"},
		{name: "wrong arity", sig: builtins.Sig([]types.Type{types.INTEGER}, types.STRING), want: types.STRING, errorText: "expects 1 argument"},
		{name: "optional", sig: builtins.SigOptional([]types.Type{types.INTEGER, types.STRING}, types.INTEGER, 1), args: []ast.Expression{&ast.IntegerLiteral{Value: 42}}, want: types.INTEGER},
		{name: "variadic procedure", sig: builtins.SigVariadic(nil, nil, 0), args: []ast.Expression{&ast.IntegerLiteral{Value: 42}, &ast.StringLiteral{Value: "text"}}, want: types.VOID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer()
			a.builtinRegistry = builtins.NewRegistry()
			a.builtinRegistry.RegisterWithSignature("RegisteredOnly", nil, builtins.CategorySystem, "test", tt.sig)
			call := &ast.CallExpression{Function: &ast.Identifier{Value: "rEgIsTeReDoNlY"}, Arguments: tt.args}
			got, recognized := a.analyzeBuiltinFunction("rEgIsTeReDoNlY", tt.args, call)
			if !recognized || got != tt.want {
				t.Fatalf("analysis = %v, %v; want %v, true", got, recognized, tt.want)
			}
			errors := strings.Join(a.Errors(), "\n")
			if tt.errorText == "" && errors != "" {
				t.Fatalf("unexpected errors: %s", errors)
			}
			if tt.errorText != "" && !strings.Contains(errors, tt.errorText) {
				t.Fatalf("errors = %q; want %q", errors, tt.errorText)
			}
			if got, ok := a.getBuiltinReturnType("REGISTEREDONLY"); !ok || got != tt.want {
				t.Fatalf("return type = %v, %v", got, ok)
			}
		})
	}
}

func TestBuiltinAnalysis_UnknownAndUnsigned(t *testing.T) {
	a := NewAnalyzer()
	a.builtinRegistry = builtins.NewRegistry()
	a.builtinRegistry.Register("Unsigned", nil, builtins.CategorySystem, "test")
	for _, name := range []string{"Unknown", "Unsigned"} {
		call := &ast.CallExpression{Function: &ast.Identifier{Value: name}}
		if _, ok := a.analyzeBuiltinFunction(name, nil, call); ok {
			t.Errorf("recognized %s without signature", name)
		}
		if _, ok := a.getBuiltinReturnType(name); ok {
			t.Errorf("return type recognized %s without signature", name)
		}
	}
}

func TestBuiltinAnalysis_ArgumentDependentResults(t *testing.T) {
	tests := []struct{ name, code string }{
		{"map array", `var a: array of Integer; var b := Map(a, lambda(x: Integer): Integer => x); var n := Length(b);`},
		{"filter array", `var a: array of Integer; var b := Filter(a, lambda(x: Integer): Boolean => x > 0); var n := Length(b);`},
		{"slice array", `var a: array of Integer; var b := Slice(a, 0, 0); var n := Length(b);`},
		{"reduce accumulator", `var a: array of String; var b := Reduce(a, lambda(acc: String; x: String): String => acc + x, ''); var n := Length(b);`},
		{"find element", `var a: array of String; var b := Find(a, lambda(x: String): Boolean => x = ''); var n := Length(b);`},
		{"call stack array", `var frames := GetCallStack(); var n := Length(frames);`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runBuiltinTest(t, tt.code, false, "") })
	}
}

func TestBuiltinAnalysis_SpecializedDiagnostics(t *testing.T) {
	tests := []struct{ name, code, want string }{
		{"filter arity", `Filter();`, "expects 2 arguments (array, predicate)"},
		{"reduce arity", `Reduce();`, "expects 3 arguments (array, lambda, initial)"},
		{"slice arity", `Slice();`, "expects 3 arguments (array, start, end)"},
		{"assert condition", `Assert(1);`, "first argument must be Boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runBuiltinTest(t, tt.code, true, tt.want) })
	}
}

func TestBuiltinAnalysis_DeclaredResultTypes(t *testing.T) {
	tests := []struct {
		name string
		want types.Type
	}{
		{"Add", types.VOID},
		{"StrArrayPack", types.NewDynamicArrayType(types.STRING)},
		{"TestBit", types.BOOLEAN},
		{"Clamp", types.FLOAT},
		{"ClampInt", types.INTEGER},
		{"MinInt", types.INTEGER},
		{"MaxInt", types.INTEGER},
		{"Abs", types.VARIANT},
		{"Sqr", types.VARIANT},
		{"Copy", types.VARIANT},
		{"Concat", types.VARIANT},
	}
	a := NewAnalyzer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := a.getBuiltinReturnType(tt.name)
			if !ok || !tt.want.Equals(got) {
				t.Fatalf("declared result = %v, %v; want %v", got, ok, tt.want)
			}
		})
	}
}
