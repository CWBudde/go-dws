package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestPropertyIndexGrouping_SurfaceRetention(t *testing.T) {
	for _, source := range []string{"(I + 1)", "(H(I + 1))", "(-I)", "(O is T)", "(O as T)", "(O implements ITest)", "(A[I + 1])"} {
		t.Run(source, func(t *testing.T) {
			p := testParser(source)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			expr := program.Statements[0].(*ast.ExpressionStatement).Expression
			group, ok := expr.(*ast.GroupedExpression)
			if !ok {
				t.Fatalf("got %T, want retained group", expr)
			}
			if group.Pos().Column != 1 || group.End().Column != len(source)+1 {
				t.Fatalf("group positions %v..%v", group.Pos(), group.End())
			}
		})
	}
}

func TestPropertyIndexGrouping_ParenthesizedWriterKind(t *testing.T) {
	p := testParser("type T = class F: Integer; property P: Integer read F write (F); property Q: Integer read F write F; end;")
	program := p.ParseProgram()
	checkParserErrors(t, p)
	props := program.Statements[0].(*ast.ClassDecl).Properties
	if props[0].WriteStmt == nil || props[0].WriteSpec != nil {
		t.Fatal("parenthesized writer lost expression kind")
	}
	if props[1].WriteSpec == nil || props[1].WriteStmt != nil {
		t.Fatal("bare writer lost field kind")
	}
}

// ungroupForAssertion preserves the original operator tree asserted by legacy
// precedence tests, while the production AST retains the source brackets.
func ungroupForAssertion(expr ast.Expression) ast.Expression {
	switch e := expr.(type) {
	case *ast.GroupedExpression:
		return ungroupForAssertion(e.Expression)
	case *ast.BinaryExpression:
		projected := *e
		projected.Left = ungroupForAssertion(e.Left)
		projected.Right = ungroupForAssertion(e.Right)
		return &projected
	case *ast.UnaryExpression:
		projected := *e
		projected.Right = ungroupForAssertion(e.Right)
		return &projected
	case *ast.IsExpression:
		projected := *e
		projected.Left = ungroupForAssertion(e.Left)
		projected.Right = ungroupForAssertion(e.Right)
		return &projected
	case *ast.AsExpression:
		projected := *e
		projected.Left = ungroupForAssertion(e.Left)
		projected.Right = ungroupForAssertion(e.Right)
		return &projected
	case *ast.ImplementsExpression:
		projected := *e
		projected.Left = ungroupForAssertion(e.Left)
		return &projected
	}
	return expr
}

func precedenceStringForAssertion(program *ast.Program) string {
	var result string
	for _, stmt := range program.Statements {
		if expression, ok := stmt.(*ast.ExpressionStatement); ok {
			result += ungroupForAssertion(expression.Expression).String()
		} else {
			result += stmt.String()
		}
	}
	return result
}

func typeStringForAssertion(typ ast.TypeExpression) string {
	if annotation, ok := typ.(*ast.TypeAnnotation); ok && annotation.InlineType != nil {
		return typeStringForAssertion(annotation.InlineType)
	}
	if array, ok := typ.(*ast.ArrayTypeNode); ok {
		projected := *array
		projected.LowBound = ungroupForAssertion(array.LowBound)
		projected.HighBound = ungroupForAssertion(array.HighBound)
		return projected.String()
	}
	if array, ok := typ.(*ast.ArrayTypeAnnotation); ok {
		projected := *array
		projected.LowBound = ungroupForAssertion(array.LowBound)
		projected.HighBound = ungroupForAssertion(array.HighBound)
		return projected.String()
	}
	return typ.String()
}

func TestPropertyIndexGrouping_UnreadBinaryTailRetention(t *testing.T) {
	p := testParser("PrintLn(O.P[I + ]);")
	index := retainedUnreadPropertyIndex(t, p)
	binary, ok := index.Index.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("index %T", index.Index)
	}
	left, ok := binary.Left.(*ast.Identifier)
	if !ok || left.Value != "I" || binary.Token.Pos.Column != 15 {
		t.Fatalf("binary prefix %v", binary)
	}
	invalid, ok := binary.Right.(*ast.InvalidExpression)
	if !ok || invalid.Pos().Column != 17 {
		t.Fatalf("invalid RHS %#v", binary.Right)
	}
	if index.MissingClosePos.IsValid() || index.End().Column != 18 {
		t.Fatalf("bracket span %#v", index)
	}
	errs := p.Errors()
	if len(errs) != 1 || errs[0].Code != ErrInvalidExpression || !errs[0].Stop || errs[0].DeferredIndex != index || errs[0].Pos.Column != 17 {
		t.Fatalf("deferred identity %#v", errs)
	}
}

func retainedUnreadPropertyIndex(t *testing.T, p *Parser) *ast.IndexExpression {
	t.Helper()
	program := p.ParseProgram()
	if len(program.Statements) != 1 {
		t.Fatalf("retained statements %d, errors %v", len(program.Statements), p.Errors())
	}
	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("retained statement %T", program.Statements[0])
	}
	call, ok := statement.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("retained call %T", statement.Expression)
	}
	if len(call.Arguments) != 1 {
		t.Fatalf("retained arguments %d", len(call.Arguments))
	}
	index, ok := call.Arguments[0].(*ast.IndexExpression)
	if !ok {
		t.Fatalf("argument %T", call.Arguments[0])
	}
	return index
}
