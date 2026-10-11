package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParser_TruncatedTryKeepsReachedHandlers(t *testing.T) {
	p := testParser("try\nexcept\non e: Integer do ;\non e: Exception do ;")
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("statements = %d, want retained try", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.TryStatement)
	if !ok || stmt.ExceptClause == nil || len(stmt.ExceptClause.Handlers) != 2 {
		t.Fatalf("reached handlers lost: %#v", program.Statements[0])
	}
	if errs := p.Errors(); len(errs) != 1 || errs[0].Message != "END expected" || !errs[0].Stop {
		t.Fatalf("errors = %v, want END compiler stop", errs)
	}
}

func TestParser_TryBodyStopDoesNotReadHandlers(t *testing.T) {
	p := testParser("try\nPrintLn(1 + );\nexcept\non e: Unread do ;\nend;\nPrintLn(UnreadTail);")
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("statements = %d, want retained try prefix", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.TryStatement)
	if !ok || stmt.TryBlock == nil || !stmt.TryBlock.Truncated || stmt.ExceptClause != nil {
		t.Fatalf("body stop must retain a truncated body without unread handlers: %#v", program.Statements[0])
	}
}

func TestParser_HandlerStopDoesNotReadLaterHandlers(t *testing.T) {
	p := testParser("try\nexcept\non e: Integer do PrintLn(1 + );\non e: Unread do ;\nend;\nPrintLn(UnreadTail);")
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("statements = %d, want retained try prefix", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.TryStatement)
	if !ok || stmt.ExceptClause == nil || len(stmt.ExceptClause.Handlers) != 1 {
		t.Fatalf("handler stop must retain only the reached handler: %#v", program.Statements[0])
	}
}

func TestParser_ClassAncestryStopKeepsReachedEntries(t *testing.T) {
	for _, source := range []string{
		"type TTest = class(TObject, Integer",
		"type TTest = class(TObject, Integer\nprivate\nF: Unread;\nprocedure P;\nend;",
	} {
		t.Run(source, func(t *testing.T) {
			p := testParser(source)
			program := p.ParseProgram()
			if len(program.Statements) != 1 {
				t.Fatalf("statements = %d, want retained class", len(program.Statements))
			}
			decl, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok || decl.Parent == nil || decl.Parent.Value != "TObject" || len(decl.Interfaces) != 1 || decl.Interfaces[0].Value != "Integer" {
				t.Fatalf("reached ancestry lost: %#v", program.Statements[0])
			}
			if len(decl.Fields)+len(decl.Methods)+len(decl.VisibilitySections) != 0 || decl.IsForward {
				t.Fatalf("unread body/forward completion retained: %#v", decl)
			}
			if errs := p.Errors(); len(errs) != 1 || errs[0].Message != `")" expected` || !errs[0].Stop {
				t.Fatalf("errors = %v, want closing parenthesis compiler stop", errs)
			}
		})
	}
}
