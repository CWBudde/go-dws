package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// declaredPrelude declares the types and variables the Declared lookups below
// are resolved against.
const declaredPrelude = `
type TMyRecord = record
   Dummy : Integer;
end;

type TMyClass = class
   FValue : Integer;
   procedure Run;
   property Value : Integer read FValue;
end;

procedure TMyClass.Run;
begin
end;

var myVar : String;

const myConst = 'TMyRecord';
`

// analyzeDeclaredSource analyzes a script with pedantic hints enabled and
// returns the analyzer together with the analyzed program.
func analyzeDeclaredSource(t *testing.T, input string) (*Analyzer, *ast.Program) {
	t.Helper()

	program := parser.New(lexer.New(input)).ParseProgram()
	analyzer := NewAnalyzer()
	analyzer.SetHintsLevel(HintsLevelPedantic)
	_ = analyzer.Analyze(program)
	return analyzer, program
}

// declaredCallResults returns the compile-time value the analyzer folded for
// every Declared/ConditionalDefined call in the program, in source order.
func declaredCallResults(t *testing.T, analyzer *Analyzer, program *ast.Program) []bool {
	t.Helper()

	var results []bool
	ast.Inspect(program, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpression)
		if !ok {
			return true
		}
		callee, ok := call.Function.(*ast.Identifier)
		if !ok {
			return true
		}
		if folded, ok := analyzer.GetSemanticInfo().FoldedPredicate(callee); ok {
			results = append(results, folded)
		}
		return true
	})
	return results
}

func TestAnalyzeDeclared_NameResolution(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"TMyRecord", true},
		{"tmyrecord", true},
		{"TMyRecord.Dummy", true},
		{"TMyRecord.DUMMY", true},
		{"TMyRecord.Oops", false},
		{"Dummy", false},
		{"dummy", false},
		{"TObject", true},
		{"TObject.Create", true},
		{"TObject.ClassName", true},
		{"TObject.Nope", false},
		{"Internal.TObject.Create", true},
		{"Internal.Sin", true},
		{"internal.sin", true},
		{"Internal.NoSuchThing", false},
		{"Sin", true},
		{"myVar", true},
		{"MYVAR", true},
		{"TMyClass", true},
		{"TMyClass.FValue", true},
		{"TMyClass.Run", true},
		{"TMyClass.Value", true},
		{"TMyClass.Missing", false},
		{"FValue", false},
		{"Internal", false},
		{"", false},
		{"TMyRecord.", false},
	}

	analyzer, _ := analyzeDeclaredSource(t, declaredPrelude)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analyzer.isDeclaredName(tt.name); got != tt.want {
				t.Errorf("isDeclaredName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAnalyzeDeclared_FoldsCallSites(t *testing.T) {
	input := declaredPrelude + `
var a := Declared('TMyRecord');
var b := Declared('dummy');
var c := Declared(myConst);
var d := ConditionalDefined('ANYTHING');
`

	analyzer, program := analyzeDeclaredSource(t, input)
	if errs := analyzer.Errors(); len(errs) > 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	got := declaredCallResults(t, analyzer, program)
	want := []bool{true, false, true, false}
	if len(got) != len(want) {
		t.Fatalf("folded %d calls, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call %d folded to %v, want %v", i, got[i], want[i])
		}
	}
}

func TestAnalyzeDeclared_EmitsNoHints(t *testing.T) {
	input := declaredPrelude + `
var a := Declared('tmyrecord.dummy');
var b := Declared('internal.sin');
`

	analyzer, _ := analyzeDeclaredSource(t, input)
	for _, err := range analyzer.Errors() {
		if strings.Contains(err, "does not match case") {
			t.Errorf("unexpected case-mismatch hint: %s", err)
		}
	}
}

func TestAnalyzeDeclared_ArgumentErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "declared with integer argument",
			input:   "var i : Integer;\nvar b := Declared(i);",
			wantErr: "String expected",
		},
		{
			name:    "conditional defined with integer argument",
			input:   "var i : Integer;\nvar b := ConditionalDefined(i);",
			wantErr: "String expected",
		},
		{
			name:    "declared with non-constant string variable",
			input:   "var s : String;\nvar b := Declared(s);",
			wantErr: "String expected",
		},
		{
			name:    "declared without arguments",
			input:   "var b := Declared();",
			wantErr: "expects 1 argument",
		},
		{
			name:    "declared with two arguments",
			input:   "var b := Declared('a', 'b');",
			wantErr: "expects 1 argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, _ := analyzeDeclaredSource(t, tt.input)
			errs := analyzer.Errors()
			for _, err := range errs {
				if strings.Contains(err, tt.wantErr) {
					return
				}
			}
			t.Errorf("expected error containing %q, got %v", tt.wantErr, errs)
		})
	}
}

func TestAnalyzeDeclared_NoCascadingUnknownName(t *testing.T) {
	analyzer, _ := analyzeDeclaredSource(t, "var i : Integer;\nvar b := Declared(i);\nb := ConditionalDefined(i);")

	for _, err := range analyzer.Errors() {
		if strings.Contains(err, "Unknown name") || strings.Contains(err, "Undefined variable") {
			t.Errorf("unexpected cascading error: %s", err)
		}
	}
}
