package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestValidationPass_HelperMethodSupport tests helper method access validation
// Note: Full helper implementation tests are in helpers_test.go
// These tests focus on the validation pass behavior with helpers
func TestValidationPass_HelperMethodSupport(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "non-existent helper member on Integer",
			input: `
				var x: Integer := 42;
				var s: String;
				s := x.NonExistent;  // No helper with this member
			`,
			expectError: true,
			errorMsg:    "helper",
		},
		{
			name: "member access requires helper when type has no native members",
			input: `
				var x: Integer := 42;
				var y: Integer;
				y := x.SomeField;  // Integers don't have fields, need helper
			`,
			expectError: true,
			errorMsg:    "helper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			_ = analyzer.Analyze(program)

			hasErrors := len(analyzer.Errors()) > 0
			if tt.expectError != hasErrors {
				if tt.expectError {
					t.Errorf("Expected error containing '%s', got no error", tt.errorMsg)
				} else {
					t.Errorf("Expected no error, got: %v", analyzer.Errors())
				}
			}

			if tt.expectError && tt.errorMsg != "" {
				foundExpectedError := false
				for _, err := range analyzer.Errors() {
					if containsError([]string{err}, tt.errorMsg) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			}
		})
	}
}

// TestValidationPass_ContextAwareTypeInference tests type inference with expected types
func TestValidationPass_ContextAwareTypeInference(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "record literal with expected type",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;

				var p: TPoint;
				p := (X: 10, Y: 20);  // Field types inferred from TPoint
			`,
			expectError: false,
		},
		{
			name: "record literal with invalid field",
			input: `
				type TPoint = record
					X, Y: Integer;
				end;

				var p: TPoint;
				p := (X: 10, Z: 20);  // Z doesn't exist in TPoint
			`,
			expectError: true,
			errorMsg:    "field",
		},
		{
			name: "array literal with expected element type",
			input: `
				var arr: array of Integer;
				arr := [1, 2, 3];  // Elements inferred as Integer
			`,
			expectError: false,
		},
		{
			name: "set literal with expected element type",
			input: `
				type TColors = (Red, Green, Blue);
				var colors: set of TColors;
				colors := [Red, Green];  // Elements inferred from set type
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			_ = analyzer.Analyze(program)

			hasErrors := len(analyzer.Errors()) > 0
			if tt.expectError != hasErrors {
				if tt.expectError {
					t.Errorf("Expected error containing '%s', got no error", tt.errorMsg)
				} else {
					t.Errorf("Expected no error, got: %v", analyzer.Errors())
				}
			}

			if tt.expectError && tt.errorMsg != "" {
				foundExpectedError := false
				for _, err := range analyzer.Errors() {
					if containsError([]string{err}, tt.errorMsg) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			}
		})
	}
}

// TestValidationPass_StatementValidation tests various statement validations
func TestValidationPass_StatementValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid for loop",
			input: `
				var i: Integer;
				for i := 1 to 10 do begin
					// Loop body
				end;
			`,
			expectError: false,
		},
		{
			name: "break inside loop",
			input: `
				var i: Integer;
				for i := 1 to 10 do begin
					if i = 5 then
						break;
				end;
			`,
			expectError: false,
		},
		{
			name: "continue inside loop",
			input: `
				var i: Integer;
				for i := 1 to 10 do begin
					if i = 5 then
						continue;
				end;
			`,
			expectError: false,
		},
		{
			name: "break outside loop",
			input: `
				var x: Integer;
				x := 5;
				break;  // Not in a loop
			`,
			expectError: true,
			errorMsg:    "break",
		},
		{
			name: "continue outside loop",
			input: `
				var x: Integer;
				x := 5;
				continue;  // Not in a loop
			`,
			expectError: true,
			errorMsg:    "continue",
		},
		{
			name: "valid case statement",
			input: `
				var x: Integer := 5;
				case x of
					1: begin end;
					2: begin end;
					else begin end;
				end;
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			_ = analyzer.Analyze(program)

			hasErrors := len(analyzer.Errors()) > 0
			if tt.expectError != hasErrors {
				if tt.expectError {
					t.Errorf("Expected error containing '%s', got no error", tt.errorMsg)
				} else {
					t.Errorf("Expected no error, got: %v", analyzer.Errors())
				}
			}

			if tt.expectError && tt.errorMsg != "" {
				foundExpectedError := false
				for _, err := range analyzer.Errors() {
					if containsError([]string{err}, tt.errorMsg) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			}
		})
	}
}

// TestValidationPass_OOPValidation tests OOP-related validations
func TestValidationPass_OOPValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid class member access",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;
				end;

				var p: TPoint;
				var x: Integer;
				x := p.X;
			`,
			expectError: false,
		},
		{
			name: "invalid class member access",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;
				end;

				var p: TPoint;
				var z: Integer;
				z := p.Z;  // Z doesn't exist
			`,
			expectError: true,
			errorMsg:    "member",
		},
		{
			name: "valid method call with auto-invocation",
			input: `
				type TPoint = class
					X, Y: Integer;
					function GetX: Integer;
				end;

				function TPoint.GetX: Integer;
				begin
					Result := X;
				end;

				var p: TPoint;
				var x: Integer;
				x := p.GetX;  // Auto-invoke parameterless method
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			_ = analyzer.Analyze(program)

			hasErrors := len(analyzer.Errors()) > 0
			if tt.expectError != hasErrors {
				if tt.expectError {
					t.Errorf("Expected error containing '%s', got no error", tt.errorMsg)
				} else {
					t.Errorf("Expected no error, got: %v", analyzer.Errors())
				}
			}

			if tt.expectError && tt.errorMsg != "" {
				foundExpectedError := false
				for _, err := range analyzer.Errors() {
					if containsError([]string{err}, tt.errorMsg) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			}
		})
	}
}

// TestValidationPass_ErrorRecovery tests that the validator continues after errors
func TestValidationPass_ErrorRecovery(t *testing.T) {
	input := `
		var x: Integer := "hello";  // Error: type mismatch
		var y: String;
		y := z;                      // Error: undefined variable
		break;                       // Error: break outside loop
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	_ = analyzer.Analyze(program)

	// Should collect all errors, not stop at the first one
	if len(analyzer.Errors()) < 3 {
		t.Errorf("Expected at least 3 errors (type mismatch, undefined var, break outside loop), got %d: %v",
			len(analyzer.Errors()), analyzer.Errors())
	}
}
