package semantic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// These cases pin the result, accepted arguments, and diagnostic wording of the
// specialized analyzers before their migration to the shared registry.
func TestBuiltinAnalysis_MigratedSignatures(t *testing.T) {
	tests := []struct {
		name         string
		params       string
		result       types.Type
		arity        string
		expectations []string
	}{
		{"StrToFloat", "S", types.FLOAT, "1 argument", []string{"String as argument"}},
		{"VarToStr", "V", types.STRING, "1 argument", []string{""}},
		{"StrToBool", "S", types.BOOLEAN, "1 argument", []string{"String as argument"}},
		{"HexToInt", "S", types.INTEGER, "1 argument", []string{"String as argument"}},
		{"BinToInt", "S", types.INTEGER, "1 argument", []string{"String as argument"}},
		{"Sqrt", "F", types.FLOAT, "1 argument", []string{"Integer or Float as argument"}},
		{"Exp", "F", types.FLOAT, "1 argument", []string{"Integer or Float as argument"}},
		{"Ln", "F", types.FLOAT, "1 argument", []string{"Integer or Float as argument"}},
		{"Log2", "F", types.FLOAT, "1 argument", []string{"Integer or Float as argument"}},
		{"Trunc", "F", types.INTEGER, "1 argument", []string{"Integer or Float as argument"}},
		{"Ceil", "F", types.INTEGER, "1 argument", []string{"Integer or Float as argument"}},
		{"Floor", "F", types.INTEGER, "1 argument", []string{"Integer or Float as argument"}},
		{"UnixTime", "", types.INTEGER, "0 arguments", []string{}},
		{"UnixTimeMSec", "", types.INTEGER, "0 arguments", []string{}},
		{"Assigned", "V", types.BOOLEAN, "1 argument", []string{""}},
		{"ToJSON", "V", types.STRING, "1 argument", []string{""}},
		{"JSONLength", "V", types.INTEGER, "1 argument", []string{""}},
		{"Log10", "F", types.FLOAT, "1 argument", []string{"Float or Integer"}},
		{"Frac", "F", types.FLOAT, "1 argument", []string{"Float or Integer"}},
		{"Int", "F", types.FLOAT, "1 argument", []string{"Float or Integer"}},
		{"IsFinite", "F", types.BOOLEAN, "1 argument", []string{"Float or Integer"}},
		{"IsInfinite", "F", types.BOOLEAN, "1 argument", []string{"Float or Integer"}},
		{"Unsigned32", "I", types.INTEGER, "1 argument", []string{"Integer argument"}},
		{"Factorial", "I", types.INTEGER, "1 argument", []string{"Integer argument"}},
		{"IsPrime", "I", types.BOOLEAN, "1 argument", []string{"Integer argument"}},
		{"LeastFactor", "I", types.INTEGER, "1 argument", []string{"Integer argument"}},
		{"PopCount", "I", types.INTEGER, "1 argument", []string{"Integer argument"}},
		{"RandomInt", "I", types.INTEGER, "1 argument", []string{"Integer argument"}},
		{"Odd", "I", types.BOOLEAN, "1 argument", []string{"Integer"}},
		{"Pi", "", types.FLOAT, "no arguments", []string{}},
		{"Infinity", "", types.FLOAT, "no arguments", []string{}},
		{"NaN", "", types.FLOAT, "no arguments", []string{}},
		{"Random", "", types.FLOAT, "no arguments", []string{}},
		{"RandSeed", "", types.INTEGER, "no arguments", []string{}},
		{"UpperCase", "S", types.STRING, "1 argument", []string{"string as argument"}},
		{"LowerCase", "S", types.STRING, "1 argument", []string{"string as argument"}},
		{"StrBeginsWith", "SS", types.BOOLEAN, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrEndsWith", "SS", types.BOOLEAN, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrContains", "SS", types.BOOLEAN, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"PosEx", "SSI", types.INTEGER, "3 arguments", []string{"string as first argument", "string as second argument", "integer as third argument"}},
		{"RevPos", "SS", types.INTEGER, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"IsDelimiter", "SSI", types.BOOLEAN, "3 arguments", []string{"string as first argument", "string as second argument", "integer as third argument"}},
		{"LastDelimiter", "SS", types.INTEGER, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"SameText", "SS", types.BOOLEAN, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"CompareText", "SS", types.INTEGER, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"CompareStr", "SS", types.INTEGER, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"AnsiCompareText", "SS", types.INTEGER, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"AnsiCompareStr", "SS", types.INTEGER, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrMatches", "SS", types.BOOLEAN, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"RightStr", "SI", types.STRING, "2 arguments", []string{"string as first argument", "integer as second argument"}},
		{"StrBefore", "SS", types.STRING, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrBeforeLast", "SS", types.STRING, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrAfter", "SS", types.STRING, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrAfterLast", "SS", types.STRING, "2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StrBetween", "SSS", types.STRING, "3 arguments", []string{"string as first argument", "string as second argument", "string as third argument"}},
		{"PadLeft", "SIS", types.STRING, "2 or 3 arguments", []string{"string as first argument", "integer as second argument", "string as third argument"}},
		{"PadRight", "SIS", types.STRING, "2 or 3 arguments", []string{"string as first argument", "integer as second argument", "string as third argument"}},
		{"StrDeleteLeft", "SI", types.STRING, "2 arguments", []string{"string as first argument", "integer as second argument"}},
		{"StrDeleteRight", "SI", types.STRING, "2 arguments", []string{"string as first argument", "integer as second argument"}},
		{"ReverseString", "S", types.STRING, "1 argument", []string{"string as argument"}},
		{"QuotedStr", "SS", types.STRING, "1 or 2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StringOfString", "SI", types.STRING, "2 arguments", []string{"string as first argument", "integer as second argument"}},
		{"NormalizeString", "SS", types.STRING, "1 or 2 arguments", []string{"string as first argument", "string as second argument"}},
		{"StripAccents", "S", types.STRING, "1 argument", []string{"string as argument"}},
		{"StrIsASCII", "S", types.BOOLEAN, "1 argument", []string{"string as argument"}},
		{"ByteSizeToStr", "I", types.STRING, "1 argument", []string{"integer as argument"}},
		{"GetText", "S", types.STRING, "1 argument", []string{"string as argument"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := builtinMigrationArguments(tt.params)
			analyze := func(args []ast.Expression) []string {
				t.Helper()
				a := NewAnalyzer()
				name := ident.Normalize(tt.name)
				call := &ast.CallExpression{Function: &ast.Identifier{Value: name}, Arguments: args}
				got, ok := a.analyzeBuiltinFunction(name, args, call)
				if !ok || !tt.result.Equals(got) {
					t.Fatalf("result = %v, %v; want %v", got, ok, tt.result)
				}
				return a.Errors()
			}
			if errs := analyze(args); len(errs) != 0 {
				t.Fatalf("valid arguments: %v", errs)
			}
			// A Boolean is invalid for every typed parameter in this family. Verify
			// errors from each position separately, including optional parameters.
			for i, expectation := range tt.expectations {
				if expectation == "" {
					continue
				}
				invalid := append([]ast.Expression(nil), args...)
				invalid[i] = &ast.BooleanLiteral{Value: true}
				want := fmt.Sprintf("function '%s' expects %s, got Boolean at 0:0", tt.name, expectation)
				if got := strings.Join(analyze(invalid), "\n"); got != want {
					t.Errorf("argument %d: got %q; want %q", i+1, got, want)
				}
			}
			invalid := append([]ast.Expression(nil), args...)
			invalid = append(invalid, &ast.BooleanLiteral{Value: true})
			want := fmt.Sprintf("function '%s' expects %s, got %d at 0:0", tt.name, tt.arity, len(invalid))
			if got := strings.Join(analyze(invalid), "\n"); got != want {
				t.Errorf("arity: got %q; want %q", got, want)
			}
			if strings.Contains(tt.arity, " or ") {
				if errs := analyze(args[:len(args)-1]); len(errs) != 0 {
					t.Errorf("optional arguments: %v", errs)
				}
			}
			t.Run("registry owns signature", func(t *testing.T) {
				a := NewAnalyzer()
				a.builtinRegistry = builtins.NewRegistry()
				a.builtinRegistry.RegisterWithSignature(tt.name, nil, builtins.CategorySystem, "test",
					builtins.Sig([]types.Type{types.BOOLEAN}, types.STRING))
				args := []ast.Expression{&ast.BooleanLiteral{Value: true}}
				call := &ast.CallExpression{Function: &ast.Identifier{Value: tt.name}, Arguments: args}
				got, ok := a.analyzeBuiltinFunction(tt.name, args, call)
				if !ok || got != types.STRING || len(a.Errors()) != 0 {
					t.Fatalf("replacement signature: result = %v, %v; errors = %v", got, ok, a.Errors())
				}
			})
		})
	}
}

func builtinMigrationArguments(params string) []ast.Expression {
	args := make([]ast.Expression, len(params))
	for i, kind := range params {
		switch kind {
		case 'S':
			args[i] = &ast.StringLiteral{Value: "text"}
		case 'I', 'F':
			args[i] = &ast.IntegerLiteral{Value: 2}
		case 'V':
			args[i] = &ast.BooleanLiteral{Value: true}
		}
	}
	return args
}

func TestBuiltinAnalysis_MigratedAliases(t *testing.T) {
	for _, tt := range []struct{ name, canonical string }{
		{"ASCIIUpperCase", "UpperCase"}, {"AnsiUpperCase", "UpperCase"},
		{"ASCIILowerCase", "LowerCase"}, {"AnsiLowerCase", "LowerCase"},
		{"DeleteLeft", "StrDeleteLeft"}, {"DeleteRight", "StrDeleteRight"},
		{"DupeString", "StringOfString"}, {"Normalize", "NormalizeString"},
		{"_", "GetText"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer()
			sig, ok := a.builtinRegistry.GetSignature(tt.name)
			if !ok {
				t.Fatal("missing alias signature")
			}
			args := make([]ast.Expression, sig.MinArgs)
			for i := range args {
				args[i] = &ast.BooleanLiteral{Value: true}
			}
			call := &ast.CallExpression{Function: &ast.Identifier{Value: tt.name}, Arguments: args}
			a.analyzeBuiltinFunction(tt.name, args, call)
			for _, err := range a.Errors() {
				if !strings.HasPrefix(err, "function '"+tt.canonical+"' expects ") {
					t.Errorf("alias changed diagnostic name: %s", err)
				}
			}
			if len(a.Errors()) != len(args) {
				t.Errorf("errors = %v; want one per argument", a.Errors())
			}
		})
	}
}

// LeftStr's first parameter permits Variant conversion; its count parameter
// remains strictly Integer. Keep this on the specialized path until signatures
// can express conversion rules separately for each parameter.
func TestBuiltinAnalysis_LeftStrVariant(t *testing.T) {
	for _, code := range []string{
		`var text: Variant := 'abc'; PrintLn(LeftStr(text, 2));`,
		`type TText = Variant; var text: TText := 'abc'; PrintLn(LeftStr(text, 2));`,
	} {
		runBuiltinTest(t, code, false, "")
	}
	runBuiltinTest(t, `var count: Variant := 2; PrintLn(LeftStr('abc', count));`, true,
		"function 'LeftStr' expects integer as second argument, got Variant")
}
