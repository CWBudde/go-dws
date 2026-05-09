package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Syntax: with <name> [":" <type>] (":=" | "=") <expr> [, ...] do <statement>
// PRE: cursor is on WITH token
// POST: cursor is on last token of body statement
func (p *Parser) parseWithStatement() *ast.WithStatement {
	builder := p.StartNode()

	withToken := p.cursor.Current()
	stmt := &ast.WithStatement{
		BaseNode:     ast.BaseNode{Token: withToken},
		Declarations: []*ast.VarDeclStatement{},
	}

	p.pushBlockContext("with", withToken.Pos)
	defer p.popBlockContext()

	p.cursor = p.cursor.Advance()
	for {
		decl := p.parseWithDeclaration()
		if decl == nil {
			p.synchronize([]lexer.TokenType{lexer.DO, lexer.SEMICOLON, lexer.END})
			return nil
		}
		stmt.Declarations = append(stmt.Declarations, decl)

		nextToken := p.cursor.Peek(1)
		if nextToken.Type != lexer.COMMA {
			break
		}
		p.cursor = p.cursor.Advance() // move to comma
		p.cursor = p.cursor.Advance() // move to next declaration name
	}

	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.DO {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingDo).
			WithMessage("expected 'do' after with declarations").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("'do'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add 'do' before the with statement body").
			WithParsePhase("with statement").
			Build()
		p.addStructuredError(err)
		return nil
	}

	p.cursor = p.cursor.Advance() // move to do
	p.cursor = p.cursor.Advance() // move to body
	stmt.Body = p.parseStatement()
	if isNilStatement(stmt.Body) {
		currentToken := p.cursor.Current()
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected statement after 'do'").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithSuggestion("add a statement for the with body").
			WithParsePhase("with statement body").
			Build()
		p.addStructuredError(err)
		return nil
	}

	stmt = builder.FinishWithNode(stmt, stmt.Body).(*ast.WithStatement)
	return stmt
}

func (p *Parser) parseWithDeclaration() *ast.VarDeclStatement {
	nameToken := p.cursor.Current()
	if !p.isIdentifierToken(nameToken.Type) {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected identifier in with declaration").
			WithPosition(nameToken.Pos, nameToken.Length()).
			WithExpectedString("variable name").
			WithActual(nameToken.Type, nameToken.Literal).
			WithSuggestion("provide a variable name after 'with'").
			WithParsePhase("with declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}

	decl := &ast.VarDeclStatement{
		BaseNode: ast.BaseNode{Token: nameToken},
		Names: []*ast.Identifier{
			{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: nameToken},
				},
				Value: nameToken.Literal,
			},
		},
	}

	p.parseVarType(decl)

	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.ASSIGN && nextToken.Type != lexer.EQ {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrInvalidSyntax).
			WithMessage("expected initializer in with declaration").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("':=' or '='").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("initialize the with variable before 'do'").
			WithParsePhase("with declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}

	p.cursor = p.cursor.Advance() // move to assignment operator
	if decl.Type == nil {
		decl.Inferred = true
	}
	p.cursor = p.cursor.Advance() // move to value expression
	decl.Value = p.parseExpression(ASSIGN)
	if decl.Value == nil {
		currentToken := p.cursor.Current()
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidExpression).
			WithMessage("expected expression for with declaration initializer").
			WithPosition(currentToken.Pos, currentToken.Length()).
			WithSuggestion("provide a value for the with variable").
			WithParsePhase("with declaration").
			Build()
		p.addStructuredError(err)
		return nil
	}

	decl.EndPos = decl.Value.End()
	return decl
}
