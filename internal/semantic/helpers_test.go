package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestHelperDeclaration tests basic helper declaration analysis
func TestHelperDeclaration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "simple helper for String",
			input: `
				type TStringHelper = helper for String
					function ToUpper: String;
				end;
			`,
			expectError: false,
		},
		{
			name: "record helper",
			input: `
				type TPoint = record
					X: Integer;
					Y: Integer;
				end;

				type TPointHelper = record helper for TPoint
					function Distance: Float;
				end;
			`,
			expectError: false,
		},
		{
			name: "helper for unknown type",
			input: `
				type THelper = helper for UnknownType
					function Test: Integer;
				end;
			`,
			expectError: true,
			errorMsg:    "unknown target type",
		},
		{
			name: "helper with class var",
			input: `
				type TIntHelper = helper for Integer
					class var DefaultValue: Integer;
					function IsPositive: Boolean;
				end;
			`,
			expectError: false,
		},
		{
			name: "helper with class const",
			input: `
				type TMathHelper = helper for Float
					class const PI = 3.14159;
					function Round: Integer;
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
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got no error", tt.errorMsg)
				}
				if tt.errorMsg != "" && !containsError(analyzer.Errors(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestHelperMethodResolution tests that helper methods are resolved correctly
func TestHelperMethodResolution(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "call helper method on String",
			input: `
				type TStringHelper = helper for String
					function ToUpper: String;
				end;

				var s: String;
				begin
					s := 'hello';
					s.ToUpper();
				end.
			`,
			expectError: false,
		},
		{
			name: "call helper method on Integer",
			input: `
				type TIntHelper = helper for Integer
					function IsEven: Boolean;
				end;

				var n: Integer;
				begin
					n := 42;
					n.IsEven();
				end.
			`,
			expectError: false,
		},
		{
			name: "call non-existent helper method",
			input: `
				type TStringHelper = helper for String
					function ToUpper: String;
				end;

				var s: String;
				begin
					s := 'hello';
					s.ToLower();
				end.
			`,
			expectError: true,
			errorMsg:    "no helper with method",
		},
		{
			name: "access helper property",
			input: `
				type TStringHelper = helper for String
					property Length: Integer read GetLength;
				end;

				var s: String;
				var len: Integer;
				begin
					s := 'hello';
					len := s.Length;
				end.
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
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got no error", tt.errorMsg)
				}
				if tt.errorMsg != "" && !containsError(analyzer.Errors(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestMultipleHelpers tests that multiple helpers for the same type work correctly
func TestMultipleHelpers(t *testing.T) {
	input := `
		type TStringHelper1 = helper for String
			function ToUpper: String;
		end;

		type TStringHelper2 = helper for String
			function ToLower: String;
		end;

		var s: String;
		begin
			s := 'Hello';
			s.ToUpper();
			s.ToLower();
		end.
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("Expected no error with multiple helpers, got: %v", err)
	}
}

// TestHelperMethodParameters tests parameter validation for helper methods
func TestHelperMethodParameters(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "correct parameters",
			input: `
				type TStringHelper = helper for String
					function Contains(s: String): Boolean;
				end;

				var str: String;
				var result: Boolean;
				begin
					str := 'hello';
					result := str.Contains('lo');
				end.
			`,
			expectError: false,
		},
		{
			name: "wrong parameter count",
			input: `
				type TStringHelper = helper for String
					function Contains(s: String): Boolean;
				end;

				var str: String;
				begin
					str := 'hello';
					str.Contains();
				end.
			`,
			expectError: true,
			errorMsg:    "expects 1 arguments, got 0",
		},
		{
			name: "wrong parameter type",
			input: `
				type TStringHelper = helper for String
					function Contains(s: String): Boolean;
				end;

				var str: String;
				begin
					str := 'hello';
					str.Contains(42);
				end.
			`,
			expectError: true,
			errorMsg:    "has type Integer, expected String",
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
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got no error", tt.errorMsg)
				}
				if tt.errorMsg != "" && !containsError(analyzer.Errors(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, analyzer.Errors())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if error list contains a specific error message
func containsError(errors []string, substr string) bool {
	for _, err := range errors {
		if strings.Contains(err, substr) {
			return true
		}
	}
	return false
}
