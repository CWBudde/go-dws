package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseUnit parses a complete unit declaration.
// Syntax:
//
//	unit UnitName;
//	interface
//	  uses ...;
//	  // declarations
//	implementation
//	  uses ...;
//	  // implementations
//	initialization
//	  // init code
//	finalization
//	  // cleanup code
//	end.
//
// PRE: cursor is UNIT
// POST: cursor is DOT

// PRE: cursor is UNIT
// POST: cursor is DOT
func (p *Parser) parseUnit() *ast.UnitDeclaration {
	builder := p.StartNode()
	unitDecl := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: p.cursor.Current()}, // 'unit' token
	}

	// Expect unit name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	unitDecl.Name = &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.cursor.Current(),
			},
		},
		Value: p.cursor.Current().Literal,
	}

	// Expect semicolon after unit name
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Move past semicolon
	p.nextToken()

	// Parse interface section (optional but common)
	if p.curTokenIs(lexer.INTERFACE) {
		unitDecl.InterfaceSection = p.parseInterfaceSection()
	}

	// Parse implementation section (optional but common)
	if p.curTokenIs(lexer.IMPLEMENTATION) {
		unitDecl.ImplementationSection = p.parseImplementationSection()
	}

	// Parse initialization section (optional)
	if p.curTokenIs(lexer.INITIALIZATION) {
		unitDecl.InitSection = p.parseInitializationSection()
	}

	// Parse finalization section (optional)
	if p.curTokenIs(lexer.FINALIZATION) {
		unitDecl.FinalSection = p.parseFinalizationSection()
	}

	// Expect 'end.' to close the unit
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close unit declaration", ErrMissingEnd)
		return nil
	}

	// Expect '.' after 'end'
	if !p.expectPeek(lexer.DOT) {
		p.addError("expected '.' after 'end' in unit declaration", ErrUnexpectedToken)
		return nil
	}

	return builder.Finish(unitDecl).(*ast.UnitDeclaration)
}

// parseUsesClause parses a uses statement.
// Syntax: uses Unit1, Unit2, Unit3;
// PRE: cursor is USES
// POST: cursor is SEMICOLON

// Syntax: uses Unit1, Unit2, Unit3;
// PRE: cursor is USES
// POST: cursor is SEMICOLON
func (p *Parser) parseUsesClause() *ast.UsesClause {
	builder := p.StartNode()
	currentToken := p.cursor.Current() // Store USES token

	usesClause := &ast.UsesClause{
		BaseNode: ast.BaseNode{Token: currentToken}, // 'uses' token
		Units:    []*ast.Identifier{},
	}

	// Expect at least one unit name
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrExpectedIdent).
			WithMessage("expected unit name after 'uses'").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("unit name").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("provide a unit name after 'uses'").
			WithParsePhase("uses clause").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to first unit name
	currentToken = p.cursor.Current()

	// Add first unit
	usesClause.Units = append(usesClause.Units, &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: currentToken,
			},
		},
		Value: currentToken.Literal,
	})

	// Parse remaining units (comma-separated)
	for {
		nextToken = p.cursor.Peek(1)
		if nextToken.Type != lexer.COMMA {
			break
		}

		p.cursor = p.cursor.Advance() // move to comma
		p.cursor = p.cursor.Advance() // move to next unit name
		currentToken = p.cursor.Current()

		if currentToken.Type != lexer.IDENT {
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrExpectedIdent).
				WithMessage("expected unit name after comma in uses clause").
				WithPosition(currentToken.Pos, currentToken.Length()).
				WithExpectedString("unit name").
				WithActual(currentToken.Type, currentToken.Literal).
				WithSuggestion("provide a unit name after comma").
				WithParsePhase("uses clause").
				Build()
			p.addStructuredError(err)
			return nil
		}

		usesClause.Units = append(usesClause.Units, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: currentToken,
				},
			},
			Value: currentToken.Literal,
		})
	}

	// Expect semicolon
	nextToken = p.cursor.Peek(1)
	if nextToken.Type != lexer.SEMICOLON {
		err := NewStructuredError(ErrKindMissing).
			WithCode(ErrMissingSemicolon).
			WithMessage("expected ';' after uses clause").
			WithPosition(nextToken.Pos, nextToken.Length()).
			WithExpectedString("';'").
			WithActual(nextToken.Type, nextToken.Literal).
			WithSuggestion("add ';' after unit names").
			WithParsePhase("uses clause").
			Build()
		p.addStructuredError(err)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to semicolon

	// Cursor is now on semicolon as expected by POST condition
	return builder.Finish(usesClause).(*ast.UsesClause)
}

// parseInterfaceSection parses the interface section of a unit.
// The interface section contains public declarations.
// PRE: cursor is INTERFACE
// POST: cursor is IMPLEMENTATION, INITIALIZATION, FINALIZATION, or END
func (p *Parser) parseInterfaceSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode:   ast.BaseNode{Token: p.cursor.Current()}, // 'interface' token
		Statements: []ast.Statement{},
	}

	p.nextToken() // move past 'interface'

	// Parse uses clause if present
	if p.curTokenIs(lexer.USES) {
		usesClause := p.parseUsesClause()
		if usesClause != nil {
			block.Statements = append(block.Statements, usesClause)
		}
	}

	// Parse declarations until we hit implementation, initialization, finalization, or end
	for !p.curTokenIs(lexer.IMPLEMENTATION) &&
		!p.curTokenIs(lexer.INITIALIZATION) &&
		!p.curTokenIs(lexer.FINALIZATION) &&
		!p.curTokenIs(lexer.END) &&
		!p.curTokenIs(lexer.EOF) {

		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseImplementationSection parses the implementation section of a unit.
// The implementation section contains private declarations and function implementations.
// PRE: cursor is IMPLEMENTATION
// POST: cursor is INITIALIZATION, FINALIZATION, or END
func (p *Parser) parseImplementationSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode:   ast.BaseNode{Token: p.cursor.Current()}, // 'implementation' token
		Statements: []ast.Statement{},
	}

	p.nextToken() // move past 'implementation'

	// Parse uses clause if present
	if p.curTokenIs(lexer.USES) {
		usesClause := p.parseUsesClause()
		if usesClause != nil {
			block.Statements = append(block.Statements, usesClause)
		}
	}

	// Parse declarations and implementations until we hit initialization, finalization, or end
	for !p.curTokenIs(lexer.INITIALIZATION) &&
		!p.curTokenIs(lexer.FINALIZATION) &&
		!p.curTokenIs(lexer.END) &&
		!p.curTokenIs(lexer.EOF) {

		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseInitializationSection parses the initialization section of a unit.
// The initialization section contains code that runs when the unit is loaded.
// PRE: cursor is INITIALIZATION
// POST: cursor is FINALIZATION or END
func (p *Parser) parseInitializationSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode:   ast.BaseNode{Token: p.cursor.Current()}, // 'initialization' token
		Statements: []ast.Statement{},
	}

	p.nextToken() // move past 'initialization'

	// Parse statements until we hit finalization or end
	for !p.curTokenIs(lexer.FINALIZATION) &&
		!p.curTokenIs(lexer.END) &&
		!p.curTokenIs(lexer.EOF) {

		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseFinalizationSection parses the finalization section of a unit.
// The finalization section contains cleanup code that runs when the program exits.
// PRE: cursor is FINALIZATION
// POST: cursor is END
func (p *Parser) parseFinalizationSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		BaseNode:   ast.BaseNode{Token: p.cursor.Current()}, // 'finalization' token
		Statements: []ast.Statement{},
	}

	p.nextToken() // move past 'finalization'

	// Parse statements until we hit end
	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Skip semicolons
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}
