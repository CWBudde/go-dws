package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
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
// PRE: cursor is PROPERTY
// POST: cursor is SEMICOLON
//
//nolint:gocyclo // Property parser handling multiple directives and indexed params
func (p *Parser) parsePropertyDeclaration() *ast.PropertyDecl {
	builder := p.StartNode()
	propToken := p.cursor.Current() // 'property' token

	// Parse property name
	if !p.peekTokenIs(lexer.IDENT) && !p.peekTokenIs(lexer.READONLY) {
		p.addExpectedStop(lexer.IDENT)
		return nil
	}
	p.nextToken()
	propName := &ast.Identifier{
		BaseNode: ast.BaseNode{
			Token: p.cursor.Current(),
		},
		Value: p.cursor.Current().Literal,
	}

	// Check for indexed property parameters: property Items[index: Integer]
	var indexParams []*ast.Parameter
	if p.peekTokenIs(lexer.LBRACK) {
		var ok bool
		indexParams, ok = p.parsePropertyIndexParameters(false)
		if !ok {
			return nil
		}
	}

	// Promotion form: `property Prop;` (no index, no type, no accessors) —
	// redeclares an inherited property under the current visibility, inheriting
	// its type and accessors from the parent. Only valid without index params.
	if indexParams == nil && p.peekTokenIs(lexer.SEMICOLON) {
		prop := &ast.PropertyDecl{
			BaseNode:    ast.BaseNode{Token: propToken},
			Name:        propName,
			IsPromotion: true,
		}
		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}
		if result, ok := builder.Finish(prop).(*ast.PropertyDecl); ok {
			return result
		}
		return prop
	}

	// Expect colon before type
	if !p.peekTokenIs(lexer.COLON) {
		p.addExpectedStop(lexer.COLON)
		return nil
	}
	p.nextToken()

	// Parse property type via the shared type parser so composite types
	// (e.g. `array of String`) are supported, not just a bare identifier.
	p.nextToken() // move onto the type's first token
	propType := p.parseTypeExpression()
	if propType == nil {
		return nil
	}

	prop := &ast.PropertyDecl{
		BaseNode: ast.BaseNode{
			Token: propToken,
		},
		Name:        propName,
		Type:        propType,
		IndexParams: indexParams,
		IsDefault:   false,
	}

	// Track optional specifiers
	var indexValue ast.Expression

	// Parse optional index/read/write directives (order-insensitive but only one of each)
parseDirectives:
	for {
		switch {
		case p.peekTokenIs(lexer.INDEX):
			p.nextToken() // move to 'index'
			p.nextToken() // move to expression start
			if indexValue != nil {
				p.addError("duplicate index directive on property", ErrUnexpectedToken)
				return nil
			}
			indexValue = p.parseExpression(LOWEST)
		case p.peekTokenIs(lexer.EXTERNAL):
			// external 'name' renames the property for JSON serialization.
			p.nextToken() // move to 'external'
			prop.IsExternal = true
			if p.peekTokenIs(lexer.STRING) {
				p.nextToken() // move to the name literal
				prop.ExternalName = p.cursor.Current().Literal
			}
		case p.peekTokenIs(lexer.READ):
			// Parse optional 'read' clause
			// ReadSpec can be:
			// - Identifier (field or method name)
			// - Expression in parentheses: read (FValue * 2)
			p.nextToken() // move to 'read'
			p.nextToken() // move to read specifier

			// Check if read spec is an expression in parentheses
			if p.curTokenIs(lexer.LPAREN) {
				// The accessor owns its outer parentheses. Inner groups still use
				// ReadBracket's stopping expression grammar.
				prop.ReadSpec = p.parsePropertyReadExpression()
				if p.stopped() {
					builder.Finish(prop)
					return prop
				}
			} else if p.isMemberNameToken(p.cursor.Current().Type) {
				// Simple field/method name (may be a reserved word, e.g. `read Set`)
				prop.ReadSpec = &ast.Identifier{
					BaseNode: ast.BaseNode{
						Token: p.cursor.Current(),
					},
					Value: p.cursor.Current().Literal,
				}
			} else {
				p.addExpectedStopCurrent(lexer.IDENT)
				return nil
			}
		case p.peekTokenIs(lexer.WRITE):
			// Parse optional 'write' clause
			// WriteSpec can be:
			// - Identifier (field or method name)
			// - Parenthesized lvalue expression: write (FSub.Field)
			// - Parenthesized assignment statement: write (Field := Value div 2)
			p.nextToken() // move to 'write'
			p.nextToken() // move to write specifier start

			switch {
			case p.curTokenIs(lexer.LPAREN):
				if !p.parsePropertyWriteClause(prop) {
					builder.Finish(prop)
					return prop
				}
			case p.isMemberNameToken(p.cursor.Current().Type):
				// Simple field/method name (may be a reserved word, e.g. `write Set`)
				prop.WriteSpec = &ast.Identifier{
					BaseNode: ast.BaseNode{
						Token: p.cursor.Current(),
					},
					Value: p.cursor.Current().Literal,
				}
			default:
				p.addExpectedStopCurrent(lexer.IDENT)
				return nil
			}
		default:
			break parseDirectives
		}
	}

	// Attach parsed index value
	prop.IndexValue = indexValue

	// If neither read nor write was specified, generate auto-property
	// Auto-property generates backing field FName (F + property name)
	if prop.ReadSpec == nil && prop.WriteSpec == nil && prop.WriteStmt == nil {
		// Generate backing field name: F + property name
		backingFieldName := "F" + propName.Value
		backingField := &ast.Identifier{
			BaseNode: ast.BaseNode{
				Token: propName.Token,
			},
			Value: backingFieldName,
		}

		// Auto-property has both read and write access to the backing field
		prop.ReadSpec = backingField
		prop.WriteSpec = backingField
		prop.IsAutoProperty = true
	}

	// ReadPropertyDecl accepts a literal description before reintroduce. Leave
	// an invalid value untouched for semicolon and class-member recovery.
	if p.peekTokenIs(lexer.DESCRIPTION) {
		p.nextToken()
		if p.peekTokenIs(lexer.STRING) {
			p.nextToken()
			prop.Description = p.cursor.Current().Literal
			prop.HasDescription = true
		} else {
			anchor := p.foundToken()
			p.recordError(NewParserError(anchor.Pos, anchor.Length(), "String expected", ErrUnexpectedToken))
		}
	}

	// The compatibility marker belongs before the declaration semicolon.
	if p.peekTokenIs(lexer.REINTRODUCE) {
		p.nextToken()
		prop.IsReintroduce = true
	}

	// A missing semicolon is recoverable: the property has already been
	// declared, and the next token still belongs to the class member loop.
	if !p.expectPeek(lexer.SEMICOLON) {
		builder.Finish(prop)
		return prop
	}

	// Parse optional 'default;' keyword
	// This comes after the semicolon: property Items[i: Integer]: String read GetItem; default;
	if p.peekTokenIs(lexer.DEFAULT) {
		p.nextToken() // move to 'default'
		prop.IsDefault = true
		prop.DefaultPos = p.cursor.Peek(1).Pos

		// Expect another semicolon after 'default'
		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}
	}

	// Parse optional 'deprecated ['msg'];', which follows the semicolon and may
	// itself follow 'default;'.
	if !p.parsePropertyDeprecatedDirective(prop) {
		return nil
	}

	decl, _ := builder.Finish(prop).(*ast.PropertyDecl)

	return decl
}

// parsePropertyReadExpression reads an accessor's declaration parentheses,
// retaining the expression barrier even after an ordinary missing close.
func (p *Parser) parsePropertyReadExpression() ast.Expression {
	opening := p.cursor.Current()
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	if p.stopped() || expr == nil {
		return expr
	}
	p.expectPeek(lexer.RPAREN)
	return &ast.GroupedExpression{
		BaseNode:   ast.BaseNode{Token: opening, EndPos: p.cursor.Current().End()},
		Expression: expr,
	}
}

// parsePropertyWriteClause retains one instruction as an expression accessor.
// A missing declaration close is ordinary recovery; a child stop stays stopping.
func (p *Parser) parsePropertyWriteClause(prop *ast.PropertyDecl) bool {
	prop.WriteStmt, prop.WriteSpec, prop.WriteSourceExpression = p.parsePropertyWriteInstruction()
	return !p.stopped()
}

// parsePropertyWriteInstruction leaves an unrecognized instruction starter
// untouched, matching ReadInstr's null instruction. Empty writers remain
// writable and carry the opening token for their diagnostic anchor.
func (p *Parser) parsePropertyWriteInstruction() (ast.Statement, ast.Expression, ast.Expression) {
	opening := p.cursor.Current()
	var stmt ast.Statement
	var spec, source ast.Expression
	switch p.cursor.Peek(1).Type {
	case lexer.BEGIN, lexer.IF, lexer.WHILE, lexer.REPEAT, lexer.FOR,
		lexer.CASE, lexer.TRY, lexer.RAISE, lexer.BREAK, lexer.CONTINUE,
		lexer.EXIT, lexer.WITH:
		p.nextToken()
		stmt = p.parseStatement()
	case lexer.SEMICOLON:
		p.nextToken()
		stmt = &ast.EmptyStatement{BaseNode: ast.BaseNode{Token: opening}}
	default:
		if p.isMemberNameToken(p.cursor.Peek(1).Type) || p.peekTokenIs(lexer.LPAREN) {
			p.nextToken()
			lhs := p.parseExpression(LOWEST)
			if p.stopped() || lhs == nil {
				return nil, nil, nil
			}
			stmt, spec = p.buildPropertyWriteSpec(lhs, opening)
			if assignment, ok := stmt.(*ast.AssignmentStatement); ok && assignment.Token.Type == lexer.LPAREN {
				source = lhs
			} else if expression, ok := stmt.(*ast.ExpressionStatement); ok && expression.Expression != lhs {
				source = lhs
			}
		} else {
			stmt = &ast.EmptyStatement{BaseNode: ast.BaseNode{Token: opening}}
		}
	}
	if !p.stopped() {
		p.expectPeek(lexer.RPAREN)
	}
	return stmt, spec, source
}

// buildPropertyWriteSpec turns a parsed parenthesized write specifier into either
// a write statement. Parentheses always create an expression accessor, including
// a single field name. Called with the cursor on the
// last token of the left-hand expression. Handles three shapes:
//   - assignment  (target := expr)         -> the assignment statement
//   - call/other  (SetField(Value div 2))  -> an expression statement
//   - plain lvalue (FSub.Field)            -> normalized to `lvalue := Value`
//   - identifier   (Field)                 -> normalized to `Field := Value`
func (p *Parser) buildPropertyWriteSpec(lhs ast.Expression, writeToken lexer.Token) (ast.Statement, ast.Expression) {
	instruction := lhs
	for {
		group, ok := lhs.(*ast.GroupedExpression)
		if !ok {
			break
		}
		lhs = group.Expression
	}
	if isAssignmentOperator(p.cursor.Peek(1).Type) {
		p.nextToken() // move to ':='
		assignOp := p.cursor.Current().Type
		assignToken := p.cursor.Current()
		p.nextToken() // move to right-hand expression
		rhs := p.parseExpression(LOWEST)
		if rhs == nil {
			return nil, nil
		}
		return &ast.AssignmentStatement{
			BaseNode: ast.BaseNode{Token: assignToken},
			Target:   lhs,
			Operator: assignOp,
			Value:    rhs,
		}, nil
	}

	switch lhs.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.BooleanLiteral, *ast.CharLiteral:
		// A literal reached through an inner bracket is a real instruction;
		// retain its source term for semantic constant/null classification.
		return &ast.ExpressionStatement{BaseNode: ast.BaseNode{Token: writeToken}, Expression: instruction}, nil
	case *ast.CallExpression:
		// A call such as SetField(Value) executes directly.
		return &ast.ExpressionStatement{
			BaseNode:   ast.BaseNode{Token: writeToken},
			Expression: lhs,
		}, nil
	default:
		// General lvalue (member/index access): normalize to `lhs := Value`.
		return &ast.AssignmentStatement{
			BaseNode: ast.BaseNode{Token: writeToken},
			Target:   lhs,
			Operator: lexer.ASSIGN,
			Value: &ast.Identifier{
				BaseNode: ast.BaseNode{Token: writeToken},
				Value:    "Value",
			},
		}, nil
	}
}

// parseIndexedPropertyParameterGroup parses a group of indexed property parameters with the same type.
// Syntax: [var | const] name1, name2: Type. Each group resets the modifier.
// Defaults, lazy parameters and const(ref) belong to routine parameter grammar.
// PRE: cursor is parameter name IDENT, VAR or CONST
// POST: cursor is type IDENT
func (p *Parser) parseIndexedPropertyParameterGroup(composite bool) []*ast.Parameter {
	params := []*ast.Parameter{}
	byRef, isConst := p.curTokenIs(lexer.VAR), p.curTokenIs(lexer.CONST)
	if byRef || isConst {
		p.nextToken()
	}

	// Collect parameter names separated by commas
	names := []*ast.Identifier{}

	for {
		// Parse parameter name (can be IDENT or keyword used as identifier)
		if !p.curTokenIs(lexer.IDENT) && !p.curTokenIs(lexer.INDEX) {
			p.addExpectedStopCurrent(lexer.IDENT)
			return nil
		}

		names = append(names, &ast.Identifier{
			BaseNode: ast.BaseNode{
				Token: p.cursor.Current(),
			},
			Value: p.cursor.Current().Literal,
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
	if !p.peekTokenIs(lexer.COLON) {
		p.addExpectedStop(lexer.COLON)
		return nil
	}
	p.nextToken()

	paramType := p.parsePropertyIndexParameterType(composite)
	if paramType == nil {
		return nil
	}

	// Create parameter for each name
	for _, name := range names {
		param := &ast.Parameter{
			Token:   name.Token,
			Name:    name,
			Type:    paramType,
			ByRef:   byRef,
			IsConst: isConst,
		}
		params = append(params, param)
	}

	return params
}

// parsePropertyDeprecatedDirective parses an optional `deprecated` marker after a
// property declaration's terminating semicolon, with or without a message. It
// returns false only when the directive is present but unterminated, which is a
// parse error the caller must propagate.
func (p *Parser) parsePropertyDeprecatedDirective(prop *ast.PropertyDecl) bool {
	if !p.peekTokenIs(lexer.DEPRECATED) {
		return true
	}
	p.nextToken() // move to 'deprecated'
	prop.IsDeprecated = true

	if p.peekTokenIs(lexer.STRING) {
		p.nextToken()
		prop.DeprecatedMessage = p.cursor.Current().Literal
	}

	return p.expectPeek(lexer.SEMICOLON)
}

// parsePropertyIndexParameters shares the declaration list grammar. Records retain
// their existing composite type annotations; classes use their named type grammar.
func (p *Parser) parsePropertyIndexParameters(composite bool) ([]*ast.Parameter, bool) {
	p.nextToken() // '['
	if p.peekTokenIs(lexer.RBRACK) {
		p.nextToken()
		tok := p.cursor.Current()
		p.recordError(NewParserError(tok.Pos, tok.Length(), "Parameters expected", ErrInvalidSyntax))
		return []*ast.Parameter{}, true
	}
	p.nextToken()
	var params []*ast.Parameter
	for {
		group := p.parseIndexedPropertyParameterGroup(composite)
		if group == nil {
			return params, false
		}
		params = append(params, group...)
		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			p.nextToken()
			continue
		}
		if !p.peekTokenIs(lexer.RBRACK) {
			p.addExpectedStop(lexer.RBRACK)
			return params, false
		}
		p.nextToken()
		return params, true
	}
}

// parsePropertyIndexParameterType retains the record reader's composite grammar
// and ReadType's ordinary missing-type recovery without consuming the separator.
func (p *Parser) parsePropertyIndexParameterType(composite bool) ast.TypeExpression {
	if composite && !isTypeExpressionStartToken(p.cursor.Peek(1).Type) && !p.peekTokenIs(lexer.TYPE) {
		tok := p.cursor.Peek(1)
		p.addTypeExpectedAt(tok)
		return &ast.TypeAnnotation{Token: tok, Name: "Variant"}
	}
	p.nextToken()
	if composite {
		return p.parseTypeExpression()
	}
	if !p.curTokenIs(lexer.IDENT) {
		p.addExpectedStopCurrent(lexer.IDENT)
		return nil
	}
	return &ast.TypeAnnotation{Token: p.cursor.Current(), Name: p.cursor.Current().Literal}
}
