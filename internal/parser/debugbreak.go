package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseDebugBreak reads the intrinsic without reading an argument expression.
// PRE: cursor is the unqualified DebugBreak identifier.
// POST: cursor is the identifier, optional ')', or '(' on a compiler stop.
func (p *Parser) parseDebugBreak() ast.Expression {
	node := &ast.DebugBreakExpression{BaseNode: ast.BaseNode{Token: p.cursor.Current()}}
	if p.cursor.Peek(1).Type == lexer.LPAREN {
		p.cursor = p.cursor.Advance()
		if p.cursor.Peek(1).Type == lexer.RPAREN {
			p.cursor = p.cursor.Advance()
		} else {
			p.addExpectedStop(lexer.RPAREN)
			node.Incomplete = true
		}
	}
	node.EndPos = p.endPosFromToken(p.cursor.Current())
	node.NextTokenPos = p.anchorFor(p.cursor.Peek(1)).Pos
	return node
}
