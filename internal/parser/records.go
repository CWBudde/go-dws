package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseRecordOrHelperDeclaration determines if this is a record or helper declaration.
// Called when we see 'type Name = record' - need to check if followed by 'helper'.
// Current token is positioned at '=' and peek token is 'record'.
// PRE: curToken is EQ, peekToken is RECORD
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordOrHelperDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()
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
		BaseNode:   ast.BaseNode{Token: typeToken},
		Name:       nameIdent,
		Fields:     []*ast.FieldDecl{},
		Methods:    []*ast.FunctionDecl{},
		Properties: []ast.RecordPropertyDecl{},
		Constants:  []*ast.ConstDecl{},
		ClassVars:  []*ast.FieldDecl{},
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

		// Check for 'const' (record constant)
		if p.curTokenIs(lexer.CONST) {
			p.nextToken() // move past 'const'
			constant := p.parseClassConstantDeclaration(currentVisibility, false)
			if constant != nil {
				recordDecl.Constants = append(recordDecl.Constants, constant)
			}
			p.nextToken()
			continue
		}

		// Check for 'class function' / 'class procedure' / 'class var' / 'class const' (class members)
		if p.curTokenIs(lexer.CLASS) {
			p.nextToken() // move past 'class'

			if p.curTokenIs(lexer.VAR) {
				// Class variable: class var FieldName: Type;
				p.nextToken() // move past 'var'
				fields := p.parseRecordFieldDeclarations(currentVisibility)
				for _, field := range fields {
					if field != nil {
						field.IsClassVar = true // Mark as class variable
						recordDecl.ClassVars = append(recordDecl.ClassVars, field)
					}
				}
				p.nextToken()
				continue
			} else if p.curTokenIs(lexer.CONST) {
				// Class constant: class const Name = Value;
				p.nextToken() // move past 'const'
				constant := p.parseClassConstantDeclaration(currentVisibility, true)
				if constant != nil {
					recordDecl.Constants = append(recordDecl.Constants, constant)
				}
				p.nextToken()
				continue
			} else if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true // Mark as static method
					recordDecl.Methods = append(recordDecl.Methods, method)
				}
				p.nextToken()
				continue
			} else {
				p.addError("expected 'var', 'const', 'function' or 'procedure' after 'class' keyword in record", ErrUnexpectedToken)
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
		p.addError("expected 'end' to close record declaration", ErrMissingEnd)
		return nil
	}

	// Expect semicolon after 'end'
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return builder.Finish(recordDecl).(*ast.RecordDecl)
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
// PRE: curToken is EQ; peekToken is RECORD
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.RecordDecl {
	builder := p.StartNode()
	recordDecl := &ast.RecordDecl{
		BaseNode:   ast.BaseNode{Token: typeToken}, // The 'type' token
		Name:       nameIdent,
		Fields:     []*ast.FieldDecl{},
		Methods:    []*ast.FunctionDecl{},
		Properties: []ast.RecordPropertyDecl{},
		Constants:  []*ast.ConstDecl{},
		ClassVars:  []*ast.FieldDecl{},
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

		// Check for 'const' (record constant)
		if p.curTokenIs(lexer.CONST) {
			p.nextToken() // move past 'const'
			constant := p.parseClassConstantDeclaration(currentVisibility, false)
			if constant != nil {
				recordDecl.Constants = append(recordDecl.Constants, constant)
			}
			p.nextToken()
			continue
		}

		// Check for 'class function' / 'class procedure' / 'class var' / 'class const' (class members)
		if p.curTokenIs(lexer.CLASS) {
			p.nextToken() // move past 'class'

			if p.curTokenIs(lexer.VAR) {
				// Class variable: class var FieldName: Type;
				p.nextToken() // move past 'var'
				fields := p.parseRecordFieldDeclarations(currentVisibility)
				for _, field := range fields {
					if field != nil {
						field.IsClassVar = true // Mark as class variable
						recordDecl.ClassVars = append(recordDecl.ClassVars, field)
					}
				}
				p.nextToken()
				continue
			} else if p.curTokenIs(lexer.CONST) {
				// Class constant: class const Name = Value;
				p.nextToken() // move past 'const'
				constant := p.parseClassConstantDeclaration(currentVisibility, true)
				if constant != nil {
					recordDecl.Constants = append(recordDecl.Constants, constant)
				}
				p.nextToken()
				continue
			} else if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true // Mark as static method
					recordDecl.Methods = append(recordDecl.Methods, method)
				}
				p.nextToken()
				continue
			} else {
				p.addError("expected 'var', 'const', 'function' or 'procedure' after 'class' keyword in record", ErrUnexpectedToken)
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
		p.addError("expected 'end' to close record declaration", ErrMissingEnd)
		return nil
	}

	// Expect semicolon after 'end'
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	return builder.Finish(recordDecl).(*ast.RecordDecl)
}

// parseRecordFieldDeclarations parses one or more field declarations with the same type.
// Pattern: Name1, Name2, Name3: Type;
// OR: Name := Value; (type inferred from initializer)
// Returns a slice of FieldDecl, one for each field name.
// PRE: curToken is field name IDENT
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordFieldDeclarations(visibility ast.Visibility) []*ast.FieldDecl {
	// Use IdentifierList combinator to parse comma-separated field names (Task 2.3.3)
	fieldNames := p.IdentifierList(IdentifierListConfig{
		ErrorContext:      "record field declaration",
		RequireAtLeastOne: true,
	})
	if fieldNames == nil {
		return nil
	}

	var fieldType ast.TypeExpression
	var initValue ast.Expression

	// Check if this is type inference (Name := Value) or explicit type (Name : Type [= Value])
	if p.peekTokenIs(lexer.ASSIGN) {
		// Type inference: Name := Value
		if len(fieldNames) > 1 {
			p.addError("type inference not allowed for comma-separated field declarations", ErrInvalidExpression)
			return nil
		}

		p.nextToken() // move to :=
		p.nextToken() // move to value expression

		// Parse initialization expression
		initValue = p.parseExpression(LOWEST)
		if initValue == nil {
			p.addError("expected initialization expression after :=", ErrInvalidExpression)
			return nil
		}

		// Type will be inferred during semantic analysis (set to nil for now)
		fieldType = nil
	} else {
		// Explicit type: Name : Type [= Value]
		// Expect colon
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		// Parse type expression (supports simple types, array types, function pointer types)
		p.nextToken() // move to type
		fieldType = p.parseTypeExpression()
		if fieldType == nil {
			return nil
		}

		// Parse optional field initializer
		initValue = p.parseFieldInitializer(fieldNames)
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Create a FieldDecl for each field name
	var fields []*ast.FieldDecl
	for _, name := range fieldNames {
		fields = append(fields, &ast.FieldDecl{
			BaseNode: ast.BaseNode{
				Token: name.Token,
			},
			Name:       name,
			Type:       fieldType,
			Visibility: visibility,
			InitValue:  initValue, // May be nil if no initialization
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
// PRE: curToken is LPAREN
// POST: curToken is RPAREN
func (p *Parser) parseRecordLiteral() *ast.RecordLiteralExpression {
	builder := p.StartNode()
	recordLit := &ast.RecordLiteralExpression{
		BaseNode: ast.BaseNode{Token: p.curToken}, // '(' token
		TypeName: nil,                             // Anonymous record (type inferred from context)
		Fields:   []*ast.FieldInitializer{},
	}

	// Check for empty literal
	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // move to ')'
		// Set EndPos to after the ')'
		return builder.Finish(recordLit).(*ast.RecordLiteralExpression)
	}

	// Move to first element
	p.nextToken()

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		// Check if this is named field initialization (Name: Value)
		// We need to look ahead to see if there's a colon
		if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
			// Named field initialization
			fieldNameToken := p.curToken
			fieldName := &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: fieldNameToken,
					},
				},
				Value: fieldNameToken.Literal,
			}

			p.nextToken() // move to ':'
			p.nextToken() // move to value expression

			// Parse value expression
			value := p.parseExpression(LOWEST)
			if value == nil {
				p.addError("expected expression after ':' in record literal field", ErrInvalidExpression)
				return nil
			}

			fieldInit := &ast.FieldInitializer{
				BaseNode: ast.BaseNode{Token: fieldNameToken},
				Name:     fieldName,
				Value:    value,
			}

			recordLit.Fields = append(recordLit.Fields, fieldInit)
		} else {
			// Positional field - not yet implemented
			p.addError("positional record field initialization not yet supported", ErrInvalidSyntax)
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
			p.addError("expected ',' or ';' or ')' in record literal", ErrUnexpectedToken)
			return nil
		}
	}

	// Expect closing paren
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Set EndPos to after the ')'
	return builder.Finish(recordLit).(*ast.RecordLiteralExpression)
}

// parseRecordPropertyDeclaration parses a record property declaration.
// Pattern: property Name: Type read FieldName write FieldName;
//
// Note: This is different from class properties (parsePropertyDeclaration)
// PRE: curToken is PROPERTY
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordPropertyDeclaration() *ast.RecordPropertyDecl {
	propToken := p.curToken // 'property' token

	// Expect property name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	propName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.curToken,
			},
		},
		Value: p.curToken.Literal,
	}

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
		BaseNode:   ast.BaseNode{Token: propToken},
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
