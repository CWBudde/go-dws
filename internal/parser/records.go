package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseRecordOrHelperDeclaration determines if this is a record or helper declaration.
// Called when we see 'type Name = record' - need to check if followed by 'helper'.
// Current token is positioned at '=' and peek token is 'record'.
func (p *Parser) parseRecordOrHelperDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	// Move to RECORD
	if !p.expectPeek(lexer.RECORD) {
		return nil
	}

	// Check if next token is HELPER
	if p.peekTokenIs(lexer.HELPER) {
		// It's a helper declaration!
		p.nextToken() // Move to HELPER
		return p.parseHelperDeclaration(nameIdent, typeToken, true)
	}

	// It's a regular record declaration
	// We're currently at RECORD, move to first token inside record
	p.nextToken()

	// Build the record declaration inline (similar to parseRecordDeclaration)
	recordDecl := &ast.RecordDecl{
		Token:      typeToken,
		Name:       nameIdent,
		Fields:     []*ast.FieldDecl{},
		Methods:    []*ast.FunctionDecl{},
		Properties: []ast.RecordPropertyDecl{},
	}

	// Track current visibility level (default to public for records)
	currentVisibility := ast.VisibilityPublic

	// Parse record body until 'end'
	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Check for visibility modifiers
		if p.curTokenIs(lexer.PRIVATE) {
			currentVisibility = ast.VisibilityPrivate
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLIC) {
			currentVisibility = ast.VisibilityPublic
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLISHED) {
			currentVisibility = ast.VisibilityPublic
			p.nextToken()
			continue
		}

		// Check for 'class function' / 'class procedure' (static methods)
		if p.curTokenIs(lexer.CLASS) {
			p.nextToken() // move past 'class'

			if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true // Mark as static method
					recordDecl.Methods = append(recordDecl.Methods, method)
				}
				p.nextToken()
				continue
			} else {
				p.addError("expected 'function' or 'procedure' after 'class' keyword in record")
				p.nextToken()
				continue
			}
		}

		// Check for method declarations (instance methods)
		if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
			method := p.parseFunctionDeclaration()
			if method != nil {
				recordDecl.Methods = append(recordDecl.Methods, method)
			}
			p.nextToken()
			continue
		}

		// Check for property declarations
		if p.curTokenIs(lexer.PROPERTY) {
			prop := p.parseRecordPropertyDeclaration()
			if prop != nil {
				recordDecl.Properties = append(recordDecl.Properties, *prop)
			}
			p.nextToken()
			continue
		}

		// Parse field declaration(s)
		fields := p.parseRecordFieldDeclarations(currentVisibility)
		if fields != nil {
			recordDecl.Fields = append(recordDecl.Fields, fields...)
		}

		p.nextToken()
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close record declaration")
		return nil
	}

	// Expect semicolon after 'end'
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return recordDecl
}

// parseRecordDeclaration parses a record type declaration.
// Called after 'type Name =' has already been parsed.
// Current token should be 'record'.
//
// Syntax:
//   - type TPoint = record X, Y: Integer; end;
//   - type TPerson = record Name: String; Age: Integer; end;
//   - type TPoint = record
//     private
//     FX, FY: Integer;
//     public
//     property X: Integer read FX write FX;
//     end;
//
// Task 8.61: Parse record declarations
func (p *Parser) parseRecordDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.RecordDecl {
	recordDecl := &ast.RecordDecl{
		Token:      typeToken, // The 'type' token
		Name:       nameIdent,
		Fields:     []*ast.FieldDecl{},
		Methods:    []*ast.FunctionDecl{},
		Properties: []ast.RecordPropertyDecl{},
	}

	// Expect 'record' keyword
	if !p.expectPeek(lexer.RECORD) {
		return nil
	}

	// Move to first token inside record
	p.nextToken()

	// Track current visibility level (default to public for records)
	currentVisibility := ast.VisibilityPublic

	// Parse record body until 'end'
	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.EOF) {
		// Check for visibility modifiers
		if p.curTokenIs(lexer.PRIVATE) {
			currentVisibility = ast.VisibilityPrivate
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLIC) {
			currentVisibility = ast.VisibilityPublic
			p.nextToken()
			continue
		} else if p.curTokenIs(lexer.PUBLISHED) {
			// Published is treated as public for records
			currentVisibility = ast.VisibilityPublic
			p.nextToken()
			continue
		}

		// Check for 'class function' / 'class procedure' (static methods)
		if p.curTokenIs(lexer.CLASS) {
			p.nextToken() // move past 'class'

			if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true // Mark as static method
					recordDecl.Methods = append(recordDecl.Methods, method)
				}
				p.nextToken()
				continue
			} else {
				p.addError("expected 'function' or 'procedure' after 'class' keyword in record")
				p.nextToken()
				continue
			}
		}

		// Check for method declarations (instance methods)
		if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
			method := p.parseFunctionDeclaration()
			if method != nil {
				recordDecl.Methods = append(recordDecl.Methods, method)
			}
			p.nextToken()
			continue
		}

		// Check for property declarations
		if p.curTokenIs(lexer.PROPERTY) {
			prop := p.parseRecordPropertyDeclaration()
			if prop != nil {
				recordDecl.Properties = append(recordDecl.Properties, *prop)
			}
			p.nextToken()
			continue
		}

		// Parse field declaration(s)
		// Pattern: Name1, Name2: Type; or Name: Type;
		fields := p.parseRecordFieldDeclarations(currentVisibility)
		if fields != nil {
			recordDecl.Fields = append(recordDecl.Fields, fields...)
		}

		// Move to next declaration
		p.nextToken()
	}

	// Expect 'end' keyword
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close record declaration")
		return nil
	}

	// Expect semicolon after 'end'
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return recordDecl
}

// parseRecordFieldDeclarations parses one or more field declarations with the same type.
// Pattern: Name1, Name2, Name3: Type;
// Returns a slice of FieldDecl, one for each field name.
func (p *Parser) parseRecordFieldDeclarations(visibility ast.Visibility) []*ast.FieldDecl {
	if !p.curTokenIs(lexer.IDENT) {
		p.addError("expected field name")
		return nil
	}

	// Collect all field names
	var fieldNames []*ast.Identifier
	fieldNames = append(fieldNames, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

	// Check for comma-separated names
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // move to comma
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		fieldNames = append(fieldNames, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	// Expect colon
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Expect type
	if !p.expectPeek(lexer.IDENT) {
		p.addError("expected type name after ':'")
		return nil
	}

	typeAnnotation := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Create a FieldDecl for each field name
	var fields []*ast.FieldDecl
	for _, name := range fieldNames {
		fields = append(fields, &ast.FieldDecl{
			Token:      name.Token,
			Name:       name,
			Type:       typeAnnotation,
			Visibility: visibility,
		})
	}

	return fields
}

// parseRecordLiteral parses a record literal expression.
// DWScript supports both semicolon and comma as separators:
//   - (x: 10; y: 20) - semicolon separator (preferred)
//   - (x: 10, y: 20) - comma separator (also valid)
//   - (10, 20) - positional initialization (not yet implemented)
//
// Called when we see '(' and need to determine if it's a record literal or grouped expression.
// This is called from parseGroupedExpression when it detects a record literal pattern.
func (p *Parser) parseRecordLiteral() *ast.RecordLiteralExpression {
	recordLit := &ast.RecordLiteralExpression{
		Token:    p.curToken, // '(' token
		TypeName: nil,        // Anonymous record (type inferred from context)
		Fields:   []*ast.FieldInitializer{},
	}

	// Check for empty literal
	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // move to ')'
		return recordLit
	}

	// Move to first element
	p.nextToken()

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		// Check if this is named field initialization (Name: Value)
		// We need to look ahead to see if there's a colon
		if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
			// Named field initialization
			fieldNameToken := p.curToken
			fieldName := &ast.Identifier{Token: fieldNameToken, Value: fieldNameToken.Literal}

			p.nextToken() // move to ':'
			p.nextToken() // move to value expression

			// Parse value expression
			value := p.parseExpression(LOWEST)
			if value == nil {
				p.addError("expected expression after ':' in record literal field")
				return nil
			}

			fieldInit := &ast.FieldInitializer{
				Token: fieldNameToken,
				Name:  fieldName,
				Value: value,
			}

			recordLit.Fields = append(recordLit.Fields, fieldInit)
		} else {
			// Positional field - not yet implemented
			p.addError("positional record field initialization not yet supported")
			return nil
		}

		// Check for separator (comma or semicolon - both valid in DWScript)
		if p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to separator
			// Allow optional trailing separator
			if p.peekTokenIs(lexer.RPAREN) {
				break
			}
			p.nextToken() // move to next field
		} else if !p.peekTokenIs(lexer.RPAREN) {
			p.addError("expected ',' or ';' or ')' in record literal")
			return nil
		}
	}

	// Expect closing paren
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return recordLit
}

// parseRecordPropertyDeclaration parses a record property declaration.
// Pattern: property Name: Type read FieldName write FieldName;
// Task 8.61d: Parse record properties
// Note: This is different from class properties (parsePropertyDeclaration)
func (p *Parser) parseRecordPropertyDeclaration() *ast.RecordPropertyDecl {
	propToken := p.curToken // 'property' token

	// Expect property name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	propName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Expect colon
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Expect type
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	propType := &ast.TypeAnnotation{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	prop := &ast.RecordPropertyDecl{
		Token:      propToken,
		Name:       propName,
		Type:       propType,
		ReadField:  "",
		WriteField: "",
	}

	// Parse optional 'read' clause
	if p.peekTokenIs(lexer.READ) {
		p.nextToken() // move to 'read'
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		prop.ReadField = p.curToken.Literal
	}

	// Parse optional 'write' clause
	if p.peekTokenIs(lexer.WRITE) {
		p.nextToken() // move to 'write'
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		prop.WriteField = p.curToken.Literal
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return prop
}
