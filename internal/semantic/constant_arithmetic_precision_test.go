package semantic

import (
	"math"
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Inspect the constant declaration produced by real source analysis: these
// values also feed array bounds, subsequent constants and default snapshots.
func TestConstantArithmetic_PrecisionAndNumericContracts(t *testing.T) {
	for _, tt := range []struct {
		want       interface{}
		expression string
	}{
		{int(9007199254740993), "9007199254740993 + 0"},
		{int(9007199254740993), "9007199254740991 + 2"},
		{int(9007199254740993), "9007199254740993 - 0"},
		{int(9000000006000000001), "3000000001 * 3000000001"},
		{int(9223372036854775807), "9223372036854775807 - 0"},
		{int(-9223372036854775806), "-9223372036854775807 + 1"},
		{int(-9223372036854775808), "9223372036854775807 + 1"},
		{int(9007199254740993), "9007199254740993 div 1"},
		{int(1), "9007199254740993 mod 2"},
		{int(-4503599627370496), "-9007199254740993 div 2"},
		{int(-1), "-9007199254740993 mod 2"},
		{int(1), "9007199254740993 mod -2"},
		{float64(1.5), "1 + 0.5"},
		{float64(1.5), "0.5 + 1"},
		{float64(0.5), "1.5 - 1"},
		{float64(1.5), "1 * 1.5"},
		{float64(9007199254740992), "9007199254740993 + 0.0"},
		{float64(1.5), "3 / 2"},
		{float64(0), "5 mod 2.5"},
		{float64(1.5), "5.5 mod 2"},
		{float64(1.5), "5.5 mod 2.0"},
		{float64(-0.5), "(-5) mod 2.25"},
		{float64(1.5), "5.5 mod (-2)"},
		{float64(0), "(-5) mod 2.5"},
		{float64(0), "(0-5.0) mod 2.5"},
		{float64(1), "(5 mod 2.5) + 1"},
		{float64(-1), "(5 mod 2.5) - 1"},
		{float64(0), "(5 mod 2.5) * 2"},
		{float64(0), "(5 mod 2.5) mod 2"},
		{float64(2), "((5 mod 2.5) + 1) * 2"},
		{float64(0.5), "(5 mod 2.5) + 0.5"},
		{"ab", "'a' + 'b'"},
	} {
		t.Run(tt.expression, func(t *testing.T) {
			analyzer, err := analyzeSource(t, "const C = "+tt.expression+";")
			if err != nil {
				t.Fatalf("analysis: %v; complete diagnostics %v", err, analyzer.Errors())
			}
			symbol, ok := analyzer.GetSymbolTable().Resolve("C")
			if !ok || !reflect.DeepEqual(symbol.Value, tt.want) {
				t.Fatalf("constant declaration %+v; want %T(%v)", symbol, tt.want, tt.want)
			}
			if expected, floating := tt.want.(float64); floating {
				if value := symbol.Value.(float64); expected == 0 && math.Signbit(value) {
					t.Fatal("Float modulo retained negative zero")
				}
			}
		})
	}
}

func TestConstantArithmetic_MixedCompositionTypeAndIntegerSyntaxDomain(t *testing.T) {
	for _, tt := range []struct {
		expression string
		want       float64
	}{{"5 mod 2.5", 0}, {"(5 mod 2.5) + 1", 1}} {
		t.Run(tt.expression, func(t *testing.T) {
			source := "const C = " + tt.expression + ";"
			analyzer, err := analyzeSource(t, source)
			if err != nil {
				t.Fatalf("complete diagnostics: %v", analyzer.Errors())
			}
			symbol, ok := analyzer.GetSymbolTable().Resolve("C")
			if !ok || !symbol.Type.Equals(types.FLOAT) || !reflect.DeepEqual(symbol.Value, tt.want) {
				t.Fatalf("mixed modulo declaration: %+v", symbol)
			}
			program := parseProgram(t, source)
			declaration := program.Statements[0].(*ast.ConstDecl)
			_, err = NewAnalyzer().evaluateConstantInt(declaration.Value)
			if err == nil || err.Error() != "expression is not a compile-time constant integer" {
				t.Fatalf("original integer-only syntax domain: %v", err)
			}
		})
	}
}

func TestConstantArithmetic_ZeroAndRejectedOperandContracts(t *testing.T) {
	for _, tt := range []struct{ expression, want, diagnostic string }{
		{"9007199254740993 div 0", "division by zero", "Constant expression expected at 1:28"},
		{"9007199254740993 mod 0", "modulo by zero", "Constant expression expected at 1:28"},
		{"1 / 0", "division by zero", "Constant expression expected at 1:13"},
		{"1.5 / 0", "division by zero", "Constant expression expected at 1:15"},
		{"(5 mod 2.5) div 0", "division by zero", "Syntax Error: Invalid Operands at 1:23"},
		{"(5 mod 2.5) mod 0", "modulo by zero", "Constant expression expected at 1:23"},
		{"5.5 mod 0", "modulo by zero", "Constant expression expected at 1:15"},
	} {
		t.Run(tt.expression, func(t *testing.T) {
			program := parseProgram(t, "const C = "+tt.expression+";")
			declaration := program.Statements[0].(*ast.ConstDecl)
			_, err := NewAnalyzer().evaluateConstant(declaration.Value)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("constant failure: %v; want %q", err, tt.want)
			}
			analyzer, err := analyzeSource(t, "const C = "+tt.expression+";")
			if err == nil || !reflect.DeepEqual(analyzer.Errors(), []string{tt.diagnostic}) {
				t.Fatalf("complete source diagnostics: %v", analyzer.Errors())
			}
		})
	}
	analyzer, err := analyzeSource(t, "const C = 5.5 div 2;")
	if err == nil || !reflect.DeepEqual(analyzer.Errors(), []string{"Syntax Error: Invalid Operands at 1:15"}) {
		t.Fatalf("mixed Float div diagnostics: %v", analyzer.Errors())
	}
}
