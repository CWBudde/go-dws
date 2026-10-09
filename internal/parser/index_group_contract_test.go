package parser_test

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Consumers distinguish declared property arguments from result indexing by
// bracket groups, even though comma indices are desugared into nested nodes.
func TestParseIndexGroup_Boundaries(t *testing.T) {
	p := parser.New(lexer.New("a[1,\n 2][3];"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 || len(program.Statements) != 1 {
		t.Fatalf("statements=%d, errors=%v", len(program.Statements), p.Errors())
	}
	outer := contractIndex(t, contractExpression(t, program.Statements[0]))
	comma := contractIndex(t, outer.Left)
	first := contractIndex(t, comma.Left)
	for _, want := range []struct {
		node                *ast.IndexExpression
		line, column, comma int
		value               int64
	}{{outer, 2, 4, 0, 3}, {comma, 1, 2, 4, 2}, {first, 1, 2, 0, 1}} {
		if want.node.Token.Pos.Line != want.line || want.node.Token.Pos.Column != want.column || want.node.CommaPos.Column != want.comma {
			t.Errorf("bracket=%v, comma=%v; want %d:%d, comma column %d", want.node.Token.Pos, want.node.CommaPos, want.line, want.column, want.comma)
		}
		value, ok := want.node.Index.(*ast.IntegerLiteral)
		if !ok || value.Value != want.value || want.node.Empty || want.node.MissingClosePos.IsValid() {
			t.Errorf("index=%#v, empty=%v, missing=%v", want.node.Index, want.node.Empty, want.node.MissingClosePos)
		}
	}
	if first.Left.TokenLiteral() != "a" || outer.End().Line != 2 || outer.End().Column != 7 {
		t.Errorf("base=%s, end=%v; want a and 2:7", first.Left.TokenLiteral(), outer.End())
	}
}

// Empty brackets must retain their own node and closing-token span. The parser
// cannot classify a property, so its provisional stop must identify that node
// while retaining following syntax for checked semantic resolution.
func TestParseIndexGroup_EmptyDeferredIdentity(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		bracket, end int
	}{
		{"ordinary", "a[]; tail[2];", 2, 4},
		{"inherited compatibility", "inherited Prop()[]; tail[2];", 17, 19},
		{"result indexing", "a[1][]; tail[2];", 5, 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.source))
			program := p.ParseProgram()
			if len(program.Statements) != 2 || len(p.Errors()) != 1 {
				t.Fatalf("statements=%d, errors=%v", len(program.Statements), p.Errors())
			}
			node := contractIndex(t, contractExpression(t, program.Statements[0]))
			err := p.Errors()[0]
			assertDeferredIndexDiagnostic(t, node, err, "Expression expected", parser.ErrInvalidExpression, true)
			assertEmptyIndexGroup(t, node, err, tt.bracket, tt.end)
			tail := contractIndex(t, contractExpression(t, program.Statements[1]))
			if tail.Left.TokenLiteral() != "tail" || tail.Empty || tail.MissingClosePos.IsValid() {
				t.Fatalf("later bracket group was damaged: %#v", tail)
			}
		})
	}
}

func TestParseIndexGroup_ConsecutiveEmptyGroupsHaveSeparateDiagnostics(t *testing.T) {
	p := parser.New(lexer.New("a[][];"))
	program := p.ParseProgram()
	if len(program.Statements) != 1 || len(p.Errors()) != 2 {
		t.Fatalf("statements=%d, errors=%v", len(program.Statements), p.Errors())
	}
	outer := contractIndex(t, contractExpression(t, program.Statements[0]))
	inner := contractIndex(t, outer.Left)
	for i, node := range []*ast.IndexExpression{inner, outer} {
		if !node.Empty || node.CommaPos.IsValid() || p.Errors()[i].DeferredIndex != node || p.Errors()[i].Pos.Column != 3+2*i {
			t.Errorf("group %d has mismatched identity/position: node=%#v, error=%#v", i, node, p.Errors()[i])
		}
	}
}

// Ordinary missing ']' is recoverable. Only the final node in that bracket
// group owns MissingClosePos and the provisional diagnostic; a later group is
// independent. Semantic property resolution may turn this error into a stop.
func TestParseIndexGroup_UnfinishedRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source            string
		bracket, comma, missing int
	}{
		{"comma group", "a[1,2; tail[3];", 2, 4, 6},
		{"consecutive group", "a[1][2; tail[3];", 5, 0, 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.source))
			program := p.ParseProgram()
			if len(program.Statements) != 2 || len(p.Errors()) != 1 {
				t.Fatalf("statements=%d, errors=%v", len(program.Statements), p.Errors())
			}
			node := contractIndex(t, contractExpression(t, program.Statements[0]))
			err := p.Errors()[0]
			assertDeferredIndexDiagnostic(t, node, err, `"]" expected`, parser.ErrMissingRBracket, false)
			assertUnfinishedIndexGroup(t, node, err, tt.bracket, tt.comma, tt.missing)
			if tail := contractIndex(t, contractExpression(t, program.Statements[1])); tail.Left.TokenLiteral() != "tail" || tail.MissingClosePos.IsValid() {
				t.Fatalf("recovery damaged following statement: %#v", tail)
			}
		})
	}
}

// A missing child expression is a definite stop rather than a deferred closing
// delimiter. Keep the reached comma-list children without reporting later noise.
func TestParseIndexGroup_MissingChildStopsDiagnostics(t *testing.T) {
	p := parser.New(lexer.New("a[1,]; var later := ;"))
	program := p.ParseProgram()
	if len(p.Errors()) != 1 {
		t.Fatalf("later recovery diagnostics escaped the stop: %v", p.Errors())
	}
	err := p.Errors()[0]
	if err.Message != "Expression expected" || !err.Stop || err.DeferredIndex != nil || err.Pos.Column != 5 {
		t.Fatalf("unexpected definite stop: %#v", err)
	}
	if len(program.Statements) == 0 {
		t.Fatal("reached index children were discarded")
	}
	node := contractIndex(t, contractExpression(t, program.Statements[0]))
	first := contractIndex(t, node.Left)
	if _, ok := node.Index.(*ast.InvalidExpression); !ok || first.Index.TokenLiteral() != "1" || node.Empty || node.MissingClosePos.IsValid() || node.End().Column != 6 {
		t.Fatalf("truncated comma group lost reached children/span: %#v", node)
	}
}

func assertDeferredIndexDiagnostic(t *testing.T, node *ast.IndexExpression, err *parser.ParserError, message, code string, stop bool) {
	t.Helper()
	if err.Message != message || err.Code != code || err.Stop != stop || err.DeferredIndex != node || err.DeferredCall != nil {
		t.Fatalf("diagnostic does not identify its bracket group: %#v", err)
	}
}

func assertEmptyIndexGroup(t *testing.T, node *ast.IndexExpression, err *parser.ParserError, bracket, end int) {
	t.Helper()
	invalid, ok := node.Index.(*ast.InvalidExpression)
	if !node.Empty || !ok || invalid.Token.Type != lexer.RBRACK || node.MissingClosePos.IsValid() {
		t.Fatalf("empty group lost its recovery child: %#v", node)
	}
	if node.Token.Pos.Column != bracket || node.End().Column != end || err.Pos.Column != end-1 || err.Pos.Offset != end-2 || err.Length != 1 {
		t.Errorf("bracket=%v, end=%v, error=%#v", node.Token.Pos, node.End(), err)
	}
}

func assertUnfinishedIndexGroup(t *testing.T, node *ast.IndexExpression, err *parser.ParserError, bracket, comma, missing int) {
	t.Helper()
	if node.Empty || node.Token.Pos.Column != bracket || node.CommaPos.Column != comma || node.MissingClosePos != err.Pos || err.Pos.Column != missing || err.Pos.Offset != missing-1 || err.Length != 1 {
		t.Errorf("unfinished group metadata=%#v, error=%#v", node, err)
	}
	previous := contractIndex(t, node.Left)
	if previous.MissingClosePos.IsValid() || previous.Empty || previous.Index.TokenLiteral() != "1" || node.Index.TokenLiteral() != "2" {
		t.Fatalf("reached children or previous bracket were changed: %#v", node)
	}
}

func contractExpression(t *testing.T, statement ast.Statement) ast.Expression {
	t.Helper()
	expression, ok := statement.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement=%T; want expression statement", statement)
	}
	return expression.Expression
}

func contractIndex(t *testing.T, expression ast.Expression) *ast.IndexExpression {
	t.Helper()
	index, ok := expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expression=%T; want index expression", expression)
	}
	return index
}
