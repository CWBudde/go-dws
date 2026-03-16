package frontend

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_CollectsParserDiagnostics(t *testing.T) {
	result := Compile("var x := );", "parser_error.pas", semantic.HintsLevelPedantic)

	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected parser diagnostics")
	}
	if !result.HasFatalDiagnostics() {
		t.Fatal("expected fatal parser diagnostics")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run on recovered parse output")
	}
	foundParsing := false
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseParsing {
			foundParsing = true
			break
		}
	}
	if !foundParsing {
		t.Fatal("expected at least one parsing diagnostic")
	}
}

func TestCompile_CollectsSemanticDiagnostics(t *testing.T) {
	source := `
var i: Integer;
begin
	i := 'oops';
end;
`

	result := Compile(source, "semantic_error.pas", semantic.HintsLevelPedantic)

	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run")
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected semantic diagnostics")
	}

	foundSemantic := false
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseSemantic {
			foundSemantic = true
			break
		}
	}
	if !foundSemantic {
		t.Fatal("expected at least one semantic diagnostic")
	}
}

func TestExtractPosition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLine int
		wantCol  int
		wantMsg  string
	}{
		{
			name:     "dwscript format",
			input:    `Syntax Error: Unknown name "Bug" [line: 2, column: 15]`,
			wantLine: 2,
			wantCol:  15,
			wantMsg:  `Syntax Error: Unknown name "Bug"`,
		},
		{
			name:     "at suffix",
			input:    "cannot access private field 'Field' of class 'TTest' at 16:2",
			wantLine: 16,
			wantCol:  2,
			wantMsg:  "cannot access private field 'Field' of class 'TTest'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col, msg := extractPosition(tt.input)
			if line != tt.wantLine || col != tt.wantCol || msg != tt.wantMsg {
				t.Fatalf("extractPosition(%q) = (%d, %d, %q), want (%d, %d, %q)",
					tt.input, line, col, msg, tt.wantLine, tt.wantCol, tt.wantMsg)
			}
		})
	}
}
