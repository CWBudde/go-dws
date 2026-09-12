package frontend

import (
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestMalformedInputDoesNotCrash pins the compile pipeline against the inputs
// that used to segfault it.
//
// Each of these parses into a tree the parser could not complete, and each one
// killed the process rather than reporting a diagnostic: a malformed routine
// header left a typed-nil *ast.FunctionDecl at the top level and the generic
// monomorphizer faulted on it (in scripts using no generics at all), an
// unsupported function-pointer return type left a typed-nil element type that
// faulted inside the parser, and an empty `uses` left a nil unit name.
//
// The contract this test defends is narrow and absolute: whatever the compiler
// makes of a malformed program, it reports it. Message parity for these inputs
// is a separate matter and deliberately not asserted here.
func TestMalformedInputDoesNotCrash(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"lazy var parameter", "procedure Bug1(lazy var a : Integer);\nbegin\nend;\n"},
		{"type before name in const parameter", "procedure Test(const Integer v);\nbegin\n   v:=1;\nend;"},
		{"unterminated ensure", "procedure Test;\nbegin\nensure"},
		{"unterminated require", "procedure Test;\nbegin\nrequire"},
		{"function pointer returning a procedure", "procedure t(const a : Array Of function : procedure);\nbegin\nend;\n"},
		{"operator uses nothing", "operator + (TObject, TObject) : Integer uses ;"},
		{"uses with a file qualifier", "uses Classes in 'Classes.pas';\n"},
		{"parameter list runs off the end", "procedure Dummy7(a : Integer"},
		{"record field after a method", "type\n   TRec = record\n      A : Integer;\n      function Test : TRec;\n      begin\n         Result.A:=1;\n      end;\n      B : Integer;\n   end;\n"},
		// These reach the tree through a wrapper that converts the concrete
		// pointer to ast.Statement itself, so they bypass any per-dispatch-case
		// normalization; they are covered by the statement boundary instead.
		{"incomplete class method", "class function Foo(a"},
		{"incomplete constructor", "constructor Create(a"},
		{"incomplete destructor", "destructor Destroy(a"},
		{"incomplete for-in", "for x in "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A panic here fails the test rather than taking the process down.
			result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
			if result == nil {
				t.Fatal("Compile returned nil")
			}
			if len(result.Diagnostics) == 0 {
				t.Fatal("malformed input compiled clean; expected at least one diagnostic")
			}
			for _, diag := range result.Diagnostics {
				if strings.Contains(diag.Code, "PANIC") {
					t.Fatalf("pipeline panicked instead of reporting: %s", diag.Rendered)
				}
			}
		})
	}
}

// TestMalformedInputTerminates pins the compiler against inputs that made it
// loop forever rather than report.
//
// A misplaced record field used to spin: the recovery helper lists IDENT among
// its safe points, so asked to recover *from* an identifier it returned without
// advancing, and the record-body loop reported the same token until memory ran
// out. A hang is worse than a crash — nothing is reported and nothing fails —
// so this asserts termination explicitly rather than leaving it to the package
// test timeout.
func TestMalformedInputTerminates(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			"record field after a method",
			"type\n   TRec = record\n      A : Integer;\n      function Test : TRec;\n      begin\n         Result.A:=1;\n      end;\n      B : Integer;\n   end;\n",
		},
		{
			"several fields after a method",
			"type\n   TRec = record\n      function Test : Integer;\n      begin\n      end;\n      A : Integer;\n      B : Integer;\n      C : Integer;\n   end;\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan int, 1)
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						done <- -1
					}
				}()
				result := Compile(tt.source, "<test>", semantic.HintsLevelPedantic)
				done <- len(result.Diagnostics)
			}()

			select {
			case count := <-done:
				if count < 0 {
					t.Fatal("compile panicked")
				}
				if count == 0 {
					t.Fatal("malformed input compiled clean; expected at least one diagnostic")
				}
			case <-time.After(10 * time.Second):
				// Leaving the goroutine running is deliberate: it is allocating,
				// and this process is about to fail anyway.
				t.Fatal("compile did not terminate within 10s")
			}
		})
	}
}

// TestMalformedInputRendersBestEffortAST pins the public promise that a parse
// result can be traversed.
//
// pkg/dwscript hands callers a best-effort AST alongside the accumulated parser
// diagnostics, which an editor integration walks — Program.String() being the
// simplest way to do it. A typed-nil statement in that tree faults on the walk
// rather than on the parse, so the caller crashes while doing exactly what the
// API invites. Partially typed declarations are the normal state of a file being
// edited, so these are ordinary inputs, not exotic ones.
func TestMalformedInputRendersBestEffortAST(t *testing.T) {
	sources := []string{
		"class function Foo(a",
		"constructor Create(a",
		"destructor Destroy(a",
		"procedure Bug1(lazy var a : Integer);\nbegin\nend;\n",
		"for x in ",
		"procedure t(const a : Array Of function : procedure);\nbegin\nend;\n",
	}

	for _, source := range sources {
		t.Run(source, func(_ *testing.T) {
			result := ParseWithOptions(source, Options{Filename: "<test>"})
			if result.Program == nil {
				return
			}
			// Walking the tree must not fault, whatever the parser made of it.
			// A fault here fails the test as a panic; there is nothing to assert.
			_ = result.Program.String()
		})
	}
}
