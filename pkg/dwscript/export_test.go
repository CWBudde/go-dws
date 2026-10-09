package dwscript

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExport_Execution(t *testing.T) {
	tests := []struct{ name, directive string }{
		{"bare", "export"}, {"named", "export 'PublicName'"}, {"empty", "ExPoRt ''"},
		{"escaped", "export 'can''t'"}, {"multiline", "export \"first\nsecond\""},
		{"inline", "export 'PublicName'; inline"}, {"overload", "overload; export"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			source := "procedure P; " + tt.directive + "; begin PrintLn('P'); end; function F: Integer; export 'PublicF'; begin Result := 42; end; P; PrintLn(F());"
			program, err := engine.Compile(source)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "P\n42\n" {
				t.Fatalf("got %q; want P and 42", got)
			}
			output.Reset()
			if _, err := engine.Eval(source); err != nil {
				t.Fatalf("Eval: %v", err)
			}
			if got := output.String(); got != "P\n42\n" {
				t.Fatalf("Eval output %q", got)
			}
		})
	}
}

func TestExport_UnitInterfaceExecution(t *testing.T) {
	dir := t.TempDir()
	const source = "unit U; interface function F: Integer; export 'PublicF'; implementation function F: Integer; begin Result := 42; end; end."
	if err := os.WriteFile(filepath.Join(dir, "U.dws"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	engine, err := New(WithUnitSearchPaths(dir), WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Eval("uses U; PrintLn(U.F());"); err != nil {
		t.Fatalf("unit compile/run: %v", err)
	}
	if got := output.String(); got != "42\n" {
		t.Fatalf("output %q; want 42", got)
	}
}

func TestExport_LocalsContractsAndForwardExecution(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"locals", "procedure P; export; const N = 7; var X: Integer := N; begin PrintLn(X); end; P;", "7\n"},
		{"contracts", "function F(X: Integer): Integer; export; require X > 0; begin Result := X+1; ensure Result > X; end; PrintLn(F(41));", "42\n"},
		{"forward", "procedure P; forward; export 'PublicName'; procedure P; begin PrintLn('P'); end; P;", "P\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Eval(tt.source); err != nil {
				t.Fatalf("Eval: %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestExport_CompileReturnsCompleteErrors(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile("procedure P; export Foo; begin Bad; end;")
	var compileErr *CompileError
	if program != nil || !errors.As(err, &compileErr) {
		t.Fatalf("program %v error %v", program, err)
	}
	want := []string{`";" expected`, "BEGIN expected"}
	if len(compileErr.Errors) != len(want) {
		t.Fatalf("complete errors: %v", compileErr.Errors)
	}
	for i, e := range compileErr.Errors {
		if e.Message != want[i] || e.Line != 1 || e.Column != 21 || !e.IsError() {
			t.Fatalf("error %d: %+v", i, e)
		}
	}
}

func TestExport_EOFRequiresBody(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		messages     []string
		line, column int
	}{
		{"semicolon", "procedure P; export;", []string{"BEGIN expected"}, 1, 20},
		{"no semicolon", "procedure P; export", []string{`";" expected`, "BEGIN expected"}, 1, 14},
		{"name without semicolon", "procedure P; export 'name'", []string{`";" expected`, "BEGIN expected"}, 1, 21},
		{"named semicolon", "procedure P; export 'name';", []string{"BEGIN expected"}, 1, 27},
		{"comments and newlines", "procedure P;\n export; {comment}\n // trailing\n", []string{"BEGIN expected"}, 2, 8},
	} {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tt.source)
			var compileErr *CompileError
			if program != nil || !errors.As(err, &compileErr) {
				t.Fatalf("program %v error %v", program, err)
			}
			if len(compileErr.Errors) != len(tt.messages) {
				t.Fatalf("complete errors: %v; want %v", compileErr.Errors, tt.messages)
			}
			for i, e := range compileErr.Errors {
				if e.Message != tt.messages[i] || e.Line != tt.line || e.Column != tt.column || !e.IsError() {
					t.Fatalf("error %d: %+v", i, e)
				}
			}
		})
	}
}
