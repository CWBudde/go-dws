package dwscript

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestForwardDefaultDirection_ExportStops(t *testing.T) {
	for _, suffix := range []string{"export; begin end;", "export;", "export", "export 42; cdecl; begin Bad; end; {$ERROR 'late'} Later;"} {
		t.Run(suffix, func(t *testing.T) {
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile("procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Integer); " + suffix)
			var compileErr *CompileError
			if program != nil || !errors.As(err, &compileErr) {
				t.Fatalf("program %v, error %v", program, err)
			}
			if len(compileErr.Errors) != 1 {
				t.Fatalf("complete diagnostics: %+v", compileErr.Errors)
			}
			got := compileErr.Errors[0]
			if got.Message != "BEGIN expected" || got.Line != 2 || got.Column != 26 || !got.IsError() {
				t.Fatalf("complete diagnostic: %+v", got)
			}
		})
	}
}

func TestForwardDefaultDirection_SourceAndPrinter(t *testing.T) {
	const source = "procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; P; P(7);"
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if err != nil || program == nil {
		t.Fatalf("Compile: %v", err)
	}
	forward := program.AST().Statements[0].(*ast.FunctionDecl)
	implementation := program.AST().Statements[1].(*ast.FunctionDecl)
	originalDefault := forward.Parameters[0].DefaultValue
	originalParameter := implementation.Parameters[0]
	originalBody := implementation.Body
	originalString := implementation.String()
	if originalDefault == nil || originalParameter.DefaultValue != nil || strings.Contains(originalString, " = ") {
		t.Fatalf("source defaults: %s / %s", forward.String(), originalString)
	}
	rendered := printer.New(printer.DefaultOptions()).Print(program.AST())
	for i := 0; i < 2; i++ {
		if _, err := engine.Run(program); err != nil {
			t.Fatal(err)
		}
		if forward.Parameters[0].DefaultValue != originalDefault || implementation.Parameters[0] != originalParameter || originalParameter.DefaultValue != nil || implementation.Body != originalBody || implementation.String() != originalString || printer.New(printer.DefaultOptions()).Print(program.AST()) != rendered {
			t.Fatal("Run mutated source AST or printing")
		}
	}
	assertForwardDefaultDirectionPrinterRoundtrip(t, engine, &output, rendered)
}

func assertForwardDefaultDirectionPrinterRoundtrip(t *testing.T, engine *Engine, output *bytes.Buffer, rendered string) {
	t.Helper()
	if got := output.String(); got != "1\n7\n1\n7\n" {
		t.Fatalf("repeat Run output %q", got)
	}
	output.Reset()
	program, err := engine.Compile(rendered)
	if err != nil || program == nil {
		t.Fatalf("printer Compile: %v\n%s", err, rendered)
	}
	if program.AST().Statements[1].(*ast.FunctionDecl).Parameters[0].DefaultValue != nil {
		t.Fatal("printer inserted syntactically omitted default")
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "1\n7\n" {
		t.Fatalf("printer Run output %q", got)
	}
}

func TestForwardDefaultDirection_RejectedControls(t *testing.T) {
	type diagnostic struct {
		message      string
		line, column int
	}
	for _, tt := range []struct {
		name, source string
		want         []diagnostic
	}{
		{"reverse adding default", "procedure P(X: Integer); overload; forward;\nprocedure P(X: Integer = 1); overload; export; begin end;", []diagnostic{{"Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version", 2, 1}, {"The function \"P\" was forward declared but not implemented", 1, 11}}},
		{"missing directive preserves original", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String); export; begin end;\nprocedure P(X: Integer); begin end;", []diagnostic{{"Syntax Error: Overloaded procedure \"P\" must be marked with the \"overload\" directive", 2, 1}}},
		{"ambiguous member preserves original", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: String = 's'); overload; export; begin end;\nprocedure P(X: Integer); begin end;", []diagnostic{{"Syntax Error: Overload of \"P\" will be ambiguous with a previously declared version", 2, 1}}},
		{"only selected forward cleared", "procedure P(X: Integer = 1); overload; forward;\nprocedure P(X: Boolean); overload; forward;\nprocedure P(X: Integer); begin end;", []diagnostic{{"The function \"P\" was forward declared but not implemented", 2, 11}}},
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
				t.Fatalf("complete diagnostics: %+v", compileErr.Errors)
			}
			for i, want := range tt.want {
				got := compileErr.Errors[i]
				if got.Message != want.message || got.Line != want.line || got.Column != want.column || !got.IsError() {
					t.Fatalf("complete diagnostic[%d]: %+v; want %+v", i, got, want)
				}
			}
		})
	}
}

func TestForwardDefaultDirection_Run(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"omitted directive", "procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; P; P(7);", "1\n7\n"},
		{"repeated directive", "procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer); overload; begin PrintLn(X); end; P; P(7);", "1\n7\n"},
		{"explicit empty arguments", "procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; P(); P(7);", "1\n7\n"},
		{"zero", "procedure P(X: Integer = 0); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; P;", "0\n"},
		{"false", "procedure P(X: Boolean = False); overload; forward; procedure P(X: Boolean); begin PrintLn(X); end; P; P(True);", "False\nTrue\n"},
		{"empty", "procedure P(X: String = ''); overload; forward; procedure P(X: String); begin PrintLn('['+X+']'); end; P; P('s');", "[]\n[s]\n"},
		{"mixed slots", "procedure P(A: Integer; B: Integer = 2; C: Integer = 3); overload; forward; procedure P(A: Integer; B: Integer; C: Integer = 3); begin PrintLn(A+B+C); end; P(1); P(1,4); P(1,4,5);", "6\n8\n10\n"},
		{"case only names", "procedure Pending(X: Integer = 1); overload; forward; procedure pENDING(x: Integer); begin PrintLn(x); end; PENDING;", "1\n"},
		{"different new member", "procedure P(X: Integer = 1); overload; forward; procedure P(X: String); overload; export; begin PrintLn(X); end; procedure P(X: Integer); begin PrintLn(X); end; P; P(7); P('s');", "1\n7\ns\n"},
		{"constant snapshot survives conversion", "const D = 4; procedure P(X: Integer = D); overload; forward; procedure P(X: String); overload; export; begin PrintLn(X); end; procedure P(X: Integer); begin PrintLn(X); end; procedure Outer; begin const D = 9; P(); P('s'); end; Outer; P(7);", "4\ns\n7\n"},
		{"genuine set", "procedure P(X: Integer = 1); overload; forward; procedure P(X: Boolean); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; procedure P(X: Boolean); begin PrintLn('bool'); end; P; P(True);", "1\nbool\n"},
		{"constant declaration context", "const D = 4; procedure P(X: Integer = D); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; procedure Outer; begin const D = 9; P; end; Outer; P;", "4\n4\n"},
		{"folded constant declaration context", "const D = 4; procedure P(X: Integer = D+2); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; procedure Outer; begin const D = 9; P(); end; Outer; P;", "6\n6\n"},
		{"float constant declaration context", "const D = 2.5; procedure P(X: Float = D); overload; forward; procedure P(X: Float); begin PrintLn(X); end; procedure Outer; begin const D = 9.5; P(); end; Outer; P(7.5);", "2.5\n7.5\n"},
		{"mixed Float arithmetic declaration context", "const D = 1; procedure P(X: Float = D+0.5); overload; forward; procedure P(X: Float); begin PrintLn(X); end; procedure Outer; begin const D = 9; P(); end; Outer; P(2.5);", "1.5\n2.5\n"},
		{"false constant declaration context", "const D = False; procedure P(X: Boolean = D); overload; forward; procedure P(X: Boolean); begin PrintLn(X); end; procedure Outer; begin const D = True; P(); end; Outer; P(True);", "False\nTrue\n"},
		{"empty constant declaration context", "const D = ''; procedure P(X: String = D); overload; forward; procedure P(X: String); begin PrintLn('['+X+']'); end; procedure Outer; begin const D = 'shadow'; P(); end; Outer; P('s');", "[]\n[s]\n"},
		{"nested matched forward", "procedure Outer; begin const D = 4; procedure P(X: Integer = D); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; procedure Inner; begin const D = 9; P(); end; Inner; P(7); end; Outer;", "4\n7\n"},
		{"local shadow inside Outer begin", "procedure P(X: Integer = 1); overload; forward; procedure Outer; begin procedure P(X: String); export; begin PrintLn(X); end; P('local'); end; procedure P(X: Integer); begin PrintLn(X); end; Outer; P;", "local\n1\n"},
		{"equal present unchanged", "procedure P(X: Integer = 1); overload; forward; procedure P(X: Integer = 1); begin PrintLn(X); end; P;", "1\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tt.source)
			if err != nil || program == nil {
				t.Fatalf("Compile: program %v, complete diagnostics %v", program, err)
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

func TestForwardDefaultDirection_IntegerArithmetic(t *testing.T) {
	for _, tt := range []struct{ name, expression, want string }{
		{"large addition", "9007199254740993 + 0", "9007199254740993\n"},
		{"precision boundary addition", "9007199254740991 + 2", "9007199254740993\n"},
		{"large subtraction", "9007199254740993 - 0", "9007199254740993\n"},
		{"large multiplication", "3000000001 * 3000000001", "9000000006000000001\n"},
		{"negative addition", "-9007199254740993 + 0", "-9007199254740993\n"},
		{"maximum signed integer", "9223372036854775807 - 0", "9223372036854775807\n"},
		{"near minimum signed integer", "-9223372036854775807 + 1", "-9223372036854775806\n"},
		{"large division", "9007199254740993 div 1", "9007199254740993\n"},
		{"large modulo", "9007199254740993 mod 2", "1\n"},
		{"negative division", "-9007199254740993 div 1", "-9007199254740993\n"},
		{"negative modulo", "-9007199254740993 mod 2", "-1\n"},
		{"signed overflow", "9223372036854775807 + 1", "-9223372036854775808\n"},
	} {
		for _, forward := range []bool{false, true} {
			name := "direct implementation/"
			if forward {
				name = "omitted forward/"
			}
			t.Run(name+tt.name, func(t *testing.T) {
				source := "procedure P(X: Integer = " + tt.expression + "); begin PrintLn(X); end; P;"
				if forward {
					source = "procedure P(X: Integer = " + tt.expression + "); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; P;"
				}
				var output bytes.Buffer
				engine, err := New(WithOutput(&output))
				if err != nil {
					t.Fatal(err)
				}
				program, err := engine.Compile(source)
				if err != nil || program == nil {
					t.Fatalf("Compile: program %v, complete diagnostics %v", program, err)
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
}

func TestForwardDefaultDirection_LargeConstantDeclarationScope(t *testing.T) {
	const source = "const D = 9007199254740991 + 2; procedure P(X: Integer = D + 0); overload; forward; procedure P(X: Integer); begin PrintLn(X); end; procedure Outer; begin const D = 1; P(); end; Outer; P(7); P;"
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if err != nil || program == nil {
		t.Fatalf("Compile: program %v, complete diagnostics %v", program, err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "9007199254740993\n7\n9007199254740993\n" {
		t.Fatalf("declaration-bound output %q", got)
	}
}

func TestForwardDefaultDirection_MixedConstantComposition(t *testing.T) {
	for _, tt := range []struct{ name, source, runtimeOutput string }{
		{"standalone acceptance", "const C = 5 mod 2.5; PrintLn(C);", "0\n"},
		{"composed acceptance", "const C = (5 mod 2.5) + 1; PrintLn(C);", "1\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(tt.source)
			if err != nil {
				var compileErr *CompileError
				if program != nil || !errors.As(err, &compileErr) {
					t.Fatalf("rejection: program %v, error %v", program, err)
				}
				t.Fatalf("want accepted nonnil Program and empty rejection list; complete raw diagnostics: %+v", compileErr.Errors)
			}
			if program == nil {
				t.Fatal("accepted Compile returned nil Program")
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			// Preserve runtime execution along with Compile acceptance.
			if got := output.String(); got != tt.runtimeOutput {
				t.Fatalf("runtime output %q; want %q", got, tt.runtimeOutput)
			}
		})
	}
}

func TestForwardDefaultDirection_FloatModuloDefaults(t *testing.T) {
	for _, tt := range []struct{ expression, want string }{
		{"5 mod 2.5", "0\n7.5\n"},
		{"5.5 mod 2", "1.5\n7.5\n"},
		{"5.5 mod 2.0", "1.5\n7.5\n"},
		{"(-5) mod 2.25", "-0.5\n7.5\n"},
		{"5.5 mod (-2)", "1.5\n7.5\n"},
		{"(-5) mod 2.5", "0\n7.5\n"},
		{"(5 mod 2.5) + 1", "1\n7.5\n"},
	} {
		for _, forward := range []bool{false, true} {
			name := "direct implementation/"
			if forward {
				name = "omitted forward/"
			}
			t.Run(name+tt.expression, func(t *testing.T) {
				source := "procedure P(X: Float = " + tt.expression + "); begin PrintLn(X); end; P; P(7.5);"
				if forward {
					source = "procedure P(X: Float = " + tt.expression + "); overload; forward; procedure P(X: Float); begin PrintLn(X); end; P; P(7.5);"
				}
				var output bytes.Buffer
				engine, err := New(WithOutput(&output))
				if err != nil {
					t.Fatal(err)
				}
				program, err := engine.Compile(source)
				if err != nil || program == nil {
					t.Fatalf("Compile: program %v, complete diagnostics %v", program, err)
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
}

func TestForwardDefaultDirection_FloatModuloConstantDeclarationScope(t *testing.T) {
	const source = "const D = 5.5 mod 2; procedure P(X: Float = D+0.5); overload; forward; procedure P(X: Float); begin PrintLn(X); end; procedure Outer; begin const D = 9; P(); end; Outer; P(7.5); P;"
	var output bytes.Buffer
	engine, err := New(WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if err != nil || program == nil {
		t.Fatalf("Compile: program %v, complete diagnostics %v", program, err)
	}
	if _, err := engine.Run(program); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "2\n7.5\n2\n" {
		t.Fatalf("declaration-bound output %q", got)
	}
}
