package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestVarDeclarationTypeInference(t *testing.T) {
	t.Helper()

	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectInferred bool
		expectType     string
		expectError    string
		assertValue    func(t *testing.T, expr ast.Expression)
	}{
		{
			name:           "integer inference with equals",
			input:          "var x = 42;",
			expectedName:   "x",
			expectInferred: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testIntegerLiteral(t, expr, 42) {
					t.Fatalf("expected integer literal 42")
				}
			},
		},
		{
			name:           "string inference with equals",
			input:          `var s = "hello";`,
			expectedName:   "s",
			expectInferred: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				str, ok := expr.(*ast.StringLiteral)
				if !ok {
					t.Fatalf("expected *ast.StringLiteral, got %T", expr)
				}
				if str.Value != "hello" {
					t.Fatalf("expected string literal value %q, got %q", "hello", str.Value)
				}
			},
		},
		{
			name:           "float inference with equals",
			input:          "var f = 3.14;",
			expectedName:   "f",
			expectInferred: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				floatLit, ok := expr.(*ast.FloatLiteral)
				if !ok {
					t.Fatalf("expected *ast.FloatLiteral, got %T", expr)
				}
				if floatLit.Value != 3.14 {
					t.Fatalf("expected float literal value 3.14, got %v", floatLit.Value)
				}
			},
		},
		{
			name:           "array literal inference",
			input:          "var arr = [1, 2, 3];",
			expectedName:   "arr",
			expectInferred: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				arrayLit, ok := expr.(*ast.ArrayLiteralExpression)
				if !ok {
					t.Fatalf("expected *ast.ArrayLiteralExpression, got %T", expr)
				}
				if len(arrayLit.Elements) != 3 {
					t.Fatalf("expected 3 array elements, got %d", len(arrayLit.Elements))
				}
				testIntegerLiteral(t, arrayLit.Elements[0], 1)
				testIntegerLiteral(t, arrayLit.Elements[1], 2)
				testIntegerLiteral(t, arrayLit.Elements[2], 3)
			},
		},
		{
			name:           "record literal inference",
			input:          "var rec = (x: 10; y: 20);",
			expectedName:   "rec",
			expectInferred: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				recordLit, ok := expr.(*ast.RecordLiteralExpression)
				if !ok {
					t.Fatalf("expected *ast.RecordLiteralExpression, got %T", expr)
				}
				if recordLit.TypeName != nil {
					t.Fatalf("expected anonymous record literal, got %s", recordLit.TypeName.String())
				}
				if len(recordLit.Fields) != 2 {
					t.Fatalf("expected 2 record fields, got %d", len(recordLit.Fields))
				}
				if recordLit.Fields[0].Name.Value != "x" {
					t.Fatalf("expected first field name x, got %s", recordLit.Fields[0].Name.Value)
				}
				if recordLit.Fields[1].Name.Value != "y" {
					t.Fatalf("expected second field name y, got %s", recordLit.Fields[1].Name.Value)
				}
				testIntegerLiteral(t, recordLit.Fields[0].Value, 10)
				testIntegerLiteral(t, recordLit.Fields[1].Value, 20)
			},
		},
		{
			name:           "inference with assign operator",
			input:          "var i := 21;",
			expectedName:   "i",
			expectInferred: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testIntegerLiteral(t, expr, 21) {
					t.Fatalf("expected integer literal 21")
				}
			},
		},
		{
			name:        "missing type and initializer error",
			input:       "var x;",
			expectError: "variable declaration requires a type or initializer",
		},
		{
			name:         "explicit type still supported",
			input:        "var x: Integer := 42;",
			expectedName: "x",
			expectType:   "Integer",
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testIntegerLiteral(t, expr, 42) {
					t.Fatalf("expected integer literal 42")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.expectError != "" {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected parser error containing %q but got none", tt.expectError)
				}
				var found bool
				for _, err := range p.Errors() {
					if strings.Contains(err, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected parser error containing %q, got %v", tt.expectError, p.Errors())
				}
				return
			}

			checkParserErrors(t, p)

			if program == nil {
				t.Fatal("ParseProgram returned nil")
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("expected VarDeclStatement, got %T", program.Statements[0])
			}

			if len(stmt.Names) != 1 {
				t.Fatalf("expected 1 variable name, got %d", len(stmt.Names))
			}

			if tt.expectedName != "" && stmt.Names[0].Value != tt.expectedName {
				t.Fatalf("expected variable name %q, got %q", tt.expectedName, stmt.Names[0].Value)
			}

			if tt.expectType != "" {
				if stmt.Type == nil {
					t.Fatalf("expected type annotation %q, got nil", tt.expectType)
				}
				if stmt.Type.String() != tt.expectType {
					t.Fatalf("expected type %q, got %q", tt.expectType, stmt.Type.String())
				}
				if stmt.Inferred {
					t.Fatalf("expected inferred=false for explicit type")
				}
			} else {
				if stmt.Type != nil {
					t.Fatalf("expected no type annotation, got %s", stmt.Type.String())
				}
				if stmt.Inferred != tt.expectInferred {
					t.Fatalf("expected inferred=%v, got %v", tt.expectInferred, stmt.Inferred)
				}
			}

			if tt.assertValue != nil {
				if stmt.Value == nil {
					t.Fatalf("expected initializer expression, got nil")
				}
				tt.assertValue(t, stmt.Value)
			}
		})
	}
}
