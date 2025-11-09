package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestEnumScopedAccess tests scoped enum access via helper class constants (Task 9.54)
func TestEnumScopedAccess(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "scoped enum access - TColor.Red",
			input: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				begin
					c := TColor.Red;
				end;
			`,
			expectError: false,
		},
		{
			name: "scoped enum access - case insensitive",
			input: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				begin
					c := TColor.RED;
					c := tcolor.red;
					c := TCOLOR.GREEN;
				end;
			`,
			expectError: false,
		},
		{
			name: "unscoped enum access still works",
			input: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				begin
					c := Red;
					c := Green;
					c := Blue;
				end;
			`,
			expectError: false,
		},
		{
			name: "mixed scoped and unscoped access",
			input: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				begin
					c := TColor.Red;
					c := Green;
					c := TColor.Blue;
				end;
			`,
			expectError: false,
		},
		{
			name: "scoped access to non-existent enum value",
			input: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				begin
					c := TColor.Yellow;
				end;
			`,
			expectError: true,
			errorMsg:    "no helper with member 'Yellow'",
		},
		{
			name: "multiple enums with scoped access",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				var c: TColor;
				var s: TSize;
				begin
					c := TColor.Red;
					s := TSize.Medium;
					c := TColor.Blue;
					s := TSize.Large;
				end;
			`,
			expectError: false,
		},
		{
			name: "scoped enum access in expression",
			input: `
				type TColor = (Red, Green, Blue);
				var c: TColor;
				var b: Boolean;
				begin
					c := TColor.Red;
					b := (c = TColor.Red);
					b := (c <> TColor.Green);
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
			analyzer.Analyze(program)

			hasError := len(analyzer.Errors()) > 0

			if tt.expectError && !hasError {
				t.Errorf("Expected error containing '%s', but got no errors", tt.errorMsg)
			}

			if !tt.expectError && hasError {
				t.Errorf("Expected no errors, but got: %v", analyzer.Errors())
			}

			if tt.expectError && hasError && tt.errorMsg != "" {
				// Check if error message contains expected substring
				found := false
				for _, err := range analyzer.Errors() {
					if strings.Contains(strings.ToLower(err), strings.ToLower(tt.errorMsg)) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errorMsg, analyzer.Errors())
				}
			}
		})
	}
}
