package dwscript

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These cases catch reading the empty pair as a method call, losing the selected
// accessor, and evaluating receivers/indices/getters twice or after an exception.
func TestInheritedIndexedCompatibility_Execution(t *testing.T) {
	paths, err := filepath.Glob("../../testdata/property_inherited_indexed/*.dws")
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

func TestInheritedIndexedCompatibility_CompileArguments(t *testing.T) {
	const prefix = "type TBase = class\n function Get(I: Integer): Integer; begin Result := I; end;\n property Prop[I: Integer]: Integer read Get reintroduce;\nend;\ntype TChild = class(TBase)\n"
	const hint = `hint 6:56 Hint: Property "Prop" reintroduced a method, you should remove empty brackets ()`
	tests := []struct{ name, read, want string }{
		{"omitted", "Prop()", hint + "\nerror 6:57 More arguments expected"},
		{"empty", "Prop()[]", hint + "\nerror 6:57 More arguments expected"},
		{"unfinished type", "Prop()[True", hint + `
error 6:63 ")" expected`},
		{"unfinished child", "Prop()[Missing", hint + `
error 6:59 Unknown name "Missing"
error 6:66 ")" expected`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := prefix + " function Read: Integer; begin Result := inherited " + tt.read + "; end;\nend;"
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if err == nil || program != nil {
				t.Fatalf("invalid argument list compiled: program=%v error=%v", program, err)
			}
			compileError, ok := err.(*CompileError)
			if !ok {
				t.Fatalf("expected CompileError, got %T: %v", err, err)
			}
			got := make([]string, len(compileError.Errors))
			for i, diagnostic := range compileError.Errors {
				got[i] = fmt.Sprintf("%s %d:%d %s", diagnostic.Severity, diagnostic.Line, diagnostic.Column, diagnostic.Message)
			}
			if messages := strings.Join(got, "\n"); messages != tt.want {
				t.Fatalf("got %q; want %q", messages, tt.want)
			}
		})
	}
}
