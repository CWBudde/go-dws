package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

func TestParseConstDeclaration(t *testing.T) {
	input := `const PI = 3.14;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "PI" {
		t.Errorf("stmt.Name.Value not 'PI'. got=%s", stmt.Name.Value)
	}

	if stmt.Type != nil {
		t.Errorf("stmt.Type should be nil for untyped const. got=%v", stmt.Type)
	}

	floatLit, ok := stmt.Value.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.FloatLiteral. got=%T", stmt.Value)
	}

	if floatLit.Value != 3.14 {
		t.Errorf("floatLit.Value not 3.14. got=%f", floatLit.Value)
	}
}

func TestParseConstDeclarationTyped(t *testing.T) {
	input := `const MAX_USERS: Integer = 1000;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "MAX_USERS" {
		t.Errorf("stmt.Name.Value not 'MAX_USERS'. got=%s", stmt.Name.Value)
	}

	if stmt.Type == nil {
		t.Fatal("stmt.Type should not be nil for typed const")
	}

	if stmt.Type.Name != "Integer" {
		t.Errorf("stmt.Type.Name not 'Integer'. got=%s", stmt.Type.Name)
	}

	intLit, ok := stmt.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.IntegerLiteral. got=%T", stmt.Value)
	}

	if intLit.Value != 1000 {
		t.Errorf("intLit.Value not 1000. got=%d", intLit.Value)
	}
}

func TestParseConstDeclarationString(t *testing.T) {
	input := `const APP_NAME = 'MyApp';`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "APP_NAME" {
		t.Errorf("stmt.Name.Value not 'APP_NAME'. got=%s", stmt.Name.Value)
	}

	stringLit, ok := stmt.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.StringLiteral. got=%T", stmt.Value)
	}

	if stringLit.Value != "MyApp" {
		t.Errorf("stringLit.Value not 'MyApp'. got=%s", stringLit.Value)
	}
}

func TestParseMultipleConstDeclarations(t *testing.T) {
	input := `
const PI = 3.14;
const MAX = 100;
const NAME = 'test';
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedName  string
		expectedValue interface{}
	}{
		{"PI", 3.14},
		{"MAX", int64(100)},
		{"NAME", "test"},
	}

	for i, tt := range tests {
		stmt, ok := program.Statements[i].(*ast.ConstDecl)
		if !ok {
			t.Fatalf("program.Statements[%d] is not *ast.ConstDecl. got=%T",
				i, program.Statements[i])
		}

		if stmt.Name.Value != tt.expectedName {
			t.Errorf("stmt.Name.Value not '%s'. got=%s", tt.expectedName, stmt.Name.Value)
		}

		switch v := tt.expectedValue.(type) {
		case float64:
			floatLit, ok := stmt.Value.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.FloatLiteral. got=%T", stmt.Value)
			}
			if floatLit.Value != v {
				t.Errorf("floatLit.Value not %f. got=%f", v, floatLit.Value)
			}
		case int64:
			intLit, ok := stmt.Value.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.IntegerLiteral. got=%T", stmt.Value)
			}
			if intLit.Value != v {
				t.Errorf("intLit.Value not %d. got=%d", v, intLit.Value)
			}
		case string:
			stringLit, ok := stmt.Value.(*ast.StringLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.StringLiteral. got=%T", stmt.Value)
			}
			if stringLit.Value != v {
				t.Errorf("stringLit.Value not %s. got=%s", v, stringLit.Value)
			}
		}
	}
}

func TestParseConstDeclarationErrors(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{"const PI;", "expected '=' after const name"},
		{"const PI =;", "no prefix parse function for SEMICOLON found"},
		{"const = 3.14;", "expected next token to be IDENT"},
		{"const PI: = 3.14;", "expected type identifier after ':' in const declaration"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		p.ParseProgram()

		if len(p.errors) == 0 {
			t.Errorf("expected parser error for input %q, but got none", tt.input)
			continue
		}

		found := false
		for _, err := range p.errors {
			if strings.Contains(err, tt.expectedError) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected error containing %q, got errors: %v", tt.expectedError, p.errors)
		}
	}
}
