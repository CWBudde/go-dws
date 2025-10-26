package parser

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
)

// parsePropertyDeclaration parses a class property declaration.
// Called when current token is 'property'.
//
// Syntax variations:
//   - property Name: Type read ReadSpec write WriteSpec;
//   - property Name: Type read ReadSpec;  (read-only)
//   - property Name: Type write WriteSpec; (write-only)
//   - property Name: Type read (Expression); (expression-based read)
//   - property Items[index: Integer]: Type read GetItem write SetItem; (indexed)
//   - property Items[i: Integer]: Type read GetItem; default; (default indexed)
//   - property Name: Type; (auto-property, generates backing field)
//
// Task 8.36: Main entry point for property parsing
// Task 8.37: Parse 'property Name : Type read ReadSpec;'
// Task 8.38: Parse 'write WriteSpec' clause
func (p *Parser) parsePropertyDeclaration() *ast.PropertyDecl {
	propToken := p.curToken // 'property' token

	// Task 8.37: Parse property name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	propName := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// Task 8.39: Check for indexed property parameters: property Items[index: Integer]
	var indexParams []*ast.Parameter
	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Check for empty brackets (not allowed)
		if p.peekTokenIs(lexer.RBRACK) {
			p.addError("indexed property cannot have empty parameter list")
			return nil
		}

		p.nextToken() // move to first parameter name

		// Parse indexed property parameters (similar to function parameters but with brackets)
		for {
			// Parse parameter group (may have multiple names with same type)
			groupParams := p.parseIndexedPropertyParameterGroup()
			if groupParams == nil {
				return nil
			}
			indexParams = append(indexParams, groupParams...)

			// After parsing a parameter group, check what comes next:
			// - ']' : end of parameters
			// - ';' : more parameter groups follow
			if !p.peekTokenIs(lexer.RBRACK) && !p.peekTokenIs(lexer.SEMICOLON) {
				p.addError("expected ']' or ';' after indexed property parameter")
				return nil
			}

			if p.peekTokenIs(lexer.SEMICOLON) {
				p.nextToken() // move to ';'
				p.nextToken() // move past ';' to next parameter name
				continue
			}

			// Must be at ']', exit loop
			break
		}

		// Expect closing bracket
		if !p.expectPeek(lexer.RBRACK) {
			return nil
		}
	}

	// Task 8.37: Expect colon before type
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Task 8.37: Parse property type
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	propType := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	prop := &ast.PropertyDecl{
		Token:       propToken,
		Name:        propName,
		Type:        propType,
		IndexParams: indexParams,
		IsDefault:   false,
	}

	// Task 8.37: Parse optional 'read' clause
	// ReadSpec can be:
	// - Identifier (field or method name)
	// - Expression in parentheses: read (FValue * 2)
	if p.peekTokenIs(lexer.READ) {
		p.nextToken() // move to 'read'
		p.nextToken() // move to read specifier

		// Check if read spec is an expression in parentheses
		if p.curTokenIs(lexer.LPAREN) {
			// Parse expression-based read spec
			readExpr := p.parseExpression(LOWEST)
			prop.ReadSpec = readExpr
		} else if p.curTokenIs(lexer.IDENT) {
			// Simple identifier (field or method name)
			prop.ReadSpec = &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			}
		} else {
			p.addError("expected identifier or expression after 'read'")
			return nil
		}
	}

	// Task 8.38: Parse optional 'write' clause
	// WriteSpec can be:
	// - Identifier (field or method name)
	if p.peekTokenIs(lexer.WRITE) {
		p.nextToken() // move to 'write'

		if !p.expectPeek(lexer.IDENT) {
			return nil
		}

		// Simple identifier (field or method name)
		prop.WriteSpec = &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	}

	// Task 8.41: If neither read nor write was specified, generate auto-property
	// Auto-property generates backing field FName (F + property name)
	if prop.ReadSpec == nil && prop.WriteSpec == nil {
		// Generate backing field name: F + property name
		backingFieldName := "F" + propName.Value
		backingField := &ast.Identifier{
			Token: propName.Token,
			Value: backingFieldName,
		}

		// Auto-property has both read and write access to the backing field
		prop.ReadSpec = backingField
		prop.WriteSpec = backingField
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Task 8.40: Parse optional 'default;' keyword
	// This comes after the semicolon: property Items[i: Integer]: String read GetItem; default;
	if p.peekTokenIs(lexer.DEFAULT) {
		p.nextToken() // move to 'default'
		prop.IsDefault = true

		// Expect another semicolon after 'default'
		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}
	}

	return prop
}

// parseIndexedPropertyParameterGroup parses a group of indexed property parameters with the same type.
// Syntax: name: Type  or  name1, name2: Type
// Similar to parseParameterGroup but without 'var' keyword support.
func (p *Parser) parseIndexedPropertyParameterGroup() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Collect parameter names separated by commas
	names := []*ast.Identifier{}

	for {
		// Parse parameter name (can be IDENT or keyword used as identifier)
		if !p.curTokenIs(lexer.IDENT) && !p.curTokenIs(lexer.INDEX) {
			p.addError("expected parameter name in indexed property")
			return nil
		}

		names = append(names, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})

		// Check if there are more names (comma-separated)
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to ','
			p.nextToken() // move past ','
			continue
		}

		break
	}

	// Expect colon before type
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Parse type annotation
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	paramType := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Create parameter for each name
	for _, name := range names {
		param := &ast.Parameter{
			Name: name,
			Type: paramType,
		}
		params = append(params, param)
	}

	return params
}
