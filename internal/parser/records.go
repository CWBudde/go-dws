package parser

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// parseRecordOrHelperDeclaration determines if this is a record or helper declaration (dispatcher).
// Task 2.7.3: Dual-mode dispatcher for record/helper parsing.
func (p *Parser) parseRecordOrHelperDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	return p.parseRecordOrHelperDeclarationCursor(nameIdent, typeToken)
}

// parseRecordOrHelperDeclarationTraditional determines if this is a record or helper declaration.
// Called when we see 'type Name = record' - need to check if followed by 'helper'.
// Current token is positioned at '=' and peek token is 'record'.
// PRE: curToken is EQ, peekToken is RECORD
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordOrHelperDeclarationTraditional(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()
	// Move to RECORD
	if !p.expectPeek(lexer.RECORD) {
		return nil
	}

	// Check if next token is HELPER
	if p.peekTokenIs(lexer.HELPER) {
		// It's a helper declaration!
		p.nextToken() // Move to HELPER
		return p.parseHelperDeclarationCursor(nameIdent, typeToken, true)
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
			constant := p.parseClassConstantDeclarationCursor(currentVisibility, false)
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
				fields := p.parseRecordFieldDeclarationsCursor(currentVisibility)
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
				constant := p.parseClassConstantDeclarationCursor(currentVisibility, true)
				if constant != nil {
					recordDecl.Constants = append(recordDecl.Constants, constant)
				}
				p.nextToken()
				continue
			} else if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclarationCursor()
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
			method := p.parseFunctionDeclarationCursor()
			if method != nil {
				recordDecl.Methods = append(recordDecl.Methods, method)
			}
			p.nextToken()
			continue
		}

		// Check for property declarations
		if p.curTokenIs(lexer.PROPERTY) {
			prop := p.parseRecordPropertyDeclarationCursor()
			if prop != nil {
				recordDecl.Properties = append(recordDecl.Properties, *prop)
			}
			p.nextToken()
			continue
		}

		// Parse field declaration(s)
		fields := p.parseRecordFieldDeclarationsCursor(currentVisibility)
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

// parseRecordOrHelperDeclarationCursor determines if this is a record or helper declaration (cursor mode).
// Task 2.7.3.4: Cursor-based implementation for immutable parsing.
// PRE: cursor is at EQ, next token is RECORD
// POST: cursor is at SEMICOLON
func (p *Parser) parseRecordOrHelperDeclarationCursor(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()
	cursor := p.cursor

	// Expect RECORD token
	if cursor.Peek(1).Type != lexer.RECORD {
		p.addError("expected 'record' keyword", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to RECORD
	p.cursor = cursor

	// Check if next token is HELPER
	if cursor.Peek(1).Type == lexer.HELPER {
		cursor = cursor.Advance() // move to HELPER
		p.cursor = cursor
		return p.parseHelperDeclarationCursor(nameIdent, typeToken, true)
	}

	// It's a regular record declaration - advance to first token inside record
	cursor = cursor.Advance()
	p.cursor = cursor

	// Build the record declaration inline
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

	// Parse record body using shared helper
	currentVisibility = p.parseRecordBodyCursor(recordDecl, currentVisibility)
	cursor = p.cursor

	// Expect 'end' keyword
	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close record declaration", ErrMissingEnd)
		return nil
	}

	// Expect semicolon after 'end'
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	return builder.Finish(recordDecl).(*ast.RecordDecl)
}

// parseRecordDeclaration parses a record type declaration (dispatcher).
// Task 2.7.3: Dual-mode dispatcher for record parsing.
func (p *Parser) parseRecordDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.RecordDecl {
	return p.parseRecordDeclarationCursor(nameIdent, typeToken)
}

// parseRecordDeclarationTraditional parses a record type declaration.
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
func (p *Parser) parseRecordDeclarationTraditional(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.RecordDecl {
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
			constant := p.parseClassConstantDeclarationCursor(currentVisibility, false)
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
				fields := p.parseRecordFieldDeclarationsCursor(currentVisibility)
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
				constant := p.parseClassConstantDeclarationCursor(currentVisibility, true)
				if constant != nil {
					recordDecl.Constants = append(recordDecl.Constants, constant)
				}
				p.nextToken()
				continue
			} else if p.curTokenIs(lexer.FUNCTION) || p.curTokenIs(lexer.PROCEDURE) {
				// Class method: class function/procedure ...
				method := p.parseFunctionDeclarationCursor()
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
			method := p.parseFunctionDeclarationCursor()
			if method != nil {
				recordDecl.Methods = append(recordDecl.Methods, method)
			}
			p.nextToken()
			continue
		}

		// Check for property declarations
		if p.curTokenIs(lexer.PROPERTY) {
			prop := p.parseRecordPropertyDeclarationCursor()
			if prop != nil {
				recordDecl.Properties = append(recordDecl.Properties, *prop)
			}
			p.nextToken()
			continue
		}

		// Parse field declaration(s)
		// Pattern: Name1, Name2: Type; or Name: Type;
		fields := p.parseRecordFieldDeclarationsCursor(currentVisibility)
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

// parseRecordDeclarationCursor parses a record type declaration (cursor mode).
// Task 2.7.3.4: Cursor-based implementation for immutable parsing.
// PRE: cursor is at EQ; next token is RECORD
// POST: cursor is at SEMICOLON
func (p *Parser) parseRecordDeclarationCursor(nameIdent *ast.Identifier, typeToken lexer.Token) *ast.RecordDecl {
	builder := p.StartNode()
	cursor := p.cursor

	recordDecl := &ast.RecordDecl{
		BaseNode:   ast.BaseNode{Token: typeToken},
		Name:       nameIdent,
		Fields:     []*ast.FieldDecl{},
		Methods:    []*ast.FunctionDecl{},
		Properties: []ast.RecordPropertyDecl{},
		Constants:  []*ast.ConstDecl{},
		ClassVars:  []*ast.FieldDecl{},
	}

	// Expect 'record' keyword
	if cursor.Peek(1).Type != lexer.RECORD {
		p.addError("expected 'record' keyword", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to RECORD
	p.cursor = cursor

	// Move to first token inside record
	cursor = cursor.Advance()
	p.cursor = cursor

	// Track current visibility level (default to public for records)
	currentVisibility := ast.VisibilityPublic

	// Parse record body using shared helper
	currentVisibility = p.parseRecordBodyCursor(recordDecl, currentVisibility)
	cursor = p.cursor

	// Expect 'end' keyword
	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close record declaration", ErrMissingEnd)
		return nil
	}

	// Expect semicolon after 'end'
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	return builder.Finish(recordDecl).(*ast.RecordDecl)
}

// parseRecordBodyCursor parses the body of a record declaration (cursor mode).
// This helper function extracts the common record body parsing logic used by both
// parseRecordOrHelperDeclarationCursor and parseRecordDeclarationCursor.
// PRE: cursor is positioned at the first token inside the record body
// POST: cursor is positioned at END keyword
func (p *Parser) parseRecordBodyCursor(recordDecl *ast.RecordDecl, currentVisibility ast.Visibility) ast.Visibility {
	cursor := p.cursor

	// Parse record body until 'end'
	for cursor.Current().Type != lexer.END && cursor.Current().Type != lexer.EOF {
		// Check for visibility modifiers
		if cursor.Current().Type == lexer.PRIVATE {
			currentVisibility = ast.VisibilityPrivate
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else if cursor.Current().Type == lexer.PUBLIC {
			currentVisibility = ast.VisibilityPublic
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		} else if cursor.Current().Type == lexer.PUBLISHED {
			// Published is treated as public for records
			currentVisibility = ast.VisibilityPublic
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for 'const' (record constant)
		if cursor.Current().Type == lexer.CONST {
			cursor = cursor.Advance() // move past 'const'
			p.cursor = cursor
			constant := p.parseClassConstantDeclarationCursor(currentVisibility, false)
			if constant != nil {
				recordDecl.Constants = append(recordDecl.Constants, constant)
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for 'class function' / 'class procedure' / 'class var' / 'class const'
		if cursor.Current().Type == lexer.CLASS {
			cursor = cursor.Advance() // move past 'class'
			p.cursor = cursor

			if cursor.Current().Type == lexer.VAR {
				// Class variable: class var FieldName: Type;
				cursor = cursor.Advance() // move past 'var'
				p.cursor = cursor
				fields := p.parseRecordFieldDeclarationsCursor(currentVisibility)
				for _, field := range fields {
					if field != nil {
						field.IsClassVar = true
						recordDecl.ClassVars = append(recordDecl.ClassVars, field)
					}
				}
				cursor = p.cursor.Advance()
				p.cursor = cursor
				continue
			} else if cursor.Current().Type == lexer.CONST {
				// Class constant: class const Name = Value;
				cursor = cursor.Advance() // move past 'const'
				p.cursor = cursor
				constant := p.parseClassConstantDeclarationCursor(currentVisibility, true)
				if constant != nil {
					recordDecl.Constants = append(recordDecl.Constants, constant)
				}
				cursor = p.cursor.Advance()
				p.cursor = cursor
				continue
			} else if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE {
				// Class method
				method := p.parseFunctionDeclarationCursor()
				if method != nil {
					method.IsClassMethod = true
					recordDecl.Methods = append(recordDecl.Methods, method)
				}
				cursor = p.cursor.Advance()
				p.cursor = cursor
				continue
			} else {
				p.addError("expected 'var', 'const', 'function' or 'procedure' after 'class' keyword in record", ErrUnexpectedToken)
				cursor = cursor.Advance()
				p.cursor = cursor
				continue
			}
		}

		// Check for method declarations (instance methods)
		if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE {
			method := p.parseFunctionDeclarationCursor()
			if method != nil {
				recordDecl.Methods = append(recordDecl.Methods, method)
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for property declarations
		if cursor.Current().Type == lexer.PROPERTY {
			prop := p.parseRecordPropertyDeclarationCursor()
			if prop != nil {
				recordDecl.Properties = append(recordDecl.Properties, *prop)
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Parse field declaration(s)
		fields := p.parseRecordFieldDeclarationsCursor(currentVisibility)
		if fields != nil {
			recordDecl.Fields = append(recordDecl.Fields, fields...)
		}

		cursor = p.cursor.Advance()
		p.cursor = cursor
	}

	return currentVisibility
}

// parseRecordFieldDeclarations parses one or more field declarations (dispatcher).
// Task 2.7.3: Dual-mode dispatcher for record field parsing.
func (p *Parser) parseRecordFieldDeclarations(visibility ast.Visibility) []*ast.FieldDecl {
	return p.parseRecordFieldDeclarationsCursor(visibility)
}

// parseRecordFieldDeclarationsTraditional parses one or more field declarations with the same type.
// Pattern: Name1, Name2, Name3: Type;
// OR: Name := Value; (type inferred from initializer)
// Returns a slice of FieldDecl, one for each field name.
// PRE: curToken is field name IDENT
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordFieldDeclarationsTraditional(visibility ast.Visibility) []*ast.FieldDecl {
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
		initValue = p.parseExpressionCursor(LOWEST)
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
		fieldType = p.parseTypeExpressionCursor()
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

// parseRecordFieldDeclarationsCursor parses one or more field declarations (cursor mode).
// Task 2.7.3.4: Cursor-based implementation for immutable parsing.
// PRE: cursor is at field name IDENT
// POST: cursor is at SEMICOLON
func (p *Parser) parseRecordFieldDeclarationsCursor(visibility ast.Visibility) []*ast.FieldDecl {
	cursor := p.cursor

	// Use IdentifierList combinator to parse comma-separated field names
	fieldNames := p.IdentifierList(IdentifierListConfig{
		ErrorContext:      "record field declaration",
		RequireAtLeastOne: true,
	})
	if fieldNames == nil {
		return nil
	}

	cursor = p.cursor // Update cursor after IdentifierList
	var fieldType ast.TypeExpression
	var initValue ast.Expression

	// Check if this is type inference (Name := Value) or explicit type (Name : Type [= Value])
	if cursor.Peek(1).Type == lexer.ASSIGN {
		// Type inference: Name := Value
		if len(fieldNames) > 1 {
			p.addError("type inference not allowed for comma-separated field declarations", ErrInvalidExpression)
			return nil
		}

		cursor = cursor.Advance() // move to :=
		cursor = cursor.Advance() // move to value expression
		p.cursor = cursor

		// Parse initialization expression
		initValue = p.parseExpressionCursor(LOWEST)
		if initValue == nil {
			p.addError("expected initialization expression after :=", ErrInvalidExpression)
			return nil
		}

		cursor = p.cursor // synchronize cursor after parseExpression

		// Type will be inferred during semantic analysis (set to nil for now)
		fieldType = nil
	} else {
		// Explicit type: Name : Type [= Value]
		// Expect colon
		if cursor.Peek(1).Type != lexer.COLON {
			p.addError("expected ':' after field name", ErrUnexpectedToken)
			return nil
		}
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move to type
		p.cursor = cursor

		// Parse type expression
		fieldType = p.parseTypeExpressionCursor()
		if fieldType == nil {
			return nil
		}

		cursor = p.cursor // Update cursor after parseTypeExpression

		// Parse optional field initializer
		initValue = p.parseFieldInitializer(fieldNames)
		cursor = p.cursor // Update cursor after parseFieldInitializer
	}

	// Expect semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after field declaration", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

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
			InitValue:  initValue,
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

	// Task 2.7.6: Dual-mode - get current token for '(' lparen
	var lparenTok lexer.Token
	if p.cursor != nil {
		lparenTok = p.cursor.Current()
	} else {
		lparenTok = p.curToken
	}

	recordLit := &ast.RecordLiteralExpression{
		BaseNode: ast.BaseNode{Token: lparenTok}, // '(' token
		TypeName: nil,                            // Anonymous record (type inferred from context)
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
			// Task 2.7.6: Dual-mode - get current token for field name
			var fieldNameToken lexer.Token
			if p.cursor != nil {
				fieldNameToken = p.cursor.Current()
			} else {
				fieldNameToken = p.curToken
			}

			// Named field initialization
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
			value := p.parseExpressionCursor(LOWEST)
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

// parseRecordPropertyDeclaration parses a record property declaration (dispatcher).
// Task 2.7.3: Dual-mode dispatcher for record property parsing.
func (p *Parser) parseRecordPropertyDeclaration() *ast.RecordPropertyDecl {
	return p.parseRecordPropertyDeclarationCursor()
}

// parseRecordPropertyDeclarationTraditional parses a record property declaration.
// Pattern: property Name: Type read FieldName write FieldName;
// Also supports array properties: property Name[Index: Type]: Type read GetMethod;
//
// Note: This is different from class properties (parsePropertyDeclaration)
// PRE: curToken is PROPERTY
// POST: curToken is SEMICOLON
func (p *Parser) parseRecordPropertyDeclarationTraditional() *ast.RecordPropertyDecl {
	propToken := p.curToken // 'property' token

	// Expect property name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	propName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: p.cursor.Current(),
			},
		},
		Value: p.cursor.Current().Literal,
	}

	// Parse optional index parameters for array properties
	// Pattern: property Items[Index: Integer; Key: String]: String
	var indexParams []*ast.Parameter
	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Parse parameter list using the same pattern as function parameters
		for !p.peekTokenIs(lexer.RBRACK) && !p.peekTokenIs(lexer.EOF) {
			p.nextToken() // move to parameter name

			// Parse parameter name
			if !p.curTokenIs(lexer.IDENT) {
				p.addError("expected parameter name in property index", ErrUnexpectedToken)
				return nil
			}
			paramName := &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: p.cursor.Current()},
				},
				Value: p.cursor.Current().Literal,
			}

			// Expect colon
			if !p.expectPeek(lexer.COLON) {
				return nil
			}

			// Parse type
			p.nextToken() // move to type
			paramType := p.parseTypeExpressionCursor()
			if paramType == nil {
				return nil
			}

			param := &ast.Parameter{
				Token: paramName.Token,
				Name:  paramName,
				Type:  paramType,
			}
			indexParams = append(indexParams, param)

			// Check for more parameters (separated by semicolon or comma)
			if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // move to separator
				continue
			}

			// No more parameters - expect closing bracket
			break
		}

		// Expect closing bracket
		if !p.expectPeek(lexer.RBRACK) {
			return nil
		}
	}

	// Expect colon
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Expect type
	p.nextToken() // move to type
	propType := p.parseTypeExpressionCursor()
	if propType == nil {
		return nil
	}

	prop := &ast.RecordPropertyDecl{
		BaseNode:    ast.BaseNode{Token: propToken},
		Name:        propName,
		Type:        propType,
		IndexParams: indexParams,
		ReadField:   "",
		WriteField:  "",
		IsDefault:   false,
	}

	// Parse optional 'read' clause
	if p.peekTokenIs(lexer.READ) {
		p.nextToken() // move to 'read'
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		prop.ReadField = p.cursor.Current().Literal
	}

	// Parse optional 'write' clause
	if p.peekTokenIs(lexer.WRITE) {
		p.nextToken() // move to 'write'
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		prop.WriteField = p.cursor.Current().Literal
	}

	// Expect semicolon first
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Then check for optional 'default' keyword after the semicolon
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

// parseRecordPropertyDeclarationCursor parses a record property declaration (cursor mode).
// Task 2.7.3.4: Cursor-based implementation for immutable parsing.
// PRE: cursor is at PROPERTY
// POST: cursor is at SEMICOLON
func (p *Parser) parseRecordPropertyDeclarationCursor() *ast.RecordPropertyDecl {
	cursor := p.cursor
	propToken := cursor.Current() // 'property' token

	// Expect property name
	if cursor.Peek(1).Type != lexer.IDENT {
		p.addError("expected property name", ErrExpectedIdent)
		return nil
	}
	cursor = cursor.Advance() // move to IDENT
	p.cursor = cursor

	propName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	// Parse optional index parameters for array properties
	var indexParams []*ast.Parameter
	if cursor.Peek(1).Type == lexer.LBRACK {
		cursor = cursor.Advance() // move to '['
		p.cursor = cursor

		// Parse parameter list
		for cursor.Peek(1).Type != lexer.RBRACK && cursor.Peek(1).Type != lexer.EOF {
			cursor = cursor.Advance() // move to parameter name
			p.cursor = cursor

			// Parse parameter name
			if cursor.Current().Type != lexer.IDENT {
				p.addError("expected parameter name in property index", ErrUnexpectedToken)
				return nil
			}
			paramName := &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: cursor.Current()},
				},
				Value: cursor.Current().Literal,
			}

			// Expect colon
			if cursor.Peek(1).Type != lexer.COLON {
				p.addError("expected ':' after parameter name", ErrUnexpectedToken)
				return nil
			}
			cursor = cursor.Advance() // move to ':'
			cursor = cursor.Advance() // move to type
			p.cursor = cursor

			// Parse type
			paramType := p.parseTypeExpressionCursor()
			if paramType == nil {
				return nil
			}

			cursor = p.cursor // Update cursor after parseTypeExpression

			param := &ast.Parameter{
				Token: paramName.Token,
				Name:  paramName,
				Type:  paramType,
			}
			indexParams = append(indexParams, param)

			// Check for more parameters (separated by semicolon or comma)
			if cursor.Peek(1).Type == lexer.SEMICOLON || cursor.Peek(1).Type == lexer.COMMA {
				cursor = cursor.Advance() // move to separator
				p.cursor = cursor
				continue
			}

			// No more parameters - expect closing bracket
			break
		}

		// Expect closing bracket
		if cursor.Peek(1).Type != lexer.RBRACK {
			p.addError("expected ']' to close property index", ErrMissingRBracket)
			return nil
		}
		cursor = cursor.Advance() // move to ']'
		p.cursor = cursor
	}

	// Expect colon
	if cursor.Peek(1).Type != lexer.COLON {
		p.addError("expected ':' after property name", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to ':'
	cursor = cursor.Advance() // move to type
	p.cursor = cursor

	// Parse type
	propType := p.parseTypeExpressionCursor()
	if propType == nil {
		return nil
	}

	cursor = p.cursor // Update cursor after parseTypeExpression

	prop := &ast.RecordPropertyDecl{
		BaseNode:    ast.BaseNode{Token: propToken},
		Name:        propName,
		Type:        propType,
		IndexParams: indexParams,
		ReadField:   "",
		WriteField:  "",
		IsDefault:   false,
	}

	// Parse optional 'read' clause
	if cursor.Peek(1).Type == lexer.READ {
		cursor = cursor.Advance() // move to 'read'
		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected identifier after 'read'", ErrExpectedIdent)
			return nil
		}
		cursor = cursor.Advance() // move to identifier
		p.cursor = cursor
		prop.ReadField = cursor.Current().Literal
	}

	// Parse optional 'write' clause
	if cursor.Peek(1).Type == lexer.WRITE {
		cursor = cursor.Advance() // move to 'write'
		if cursor.Peek(1).Type != lexer.IDENT {
			p.addError("expected identifier after 'write'", ErrExpectedIdent)
			return nil
		}
		cursor = cursor.Advance() // move to identifier
		p.cursor = cursor
		prop.WriteField = cursor.Current().Literal
	}

	// Expect semicolon first
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after property declaration", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// Then check for optional 'default' keyword after the semicolon
	if cursor.Peek(1).Type == lexer.DEFAULT {
		cursor = cursor.Advance() // move to 'default'
		p.cursor = cursor
		prop.IsDefault = true
		// Expect another semicolon after 'default'
		if cursor.Peek(1).Type != lexer.SEMICOLON {
			p.addError("expected ';' after 'default'", ErrMissingSemicolon)
			return nil
		}
		cursor = cursor.Advance() // move to SEMICOLON
		p.cursor = cursor
	}

	return prop
}
