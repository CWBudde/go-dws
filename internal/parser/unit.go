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
func (p *Parser) parseUnit() *ast.UnitDeclaration {
	unitDecl := &ast.UnitDeclaration{
		BaseNode: ast.BaseNode{Token: p.curToken}, // 'unit' token
	}

	// Expect unit name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	unitDecl.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
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

	// Expect 'end'
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close unit", ErrMissingEnd)
		return nil
	}

	// Expect '.' after 'end'
	if !p.expectPeek(lexer.DOT) {
		p.addError("expected '.' after 'end' in unit declaration", ErrUnexpectedToken)
		return nil
	}

	// Set end position to the '.' token
	unitDecl.EndPos = p.endPosFromToken(p.curToken)

	return unitDecl
}

// parseUsesClause parses a uses statement.
// Syntax: uses Unit1, Unit2, Unit3;
func (p *Parser) parseUsesClause() *ast.UsesClause {
	usesClause := &ast.UsesClause{
		BaseNode: ast.BaseNode{Token: p.curToken}, // 'uses' token
		Units:    []*ast.Identifier{},
	}

	// Expect at least one unit name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	// Add first unit
	usesClause.Units = append(usesClause.Units, &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	})

	// Parse remaining units (comma-separated)
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // move to comma
		p.nextToken() // move to next unit name

		if !p.curTokenIs(lexer.IDENT) {
			p.addError("expected unit name after comma in uses clause", ErrExpectedIdent)
			return nil
		}

		usesClause.Units = append(usesClause.Units, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Set end position to the semicolon
	usesClause.EndPos = p.endPosFromToken(p.curToken)

	// Don't move past semicolon - ParseProgram will do that
	return usesClause
}

// parseInterfaceSection parses the interface section of a unit.
// The interface section contains public declarations.
func (p *Parser) parseInterfaceSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken, // 'interface' token
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
func (p *Parser) parseImplementationSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken, // 'implementation' token
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
func (p *Parser) parseInitializationSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken, // 'initialization' token
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
func (p *Parser) parseFinalizationSection() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken, // 'finalization' token
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
