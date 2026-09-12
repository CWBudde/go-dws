package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// parseRecordOrHelperDeclaration determines if this is a record or helper declaration (dispatcher).

// Called when we see 'type Name = record' - need to check if followed by 'helper'.
// PRE: cursor is RECORD (already advanced in parseSingleTypeDeclaration)
// POST: cursor is SEMICOLON
func (p *Parser) parseRecordOrHelperDeclaration(nameIdent *ast.Identifier, typeToken lexer.Token) ast.Statement {
	builder := p.StartNode()
	cursor := p.cursor

	// Cursor should already be on RECORD token
	if cursor.Current().Type != lexer.RECORD {
		p.addError("expected 'record' keyword", ErrUnexpectedToken)
		return nil
	}

	// Check if next token is HELPER
	if cursor.Peek(1).Type == lexer.HELPER {
		cursor = cursor.Advance() // move to HELPER
		p.cursor = cursor
		return p.parseHelperDeclaration(nameIdent, true)
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
	p.parseRecordBody(recordDecl, currentVisibility)
	cursor = p.cursor

	// Expect 'end' keyword
	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close record declaration", ErrMissingEnd)
		return nil
	}
	recordDecl.EndKeywordPos = cursor.Current().Pos

	// Expect semicolon after 'end'
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	decl, _ := builder.Finish(recordDecl).(*ast.RecordDecl)

	return decl
}

func (p *Parser) parseInlineRecordType() ast.TypeExpression {
	cursor := p.cursor
	recordToken := cursor.Current()

	if recordToken.Type != lexer.RECORD {
		p.addError("expected 'record' keyword", ErrUnexpectedToken)
		return invalidTypeExpression(recordToken, "record type expected")
	}

	cursor = cursor.Advance()
	p.cursor = cursor

	recordDecl := &ast.RecordDecl{
		BaseNode:   ast.BaseNode{Token: recordToken},
		Name:       &ast.Identifier{Value: ""},
		Fields:     []*ast.FieldDecl{},
		Methods:    []*ast.FunctionDecl{},
		Properties: []ast.RecordPropertyDecl{},
		Constants:  []*ast.ConstDecl{},
		ClassVars:  []*ast.FieldDecl{},
	}

	p.parseRecordBody(recordDecl, ast.VisibilityPublic)
	cursor = p.cursor

	if cursor.Current().Type != lexer.END {
		p.addError("expected 'end' to close record type", ErrMissingEnd)
		return &ast.InvalidTypeExpression{
			BaseNode: ast.BaseNode{Token: recordToken},
			Reason:   "unterminated record type",
		}
	}

	return &ast.RecordTypeNode{
		Token:              recordToken,
		EndPos:             cursor.Current().End(),
		Fields:             recordDecl.Fields,
		Methods:            recordDecl.Methods,
		Properties:         recordDecl.Properties,
		Constants:          recordDecl.Constants,
		ClassVars:          recordDecl.ClassVars,
		VisibilitySections: recordDecl.VisibilitySections,
	}
}

// This helper function extracts the common record body parsing logic used by both
// parseRecordOrHelperDeclaration and parseRecordDeclaration.
// PRE: cursor is positioned at the first token inside the record body
// POST: cursor is positioned at END keyword
//
//nolint:gocyclo // Record body parser handling multiple member types
func (p *Parser) parseRecordBody(recordDecl *ast.RecordDecl, currentVisibility ast.Visibility) ast.Visibility {
	cursor := p.cursor
	seenMethod := false

	// Parse record body until 'end'
	for cursor.Current().Type != lexer.END && cursor.Current().Type != lexer.EOF {
		// Check for visibility modifiers. `protected` is not a legal record
		// section, but it is consumed here so the analyzer can report it and
		// the rest of the body still parses.
		switch cursor.Current().Type {
		case lexer.PRIVATE, lexer.PUBLIC, lexer.PUBLISHED, lexer.PROTECTED:
			specifier := pkgident.Normalize(cursor.Current().Literal)
			recordDecl.VisibilitySections = append(recordDecl.VisibilitySections,
				ast.RecordVisibilitySection{Specifier: specifier, Pos: cursor.Current().Pos})
			switch cursor.Current().Type {
			case lexer.PRIVATE:
				currentVisibility = ast.VisibilityPrivate
			case lexer.PUBLIC, lexer.PUBLISHED:
				// Published is treated as public for records
				currentVisibility = ast.VisibilityPublic
			}
			cursor = cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for 'const' (record constant)
		if cursor.Current().Type == lexer.CONST {
			cursor = cursor.Advance() // move past 'const'
			p.cursor = cursor
			constant := p.parseClassConstantDeclaration(currentVisibility, false)
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
				fields := p.parseRecordFieldDeclarations(currentVisibility)
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
				constant := p.parseClassConstantDeclaration(currentVisibility, true)
				if constant != nil {
					recordDecl.Constants = append(recordDecl.Constants, constant)
				}
				cursor = p.cursor.Advance()
				p.cursor = cursor
				continue
			} else if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE {
				// Class method
				method := p.parseFunctionDeclaration()
				if method != nil {
					method.IsClassMethod = true
					recordDecl.Methods = append(recordDecl.Methods, method)
				}
				cursor = p.cursor.Advance()
				p.cursor = cursor
				continue
			} else if cursor.Current().Type == lexer.PROPERTY {
				// Class property: class property Name: Type read X write Y;
				prop := p.parseRecordPropertyDeclaration()
				if prop != nil {
					prop.IsClassProperty = true
					recordDecl.Properties = append(recordDecl.Properties, *prop)
					p.addRecordAutoPropertyBackingField(recordDecl, prop, currentVisibility)
				}
				cursor = p.cursor.Advance()
				p.cursor = cursor
				continue
			} else {
				p.addError("expected 'var', 'const', 'function', 'procedure' or 'property' after 'class' keyword in record", ErrUnexpectedToken)
				cursor = cursor.Advance()
				p.cursor = cursor
				continue
			}
		}

		// Check for method declarations (instance methods)
		if cursor.Current().Type == lexer.FUNCTION || cursor.Current().Type == lexer.PROCEDURE {
			errorCount := len(p.errors)
			method := p.parseFunctionDeclaration()
			if method != nil {
				recordDecl.Methods = append(recordDecl.Methods, method)
				seenMethod = true
			}
			if len(p.errors) > errorCount && method != nil && method.Body == nil {
				firstErr := p.errors[errorCount]
				p.addParserErrorAt(firstErr.Pos, firstErr.Length, "Record fields must be declared before record methods", ErrUnexpectedToken)
				p.synchronize([]lexer.TokenType{lexer.END, lexer.EOF})
				return currentVisibility
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Check for property declarations
		if cursor.Current().Type == lexer.PROPERTY {
			prop := p.parseRecordPropertyDeclaration()
			if prop != nil {
				recordDecl.Properties = append(recordDecl.Properties, *prop)
				p.addRecordAutoPropertyBackingField(recordDecl, prop, currentVisibility)
			}
			cursor = p.cursor.Advance()
			p.cursor = cursor
			continue
		}

		// Parse field declaration(s)
		if seenMethod && cursor.Current().Type == lexer.IDENT {
			p.addError("Record fields must be declared before record methods", ErrUnexpectedToken)
			p.synchronize([]lexer.TokenType{lexer.END, lexer.EOF})
			cursor = p.cursor
			continue
		}

		fields := p.parseRecordFieldDeclarations(currentVisibility)
		if fields != nil {
			recordDecl.Fields = append(recordDecl.Fields, fields...)
		}

		cursor = p.cursor.Advance()
		p.cursor = cursor
	}

	return currentVisibility
}

// parseRecordFieldDeclarations parses one or more field declarations (dispatcher).

// Pattern: Name1, Name2, Name3: Type;
// OR: Name := Value; (type inferred from initializer)
// Returns a slice of FieldDecl, one for each field name.
// PRE: cursor is field name IDENT
// POST: cursor is SEMICOLON
func (p *Parser) parseRecordFieldDeclarations(visibility ast.Visibility) []*ast.FieldDecl {
	// Use IdentifierList combinator to parse comma-separated field names
	fieldNames := p.IdentifierList(IdentifierListConfig{
		ErrorContext:      "record field declaration",
		RequireAtLeastOne: true,
	})
	if fieldNames == nil {
		return nil
	}

	cursor := p.cursor // Update cursor after IdentifierList
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
		initValue = p.parseExpression(LOWEST)
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
		fieldType = p.parseTypeExpression()
		if fieldType == nil {
			return nil
		}

		// Parse optional field initializer
		initValue = p.parseFieldInitializer(fieldNames)
		cursor = p.cursor // Update cursor after parseFieldInitializer
	}

	// Expect semicolon
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		if cursor.Peek(1).Type == lexer.END {
			p.cursor = cursor
			return fieldsFromRecordFieldNames(fieldNames, fieldType, initValue, visibility)
		}
		p.addError("expected ';' after field declaration", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	return fieldsFromRecordFieldNames(fieldNames, fieldType, initValue, visibility)
}

func fieldsFromRecordFieldNames(
	fieldNames []*ast.Identifier,
	fieldType ast.TypeExpression,
	initValue ast.Expression,
	visibility ast.Visibility,
) []*ast.FieldDecl {
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

// parseRecordPropertyWriteClause parses a parenthesized record property write
// specifier. Mirrors parsePropertyWriteClause for classes.
//
// PRE: cursor is LPAREN
// POST: cursor is RPAREN
func (p *Parser) parseRecordPropertyWriteClause(prop *ast.RecordPropertyDecl) bool {
	writeToken := p.cursor.Current()

	p.nextToken() // move into parentheses, to the lvalue start
	lhs := p.parseExpression(LOWEST)
	if lhs == nil {
		return false
	}

	writeStmt, writeSpec := p.buildPropertyWriteSpec(lhs, writeToken)
	prop.WriteStmt = writeStmt
	if identExpr, ok := writeSpec.(*ast.Identifier); ok {
		prop.WriteField = identExpr.Value
	}

	return p.expectPeek(lexer.RPAREN)
}

// parseRecordPropertyDeclaration parses a record property declaration (dispatcher).

// Pattern: property Name: Type read FieldName write FieldName;
// Also supports array properties: property Name[Index: Type]: Type read GetMethod;
//
// Note: This is different from class properties (parsePropertyDeclaration)
// PRE: cursor is PROPERTY
// POST: cursor is SEMICOLON
//
// addRecordAutoPropertyBackingField synthesizes the backing member for a
// field-less (auto) record property such as `property Alpha: Integer;`. The
// parser has already pointed the property's read/write specifiers at `F<Name>`;
// here we add the matching member so the analyzer and runtime have real storage.
// An instance property backs onto a field, a class property onto a class var.
// Record counterpart of Parser.addAutoPropertyBackingField.
//
//nolint:gocyclo // Property parser handling multiple directives
func (p *Parser) addRecordAutoPropertyBackingField(recordDecl *ast.RecordDecl, property *ast.RecordPropertyDecl, visibility ast.Visibility) {
	if property == nil || !property.IsAutoProperty || property.Name == nil || property.Type == nil {
		return
	}
	backingName := "F" + property.Name.Value
	// Do not duplicate a member the user declared explicitly. Match on storage
	// kind too: an existing F<Name> of the opposite kind must not suppress the
	// backing member this property actually needs.
	existing := recordDecl.Fields
	if property.IsClassProperty {
		existing = recordDecl.ClassVars
	}
	for _, f := range existing {
		if f != nil && f.Name != nil && pkgident.Equal(f.Name.Value, backingName) {
			return
		}
	}
	field := &ast.FieldDecl{
		Name: &ast.Identifier{
			BaseNode: ast.BaseNode{Token: property.Name.Token},
			Value:    backingName,
		},
		Type:       property.Type,
		Visibility: visibility,
		IsClassVar: property.IsClassProperty,
	}
	field.Token = property.Name.Token
	if property.IsClassProperty {
		recordDecl.ClassVars = append(recordDecl.ClassVars, field)
	} else {
		recordDecl.Fields = append(recordDecl.Fields, field)
	}
}

func (p *Parser) parseRecordPropertyDeclaration() *ast.RecordPropertyDecl {
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
		BaseNode: ast.BaseNode{
			Token: cursor.Current(),
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
				BaseNode: ast.BaseNode{Token: cursor.Current()},
				Value:    cursor.Current().Literal,
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
			paramType := p.parseTypeExpression()
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
	propType := p.parseTypeExpression()
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

	// Parse optional 'external' clause: property P : T external 'name' read F;
	// The external name replaces the property name in JSON serialization.
	if cursor.Peek(1).Type == lexer.EXTERNAL {
		cursor = cursor.Advance() // move to 'external'
		p.cursor = cursor
		prop.IsExternal = true
		if cursor.Peek(1).Type == lexer.STRING {
			cursor = cursor.Advance() // move to the name literal
			p.cursor = cursor
			prop.ExternalName = cursor.Current().Literal
		}
	}

	// Parse optional 'read' clause
	if cursor.Peek(1).Type == lexer.READ {
		cursor = cursor.Advance() // move to 'read'
		switch cursor.Peek(1).Type {
		case lexer.LPAREN:
			cursor = cursor.Advance() // move to '('
			p.cursor = cursor
			// parseExpression at '(' consumes the whole grouped expression and
			// leaves the cursor on the closing ')'.
			readExpr := p.parseExpression(LOWEST)
			if readExpr == nil {
				return nil
			}
			// A single-identifier expression behaves like a plain field/method read.
			if identExpr, ok := readExpr.(*ast.Identifier); ok {
				prop.ReadField = identExpr.Value
			} else {
				prop.ReadExpr = readExpr
			}
			cursor = p.cursor
		case lexer.IDENT:
			cursor = cursor.Advance() // move to identifier
			p.cursor = cursor
			prop.ReadField = cursor.Current().Literal
		default:
			p.addError("expected identifier after 'read'", ErrExpectedIdent)
			return nil
		}
	}

	// Parse optional 'write' clause
	if cursor.Peek(1).Type == lexer.WRITE {
		cursor = cursor.Advance() // move to 'write'
		switch cursor.Peek(1).Type {
		case lexer.LPAREN:
			cursor = cursor.Advance() // move to '('
			p.cursor = cursor
			if !p.parseRecordPropertyWriteClause(prop) {
				return nil
			}
			cursor = p.cursor
		case lexer.IDENT:
			cursor = cursor.Advance() // move to identifier
			p.cursor = cursor
			prop.WriteField = cursor.Current().Literal
		default:
			p.addError("expected identifier after 'write'", ErrExpectedIdent)
			return nil
		}
	}

	// A property declared without read/write specifiers is an auto-property: it
	// reads and writes a compiler-synthesized backing member named F<Name>.
	// Mirrors the class-side desugaring in parsePropertyDeclaration.
	if prop.ReadField == "" && prop.WriteField == "" && prop.ReadExpr == nil && prop.WriteStmt == nil {
		backingName := "F" + propName.Value
		prop.ReadField = backingName
		prop.WriteField = backingName
		prop.IsAutoProperty = true
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

	// Then the optional `deprecated ['msg'];` directive, which may follow either
	// the property's own semicolon or the one after `default`.
	if cursor.Peek(1).Type == lexer.DEPRECATED {
		cursor = cursor.Advance() // move to 'deprecated'
		p.cursor = cursor
		prop.IsDeprecated = true
		if cursor.Peek(1).Type == lexer.STRING {
			cursor = cursor.Advance() // move to the message literal
			p.cursor = cursor
			prop.DeprecatedMessage = cursor.Current().Literal
		}
		if cursor.Peek(1).Type != lexer.SEMICOLON {
			p.addError("expected ';' after 'deprecated'", ErrMissingSemicolon)
			return nil
		}
		cursor = cursor.Advance() // move to SEMICOLON
		p.cursor = cursor
	}

	return prop
}
