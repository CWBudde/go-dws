package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_HeaderDiagnostics(t *testing.T) {
	for _, fixture := range []string{
		"FailureScripts/class_class", "FailureScripts/class_error1",
		"FailureScripts/record_syntax1", "FailureScripts/record_syntax2",
		"HelpersFail/helper_error5", "HelpersFail/helper_scopes1",
	} {
		t.Run(fixture, func(t *testing.T) {
			base := filepath.Join("../../testdata/fixtures", fixture)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertDiagnostics(t, string(source), base+".pas", strings.Split(strings.TrimSpace(string(expected)), "\n"))
		})
	}
}

func TestCompile_HelperVisibilityBeforeHeaderStop(t *testing.T) {
	source := "type T = helper for Integer\n public\n public\n protected\n private\n class bug;\nend;"
	for _, tt := range []struct {
		name  string
		want  []string
		level semantic.HintsLevel
	}{
		{name: "normal hints", level: semantic.HintsLevelNormal, want: []string{
			"Hint: Redundant specifier, visibility is already \"public\" [line: 2, column: 2]",
			"Hint: Redundant specifier, visibility is already \"public\" [line: 3, column: 2]",
			"Syntax Error: Helpers do not supported \"protected\" visibility specifier [line: 4, column: 2]",
			"Syntax Error: PROCEDURE or FUNCTION expected [line: 6, column: 8]",
		}},
		{name: "disabled hints", level: semantic.HintsLevelDisabled, want: []string{
			"Syntax Error: Helpers do not supported \"protected\" visibility specifier [line: 4, column: 2]",
			"Syntax Error: PROCEDURE or FUNCTION expected [line: 6, column: 8]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(source, "<test>", tt.level).DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("got %v; want %v", got, tt.want)
			}
		})
	}
}

func TestCompile_ClassHeaderStop(t *testing.T) {
	for _, kind := range []string{"class", "record", "helper for Integer"} {
		t.Run(kind, func(t *testing.T) {
			source := "type T = " + kind + "\n class {padding}\n bug;\n property P: Unknown;\nend; Missing;"
			assertDiagnostics(t, source, "<test>", []string{
				"Syntax Error: PROCEDURE or FUNCTION expected [line: 3, column: 2]",
			})
			assertDiagnostics(t, "type T = "+kind+"\n class {trailing}\n", "<test>", []string{
				"Syntax Error: PROCEDURE or FUNCTION expected [line: 2, column: 2]",
			})
		})
	}
	assertDiagnostics(t, "class {padding}\n bug; Missing;", "<test>", []string{
		"Syntax Error: PROCEDURE or FUNCTION expected [line: 2, column: 2]",
	})
	assertDiagnostics(t, "class {trailing}\n", "<test>", []string{
		"Syntax Error: PROCEDURE or FUNCTION expected [line: 1, column: 1]",
	})
}

func TestCompile_UnterminatedTypeBody(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"record", "type T = record\n F: Integer;", []string{
			"Syntax Error: END expected [line: 2, column: 12]",
		}},
		{"helper", "type T = helper for Integer\n class const C = 1;", []string{
			"Syntax Error: END expected [line: 2, column: 19]",
		}},
		{"helper unexpected member", "type T = helper for Integer\n bug;\n Missing;", []string{
			"Syntax Error: END expected [line: 2, column: 2]",
		}},
		{"earlier record duplicates", "type T = record\n MiXeD, mixed: Integer;\n class bug;\nend;", []string{
			"Syntax Error: There is already a field with name \"MiXeD\" [line: 2, column: 9]",
			"Syntax Error: PROCEDURE or FUNCTION expected [line: 3, column: 8]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertDiagnostics(t, tt.source, "<test>", tt.want) })
	}
}

func TestCompile_ValidClassMemberHeaders(t *testing.T) {
	for _, kind := range []string{"class", "record", "helper for Integer"} {
		t.Run(kind, func(t *testing.T) {
			source := "type T = " + kind + "\n class var v: Integer;\n class const c = 1;\n class function F: Integer; begin Result := c; end;\n class procedure P; begin v := F; end;\nend;"
			if kind == "record" {
				source = strings.Replace(source, "\n class var", "\n Field: Integer;\n class var", 1)
			}
			assertDiagnostics(t, source, "<test>", nil)
		})
	}
	assertDiagnostics(t, "class function F: Integer; begin Result := 1; end;", "<test>", nil)
}

func TestCompile_EarlierInlineRecordDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"duplicate", "var r: record\n MiXeD, mixed: Integer;\n class bug;\nend;", []string{
			"Syntax Error: There is already a field with name \"MiXeD\" [line: 2, column: 9]",
			"Syntax Error: PROCEDURE or FUNCTION expected [line: 3, column: 8]",
		}},
		{"visibility", "var r: record\n protected\n F: Integer;", []string{
			"Syntax Error: Records do not supported \"protected\" visibility specifier [line: 2, column: 2]",
			"Syntax Error: END expected [line: 3, column: 12]",
		}},
		{"record field visibility", "type T = record\n F: record\n  protected\n  X: Integer;", []string{
			"Syntax Error: Records do not supported \"protected\" visibility specifier [line: 3, column: 3]",
			"Syntax Error: END expected [line: 4, column: 13]",
		}},
		{"record field duplicate", "type T = record\n F: record\n  MiXeD, mixed: Integer;\n  class bug;\n end;\nend;", []string{
			"Syntax Error: There is already a field with name \"MiXeD\" [line: 3, column: 10]",
			"Syntax Error: PROCEDURE or FUNCTION expected [line: 4, column: 9]",
		}},
		{"class field visibility", "type T = class\n F: record\n  protected\n  X: Integer;", []string{
			"Syntax Error: Records do not supported \"protected\" visibility specifier [line: 3, column: 3]",
			"Syntax Error: END expected [line: 4, column: 13]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) { assertDiagnostics(t, tt.source, "<test>", tt.want) })
	}
}

func TestCompile_DeclarationVisibilityOrder(t *testing.T) {
	assertDiagnostics(t, "type T = helper for Integer\n function f: Integer; begin Result := 'bad'; end;\n public\n class bug;\nend;", "<test>", []string{
		"Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 2, column: 39]",
		"Hint: Redundant specifier, visibility is already \"public\" [line: 3, column: 2]",
		"Syntax Error: PROCEDURE or FUNCTION expected [line: 4, column: 8]",
	})
	// The body error must not be a compiler stop (an unknown name is one), or
	// upstream would never read the later declarations.
	assertDiagnostics(t, "type T = helper for Integer\n procedure P; begin PrintLn(1 as String); end;\n public\n class bug;\nend;", "<test>", []string{
		"Syntax Error: Cannot cast \"Integer\" as \"String\" [line: 2, column: 31]",
		"Hint: Redundant specifier, visibility is already \"public\" [line: 3, column: 2]",
		"Syntax Error: PROCEDURE or FUNCTION expected [line: 4, column: 8]",
	})
	assertDiagnostics(t, "type T = record\n public\n F;\n class bug;\nend;", "<test>", []string{
		"Hint: Redundant specifier, visibility is already \"public\" [line: 2, column: 2]",
		"Syntax Error: Colon \":\" expected [line: 3, column: 3]",
		"Syntax Error: PROCEDURE or FUNCTION expected [line: 4, column: 8]",
	})
}
