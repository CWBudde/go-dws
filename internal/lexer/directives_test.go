// Package lexer provides lexical analysis for DWScript source code.
// This file contains tests for compiler directive support.
package lexer

import (
	"testing"
)

// TestCompilerDirectiveDefine tests {$DEFINE} directives.
func TestCompilerDirectiveDefine(t *testing.T) {
	input := `{$DEFINE DEBUG}
	x := 5;`

	tests := []struct {
		expectedType TokenType
	}{
		{IDENT}, // x
		{ASSIGN},
		{INT},
		{SEMICOLON},
		{EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

// TestCompilerDirectiveIfDef tests {$IFDEF} and {$IFNDEF} directives.
func TestCompilerDirectiveIfDef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name: "IFDEF defined symbol",
			input: `{$DEFINE DEBUG}
{$IFDEF DEBUG}
x := 1;
{$ENDIF}`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
		{
			name: "IFDEF undefined symbol",
			input: `{$IFDEF NOTDEFINED}
x := 1;
{$ENDIF}
y := 2;`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
		{
			name: "IFNDEF undefined symbol",
			input: `{$IFNDEF NOTDEFINED}
x := 1;
{$ENDIF}`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
		{
			name: "IFNDEF defined symbol",
			input: `{$DEFINE DEBUG}
{$IFNDEF DEBUG}
x := 1;
{$ENDIF}
y := 2;`,
			expected: []TokenType{IDENT, ASSIGN, INT, SEMICOLON, EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			for i, expectedType := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expectedType {
					t.Errorf("token[%d] - wrong type. expected=%q, got=%q (literal=%q)",
						i, expectedType, tok.Type, tok.Literal)
				}
			}
		})
	}
}

// TestCompilerDirectiveElse tests {$ELSE} directives.
func TestCompilerDirectiveElse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "IFDEF with ELSE - defined",
			input: `{$DEFINE DEBUG}
{$IFDEF DEBUG}
x := 1;
{$ELSE}
y := 2;
{$ENDIF}`,
			expected: []string{"x"},
		},
		{
			name: "IFDEF with ELSE - undefined",
			input: `{$IFDEF NOTDEFINED}
x := 1;
{$ELSE}
y := 2;
{$ENDIF}`,
			expected: []string{"y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			var identifiers []string
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == IDENT {
					identifiers = append(identifiers, tok.Literal)
				}
			}

			if len(identifiers) != len(tt.expected) {
				t.Fatalf("wrong number of identifiers. expected=%d, got=%d",
					len(tt.expected), len(identifiers))
			}

			for i, expected := range tt.expected {
				if identifiers[i] != expected {
					t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
						i, expected, identifiers[i])
				}
			}
		})
	}
}

// TestCompilerDirectiveNested tests nested {$IFDEF} blocks.
func TestCompilerDirectiveNested(t *testing.T) {
	input := `{$DEFINE A}
{$DEFINE B}
{$IFDEF A}
x := 1;
{$IFDEF B}
y := 2;
{$ENDIF}
z := 3;
{$ENDIF}`

	expected := []string{"x", "y", "z"}

	l := New(input)

	var identifiers []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == IDENT {
			identifiers = append(identifiers, tok.Literal)
		}
	}

	if len(identifiers) != len(expected) {
		t.Fatalf("wrong number of identifiers. expected=%d, got=%d",
			len(expected), len(identifiers))
	}

	for i, exp := range expected {
		if identifiers[i] != exp {
			t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
				i, exp, identifiers[i])
		}
	}
}

// TestCompilerDirectiveUndef tests {$UNDEF} directives.
func TestCompilerDirectiveUndef(t *testing.T) {
	input := `{$DEFINE DEBUG}
{$IFDEF DEBUG}
x := 1;
{$ENDIF}
{$UNDEF DEBUG}
{$IFDEF DEBUG}
y := 2;
{$ENDIF}
z := 3;`

	expected := []string{"x", "z"}

	l := New(input)

	var identifiers []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == IDENT {
			identifiers = append(identifiers, tok.Literal)
		}
	}

	if len(identifiers) != len(expected) {
		t.Fatalf("wrong number of identifiers. expected=%d, got=%d",
			len(expected), len(identifiers))
	}

	for i, exp := range expected {
		if identifiers[i] != exp {
			t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
				i, exp, identifiers[i])
		}
	}
}

// TestCompilerDirectiveIf tests {$IF} expression directives.
func TestCompilerDirectiveIf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "$IF with defined()",
			input: `{$DEFINE DEBUG}
{$IF defined(DEBUG)}
x := 1;
{$ENDIF}`,
			expected: []string{"x"},
		},
		{
			name: "$IF with not defined()",
			input: `{$IF not defined(NOTDEFINED)}
x := 1;
{$ENDIF}`,
			expected: []string{"x"},
		},
		{
			name: "$IF with boolean operators",
			input: `{$DEFINE A}
{$DEFINE B}
{$IF defined(A) and defined(B)}
x := 1;
{$ENDIF}`,
			expected: []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			var identifiers []string
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == IDENT {
					identifiers = append(identifiers, tok.Literal)
				}
			}

			if len(identifiers) != len(tt.expected) {
				t.Fatalf("wrong number of identifiers. expected=%v, got=%v",
					tt.expected, identifiers)
			}

			for i, exp := range tt.expected {
				if identifiers[i] != exp {
					t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
						i, exp, identifiers[i])
				}
			}
		})
	}
}

// TestCompilerDirectiveErrors tests error handling in directives.
func TestCompilerDirectiveErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "unterminated directive",
			input:         "{$DEFINE DEBUG",
			expectedError: `"}" expected`,
		},
		{
			name:          "empty directive",
			input:         "{$}",
			expectedError: "empty compiler directive",
		},
		{
			name:          "DEFINE without name",
			input:         "{$DEFINE}\nx := 1;",
			expectedError: "name expected after $define",
		},
		{
			name:          "UNDEF without name",
			input:         "{$UNDEF}\nx := 1;",
			expectedError: "name expected after $undef",
		},
		{
			name:          "IFDEF without name",
			input:         "{$IFDEF}\nx := 1;",
			expectedError: "name expected after $ifdef",
		},
		{
			name:          "unbalanced ELSE",
			input:         "{$ELSE}\nx := 1;",
			expectedError: "Unbalanced conditional directive",
		},
		{
			name:          "unbalanced ENDIF",
			input:         "{$ENDIF}\nx := 1;",
			expectedError: "Unbalanced conditional directive",
		},
		{
			name:          "double ELSE",
			input:         "{$IFDEF DEBUG}\nx := 1;\n{$ELSE}\ny := 2;\n{$ELSE}\nz := 3;\n{$ENDIF}",
			expectedError: "Unfinished conditional directive",
		},
		{
			name:          "unknown directive",
			input:         "{$UNKNOWN}\nx := 1;",
			expectedError: `Compiler switch "UNKNOWN" unknown`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) == 0 {
				t.Fatalf("expected error containing %q, but got no errors", tt.expectedError)
			}

			found := false
			for _, err := range errors {
				if contains(err.Message, tt.expectedError) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error containing %q, but got errors: %v",
					tt.expectedError, errors)
			}
		})
	}
}

// TestCompilerDirectiveCaseInsensitive tests case insensitivity of directives.
func TestCompilerDirectiveCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lowercase define",
			input: "{$define DEBUG}",
		},
		{
			name:  "uppercase DEFINE",
			input: "{$DEFINE DEBUG}",
		},
		{
			name:  "mixed Define",
			input: "{$DeFiNe DEBUG}",
		},
		{
			name:  "lowercase ifdef",
			input: "{$define DEBUG}\n{$ifdef DEBUG}\nx := 1;\n{$endif}",
		},
		{
			name:  "uppercase IFDEF",
			input: "{$DEFINE DEBUG}\n{$IFDEF DEBUG}\nx := 1;\n{$ENDIF}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			// Consume all tokens - should not produce errors
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
			}

			errors := l.Errors()
			if len(errors) > 0 {
				t.Errorf("expected no errors, but got: %v", errors)
			}
		})
	}
}

// TestCompilerDirectiveConstTracking tests constant value tracking for $IF expressions.
// Note: Const tracking works by monitoring the token stream, so consts must appear
// BEFORE directives, and only simple "const NAME = INTEGER" patterns are tracked.
func TestCompilerDirectiveConstTracking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "const before directive - not yet implemented",
			input: `const VERSION = 5;
x := 1;`,
			expected: []string{"VERSION", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			var identifiers []string
			for {
				tok := l.NextToken()
				if tok.Type == EOF {
					break
				}
				if tok.Type == IDENT {
					identifiers = append(identifiers, tok.Literal)
				}
			}

			if len(identifiers) != len(tt.expected) {
				t.Fatalf("wrong number of identifiers. expected=%v, got=%v",
					tt.expected, identifiers)
			}

			for i, exp := range tt.expected {
				if identifiers[i] != exp {
					t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
						i, exp, identifiers[i])
				}
			}
		})
	}
}

// TestCompilerDirectiveMultiline tests multiline directives.
func TestCompilerDirectiveMultiline(t *testing.T) {
	input := `{$DEFINE
	DEBUG}
x := 1;`

	l := New(input)

	tok := l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("expected identifier 'x', got type=%q literal=%q", tok.Type, tok.Literal)
	}
}

// TestCompilerDirectiveComplex tests complex directive combinations.
func TestCompilerDirectiveComplex(t *testing.T) {
	input := `{$DEFINE PLATFORM_WINDOWS}
{$DEFINE VERSION_5}

{$IF defined(PLATFORM_WINDOWS) and defined(VERSION_5)}
x := 1;
{$ENDIF}

{$IFDEF PLATFORM_WINDOWS}
	y := 2;
{$ELSE}
	z := 3;
{$ENDIF}`

	expected := []string{"x", "y"}

	l := New(input)

	var identifiers []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == IDENT {
			identifiers = append(identifiers, tok.Literal)
		}
	}

	if len(identifiers) != len(expected) {
		t.Fatalf("wrong number of identifiers. expected=%v, got=%v",
			expected, identifiers)
	}

	for i, exp := range expected {
		if identifiers[i] != exp {
			t.Errorf("identifier[%d] wrong. expected=%q, got=%q",
				i, exp, identifiers[i])
		}
	}
}

// Helper function to check if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (hasPrefix(s, substr) || hasSuffix(s, substr) || containsMiddle(s, substr)))
}

func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// drainTokens runs the lexer to completion and returns the identifier literals it
// produced, which is enough to tell which conditional branches were taken.
func drainTokens(l *Lexer) []string {
	var idents []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			return idents
		}
		if tok.Type == IDENT || tok.Type == STRING {
			idents = append(idents, tok.Literal)
		}
	}
}

// branchMarkers keeps only the "yes"/"no" branch markers from a token literal list.
func branchMarkers(idents []string) []string {
	var markers []string
	for _, id := range idents {
		if id == "yes" || id == "no" {
			markers = append(markers, id)
		}
	}
	return markers
}

// declaredMarker wraps body in a {$IF Declared(name)} whose active branch emits the
// identifier "yes" and whose inactive branch emits "no".
func declaredMarker(decls, name string) string {
	return decls + "\n{$IF Declared('" + name + "')}\nyes;\n{$ELSE}\nno;\n{$ENDIF}\n"
}

// TestCompilerDirectiveDeclared covers the heuristic declaration tracker behind
// {$IF Declared('X')}: types, dotted members, helpers, case-insensitivity and the
// forward-visibility rule.
func TestCompilerDirectiveDeclared(t *testing.T) {
	const recordDecl = "type TMyRecord = record\n  Dummy : Integer;\nend;"
	const classDecl = "type TFoo = class\n  procedure Bar;\n  FField : Integer;\nend;"
	const helperDecl = "type THelper = helper for TObject\n" +
		"  procedure Proc;\n  begin\n    PrintLn('x');\n  end;\nend;"

	tests := []struct {
		name  string
		decls string
		query string
		want  bool
	}{
		{name: "record type present", decls: recordDecl, query: "TMyRecord", want: true},
		{name: "record type absent", decls: recordDecl, query: "TNoSuchRecord", want: false},
		{name: "record member dotted", decls: recordDecl, query: "TMyRecord.Dummy", want: true},
		{name: "record member unknown", decls: recordDecl, query: "TMyRecord.Oops", want: false},
		{name: "bare member invisible", decls: recordDecl, query: "dummy", want: false},
		{name: "type name case insensitive", decls: recordDecl, query: "tmyRECORD", want: true},
		{name: "member case insensitive", decls: recordDecl, query: "TMYRECORD.DUMMY", want: true},

		{name: "class type", decls: classDecl, query: "TFoo", want: true},
		{name: "class method", decls: classDecl, query: "TFoo.Bar", want: true},
		{name: "class field", decls: classDecl, query: "TFoo.FField", want: true},
		{name: "class bare method invisible", decls: classDecl, query: "Bar", want: false},

		{name: "helper type", decls: helperDecl, query: "THelper", want: true},
		{name: "helper method", decls: helperDecl, query: "THelper.Proc", want: true},
		{name: "helper method through helped type", decls: helperDecl, query: "TObject.Proc", want: true},
		{name: "helper method unknown", decls: helperDecl, query: "THelper.ProcBug", want: false},
		{name: "helped type unknown member", decls: helperDecl, query: "TObject.ProcBug", want: false},
		{name: "helper body local invisible", decls: helperDecl, query: "PrintLn", want: false},

		{name: "seeded TObject", decls: "", query: "TObject", want: true},
		{name: "seeded TObject.Create", decls: "", query: "tobject.CREATE", want: true},
		{name: "seeded TObject.Free", decls: "", query: "TObject.Free", want: true},

		{name: "top level var", decls: "var Alpha : Integer;", query: "Alpha", want: true},
		{name: "top level var list", decls: "var Alpha, Beta : Integer;", query: "Beta", want: true},
		{name: "top level const", decls: "const Gamma = 'g';", query: "Gamma", want: true},
		{name: "top level function", decls: "function Delta : Integer;\nbegin\n  Result := 1;\nend;",
			query: "Delta", want: true},
		{name: "top level procedure", decls: "procedure Epsilon;\nbegin\nend;", query: "Epsilon", want: true},
		{name: "routine local invisible",
			decls: "procedure Epsilon;\nvar Local : Integer;\nbegin\nend;", query: "Local", want: false},

		{name: "forward declaration not visible", decls: "", query: "Later", want: false},

		// A declaration section stays in effect across semicolons, so every entry is
		// recorded — not just the first.
		{name: "var section second entry", decls: "var Alpha : Integer; Beta : Integer;",
			query: "Beta", want: true},
		{name: "var section third entry",
			decls: "var Alpha : Integer; Beta : Integer; Gamma : String;", query: "Gamma", want: true},
		{name: "var section second entry list",
			decls: "var Alpha : Integer; Beta, Delta : Integer;", query: "Delta", want: true},
		{name: "const section second entry", decls: "const Gamma = 'g'; Theta = 't';",
			query: "Theta", want: true},
		// ... but a statement after a section must not be mistaken for a declaration.
		{name: "statement after var section not tracked",
			decls: "var Alpha : Integer;\nSomeCall('x');", query: "SomeCall", want: false},
		{name: "assignment after var section not tracked",
			decls: "var Alpha : Integer;\nUnknownVar := 1;", query: "UnknownVar", want: false},
		// A visibility section introduces members; it must not interrupt them.
		{name: "field after private",
			decls: "type TWidget = class\nprivate\n  FPriv : Integer;\nend;",
			query: "TWidget.FPriv", want: true},
		{name: "field after public",
			decls: "type TWidget = class\nprivate\n  FPriv : Integer;\npublic\n  FPub : Integer;\nend;",
			query: "TWidget.FPub", want: true},
		{name: "field after strict private",
			decls: "type TWidget = class\nstrict private\n  FHidden : Integer;\nend;",
			query: "TWidget.FHidden", want: true},
		{name: "method after protected",
			decls: "type TWidget = class\nprotected\n  procedure Run;\nend;",
			query: "TWidget.Run", want: true},
		{name: "visibility keyword not a member",
			decls: "type TWidget = class\nprivate\n  FPriv : Integer;\nend;",
			query: "TWidget.private", want: false},

		{name: "var section ends at begin",
			decls: "var Alpha : Integer;\nbegin\n  Local := 1;\nend;", query: "Local", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := declaredMarker(tt.decls, tt.query)
			// The forward case declares the name only after the {$IF}.
			if tt.name == "forward declaration not visible" {
				src += "var Later : Integer;\n"
			}
			l := New(src)
			markers := branchMarkers(drainTokens(l))

			want := "no"
			if tt.want {
				want = "yes"
			}
			if len(markers) != 1 || markers[0] != want {
				t.Fatalf("Declared(%q): got markers %v, want [%s]", tt.query, markers, want)
			}
		})
	}
}

// TestCompilerDirectiveDeclaredAfterDeclaration verifies the point-of-use rule: the same
// query is false before the declaration and true after it.
func TestCompilerDirectiveDeclaredAfterDeclaration(t *testing.T) {
	src := "{$IF Declared('Later')}early;{$ENDIF}\n" +
		"var Later : Integer;\n" +
		"{$IF Declared('Later')}late;{$ENDIF}\n"

	var markers []string
	for _, id := range drainTokens(New(src)) {
		if id == "early" || id == "late" {
			markers = append(markers, id)
		}
	}
	if len(markers) != 1 || markers[0] != "late" {
		t.Fatalf("got markers %v, want [late]", markers)
	}
}

// TestCompilerDirectiveDefinedIsPreprocessorOnly verifies that Defined() answers only the
// {$DEFINE} question while Declared() answers the symbol-table one.
func TestCompilerDirectiveDefinedIsPreprocessorOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Defined does not see declarations",
			input: "var Alpha : Integer;\n{$IF Defined('Alpha')}yes;{$ELSE}no;{$ENDIF}",
			want:  "no",
		},
		{
			name:  "Defined sees DEFINE symbols",
			input: "{$DEFINE Alpha}\n{$IF Defined(Alpha)}yes;{$ELSE}no;{$ENDIF}",
			want:  "yes",
		},
		{
			name:  "Declared does not see DEFINE symbols",
			input: "{$DEFINE Alpha}\n{$IF Declared('Alpha')}yes;{$ELSE}no;{$ENDIF}",
			want:  "no",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markers := branchMarkers(drainTokens(New(tt.input)))
			if len(markers) != 1 || markers[0] != tt.want {
				t.Fatalf("got markers %v, want [%s]", markers, tt.want)
			}
		})
	}
}

// TestCompilerDirectiveIfArgumentDiagnostics pins the DWScript messages and columns for a
// malformed Defined()/Declared() argument, and verifies that a directive in an inactive
// branch reports nothing.
func TestCompilerDirectiveIfArgumentDiagnostics(t *testing.T) {
	type diag struct {
		message string
		line    int
		column  int
	}

	tests := []struct {
		name  string
		input string
		want  []diag
	}{
		{
			name:  "Defined with integer argument",
			input: "var i : Integer;\n\n{$if Defined(123)}{$endif}\n",
			want:  []diag{{"String expected", 3, 14}},
		},
		{
			name:  "Declared with call argument",
			input: "var i : Integer;\n\n{$if Declared(IntToStr(i))}{$endif}\n",
			want:  []diag{{"Constant expression expected", 3, 15}},
		},
		{
			name: "both, matching FailureScripts/special_funcs5",
			input: "var i : Integer;\n\n{$if Defined(123)}{$endif}\n\n" +
				"{$if Declared(IntToStr(i))}{$endif}\n",
			want: []diag{
				{"String expected", 3, 14},
				{"Constant expression expected", 5, 15},
			},
		},
		{
			name:  "inactive branch reports nothing",
			input: "{$IFDEF NOT_DEFINED}\n{$if Defined(123)}{$endif}\n{$ENDIF}\n",
			want:  nil,
		},
		{
			name:  "well formed arguments report nothing",
			input: "var i : Integer;\n{$if Declared('i')}{$endif}\n{$if Defined(DWSCRIPT)}{$endif}\n",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			drainTokens(l)

			got := l.DirectiveDiagnostics()
			if len(got) != len(tt.want) {
				t.Fatalf("got %d diagnostics %v, want %d", len(got), got, len(tt.want))
			}
			for i, want := range tt.want {
				if got[i].Message != want.message ||
					got[i].Pos.Line != want.line ||
					got[i].Pos.Column != want.column {
					t.Errorf("diagnostic %d = %q [line: %d, column: %d], want %q [line: %d, column: %d]",
						i, got[i].Message, got[i].Pos.Line, got[i].Pos.Column,
						want.message, want.line, want.column)
				}
			}
		})
	}
}

// TestDirectiveStringArgumentDecoding covers DWScript string escaping in a directive
// argument: an inner quote is written by doubling it.
func TestDirectiveStringArgumentDecoding(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
		ok   bool
	}{
		{name: "plain single quoted", arg: "'done'", want: "done", ok: true},
		{name: "plain double quoted", arg: `"done"`, want: "done", ok: true},
		{name: "doubled single quote", arg: "'it''s ready'", want: "it's ready", ok: true},
		{name: "doubled double quote", arg: `"say ""hi"""`, want: `say "hi"`, ok: true},
		{name: "double quote inside single quoted", arg: `'say "hi"'`, want: `say "hi"`, ok: true},
		{name: "empty string", arg: "''", want: "", ok: true},
		{name: "unterminated", arg: "'oops", ok: false},
		{name: "unquoted", arg: "oops", ok: false},
		{name: "undoubled inner quote", arg: "'a'b'", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := unquoteDirectiveString(tt.arg)
			if ok != tt.ok {
				t.Fatalf("unquoteDirectiveString(%q) ok = %v, want %v", tt.arg, ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("unquoteDirectiveString(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

// TestDirectiveMessageDecodesDoubledQuote checks the decoded text reaches the diagnostic.
func TestDirectiveMessageDecodesDoubledQuote(t *testing.T) {
	l := New("{$HINT 'it''s ready'}\n")
	drainTokens(l)

	diags := l.DirectiveDiagnostics()
	if len(diags) != 1 {
		t.Fatalf("got %d diagnostics, want 1: %v", len(diags), diags)
	}
	if diags[0].Message != "it's ready" {
		t.Errorf("message = %q, want %q", diags[0].Message, "it's ready")
	}
	if want := "Hint: it's ready [line: 1, column: 3]"; diags[0].Rendered != want {
		t.Errorf("rendered = %q, want %q", diags[0].Rendered, want)
	}
}

// TestFatalStopIsRewoundByRestoreState guards the {$FATAL} stop flag against parser
// backtracking: a speculative read that runs past the directive must not leave the lexer
// permanently stopped.
func TestFatalStopIsRewoundByRestoreState(t *testing.T) {
	l := New("alpha beta {$FATAL 'stop'} gamma")

	first := l.NextToken()
	if first.Literal != "alpha" {
		t.Fatalf("first token = %q, want alpha", first.Literal)
	}

	state := l.SaveState()

	// Speculative read that runs into the fatal directive.
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
	}
	if !l.StoppedByFatal() {
		t.Fatal("expected the speculative read to hit {$FATAL}")
	}

	l.RestoreState(state)
	if l.StoppedByFatal() {
		t.Fatal("RestoreState did not rewind the fatal stop flag")
	}
	if tok := l.NextToken(); tok.Literal != "beta" {
		t.Fatalf("after restore got %q, want beta", tok.Literal)
	}
}

// TestHintsAndWarningsSwitchesGateMessages covers {$HINTS}/{$WARNINGS}: they switch
// whether later {$HINT}/{$WARNING} directives report anything. {$ERROR} is not switchable.
func TestHintsAndWarningsSwitchesGateMessages(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "hints off suppresses hint",
			src:  "{$HINT 'first'}{$HINTS OFF}{$HINT 'gone'}",
			want: []string{"first"},
		},
		{
			name: "hints on restores hint",
			src:  "{$HINTS OFF}{$HINT 'gone'}{$HINTS ON}{$HINT 'back'}",
			want: []string{"back"},
		},
		{
			name: "warnings off suppresses warning",
			src:  "{$WARNINGS OFF}{$WARNING 'gone'}{$WARNINGS ON}{$WARNING 'back'}",
			want: []string{"back"},
		},
		{
			name: "switches are independent",
			src:  "{$WARNINGS OFF}{$HINT 'hint survives'}{$WARNING 'gone'}",
			want: []string{"hint survives"},
		},
		{
			name: "hint level names mean on",
			src:  "{$HINTS OFF}{$HINTS PEDANTIC}{$HINT 'pedantic on'}",
			want: []string{"pedantic on"},
		},
		{
			name: "errors are not switchable",
			src:  "{$HINTS OFF}{$WARNINGS OFF}{$ERROR 'still reported'}",
			want: []string{"still reported"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.src)
			drainTokens(l)

			var got []string
			for _, d := range l.DirectiveDiagnostics() {
				got = append(got, d.Message)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("message %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestHintsSwitchIsRewoundByRestoreState keeps the switches consistent under parser
// backtracking, like the other directive side effects.
func TestHintsSwitchIsRewoundByRestoreState(t *testing.T) {
	l := New("alpha {$HINTS OFF} beta")

	if tok := l.NextToken(); tok.Literal != "alpha" {
		t.Fatalf("first token = %q, want alpha", tok.Literal)
	}
	state := l.SaveState()

	drainTokens(l)
	if l.hintsEnabled {
		t.Fatal("expected the speculative read to switch hints off")
	}

	l.RestoreState(state)
	if !l.hintsEnabled {
		t.Fatal("RestoreState did not rewind the {$HINTS} switch")
	}
}
