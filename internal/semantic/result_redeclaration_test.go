package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ident"
)

func TestResultRedeclaration(t *testing.T) {
	for _, tt := range []struct {
		name   string
		source string
	}{
		{"function", "function F: Integer; begin var result := 1; end;"},
		{"nested block", "function F: Integer; begin begin var ReSuLt := 1; end; end;"},
		{"inline for variable", "function F: Integer; begin for var result := 1 to 2 do PrintLn(result); end;"},
		{"inline for in variable", "function F: Integer; begin for var result in [1, 2] do PrintLn(result); end;"},
		{"lambda", "var f := lambda(): Integer begin var result := 1; end;"},
		{"inferred lambda", "var f := lambda() begin var result := 1; Result := 2; end;"},
		{"inferred lambda result reference", "var f := lambda() begin var result := 1; Result := result; end;"},
		{"contextual lambda", "type TFunc = function: Integer; var f: TFunc := lambda() begin var result := 1; Result := 2; end;"},
		{"contextual lambda without assignment", "type TFunc = function: Integer; var f: TFunc := lambda() begin var result := 1; end;"},
		{"class method", "type TTest = class function F: Integer; begin var result := 1; end; end;"},
		{"record method", "type TTest = record function F: Integer; begin var result := 1; end; end;"},
		{"helper method", "type TTest = helper for Integer function F: Integer; begin var result := 1; end; end;"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := analyzeSource(t, tt.source)
			if err == nil || !strings.Contains(ident.Normalize(err.Error()), `name "result" already exists`) {
				t.Fatalf("expected Result redeclaration error, got %v (%v)", err, analyzer.Errors())
			}
		})
	}
}

func TestResultRedeclaration_RoutineBoundary(t *testing.T) {
	for _, tt := range []struct{ name, source string }{
		{"global", "var result := 1; begin var result := 2; PrintLn(result); end;"},
		{"procedure", "procedure P; begin var result := 1; PrintLn(result); end;"},
		{"nested procedure", "function F: Integer; begin procedure P; begin var result := 1; PrintLn(result); end; Result := 2; end;"},
		{"nested lambda", "function F: Integer; begin var p := lambda() begin var result := 1; PrintLn(result); end; Result := 2; end;"},
		{"inferred lambda nested routine", "var f := lambda() begin procedure P; begin var result := 1; PrintLn(result); end; Result := 2; end;"},
		{"routine restoration", "function F: Integer; begin var p := lambda(): Integer begin Result := 1; end; Result := 2; end; procedure P; begin var result := 1; end; var result := 3;"},
	} {
		t.Run(tt.name, func(t *testing.T) { expectNoErrors(t, tt.source) })
	}
}

func TestResultRedeclaration_DiagnosticAndSelfAssignment(t *testing.T) {
	got := analyzeWithHints(t, "function Test: String;\nbegin\n var result := 'hello';\nend;", HintsLevelPedantic)
	for _, want := range []string{
		`Syntax Error: Name "result" already exists [line: 3, column: 6]`,
		`Hint: Result is never used [line: 4, column: 1]`,
	} {
		if !containsDiagnostic(got, want) {
			t.Errorf("missing %q in %v", want, got)
		}
	}
	got = analyzeWithHints(t, "function F: Integer; begin Result := Result; end;", HintsLevelPedantic)
	if !containsDiagnostic(got, "Hint: Assigning Result to itself [line: 1, column: 35]") {
		t.Errorf("missing Result self-assignment hint in %v", got)
	}
	got = analyzeWithHints(t, "var Result := 1; Result := Result;", HintsLevelPedantic)
	if !containsDiagnostic(got, "Hint: Assigning Result to itself [line: 1, column: 25]") {
		t.Errorf("missing global Result self-assignment hint in %v", got)
	}
}
