package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestParserWithSemanticAnalysis tests that the parser can optionally run semantic analysis
func TestParserWithSemanticAnalysis(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		errorCount  int
	}{
		{
			name:        "valid program with type checking",
			input:       "var x: Integer := 5; var y: Integer := x + 3;",
			shouldError: false,
		},
		{
			name:        "type mismatch should error",
			input:       "var x: Integer := 'hello';",
			shouldError: true,
			errorCount:  1,
		},
		{
			name:        "undefined variable should error",
			input:       "var x: Integer := y;",
			shouldError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			// Run semantic analysis
			analyzer := semantic.NewAnalyzer()
			_ = analyzer.Analyze(program)
			semanticErrors := analyzer.Errors()

			if tt.shouldError {
				if len(semanticErrors) == 0 {
					t.Fatalf("expected semantic errors but got none")
				}
				if tt.errorCount > 0 && len(semanticErrors) != tt.errorCount {
					t.Errorf("expected %d semantic errors, got %d: %v",
						tt.errorCount, len(semanticErrors), semanticErrors)
				}
			} else {
				if len(semanticErrors) > 0 {
					t.Fatalf("unexpected semantic errors: %v", semanticErrors)
				}
			}

			if program == nil {
				t.Fatal("ParseProgram() returned nil")
			}
		})
	}
}

// TestParserWithoutSemanticAnalysis ensures semantic analysis is optional
func TestParserWithoutSemanticAnalysis(t *testing.T) {
	input := "var x: Integer := 'hello';" // Type mismatch

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// Should have no parser errors (only semantic error)
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Without running semantic analysis, should have no errors stored
	semanticErrors := p.SemanticErrors()
	if len(semanticErrors) > 0 {
		t.Fatalf("unexpected semantic errors when analysis not run: %v", semanticErrors)
	}

	if program == nil {
		t.Fatal("ParseProgram() returned nil")
	}
}
