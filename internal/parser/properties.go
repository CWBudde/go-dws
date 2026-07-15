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

	// Check for indexed property parameters: property Items[index: Integer]
	var indexParams []*ast.Parameter
	if p.peekTokenIs(lexer.LBRACK) {
		p.nextToken() // move to '['

		// Check for empty brackets (not allowed)
		if p.peekTokenIs(lexer.RBRACK) {
			p.addError("indexed property cannot have empty parameter list", ErrInvalidSyntax)
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
				p.addError("expected ']' or ';' after indexed property parameter", ErrUnexpectedToken)
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
		return builder.Finish(prop).(*ast.PropertyDecl)
	}

	// Expect colon before type
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

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
		case p.peekTokenIs(lexer.READ):
			// Parse optional 'read' clause
			// ReadSpec can be:
			// - Identifier (field or method name)
			// - Expression in parentheses: read (FValue * 2)
			p.nextToken() // move to 'read'
			p.nextToken() // move to read specifier

			// Check if read spec is an expression in parentheses
			if p.curTokenIs(lexer.LPAREN) {
				// Parse expression-based read spec
				readExpr := p.parseExpression(LOWEST)
				prop.ReadSpec = readExpr
			} else if p.isMemberNameToken(p.cursor.Current().Type) {
				// Simple field/method name (may be a reserved word, e.g. `read Set`)
				prop.ReadSpec = &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: p.cursor.Current(),
						},
					},
					Value: p.cursor.Current().Literal,
				}
			} else {
				p.addError("expected identifier or expression after 'read'", ErrExpectedIdent)
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
					return nil
				}
			case p.isMemberNameToken(p.cursor.Current().Type):
				// Simple field/method name (may be a reserved word, e.g. `write Set`)
				prop.WriteSpec = &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: p.cursor.Current(),
						},
					},
					Value: p.cursor.Current().Literal,
				}
			default:
				p.addError("expected identifier or expression after 'write'", ErrExpectedIdent)
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
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: propName.Token,
				},
			},
			Value: backingFieldName,
		}

		// Auto-property has both read and write access to the backing field
		prop.ReadSpec = backingField
		prop.WriteSpec = backingField
		prop.IsAutoProperty = true
	}

	// Expect semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	// Parse optional 'default;' keyword
	// This comes after the semicolon: property Items[i: Integer]: String read GetItem; default;
	if p.peekTokenIs(lexer.DEFAULT) {
		p.nextToken() // move to 'default'
		prop.IsDefault = true

		// Expect another semicolon after 'default'
		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}
	}

	decl, _ := builder.Finish(prop).(*ast.PropertyDecl)

	return decl
}

// parsePropertyWriteClause parses a parenthesized property write specifier.
// Two forms are supported:
//   - lvalue:     write (FSub.Field)      -> normalized to `FSub.Field := Value`
//   - assignment: write (Field := Value)  -> stored as-is
//
// A single-identifier lvalue (write (Field)) is stored as an ordinary field/method
// write specifier so it flows through the existing field-backed write path.
//
// The special identifier `Value` refers to the value being assigned.
//
// PRE: cursor is LPAREN
// POST: cursor is RPAREN
func (p *Parser) parsePropertyWriteClause(prop *ast.PropertyDecl) bool {
	writeToken := p.cursor.Current()

	p.nextToken() // move into parentheses, to the lvalue start
	lhs := p.parseExpression(LOWEST)
	if lhs == nil {
		return false
	}

	prop.WriteStmt, prop.WriteSpec = p.buildPropertyWriteSpec(lhs, writeToken)

	// Expect closing parenthesis
	return p.expectPeek(lexer.RPAREN)
}

// buildPropertyWriteSpec turns a parsed parenthesized write specifier into either
// a write statement or a field/method write spec. Called with the cursor on the
// last token of the left-hand expression. Handles three shapes:
//   - assignment  (target := expr)         -> the assignment statement
//   - call/other  (SetField(Value div 2))  -> an expression statement
//   - plain lvalue (FSub.Field)            -> normalized to `lvalue := Value`
//   - identifier   (Field)                 -> a plain field/method write spec
func (p *Parser) buildPropertyWriteSpec(lhs ast.Expression, writeToken lexer.Token) (ast.Statement, ast.Expression) {
	if p.peekTokenIs(lexer.ASSIGN) {
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
	case *ast.Identifier:
		// Single identifier lvalue: behaves like `write Field`.
		return nil, lhs
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
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: writeToken},
				},
				Value: "Value",
			},
		}, nil
	}
}

// parseIndexedPropertyParameterGroup parses a group of indexed property parameters with the same type.
// Syntax: name: Type  or  name1, name2: Type
// Similar to parseParameterGroup but without 'var' keyword support.
// PRE: cursor is parameter name IDENT
// POST: cursor is type IDENT
func (p *Parser) parseIndexedPropertyParameterGroup() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Collect parameter names separated by commas
	names := []*ast.Identifier{}

	for {
		// Parse parameter name (can be IDENT or keyword used as identifier)
		if !p.curTokenIs(lexer.IDENT) && !p.curTokenIs(lexer.INDEX) {
			p.addError("expected parameter name in indexed property", ErrExpectedIdent)
			return nil
		}

		names = append(names, &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: p.cursor.Current(),
				},
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
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	// Parse type annotation
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	paramType := &ast.TypeAnnotation{
		Token: p.cursor.Current(),
		Name:  p.cursor.Current().Literal,
	}

	// Create parameter for each name
	for _, name := range names {
		param := &ast.Parameter{
			Token: name.Token,
			Name:  name,
			Type:  paramType,
		}
		params = append(params, param)
	}

	return params
}
