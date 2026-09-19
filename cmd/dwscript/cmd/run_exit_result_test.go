package cmd

import (
	"strings"
	"testing"
)

func TestRun_ExitValueUsesResult(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"function", "function F: String; begin Exit('ok') end; PrintLn(F);", "ok\n"},
		{"inline method", "type T = class function F: String; begin Exit('ok') end; end; PrintLn(T.Create.F);", "ok\n"},
		{"out of line method", "type T = class function F: String; end; function T.F: String; begin Exit('ok') end; PrintLn(T.Create.F);", "ok\n"},
		{"nested function", "function Outer: String; begin function Inner: String; begin Exit('ok') end; Exit(Inner()) end; PrintLn(Outer);", "ok\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := captureRun(t, tt.source, nil, func() { testEnvelope = true; hintsLevel = "pedantic" })
			if err != nil || out != tt.want {
				t.Fatalf("err=%v output=%q, want %q", err, out, tt.want)
			}
		})
	}
}

func TestRun_ExitValueResultUsageStaysInScope(t *testing.T) {
	for _, source := range []string{
		"function F: Integer; begin Exit end; PrintLn(F);",
		"function Outer: Integer; begin function Inner: Integer; begin Exit(1) end; Inner; end; PrintLn(Outer);",
	} {
		out, err := captureRun(t, source, nil, func() { testEnvelope = true; hintsLevel = "pedantic" })
		if err != nil || strings.Count(out, "Result is never used") != 1 || !strings.HasSuffix(out, "Result >>>>\n0\n") {
			t.Fatalf("err=%v output=%q, want one unused outer Result hint and default result", err, out)
		}
	}
}

func TestRun_ExitValueValidation(t *testing.T) {
	for _, tt := range []struct{ source, want string }{
		{"Exit(1);", "exit with value not allowed at program level"},
		{"procedure P; begin Exit(1) end; P;", "exit with value not allowed in procedure"},
		{"function F: Integer; begin Exit('bad') end; PrintLn(F);", "incompatible with function return type"},
		{"function F: Integer; begin try Result := 1 finally Exit(2) end end; PrintLn(F);", "exit statement not allowed in finally block"},
	} {
		out, err := captureRun(t, tt.source, nil, func() { diagnosticsMode = "plain" })
		if err == nil || !strings.Contains(out, tt.want) {
			t.Fatalf("err=%v output=%q, want error containing %q", err, out, tt.want)
		}
	}
}
