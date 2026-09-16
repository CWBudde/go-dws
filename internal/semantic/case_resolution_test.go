package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestCaseHints_ResolvedDeclarations(t *testing.T) {
	tests := []struct {
		name, source string
		want         []string
	}{
		{"assignment", "var Value := 1; value := 2;", []string{`"value" does not match case of declaration ("Value")`}},
		{"implicit result", "function Make: Integer; begin result := 1; end; PrintLn(Make);", []string{`"result" does not match case of declaration ("Result")`}},
		{"function call", "procedure Work; begin end; work();", []string{`"work" does not match case of declaration ("Work")`}},
		{"method call", "type T = class procedure Work; begin end; end; var Obj := T.Create; Obj.work();", []string{`"work" does not match case of declaration ("Work")`}},
		{"bare method", "type T = class procedure Work; begin end; end; var Obj := T.Create; Obj.work;", []string{`"work" does not match case of declaration ("Work")`}},
		{"property", "type T = class F: Integer; property Value: Integer read F write F; end; var Obj := T.Create; PrintLn(Obj.value);", []string{`"value" does not match case of declaration ("Value")`}},
		{"revisited overload argument", "procedure Work(X: Integer); overload; begin end; procedure Work(X: String); overload; begin end; var Value := 1; Work(value);", []string{`"value" does not match case of declaration ("Value")`}},
		{"builtin assigned once", "var Obj: TObject; PrintLn(assigned(Obj));", []string{`"assigned" does not match case of declaration ("Assigned")`}},
		{"implicit call value", "function Make: Integer; begin Result := 1; end; var V: Integer := make;", []string{`"make" does not match case of declaration ("Make")`}},
		{"address of function", "function Make: Integer; begin Result := 1; end; var P := @make;", []string{`"make" does not match case of declaration ("Make")`}},
		{"interface helper", "type I = interface procedure SayIt; end; type H = helper for I procedure Work; begin SayIt(); end; end;", nil},
		{"helper method", "type H = helper for Integer procedure Work; begin end; end; var V := 1; V.work;", []string{`"work" does not match case of declaration ("Work")`}},
		{"property write", "type T = class F: Integer; property Value: Integer read F write F; end; var Obj := T.Create; Obj.value := 1;", []string{`"value" does not match case of declaration ("Value")`}},
		{"same spelling", "procedure Work; begin end; Work();", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := analyzeWithHints(t, tt.source, HintsLevelPedantic)
			var hints []string
			for _, d := range diagnostics {
				if strings.Contains(d, "does not match case of declaration") {
					hints = append(hints, d)
				}
			}
			if len(hints) != len(tt.want) {
				t.Fatalf("case hints = %v, want %v (all diagnostics %v)", hints, tt.want, diagnostics)
			}
			for i, want := range tt.want {
				if !strings.Contains(hints[i], want) {
					t.Errorf("hint %q does not contain %q", hints[i], want)
				}
			}
		})
	}
}

func TestCaseHints_IdentifierIdentity(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.SetHintsLevel(HintsLevelPedantic)
	pos := token.Position{Line: 1, Column: 1}
	first := &ast.Identifier{BaseNode: ast.BaseNode{Token: token.Token{Pos: pos}}, Value: "value"}
	second := &ast.Identifier{BaseNode: ast.BaseNode{Token: token.Token{Pos: pos}}, Value: "value"}
	analyzer.addIdentifierCaseHint(first, "")
	analyzer.addIdentifierCaseHint(first, "Other")
	analyzer.addIdentifierCaseHint(first, "value")
	analyzer.addIdentifierCaseHint(first, "Value")
	analyzer.addIdentifierCaseHint(first, "Value")
	analyzer.addIdentifierCaseHint(second, "Value")
	if got := analyzer.Errors(); len(got) != 2 {
		t.Fatalf("distinct identifiers should each emit once: %v", got)
	}
}
