package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// isCallingConvention checks if a string is a calling convention keyword.
// Calling conventions are contextual identifiers, not reserved keywords,
// so they're tokenized as IDENT by the lexer.
func isCallingConvention(literal string) bool {
	return ident.Equal(literal, "register") || ident.Equal(literal, "pascal") ||
		ident.Equal(literal, "cdecl") || ident.Equal(literal, "safecall") ||
		ident.Equal(literal, "stdcall") || ident.Equal(literal, "fastcall") ||
		ident.Equal(literal, "reference")
}

// parseFunctionLocalDeclarations parses local VAR and CONST declarations inside a function.
// PRE: cursor position such that Peek(1) might be VAR or CONST
// POST: cursor is at last token of last declaration
func (p *Parser) parseFunctionLocalDeclarations(fn *ast.FunctionDecl) bool {
	cursor := p.cursor

	for cursor.Peek(1).Type == lexer.VAR || cursor.Peek(1).Type == lexer.CONST {
		cursor = cursor.Advance() // move to 'var' or 'const'
		p.cursor = cursor

		if cursor.Current().Type == lexer.VAR {
			// Parse multiple variable declarations
			for {
				if cursor.Current().Type == lexer.VAR {
					if cursor.Peek(1).Type != lexer.IDENT {
						break
					}
				} else if cursor.Current().Type != lexer.IDENT {
					break
				}

				varDecl := p.parseVarDeclaration()
				if varDecl == nil {
					break
				}
				cursor = p.cursor

				if fn.Body == nil {
					fn.Body = &ast.BlockStatement{
						BaseNode: ast.BaseNode{Token: cursor.Current()},
					}
				}

				if blockStmt, ok := varDecl.(*ast.BlockStatement); ok && p.isVarDeclBlock(blockStmt) {
					fn.Body.Statements = append(fn.Body.Statements, blockStmt.Statements...)
				} else {
					fn.Body.Statements = append(fn.Body.Statements, varDecl)
				}

				if cursor.Peek(1).Type != lexer.IDENT {
					break
				}

				cursor = cursor.Advance() // move to next identifier
				p.cursor = cursor
			}
		} else if cursor.Current().Type == lexer.CONST {
			constDecl := p.parseConstDeclaration()
			if constDecl != nil {
				if fn.Body == nil {
					fn.Body = &ast.BlockStatement{
						BaseNode: ast.BaseNode{Token: cursor.Current()},
					}
				}
				fn.Body.Statements = append(fn.Body.Statements, constDecl)
			}
			cursor = p.cursor
		}
	}

	return true
}

// parseFunctionDirectives parses function/procedure directives (static, virtual, override, etc.).
// PRE: cursor is at semicolon after function signature
// POST: cursor is at last semicolon after last directive
func (p *Parser) parseFunctionDirectives(fn *ast.FunctionDecl) bool {
	cursor := p.cursor

	for {
		nextTok := cursor.Peek(1)

		if nextTok.Type == lexer.STATIC {
			cursor = cursor.Advance() // move to 'static'
			p.cursor = cursor

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after static", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.VIRTUAL {
			cursor = cursor.Advance() // move to 'virtual'
			p.cursor = cursor
			fn.IsVirtual = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after virtual", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.OVERRIDE {
			cursor = cursor.Advance() // move to 'override'
			p.cursor = cursor
			fn.IsOverride = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after override", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.REINTRODUCE {
			cursor = cursor.Advance() // move to 'reintroduce'
			p.cursor = cursor
			fn.IsReintroduce = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after reintroduce", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.DEFAULT {
			cursor = cursor.Advance() // move to 'default'
			p.cursor = cursor
			fn.IsDefault = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after default", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.ABSTRACT {
			cursor = cursor.Advance() // move to 'abstract'
			p.cursor = cursor
			fn.IsAbstract = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after abstract", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.EXTERNAL {
			cursor = cursor.Advance() // move to 'external'
			p.cursor = cursor
			fn.IsExternal = true

			// Check for optional external name string
			if cursor.Peek(1).Type == lexer.STRING {
				cursor = cursor.Advance() // move to string
				p.cursor = cursor
				fn.ExternalName = cursor.Current().Literal
			}

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after external", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.OVERLOAD {
			cursor = cursor.Advance() // move to 'overload'
			p.cursor = cursor
			fn.IsOverload = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after overload", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.FORWARD {
			cursor = cursor.Advance() // move to 'forward'
			p.cursor = cursor
			fn.IsForward = true

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after forward", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.IDENT && isCallingConvention(nextTok.Literal) {
			cursor = cursor.Advance() // move to calling convention
			p.cursor = cursor
			fn.CallingConvention = ident.Normalize(cursor.Current().Literal)
			fn.CallingConventionPos = cursor.Current().Pos

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after calling convention", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else if nextTok.Type == lexer.DEPRECATED {
			cursor = cursor.Advance() // move to 'deprecated'
			p.cursor = cursor
			fn.IsDeprecated = true

			// Check for optional deprecation message
			if cursor.Peek(1).Type == lexer.STRING {
				cursor = cursor.Advance() // move to string
				p.cursor = cursor
				fn.DeprecatedMessage = cursor.Current().Literal
			}

			if cursor.Peek(1).Type != lexer.SEMICOLON {
				p.addError("expected ';' after deprecated", ErrMissingSemicolon)
				return false
			}
			cursor = cursor.Advance() // move to SEMICOLON
			p.cursor = cursor

		} else {
			break // No more directives
		}
	}

	return true
}

// parseFunctionReturnType parses and converts a return type expression to TypeAnnotation.
// PRE: cursor is at type expression
// POST: cursor is at last token of type expression
func (p *Parser) parseFunctionReturnType() *ast.TypeAnnotation {
	typeExpr := p.parseTypeExpression()
	if typeExpr == nil {
		p.addError("expected return type after ':'", ErrExpectedType)
		return nil
	}

	// Convert TypeExpression to TypeAnnotation
	switch te := typeExpr.(type) {
	case *ast.TypeAnnotation:
		return te
	case *ast.FunctionPointerTypeNode:
		return &ast.TypeAnnotation{
			Token: te.Token,
			Name:  te.String(),
		}
	case *ast.SetTypeNode:
		return &ast.TypeAnnotation{
			Token: te.Token,
			Name:  te.String(),
		}
	case *ast.ArrayTypeNode:
		if te == nil {
			p.addError("array type expression is nil in return type", ErrInvalidType)
			return nil
		}
		token := te.Token
		if token.Type == 0 || token.Literal == "" {
			token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
		}
		return &ast.TypeAnnotation{
			Token: token,
			Name:  te.String(),
		}
	default:
		p.addError("unsupported type expression in return type", ErrInvalidType)
		return nil
	}
}

// parseFunctionQualifiedName parses a function name, which may be qualified (ClassName.MethodName).
// PRE: cursor is at function/procedure name
// POST: cursor is at function name (last identifier)
func (p *Parser) parseFunctionQualifiedName() (name, className *ast.Identifier) {
	cursor := p.cursor

	firstIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}

	// Check for qualified name (ClassName.MethodName for method implementations)
	if cursor.Peek(1).Type == lexer.DOT {
		cursor = cursor.Advance() // move to '.'
		cursor = cursor.Advance() // move past '.'
		p.cursor = cursor

		className = firstIdent
		name = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}
	} else {
		// Even for simple names, we're already positioned at it
		name = firstIdent
		className = nil
		// p.cursor is already at the name, no need to update
	}

	return name, className
}

// Syntax: function Name(params): Type; begin ... end;
//
//	procedure Name(params); begin ... end;
//
// PRE: cursor is FUNCTION or PROCEDURE
// POST: cursor is END or SEMICOLON (forward declaration) or last token of body
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDecl {
	cursor := p.cursor
	builder := p.StartNode()

	fn := &ast.FunctionDecl{
		BaseNode: ast.BaseNode{
			Token: cursor.Current(),
		},
	}

	// Parse function name (may be qualified: ClassName.MethodName)
	cursor = cursor.Advance() // move to name
	p.cursor = cursor

	fn.Name, fn.ClassName = p.parseFunctionQualifiedName()
	cursor = p.cursor // reload cursor after parsing qualified name

	// Parse parameter list (if present)
	if cursor.Peek(1).Type == lexer.LPAREN {
		cursor = cursor.Advance() // move to '('
		p.cursor = cursor

		fn.Parameters = p.parseParameterList()
		cursor = p.cursor

		if cursor.Current().Type != lexer.RPAREN {
			p.addError("expected ')' after parameter list", ErrMissingRParen)
			return nil
		}
	}

	// Parse return type for functions (not procedures)
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move past ':' to type expression
		p.cursor = cursor

		returnType := p.parseFunctionReturnType()
		if returnType == nil {
			return nil
		}
		fn.ReturnType = returnType
		cursor = p.cursor
	}

	// Expect semicolon after signature
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after function signature", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// Parse directives (static, virtual, override, etc.)
	if !p.parseFunctionDirectives(fn) {
		return nil
	}
	cursor = p.cursor

	// Parse preconditions (require block) if present
	if cursor.Peek(1).Type == lexer.REQUIRE {
		cursor = cursor.Advance() // move to REQUIRE
		p.cursor = cursor
		fn.PreConditions = p.parsePreConditions()
		cursor = p.cursor
	}

	// Check if this is a forward declaration (no body)
	nextTok := cursor.Peek(1)
	if fn.IsForward || (nextTok.Type != lexer.BEGIN && nextTok.Type != lexer.VAR && nextTok.Type != lexer.CONST && nextTok.Type != lexer.REQUIRE) {
		return builder.Finish(fn).(*ast.FunctionDecl)
	}

	// Parse local variable/constant declarations
	if !p.parseFunctionLocalDeclarations(fn) {
		return nil
	}
	cursor = p.cursor

	// Parse function body (begin...end block)
	if cursor.Peek(1).Type != lexer.BEGIN {
		p.addError("expected 'begin' for function body", ErrUnexpectedToken)
		return nil
	}
	cursor = cursor.Advance() // move to BEGIN
	p.cursor = cursor

	bodyBlock := p.parseBlockStatement()
	cursor = p.cursor

	if bodyBlock != nil {
		if fn.Body == nil {
			fn.Body = bodyBlock
		} else {
			fn.Body.Statements = append(fn.Body.Statements, bodyBlock.Statements...)
		}
	}

	// Check if we stopped at ENSURE inside the begin...end block
	if cursor.Current().Type == lexer.ENSURE {
		fn.PostConditions = p.parsePostConditions()
		cursor = p.cursor

		// Skip semicolons
		for cursor.Current().Type == lexer.SEMICOLON {
			cursor = cursor.Advance()
			p.cursor = cursor
		}

		// Advance to END if not already there
		if cursor.Current().Type != lexer.END {
			if cursor.Peek(1).Type != lexer.END {
				p.addError("expected 'end' after postconditions", ErrUnexpectedToken)
				return nil
			}
			cursor = cursor.Advance() // move to END
			p.cursor = cursor
		}
	}

	// Expect semicolon after end
	if cursor.Peek(1).Type != lexer.SEMICOLON {
		p.addError("expected ';' after 'end'", ErrMissingSemicolon)
		return nil
	}
	cursor = cursor.Advance() // move to SEMICOLON
	p.cursor = cursor

	// Parse postconditions after end (if present and not already parsed inline)
	if cursor.Peek(1).Type == lexer.ENSURE {
		if fn.PostConditions != nil {
			p.addError("postconditions already defined inline; cannot define them again after 'end'", ErrInvalidSyntax)
			return nil
		}
		cursor = cursor.Advance() // move to ENSURE
		p.cursor = cursor

		fn.PostConditions = p.parsePostConditions()
		if fn.PostConditions != nil {
			return builder.FinishWithNode(fn, fn.PostConditions).(*ast.FunctionDecl)
		} else {
			return builder.Finish(fn).(*ast.FunctionDecl)
		}
	} else {
		return builder.Finish(fn).(*ast.FunctionDecl)
	}
}

// parseParameterList parses a function parameter list.
// Syntax: (param: Type; var param: Type; a, b, c: Type)
// PRE: cursor is LPAREN
// POST: cursor is RPAREN

// Syntax: (param: Type; var param: Type; a, b, c: Type)
// PRE: cursor is at LPAREN
// POST: cursor is at RPAREN
func (p *Parser) parseParameterList() []*ast.Parameter {
	cursor := p.cursor
	params := []*ast.Parameter{}

	cursor = cursor.Advance() // move past '('
	p.cursor = cursor

	// Check for empty parameter list
	if cursor.Current().Type == lexer.RPAREN {
		return params
	}

	// Parse first parameter group
	groupParams := p.parseParameterGroup()
	if groupParams == nil {
		return nil
	}
	params = append(params, groupParams...)

	// Update cursor after parameter group parsing
	cursor = p.cursor

	// Parse remaining parameter groups separated by semicolons
	for cursor.Peek(1).Type == lexer.SEMICOLON {
		cursor = cursor.Advance() // move to ';'
		cursor = cursor.Advance() // move past ';'
		p.cursor = cursor

		groupParams = p.parseParameterGroup()
		if groupParams == nil {
			return nil
		}
		params = append(params, groupParams...)

		// Update cursor after parameter group parsing
		cursor = p.cursor
	}

	// Expect closing parenthesis
	if cursor.Peek(1).Type != lexer.RPAREN {
		p.addError("expected ')' after parameter list", ErrMissingRParen)
		return nil
	}
	cursor = cursor.Advance() // move to RPAREN
	p.cursor = cursor

	return params
}

// parseParameterGroup parses a group of parameters with the same type.
// Syntax: name: Type  or  name1, name2, name3: Type  or  var name: Type  or  lazy name: Type  or  const name: Type
// PRE: cursor is VAR, CONST, LAZY, or first parameter name IDENT
// POST: cursor is last token of type expression or default value

// Syntax: name: Type  or  name1, name2, name3: Type  or  var name: Type  or  lazy name: Type  or  const name: Type
// PRE: cursor is at VAR, CONST, LAZY, or first parameter name IDENT
// POST: cursor is at last token of type expression or default value
func (p *Parser) parseParameterGroup() []*ast.Parameter {
	cursor := p.cursor
	params := []*ast.Parameter{}

	// Parse optional modifiers
	isConst := false
	isLazy := false
	byRef := false

	// Check for 'const' keyword (pass by const-reference)
	if cursor.Current().Type == lexer.CONST {
		isConst = true
		cursor = cursor.Advance() // move past 'const'
		p.cursor = cursor
	}

	// Check for 'lazy' keyword (expression capture)
	if cursor.Current().Type == lexer.LAZY {
		isLazy = true
		cursor = cursor.Advance() // move past 'lazy'
		p.cursor = cursor
	}

	// Check for 'var' keyword (pass by reference)
	if cursor.Current().Type == lexer.VAR {
		byRef = true
		cursor = cursor.Advance() // move past 'var'
		p.cursor = cursor
	}

	// Check for mutually exclusive modifiers
	if (isLazy && byRef) || (isConst && byRef) || (isConst && isLazy) {
		err := NewStructuredError(ErrKindInvalid).
			WithCode(ErrInvalidSyntax).
			WithMessage("parameter modifiers are mutually exclusive").
			WithPosition(cursor.Current().Pos, cursor.Current().Length()).
			WithSuggestion("use only one of: var, const, or lazy").
			WithParsePhase("function parameter").
			Build()
		p.addStructuredError(err)
		return nil
	}

	// Parse identifier list (name1, name2, name3)
	names := []*ast.Identifier{}

	// First identifier (allow contextual keywords like STEP)
	if !p.isIdentifierToken(cursor.Current().Type) {
		p.addError("expected parameter name", ErrExpectedIdent)
		return nil
	}

	name := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: cursor.Current(),
			},
		},
		Value: cursor.Current().Literal,
	}
	names = append(names, name)

	// Additional identifiers separated by commas
	for cursor.Peek(1).Type == lexer.COMMA {
		cursor = cursor.Advance() // move to comma
		cursor = cursor.Advance() // move past comma
		p.cursor = cursor

		if !p.isIdentifierToken(cursor.Current().Type) {
			p.addError("expected parameter name after ','", ErrExpectedIdent)
			return nil
		}

		name = &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: cursor.Current(),
				},
			},
			Value: cursor.Current().Literal,
		}
		names = append(names, name)
	}

	// Check for ':' and type (optional for lambda parameters)
	var typeExpr ast.TypeExpression
	if cursor.Peek(1).Type == lexer.COLON {
		cursor = cursor.Advance() // move to ':'
		cursor = cursor.Advance() // move past ':' to type expression
		p.cursor = cursor

		typeExpr = p.parseTypeExpression()
		if typeExpr == nil {
			// Error already reported by parseTypeExpression
			return nil
		}

		// Update cursor after type parsing
		cursor = p.cursor
	}
	// If no colon, typeExpr remains nil (valid for lambda parameters)

	// Parse optional default value
	var defaultValue ast.Expression
	if cursor.Peek(1).Type == lexer.EQ {
		// Validate that optional parameters don't have modifiers (lazy, var, const)
		if isLazy || byRef || isConst {
			err := NewStructuredError(ErrKindInvalid).
				WithCode(ErrInvalidSyntax).
				WithMessage("optional parameters cannot have lazy, var, or const modifiers").
				WithPosition(cursor.Current().Pos, cursor.Current().Length()).
				WithSuggestion("remove the modifier or remove the default value").
				WithParsePhase("function parameter").
				Build()
			p.addStructuredError(err)
			return nil
		}

		cursor = cursor.Advance() // move to '='
		cursor = cursor.Advance() // move past '='
		p.cursor = cursor

		defaultValue = p.parseExpression(LOWEST)
		if defaultValue == nil {
			err := NewStructuredError(ErrKindMissing).
				WithCode(ErrInvalidExpression).
				WithMessage("expected default value expression after '='").
				WithPosition(cursor.Current().Pos, cursor.Current().Length()).
				WithExpectedString("expression").
				WithParsePhase("function parameter").
				Build()
			p.addStructuredError(err)
			return nil
		}

		// Update cursor after default value parsing
		cursor = p.cursor
	}

	// Create parameter nodes for each name
	for _, name := range names {
		param := &ast.Parameter{
			Token:        name.Token,
			Name:         name,
			Type:         typeExpr,
			ByRef:        byRef,
			IsConst:      isConst,
			IsLazy:       isLazy,
			DefaultValue: defaultValue,
		}
		params = append(params, param)
	}

	return params
}

// positioned at the first parameter token (not at LPAREN).
// This is a wrapper used by function pointer type parsing.
// Syntax: name: Type; name2: Type; ...
// PRE: cursor is first parameter token (VAR, CONST, LAZY, or IDENT)
// POST: cursor is RPAREN
func (p *Parser) parseParameterListAtToken() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Parse first parameter group (we're already at first token)
	groupParams := p.parseParameterGroup()
	if groupParams == nil {
		return nil
	}
	params = append(params, groupParams...)

	// Update cursor after parameter group parsing
	cursor := p.cursor

	// Parse remaining parameter groups separated by semicolons
	for cursor.Peek(1).Type == lexer.SEMICOLON {
		cursor = cursor.Advance() // move to ';'
		cursor = cursor.Advance() // move past ';'
		p.cursor = cursor

		groupParams = p.parseParameterGroup()
		if groupParams == nil {
			return nil
		}
		params = append(params, groupParams...)

		// Update cursor after parameter group parsing
		cursor = p.cursor
	}

	// Expect closing parenthesis
	if cursor.Peek(1).Type != lexer.RPAREN {
		p.addError("expected ')' after parameter list", ErrMissingRParen)
		return nil
	}
	cursor = cursor.Advance() // move to RPAREN
	p.cursor = cursor

	return params
}

// Used for shorthand function pointer syntax: function(Integer, String): Boolean
//
// Syntax:
//   - function(Integer): Boolean                  - single param
//   - function(Integer, String): Boolean          - comma-separated
//   - function(Integer; String; Boolean): Float   - semicolon-separated
//
// The parser is positioned at the first type token when this is called.
// Parameters will have nil Name fields.
//
// This format is used in type declarations but not in actual function definitions.
// Example: type TFunc = function(Integer, String): Boolean;
// PRE: cursor is first type token or modifier (CONST, VAR, LAZY)
// POST: cursor is RPAREN
func (p *Parser) parseTypeOnlyParameterListAtToken() []*ast.Parameter {
	params := []*ast.Parameter{}
	cursor := p.cursor

	// Current token is first type
	for {
		// Check for modifiers (const, var, lazy)
		isConst := false
		isLazy := false
		byRef := false

		if cursor.Current().Type == lexer.CONST {
			isConst = true
			cursor = cursor.Advance()
			p.cursor = cursor
		}
		if cursor.Current().Type == lexer.LAZY {
			isLazy = true
			cursor = cursor.Advance()
			p.cursor = cursor
		}
		if cursor.Current().Type == lexer.VAR {
			byRef = true
			cursor = cursor.Advance()
			p.cursor = cursor
		}

		// Parse type expression (could be complex like "array of Integer" or "function(Integer): Integer")
		typeExpr := p.parseTypeExpression()
		if typeExpr == nil {
			p.addError("expected type in function pointer parameter list", ErrExpectedType)
			return nil
		}

		// Update cursor after type expression parsing
		cursor = p.cursor

		// Convert TypeExpression to TypeAnnotation
		var typeAnnotation *ast.TypeAnnotation
		switch te := typeExpr.(type) {
		case *ast.TypeAnnotation:
			typeAnnotation = te
		case *ast.FunctionPointerTypeNode:
			// For nested function pointers, use the string representation as type name
			typeAnnotation = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		case *ast.ArrayTypeNode:
			// For array types, use string representation
			token := te.Token
			if token.Type == 0 || token.Literal == "" {
				token = lexer.Token{Type: lexer.ARRAY, Literal: "array", Pos: lexer.Position{}}
			}
			typeAnnotation = &ast.TypeAnnotation{
				Token: token,
				Name:  te.String(),
			}
		case *ast.SetTypeNode:
			// For set types, use string representation
			typeAnnotation = &ast.TypeAnnotation{
				Token: te.Token,
				Name:  te.String(),
			}
		default:
			p.addError("unsupported type expression in function pointer parameter", ErrInvalidType)
			return nil
		}

		// Create parameter with nil name (shorthand syntax)
		param := &ast.Parameter{
			Token:   typeAnnotation.Token,
			Name:    nil, // Shorthand syntax has no parameter names
			Type:    typeAnnotation,
			IsLazy:  isLazy,
			ByRef:   byRef,
			IsConst: isConst,
		}
		params = append(params, param)

		// Check what comes next
		if cursor.Peek(1).Type == lexer.COMMA {
			// More parameters in same group
			cursor = cursor.Advance() // move to comma
			cursor = cursor.Advance() // move past comma to next type
			p.cursor = cursor
			continue
		} else if cursor.Peek(1).Type == lexer.SEMICOLON {
			// Next parameter group
			cursor = cursor.Advance() // move to semicolon
			cursor = cursor.Advance() // move past semicolon to next type
			p.cursor = cursor
			continue
		} else if cursor.Peek(1).Type == lexer.RPAREN {
			// End of parameter list
			cursor = cursor.Advance() // move to RPAREN
			p.cursor = cursor
			break
		} else {
			p.addError("expected ',', ';', or ')' in function pointer parameter list", ErrUnexpectedToken)
			return nil
		}
	}

	return params
}
