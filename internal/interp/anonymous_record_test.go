package interp

import (
	"strings"
	"testing"
)

// runScriptTestWithSemantic is runScriptTest with semantic analysis enabled,
// which the CLI always runs. Empty array literals need it: their element type
// is inferred by the analyzer, not by the evaluator.
func runScriptTestWithSemantic(t *testing.T, script, expectedOutput string) {
	t.Helper()

	result, output := testEvalWithOutputAndSemantic(t, script)
	if isError(result) {
		t.Fatalf("Script execution failed: %v", result.String())
	}

	if got, want := strings.TrimSpace(output), strings.TrimSpace(expectedOutput); got != want {
		t.Errorf("Output mismatch:\nExpected:\n%s\n\nGot:\n%s", want, got)
	}
}

// ============================================================================
// Anonymous record constructor expressions: record a := 1; ... end
//
// This form is structurally typed - it takes no record type from context - so
// it can be declared, read back and serialized without any type annotation.
// ============================================================================

func TestAnonymousRecordExpression_FieldAccess(t *testing.T) {
	runScriptTest(t, `
		var r := record a := 1; b := 'x'; end;
		PrintLn(r.a);
		PrintLn(r.b);
	`, "1\nx")
}

func TestAnonymousRecordExpression_CaseInsensitiveFieldAccess(t *testing.T) {
	runScriptTest(t, `
		var r := record SomeField := 42; end;
		PrintLn(r.somefield);
	`, "42")
}

func TestAnonymousRecordExpression_Stringify(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name:   "mixed scalar types",
			script: `var s := 'hello'; PrintLn(JSON.Stringify(record a := s; b := True; c := False; d := s + s; e := 123; f := 12.5; g := Null; end));`,
			want:   `{"a":"hello","b":true,"c":false,"d":"hellohello","e":123,"f":12.5,"g":null}`,
		},
		{
			name:   "keys are serialized in sorted order",
			script: `var i := 123; PrintLn(JSON.Stringify(record "is" := i.ToString; "2i" := 2*i; end));`,
			want:   `{"2i":246,"is":"123"}`,
		},
		{
			name:   "quoted names that are not identifiers survive",
			script: `var i := 123; PrintLn(JSON.Stringify(record "i*i" := i * i; "i-1" := (i-1).ToString; end));`,
			want:   `{"i*i":15129,"i-1":"122"}`,
		},
		{
			name:   "nested arrays",
			script: `PrintLn(JSON.Stringify(record a := []; b := [ [] ]; c := [ [], [] ]; end));`,
			want:   `{"a":[],"b":[[]],"c":[[],[]]}`,
		},
		{
			name:   "single line without trailing semicolon",
			script: `PrintLn(JSON.Stringify(record Field := 123 end));`,
			want:   `{"Field":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runScriptTestWithSemantic(t, tt.script, tt.want)
		})
	}
}

// Assigning a record expression to a second variable must copy it, matching the
// value semantics of named records.
func TestAnonymousRecordExpression_AssignmentCopies(t *testing.T) {
	runScriptTestWithSemantic(t, `
		var ra := record a := 1; b := []; end;
		var ra1 := ra;
		PrintLn(JSON.Stringify(ra));
		PrintLn(JSON.Stringify(ra1));
	`, `{"a":1,"b":[]}`+"\n"+`{"a":1,"b":[]}`)
}

func TestAnonymousRecordExpression_DuplicateFieldIsAnError(t *testing.T) {
	result, _ := testEvalWithOutput(`var r := record a := 1; a := 2; end;`)
	if !isError(result) {
		t.Errorf("expected an error for a duplicate field, got %v", result)
	}
}
