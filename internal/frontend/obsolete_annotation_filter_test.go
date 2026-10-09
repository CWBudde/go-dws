package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// A missing colon does not make a separate semantic error on its line a
// parser cascade. Only the parser's redundant declaration error is suppressed.
func TestFilterDiagnostics_ColonDoesNotHideSemanticAnnotation(t *testing.T) {
	colon := Diagnostic{
		Message: `Colon ":" expected`, Phase: PhaseParsing,
		Line: 1, Column: 6, Severity: SeverityError, Fatal: true,
	}
	tests := []struct {
		name      string
		candidate Diagnostic
		drop      bool
	}{
		{
			name: "separate semantic annotation on the same line",
			candidate: Diagnostic{
				Message: "variable 'other' must have either a type annotation or an initializer",
				Phase:   PhaseSemantic, Line: 1, Column: 12, Severity: SeverityError, Fatal: true,
			},
		},
		{
			name: "redundant parser declaration cascade on the same line",
			candidate: Diagnostic{
				Message: "variable declaration requires a type or initializer",
				Phase:   PhaseParsing, Line: 1, Column: 6, Severity: SeverityError, Fatal: true,
			},
			drop: true,
		},
		{
			name: "unrelated semantic name error on the same line",
			candidate: Diagnostic{
				Message: `Unknown name "Missing"`,
				Phase:   PhaseSemantic, Line: 1, Column: 12, Severity: SeverityError, Fatal: true,
			},
		},
		{
			name: "parser declaration error on another line",
			candidate: Diagnostic{
				Message: "variable declaration requires a type or initializer",
				Phase:   PhaseParsing, Line: 2, Column: 6, Severity: SeverityError, Fatal: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []Diagnostic{colon}
			if !tt.drop {
				want = append(want, tt.candidate)
			}
			got := filterDiagnostics([]Diagnostic{colon, tt.candidate})
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("filtered diagnostics = %+v; want %+v", got, want)
			}
		})
	}
}

// Missing declarations stop in upstream ReadNamedVarsDecl with Colon expected;
// failed initializer resolution reports its child error without an annotation
// cascade. The complete lists also keep unrelated same-line name errors visible.
func TestCompile_DeclarationDiagnosticsWithoutAnnotationFilter(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name: "missing type and initializer", source: "var s;",
			want: []string{`Syntax Error: Colon ":" expected [line: 1, column: 6]`},
		},
		{
			name: "multiple names missing type", source: "var i, j;",
			want: []string{`Syntax Error: Colon ":" expected [line: 1, column: 9]`},
		},
		{
			name: "multiple names with untyped initializer", source: "var i, j = ;",
			want: []string{`Syntax Error: Colon ":" expected [line: 1, column: 10]`},
		},
		{
			name: "failed initializer name", source: "var s := Missing;",
			want: []string{`Syntax Error: Unknown name "Missing" [line: 1, column: 10]`},
		},
		{
			name: "same line unrelated name after missing type", source: "var s; PrintLn(Missing);",
			want: []string{
				`Syntax Error: Colon ":" expected [line: 1, column: 6]`,
				`Syntax Error: Unknown name "Missing" [line: 1, column: 16]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "obsolete_annotation_filter.pas", semantic.HintsLevelPedantic)
			if !result.SemanticAttempted {
				t.Fatal("expected semantic analysis for recoverable declarations")
			}
			if got := result.DiagnosticStrings(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics = %q; want %q", got, tt.want)
			}
		})
	}
}
