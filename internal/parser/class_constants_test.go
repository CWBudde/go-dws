package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestParseClassConstants(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedConstName string
		expectedValue     string
		expectedIsClass   bool
		expectedVis       ast.Visibility
	}{
		{
			name: "simple class constant",
			input: `type TBase = class
				const c1 = 1;
			end;`,
			expectedConstName: "c1",
			expectedValue:     "1",
			expectedIsClass:   false,
			expectedVis:       ast.VisibilityPublic,
		},
		{
			name: "class const with class keyword",
			input: `type TBase = class
				public class const cPublic = 3;
			end;`,
			expectedConstName: "cPublic",
			expectedValue:     "3",
			expectedIsClass:   true,
			expectedVis:       ast.VisibilityPublic,
		},
		{
			name: "private class constant",
			input: `type TBase = class
				private const cPrivate = 1;
			end;`,
			expectedConstName: "cPrivate",
			expectedValue:     "1",
			expectedIsClass:   false,
			expectedVis:       ast.VisibilityPrivate,
		},
		{
			name: "protected class constant",
			input: `type TBase = class
				protected const cProtected = 2;
			end;`,
			expectedConstName: "cProtected",
			expectedValue:     "2",
			expectedIsClass:   false,
			expectedVis:       ast.VisibilityProtected,
		},
		{
			name: "class constant with expression",
			input: `type TBase = class
				const c1 = 1;
				const c2 = c1 + 1;
			end;`,
			expectedConstName: "c2",
			expectedValue:     "(c1 + 1)",
			expectedIsClass:   false,
			expectedVis:       ast.VisibilityPublic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			if program == nil {
				t.Fatal("ParseProgram() returned nil")
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			classDecl, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok {
				t.Fatalf("statement is not *ast.ClassDecl, got %T", program.Statements[0])
			}

			if len(classDecl.Constants) == 0 {
				t.Fatal("expected at least one constant, got 0")
			}

			// Get the last constant (in case of multiple constants like c1 and c2)
			constant := classDecl.Constants[len(classDecl.Constants)-1]

			if constant.Name.Value != tt.expectedConstName {
				t.Errorf("constant name = %q, want %q", constant.Name.Value, tt.expectedConstName)
			}

			if constant.Value.String() != tt.expectedValue {
				t.Errorf("constant value = %q, want %q", constant.Value.String(), tt.expectedValue)
			}

			if constant.IsClassConst != tt.expectedIsClass {
				t.Errorf("IsClassConst = %v, want %v", constant.IsClassConst, tt.expectedIsClass)
			}

			if constant.Visibility != tt.expectedVis {
				t.Errorf("Visibility = %v, want %v", constant.Visibility, tt.expectedVis)
			}
		})
	}
}

func TestParseClassConstantsWithType(t *testing.T) {
	input := `type TBase = class
		const cTyped: Integer = 42;
	end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	classDecl := program.Statements[0].(*ast.ClassDecl)

	if len(classDecl.Constants) != 1 {
		t.Fatalf("expected 1 constant, got %d", len(classDecl.Constants))
	}

	constant := classDecl.Constants[0]

	if constant.Name.Value != "cTyped" {
		t.Errorf("constant name = %q, want %q", constant.Name.Value, "cTyped")
	}

	if constant.Type == nil {
		t.Fatal("expected type annotation, got nil")
	}
}

func TestParseMultipleClassConstants(t *testing.T) {
	input := `type TBase = class
		private const cPrivate = 1;
		protected const cProtected = 2;
		public class const cPublic = 3;
	end;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	classDecl := program.Statements[0].(*ast.ClassDecl)

	if len(classDecl.Constants) != 3 {
		t.Fatalf("expected 3 constants, got %d", len(classDecl.Constants))
	}

	// Check first constant
	if classDecl.Constants[0].Name.Value != "cPrivate" {
		t.Errorf("first constant name = %q, want %q", classDecl.Constants[0].Name.Value, "cPrivate")
	}
	if classDecl.Constants[0].Visibility != ast.VisibilityPrivate {
		t.Errorf("first constant visibility = %v, want %v", classDecl.Constants[0].Visibility, ast.VisibilityPrivate)
	}

	// Check second constant
	if classDecl.Constants[1].Name.Value != "cProtected" {
		t.Errorf("second constant name = %q, want %q", classDecl.Constants[1].Name.Value, "cProtected")
	}
	if classDecl.Constants[1].Visibility != ast.VisibilityProtected {
		t.Errorf("second constant visibility = %v, want %v", classDecl.Constants[1].Visibility, ast.VisibilityProtected)
	}

	// Check third constant
	if classDecl.Constants[2].Name.Value != "cPublic" {
		t.Errorf("third constant name = %q, want %q", classDecl.Constants[2].Name.Value, "cPublic")
	}
	if classDecl.Constants[2].Visibility != ast.VisibilityPublic {
		t.Errorf("third constant visibility = %v, want %v", classDecl.Constants[2].Visibility, ast.VisibilityPublic)
	}
	if !classDecl.Constants[2].IsClassConst {
		t.Errorf("third constant IsClassConst = false, want true")
	}
}
