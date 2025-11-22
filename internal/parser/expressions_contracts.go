package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseCondition parses a single contract condition.
// Syntax: boolean_expression [: "error message"]
// Returns a Condition node with the test expression and optional custom message.
// PRE: cursor is first token of condition expression
// POST: cursor is last token of condition (message STRING or test expression)
func (p *Parser) parseCondition() *ast.Condition {
	builder := p.StartNode()

	startToken := p.cursor.Current()

	// Parse the test expression (should be boolean, but type checking is done in semantic phase)
	testExpr := p.parseExpression(LOWEST)
	if testExpr == nil {
		return nil
	}

	condition := &ast.Condition{
		BaseNode: ast.BaseNode{Token: startToken},
		Test:     testExpr,
	}

	// Check for optional custom message: : "message"
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume the colon

		// Expect a string literal for the error message
		if !p.expectPeek(lexer.STRING) {
			p.addError("expected string literal after ':' in contract condition", ErrUnexpectedToken)
			return nil
		}

		msgToken := p.cursor.Current()

		condition.Message = &ast.StringLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: msgToken,
				},
			},
			Value: msgToken.Literal,
		}
		// EndPos is the end of the message string literal
		return builder.Finish(condition).(*ast.Condition)
	} else {
		// EndPos is the end of the test expression
		return builder.FinishWithNode(condition, testExpr).(*ast.Condition)
	}
}

// parseOldExpression parses an 'old' expression for contract postconditions.
// The 'old' keyword refers to the value of a variable before function execution (in postconditions).
// DWScript syntax: old identifier
// The 'old' keyword can only be used in postconditions to reference pre-execution values.
// PRE: cursor is on OLD token
// POST: cursor is on identifier
func (p *Parser) parseOldExpression() ast.Expression {
	currentToken := p.cursor.Current() // the OLD token

	// Validate that we're in a postcondition context
	// Use new context API (Task 2.1.2) instead of direct field access
	if !p.ctx.ParsingPostCondition() {
		msg := fmt.Sprintf("'old' keyword can only be used in postconditions at line %d, column %d",
			currentToken.Pos.Line, currentToken.Pos.Column)
		p.addError(msg, ErrInvalidSyntax)
		return nil
	}

	// Expect an identifier after 'old'
	nextToken := p.cursor.Peek(1)
	if nextToken.Type != lexer.IDENT {
		p.addError("expected identifier after 'old'", ErrExpectedIdent)
		return nil
	}
	p.cursor = p.cursor.Advance() // move to identifier
	identToken := p.cursor.Current()

	identifier := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: identToken,
			},
		},
		Value: identToken.Literal,
	}

	return &ast.OldExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token:  currentToken,
				EndPos: identifier.End(),
			},
		},
		Identifier: identifier,
	}
}

// parsePreConditions parses function preconditions (require block).
// Syntax: require condition1; condition2; ...
// Returns a PreConditions node containing all parsed conditions.
// PRE: cursor is REQUIRE
// POST: cursor is last token of last condition
func (p *Parser) parsePreConditions() *ast.PreConditions {
	builder := p.StartNode()

	requireToken := p.cursor.Current()

	// Advance to the first condition
	p.nextToken()

	var conditions []*ast.Condition

	// Parse first condition
	condition := p.parseCondition()
	if condition == nil {
		p.addError("expected at least one condition after 'require'", ErrInvalidExpression)
		return nil
	}
	conditions = append(conditions, condition)

	// Parse additional conditions separated by semicolons
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume the semicolon

		// Check if we've reached the end of preconditions (peek at next token)
		// (beginning of var/const/begin or postconditions or EOF)
		if p.peekTokenIs(lexer.VAR) || p.peekTokenIs(lexer.CONST) ||
			p.peekTokenIs(lexer.BEGIN) || p.peekTokenIs(lexer.ENSURE) ||
			p.peekTokenIs(lexer.EOF) {
			break
		}

		p.nextToken() // move to the next condition

		condition := p.parseCondition()
		if condition == nil {
			break
		}
		conditions = append(conditions, condition)
	}

	preConditions := &ast.PreConditions{
		BaseNode:   ast.BaseNode{Token: requireToken},
		Conditions: conditions,
	}

	// EndPos is the end of the last condition
	if len(conditions) > 0 {
		return builder.FinishWithNode(preConditions, conditions[len(conditions)-1]).(*ast.PreConditions)
	}

	return preConditions
}

// parsePostConditions parses function postconditions (ensure block).
// Syntax: ensure condition1; condition2; ...
// Returns a PostConditions node containing all parsed conditions.
// Sets parsingPostCondition flag to enable 'old' keyword parsing.
// PRE: cursor is ENSURE
// POST: cursor is last token of last condition
func (p *Parser) parsePostConditions() *ast.PostConditions {
	builder := p.StartNode()

	ensureToken := p.cursor.Current()

	// Enable 'old' keyword parsing
	// Synchronize both old field and new context (Task 2.1.2)
	p.parsingPostCondition = true
	p.ctx.SetParsingPostCondition(true)
	defer func() {
		p.parsingPostCondition = false
		p.ctx.SetParsingPostCondition(false)
	}()

	// Advance to the first condition
	p.nextToken()

	var conditions []*ast.Condition

	// Parse first condition
	condition := p.parseCondition()
	if condition == nil {
		p.addError("expected at least one condition after 'ensure'", ErrInvalidExpression)
		return nil
	}
	conditions = append(conditions, condition)

	// Parse additional conditions separated by semicolons
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume the semicolon

		// Check if we've reached the end of postconditions (peek at next token)
		// (next function/procedure/type/begin/end/etc. or EOF)
		if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.PROCEDURE) ||
			p.peekTokenIs(lexer.TYPE) || p.peekTokenIs(lexer.VAR) ||
			p.peekTokenIs(lexer.CONST) || p.peekTokenIs(lexer.BEGIN) ||
			p.peekTokenIs(lexer.END) || p.peekTokenIs(lexer.IMPLEMENTATION) ||
			p.peekTokenIs(lexer.EOF) {
			break
		}

		p.nextToken() // move to the next condition

		condition := p.parseCondition()
		if condition == nil {
			break
		}
		conditions = append(conditions, condition)
	}

	postConditions := &ast.PostConditions{
		BaseNode:   ast.BaseNode{Token: ensureToken},
		Conditions: conditions,
	}

	// EndPos is the end of the last condition
	if len(conditions) > 0 {
		return builder.FinishWithNode(postConditions, conditions[len(conditions)-1]).(*ast.PostConditions)
	}

	return postConditions
}

// parseInvariantClause parses class invariants (invariant block).
// Syntax: invariants condition1; condition2; ...
// Returns an InvariantClause node containing all parsed conditions.
// PRE: cursor is INVARIANTS
// POST: cursor is last token of last condition
func (p *Parser) parseInvariantClause() *ast.InvariantClause {
	builder := p.StartNode()

	invariantToken := p.cursor.Current()

	// Advance to the first condition
	p.nextToken()

	var conditions []*ast.Condition

	// Parse first condition
	condition := p.parseCondition()
	if condition == nil {
		p.addError("expected at least one condition after 'invariants'", ErrInvalidExpression)
		return nil
	}
	conditions = append(conditions, condition)

	// Parse additional conditions separated by semicolons
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume the semicolon

		// Check if we've reached the end of invariants (peek at next token)
		// (next class member keyword, end, or EOF)
		if p.peekTokenIs(lexer.PRIVATE) || p.peekTokenIs(lexer.PROTECTED) ||
			p.peekTokenIs(lexer.PUBLIC) || p.peekTokenIs(lexer.FUNCTION) ||
			p.peekTokenIs(lexer.PROCEDURE) || p.peekTokenIs(lexer.CONSTRUCTOR) ||
			p.peekTokenIs(lexer.DESTRUCTOR) || p.peekTokenIs(lexer.PROPERTY) ||
			p.peekTokenIs(lexer.CLASS) || p.peekTokenIs(lexer.CONST) ||
			p.peekTokenIs(lexer.END) || p.peekTokenIs(lexer.EOF) {
			break
		}

		p.nextToken() // move to the next condition

		condition := p.parseCondition()
		if condition == nil {
			break
		}
		conditions = append(conditions, condition)
	}

	invariantClause := &ast.InvariantClause{
		BaseNode:   ast.BaseNode{Token: invariantToken},
		Conditions: conditions,
	}

	// EndPos is the end of the last condition
	if len(conditions) > 0 {
		return builder.FinishWithNode(invariantClause, conditions[len(conditions)-1]).(*ast.InvariantClause)
	}

	return invariantClause
}
