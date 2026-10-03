package frontend

import (
	"slices"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// Exercise the production compile path, including diagnostic recovery and spans.
func TestCompile_CallArgumentDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{
			name: "constructor",
			source: "type TTest = class constructor Create(v: Integer); begin end; end;\n" +
				"var o := TTest.Create(\n  'bad');",
			want: []string{`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 3, column: 3]`},
		},
		{
			name: "constructor through metaclass",
			source: "type TTest = class constructor Create(v: Integer); begin end; end;\n" +
				"var cls: class of TTest := TTest;\nvar o := cls.Create(\n  'bad');",
			want: []string{`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 4, column: 3]`},
		},
		{
			name: "implicit class method",
			source: "type TTest = class\n procedure Take(v: Integer); begin end;\n" +
				" procedure Run; begin Take(\n  'bad'); end;\nend;",
			want: []string{`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 4, column: 3]`},
		},
		{
			name: "strict implicit method",
			source: "type TTest = class\n procedure Take(v: type Float); begin end;\n" +
				" procedure Run; begin Take(\n  123); end;\nend;",
			want: []string{`Syntax Error: Argument 0 expects type "Float" instead of "Integer" [line: 4, column: 3]`},
		},
		{
			name: "strict explicit method",
			source: "type TTest = class procedure Take(v: type Float); begin end; end;\n" +
				"var o: TTest;\no.Take(\n  123);",
			want: []string{`Syntax Error: Argument 0 expects type "Float" instead of "Integer" [line: 4, column: 3]`},
		},
		{
			name: "strict explicit class method",
			source: "type TTest = class class procedure Take(v: type Float); begin end; end;\n" +
				"TTest.Take(\n  123);",
			want: []string{`Syntax Error: Argument 0 expects type "Float" instead of "Integer" [line: 3, column: 3]`},
		},
		{
			name: "convertible implicit method",
			source: "type TTest = class\n procedure Take(v: Float); begin end;\n" +
				" procedure Run; begin Take(\n  123); end;\nend;",
		},
		{
			name: "valueless implicit method argument",
			source: "type TTest = class\n procedure Take(v: Integer); begin end;\n" +
				" procedure Run; begin Take(\n  Print('')); end;\nend;",
			want: []string{`Syntax Error: Argument 0 expects type "Integer" [line: 4, column: 3]`},
		},
		{
			name: "poisoned implicit method argument",
			source: "type TTest = class\n procedure Take(v: Integer); begin end;\n" +
				" procedure Run; begin Take(\n  IntToStr(bug)); end;\nend;",
			want: []string{`Syntax Error: Unknown name "bug" [line: 4, column: 12]`},
		},
		{
			name: "strict constructor",
			source: "type TTest = class constructor Create(v: type Float); begin end; end;\n" +
				"var o := TTest.Create(\n  123);",
			want: []string{`Syntax Error: Argument 0 expects type "Float" instead of "Integer" [line: 3, column: 3]`},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_ReceiverCallArgumentDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{
			name: "implicit helper receiver",
			source: "type TTest = class procedure Run; end;\n" +
				"type THelp = helper for TTest procedure Take(v: Integer); begin end; end;\n" +
				"procedure TTest.Run; begin Take(\n  'bad'); end;",
			want: []string{`Syntax Error: Argument 1 expects type "Integer" instead of "String" [line: 3, column: 28]`},
		},
		{
			name: "implicit helper shifted position",
			source: "type TTest = class procedure Run; end;\n" +
				"type THelp = helper for TTest procedure Take(v: Integer; w: String); begin end; end;\n" +
				"procedure TTest.Run; begin Take(\n  'bad',\n  'ok'); end;",
			want: []string{`Syntax Error: Argument 1 expects type "Integer" instead of "String" [line: 5, column: 3]`},
		},
		{
			name: "inherited method",
			source: "type TBase = class procedure Take(v: Integer); begin end; end;\n" +
				"type TChild = class(TBase) procedure Run; begin inherited Take(\n  'bad'); end; end;",
			want: []string{`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 3, column: 3]`},
		},
		{
			// Constructor candidates include an implicit parameterless overload.
			// Keep the existing resolution failure rather than changing selection.
			name: "inherited constructor overload failure",
			source: "type TBase = class constructor Build(v: Integer); begin end; end;\n" +
				"type TChild = class(TBase) constructor Build; begin inherited Build(\n  'bad'); end; end;",
			want: []string{`Syntax Error: There is no overloaded version of "Build" that can be called with these arguments [line: 2, column: 53]`},
		},
		{
			name: "inherited helper receiver",
			source: "type TBase = helper for Integer procedure Take(v: String); begin end; end;\n" +
				"type TChild = helper(TBase) for Integer procedure Run; begin inherited Take(\n  123); end; end;",
			want: []string{`Syntax Error: Argument 1 expects type "String" instead of "Integer" [line: 2, column: 72]`},
		},
		{
			name: "record instance receiver shifted position",
			source: "type TRec = record procedure Take(v: Integer; w: String); begin end; end;\n" +
				"var r: TRec;\nr.Take(\n  'bad',\n  'ok');",
			want: []string{`Syntax Error: Argument 1 expects type "Integer" instead of "String" [line: 5, column: 3]`},
		},
		{
			name:   "set include",
			source: "type TEnum = (one, two); type TSet = set of TEnum; var s: TSet;\ns.Include(\n  'bad');",
			want:   []string{`Syntax Error: Argument 0 expects type "TEnum" instead of "String" [line: 3, column: 3]`},
		},
		{
			name:   "set exclude valueless argument",
			source: "type TEnum = (one, two); type TSet = set of TEnum; var s: TSet;\ns.Exclude(\n  Print(''));",
			want:   []string{`Syntax Error: Argument 0 expects type "TEnum" [line: 3, column: 3]`},
		},
		{
			name:   "set poisoned argument",
			source: "type TEnum = (one, two); type TSet = set of TEnum; var s: TSet;\ns.Include(\n  IntToStr(bug));",
			want:   []string{`Syntax Error: Unknown name "bug" [line: 3, column: 12]`},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestCompile_OverloadCountDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, declaration, call, sentence string
	}{
		{"marked too few", "procedure Take(v: Integer); overload; begin end;", "Take();", `There is no overloaded version of "Take" that can be called with these arguments`},
		{"marked too many", "procedure Take(v: Integer); overload; begin end;", "Take(1, 2);", `There is no overloaded version of "Take" that can be called with these arguments`},
		{"marked parameterless", "procedure Take; overload; begin end;", "Take(1);", `There is no overloaded version of "Take" that can be called with these arguments`},
		{"unmarked too few", "procedure Take(v: Integer); begin end;", "Take();", "More arguments expected"},
		{"unmarked too many", "procedure Take(v: Integer); begin end;", "Take(1, 2);", "Too many arguments"},
		{"unmarked parameterless", "procedure Take; begin end;", "Take(1);", "No arguments expected"},
		{"marked optional", "procedure Take(v: Integer = 1); overload; begin end;", "Take();", ""},
		{"marked valid", "procedure Take(v: Integer); overload; begin end;", "Take(1);", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var want []string
			if tt.sentence != "" {
				want = []string{"Syntax Error: " + tt.sentence + " [line: 2, column: 1]"}
			}
			got := Compile(tt.declaration+"\n"+tt.call, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestCompile_UnmarkedHelperCountDiagnostic(t *testing.T) {
	source := "type H = helper for Integer procedure Take(v: Integer); begin\n  Take(); end; end;"
	assertDiagnostics(t, source, "<test>", []string{
		`Syntax Error: More arguments expected [line: 2, column: 3]`,
	})
}

func TestCompile_HelperOverloadDirectiveRespectsSourceOrder(t *testing.T) {
	// A later marked overload must not change the earlier unmarked call.
	source := "type H = helper for Integer\n" +
		" procedure Take(v: Integer); begin Take(); end;\n" +
		" procedure Take(v, w: Integer); overload; begin end;\nend;"
	assertDiagnostics(t, source, "<test>", []string{
		`Syntax Error: More arguments expected [line: 2, column: 36]`,
	})
}

func TestCompile_HelperVisibleOverloadsRetainUnmarkedFirst(t *testing.T) {
	source := "type H = helper for Integer\n" +
		" procedure Take(v: Integer); begin end;\n" +
		" procedure Take(v, w: Integer); overload; begin end;\n" +
		" procedure Run; begin Take(1); Take(1, 2); end;\nend;"
	got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
	if len(got) != 0 {
		t.Fatalf("visible overloads produced diagnostics: %q", got)
	}
}

func TestCompile_DuplicateClassMethodDiagnostics(t *testing.T) {
	t.Run("empty_body fixture", func(t *testing.T) {
		assertDiagnostics(t, fixtureSource(t, "empty_body.pas"), "empty_body.pas", fixtureExpectation(t, "empty_body.txt"))
	})
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{
			name: "out of line implementation of inline body",
			source: "type TTest = class procedure Take; begin end; end;\n" +
				"procedure TTest.Take;\nbegin end;",
			want: []string{`Syntax Error: There is already a method with name "Take" [line: 2, column: 21]`},
		},
		{
			name: "duplicate static class method",
			source: "type TTest = class class procedure Take; static; begin end; end;\n" +
				"class procedure TTest.Take; static;\nbegin end;",
			want: []string{`Syntax Error: There is already a method with name "Take" [line: 3, column: 1]`},
		},
		{
			name: "forward implementation remains valid",
			source: "type TTest = class procedure Take(v: Integer = 1); end;\n" +
				"procedure TTest.Take(v: Integer); begin end;",
		},
		{
			name: "distinct overloads remain valid",
			source: "type TTest = class\n procedure Take(v: Integer); overload; begin end;\n" +
				" procedure Take(v: String); overload; begin end;\nend;",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if !slices.Equal(got, tt.want) {
				t.Fatalf("diagnostics:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}
