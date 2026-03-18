package parser

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func isKnownUnitPrefixName(name string) bool {
	switch ident.Normalize(name) {
	case "system", "internal":
		return true
	default:
		return false
	}
}

func (p *Parser) addParserErrorAt(pos lexer.Position, length int, message, code string) {
	p.errors = append(p.errors, NewParserError(pos, length, message, code))
}

func (p *Parser) parseQualifiedIdentifierAtCurrent() (*ast.Identifier, bool) {
	currentToken := p.cursor.Current()
	parts := []string{currentToken.Literal}
	endToken := currentToken
	cursor := p.cursor

	for cursor.Peek(1).Type == lexer.DOT {
		nextToken := cursor.Peek(2)
		if !p.isMemberNameToken(nextToken.Type) {
			cursor = cursor.Advance()
			cursor = cursor.Advance()
			p.cursor = cursor
			p.addParserErrorAt(nextToken.Pos, nextToken.Length(), "Name expected", ErrExpectedIdent)
			return &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token:  currentToken,
						EndPos: p.endPosFromToken(endToken),
					},
				},
				Value: strings.Join(parts, "."),
			}, false
		}

		cursor = cursor.Advance()
		cursor = cursor.Advance()
		endToken = cursor.Current()
		parts = append(parts, endToken.Literal)
	}

	p.cursor = cursor
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: p.endPosFromToken(endToken),
			},
		},
		Value: strings.Join(parts, "."),
	}, true
}
