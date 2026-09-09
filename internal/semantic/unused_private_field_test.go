package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// pedanticHints analyzes input at the pedantic hint level and returns every
// diagnostic the analyzer accumulated.
func pedanticHints(t *testing.T, input string) []string {
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

// hasUnusedPrivateFieldHint reports whether the diagnostics flag fieldName as an
// unused private field.
func hasUnusedPrivateFieldHint(diagnostics []string, fieldName string) bool {
	needle := `Private field "` + fieldName + `" declared but never used`
	for _, d := range diagnostics {
		if strings.Contains(d, needle) {
			return true
		}
	}
	return false
}

// TestUnusedPrivateFieldHint_UsageTracking covers the ways a private field can be
// referenced. A bare name inside a method body or a property expression accessor
// resolves through the symbol table rather than the currentClass.GetField
// fallback, so each of these used to be reported as unused.
func TestUnusedPrivateFieldHint_UsageTracking(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		field      string
		wantHinted bool
	}{
		{
			name: "qualified read through Self",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
	begin
		PrintLn(Self.FValue);
	end;
end;
`,
			field: "FValue",
		},
		{
			name: "bare read in an inline method body",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
	begin
		PrintLn(FValue);
	end;
end;
`,
			field: "FValue",
		},
		{
			name: "bare read in an out-of-line method body",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
end;

procedure TTest.Touch;
begin
	PrintLn(FValue);
end;
`,
			field: "FValue",
		},
		{
			name: "bare assignment in a method body",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
	begin
		FValue := 1;
	end;
end;
`,
			field: "FValue",
		},
		{
			name: "bare compound assignment in a method body",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
	begin
		FValue += 1;
	end;
end;
`,
			field: "FValue",
		},
		{
			name: "expression-form property read accessor",
			input: `
type TTest = class
private
	FValue: Integer;
public
	property Doubled: Integer read (FValue * 2);
end;
`,
			field: "FValue",
		},
		{
			name: "expression-form property write accessor",
			input: `
type TTest = class
private
	FValue: Integer;
public
	property Halved: Integer read (FValue) write (FValue := Value * 2);
end;
`,
			field: "FValue",
		},
		{
			name: "identifier property read specifier",
			input: `
type TTest = class
private
	FValue: Integer;
public
	property Value: Integer read FValue write FValue;
end;
`,
			field: "FValue",
		},
		{
			name: "genuinely unused private field is still hinted",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
	begin
	end;
end;
`,
			field:      "FValue",
			wantHinted: true,
		},
		{
			name: "a local shadowing the field does not count as usage",
			input: `
type TTest = class
private
	FValue: Integer;
public
	procedure Touch;
	var
		FValue: Integer;
	begin
		FValue := 1;
		PrintLn(FValue);
	end;
end;
`,
			field:      "FValue",
			wantHinted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := pedanticHints(t, tt.input)
			got := hasUnusedPrivateFieldHint(diagnostics, tt.field)
			if got != tt.wantHinted {
				t.Errorf("unused-private-field hint for %q = %v, want %v\ndiagnostics: %v",
					tt.field, got, tt.wantHinted, diagnostics)
			}
		})
	}
}

// TestUnusedPrivateFieldHint_SiblingFieldsTrackedIndependently guards against a
// scope-wide "any field used marks all fields used" regression.
func TestUnusedPrivateFieldHint_SiblingFieldsTrackedIndependently(t *testing.T) {
	input := `
type TTest = class
private
	FUsed: Integer;
	FUnused: Integer;
public
	procedure Touch;
	begin
		PrintLn(FUsed);
	end;
end;
`
	diagnostics := pedanticHints(t, input)
	if hasUnusedPrivateFieldHint(diagnostics, "FUsed") {
		t.Errorf("FUsed should not be hinted\ndiagnostics: %v", diagnostics)
	}
	if !hasUnusedPrivateFieldHint(diagnostics, "FUnused") {
		t.Errorf("FUnused should be hinted\ndiagnostics: %v", diagnostics)
	}
}

// TestUnusedPrivateFieldHint_NotEmittedBelowPedantic keeps the hint gated.
func TestUnusedPrivateFieldHint_NotEmittedBelowPedantic(t *testing.T) {
	input := `
type TTest = class
private
	FValue: Integer;
end;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.SetHintsLevel(HintsLevelNormal)
	_ = analyzer.Analyze(program)
	if hasUnusedPrivateFieldHint(analyzer.Errors(), "FValue") {
		t.Errorf("hint should be pedantic-only, got: %v", analyzer.Errors())
	}
}
