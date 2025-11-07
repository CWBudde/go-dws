package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseHelperDeclaration parses a helper type declaration.
// Helper syntax variants:
//   - type THelper = record helper for TypeName ... end;
//   - type THelper = helper for TypeName ... end;
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
// Helpers can contain:
//   - Methods (functions and procedures)
//   - Properties
//   - Class variables (class var)
//   - Class constants (class const)
//   - Visibility sections (private/public)
func (p *Parser) parseHelperDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token, isRecordHelper bool) *ast.HelperDecl {
	helperDecl := &ast.HelperDecl{
		Token:          p.curToken, // The HELPER token
		Name:           nameIdent,
		IsRecordHelper: isRecordHelper,
		Methods:        []*ast.FunctionDecl{},
		Properties:     []*ast.PropertyDecl{},
		ClassVars:      []*ast.FieldDecl{},
		ClassConsts:    []*ast.ConstDecl{},
		PrivateMembers: []ast.Statement{},
		PublicMembers:  []ast.Statement{},
	}

	// Expect 'for' keyword after 'helper'
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

	return helperDecl
}
