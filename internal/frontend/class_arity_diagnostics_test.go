package frontend

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_ClassArityDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"explicit too few", "type T = class procedure Take(v: Integer; w: String); begin end; end;\nvar obj: T;\nobj.Take(1);", []string{
			"Syntax Error: More arguments expected [line: 3, column: 5]",
		}},
		{"explicit parameterless", "type T = class procedure Take; begin end; end;\nvar obj: T;\nobj.Take(1);", []string{
			"Syntax Error: Too many arguments [line: 3, column: 5]",
		}},
		{"explicit type before count", "type T = class procedure Take(v: Integer; w: String); begin end; end;\nvar obj: T;\nobj.Take(\n  true);", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 3]",
		}},
		{"explicit excess type before count", "type T = class procedure Take(v: Integer); begin end; end;\nvar obj: T;\nobj.Take(\n  true, 1);", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 3]",
		}},
		{"explicit child before count", "type T = class procedure Take(v: Integer; w: String); begin end; end;\nvar obj: T;\nobj.Take(\n  Missing);", []string{
			"Syntax Error: Unknown name \"Missing\" [line: 4, column: 3]",
			"Syntax Error: More arguments expected [line: 3, column: 5]",
		}},
		{"explicit excess child", "type T = class procedure Take(v: Integer); begin end; end;\nvar obj: T;\nobj.Take(1,\n  Missing);", []string{
			"Syntax Error: Unknown name \"Missing\" [line: 4, column: 3]",
			"Syntax Error: Too many arguments [line: 3, column: 5]",
		}},
		{"sole marked method", "type T = class procedure Take(v: Integer); overload; begin end; end;\nvar obj: T;\nobj.Take();", []string{
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 5]",
		}},
		{"sole marked wrong type", "type T = class procedure Take(v: Integer); overload; begin end; end;\nvar obj: T;\nobj.Take(true);", []string{
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 5]",
		}},
		{"multiple overload anchor", "type T = class procedure Take(v: Integer); overload; begin end; procedure Take(v: String); overload; begin end; end;\nvar obj: T;\nobj.Take(true);", []string{
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 5]",
		}},
		{"class method", "type T = class class procedure Take(v: Integer); overload; begin end; end;\nT.Take();", []string{
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 2, column: 3]",
		}},
		{"marked instance method through metaclass", "type T = class procedure Take(v: Integer); overload; begin end; end;\nvar cls: class of T;\ncls.Take(1);", []string{
			"Syntax Error: Class method or constructor expected [line: 3, column: 5]",
		}},
		{"implicit too few", "type T = class procedure Take(v: Integer); begin end;\nprocedure Run; begin\n  Take();\nend; end;", []string{
			"Syntax Error: More arguments expected [line: 3, column: 3]",
		}},
		{"implicit optional", "type T = class procedure Take(v: Integer = 7); begin end;\nprocedure Run; begin\n  Take();\nend; end;", nil},
		{"implicit type before count", "type T = class procedure Take(v: Integer; w: String); begin end;\nprocedure Run; begin\n  Take(true);\nend; end;", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 8]",
		}},
		{"implicit marked", "type T = class procedure Take(v: Integer); overload; begin end;\nprocedure Run; begin\n  Take();\nend; end;", []string{
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 3]",
		}},
		{"method recovery result", "type T = class function Take(v: Integer): String; begin Result := ''; end; end;\nvar obj: T;\nvar n: Integer;\nn := obj.Take();", []string{
			"Syntax Error: More arguments expected [line: 4, column: 10]",
			"Syntax Error: Incompatible types: Cannot assign \"String\" to \"Integer\" [line: 4, column: 6]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_NamedInheritedArity(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"too few", "type B = class procedure Take(v: Integer); begin end; end;\ntype C = class(B) procedure Run; begin\n  inherited Take();\nend; end;", []string{
			"Syntax Error: More arguments expected [line: 3, column: 13]",
		}},
		{"parameterless", "type B = class procedure Take; begin end; end;\ntype C = class(B) procedure Run; begin\n  inherited Take(1);\nend; end;", []string{
			"Syntax Error: Too many arguments [line: 3, column: 13]",
		}},
		{"optional", "type B = class procedure Take(v: Integer = 7); begin end; end;\ntype C = class(B) procedure Run; begin\n  inherited Take();\nend; end;", nil},
		{"type before count", "type B = class procedure Take(v: Integer; w: String); begin end; end;\ntype C = class(B) procedure Run; begin\n  inherited Take(\n    true);\nend; end;", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 4, column: 5]",
		}},
		{"marked", "type B = class procedure Take(v: Integer); overload; begin end; end;\ntype C = class(B) procedure Run; begin\n  inherited Take();\nend; end;", []string{
			"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 13]",
		}},
		{"constructor overload", "type B = class constructor Build(v: Integer); begin end; end;\ntype C = class(B) constructor Build; begin\n  inherited Build(true);\nend; end;", []string{
			"Syntax Error: There is no overloaded version of \"Build\" that can be called with these arguments [line: 3, column: 13]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_ConstructorArityDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"dotted count anchor", "type T = class constructor Create(v: Integer); begin end; end;\nvar obj := T.\n  Create();", []string{
			"Syntax Error: More arguments expected [line: 3, column: 3]",
		}},
		{"new parameterless", "new TObject(1);", []string{
			"Syntax Error: Too many arguments [line: 1, column: 5]",
		}},
		{"dotted parameterless", "TObject.Create(1);", []string{
			"Syntax Error: Too many arguments [line: 1, column: 9]",
		}},
		{"new type before count", "type T = class constructor Create(v: Integer; w: String); begin end; end;\nvar obj := new T(\n  true);", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 3]",
		}},
		{"dotted excess type", "type T = class constructor Create(v: Integer); begin end; end;\nvar obj := T.Create(\n  true, 1);", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"Boolean\" [line: 3, column: 3]",
		}},
		{"new marked count", "type T = class constructor Create(v: Integer); overload; begin end; end;\nvar obj := new T(1, 2);", []string{
			"Syntax Error: There is no overloaded version of \"Create\" that can be called with these arguments [line: 2, column: 16]",
		}},
		{"new marked type", "type T = class constructor Create(v: Integer); overload; begin end; end;\nvar obj := new T(true);", []string{
			"Syntax Error: There is no overloaded version of \"Create\" that can be called with these arguments [line: 2, column: 16]",
		}},
		{"custom constructor overload anchor", "type T = class constructor Build(v: Integer); overload; begin end; constructor Build(v: String); overload; begin end; end;\nvar cls: class of T := T;\nvar obj := cls.Build(true);", []string{
			"Syntax Error: There is no overloaded version of \"Build\" that can be called with these arguments [line: 3, column: 16]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_ClassOverloadsReadAllChildren(t *testing.T) {
	for _, tt := range []struct{ name, source string }{
		{"explicit", "type T = class procedure Take(v: Integer); overload; begin end; end;\nvar obj: T;\nobj.Take(\n  Missing1,\n  Missing2);"},
		{"implicit", "type T = class procedure Take(v: Integer); overload; begin end;\nprocedure Run; begin\n  Take(\n  Missing1,\n  Missing2);\nend; end;"},
		{"inherited", "type B = class procedure Take(v: Integer); overload; begin end; end;\ntype C = class(B) procedure Run; begin\n  inherited Take(\n  Missing1,\n  Missing2);\nend; end;"},
		{"new", "type T = class constructor Create(v: Integer); overload; begin end; end;\nvar obj: T;\nobj := new T(\n  Missing1,\n  Missing2);"},
		{"dotted constructor", "type T = class constructor Build(v: Integer); overload; begin end; end;\nvar obj: T;\nobj := T.Build(\n  Missing1,\n  Missing2);"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			want := []string{
				"Syntax Error: Unknown name \"Missing1\" [line: 4, column: 3]",
				"Syntax Error: Unknown name \"Missing2\" [line: 5, column: 3]",
			}
			if !slices.Equal(got, want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestCompile_ClassOverloadChildBeforeFailure(t *testing.T) {
	for _, tt := range []struct{ name, source string }{
		{"explicit", "type T = class procedure Take(v: Integer); overload; begin end; end;\nvar obj: T;\nobj.Take(\n  1 as String, 2);"},
		{"implicit", "type T = class procedure Take(v: Integer); overload; begin end;\nprocedure Run; begin\n  Take(\n  1 as String, 2);\nend; end;"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			want := []string{
				"Syntax Error: Cannot cast \"Integer\" as \"String\" [line: 4, column: 5]",
				"Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 3]",
			}
			if tt.name == "explicit" {
				want[1] = "Syntax Error: There is no overloaded version of \"Take\" that can be called with these arguments [line: 3, column: 5]"
			}
			if !slices.Equal(got, want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, want)
			}
		})
	}
}
