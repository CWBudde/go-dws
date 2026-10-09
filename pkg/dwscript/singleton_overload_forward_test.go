package dwscript

import (
	"bytes"
	"errors"
	"testing"
)

// A different parameter type must add an overload without consuming the lone
// explicitly overloaded forward. Its eventual implementation needs no directive.
func TestSingletonOverloadForward_DifferentParameterType(t *testing.T) {
	for _, export := range []string{"", "export; "} {
		t.Run(export, func(t *testing.T) {
			const prefix = "procedure P(x: Integer); overload; forward;\n"
			source := prefix + "procedure P(x: String); overload; " + export + "begin PrintLn('string'); end;\n" +
				"procedure P(x: Integer); begin PrintLn('integer'); end;"
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if err != nil || program == nil {
				t.Fatalf("Compile: program %v, complete errors %v", program, err)
			}
			program, err = engine.Compile(source + "\nP(1); P('x');")
			if err != nil || program == nil {
				t.Fatalf("Compile calls: program %v, complete errors %v", program, err)
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "integer\nstring\n" {
				t.Fatalf("output %q; want integer and string implementations", got)
			}
		})
	}
}

func TestSingletonOverloadForward_LegalControls(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"matched without repeated directive", "procedure P(x: Integer); overload; forward; procedure P(x: Integer); begin PrintLn(x); end; P(3);", "3\n"},
		{"case only names", "procedure Pending(X: Integer); overload; forward; procedure pENDING(x: String); overload; export; begin PrintLn(x); end; procedure PENDING(x: Integer); begin PrintLn(x); end; Pending(7); Pending('s');", "7\ns\n"},
		{"nested shadow inside body", "procedure P(x: Integer); overload; forward; procedure Outer; begin procedure P(x: String); export; begin PrintLn(x); end; P('nested'); end; procedure P(x: Integer); begin PrintLn(x); end; Outer; P(7);", "nested\n7\n"},
		{"genuine set differing types", "procedure P(x: Integer); overload; forward; procedure P(x: Boolean); overload; forward; procedure P(x: String); overload; export; begin PrintLn(x); end; procedure P(x: Integer); begin PrintLn(x); end; procedure P(x: Boolean); begin PrintLn('boolean'); end; P(3); P(True); P('string');", "3\nboolean\nstring\n"},
		{"function differing types", "function F(x: Integer): Integer; overload; forward; function F(x: String): String; overload; export; begin Result := x; end; function F(x: Integer): Integer; begin Result := x; end; PrintLn(F(3)); PrintLn(F('string'));", "3\nstring\n"},
		{"matched default presence", "procedure P(x: Integer = 1); overload; forward; procedure P(x: Integer = 1); begin PrintLn(x); end; P;", "1\n"},
		{"exported forward calling hint", "procedure P(x: Integer); overload; forward; export; cdecl; procedure P(x: String); overload; export; begin PrintLn(x); end; procedure P(x: Integer); begin PrintLn(x); end; P(3); P('string');", "3\nstring\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tt.source)
			if err != nil || program == nil {
				t.Fatalf("Compile: program %v, complete errors %v", program, err)
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("output %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSingletonOverloadForward_RejectedControls(t *testing.T) {
	type diagnostic struct {
		message      string
		line, column int
	}
	for _, tt := range []struct {
		name, source string
		want         []diagnostic
	}{
		{"missing new member directive", "procedure P(x: Integer); overload; forward;\nprocedure P(x: String); export; begin end;\nprocedure P(x: Integer); begin end;", []diagnostic{{"Syntax Error: Overloaded procedure \"P\" must be marked with the \"overload\" directive", 2, 1}}},
		{"ambiguous defaults preserve original", "procedure P(x: Integer = 1); overload; forward;\nprocedure P(x: String = 's'); overload; export; begin end;\nprocedure P(x: Integer = 1); begin end;", []diagnostic{{"Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version", 2, 1}}},
		{"matched export cutoff", "procedure P(x: Integer); overload; forward;\nprocedure P(x: Integer); export; cdecl; begin Bad; end; Later;", []diagnostic{{"BEGIN expected", 2, 26}}},
		{"matched export EOF", "procedure P(x: Integer); overload; forward;\nprocedure P(x: Integer); export;", []diagnostic{{"BEGIN expected", 2, 26}}},
		{"ordinary mismatch", "procedure P(x: Integer); forward;\nprocedure P(x: String); begin end;", []diagnostic{{"Syntax Error: implementation signature for 'P' does not match forward declaration", 2, 1}}},
		{"ordinary mismatch export cutoff", "procedure P(x: Integer); forward;\nprocedure P(x: String); export; begin Bad; end; Later;", []diagnostic{{"Syntax Error: implementation signature for 'P' does not match forward declaration", 2, 1}, {"BEGIN expected", 2, 25}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tt.source)
			var compileErr *CompileError
			if program != nil || !errors.As(err, &compileErr) {
				t.Fatalf("program %v, error %v", program, err)
			}
			if len(compileErr.Errors) != len(tt.want) {
				t.Fatalf("complete errors: %v", compileErr.Errors)
			}
			for i, want := range tt.want {
				got := compileErr.Errors[i]
				if got.Message != want.message || got.Line != want.line || got.Column != want.column || !got.IsError() {
					t.Fatalf("complete error[%d]: %+v; want %+v", i, got, want)
				}
			}
		})
	}
}

func TestSingletonOverloadForward_PreservesPendingDiagnostic(t *testing.T) {
	for _, export := range []string{"", "export; "} {
		t.Run(export, func(t *testing.T) {
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile("procedure P(x: Integer); overload; forward;\nprocedure P(x: String); overload; " + export + "begin end;")
			var compileErr *CompileError
			if program != nil || !errors.As(err, &compileErr) {
				t.Fatalf("program %v, error %v", program, err)
			}
			if len(compileErr.Errors) != 1 {
				t.Fatalf("complete errors: %v", compileErr.Errors)
			}
			diagnostic := compileErr.Errors[0]
			if diagnostic.Message != "The function \"P\" was forward declared but not implemented" || diagnostic.Line != 1 || diagnostic.Column != 11 || !diagnostic.IsError() {
				t.Fatalf("complete diagnostic: %+v", diagnostic)
			}
		})
	}
}
