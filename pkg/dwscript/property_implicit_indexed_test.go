package dwscript

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// These cases catch reading the empty pair as a method call, losing the selected
// accessor, and evaluating receivers/indices/getters twice or after an exception.
func TestImplicitIndexedCompatibility_Execution(t *testing.T) {
	paths, err := filepath.Glob("../../testdata/property_implicit_indexed/*.dws")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no indexed property cases")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(path[:len(path)-4] + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = engine.Eval(string(source)); err != nil {
				t.Fatalf("execution failed: %v", err)
			}
			if got := output.String(); got != string(want) {
				t.Fatalf("output %q; want %q", got, want)
			}
		})
	}
}
