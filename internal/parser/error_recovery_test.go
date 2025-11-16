package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestErrorRecoveryBlockStatement tests error recovery in begin...end blocks
func TestErrorRecoveryBlockStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  int
		errorContains []string
	}{
		{
			name: "missing end keyword",
			input: `
			begin
				var x: Integer := 42;
				var y: String := 'hello'
			`,
			expectErrors:  1,
			errorContains: []string{"expected 'end'", "begin block"},
		},
		{
			name: "multiple errors in block",
			input: `
			begin
				var x Integer; // missing colon
				y := 10; // undefined var
			end;
			`,
			expectErrors:  2,          // One for missing colon, one for the parser continuing
			errorContains: []string{}, // Just check that multiple errors are reported
		},
		{
			name: "nested blocks with missing end",
			input: `
			begin
				var x: Integer := 1;
				begin
					var y: Integer := 2
				// missing end for inner block
			end;
			`,
			expectErrors:  1,
			errorContains: []string{"expected 'end'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) < tt.expectErrors {
				t.Errorf("expected at least %d errors, got %d", tt.expectErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s", err.Message)
				}
				return
			}

			// Check that error messages contain expected strings
			allErrors := make([]string, len(errors))
			for i, err := range errors {
				allErrors[i] = err.Message
			}
			combinedErrors := strings.Join(allErrors, " | ")

			for _, contains := range tt.errorContains {
				found := false
				for _, errMsg := range allErrors {
					if strings.Contains(errMsg, contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message to contain '%s', got:\n%s", contains, combinedErrors)
				}
			}
		})
	}
}

// TestErrorRecoveryIfStatement tests error recovery in if statements
func TestErrorRecoveryIfStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  int
		errorContains []string
	}{
		{
			name: "missing then keyword",
			input: `
			if x > 10
				Print('big');
			`,
			expectErrors:  1,
			errorContains: []string{"expected 'then'", "if block"},
		},
		{
			name: "missing condition",
			input: `
			if then
				Print('invalid');
			`,
			expectErrors:  1,
			errorContains: []string{"if block"},
		},
		{
			name: "missing consequence",
			input: `
			if x > 10 then
			else
				Print('small');
			`,
			expectErrors:  1,
			errorContains: []string{}, // Error detected, context not always present
		},
		{
			name: "multiple if errors",
			input: `
			if x > 10
				y := 20; // missing then
			if z < 5 then
			// missing statement after then
			`,
			expectErrors:  2,
			errorContains: []string{"if block"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) < tt.expectErrors {
				t.Errorf("expected at least %d errors, got %d", tt.expectErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s", err.Message)
				}
				return
			}

			// Check that error messages contain expected strings
			allErrors := make([]string, len(errors))
			for i, err := range errors {
				allErrors[i] = err.Message
			}
			combinedErrors := strings.Join(allErrors, " | ")

			for _, contains := range tt.errorContains {
				found := false
				for _, errMsg := range allErrors {
					if strings.Contains(errMsg, contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message to contain '%s', got:\n%s", contains, combinedErrors)
				}
			}
		})
	}
}

// TestErrorRecoveryWhileStatement tests error recovery in while loops
func TestErrorRecoveryWhileStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  int
		errorContains []string
	}{
		{
			name: "missing do keyword",
			input: `
			while x > 0
				x := x - 1;
			`,
			expectErrors:  1,
			errorContains: []string{"expected 'do'", "while block"},
		},
		{
			name: "missing condition",
			input: `
			while do
				x := x - 1;
			`,
			expectErrors:  1,
			errorContains: []string{"while block"},
		},
		{
			name: "missing body",
			input: `
			while x > 0 do
			end;
			`,
			expectErrors:  1,
			errorContains: []string{}, // Error detected, context not always present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) < tt.expectErrors {
				t.Errorf("expected at least %d errors, got %d", tt.expectErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s", err.Message)
				}
				return
			}

			// Check that error messages contain expected strings
			allErrors := make([]string, len(errors))
			for i, err := range errors {
				allErrors[i] = err.Message
			}
			combinedErrors := strings.Join(allErrors, " | ")

			for _, contains := range tt.errorContains {
				found := false
				for _, errMsg := range allErrors {
					if strings.Contains(errMsg, contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message to contain '%s', got:\n%s", contains, combinedErrors)
				}
			}
		})
	}
}

// TestErrorRecoveryRepeatStatement tests error recovery in repeat-until loops
func TestErrorRecoveryRepeatStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  int
		errorContains []string
	}{
		{
			name: "missing until keyword",
			input: `
			repeat
				x := x + 1;
			end;
			`,
			expectErrors:  1,
			errorContains: []string{"expected 'until'", "repeat block"},
		},
		{
			name: "empty repeat body",
			input: `
			repeat
			until x > 10;
			`,
			expectErrors:  1,
			errorContains: []string{"repeat block"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) < tt.expectErrors {
				t.Errorf("expected at least %d errors, got %d", tt.expectErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s", err.Message)
				}
				return
			}

			// Check that error messages contain expected strings
			allErrors := make([]string, len(errors))
			for i, err := range errors {
				allErrors[i] = err.Message
			}
			combinedErrors := strings.Join(allErrors, " | ")

			for _, contains := range tt.errorContains {
				found := false
				for _, errMsg := range allErrors {
					if strings.Contains(errMsg, contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message to contain '%s', got:\n%s", contains, combinedErrors)
				}
			}
		})
	}
}

// TestErrorRecoveryCaseStatement tests error recovery in case statements
func TestErrorRecoveryCaseStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  int
		errorContains []string
	}{
		{
			name: "missing end keyword",
			input: `
			case x of
				1: Print('one');
				2: Print('two');
			`,
			expectErrors:  1,
			errorContains: []string{"expected 'end'", "case block"},
		},
		{
			name: "missing of keyword",
			input: `
			case x
				1: Print('one');
			end;
			`,
			expectErrors:  1,
			errorContains: []string{}, // Error detected, context not always present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) < tt.expectErrors {
				t.Errorf("expected at least %d errors, got %d", tt.expectErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s", err.Message)
				}
				return
			}

			// Check that error messages contain expected strings
			allErrors := make([]string, len(errors))
			for i, err := range errors {
				allErrors[i] = err.Message
			}
			combinedErrors := strings.Join(allErrors, " | ")

			for _, contains := range tt.errorContains {
				found := false
				for _, errMsg := range allErrors {
					if strings.Contains(errMsg, contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error message to contain '%s', got:\n%s", contains, combinedErrors)
				}
			}
		})
	}
}

// TestMultipleErrorsReported tests that parser continues after errors and reports multiple issues
func TestMultipleErrorsReported(t *testing.T) {
	input := `
	begin
		var x Integer; // missing colon
	end;

	if y > 10  // missing then
		z := 20;

	while a < 5  // missing do
		b := b + 1;
	`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	errors := p.Errors()
	if len(errors) < 3 {
		t.Errorf("expected at least 3 errors (missing colon, missing then, missing do), got %d", len(errors))
		for _, err := range errors {
			t.Logf("  Error: %s", err.Message)
		}
	}
}

// TestContextInNestedBlocks tests that error messages include proper context for nested blocks
func TestContextInNestedBlocks(t *testing.T) {
	input := `
	begin
		var x: Integer := 1;
		if x > 0 then
			begin
				var y: Integer := 2
			// missing end for inner begin block
		// also missing end for if
	end; // this end closes the outer begin
	`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected at least one error, got none")
	}

	// At least one error should mention a block context
	foundContextError := false
	for _, err := range errors {
		t.Logf("Error: %s", err.Message)
		if strings.Contains(err.Message, "block") {
			foundContextError = true
		}
	}

	if !foundContextError {
		t.Error("expected at least one error to mention block context")
	}
}

// TestSynchronizationPreventsInfiniteLoops tests that synchronization prevents parser from looping forever
func TestSynchronizationPreventsInfiniteLoops(t *testing.T) {
	// This input has many errors but parser should not hang
	input := `
	begin
		var x Integer
		y := 10
		if z then
			a := 5
		while b do
			c := 6
	end
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// If we get here without hanging, synchronization worked
	if program == nil {
		t.Fatal("expected program to be parsed despite errors")
	}

	errors := p.Errors()
	if len(errors) == 0 {
		t.Error("expected errors to be reported")
	}

	// Log errors for manual inspection
	t.Logf("Parser recovered from %d errors:", len(errors))
	for _, err := range errors {
		t.Logf("  - %s", err.Message)
	}
}
