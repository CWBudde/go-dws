package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseHelperDeclaration parses a helper type declaration (dispatcher).
// Task 2.7.3: Dual-mode dispatcher for helper parsing.
func (p *Parser) parseHelperDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token, isRecordHelper bool) *ast.HelperDecl {
	if p.useCursor {
		return p.parseHelperDeclarationCursor(nameIdent, typeToken, isRecordHelper)
	}
	return p.parseHelperDeclarationTraditional(nameIdent, typeToken, isRecordHelper)
}

// parseHelperDeclarationTraditional parses a helper type declaration.
// Helper syntax variants:
//   - type THelper = record helper for TypeName ... end;
//   - type THelper = helper for TypeName ... end;
//   - type THelper = helper(TParentHelper) for TypeName ... end;
//
// Current token should be positioned at HELPER keyword on entry.
// nameIdent is the helper's type name.
// typeToken is the original TYPE token.
// isRecordHelper indicates if "record" keyword preceded "helper".
//
// Example:
//
//	type TStringHelper = record helper for String
//	  function ToUpper: String;
//	  property Length: Integer read GetLength;
//	end;
//
//	type TChildHelper = helper(TParentHelper) for String
//	  function ToLower: String;
//	end;
//
// Helpers can contain:
//   - Methods (functions and procedures)
//   - Properties
//   - Class variables (class var)
//   - Class constants (class const)
//   - Visibility sections (private/public)
//
// PRE: curToken is HELPER
// POST: curToken is SEMICOLON
func (p *Parser) parseHelperDeclarationTraditional(nameIdent *ast.Identifier, typeToken lexer.Token, isRecordHelper bool) *ast.HelperDecl {
	builder := p.StartNode()
	helperDecl := &ast.HelperDecl{
		BaseNode: ast.BaseNode{
			Token: p.curToken, // The HELPER token
		},
		Name:           nameIdent,
		IsRecordHelper: isRecordHelper,
		Methods:        []*ast.FunctionDecl{},
		Properties:     []*ast.PropertyDecl{},
		ClassVars:      []*ast.FieldDecl{},
		ClassConsts:    []*ast.ConstDecl{},
		PrivateMembers: []ast.Statement{},
		PublicMembers:  []ast.Statement{},
	}

	// Check for optional parent helper: helper(TParentHelper)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // Move to LPAREN

		// Expect parent helper name
		if !p.expectPeek(lexer.IDENT) {
			p.addError("expected parent helper name after '(' in helper declaration", ErrExpectedIdent)
			return nil
		}

		helperDecl.ParentHelper = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.curToken,
				},
			},
			Value: p.curToken.Literal,
		}

		// Expect closing paren
		if !p.expectPeek(lexer.RPAREN) {
			p.addError("expected ')' after parent helper name", ErrMissingRParen)
			return nil
		}
	}

	// Expect 'for' keyword after 'helper' or after ')'
	if !p.expectPeek(lexer.FOR) {
		return nil
	}

	// Expect the target type name
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected type name after 'for' in helper declaration", ErrExpectedType)
		return nil
	}

	helperDecl.ForType = &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Move to first token inside helper body
	p.nextToken()

	// Track current visibility level (default to public for helpers)
	currentVisibility := ast.VisibilityPublic
	var currentSection *[]ast.Statement // Points to PrivateMembers or PublicMembers

	// Parse helper body until 'end'
	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Check for visibility modifiers
		if p.curTokenIs(lexer.PRIVATE) {
			currentVisibility = ast.VisibilityPrivate
			currentSection = &helperDecl.PrivateMembers
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLIC) {
			currentVisibility = ast.VisibilityPublic
			currentSection = &helperDecl.PublicMembers
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLISHED) {
			// Published is treated as public for helpers
			currentVisibility = ast.VisibilityPublic
			currentSection = &helperDecl.PublicMembers
			p.nextToken()
			continue
		}

		// Check for 'class const' declarations
		if p.curTokenIs(lexer.CLASS) && p.peekTokenIs(lexer.CONST) {
			p.nextToken() // Move to CONST
			classConstStmt := p.parseConstDeclaration()
			if classConstStmt != nil {
				if classConst, ok := classConstStmt.(*ast.ConstDecl); ok {
					helperDecl.ClassConsts = append(helperDecl.ClassConsts, classConst)
					if currentSection != nil {
						*currentSection = append(*currentSection, classConst)
					}
				}
			}
			p.nextToken()
			continue
		}

		// Check for 'class var' declarations
		if p.curTokenIs(lexer.CLASS) && p.peekTokenIs(lexer.VAR) {
			p.nextToken() // Move to VAR
			p.nextToken() // Move to identifier

			// Parse field declarations (can be comma-separated)
			fields := p.parseFieldDeclarations(currentVisibility)
			if fields != nil {
				for _, field := range fields {
					field.IsClassVar = true
					helperDecl.ClassVars = append(helperDecl.ClassVars, field)
					if currentSection != nil {
						*currentSection = append(*currentSection, field)
					}
				}
			}
			p.nextToken()
			continue
		}

		// Check for method declarations (function/procedure)
		if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
			method := p.parseFunctionDeclaration()
			if method != nil {
				helperDecl.Methods = append(helperDecl.Methods, method)
				if currentSection != nil {
					*currentSection = append(*currentSection, method)
				}
			}
			p.nextToken()
			continue
		}

		// Check for property declarations
		if p.curTokenIs(lexer.PROPERTY) {
			property := p.parsePropertyDeclaration()
			if property != nil {
				helperDecl.Properties = append(helperDecl.Properties, property)
				if currentSection != nil {
					*currentSection = append(*currentSection, property)
				}
			}
			p.nextToken()
			continue
		}

		// Unknown token - skip it
		p.addError("unexpected token in helper body: "+p.curToken.Literal, ErrUnexpectedToken)
		p.nextToken()
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close helper declaration", ErrMissingEnd)
		return nil
	}

	// Expect semicolon after 'end'
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return builder.Finish(helperDecl).(*ast.HelperDecl)
}

// parseHelperDeclarationCursor parses a helper type declaration (cursor mode).
// Task 2.7.3.4: Cursor-based implementation for immutable parsing.
// PRE: cursor is at HELPER
// POST: cursor is at SEMICOLON
func (p *Parser) parseHelperDeclarationCursor(nameIdent *ast.Identifier, typeToken lexer.Token, isRecordHelper bool) *ast.HelperDecl {
	builder := p.StartNode()
	cursor := p.cursor

	helperDecl := &ast.HelperDecl{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(), // The HELPER token
		},
		Name:           nameIdent,
		IsRecordHelper: isRecordHelper,
		Methods:        []*ast.FunctionDecl{},
		Properties:     []*ast.PropertyDecl{},
		ClassVars:      []*ast.FieldDecl{},
		ClassConsts:    []*ast.ConstDecl{},
		PrivateMembers: []ast.Statement{},
		PublicMembers:  []ast.Statement{},
	}

	// Check for optional parent helper: helper(TParentHelper)
	if cursor.Peek(1).Type == lexer.LPAREN {
		cursor = cursor.Advance() // Move to LPAREN
		p.cursor = cursor

		// Expect parent helper name
		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected parent helper name after '(' in helper declaration", ErrExpectedIdent)
			return nil
		}
		cursor = cursor.Advance() // move to IDENT
		p.cursor = cursor

		helperDecl.ParentHelper = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}

		// Expect closing paren
		if cursor.Peek(1).Type != lexer.RPAREN {
			p.addError("expected ')' after parent helper name", ErrMissingRParen)
			return nil
		}
		cursor = cursor.Advance() // move to ')'
		p.cursor = cursor
	}

	// Expect 'for' keyword after 'helper' or after ')'
	if cursor.Peek(1).Type != lexer.FOR {
		p.addError("expected 'for' keyword in helper declaration", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to FOR
	p.cursor = cursor

	// Expect the target type name
	if cursor.Peek(1).Type != lexer.IDENT {
		p.addError("expected type name after 'for' in helper declaration", ErrExpectedType)
		return nil
	}
	cursor = cursor.Advance() // move to IDENT
	p.cursor = cursor

	helperDecl.ForType = &ast.TypeAnnotation{
		Token: cursor.Current(),
		Name:  cursor.Current().Literal,
	}

	// Move to first token inside helper body
	cursor = cursor.Advance()
	p.cursor = cursor

	// Track current visibility level (default to public for helpers)
	currentVisibility := ast.VisibilityPublic
	var currentSection *[]ast.Statement // Points to PrivateMembers or PublicMembers

	// Parse helper body until 'end'
	for cursor.Current().Type != lexer.END && cursor.Current().Type != lexer.EOF {
		// Check for visibility modifiers
		if cursor.Current().Type == lexer.PRIVATE {
			currentVisibility = ast.VisibilityPrivate
			currentSection = &helperDecl.PrivateMembers
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else if cursor.Current().Type == lexer.PUBLIC {
			currentVisibility = ast.VisibilityPublic
			currentSection = &helperDecl.PublicMembers
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else if cursor.Current().Type == lexer.PUBLISHED {
			// Published is treated as public for helpers
			currentVisibility = ast.VisibilityPublic
			currentSection = &helperDecl.PublicMembers
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for 'class const' declarations
		if cursor.Current().Type == lexer.CLASS && cursor.Peek(1).Type == lexer.CONST {
			cursor = cursor.Advance() // Move to CONST
			p.cursor = cursor
			classConstStmt := p.parseConstDeclaration()
			if classConstStmt != nil {
				if classConst, ok := classConstStmt.(*ast.ConstDecl); ok {
					helperDecl.ClassConsts = append(helperDecl.ClassConsts, classConst)
					if currentSection != nil {
						*currentSection = append(*currentSection, classConst)
					}
				}
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for 'class var' declarations
		if cursor.Current().Type == lexer.CLASS && cursor.Peek(1).Type == lexer.VAR {
			cursor = cursor.Advance() // Move to VAR
			cursor = cursor.Advance() // Move to identifier
			p.cursor = cursor

			// Parse field declarations (can be comma-separated)
			fields := p.parseFieldDeclarations(currentVisibility)
			if fields != nil {
				for _, field := range fields {
					field.IsClassVar = true
					helperDecl.ClassVars = append(helperDecl.ClassVars, field)
					if currentSection != nil {
						*currentSection = append(*currentSection, field)
					}
				}
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for method declarations (function/procedure)
		if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE {
			method := p.parseFunctionDeclaration()
			if method != nil {
				helperDecl.Methods = append(helperDecl.Methods, method)
				if currentSection != nil {
					*currentSection = append(*currentSection, method)
				}
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for property declarations
		if cursor.Current().Type == lexer.PROPERTY {
			property := p.parsePropertyDeclaration()
			if property != nil {
				helperDecl.Properties = append(helperDecl.Properties, property)
				if currentSection != nil {
					*currentSection = append(*currentSection, property)
				}
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Unknown token - skip it
		p.addError("unexpected token in helper body: "+cursor.Current().Literal, ErrUnexpectedToken)
		cursor = cursor.Advance()
		p.cursor = cursor
	}

	// Expect 'end' keyword
	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close helper declaration", ErrMissingEnd)
		return nil
	}

	// Expect semicolon after 'end'
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	return builder.Finish(helperDecl).(*ast.HelperDecl)
}
