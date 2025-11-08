package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestVarDeclarationTypeInference(t *testing.T) {
	t.Helper()

	tests := []struct {
		assertValue    func(t *testing.T, expr ast.Expression)
		name           string
		input          string
		expectedName   string
		expectType     string
		expectError    string
		expectInferred bool
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
					if strings.Contains(err.Message, tt.expectError) {
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

func TestVarDeclarations(t *testing.T) {
	input := `
var x: Integer;
var y := 5;
var s: String := 'hello';
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		assertValue func(*testing.T, ast.Expression)
		name        string
		expectedVar string
		expectedTyp string
		expectValue bool
	}{
		{
			name:        "typed declaration without initializer",
			expectedVar: "x",
			expectedTyp: "Integer",
			expectValue: false,
		},
		{
			name:        "inferred integer declaration",
			expectedVar: "y",
			expectedTyp: "",
			expectValue: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testIntegerLiteral(t, expr, 5) {
					t.Fatalf("value is not expected integer literal")
				}
			},
		},
		{
			name:        "string declaration with type",
			expectedVar: "s",
			expectedTyp: "String",
			expectValue: true,
			assertValue: func(t *testing.T, expr ast.Expression) {
				if !testStringLiteralExpression(t, expr, "hello") {
					t.Fatalf("value is not expected string literal")
				}
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, ok := program.Statements[i].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not ast.VarDeclStatement. got=%T", program.Statements[i])
			}

			if len(stmt.Names) == 0 || stmt.Names[0].Value != tt.expectedVar {
				if len(stmt.Names) == 0 {
					t.Errorf("stmt.Names is empty, want %q", tt.expectedVar)
				} else {
					t.Errorf("stmt.Names[0].Value = %q, want %q", stmt.Names[0].Value, tt.expectedVar)
				}
			}

			if (stmt.Type == nil && tt.expectedTyp != "") || (stmt.Type != nil && stmt.Type.Name != tt.expectedTyp) {
				t.Errorf("stmt.Type = %q, want %q", stmt.Type, tt.expectedTyp)
			}

			if tt.expectValue {
				if stmt.Value == nil {
					t.Fatalf("expected initialization expression")
				}
				tt.assertValue(t, stmt.Value)
			} else if stmt.Value != nil {
				t.Fatalf("expected no initialization, got %T", stmt.Value)
			}
		})
	}
}

// TestExternalVarParsing tests parsing of external variable declarations.
//
//	var x: Integer; external;
//	var y: String; external 'externalName';
func TestExternalVarParsing(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedVar  string
		expectedType string
		externalName string
		isExternal   bool
		expectError  bool
	}{
		{
			name:         "external variable without custom name",
			input:        "var x: Integer external;",
			expectedVar:  "x",
			expectedType: "Integer",
			isExternal:   true,
			externalName: "",
		},
		{
			name:         "external variable with custom name",
			input:        "var y: String external 'customName';",
			expectedVar:  "y",
			expectedType: "String",
			isExternal:   true,
			externalName: "customName",
		},
		{
			name:         "regular variable (not external)",
			input:        "var z: Float;",
			expectedVar:  "z",
			expectedType: "Float",
			isExternal:   false,
			externalName: "",
		},
		{
			name:         "regular variable with initialization",
			input:        "var w: Integer := 42;",
			expectedVar:  "w",
			expectedType: "Integer",
			isExternal:   false,
			externalName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.expectError {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected parser error, got none")
				}
				return
			}

			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not ast.VarDeclStatement. got=%T", program.Statements[0])
			}

			if len(stmt.Names) == 0 || stmt.Names[0].Value != tt.expectedVar {
				if len(stmt.Names) == 0 {
					t.Errorf("stmt.Names is empty, want %q", tt.expectedVar)
				} else {
					t.Errorf("stmt.Names[0].Value = %q, want %q", stmt.Names[0].Value, tt.expectedVar)
				}
			}

			if stmt.Type == nil || stmt.Type.Name != tt.expectedType {
				var gotType string
				if stmt.Type != nil {
					gotType = stmt.Type.Name
				}
				t.Errorf("stmt.Type.Name = %q, want %q", gotType, tt.expectedType)
			}

			if stmt.IsExternal != tt.isExternal {
				t.Errorf("stmt.IsExternal = %v, want %v", stmt.IsExternal, tt.isExternal)
			}

			if stmt.ExternalName != tt.externalName {
				t.Errorf("stmt.ExternalName = %q, want %q", stmt.ExternalName, tt.externalName)
			}
		})
	}
}

// TestMultiIdentifierVarDeclarations tests parsing of multi-identifier variable declarations.
// DWScript allows comma-separated variable names like `var a, b, c: Integer;`.
func TestMultiIdentifierVarDeclarations(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedType  string
		errorContains string
		expectedNames []string
		expectError   bool
	}{
		{
			name:          "two variables",
			input:         "var x, y: Integer;",
			expectedNames: []string{"x", "y"},
			expectedType:  "Integer",
		},
		{
			name:          "three variables",
			input:         "var a, b, c: String;",
			expectedNames: []string{"a", "b", "c"},
			expectedType:  "String",
		},
		{
			name:          "many variables",
			input:         "var i, j, k, l, m: Integer;",
			expectedNames: []string{"i", "j", "k", "l", "m"},
			expectedType:  "Integer",
		},
		{
			name:          "reject initializer with multiple names",
			input:         "var x, y: Integer := 42;",
			expectError:   true,
			errorContains: "cannot use initializer with multiple variable names",
		},
		{
			name:          "reject initializer without type",
			input:         "var a, b := 5;",
			expectError:   true,
			errorContains: "cannot use initializer with multiple variable names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()

			if tt.expectError {
				if len(p.Errors()) == 0 {
					t.Fatalf("expected error containing %q, got no errors", tt.errorContains)
				}
				found := false
				for _, err := range p.Errors() {
					if strings.Contains(err.Message, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got %v", tt.errorContains, p.Errors())
				}
				return
			}

			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not ast.VarDeclStatement. got=%T", program.Statements[0])
			}

			if len(stmt.Names) != len(tt.expectedNames) {
				t.Fatalf("wrong number of names. got=%d, want=%d", len(stmt.Names), len(tt.expectedNames))
			}

			for i, expectedName := range tt.expectedNames {
				if stmt.Names[i].Value != expectedName {
					t.Errorf("name[%d] = %q, want %q", i, stmt.Names[i].Value, expectedName)
				}
			}

			if stmt.Type == nil || stmt.Type.Name != tt.expectedType {
				var typeName string
				if stmt.Type != nil {
					typeName = stmt.Type.Name
				}
				t.Errorf("stmt.Type.Name = %q, want %q", typeName, tt.expectedType)
			}

			// Test String() method for multi-names
			expectedStr := "var " + strings.Join(tt.expectedNames, ", ") + ": " + tt.expectedType
			if stmt.String() != expectedStr {
				t.Errorf("stmt.String() = %q, want %q", stmt.String(), expectedStr)
			}
		})
	}
}

// TestAssignmentStatements tests parsing of assignment statements.
