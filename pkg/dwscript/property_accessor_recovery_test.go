package dwscript

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestEngineCompile_PropertyAccessorRecoveryFixtures(t *testing.T) {
	for _, name := range []string{"missing_reader_bracket", "null_write_expression", "null_read_expression"} {
		t.Run(name, func(t *testing.T) {
			base := "../../testdata/fixtures/PropertyExpressionsFail/" + name
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			result := frontend.CompileWithOptions(string(source), frontend.Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if got := strings.Join(result.DiagnosticStrings(), "\n"); got != strings.TrimSpace(string(expected)) {
				t.Errorf("full frontend diagnostics:\n%s\nwant:\n%s", got, expected)
			}
			engine, err := New()
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(string(source))
			accepted := name == "null_write_expression"
			if (program != nil) != accepted || (err == nil) != accepted {
				t.Fatalf("program nil=%t, error=%v, accepted=%t", program == nil, err, accepted)
			}
			if !accepted {
				assertPropertyAccessorRecoveryCompileError(t, err, result)
			}
		})
	}
}

func assertPropertyAccessorRecoveryCompileError(t *testing.T, err error, result *frontend.Result) {
	t.Helper()
	ce, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("error=%T", err)
	}
	if len(ce.Errors) != len(result.Diagnostics) {
		t.Fatalf("public raw diagnostics=%v", ce.Errors)
	}
	for i, d := range ce.Errors {
		f := result.Diagnostics[i]
		if d.Message != f.Message || d.Line != f.Line || d.Column != f.Column || int(d.Severity) != int(f.Severity) {
			t.Errorf("raw diagnostic %d: %+v vs %+v", i, d, f)
		}
	}
}

func TestEngineRun_PropertyWriterInstructions(t *testing.T) {
	for _, checked := range []bool{true, false} {
		for _, tt := range []struct{ name, writer, want string }{
			{"empty", "()", "0\n"}, {"comment", "({comment})", "0\n"}, {"field", "(F)", "7\n"}, {"group", "((F))", "7\n"}, {"assignment", "(F := Value)", "7\n"}, {"call", "(PrintLn(Value))", "7\n0\n"}, {"named", "F", "7\n"}, {"compound", "(F += Value)", "7\n"}, {"conditional", "(if Value>0 then F := Value)", "7\n"}, {"block", "(begin F := Value; end)", "7\n"}, {"while", "(while F<Value do F += 1)", "7\n"}, {"repeat", "(repeat F += 1 until F=Value)", "7\n"}, {"try", "(try F := Value; finally Print(''); end)", "7\n"}, {"for", "(for var I := 1 to Value do F += 1)", "7\n"}, {"case", "(case Value of 7: F := Value; else F := 0; end)", "7\n"}, {"constant group", "((2))", "0\n"},
		} {
			t.Run(fmt.Sprintf("%s/checked=%t", tt.name, checked), func(t *testing.T) {
				source := "type T = class F: Integer; property P: Integer read F write " + tt.writer + "; end; var O := new T; O.P := 7; PrintLn(O.F);"
				var output bytes.Buffer
				engine, err := New(WithOutput(&output), WithTypeCheck(checked))
				if err != nil {
					t.Fatal(err)
				}
				program, err := engine.Compile(source)
				if err != nil || program == nil {
					t.Fatalf("Compile=%v, program nil=%t", err, program == nil)
				}
				if _, err := engine.Run(program); err != nil {
					t.Fatal(err)
				}
				if output.String() != tt.want {
					t.Errorf("Run output=%q want=%q", output.String(), tt.want)
				}
			})
		}
	}
}

func TestEngineRun_PropertyAccessorRecoveryPrintReparse(t *testing.T) {
	for _, body := range []string{"", "{padding}", "(2)", "K", "(K)", "F", "(F)", "F := Value", "PrintLn(Value)"} {
		t.Run(body, func(t *testing.T) {
			source := "const K=2; type T = class F: Integer; property P: Integer read (F) write (" + body + "); end; var O := new T; O.P := 7; PrintLn(O.F);"
			var original bytes.Buffer
			engine, err := New(WithOutput(&original))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(source)
			if err != nil {
				t.Fatal(err)
			}
			prop := program.AST().Statements[1].(*ast.ClassDecl).Properties[0]
			sourceTerm, writerStatement := prop.WriteSourceExpression, prop.WriteStmt
			before := prop.String()
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if prop.WriteSourceExpression != sourceTerm || prop.WriteStmt != writerStatement || prop.String() != before {
				t.Error("Run mutated source writer metadata")
			}
			rendered := printer.New(printer.DefaultOptions()).Print(program.AST())
			var output bytes.Buffer
			again, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			reparsed, err := again.Compile(rendered)
			if err != nil {
				t.Fatalf("reparse %s: %v", rendered, err)
			}
			if _, err := again.Run(reparsed); err != nil {
				t.Fatal(err)
			}
			if output.String() != original.String() {
				t.Errorf("source %q != printed %q (%s)", original.String(), output.String(), rendered)
			}
			property := reparsed.AST().Statements[1].(*ast.ClassDecl).Properties[0]
			if property.WriteStmt == nil || property.WriteSpec != nil || property.IsAutoProperty {
				t.Errorf("writer source shape lost: %+v", property)
			}
		})
	}
}

func TestEngineRun_PropertyNullWriterOtherOwners(t *testing.T) {
	for _, source := range []string{
		"type R = record F: Integer; property P: Integer write (); end; var O: R; O.P := 7; PrintLn(O.F);",
		"type T = class F: Integer; end; type H = class helper for T property P: Integer write (); end; var O := new T; O.P := 7; PrintLn(O.F);",
		"type R = record F: Integer; end; type H = record helper for R property P: Integer write (); end; var O: R; O.P := 7; PrintLn(O.F);",
	} {
		for _, checked := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/checked=%t", source, checked), func(t *testing.T) {
				var output bytes.Buffer
				engine, err := New(WithOutput(&output), WithTypeCheck(checked))
				if err != nil {
					t.Fatal(err)
				}
				program, err := engine.Compile(source)
				if err != nil || program == nil {
					t.Fatalf("Compile=%v", err)
				}
				if _, err := engine.Run(program); err != nil {
					t.Fatal(err)
				}
				if output.String() != "0\n" {
					t.Errorf("null writer had side effects: %q", output.String())
				}
			})
		}
	}
}

func TestEngineRun_PropertyAccessorRecoveryPassFixtures(t *testing.T) {
	for _, name := range []string{"simple_instance_expressions", "simple_record_expressions", "simple_interface_expressions", "helpers_property_expressions", "object_writer_statement", "record_write_statement", "double_brackets", "class_property_write_expressions"} {
		t.Run(name, func(t *testing.T) {
			base := "../../testdata/fixtures/PropertyExpressionsPass/" + name
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			engine, err := New(WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(string(source))
			if err != nil || program == nil {
				t.Fatalf("Compile=%v", err)
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(output.String()) != strings.TrimSpace(string(expected)) {
				t.Errorf("output %q want %q", output.String(), expected)
			}
		})
	}
}

func TestEngineRun_PropertyWriterRecordPrintReparse(t *testing.T) {
	for _, writer := range []string{"()", "((F))", "(F := Value)"} {
		for _, checked := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/checked=%t", writer, checked), func(t *testing.T) {
				source := "type R = record F: Integer; property P: Integer read (F) write " + writer + "; end; var O: R; O.P := 7; PrintLn(O.P);"
				var output bytes.Buffer
				engine, err := New(WithOutput(&output), WithTypeCheck(checked))
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
				want := output.String()
				output.Reset()
				rendered := printer.New(printer.DefaultOptions()).Print(program.AST())
				again, err := engine.Compile(rendered)
				if err != nil {
					t.Fatalf("printed record lost source accessors: %s: %v", rendered, err)
				}
				if _, err := engine.Run(again); err != nil {
					t.Fatal(err)
				}
				if output.String() != want {
					t.Errorf("printed record Run=%q source=%q", output.String(), want)
				}
				prop := again.AST().Statements[0].(*ast.RecordDecl).Properties[0]
				if prop.WriteStmt == nil || prop.WriteField != "" || prop.IsAutoProperty {
					t.Errorf("record expression writer shape lost: %+v", prop)
				}
			})
		}
	}
}
