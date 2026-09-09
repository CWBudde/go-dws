package semantic

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// These cases were recorded from the specialized analyzers before registry migration.
// The argument alphabet exercises scalar types, aliases, arrays, and unresolved names.
func TestBuiltinAnalysis_ArchitectureCompatibility(t *testing.T) {
	data, err := os.ReadFile("testdata/builtin_analysis_compatibility.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string][]struct {
		Arguments string   `json:"arguments"`
		Result    []string `json:"result"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for name, tests := range cases {
		t.Run(name, func(t *testing.T) {
			for _, test := range tests {
				for _, pattern := range strings.Split(test.Arguments, ",") {
					want := test.Result
					if got := builtinCompatibilityResult(name, pattern); !reflect.DeepEqual(got, want) {
						t.Errorf("arguments %q: got %q; want %q", pattern, got, want)
					}
				}
			}
		})
	}
}

func builtinCompatibilityResult(name, pattern string) []string {
	a := NewAnalyzer()
	values := map[rune]types.Type{
		'I': types.INTEGER, 'F': types.FLOAT, 'S': types.STRING, 'B': types.BOOLEAN,
		'V': types.VARIANT, 'J': types.JSON_VARIANT,
		'A': types.NewDynamicArrayType(types.STRING),
		'D': types.NewStaticArrayType(types.STRING, 1, 3),
		'G': types.NewDynamicArrayType(types.INTEGER),
		'i': &types.TypeAlias{Name: "IntAlias", AliasedType: types.INTEGER},
		'v': &types.TypeAlias{Name: "VariantAlias", AliasedType: types.VARIANT},
	}
	args := make([]ast.Expression, 0, len(pattern))
	for index, kind := range pattern {
		name := fmt.Sprintf("argument%d", index)
		if typ := values[kind]; typ != nil {
			a.symbols.Define(name, typ, token.Position{})
		}
		args = append(args, &ast.Identifier{Value: name})
	}
	call := &ast.CallExpression{Function: &ast.Identifier{Value: name}, Arguments: args}
	result, ok := a.analyzeBuiltinFunction(name, args, call)
	resultName := "<nil>"
	if result != nil {
		resultName = result.String()
	}
	return append([]string{fmt.Sprintf("%t %s", ok, resultName)}, a.Errors()...)
}

func TestBuiltinAnalysis_RegistryOwnsRemainingSignatures(t *testing.T) {
	for name := range builtinDiagnosticStyles {
		t.Run(name, func(t *testing.T) {
			a := NewAnalyzer()
			a.builtinRegistry = builtins.NewRegistry()
			a.builtinRegistry.RegisterWithSignature(name, nil, builtins.CategorySystem, "replacement",
				builtins.Sig([]types.Type{types.BOOLEAN}, types.STRING))
			args := []ast.Expression{&ast.BooleanLiteral{Value: true}}
			call := &ast.CallExpression{Function: &ast.Identifier{Value: name}, Arguments: args}
			result, ok := a.analyzeBuiltinFunction(name, args, call)
			if !ok || result != types.STRING || len(a.Errors()) != 0 {
				t.Fatalf("replacement signature: result = %v, %t; errors = %v", result, ok, a.Errors())
			}
		})
	}
}

func TestBuiltinAnalysis_CorrectedSignatureConsumers(t *testing.T) {
	a := NewAnalyzer()
	if got, ok := a.parameterlessBuiltinType("rAnDg"); !ok || got != types.FLOAT {
		t.Fatalf("implicit RandG: %v, %t", got, ok)
	}
	for _, test := range []struct {
		result types.Type
		name   string
		params int
	}{
		{types.FLOAT, "RandG", 0},
		{types.NewDynamicArrayType(types.STRING), "JSONKeys", 1},
		{types.NewDynamicArrayType(types.VARIANT), "JSONValues", 1},
		{types.NewDynamicArrayType(types.STRING), "StrSplit", 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, ok := a.getBuiltinReturnType(test.name)
			if !ok || !test.result.Equals(result) {
				t.Fatalf("return lookup: %v, %t", result, ok)
			}
			pointer := a.analyzeAddressOfFunction(test.name, &ast.AddressOfExpression{})
			fn, ok := pointer.(*types.FunctionPointerType)
			if !ok || len(fn.Parameters) != test.params || !test.result.Equals(fn.ReturnType) {
				t.Fatalf("function pointer: %v", pointer)
			}
		})
	}
}
