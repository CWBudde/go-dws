package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/printer"
)

func TestCompile_PropertyAccessorRecoveryFixtures(t *testing.T) {
	for _, name := range []string{"missing_reader_bracket", "null_write_expression", "null_read_expression"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("../../testdata/fixtures/PropertyExpressionsFail", name)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			result := CompileWithOptions(string(source), Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			want := strings.TrimSpace(string(expected))
			assertPropertyAccessorRecoveryDiagnostics(t, result, want, name)
			assertPropertyAccessorRecoveryAST(t, result, name)
		})
	}
}

func assertPropertyAccessorRecoveryDiagnostics(t *testing.T, result *Result, want, name string) {
	t.Helper()
	if got := strings.Join(result.DiagnosticStrings(), "\n"); got != want {
		t.Errorf("complete rendered list:\ngot:\n%s\nwant:\n%s", got, want)
	}
	raw := make([]string, len(result.Diagnostics))
	for i, d := range result.Diagnostics {
		raw[i] = d.Render()
		if d.Stop != (name != "null_write_expression" && i == len(result.Diagnostics)-1) {
			t.Errorf("diagnostic %d Stop=%t", i, d.Stop)
		}
	}
	if strings.Join(raw, "\n") != want {
		t.Errorf("complete raw list: %v", raw)
	}
	assertPropertyAccessorRecoveryDiagnosticTuples(t, result, strings.Split(want, "\n"))
}

func assertPropertyAccessorRecoveryDiagnosticTuples(t *testing.T, result *Result, lines []string) {
	t.Helper()
	if len(result.Diagnostics) == len(lines) {
		for i, line := range lines {
			at := strings.LastIndex(line, " [line: ")
			var row, column int
			if _, err := fmt.Sscanf(line[at:], " [line: %d, column: %d]", &row, &column); err != nil {
				t.Fatal(err)
			}
			message := strings.TrimPrefix(strings.TrimPrefix(line[:at], "Syntax Error: "), "Warning: ")
			severity, phase := SeverityError, PhaseParsing
			if strings.HasPrefix(line, "Warning: ") {
				severity, phase = SeverityWarning, PhaseSemantic
				message = line[:at]
			}
			d := result.Diagnostics[i]
			if d.Message != message || d.Line != row || d.Column != column || d.Severity != severity || d.Phase != phase {
				t.Errorf("raw diagnostic %d = %#v; want %s", i, d, line)
			}
		}
	}
}

func assertPropertyAccessorRecoveryAST(t *testing.T, result *Result, name string) {
	t.Helper()
	if result.Program == nil || !result.SemanticAttempted {
		t.Fatalf("reached AST/analysis missing: %+v", result)
	}
	var class *ast.ClassDecl
	ast.Inspect(result.Program, func(n ast.Node) bool {
		if c, ok := n.(*ast.ClassDecl); ok {
			class = c
			return false
		}
		return true
	})
	if class == nil {
		t.Fatal("reached class missing")
	}
	if name == "missing_reader_bracket" {
		assertMissingReaderRecovery(t, result, class)
	}
	if name == "null_write_expression" {
		if len(class.Properties) != 1 || class.Properties[0].WriteStmt == nil || class.Properties[0].WriteSpec != nil || class.Properties[0].IsAutoProperty {
			t.Errorf("null writer not retained: %+v", class.Properties)
		}
	}
}

func assertMissingReaderRecovery(t *testing.T, result *Result, class *ast.ClassDecl) {
	t.Helper()
	if len(class.Properties) != 5 || len(class.Fields) != 1 {
		t.Errorf("retained fields/properties: %d/%d", len(class.Fields), len(class.Properties))
	}
	for _, p := range class.Properties {
		if p.ReadSpec == nil || p.IsAutoProperty {
			t.Errorf("lost reader or fabricated storage: %s", p.Name.Value)
		}
	}
	if len(class.Properties) == 5 {
		assertMissingReaderSourceAccessors(t, class)
	}
	registered := result.Analyzer.GetClasses()[ident.Normalize("TBase")]
	if registered == nil || len(registered.Properties) != 5 || len(registered.Fields) != 1 {
		t.Errorf("reached symbols lost: %+v", registered)
	}
	// Recovery ASTs remain safely printable even when compilation is rejected.
	_ = printer.New(printer.DefaultOptions()).Print(result.Program)
	_ = result.Program.String()
}

func assertMissingReaderSourceAccessors(t *testing.T, class *ast.ClassDecl) {
	t.Helper()
	for i, name := range []string{"Test", "Direct", "Test2", "Test3", "Test4"} {
		if class.Properties[i].Name.Value != name {
			t.Errorf("property %d lost source name: %s", i, class.Properties[i].Name.Value)
		}
	}
	null, ok := class.Properties[4].WriteStmt.(*ast.EmptyStatement)
	if !ok || null.Pos().Line != 8 || null.Pos().Column != 53 {
		t.Errorf("null writer lost classification/anchor: %+v", class.Properties[4].WriteStmt)
	}
	if _, ok := class.Properties[3].ReadSpec.(*ast.GroupedExpression).Expression.(*ast.GroupedExpression); !ok {
		t.Error("successful inner group lost")
	}
}

func TestCompile_PropertyAccessorRecoveryBoundaries(t *testing.T) {
	for _, tt := range []struct {
		name, clause string
		messages     []string
		stop         bool
	}{
		{"outer reader", "read (F;", []string{`")" expected`}, false},
		{"successful inner reader", "read ((F);", []string{`")" expected`}, false},
		{"missing inner reader", "read ((F;", []string{`")" expected`}, true},
		{"outer before write", "read (F write F;", []string{`")" expected`}, false},
		{"writer assignment", "read F write (F := Value;", []string{`")" expected`}, false},
		{"writer lvalue", "read F write (F;", []string{`")" expected`}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type T = class\n F: Integer;\n property P: Integer " + tt.clause + "\n property Q: Integer read F;\nend;\nMissing;"
			line := strings.Split(source, "\n")[2]
			col := strings.Index(line, ";") + 1
			if strings.Contains(tt.clause, " write F;") {
				col = strings.Index(line, "write") + 1
			}
			want := []string{fmt.Sprintf("Syntax Error: %s [line: 3, column: %d]", tt.messages[0], col)}
			if !tt.stop {
				want = append(want, `Syntax Error: Unknown name "Missing" [line: 6, column: 1]`)
			}
			r := CompileWithOptions(source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if got := r.DiagnosticStrings(); strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Errorf("full list: %v; want %v", got, want)
			}
			if r.HasParserStop() != tt.stop {
				t.Errorf("parser stop=%t", r.HasParserStop())
			}
			var c *ast.ClassDecl
			ast.Inspect(r.Program, func(n ast.Node) bool {
				if x, ok := n.(*ast.ClassDecl); ok {
					c = x
					return false
				}
				return true
			})
			count := 2
			if tt.stop {
				count = 1
			}
			if c == nil || len(c.Properties) != count {
				t.Fatalf("retained class=%+v; want %d properties", c, count)
			}
			if !tt.stop {
				ct := r.Analyzer.GetClasses()[ident.Normalize("T")]
				if ct == nil || len(ct.Properties) != 2 {
					t.Errorf("registered class/properties=%+v", ct)
				}
			}
		})
	}
}

func TestCompile_PropertyWriterInstructionControls(t *testing.T) {
	for _, tt := range []struct {
		name, body, prefix string
		want               []string
	}{
		{"empty", "", "", []string{`Warning: Property writer does nothing [line: 3, column: 35]`}},
		{"comment", "{padding}\n {padding}", "", []string{`Warning: Property writer does nothing [line: 3, column: 35]`}},
		{"numeric untouched", "2", "", []string{`Warning: Property writer does nothing [line: 3, column: 35]`, `Syntax Error: ")" expected [line: 3, column: 36]`, `Syntax Error: ";" expected [line: 3, column: 36]`, `Syntax Error: Name expected [line: 3, column: 36]`}},
		{"field", "F", "", nil}, {"Value", "Value", "", []string{`Hint: Assigning Value to itself [line: 3, column: 35]`}}, {"grouped field", "(F)", "", nil}, {"assignment", "F := Value", "", nil}, {"call", "PrintLn(Value)", "", nil},
		{"constant group", "(2)", "", []string{`Hint: Constant Instruction - has no effect [line: 3, column: 36]`, `Warning: Property writer does nothing [line: 3, column: 35]`}},
		{"named constant", "K", "const K=2;\n", []string{`Hint: Constant Instruction - has no effect [line: 4, column: 36]`, `Warning: Property writer does nothing [line: 4, column: 35]`}},
		{"stateless call", "IntToStr(2)", "", []string{`Hint: Constant Instruction - has no effect [line: 3, column: 36]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.prefix + "type T = class\n F: Integer;\n property P: Integer read F write (" + tt.body + ");\nend;"
			want := tt.want
			r := CompileWithOptions(source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if strings.Join(r.DiagnosticStrings(), "\n") != strings.Join(want, "\n") {
				t.Errorf("got %v; want %v", r.DiagnosticStrings(), want)
			}
		})
	}
}

func TestCompile_PropertyAccessorRecoveryOtherOwners(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"record null", "type R = record\n property P: Integer write ();\nend;", `Warning: Property writer does nothing [line: 2, column: 28]`},
		{"helper null", "type T = class end;\ntype H = class helper for T\n property P: Integer write ();\nend;", `Warning: Property writer does nothing [line: 3, column: 28]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := CompileWithOptions(tt.source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if got := strings.Join(r.DiagnosticStrings(), "\n"); got != tt.want {
				t.Errorf("complete list got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestCompile_PropertyAccessorRecoveryStopsAndPrefix(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"unknown reader stops punctuation", "type T = class\n property P: Integer read (Unknown;\n property Q: Integer;\nend;\nMissing;", []string{`Syntax Error: Unknown name "Unknown" [line: 2, column: 28]`}},
		{"call reader stops punctuation", "type T = class\n property P: Integer read (Ord(;\n property Q: Integer;\nend;\nMissing;", []string{`Syntax Error: Expression expected [line: 2, column: 32]`}},
		{"completed routine hint and later cutoff", "procedure Prefix;\nbegin var Local: Integer; end;\nprocedure Pending; forward;\ntype T = class\n property P: Integer write (2);\n procedure Late; begin var Later: Integer; end;\nend;\nMissing; {$ERROR 'later'}", []string{`Hint: Variable "Local" declared but not used [line: 2, column: 11]`, `Warning: Property writer does nothing [line: 5, column: 28]`, `Syntax Error: ")" expected [line: 5, column: 29]`, `Syntax Error: ";" expected [line: 5, column: 29]`, `Syntax Error: Name expected [line: 5, column: 29]`}},
		{"missing close before END", "type T = class\n F: Integer;\n property P: Integer read (F\nend;", []string{`Syntax Error: ")" expected [line: 4, column: 1]`, `Syntax Error: ";" expected [line: 4, column: 1]`}},
		{"empty reader", "type T = class\n property P: Integer read ();\nend;\nMissing;", []string{`Syntax Error: Expression expected [line: 2, column: 28]`}},
		{"invalid explicit assignment is not null", "const K=2;\ntype T = class\n property P: Integer write (K := Value);\nend;", []string{`Syntax Error: Cannot assign a value to the left-side argument [line: 3, column: 31]`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := CompileWithOptions(tt.source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if got := strings.Join(r.DiagnosticStrings(), "\n"); got != strings.Join(tt.want, "\n") {
				t.Errorf("complete list got:\n%s\nwant:\n%s", got, strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestCompile_PropertyWriterInnerStopRetainsReachedDeclaration(t *testing.T) {
	for _, body := range []string{"(F;", "PrintLn(;"} {
		t.Run(body, func(t *testing.T) {
			source := "type T = class F: Integer; property P: Integer read F write (" + body + " property Q: Integer; end; Missing;"
			r := CompileWithOptions(source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if len(r.Diagnostics) != 1 || !r.Diagnostics[0].Stop {
				t.Errorf("full stopped list=%#v", r.Diagnostics)
			}
			c := r.Program.Statements[0].(*ast.ClassDecl)
			if len(c.Properties) != 1 || c.Properties[0].Name.Value != "P" || c.Properties[0].ReadSpec == nil || c.Properties[0].IsAutoProperty {
				t.Fatalf("reached property/reader lost: %+v", c.Properties)
			}
		})
	}
}

func TestCompile_PropertyAccessorRecoveryEOF(t *testing.T) {
	for _, tt := range []struct {
		name, clause string
		column       int
		null         bool
	}{
		{"reader", "read (F", 28, false}, {"writer", "write (F", 29, false}, {"empty writer", "write (", 28, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type T = class\n F: Integer;\n property P: Integer " + tt.clause
			var want []string
			if tt.null {
				want = append(want, fmt.Sprintf("Warning: Property writer does nothing [line: 3, column: %d]", tt.column))
			}
			for _, m := range []string{`")" expected`, `";" expected`, `Name expected`} {
				want = append(want, fmt.Sprintf("Syntax Error: %s [line: 3, column: %d]", m, tt.column))
			}
			r := CompileWithOptions(source, Options{HintsLevel: semantic.HintsLevelPedantic, DisableSymbolDictionaryDiagnostics: false})
			if got := strings.Join(r.DiagnosticStrings(), "\n"); got != strings.Join(want, "\n") {
				t.Errorf("full EOF list got:\n%s\nwant:\n%s", got, strings.Join(want, "\n"))
			}
		})
	}
}

func TestCompile_PropertyNullWriterWarningOptions(t *testing.T) {
	const source = "type T = class\n property P: Integer write ({padding});\nend;"
	for _, hints := range []semantic.HintsLevel{semantic.HintsLevelDisabled, semantic.HintsLevelNormal, semantic.HintsLevelPedantic} {
		for _, disabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("hints=%d/dictionaryDisabled=%t", hints, disabled), func(t *testing.T) {
				r := CompileWithOptions(source, Options{HintsLevel: hints, DisableSymbolDictionaryDiagnostics: disabled})
				want := `Warning: Property writer does nothing [line: 2, column: 28]`
				if got := strings.Join(r.DiagnosticStrings(), "\n"); got != want || r.HasFatalDiagnostics() {
					t.Errorf("warning depends on options: %q", got)
				}
			})
		}
	}
}
