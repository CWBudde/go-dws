// Package lexer provides lexical analysis for DWScript source code.
// This file contains tests for {$INCLUDE} / {$I} / {$INCLUDE_ONCE} directives.
package lexer

import (
	"fmt"
	"testing"
)

// mapIncludeResolver builds an IncludeResolver backed by an in-memory file map.
// The canonical path is the file name itself, so {$INCLUDE_ONCE} de-duplication
// keys on the exact name.
func mapIncludeResolver(files map[string]string) IncludeResolver {
	return func(name, _ string) (string, string, error) {
		content, ok := files[name]
		if !ok {
			return "", "", fmt.Errorf("file not found: %s", name)
		}
		return content, name, nil
	}
}

// collectLiterals drains the lexer and returns the literals of every non-EOF token.
func collectLiterals(l *Lexer) []string {
	var out []string
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			return out
		}
		out = append(out, tok.Literal)
	}
}

func TestIncludeDirectiveSplicesContent(t *testing.T) {
	files := map[string]string{
		"greeting.inc": `PrintLn('hello');`,
	}
	input := `a; {$INCLUDE 'greeting.inc'} b;`

	l := New(input, WithIncludeResolver(mapIncludeResolver(files)))
	got := collectLiterals(l)

	// a ; PrintLn ( 'hello' ) ; b ;
	want := []string{"a", ";", "PrintLn", "(", "hello", ")", ";", "b", ";"}
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d (got %#v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if errs := l.Errors(); len(errs) != 0 {
		t.Fatalf("unexpected lexer errors: %v", errs)
	}
}

func TestIncludeDirectiveShortForm(t *testing.T) {
	files := map[string]string{"x.inc": `42`}
	l := New(`{$I 'x.inc'}`, WithIncludeResolver(mapIncludeResolver(files)))
	got := collectLiterals(l)
	if len(got) != 1 || got[0] != "42" {
		t.Fatalf("got %#v, want [42]", got)
	}
}

func TestIncludeDirectiveNested(t *testing.T) {
	files := map[string]string{
		"outer.inc": `1 {$INCLUDE 'inner.inc'} 3`,
		"inner.inc": `2`,
	}
	l := New(`{$INCLUDE 'outer.inc'}`, WithIncludeResolver(mapIncludeResolver(files)))
	got := collectLiterals(l)
	want := []string{"1", "2", "3"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestIncludeNestedResolvesFromCurrentFile verifies that a relative include nested
// inside an already-included file is resolved against that file's directory, not the
// top-level entry point (matching DWScript semantics).
func TestIncludeNestedResolvesFromCurrentFile(t *testing.T) {
	// Files keyed by canonical path; the resolver joins relative names against the
	// directory of the including file (fromPath).
	files := map[string]string{
		"/root/main.dws":  "",
		"/root/sub/a.inc": `1 {$INCLUDE 'b.inc'} 3`,
		"/root/sub/b.inc": `2`,
	}
	var fromPaths []string
	resolver := func(name, fromPath string) (string, string, error) {
		fromPaths = append(fromPaths, fromPath)
		dir := "/root"
		if fromPath != "" {
			dir = pathDir(fromPath)
		}
		canonical := dir + "/" + name
		content, ok := files[canonical]
		if !ok {
			return "", "", fmt.Errorf("not found: %s", canonical)
		}
		return content, canonical, nil
	}

	// Entry point is /root/main.dws; its directive uses the sub/ subdirectory.
	l := New(`{$INCLUDE 'sub/a.inc'}`, WithIncludeResolver(resolver))
	l.currentIncludePath = "/root/main.dws"
	got := collectLiterals(l)

	want := []string{"1", "2", "3"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// The nested {$INCLUDE 'b.inc'} must have been resolved from /root/sub/a.inc.
	if len(fromPaths) != 2 || fromPaths[1] != "/root/sub/a.inc" {
		t.Fatalf("nested include fromPath = %v, want second entry /root/sub/a.inc", fromPaths)
	}
}

// pathDir returns everything up to the last '/' in p (a test-local dirname).
func pathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

func TestIncludeOnceDeduplicates(t *testing.T) {
	files := map[string]string{"once.inc": `7`}
	input := `{$INCLUDE_ONCE 'once.inc'} {$INCLUDE_ONCE 'once.inc'}`

	l := New(input, WithIncludeResolver(mapIncludeResolver(files)))
	got := collectLiterals(l)
	if len(got) != 1 || got[0] != "7" {
		t.Fatalf("got %#v, want [7] (second include_once should be skipped)", got)
	}
}

func TestIncludeOnceGuardsMutualRecursion(t *testing.T) {
	files := map[string]string{
		"a.inc": `{$INCLUDE_ONCE 'b.inc'} 1`,
		"b.inc": `{$INCLUDE_ONCE 'a.inc'} 2`,
	}
	l := New(`{$INCLUDE_ONCE 'a.inc'}`, WithIncludeResolver(mapIncludeResolver(files)))
	got := collectLiterals(l)
	// a enters -> a includes b -> b tries to re-include a (skipped) -> b yields 2
	// -> back in a, yields 1.
	want := []string{"2", "1"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPlainIncludeRepeats(t *testing.T) {
	files := map[string]string{"r.inc": `9`}
	input := `{$INCLUDE 'r.inc'} {$INCLUDE 'r.inc'}`
	l := New(input, WithIncludeResolver(mapIncludeResolver(files)))
	got := collectLiterals(l)
	want := []string{"9", "9"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestIncludeMissingFileRecordsError(t *testing.T) {
	l := New(`{$INCLUDE 'nope.inc'}`, WithIncludeResolver(mapIncludeResolver(nil)))
	_ = collectLiterals(l)
	if len(l.Errors()) == 0 {
		t.Fatal("expected a lexer error for a missing include file")
	}
}

func TestIncludeWithoutResolverIsIgnored(t *testing.T) {
	// No resolver configured: the directive is a no-op, not an error.
	l := New(`a {$INCLUDE 'x.inc'} b`)
	got := collectLiterals(l)
	want := []string{"a", "b"}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if len(l.Errors()) != 0 {
		t.Fatalf("unexpected errors: %v", l.Errors())
	}
}

// TestIncludeSurvivesSaveRestoreState exercises the parser-backtracking path: a
// SaveState taken before an {$INCLUDE} boundary must rewind the lexer to the right
// file and position when RestoreState is called.
func TestIncludeSurvivesSaveRestoreState(t *testing.T) {
	files := map[string]string{"inc.inc": `x y`}
	l := New(`a {$INCLUDE 'inc.inc'} b`, WithIncludeResolver(mapIncludeResolver(files)))

	// Consume 'a', then snapshot before the include is entered.
	if tok := l.NextToken(); tok.Literal != "a" {
		t.Fatalf("first token = %q, want a", tok.Literal)
	}
	state := l.SaveState()

	first := collectLiterals(l)
	want := []string{"x", "y", "b"}
	if fmt.Sprint(first) != fmt.Sprint(want) {
		t.Fatalf("first pass = %v, want %v", first, want)
	}

	// Rewind and re-lex: the include must expand identically.
	l.RestoreState(state)
	second := collectLiterals(l)
	if fmt.Sprint(second) != fmt.Sprint(want) {
		t.Fatalf("second pass after restore = %v, want %v", second, want)
	}
}

func TestIncludeValueSubstitutionIsSkipped(t *testing.T) {
	// {$I %FILE%} style value substitutions are not file includes and must not be
	// treated as a missing file.
	l := New(`{$I %FILE%}`, WithIncludeResolver(mapIncludeResolver(nil)))
	_ = collectLiterals(l)
	if len(l.Errors()) != 0 {
		t.Fatalf("value substitution should not error: %v", l.Errors())
	}
}
