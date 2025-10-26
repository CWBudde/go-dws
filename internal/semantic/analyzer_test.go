package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to create an analyzer and analyze source code
func analyzeSource(t *testing.T, input string) (*Analyzer, error) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)
	return analyzer, err
}

// Helper function to check that analysis succeeds
func expectNoErrors(t *testing.T, input string) {
	_, err := analyzeSource(t, input)
	if err != nil {
		t.Errorf("expected no errors, got: %v", err)
	}
}

// Helper function to check that analysis fails with specific error
func expectError(t *testing.T, input string, expectedError string) {
	_, err := analyzeSource(t, input)
	if err == nil {
		t.Errorf("expected error containing '%s', got no error", expectedError)
		return
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error containing '%s', got: %v", expectedError, err)
	}
}

// ============================================================================
// Complex Integration Tests
// ============================================================================

func TestComplexProgram(t *testing.T) {
	input := `
		function Factorial(n: Integer): Integer;
		begin
			if n <= 1 then
				Result := 1
			else
				Result := n * Factorial(n - 1);
		end;

		var result: Integer;
		for i := 1 to 5 do
		begin
			result := Factorial(i);
			PrintLn(result);
		end;
	`
	expectNoErrors(t, input)
}

func TestNestedScopes(t *testing.T) {
	input := `
		var x: Integer := 10;
		begin
			var x: String := 'outer';
			begin
				var x: Float := 3.14;
				PrintLn(x);
			end;
			PrintLn(x);
		end;
		PrintLn(x);
	`
	expectNoErrors(t, input)
}

func TestFunctionWithLocalVariables(t *testing.T) {
	input := `
		function Calculate(n: Integer): Integer;
		begin
			var temp: Integer;
			var result: Integer;
			temp := n * 2;
			result := temp + 10;
			Result := result;
		end;

		var x := Calculate(5);
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestEmptyProgram(t *testing.T) {
	input := ``
	expectNoErrors(t, input)
}

func TestOnlyComments(t *testing.T) {
	input := `
		// This is a comment
		(* This is another comment *)
		{ And another }
	`
	expectNoErrors(t, input)
}

func TestMultipleErrors(t *testing.T) {
	input := `
		var x: UnknownType;
		var y := z;
		x := 'hello';
	`
	analyzer, err := analyzeSource(t, input)
	if err == nil {
		t.Error("expected errors")
		return
	}

	if len(analyzer.Errors()) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(analyzer.Errors()))
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
