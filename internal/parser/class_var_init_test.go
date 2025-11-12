package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestParseClassVarInitialization(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedName  string
		expectedType  string
		expectedInit  string
		expectedIsVar bool
	}{
		{
			name: "class var with type and initialization",
			input: `type TBase = class
				class var Count: Integer := 42;
			end;`,
			expectedName:  "Count",
			expectedType:  "Integer",
			expectedInit:  "42",
			expectedIsVar: true,
		},
		{
			name: "class var with type inference",
			input: `type TBase = class
				class var Test := 1;
			end;`,
			expectedName:  "Test",
			expectedType:  "",
			expectedInit:  "1",
			expectedIsVar: true,
		},
		{
			name: "class var with expression initialization",
			input: `type TBase = class
				class const C = 10;
				class var Test := 5 + C;
			end;`,
			expectedName:  "Test",
			expectedType:  "",
			expectedInit:  "(5 + C)",
			expectedIsVar: true,
		},
		{
			name: "class var with string initialization",
			input: `type TBase = class
				class var Name: String := 'Hello';
			end;`,
			expectedName:  "Name",
			expectedType:  "String",
			expectedInit:  "\"Hello\"", // String literals include quotes in String()
			expectedIsVar: true,
		},
		{
			name: "class var without initialization",
			input: `type TBase = class
				class var Counter: Integer;
			end;`,
			expectedName:  "Counter",
			expectedType:  "Integer",
			expectedInit:  "",
			expectedIsVar: true,
		},
		{
			name: "private class var with initialization",
			input: `type TBase = class
				private class var Secret: Integer := 123;
			end;`,
			expectedName:  "Secret",
			expectedType:  "Integer",
			expectedInit:  "123",
			expectedIsVar: true,
		},
		{
			name: "class var with float initialization",
			input: `type TBase = class
				class var Pi: Float := 3.14;
			end;`,
			expectedName:  "Pi",
			expectedType:  "Float",
			expectedInit:  "3.14",
			expectedIsVar: true,
		},
		{
			name: "class var with boolean initialization",
			input: `type TBase = class
				class var Enabled := True;
			end;`,
			expectedName:  "Enabled",
			expectedType:  "",
			expectedInit:  "True",
			expectedIsVar: true,
		},
		{
			name: "class var with negative number",
			input: `type TBase = class
				class var Offset: Integer := -10;
			end;`,
			expectedName:  "Offset",
			expectedType:  "Integer",
			expectedInit:  "(-10)",
			expectedIsVar: true,
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
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			classDecl, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok {
				t.Fatalf("statement is not *ast.ClassDecl. got=%T", program.Statements[0])
			}

			// Find the class var field
			var field *ast.FieldDecl
			for _, f := range classDecl.Fields {
				if f.Name.Value == tt.expectedName {
					field = f
					break
				}
			}

			if field == nil {
				t.Fatalf("class var '%s' not found in class", tt.expectedName)
			}

			// Check IsClassVar flag
			if field.IsClassVar != tt.expectedIsVar {
				t.Errorf("field.IsClassVar = %v, want %v", field.IsClassVar, tt.expectedIsVar)
			}

			// Check name
			if field.Name.Value != tt.expectedName {
				t.Errorf("field.Name.Value = %q, want %q", field.Name.Value, tt.expectedName)
			}

			// Check type (may be nil for type inference)
			if tt.expectedType != "" {
				if field.Type == nil {
					t.Errorf("field.Type is nil, expected %q", tt.expectedType)
				} else if field.Type.String() != tt.expectedType {
					t.Errorf("field.Type = %q, want %q", field.Type.String(), tt.expectedType)
				}
			} else {
				if field.Type != nil {
					t.Errorf("field.Type = %q, expected nil for type inference", field.Type.String())
				}
			}

			// Check initialization value
			if tt.expectedInit != "" {
				if field.InitValue == nil {
					t.Errorf("field.InitValue is nil, expected %q", tt.expectedInit)
				} else if field.InitValue.String() != tt.expectedInit {
					t.Errorf("field.InitValue = %q, want %q", field.InitValue.String(), tt.expectedInit)
				}
			} else {
				if field.InitValue != nil {
					t.Errorf("field.InitValue = %q, expected nil", field.InitValue.String())
				}
			}
		})
	}
}

func TestClassVarInitializationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "comma-separated fields with initialization",
			input: `type TBase = class
				class var A, B: Integer := 42;
			end;`,
			expectedError: "initialization not allowed for comma-separated field declarations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Fatal("expected parser error, got none")
			}

			foundError := false
			for _, err := range p.Errors() {
				if contains(err.Error(), tt.expectedError) {
					foundError = true
					break
				}
			}

			if !foundError {
				t.Errorf("expected error containing %q, got %v", tt.expectedError, p.Errors())
			}
		})
	}
}
