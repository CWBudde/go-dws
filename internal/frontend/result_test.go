package frontend

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
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
	if result.HasSemanticBlockingDiagnostics() {
		t.Fatal("expected recovered parser diagnostics to remain semantic-recoverable")
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

func TestParserDiagnosticBlocksSemantic(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		block bool
	}{
		{name: "recoverable invalid expression", code: parser.ErrInvalidExpression, block: false},
		{name: "recoverable missing end", code: parser.ErrMissingEnd, block: false},
		{name: "recoverable expected identifier", code: parser.ErrExpectedIdent, block: false},
		{name: "unknown parser code blocks", code: "E_UNKNOWN_PARSER_STATE", block: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.NewParserError(lexer.Position{Line: 1, Column: 1}, 1, "test", tt.code)
			got := parserDiagnosticBlocksSemantic(err)
			if got != tt.block {
				t.Fatalf("parserDiagnosticBlocksSemantic(%q) = %v, want %v", tt.code, got, tt.block)
			}
		})
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
		wantMsg  string
		wantLine int
		wantCol  int
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

func TestCompile_RendersStructuredSemanticSyntaxErrorsInDWScriptFormat(t *testing.T) {
	source := `
type TEnum = (eZero, eOne, eTwo);

var ab1 : array [False..True] of Integer;
var ab2 : array [True..False] of Integer;

var v : Variant;

ab1[1]:=1;
ab1['1']:=1;
ab1[True]:=1;
ab1[Integer(True)]:=1;
ab1[eZero]:=1;
ab1[Integer(eZero)]:=1;
ab1[v]:=1;
`

	result := Compile(source, "array_index_bool.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Lower bound is greater than upper bound [line: 5, column: 23]`,
		`Syntax Error: Array index expected "Boolean" but got "Integer" [line: 9, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "String" [line: 10, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "Integer" [line: 12, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "TEnum" [line: 13, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "Integer" [line: 14, column: 5]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_DedupesStructuredAndLegacySemanticVisibilityErrors(t *testing.T) {
	source := `
type
   TTest = class
      private
         Field : Integer;
   end;
type
   TSubTest = class (TTest)
      protected
         Field2 : Integer;
         procedure Stuff;
   end;

procedure TSubTest.Stuff;
begin
   Field2:=1;
   Field:=2;
end;

var o := TSubTest.Create;
o.Field:=1;`

	result := Compile(source, "visibility4.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Member symbol "Field" is not visible from this scope [line: 17, column: 4]`,
		`Syntax Error: Member symbol "Field" is not visible from this scope [line: 21, column: 3]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredUnknownNameDiagnostics(t *testing.T) {
	source := `
Foo();
`

	result := Compile(source, "unknown_name.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Unknown name "Foo" [line: 2, column: 5]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredNoOverloadDiagnostics(t *testing.T) {
	source := `
function Pick(x: Integer): Integer;
begin
	Result := x;
end;

function Pick(x: String): String;
begin
	Result := x;
end;

Pick(True);
`

	result := Compile(source, "no_overload.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: There is no overloaded version of "Pick" that can be called with these arguments [line: 12, column: 5]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}
