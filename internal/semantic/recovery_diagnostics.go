package semantic

import "github.com/cwbudde/go-dws/pkg/ast"

// containsParserRecovery reports whether expr holds an *ast.InvalidExpression, the
// placeholder the parser leaves where it reported "Expression expected". DWScript stops
// compiling at that point, so no semantic diagnostic about the enclosing expression
// may follow it.
func containsParserRecovery(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if _, ok := n.(*ast.InvalidExpression); ok {
			found = true
		}
		return !found
	})
	return found
}

// reportUnconsumedPropertyValue mirrors how DWScript recovers from assigning to a
// read-only property. Upstream consumes the assignment operator, reports the read-only
// property and ends the statement there, leaving the value unread. The main program's
// statement loop then finds that value where it expects a ';' and stops with
// `Unexpected "<token>"` at the value's first token.
//
// Only a top-level main-program statement gets the follow-up: inside a block the loop
// words it differently. The token is named in upstream's vocabulary only for the kinds
// recognisable from the AST; anything else gets no follow-up rather than a guessed one.
func (a *Analyzer) reportUnconsumedPropertyValue(stmt *ast.AssignmentStatement) {
	if stmt == nil || stmt.Value == nil || a.mainStatement != ast.Statement(stmt) {
		return
	}
	pos := stmt.Value.Pos()
	name := ""
	ast.Inspect(stmt.Value, func(n ast.Node) bool {
		if name != "" || n == nil {
			return false
		}
		if n.Pos() != pos {
			return true
		}
		switch lit := n.(type) {
		case *ast.IntegerLiteral:
			name = "Integer Literal"
		case *ast.FloatLiteral:
			name = "Float Literal"
		case *ast.StringLiteral:
			name = "UnicodeString Literal"
		case *ast.Identifier:
			name = "name"
		case *ast.BooleanLiteral:
			name = "False"
			if lit.Value {
				name = "True"
			}
		case *ast.NilLiteral:
			name = "nil"
		}
		return name == ""
	})
	if name == "" {
		return
	}
	a.addError("Syntax Error: Unexpected \"%s\" [line: %d, column: %d]", name, pos.Line, pos.Column)
}
