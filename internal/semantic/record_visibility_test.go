package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// analyzeRecordSource parses and analyzes input with pedantic hints, returning
// every diagnostic the analyzer produced (errors, warnings and hints alike).
func analyzeRecordSource(t *testing.T, input string) []string {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.SetHintsLevel(HintsLevelPedantic)
	_ = analyzer.Analyze(program)
	return analyzer.Errors()
}

// hasDiagnostic reports whether any diagnostic contains substr.
func hasDiagnostic(diags []string, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d, substr) {
			return true
		}
	}
	return false
}

const recordVisibilityPrelude = `
type TRec = record
   private
      FHidden : String;
   public
      Pub : String;
   published
      Doh : String;
end;
`

// TestRecordMemberVisibility covers access to private, public and published
// record fields from inside and outside the owning record.
func TestRecordMemberVisibility(t *testing.T) {
	const notVisible = `is not visible from this scope`

	tests := []struct {
		name      string
		source    string
		wantError bool
	}{
		{
			name:      "private field read from outside",
			source:    recordVisibilityPrelude + "var r : TRec;\nPrintLn(r.FHidden);\n",
			wantError: true,
		},
		{
			name:      "private field assignment from outside",
			source:    recordVisibilityPrelude + "var r : TRec;\nr.FHidden := 'a';\n",
			wantError: true,
		},
		{
			name: "private field in record literal",
			source: recordVisibilityPrelude +
				"const cr : TRec = (FHidden : 'hello'; Pub : 'World'; Doh : 'Duh');\n",
			wantError: true,
		},
		{
			name:      "public field read from outside",
			source:    recordVisibilityPrelude + "var r : TRec;\nPrintLn(r.Pub);\n",
			wantError: false,
		},
		{
			name:      "published field read from outside",
			source:    recordVisibilityPrelude + "var r : TRec;\nPrintLn(r.Doh);\n",
			wantError: false,
		},
		{
			name:      "public field in record literal",
			source:    recordVisibilityPrelude + "const cr : TRec = (Pub : 'World'; Doh : 'Duh');\n",
			wantError: false,
		},
		{
			name: "private field from inline method body",
			source: `
type TRec = record
   private
      FHidden : String;
   public
      procedure SetHidden(v : String);
      begin
         FHidden := v;
      end;
end;
`,
			wantError: false,
		},
		{
			name: "private field from inline class method via Result",
			source: `
type TRec = record
   private
      FHidden : String;
   public
      class function NewOne(s : String) : TRec;
      begin
         Result.FHidden := s;
      end;
end;
`,
			wantError: false,
		},
		{
			name: "private field from out-of-line method body",
			source: `
type TRec = record
   private
      FHidden : String;
   public
      procedure SayHello;
end;

procedure TRec.SayHello;
begin
   PrintLn('Hello ' + FHidden);
end;
`,
			wantError: false,
		},
		{
			name:      "lookup is case-insensitive",
			source:    recordVisibilityPrelude + "var r : TRec;\nPrintLn(r.fhidden);\n",
			wantError: true,
		},
		{
			name: "private field of another record is not visible",
			source: `
type TOther = record
   private
      FHidden : String;
end;

type TRec = record
   public
      Other : TOther;
      procedure Peek;
      begin
         PrintLn(Other.FHidden);
      end;
end;
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := analyzeRecordSource(t, tt.source)
			got := hasDiagnostic(diags, notVisible)
			if got != tt.wantError {
				t.Errorf("visibility error = %v, want %v (diagnostics: %v)", got, tt.wantError, diags)
			}
		})
	}
}

// TestRecordVisibilitySectionDiagnostics covers the record-body visibility
// specifier diagnostics: redundant sections, the unsupported "protected"
// section, and records without any field member.
func TestRecordVisibilitySectionDiagnostics(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    []string
		notWant []string
	}{
		{
			name: "redundant public section",
			source: `
type TRec = record
   public
   public
      Pub : String;
end;
`,
			want: []string{`Hint: Redundant specifier, visibility is already "public"`},
		},
		{
			name: "leading public section is redundant with the default",
			source: `
type TRec = record
   public
      Pub : String;
end;
`,
			want: []string{`Hint: Redundant specifier, visibility is already "public"`},
		},
		{
			name: "redundant private section",
			source: `
type TRec = record
   private
   private
      FHidden : String;
end;
`,
			want: []string{`Hint: Redundant specifier, visibility is already "private"`},
		},
		{
			name: "alternating sections are not redundant",
			source: `
type TRec = record
   private
      FHidden : String;
   public
      Pub : String;
end;
`,
			notWant: []string{"Redundant specifier"},
		},
		{
			name: "protected section is rejected",
			source: `
type TRec = record
   protected
      Prot : String;
end;
`,
			want: []string{`Records do not supported "protected" visibility specifier`},
		},
		{
			name: "record without field members",
			source: `
type TRec = record
end;
`,
			want: []string{"Record has no field members"},
		},
		{
			name: "record with a field has no field-member error",
			source: `
type TRec = record
   Pub : String;
end;
`,
			notWant: []string{"Record has no field members"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := analyzeRecordSource(t, tt.source)
			for _, want := range tt.want {
				if !hasDiagnostic(diags, want) {
					t.Errorf("missing diagnostic %q (got: %v)", want, diags)
				}
			}
			for _, notWant := range tt.notWant {
				if hasDiagnostic(diags, notWant) {
					t.Errorf("unexpected diagnostic %q (got: %v)", notWant, diags)
				}
			}
		})
	}
}
