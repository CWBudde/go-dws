package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestClassDeclarationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing class name",
			"type = class end;",
		},
		{
			"missing equals sign",
			"type TPoint class end;",
		},
		{
			"missing class keyword",
			"type TPoint = end;",
		},
		{
			"missing end keyword",
			"type TPoint = class X: Integer;",
		},
		{
			"missing semicolon after end",
			"type TPoint = class end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected parser errors but got none")
			}
		})
	}
}

func TestFieldDeclarationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing type in field",
			"type TPoint = class X:; end;",
		},
		{
			"missing semicolon after field",
			"type TPoint = class X: Integer end;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected parser errors but got none")
			}
		})
	}
}

func TestMemberAccessErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing identifier after dot",
			"obj.;",
		},
		{
			"number after dot",
			"obj.123;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected parser errors but got none")
			}
		})
	}
}
