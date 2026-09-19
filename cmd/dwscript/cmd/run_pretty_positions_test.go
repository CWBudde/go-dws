package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestRun_PrettyRuntimePositionOnce(t *testing.T) {
	tests := []struct {
		name   string
		source string
		index  int
		column int
	}{
		{"array read", "var a: array of Integer := [1]; PrintLn(a[2]);", 2, 44},
		{"array write", "var a: array of Integer := [1]; a[2] := 3;", 2, 36},
		{"string read", "var s := 'ab'; PrintLn(s[3]);", 3, 25},
		{"string write", "var s := 'ab'; s[3] := '!';", 3, 17},
	}
	for _, tt := range tests {
		for _, mode := range []string{"pretty", "plain", "envelope"} {
			t.Run(tt.name+"/"+mode, func(t *testing.T) {
				out, err := captureRun(t, tt.source, nil, func() {
					diagnosticsMode = mode
					if mode == "envelope" {
						diagnosticsMode = "pretty"
						testEnvelope = true
					}
				})
				if err == nil {
					t.Fatalf("expected bounds failure, got %q", out)
				}
				message := fmt.Sprintf("Upper bound exceeded! Index %d [line: 1, column: %d]", tt.index, tt.column)
				want := "Runtime Error: " + message + "\n"
				switch mode {
				case "pretty":
					want = "Runtime Error: Exception: " + message + "\n"
				case "envelope":
					want = "Errors >>>>\n" + want + "Result >>>>\n"
				}
				if out != want {
					t.Fatalf("got %q, want %q", out, want)
				}
			})
		}
	}
}

func TestRun_PrettyPreservesAuthoredPositionText(t *testing.T) {
	for _, column := range []int{57, 75} {
		t.Run(fmt.Sprint(column), func(t *testing.T) {
			message := fmt.Sprintf("authored [line: 1, column: %d]", column)
			source := "raise Exception.Create('" + message + "');"
			out, err := captureRun(t, source, nil, nil)
			if err == nil {
				t.Fatalf("expected unhandled exception, got %q", out)
			}
			want := "Runtime Error: Exception: " + message + " [line: 1, column: 57]\n"
			if out != want {
				t.Fatalf("got %q, want %q", out, want)
			}
		})
	}
}

func TestRun_PrettyReRaisePreservesPositionsAndStack(t *testing.T) {
	const source = "procedure H; begin raise ExceptObject; end; try var x := 0; var y := 5 div x; except H; end;"
	out, err := captureRun(t, source, nil, nil)
	if err == nil {
		t.Fatalf("expected unhandled re-raise, got %q", out)
	}
	want := "Runtime Error: Exception: Division by zero [line: 1, column: 72] [line: 1, column: 38]"
	if first, _, _ := strings.Cut(out, "\n"); first != want {
		t.Fatalf("first line %q, want %q", first, want)
	}
	if !strings.Contains(out, "H [line: 1, column: 86]") {
		t.Fatalf("missing re-raise stack frame in %q", out)
	}
}
