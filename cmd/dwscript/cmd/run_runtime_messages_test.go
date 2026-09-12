package cmd

import (
	"strings"
	"testing"
)

// TestRun_RuntimeMessageVocabulary pins the sentences DWScript uses for the
// runtime errors go-dws used to phrase in its own words. Each case is a fixture
// under testdata/fixtures/SimpleScripts reduced to its essential line; the
// wording, not the position, is what these assert.
func TestRun_RuntimeMessageVocabulary(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "integer division by zero",
			source: "var a := 0; var b := 1 div a;",
			want:   "Runtime Error: Division by zero",
		},
		{
			name:   "modulo by zero is the same sentence",
			source: "var a := 0; var b := 1 mod a;",
			want:   "Runtime Error: Division by zero",
		},
		{
			name:   "string index below the first character",
			source: "var s := 'string'; PrintLn(s[0]);",
			want:   "Runtime Error: Lower bound exceeded! Index 0",
		},
		{
			name:   "string index past the last character",
			source: "var s := 'ab'; PrintLn(s[3]);",
			want:   "Runtime Error: Upper bound exceeded! Index 3",
		},
		{
			name:   "string write below the first character",
			source: "var s := 'string'; s[0] := '!';",
			want:   "Runtime Error: Lower bound exceeded! Index 0",
		},
		{
			name:   "string write past the last character",
			source: "var s := 'ab'; s[3] := '!';",
			want:   "Runtime Error: Upper bound exceeded! Index 3",
		},
		{
			name:   "call to an unbound external routine",
			source: "function Dummy(p : Integer) : String; external; Dummy(12);",
			want:   `Runtime Error: Unhandled call to external symbol "Dummy" from`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, tt.source, nil, func() { testEnvelope = true; hintsLevel = "pedantic" })
			if err == nil {
				t.Fatalf("expected a runtime error, got output %q", out)
			}
			if !strings.Contains(out, tt.want) {
				t.Fatalf("got %q\nwant it to contain %q", out, tt.want)
			}
		})
	}
}

// TestRun_StringIndexOutOfRangeIsCatchable pins that a bad string index raises a
// script exception rather than killing the program, for reads and writes alike.
// This held before the messages were unified and has to keep holding after:
// SimpleScripts/string_bounds wraps four out-of-range *writes* in try/except and
// expects execution to continue past each.
func TestRun_StringIndexOutOfRangeIsCatchable(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "read",
			source: "var s := 'a'; try PrintLn(s[0]); except PrintLn('caught'); end; PrintLn('after');",
		},
		{
			name:   "write",
			source: "var s := 'a'; try s[0] := '!'; except PrintLn('caught'); end; PrintLn('after');",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, tt.source, nil, nil)
			if err != nil {
				t.Fatalf("expected the exception to be caught, got %v\n%s", err, out)
			}
			for _, want := range []string{"caught", "after"} {
				if !strings.Contains(out, want) {
					t.Fatalf("got %q\nwant it to contain %q", out, want)
				}
			}
		})
	}
}

// TestRun_ReRaiseKeepsOriginalMessage pins SimpleScripts/re_raise: raising the
// exception that is already in flight is a re-raise, so the report keeps the
// original message and position and appends the position of the raise, instead
// of rewrapping it as a user-defined exception.
func TestRun_ReRaiseKeepsOriginalMessage(t *testing.T) {
	const source = "procedure H; begin raise ExceptObject; end; try var x := 0; var y := 5 div x; except H; end;"

	out, err := captureRun(t, source, nil, func() { testEnvelope = true })
	if err == nil {
		t.Fatalf("expected the re-raised exception to go unhandled, got %q", out)
	}
	if !strings.Contains(out, "Division by zero") {
		t.Fatalf("got %q\nwant the original runtime message", out)
	}
	if strings.Contains(out, "User defined exception") {
		t.Fatalf("got %q\nwant a re-raise, not a user-defined exception", out)
	}
	if strings.Count(out, "[line:") < 2 {
		t.Fatalf("got %q\nwant both the original and the re-raise position", out)
	}
}
