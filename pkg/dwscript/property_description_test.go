package dwscript

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestPropertyDescription_Execution(t *testing.T) {
	for _, name := range []string{"literal", "empty", "escaped", "multiline"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../testdata/property_description/" + name + ".dws")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("../../testdata/property_description/" + name + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(string(source))
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			wantDescription := map[string]string{"literal": "a description", "empty": "", "escaped": "it's \"quoted\"", "multiline": "first\nsecond \"quoted\""}[name]
			class := program.AST().Statements[0].(*ast.ClassDecl)
			for _, prop := range class.Properties {
				if !prop.HasDescription || prop.Description != wantDescription {
					t.Fatalf("description %q present %v; want %q", prop.Description, prop.HasDescription, wantDescription)
				}
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatalf("run: %v", err)
			}
			if got := output.String(); got != string(want) {
				t.Fatalf("output %q; want %q", got, want)
			}
		})
	}
}

// Natural invalid-member recovery must preserve the pre-existing success of a
// field's trailing readonly directive rather than treating it as a new member.
func TestPropertyDescription_ClassFieldReadonlyFixture(t *testing.T) {
	source, err := os.ReadFile("../../testdata/fixtures/SimpleScripts/readonly_field.pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../testdata/fixtures/SimpleScripts/readonly_field.txt")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(string(source))
	if err != nil {
		t.Fatalf("compile readonly field fixture: %v", err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	// Upstream fixture text omits the final PrintLn newline.
	if got := strings.TrimRight(output.String(), "\r\n"); got != strings.TrimRight(string(want), "\r\n") {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestPropertyDescription_ClassExternalFieldExecution(t *testing.T) {
	const source = `type T = class F: Integer := 3; external "field"; readonly; property P: Integer read F description "text"; end; var O := new T; PrintLn(O.P);`
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "3\n" {
		t.Fatalf("got %q; want 3 newline", got)
	}
}
