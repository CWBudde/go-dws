package dwscript

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExport_ForwardImplementationReturnsNilProgram(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		line, column int
	}{
		{"ordinary", "procedure P; forward;\nprocedure P; export; begin Bad; end; Later;", 2, 14},
		{"EOF after semicolon", "procedure P; forward;\nprocedure P; export;", 2, 14},
		{"EOF without semicolon", "procedure P; forward;\nprocedure P; export", 2, 14},
		{"EOF after name", "procedure P; forward;\nprocedure P; export 'name'", 2, 14},
		{"bad optional name", "procedure P; forward;\nprocedure P; export 42; begin Bad; end;", 2, 14},
		{"missing body", "procedure P; forward;\nprocedure P; export; PrintLn('late');", 2, 14},
		{"case insensitive", "procedure Pending; forward;\nprocedure pENDING; ExPoRt 'Again'; begin end;", 2, 20},
		{"unit", "unit U; interface procedure P; implementation\nprocedure P; export; begin Bad; end; end.", 2, 14},
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
			if len(compileErr.Errors) != 1 {
				t.Fatalf("complete errors: %v", compileErr.Errors)
			}
			e := compileErr.Errors[0]
			if e.Message != "BEGIN expected" || e.Line != tt.line || e.Column != tt.column || !e.IsError() {
				t.Fatalf("forward export error: %+v", e)
			}
		})
	}
}

func TestExport_ImportedForwardImplementationReturnsNilProgram(t *testing.T) {
	dir := t.TempDir()
	const source = "unit U; interface procedure P; implementation\nprocedure P; export; begin end; end."
	if err := os.WriteFile(filepath.Join(dir, "U.dws"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	engine, err := New(WithUnitSearchPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile("uses U;")
	var compileErr *CompileError
	if program != nil || !errors.As(err, &compileErr) {
		t.Fatalf("program %v error %v", program, err)
	}
	if len(compileErr.Errors) != 1 {
		t.Fatalf("complete errors: %v", compileErr.Errors)
	}
	e := compileErr.Errors[0]
	if e.Message != "BEGIN expected" || e.Line != 2 || e.Column != 14 || !e.IsError() {
		t.Fatalf("imported forward export error: %+v", e)
	}
}

func TestExport_ForwardContextLegalExecution(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"exported forward ordinary implementation", "procedure P; forward; export 'PublicP'; procedure P; begin PrintLn('P'); end; P;", "P\n"},
		{"different new overload", "procedure P(X: Integer); overload; forward; procedure P(X: String); overload; forward; procedure P(X: Float); overload; export; begin PrintLn('float'); end; procedure P(X: Integer); begin PrintLn('int'); end; procedure P(X: String); begin PrintLn('string'); end; P(1); P('x'); P(1.5);", "int\nstring\nfloat\n"},
		{"new overload return type", "function F(X: Integer): Integer; overload; forward; function F(X: String): Integer; overload; forward; function F(X: Integer): String; overload; export; begin Result := 'new'; end; function F(X: Integer): Integer; begin Result := 1; end; function F(X: String): Integer; begin Result := 2; end;", ""},
		{"nested shadow", "procedure P; forward; procedure Outer; begin procedure P; export; begin PrintLn('nested'); end; P; end; procedure P; begin PrintLn('outer'); end; Outer; P;", "nested\nouter\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := engine.Eval(tt.source); err != nil {
				t.Fatalf("execution: %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
