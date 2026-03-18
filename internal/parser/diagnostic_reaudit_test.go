package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func parserErrorMessages(p *Parser) []string {
	errors := p.Errors()
	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Message)
	}
	return messages
}

func requireErrorContaining(t *testing.T, messages []string, want string) {
	t.Helper()
	for _, msg := range messages {
		if strings.Contains(msg, want) {
			return
		}
	}
	t.Fatalf("expected an error containing %q, got %v", want, messages)
}

func requireNoErrorContaining(t *testing.T, messages []string, unwanted string) {
	t.Helper()
	for _, msg := range messages {
		if strings.Contains(msg, unwanted) {
			t.Fatalf("did not expect an error containing %q, got %v", unwanted, messages)
		}
	}
}

func TestParserReaudit_GroupedExpressionDiagnosticsStayLocalized(t *testing.T) {
	p := testParser("var x := (, 2); var y := 1;")
	program := p.ParseProgram()
	errors := parserErrorMessages(p)

	requireErrorContaining(t, errors, "Expression expected")
	requireNoErrorContaining(t, errors, "expected ';' after variable declaration")

	if len(program.Statements) != 2 {
		t.Fatalf("expected parser to continue to second statement, got %d statements", len(program.Statements))
	}
}

func TestParserReaudit_CallArgumentDiagnosticsStayLocalized(t *testing.T) {
	p := testParser("var x := Foo(, 2); var y := 1;")
	program := p.ParseProgram()
	errors := parserErrorMessages(p)

	requireErrorContaining(t, errors, "Expression expected")
	requireNoErrorContaining(t, errors, "expected ';' after variable declaration")

	if len(program.Statements) != 2 {
		t.Fatalf("expected parser to continue to second statement, got %d statements", len(program.Statements))
	}
}

func TestParserReaudit_IndexDiagnosticsStayLocalized(t *testing.T) {
	p := testParser("var x := arr[, 2]; var y := 1;")
	program := p.ParseProgram()
	errors := parserErrorMessages(p)

	requireErrorContaining(t, errors, "Expression expected")
	requireNoErrorContaining(t, errors, "expected ';' after variable declaration")
	requireNoErrorContaining(t, errors, "\"]\" expected")

	if len(program.Statements) != 2 {
		t.Fatalf("expected parser to continue to second statement, got %d statements", len(program.Statements))
	}
}

func TestParserReaudit_NestedTypeDiagnosticsStayLocalized(t *testing.T) {
	p := testParser("type TFunc = function(array of , Integer): Boolean; var y := 1;")
	program := p.ParseProgram()
	errors := parserErrorMessages(p)

	requireErrorContaining(t, errors, "expected type expression after 'array of'")
	requireNoErrorContaining(t, errors, "expected ')' after parameter list in function pointer type")
	requireNoErrorContaining(t, errors, "expected ';' after type declaration")

	if len(program.Statements) != 2 {
		t.Fatalf("expected parser to continue to second statement, got %d statements", len(program.Statements))
	}

	if _, ok := program.Statements[1].(*ast.VarDeclStatement); !ok {
		t.Fatalf("expected second statement to be var declaration, got %T", program.Statements[1])
	}
}
