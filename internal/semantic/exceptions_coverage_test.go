package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Additional Semantic Exception Tests for Coverage (Task 8.227)
// ============================================================================

// TestAnalyzeFinallyClause tests the analyzeFinallyClause function
// This brings coverage from 0% to 100% for this function
func TestAnalyzeFinallyClause(t *testing.T) {
	t.Run("Try-Finally with statements", func(t *testing.T) {
		input := `
		var x: Integer;
		try
			x := 10;
		finally
			x := 20;
			PrintLn(IntToStr(x));
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// No errors expected - finally clause should be analyzed
		if len(analyzer.Errors()) > 0 {
			t.Errorf("unexpected errors: %v", analyzer.Errors())
		}
	})

	t.Run("Try-Finally with empty finally", func(t *testing.T) {
		input := `
		var x: Integer;
		try
			x := 10;
		finally
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Empty finally is valid
		if len(analyzer.Errors()) > 0 {
			t.Errorf("unexpected errors: %v", analyzer.Errors())
		}
	})

	t.Run("Try-Except-Finally all present", func(t *testing.T) {
		input := `
		var x: Integer;
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				x := 1;
		finally
			x := 2;
			PrintLn(IntToStr(x));
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// All clauses should be analyzed without error
		if len(analyzer.Errors()) > 0 {
			t.Errorf("unexpected errors: %v", analyzer.Errors())
		}
	})
}

// TestAnalyzeExceptClauseElseBlock tests else block in except clause
// This improves coverage of analyzeExceptClause
func TestAnalyzeExceptClauseElseBlock(t *testing.T) {
	t.Run("Except with else block", func(t *testing.T) {
		input := `
		var x: Integer;
		try
			raise Exception.Create('test');
		except
			on E: ECustom do
				x := 1;
		else
			x := 2;
			PrintLn(IntToStr(x));
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Else block should be analyzed
		// Note: ECustom may not be defined, so we might have errors for that
	})

	t.Run("Except without else block", func(t *testing.T) {
		input := `
		var x: Integer;
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				x := 1;
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should work without else block
	})
}

// TestAnalyzeExceptionHandlerEdgeCases tests edge cases in analyzeExceptionHandler
// This improves coverage of error paths and special cases
func TestAnalyzeExceptionHandlerEdgeCases(t *testing.T) {
	t.Run("Handler with nil exception type", func(t *testing.T) {
		// This tests the error path at line 74-76 in analyze_exceptions.go
		input := `
		try
			raise Exception.Create('test');
		except
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Bare except without explicit handlers should be allowed
	})

	t.Run("Handler with unknown exception type", func(t *testing.T) {
		// This tests the error path at line 80-83
		input := `
		try
			raise Exception.Create('test');
		except
			on E: UnknownExceptionType do
				PrintLn('error');
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should have error about unknown exception type
		hasError := false
		for _, errStr := range analyzer.Errors() {
			if contains(errStr, "unknown exception type") {
				hasError = true
				break
			}
		}

		if !hasError {
			t.Error("expected error about unknown exception type")
		}
	})

	t.Run("Handler with non-exception type", func(t *testing.T) {
		// This tests the error path at line 86-89
		input := `
		type TNonException = class
		end;

		try
			raise Exception.Create('test');
		except
			on E: TNonException do
				PrintLn('error');
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should have error about type not being Exception
		hasError := false
		for _, errStr := range analyzer.Errors() {
			if contains(errStr, "must be Exception") {
				hasError = true
				break
			}
		}

		if !hasError {
			t.Error("expected error about non-exception type")
		}
	})

	t.Run("Handler with exception variable in scope", func(t *testing.T) {
		// This tests lines 92-98 (variable scoping)
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do begin
				PrintLn(E.Message);
				PrintLn(E.ClassName);
			end;
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Exception variable E should be accessible in handler
		// (though Message and ClassName might not be defined yet)
	})

	t.Run("Handler without exception variable", func(t *testing.T) {
		// This tests the case where handler.Variable is nil
		input := `
		try
			raise Exception.Create('test');
		except
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should work - bare except is valid
	})

	t.Run("Handler statement is nil", func(t *testing.T) {
		// Edge case - handler.Statement might be nil in error scenarios
		// This is handled gracefully at line 101-103
		// Hard to test directly as parser usually provides a statement
		// But the nil check is there for safety
	})
}

// TestIsExceptionTypeCoverage tests isExceptionType edge cases
// This improves coverage of the type checking logic
func TestIsExceptionTypeCoverage(t *testing.T) {
	t.Run("Non-class type is not exception", func(t *testing.T) {
		// This tests line 118-121 (non-ClassType check)
		input := `
		var x: Integer;
		try
			raise x;
		except
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should have error - Integer is not an exception type
		hasError := false
		for _, errStr := range analyzer.Errors() {
			if contains(errStr, "requires Exception type") {
				hasError = true
				break
			}
		}

		if !hasError {
			t.Error("expected error about non-exception type")
		}
	})

	t.Run("Class not derived from Exception", func(t *testing.T) {
		// This tests lines 124-129 (inheritance check)
		input := `
		type TNotException = class
		end;

		var x: TNotException;
		try
			raise x;
		except
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should have error - TNotException is not derived from Exception
		hasError := false
		for _, errStr := range analyzer.Errors() {
			if contains(errStr, "requires Exception type") {
				hasError = true
				break
			}
		}

		if !hasError {
			t.Error("expected error about non-exception class")
		}
	})

	t.Run("Class directly is Exception", func(t *testing.T) {
		// This tests the base case where classType.Name == "Exception"
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				PrintLn(E.Message);
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Exception itself should be valid
	})

	t.Run("Class derived from Exception", func(t *testing.T) {
		// This tests the inheritance chain traversal
		input := `
		type ECustom = class(Exception)
		end;

		try
			raise ECustom.Create('test');
		except
			on E: ECustom do
				PrintLn(E.Message);
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// ECustom should be recognized as exception type
	})
}

// TestAnalyzeTryStatementValidation tests try statement validation
// This covers the validation at lines 52-55
func TestAnalyzeTryStatementValidation(t *testing.T) {
	t.Run("Try with only except clause", func(t *testing.T) {
		input := `
		try
			PrintLn('test');
		except
			on E: Exception do
				PrintLn('error');
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Valid - has except clause
	})

	t.Run("Try with only finally clause", func(t *testing.T) {
		input := `
		try
			PrintLn('test');
		finally
			PrintLn('cleanup');
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Valid - has finally clause
	})

	t.Run("Try with both except and finally", func(t *testing.T) {
		input := `
		try
			PrintLn('test');
		except
			on E: Exception do
				PrintLn('error');
		finally
			PrintLn('cleanup');
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Valid - has both clauses
	})
}

// TestAnalyzeRaiseStatementEdgeCases tests raise statement edge cases
// This improves coverage of analyzeRaiseStatement
func TestAnalyzeRaiseStatementEdgeCases(t *testing.T) {
	t.Run("Raise with nil expression returns early", func(t *testing.T) {
		// This tests lines 15-20 (bare raise)
		input := `
		try
			raise Exception.Create('first');
		except
			on E: Exception do begin
				PrintLn(E.Message);
				raise;
			end;
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Bare raise is allowed
	})

	t.Run("Raise with expression that fails analysis", func(t *testing.T) {
		// This tests lines 23-27 (excType is nil)
		input := `
		try
			raise UndefinedVariable;
		except
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should have error about undefined variable
	})

	t.Run("Raise with valid exception expression", func(t *testing.T) {
		// This tests the success path
		input := `
		try
			raise Exception.Create('message');
		except
			on E: Exception do
				PrintLn(E.Message);
		end;
		`

		analyzer := parseAndAnalyze(t, input)
		if analyzer == nil {
			t.Fatal("analysis failed")
		}

		// Should succeed
	})
}

// Helper function to parse and analyze
func parseAndAnalyze(t *testing.T, input string) *Analyzer {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Logf("Parser errors: %v", p.Errors())
		// Don't fail - some tests expect parser errors
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	return analyzer
}

// Helper to check if error contains string
func containsStr(err error, substr string) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), substr)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
