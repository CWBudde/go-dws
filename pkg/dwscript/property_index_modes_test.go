package dwscript

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestEngineCompile_PropertyIndexDeclarationFixtures(t *testing.T) {
	for _, name := range []string{"array_params1", "array_params2", "array_params3"} {
		t.Run(name, func(t *testing.T) {
			base := "../../testdata/fixtures/FailureScripts/" + name
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(string(source))
			if program != nil || err == nil {
				t.Fatalf("invalid source compiled: %v, %v", program, err)
			}
			ce, ok := err.(*CompileError)
			if !ok {
				t.Fatalf("error = %T: %v", err, err)
			}
			got := make([]string, len(ce.Errors))
			for i, d := range ce.Errors {
				got[i] = fmt.Sprintf("Syntax Error: %s [line: %d, column: %d]", d.Message, d.Line, d.Column)
				if d.Severity != SeverityError {
					t.Fatalf("unexpected severity %q", d.Severity)
				}
			}
			if strings.Join(got, "\n") != strings.TrimSpace(string(expected)) {
				t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), expected)
			}
		})
	}
}
