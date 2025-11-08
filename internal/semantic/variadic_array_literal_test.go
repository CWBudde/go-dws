package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestVariadicArrayLiterals verifies array literal support for variadic function calls
// Task 9.4.5: Array Literal Support for Variadic Calls
//
// This test documents what works and what doesn't:
// ✓ Lambdas in array literals for variadic parameters
// ✓ Mixed types in array literal for 'array of const'
// ✗ Inline function expressions (parser limitation)
func TestVariadicArrayLiterals(t *testing.T) {
	t.Run("lambdas in array literal for variadic parameter", func(t *testing.T) {
		// This should work - lambdas are supported
		input := `
			type TIntMapper = function(x: Integer): Integer;

			procedure ApplyAll(const funcs: array of TIntMapper);
			begin
			end;

			begin
				ApplyAll([
					lambda(x) => x * 2,
					lambda(x) => x + 1,
					lambda(x) => x - 1
				]);
			end;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		analyzer := NewAnalyzer()
		err := analyzer.Analyze(program)

		if len(analyzer.Errors()) > 0 {
			t.Fatalf("analyzer errors: %v", analyzer.Errors())
		}
		if err != nil {
			t.Fatalf("analyzer error: %v", err)
		}

		t.Log("✓ Lambdas in array literals for variadic parameters: WORKS")
	})

	t.Run("mixed types in array literal for array of const", func(t *testing.T) {
		// This should work - Format accepts array of const (Variant)
		input := `
			begin
				PrintLn(Format('test %s %d %f', ['hello', 42, 3.14]));
			end;
		`

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		analyzer := NewAnalyzer()
		err := analyzer.Analyze(program)

		if len(analyzer.Errors()) > 0 {
			t.Fatalf("analyzer errors: %v", analyzer.Errors())
		}
		if err != nil {
			t.Fatalf("analyzer error: %v", err)
		}

		t.Log("✓ Mixed types in array literal for 'array of const': WORKS")
	})

	t.Run("inline function expressions in array literal", func(t *testing.T) {
		// This is expected to FAIL at parse time - inline function expressions not supported
		input := `
			type TIntMapper = function(x: Integer): Integer;

			procedure ApplyAll(const funcs: array of TIntMapper);
			begin
			end;

			begin
				ApplyAll([
					function(x: Integer): Integer begin Result := x * 2; end,
					function(x: Integer): Integer begin Result := x + 1; end
				]);
			end;
		`

		l := lexer.New(input)
		p := parser.New(l)
		_ = p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Log("✗ Inline function expressions: NOT SUPPORTED (parser limitation)")
			t.Log("  Parser errors:", p.Errors()[0])
			return // Expected failure
		}

		// If we get here, the parser unexpectedly succeeded
		t.Fatal("Unexpected: parser accepted inline function expressions")
	})
}
