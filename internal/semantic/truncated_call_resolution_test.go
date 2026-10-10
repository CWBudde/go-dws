package semantic

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// analyzeDespiteParserErrors analyzes source at pedantic hint level even when
// the parser stopped, as the frontend does.
func analyzeDespiteParserErrors(t *testing.T, source string) *Analyzer {
	t.Helper()
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	a := NewAnalyzer()
	a.SetHintsLevel(HintsLevelPedantic)
	_ = a.Analyze(program)
	return a
}

// TestTruncatedCall_EnumerationTypeCallStops pins upstream's
// ReadEnumerationSymbolName: on an enumeration type name, a call other than
// ByName(…) or an empty () stops at the member before its arguments are read,
// so the stop precedes a truncated argument list's own stop.
func TestTruncatedCall_EnumerationTypeCallStops(t *testing.T) {
	tests := []struct {
		name, source, want string
		line, column       int
	}{
		// FailureScripts/enums8
		{"truncated", "Type TElement = (etOne=1);\n\nvar b := TElement.low(;", `There is no accessible method with name "low" for type TElement`, 3, 19},
		{"with arguments", "Type TElement = (etOne=1);\nvar b := TElement.foo(1);", `There is no accessible method with name "foo" for type TElement`, 2, 19},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := analyzeDespiteParserErrors(t, tt.source).StructuredErrors()
			if len(errs) != 1 {
				t.Fatalf("errors = %v, want one stop", errs)
			}
			err := errs[0]
			if !err.Stop || err.Message != tt.want || err.Pos.Line != tt.line || err.Pos.Column != tt.column {
				t.Fatalf("error = %q at %d:%d (stop %v), want stop %q at %d:%d",
					err.Message, err.Pos.Line, err.Pos.Column, err.Stop, tt.want, tt.line, tt.column)
			}
		})
	}

	for _, source := range []string{
		"Type TElement = (etOne=1);\nvar b := TElement.Low();",
		"Type TElement = (etOne=1);\nvar b := TElement.ByName('etOne');",
	} {
		if errs := analyzeDespiteParserErrors(t, source).Errors(); len(errs) != 0 {
			t.Errorf("%q: unexpected diagnostics %v", source, errs)
		}
	}
}

// TestTruncatedCall_MethodCasingHintPrecedesStop pins that upstream resolves a
// method call's member before ReadArguments stops, so its pedantic casing hint
// is kept even when the argument list is cut short (FailureScripts/params3).
func TestTruncatedCall_MethodCasingHintPrecedesStop(t *testing.T) {
	source := "type TFoo = class external\npublic procedure translate(x : Float); end;\nvar Bar: TFoo;\nBar.Translate(-);"
	got := analyzeDespiteParserErrors(t, source).Errors()
	want := []string{`Hint: "Translate" does not match case of declaration ("translate") [line: 4, column: 5]`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostics = %q, want %q", got, want)
	}
}

func TestContainsParserRecovery_PreservesEarlierSibling(t *testing.T) {
	for _, tt := range []struct {
		expression ast.Expression
		name       string
	}{
		{&ast.InvalidExpression{}, "invalid"},
		{&ast.CallExpression{Truncated: true}, "call"},
		{&ast.NewExpression{Truncated: true}, "new"},
		{&ast.MethodCallExpression{Truncated: true}, "method"},
		{&ast.InheritedExpression{Truncated: true}, "inherited"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, completed := range []ast.Expression{&ast.CallExpression{}, &ast.NewExpression{}, &ast.MethodCallExpression{}, &ast.InheritedExpression{}} {
				expr := &ast.BinaryExpression{Left: tt.expression, Right: completed}
				if !containsParserRecovery(expr) {
					t.Errorf("completed %T erased earlier %T", completed, tt.expression)
				}
			}
		})
	}
	for _, completed := range []ast.Expression{&ast.CallExpression{}, &ast.NewExpression{}, &ast.MethodCallExpression{}, &ast.InheritedExpression{}} {
		if containsParserRecovery(completed) {
			t.Errorf("completed %T incorrectly marked as recovery", completed)
		}
	}
}
