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
		errorContains []string
		expectErrors  int
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
		errorContains []string
		expectErrors  int
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
		errorContains []string
		expectErrors  int
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
		errorContains []string
		expectErrors  int
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
		errorContains []string
		expectErrors  int
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

// Unit tests for ErrorRecovery methods

// TestNewErrorRecovery tests creation of ErrorRecovery instance
func TestNewErrorRecovery(t *testing.T) {
	input := "var x: Integer;"
	l := lexer.New(input)
	p := New(l)

	recovery := NewErrorRecovery(p)
	if recovery == nil {
		t.Fatal("NewErrorRecovery returned nil")
	}
	if recovery.parser != p {
		t.Error("ErrorRecovery.parser not set correctly")
	}
}

// TestGetSyncTokens tests synchronization token sets
func TestGetSyncTokens(t *testing.T) {
	tests := []struct {
		name        string
		contains    []lexer.TokenType
		set         SynchronizationSet
		shouldBeNil bool
	}{
		{
			name:     "statement starters",
			set:      SyncStatementStarters,
			contains: []lexer.TokenType{lexer.IF, lexer.WHILE, lexer.FOR, lexer.BEGIN, lexer.VAR},
		},
		{
			name:     "block closers",
			set:      SyncBlockClosers,
			contains: []lexer.TokenType{lexer.END, lexer.ELSE, lexer.UNTIL},
		},
		{
			name:     "declaration starters",
			set:      SyncDeclarationStarters,
			contains: []lexer.TokenType{lexer.VAR, lexer.CONST, lexer.TYPE, lexer.FUNCTION, lexer.PROCEDURE},
		},
		{
			name:     "all sync points",
			set:      SyncAll,
			contains: []lexer.TokenType{lexer.IF, lexer.END, lexer.VAR, lexer.FUNCTION},
		},
		{
			name:        "invalid set (default case)",
			set:         SynchronizationSet(999), // Invalid value
			shouldBeNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tt.set.GetSyncTokens()

			if tt.shouldBeNil {
				if tokens != nil {
					t.Errorf("expected nil for invalid set, got %v", tokens)
				}
				return
			}

			if tokens == nil {
				t.Fatal("GetSyncTokens returned nil")
			}

			// Create a map for quick lookup
			tokenMap := make(map[lexer.TokenType]bool)
			for _, tok := range tokens {
				tokenMap[tok] = true
			}

			// Check that all expected tokens are present
			for _, expected := range tt.contains {
				if !tokenMap[expected] {
					t.Errorf("expected token %s not found in sync set", expected)
				}
			}
		})
	}
}

// TestSynchronizeOn tests basic synchronization
func TestSynchronizeOn(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		syncTokens []lexer.TokenType
		shouldFind bool
	}{
		{
			name:       "find THEN token",
			input:      "if x > 10 invalid tokens then",
			syncTokens: []lexer.TokenType{lexer.THEN},
			shouldFind: true,
		},
		{
			name:       "find END token",
			input:      "begin invalid tokens end",
			syncTokens: []lexer.TokenType{lexer.END},
			shouldFind: true,
		},
		{
			name:       "reach EOF without finding token",
			input:      "10 + 20", // Only numbers and operators, no sync points
			syncTokens: []lexer.TokenType{lexer.THEN},
			shouldFind: false,
		},
		{
			name:       "find first of multiple tokens",
			input:      "if x then",
			syncTokens: []lexer.TokenType{lexer.THEN, lexer.ELSE},
			shouldFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Skip first token to simulate being in middle of parsing
			p.nextToken()
			p.nextToken()

			found := recovery.SynchronizeOn(tt.syncTokens...)
			if found != tt.shouldFind {
				t.Errorf("expected SynchronizeOn to return %v, got %v", tt.shouldFind, found)
			}
		})
	}
}

// TestSynchronizeOnSet tests synchronization with predefined sets
func TestSynchronizeOnSet(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		additionalTokens []lexer.TokenType
		set              SynchronizationSet
		shouldFind       bool
	}{
		{
			name:       "sync on statement starters",
			input:      "invalid tokens if x then",
			set:        SyncStatementStarters,
			shouldFind: true,
		},
		{
			name:       "sync on block closers",
			input:      "invalid tokens end;",
			set:        SyncBlockClosers,
			shouldFind: true,
		},
		{
			name:             "sync with additional tokens",
			input:            "invalid tokens custom_token",
			set:              SyncStatementStarters,
			additionalTokens: []lexer.TokenType{lexer.IDENT}, // custom_token will be IDENT
			shouldFind:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Skip first token
			p.nextToken()
			p.nextToken()

			found := recovery.SynchronizeOnSet(tt.set, tt.additionalTokens...)
			if found != tt.shouldFind {
				t.Errorf("expected SynchronizeOnSet to return %v, got %v", tt.shouldFind, found)
			}
		})
	}
}

// TestAddExpectError tests expect error reporting
func TestAddExpectError(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		context     string
		expected    lexer.TokenType
		shouldError bool
	}{
		{
			name:        "missing THEN with context",
			input:       "if x > 10",
			expected:    lexer.THEN,
			context:     "after if condition",
			shouldError: true,
		},
		{
			name:        "missing DO without context",
			input:       "while x > 0",
			expected:    lexer.DO,
			context:     "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Move to end of input
			for p.cursor.Peek(1).Type != lexer.EOF {
				p.nextToken()
			}

			recovery.AddExpectError(tt.expected, tt.context)

			errors := p.Errors()
			if tt.shouldError && len(errors) == 0 {
				t.Error("expected error to be added")
			}
			if tt.shouldError && len(errors) > 0 {
				errMsg := errors[0].Message
				if !strings.Contains(errMsg, tt.expected.String()) {
					t.Errorf("error message should contain expected token %s, got: %s", tt.expected, errMsg)
				}
				if tt.context != "" && !strings.Contains(errMsg, tt.context) {
					t.Errorf("error message should contain context '%s', got: %s", tt.context, errMsg)
				}
			}
		})
	}
}

// TestAddExpectErrorWithSuggestion tests error with suggestion
func TestAddExpectErrorWithSuggestion(t *testing.T) {
	input := "if x > 10"
	l := lexer.New(input)
	p := New(l)
	recovery := NewErrorRecovery(p)

	// Move to end
	for p.cursor.Peek(1).Type != lexer.EOF {
		p.nextToken()
	}

	recovery.AddExpectErrorWithSuggestion(
		lexer.THEN,
		"after if condition",
		"add 'then' keyword after condition",
	)

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected error to be added")
	}

	// Check that error was created (suggestion is stored in StructuredParserError
	// but converted to ParserError, so we just verify the error exists)
	err := errors[0]
	if !strings.Contains(strings.ToLower(err.Message), "then") {
		t.Errorf("error message should mention 'then', got: %s", err.Message)
	}
}

// TestAddContextError tests context error reporting
func TestAddContextError(t *testing.T) {
	input := "begin var x: Integer;"
	l := lexer.New(input)
	p := New(l)
	recovery := NewErrorRecovery(p)

	// Set up a block context
	p.pushBlockContext("begin", p.cursor.Current().Pos)

	recovery.AddContextError("expected 'end'", ErrMissingEnd)

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected error to be added")
	}
}

// TestAddError tests basic error reporting
func TestAddError(t *testing.T) {
	input := "var x: Integer;"
	l := lexer.New(input)
	p := New(l)
	recovery := NewErrorRecovery(p)

	recovery.AddError("test error message", "TEST_ERROR")

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected error to be added")
	}
	if errors[0].Message != "test error message" {
		t.Errorf("expected error message 'test error message', got: %s", errors[0].Message)
	}
}

// TestAddStructuredError tests structured error reporting
func TestAddStructuredError(t *testing.T) {
	input := "var x: Integer;"
	l := lexer.New(input)
	p := New(l)
	recovery := NewErrorRecovery(p)

	structErr := NewStructuredError(ErrKindMissing).
		WithCode(ErrMissingThen).
		WithPosition(p.cursor.Current().Pos, 1).
		WithExpected(lexer.THEN).
		WithActual(lexer.EOF, "EOF").
		WithSuggestion("add 'then' keyword").
		Build()

	recovery.AddStructuredError(structErr)

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected error to be added")
	}
}

// TestTryRecover tests recovery attempt
func TestTryRecover(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		syncTokens []lexer.TokenType
		shouldFind bool
	}{
		{
			name:       "successful recovery",
			input:      "invalid tokens then x := 10;",
			syncTokens: []lexer.TokenType{lexer.THEN},
			shouldFind: true,
		},
		{
			name:       "failed recovery at EOF",
			input:      "invalid tokens",
			syncTokens: []lexer.TokenType{lexer.THEN},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Skip first tokens
			p.nextToken()
			p.nextToken()

			result := recovery.TryRecover(tt.syncTokens...)
			if result != tt.shouldFind {
				t.Errorf("expected TryRecover to return %v, got %v", tt.shouldFind, result)
			}
		})
	}
}

// TestExpectWithRecovery tests combined expect and recovery
func TestExpectWithRecovery(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		context       string
		syncTokens    []lexer.TokenType
		expected      lexer.TokenType
		shouldSucceed bool
	}{
		{
			name:          "successful expect",
			input:         "x then", // Need something before THEN so THEN is in peek position
			expected:      lexer.THEN,
			context:       "test",
			shouldSucceed: true,
		},
		{
			name:          "failed expect with recovery",
			input:         "invalid other then", // THEN is not in peek position (peek is "other")
			expected:      lexer.THEN,
			context:       "test",
			syncTokens:    []lexer.TokenType{lexer.THEN},
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Cursor is already at first token after New()
			// ExpectWithRecovery checks Peek(1) (next token), so we're positioned correctly

			result := recovery.ExpectWithRecovery(tt.expected, tt.context, tt.syncTokens...)
			if result != tt.shouldSucceed {
				t.Errorf("expected ExpectWithRecovery to return %v, got %v (current: %s, peek: %s)",
					tt.shouldSucceed, result, p.cursor.Current().Type, p.cursor.Peek(1).Type)
			}

			if !tt.shouldSucceed {
				errors := p.Errors()
				if len(errors) == 0 {
					t.Error("expected error to be reported on failed expect")
				}
			}
		})
	}
}

// TestExpectOneOf tests expecting one of multiple tokens
func TestExpectOneOf(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		context     string
		expected    []lexer.TokenType
		shouldMatch lexer.TokenType
	}{
		{
			name:        "match first option",
			input:       "x if", // Need something before so IF is in peek position
			expected:    []lexer.TokenType{lexer.IF, lexer.WHILE},
			context:     "test",
			shouldMatch: lexer.IF,
		},
		{
			name:        "match second option",
			input:       "x while", // Need something before so WHILE is in peek position
			expected:    []lexer.TokenType{lexer.IF, lexer.WHILE},
			context:     "test",
			shouldMatch: lexer.WHILE,
		},
		{
			name:        "no match",
			input:       "x invalid", // Need something before so IDENT is in peek position
			expected:    []lexer.TokenType{lexer.IF, lexer.WHILE},
			context:     "test",
			shouldMatch: lexer.ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Cursor is already at first token after New()
			// ExpectOneOf checks Peek(1) (next token), so we're positioned correctly

			result := recovery.ExpectOneOf(tt.expected, tt.context, lexer.SEMICOLON)
			if result != tt.shouldMatch {
				t.Errorf("expected %s, got %s (current: %s, peek: %s)",
					tt.shouldMatch, result, p.cursor.Current().Type, p.cursor.Peek(1).Type)
			}

			if tt.shouldMatch == lexer.ILLEGAL {
				errors := p.Errors()
				if len(errors) == 0 {
					t.Error("expected error when no match found")
				}
			}
		})
	}
}

// TestErrorRecoverySkipUntil tests skipping tokens until specific token found
func TestErrorRecoverySkipUntil(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		skipUntil  []lexer.TokenType
		shouldFind bool
	}{
		{
			name:       "find semicolon",
			input:      "invalid tokens here ;",
			skipUntil:  []lexer.TokenType{lexer.SEMICOLON},
			shouldFind: true,
		},
		{
			name:       "find END",
			input:      "invalid tokens end",
			skipUntil:  []lexer.TokenType{lexer.END},
			shouldFind: true,
		},
		{
			name:       "reach EOF",
			input:      "invalid tokens",
			skipUntil:  []lexer.TokenType{lexer.SEMICOLON},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			p.nextToken() // Initialize
			p.nextToken() // Skip first token

			result := recovery.SkipUntil(tt.skipUntil...)
			if result != tt.shouldFind {
				t.Errorf("expected SkipUntil to return %v, got %v", tt.shouldFind, result)
			}
		})
	}
}

// TestIsAtSyncPoint tests sync point detection
func TestIsAtSyncPoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		skipTo   int
		expected bool
	}{
		{
			name:     "at IF statement",
			input:    "if x then",
			skipTo:   0,
			expected: true,
		},
		{
			name:     "at END keyword",
			input:    "end;",
			skipTo:   0,
			expected: true,
		},
		{
			name:     "at number (not sync point)",
			input:    "42 + 10", // Numbers and operators are not sync points
			skipTo:   0,
			expected: false,
		},
		{
			name:     "at BEGIN",
			input:    "begin",
			skipTo:   0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			recovery := NewErrorRecovery(p)

			// Cursor is already positioned at first token after New()
			// Skip additional tokens if needed
			for i := 0; i < tt.skipTo; i++ {
				p.nextToken()
			}

			result := recovery.IsAtSyncPoint()
			if result != tt.expected {
				t.Errorf("expected IsAtSyncPoint to return %v, got %v (current token: %s)",
					tt.expected, result, p.cursor.Current().Type)
			}
		})
	}
}

// TestSuggestMissingDelimiter tests delimiter suggestion
func TestSuggestMissingDelimiter(t *testing.T) {
	l := lexer.New("begin")
	p := New(l)
	recovery := NewErrorRecovery(p)

	suggestion := recovery.SuggestMissingDelimiter(lexer.END, "begin block")
	if !strings.Contains(strings.ToLower(suggestion), "end") {
		t.Errorf("suggestion should mention 'end', got: %s", suggestion)
	}
	if !strings.Contains(suggestion, "begin block") {
		t.Errorf("suggestion should mention 'begin block', got: %s", suggestion)
	}
}

// TestSuggestMissingSeparator tests separator suggestion
func TestSuggestMissingSeparator(t *testing.T) {
	l := lexer.New("var x")
	p := New(l)
	recovery := NewErrorRecovery(p)

	suggestion := recovery.SuggestMissingSeparator(lexer.SEMICOLON, "statements")
	// The suggestion will contain the token string representation (e.g., "SEMICOLON")
	if !strings.Contains(strings.ToLower(suggestion), "semicolon") {
		t.Errorf("suggestion should mention 'semicolon', got: %s", suggestion)
	}
	if !strings.Contains(suggestion, "statements") {
		t.Errorf("suggestion should mention 'statements', got: %s", suggestion)
	}
}

// TestGetErrorCodeForMissingToken tests error code mapping
func TestGetErrorCodeForMissingToken(t *testing.T) {
	tests := []struct {
		expectedCode string
		token        lexer.TokenType
	}{
		{token: lexer.THEN, expectedCode: ErrMissingThen},
		{token: lexer.DO, expectedCode: ErrMissingDo},
		{token: lexer.END, expectedCode: ErrMissingEnd},
		{token: lexer.SEMICOLON, expectedCode: ErrMissingSemicolon},
		{token: lexer.COLON, expectedCode: ErrMissingColon},
		{token: lexer.RPAREN, expectedCode: ErrMissingRParen},
		{token: lexer.RBRACE, expectedCode: ErrMissingRBrace},
		{token: lexer.RBRACK, expectedCode: ErrMissingRBracket},
		{token: lexer.ASSIGN, expectedCode: ErrMissingAssign},
		{token: lexer.OF, expectedCode: ErrMissingOf},
		{token: lexer.TO, expectedCode: ErrMissingTo},
		{token: lexer.DOWNTO, expectedCode: ErrMissingTo},
		{token: lexer.IDENT, expectedCode: ErrUnexpectedToken}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.token.String(), func(t *testing.T) {
			code := getErrorCodeForMissingToken(tt.token)
			if code != tt.expectedCode {
				t.Errorf("expected error code %s for token %s, got %s",
					tt.expectedCode, tt.token, code)
			}
		})
	}
}
