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
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if perrs := p.Errors(); len(perrs) > 0 {
		for _, err := range perrs {
			if ErrorMatches(err.Message, expectedError) {
				return
			}
		}
		t.Fatalf("expected error containing '%s', got parser errors: %v", expectedError, perrs)
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)
	if err == nil {
		t.Errorf("expected error containing '%s', got no error", expectedError)
		return
	}

	if !ErrorMatches(err.Error(), expectedError) {
		t.Errorf("expected error containing '%s', got: %v", expectedError, err)
	}
}

// ErrorMatches checks if an error message matches the expected pattern.
// It handles variations in error message formatting after error consolidation.
// This is exported so other test files can use it.
func ErrorMatches(actual, expected string) bool {
	actualLower := strings.ToLower(actual)
	expectedLower := strings.ToLower(expected)

	// Direct substring match
	if strings.Contains(actualLower, expectedLower) {
		return true
	}

	// Handle "undefined variable 'X'" → "unknown name \"X\""
	// Also handles just "undefined variable" → "unknown name"
	if strings.Contains(expectedLower, "undefined variable") {
		varName := extractQuotedName(expectedLower, "undefined variable")
		if varName != "" {
			// Specific variable name given
			if strings.Contains(actualLower, "unknown name") && strings.Contains(actualLower, strings.ToLower(varName)) {
				return true
			}
		} else {
			// No specific name, just check for "unknown name"
			if strings.Contains(actualLower, "unknown name") {
				return true
			}
		}
	}

	// Handle "undefined" alone → "unknown"
	if expectedLower == "undefined" && strings.Contains(actualLower, "unknown") {
		return true
	}

	// Handle "cannot assign X to Y" → "cannot assign \"X\" to \"Y\""
	// The quotes in the actual message prevent direct match, so remove quotes from actual
	actualNoQuotes := strings.ReplaceAll(actualLower, "\"", "")
	if strings.Contains(actualNoQuotes, expectedLower) {
		return true
	}

	// Handle "array index must be integer" → "array index expected \"integer\""
	if strings.Contains(expectedLower, "array index must be integer") {
		if strings.Contains(actualLower, "array index expected") && strings.Contains(actualLower, "integer") {
			return true
		}
	}

	// Handle "has type X, expected Y" → "expects type \"Y\" instead of \"X\""
	// or "argument N to function" → "Argument N expects type"
	if strings.Contains(expectedLower, "has type") && strings.Contains(expectedLower, "expected") {
		// Extract the types from the expected pattern: "has type X, expected Y"
		parts := strings.Split(expectedLower, ",")
		if len(parts) == 2 {
			// Extract X and Y
			hasTypePart := strings.TrimSpace(parts[0])
			expectedPart := strings.TrimSpace(parts[1])

			// Get type X (after "has type")
			if idx := strings.Index(hasTypePart, "has type"); idx != -1 {
				typeX := strings.TrimSpace(hasTypePart[idx+len("has type"):])

				// Get type Y (after "expected")
				if idx := strings.Index(expectedPart, "expected"); idx != -1 {
					typeY := strings.TrimSpace(expectedPart[idx+len("expected"):])

					// Check if actual contains both types (with or without quotes)
					actualNoQuotes := strings.ReplaceAll(actualLower, "\"", "")
					if strings.Contains(actualNoQuotes, typeX) && strings.Contains(actualNoQuotes, typeY) {
						// Make sure "expects type" is mentioned
						if strings.Contains(actualLower, "expects type") || strings.Contains(actualLower, "expected") {
							return true
						}
					}
				}
			}
		}
	}

	// Handle "argument N to function" → "Argument N expects type"
	if strings.Contains(expectedLower, "argument") && strings.Contains(expectedLower, "to function") {
		if strings.Contains(actualLower, "argument") && strings.Contains(actualLower, "expects type") {
			return true
		}
	}

	return false
}

// extractQuotedName extracts a name from a pattern like "undefined variable 'name'"
func extractQuotedName(text, prefix string) string {
	// Find the prefix
	idx := strings.Index(text, prefix)
	if idx == -1 {
		return ""
	}

	// Look for quoted name after prefix
	rest := text[idx+len(prefix):]

	// Try single quotes
	if start := strings.Index(rest, "'"); start != -1 {
		if end := strings.Index(rest[start+1:], "'"); end != -1 {
			return rest[start+1 : start+1+end]
		}
	}

	// Try double quotes
	if start := strings.Index(rest, "\""); start != -1 {
		if end := strings.Index(rest[start+1:], "\""); end != -1 {
			return rest[start+1 : start+1+end]
		}
	}

	return ""
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
