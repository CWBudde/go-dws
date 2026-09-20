package frontend

import (
	"reflect"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestCompile_DiagnosticsKeepSourceEmissionOrder pins the complete diagnostic output
// for inputs whose only failure used to be ordering (PLAN.md §4, F1). Upstream compiles
// in a single pass, so diagnostics come out in the order the compiler reached the code
// that produced them. Two go-dws design choices break that order on the way out and
// are repaired here:
//
//   - top-level routine bodies are analyzed in a deferred second pass (so mutually
//     recursive routines resolve), which reported them after the main program;
//   - a compiler switch the tokenizer rejects is a lexer diagnostic collected before
//     semantic analysis ran, which pinned it ahead of every semantic hint.
//
// The first two cases mirror fixtures; their expectations are the fixtures' `.txt`
// verbatim. The remaining cases pin the splice against class method bodies (drained
// after the last class declaration) and nested routines (analyzed inline).
func TestCompile_DiagnosticsKeepSourceEmissionOrder(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// FailureScripts/infinite_loop: the routine bodies come first, in source
			// order, and inside a body the inner loop is reported before the outer
			// one (the analyzer reaches the end of the inner loop first).
			name: "routine bodies before the main program",
			source: "procedure Trap;\n" +
				"begin\n" +
				"   while True do begin\n" +
				"      repeat\n" +
				"         PrintLn('');\n" +
				"      until False;\n" +
				"   end;\n" +
				"end;\n" +
				"\n" +
				"procedure NoTrap;\n" +
				"begin\n" +
				"   while True do begin\n" +
				"      repeat\n" +
				"         Exit;\n" +
				"      until False;\n" +
				"   end;\n" +
				"end;\n" +
				"\n" +
				"while True do ;\n" +
				"\n" +
				"while True do begin\n" +
				"   PrintLn('');\n" +
				"end;\n" +
				"while False do begin\n" +
				"   PrintLn('');\n" +
				"end;\n" +
				"\n" +
				"while 1=1 do begin\n" +
				"   PrintLn('');\n" +
				"   Break;\n" +
				"end;\n" +
				"\n" +
				"repeat\n" +
				"   PrintLn('');\n" +
				"until False;\n" +
				"repeat\n" +
				"   PrintLn('');\n" +
				"until True;\n" +
				"\n" +
				"repeat\n" +
				"   PrintLn('');\n" +
				"   Break;\n" +
				"until 1=1;\n",
			want: []string{
				`Warning: Infinite loop [line: 6, column: 13]`,
				`Warning: Infinite loop [line: 3, column: 4]`,
				`Warning: Infinite loop [line: 19, column: 1]`,
				`Warning: Infinite loop [line: 21, column: 1]`,
				`Warning: Infinite loop [line: 35, column: 7]`,
			},
		},
		{
			// FailureScripts/hint_pedantic: the tokenizer reports the unknown switch
			// when the parser reaches it, after the hint for the statement above.
			name: "unknown compiler switch after an earlier hint",
			source: "procedure Test; begin end;\n" +
				"\n" +
				"{$HINTS OFF}\n" +
				"\n" +
				"test;\n" +
				"\n" +
				"{$HINTS pedantic}\n" +
				"\n" +
				"test;\n" +
				"\n" +
				"{$FOOBAR}\n",
			want: []string{
				`Hint: "test" does not match case of declaration ("Test") [line: 9, column: 1]`,
				`Syntax Error: Compiler switch "FOOBAR" unknown [line: 11, column: 3]`,
			},
		},
		{
			// An unknown switch before the statement that produces the hint stays first.
			name: "unknown compiler switch before a later hint",
			source: "procedure Test; begin end;\n" +
				"{$FOOBAR}\n" +
				"test;\n",
			want: []string{
				`Syntax Error: Compiler switch "FOOBAR" unknown [line: 2, column: 3]`,
				`Hint: "test" does not match case of declaration ("Test") [line: 3, column: 1]`,
			},
		},
		{
			// The rule is about origin, not severity: a semantic warning emitted
			// before the tokenizer reached the switch is printed before it.
			name: "unknown compiler switch after an earlier warning",
			source: "while True do ;\n" +
				"{$FOOBAR}\n",
			want: []string{
				`Warning: Infinite loop [line: 1, column: 1]`,
				`Syntax Error: Compiler switch "FOOBAR" unknown [line: 2, column: 3]`,
			},
		},
		{
			// A {$HINT} directive is emitted where the tokenizer reached it too, so
			// it interleaves with semantic hints by line in both directions.
			name: "message directive hint interleaved with a semantic hint",
			source: "procedure Test; begin end;\n" +
				"{$HINT 'first'}\n" +
				"test;\n" +
				"{$HINT 'last'}\n",
			want: []string{
				`Hint: first [line: 2, column: 3]`,
				`Hint: "test" does not match case of declaration ("Test") [line: 3, column: 1]`,
				`Hint: last [line: 4, column: 3]`,
			},
		},
		{
			// A class method body is checked once the last class declaration is
			// complete; a routine declared after the class is spliced back after the
			// method's diagnostics and before the main program's. A nested routine is
			// analyzed inline, ahead of the enclosing body's own statements.
			name: "class method body, then routine with nested routine, then main",
			source: "type TFoo = class\n" +
				"   procedure M;\n" +
				"   begin\n" +
				"      while True do ;\n" +
				"   end;\n" +
				"end;\n" +
				"\n" +
				"procedure Outer;\n" +
				"   procedure Inner;\n" +
				"   begin\n" +
				"      while True do ;\n" +
				"   end;\n" +
				"begin\n" +
				"   while True do ;\n" +
				"end;\n" +
				"\n" +
				"while True do ;\n",
			want: []string{
				`Warning: Infinite loop [line: 4, column: 7]`,
				`Warning: Infinite loop [line: 11, column: 7]`,
				`Warning: Infinite loop [line: 14, column: 4]`,
				`Warning: Infinite loop [line: 17, column: 1]`,
			},
		},
		{
			// A routine declared before the class is spliced back ahead of the method
			// body's diagnostics: the splice index is where the body was deferred, not
			// where the method bodies were drained.
			name: "routine before class method body",
			source: "procedure Outer;\n" +
				"begin\n" +
				"   while True do ;\n" +
				"end;\n" +
				"\n" +
				"type TFoo = class\n" +
				"   procedure M;\n" +
				"   begin\n" +
				"      while True do ;\n" +
				"   end;\n" +
				"end;\n" +
				"\n" +
				"while True do ;\n",
			want: []string{
				`Warning: Infinite loop [line: 3, column: 4]`,
				`Warning: Infinite loop [line: 9, column: 7]`,
				`Warning: Infinite loop [line: 13, column: 1]`,
			},
		},
		{
			// Routines interleaved with main-program statements: each body lands at
			// its own declaration position, and later splices shift accordingly.
			name: "routines interleaved with main statements",
			source: "while True do ;\n" +
				"procedure A;\n" +
				"begin\n" +
				"   while True do ;\n" +
				"end;\n" +
				"while True do ;\n" +
				"procedure B;\n" +
				"begin\n" +
				"   while True do ;\n" +
				"end;\n" +
				"while True do ;\n",
			want: []string{
				`Warning: Infinite loop [line: 1, column: 1]`,
				`Warning: Infinite loop [line: 4, column: 4]`,
				`Warning: Infinite loop [line: 6, column: 1]`,
				`Warning: Infinite loop [line: 9, column: 4]`,
				`Warning: Infinite loop [line: 11, column: 1]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			got := result.DiagnosticStrings()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("diagnostics mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
