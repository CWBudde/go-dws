package cmd

import (
	"strings"
	"testing"
)

func TestRun_TestEnvelope(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:   "hints and result are framed",
			source: "var Foo: Integer; foo := 1; PrintLn(foo);",
			want:   "Errors >>>>\nHint: \"foo\" does not match case of declaration (\"Foo\") [line: 1, column: 37]\nResult >>>>\n1\n",
		},
		{
			name:    "runtime error after output is framed",
			source:  "PrintLn('a'); var x := 1 div 0;",
			want:    "Errors >>>>\nRuntime Error: division by zero: 1 div 0 [line: 1, column: 26]\nResult >>>>\na\n",
			wantErr: true,
		},
		{
			name:   "nothing to report means bare output",
			source: "PrintLn('a');",
			want:   "a\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, tt.source, nil, func() { testEnvelope = true; hintsLevel = "pedantic" })
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v, wantErr=%v\n%s", err, tt.wantErr, out)
			}
			if out != tt.want {
				t.Fatalf("got %q\nwant %q", out, tt.want)
			}
		})
	}
}

func TestRun_TestEnvelopeCompileFailureIsFlat(t *testing.T) {
	out, err := captureRun(t, "var x: Integer := 'hello';", nil, func() { testEnvelope = true; hintsLevel = "pedantic" })
	if err == nil {
		t.Fatal("expected a compile failure")
	}
	if !strings.HasPrefix(out, "Syntax Error:") || strings.Contains(out, ">>>>") {
		t.Fatalf("compile failures must be a flat diagnostic list, got %q", out)
	}
}
