package frontend

import (
	"fmt"
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// The fixture distinguishes narrowing at do from the generic mismatch at in.
func TestCompile_ForInAssignmentDiagnostics(t *testing.T) {
	for _, name := range []string{"for_in_subclass", "for_in1", "for_error4"} {
		t.Run(name, func(t *testing.T) {
			assertDiagnostics(t, fixtureSource(t, name+".pas"), name+".pas", fixtureExpectation(t, name+".txt"))
		})
	}
}

func TestCompile_ForInAssignmentPositionFallback(t *testing.T) {
	const source = `type TBase = class end;
type TChild = class(TBase) end;
var a: array of TBase;
var c: TChild;
for c in a do PrintLn(c.ClassName);`
	for _, tt := range []struct {
		name, want string
		clearIn    bool
	}{
		{
			name: "in keyword",
			want: `Syntax Error: Incompatible types: Cannot assign "TBase" to "TChild" [line: 5, column: 7]`,
		},
		{
			name: "statement token", clearIn: true,
			want: `Syntax Error: Incompatible types: Cannot assign "TBase" to "TChild" [line: 5, column: 1]`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{HintsLevel: semantic.HintsLevelPedantic}
			result := ParseWithOptions(source, opts)
			if result.HasFatalDiagnostics() {
				t.Fatalf("unexpected parser diagnostics: %v", result.DiagnosticStrings())
			}
			stmt := result.Program.Statements[len(result.Program.Statements)-1].(*ast.ForInStatement)
			stmt.DoPos = token.Position{}
			if tt.clearIn {
				stmt.InPos = token.Position{}
			}
			got := AnalyzeParsed(result, source, opts).DiagnosticStrings()
			if !slices.Equal(got, []string{tt.want}) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestCompile_ForInClassAssignmentKinds(t *testing.T) {
	const declarations = `type TBase = class end;
type TChild = class(TBase) end;
type TGrandChild = class(TChild) end;
type TOther = class end;
type TSibling = class(TBase) end;
type TBaseAlias = TBase;
type TChildAlias = TChild;
`
	for _, tt := range []struct {
		name, element, variable, header string
		want                            []string
	}{
		{
			name: "narrowing", element: "TBase", variable: "TChild", header: "for c in a do ;",
			want: []string{
				`Syntax Error: Incompatible types: Cannot assign "TBase" to "TChild" [line: 10, column: 12]`,
				`Hint: Empty FOR loop [line: 10, column: 15]`,
			},
		},
		{
			name: "indirect narrowing", element: "TBase", variable: "TGrandChild", header: "for c in a do ;",
			want: []string{
				`Syntax Error: Incompatible types: Cannot assign "TBase" to "TGrandChild" [line: 10, column: 12]`,
				`Hint: Empty FOR loop [line: 10, column: 15]`,
			},
		},
		{
			name: "aliases", element: "TBaseAlias", variable: "TChildAlias", header: "for c in a do ;",
			want: []string{
				`Syntax Error: Incompatible types: Cannot assign "TBaseAlias" to "TChildAlias" [line: 10, column: 12]`,
				`Hint: Empty FOR loop [line: 10, column: 15]`,
			},
		},
		{
			name: "unrelated classes", element: "TOther", variable: "TChild", header: "for c in a do ;",
			want: []string{
				`Syntax Error: Incompatible types: "TChild" and "TOther" [line: 10, column: 7]`,
				`Hint: Empty FOR loop [line: 10, column: 15]`,
			},
		},
		{
			name: "sibling classes", element: "TChild", variable: "TSibling", header: "for c in a do ;",
			want: []string{
				`Syntax Error: Incompatible types: "TSibling" and "TChild" [line: 10, column: 7]`,
				`Hint: Empty FOR loop [line: 10, column: 15]`,
			},
		},
		{
			name: "widening", element: "TChild", variable: "TBase", header: "for c in a do ;",
			want: []string{`Hint: Empty FOR loop [line: 10, column: 15]`},
		},
		{
			name: "same class", element: "TChild", variable: "TChild", header: "for c in a do ;",
			want: []string{`Hint: Empty FOR loop [line: 10, column: 15]`},
		},
		{
			name: "inline variable", element: "TBase", variable: "TChild", header: "for var c in a do ;",
			want: []string{`Hint: Empty FOR loop [line: 10, column: 19]`},
		},
		{
			name: "multiline header", element: "TBase", variable: "TChild", header: "for c in a\n    do ;",
			want: []string{
				`Syntax Error: Incompatible types: Cannot assign "TBase" to "TChild" [line: 11, column: 5]`,
				`Hint: Empty FOR loop [line: 11, column: 8]`,
			},
		},
		{
			name: "step header", element: "TBase", variable: "TChild", header: "for c in a step 2 do ;",
			want: []string{
				`Syntax Error: Incompatible types: Cannot assign "TBase" to "TChild" [line: 10, column: 19]`,
				`Hint: Empty FOR loop [line: 10, column: 22]`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := declarations + fmt.Sprintf("var a: array of %s;\nvar c: %s;\n%s", tt.element, tt.variable, tt.header)
			result := CompileWithOptions(source, Options{
				HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: true,
			})
			if got := result.DiagnosticStrings(); !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// CLI renderers print only structured errors once any exist, so the narrowing
// diagnostic must be structured to survive next to another structured error.
func TestCompile_ForInNarrowingIsStructured(t *testing.T) {
	const source = `type TBase = class end;
type TChild = class(TBase) end;
var a: array of TBase;
var c: TChild;
var s: String := 1;
for c in a do PrintLn(c.ClassName);`
	result := CompileWithOptions(source, Options{DisableSymbolDictionaryDiagnostics: true})
	var got []string
	for _, err := range result.Analyzer.StructuredErrors() {
		got = append(got, err.Message)
	}
	const want = `Syntax Error: Incompatible types: Cannot assign "TBase" to "TChild"`
	if !slices.Contains(got, want) {
		t.Fatalf("structured errors missing narrowing diagnostic\n got: %q\nwant: %q", got, want)
	}
}
