package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestInfiniteLoopDetection(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectWarnings []string // Substring match for warnings
	}{
		{
			name:           "while True - infinite loop",
			input:          "while True do ;",
			expectWarnings: []string{"Infinite loop"},
		},
		{
			name:           "while False - not infinite (never executes)",
			input:          "while False do ;",
			expectWarnings: []string{}, // No warning
		},
		{
			name:           "while True with break - not infinite",
			input:          "while True do break;",
			expectWarnings: []string{}, // Break makes it exitable
		},
		{
			name:           "while True with exit - not infinite",
			input:          "procedure Test; begin while True do exit; end;",
			expectWarnings: []string{}, // Exit makes it exitable
		},
		{
			name:           "repeat until False - infinite loop",
			input:          "repeat PrintLn(''); until False;",
			expectWarnings: []string{"Infinite loop"},
		},
		{
			name:           "repeat until True - not infinite (exits after first iteration)",
			input:          "repeat PrintLn(''); until True;",
			expectWarnings: []string{}, // No warning
		},
		{
			name:           "repeat until False with break - not infinite",
			input:          "repeat break; until False;",
			expectWarnings: []string{}, // Break makes it exitable
		},
		{
			name: "nested loops - inner infinite",
			input: `while True do begin
				while True do ;
			end;`,
			expectWarnings: []string{"Infinite loop"}, // Both loops are infinite
		},
		{
			name: "nested loops - outer with break, inner without",
			input: `while True do begin
				while True do ;
				break;
			end;`,
			expectWarnings: []string{"Infinite loop"}, // Inner loop is still infinite (only outer has break)
		},
		{
			name: "nested loops - both with break",
			input: `while True do begin
				while True do break;
				break;
			end;`,
			expectWarnings: []string{}, // Both have breaks
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

			errors := analyzer.Errors()

			// Check for expected warnings
			for _, expectedWarning := range tt.expectWarnings {
				found := false
				for _, errMsg := range errors {
					if strings.Contains(errMsg, expectedWarning) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected warning containing '%s', but got errors: %v", expectedWarning, errors)
				}
			}

			// Check that we don't have unexpected warnings
			if len(tt.expectWarnings) == 0 {
				for _, errMsg := range errors {
					if strings.Contains(errMsg, "Infinite loop") {
						t.Errorf("Unexpected infinite loop warning: %s", errMsg)
					}
				}
			}

			// Compilation errors should not prevent warning detection
			if err != nil && len(tt.expectWarnings) > 0 {
				// It's OK to have errors as long as we got the warnings we expected
			}
		})
	}
}
