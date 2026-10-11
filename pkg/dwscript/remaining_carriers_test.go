package dwscript

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineCompile_RemainingCarrierFixtures(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"try_except1", "class_error4", "class_error2", "class_error3"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("../../testdata/fixtures/FailureScripts", name)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(string(source))
			if program != nil {
				t.Fatal("Compile returned a public Program for invalid source")
			}
			compileErr, ok := err.(*CompileError)
			if !ok {
				t.Fatalf("error = %T %v, want CompileError", err, err)
			}
			var got []string
			for _, diag := range compileErr.Errors {
				if diag.Severity != SeverityError {
					t.Errorf("unexpected non-error diagnostic: %v", diag)
				}
				got = append(got, fmt.Sprintf("Syntax Error: %s [line: %d, column: %d]", diag.Message, diag.Line, diag.Column))
			}
			if strings.Join(got, "\n") != strings.TrimSpace(string(want)) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, strings.TrimSpace(string(want)))
			}
		})
	}
}

func TestEngineEval_ExceptionAliasHandler(t *testing.T) {
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Eval("type TAlias = Exception;\ntry\nraise Exception.Create('caught');\nexcept\non e: TAlias do PrintLn(e.Message);\nend;")
	if err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "caught\n" {
		t.Fatalf("output = %q, want caught exception message", got)
	}
}
