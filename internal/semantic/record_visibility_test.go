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

// TestRecordEmptyBodyDiagnostic verifies that "Record has no field members" is
// reported only for a record whose body declares no members at all. A record
// that declares static members, methods, properties or constants is legal even
// when it has no instance fields.
func TestRecordEmptyBodyDiagnostic(t *testing.T) {
	const emptyRecordDiagnostic = "Record has no field members"

	tests := []struct {
		name      string
		source    string
		wantError bool
	}{
		{
			name: "record with only class vars",
			source: `
type TRec = record
   class var Counter : Integer;
end;
`,
		},
		{
			name: "record with only methods",
			source: `
type TRec = record
   function Add(a, b : Integer) : Integer;
   begin
      Result := a + b;
   end;
end;
`,
		},
		{
			name: "record with only a class method",
			source: `
type TRec = record
   class function Zero : Integer;
   begin
      Result := 0;
   end;
end;
`,
		},
		{
			name: "record with only a property",
			source: `
type TRec = record
   property Value : Integer read GetValue;
   function GetValue : Integer;
   begin
      Result := 1;
   end;
end;
`,
		},
		{
			name: "record with only a constant",
			source: `
type TRec = record
   const Answer = 42;
end;
`,
		},
		{
			name: "genuinely empty record",
			source: `
type TRec = record
end;
`,
			wantError: true,
		},
		{
			name: "record with only visibility specifiers",
			source: `
type TRec = record
   public
   private
end;
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := analyzeRecordSource(t, tt.source)
			got := hasDiagnostic(diags, emptyRecordDiagnostic)
			if got != tt.wantError {
				t.Errorf("hasDiagnostic(%q) = %v, want %v (diagnostics: %v)",
					emptyRecordDiagnostic, got, tt.wantError, diags)
			}
		})
	}
}

// TestInlineRecordVisibilitySectionDiagnostics verifies that anonymous inline
// record types get the same visibility diagnostics as named record
// declarations. The specifiers are parsed into RecordTypeNode.VisibilitySections
// and checked when the inline type is resolved.
func TestInlineRecordVisibilitySectionDiagnostics(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    []string
		notWant []string
	}{
		{
			name:   "inline record with protected section is rejected",
			source: "var r : record protected X : Integer; end;\n",
			want:   []string{`Records do not supported "protected" visibility specifier`},
		},
		{
			name: "inline record with protected section spanning lines",
			source: `
var r : record
   protected
      X : Integer;
end;
`,
			want: []string{`Records do not supported "protected" visibility specifier`},
		},
		{
			name:   "inline record with redundant public section hints",
			source: "var r : record public X : Integer; end;\n",
			want:   []string{`Hint: Redundant specifier, visibility is already "public"`},
		},
		{
			name: "inline record with repeated private section hints",
			source: `
var r : record
   private
   private
      FHidden : Integer;
end;
`,
			want: []string{`Hint: Redundant specifier, visibility is already "private"`},
		},
		{
			name:   "plain inline record reports nothing",
			source: "var r : record X : Integer; end;\n",
			notWant: []string{
				`Records do not supported "protected" visibility specifier`,
				"Redundant specifier",
			},
		},
		{
			name: "alternating inline sections are not redundant",
			source: `
var r : record
   private
      FHidden : Integer;
   public
      Pub : Integer;
end;
`,
			notWant: []string{"Redundant specifier"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := analyzeRecordSource(t, tt.source)
			for _, want := range tt.want {
				if countDiagnostics(diags, want) != 1 {
					t.Errorf("want exactly one diagnostic %q, got %d (diagnostics: %v)",
						want, countDiagnostics(diags, want), diags)
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

// countDiagnostics returns how many diagnostics contain substr.
func countDiagnostics(diags []string, substr string) int {
	count := 0
	for _, d := range diags {
		if strings.Contains(d, substr) {
			count++
		}
	}
	return count
}
