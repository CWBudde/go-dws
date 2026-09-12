package parser

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// parseAnonymousRecordExpression parses DWScript's anonymous record constructor
// expression:
//
//	record a := 1; b := 'x'; end
//	record "i*i" := i * i; "2i" := 2 * i; end
//	record Field := 123 end
//
// Field names may be identifiers or string literals; a quoted name is kept
// verbatim so that names which are not valid identifiers survive into
// serialization. The trailing semicolon before 'end' is optional.
//
// This is distinct from parseGroupedExpression's record literal, which is
// written with parentheses and takes its type from context. The form parsed
// here is structurally typed and stands alone as an expression.
//
// PRE: cursor is RECORD
// POST: cursor is END
func (p *Parser) parseAnonymousRecordExpression() ast.Expression {
	recordToken := p.cursor.Current()

	recordExpr := &ast.AnonymousRecordExpression{
		BaseNode: ast.BaseNode{Token: recordToken},
		Fields:   []*ast.FieldInitializer{},
	}

	// Move past 'record' to the first field name (or straight to 'end')
	p.cursor = p.cursor.Advance()

	for {
		current := p.cursor.Current()
		if current.Type == lexer.END {
			break
		}
		if current.Type == lexer.EOF {
			p.addError("expected 'end' to close record expression", ErrMissingEnd)
			return nil
		}

		field := p.parseAnonymousRecordField()
		if field == nil {
			return nil
		}
		recordExpr.Fields = append(recordExpr.Fields, field)

		// After a field the cursor sits on the last token of its value.
		next := p.cursor.Peek(1)
		switch next.Type {
		case lexer.SEMICOLON:
			p.cursor = p.cursor.Advance() // move to ';'
			p.cursor = p.cursor.Advance() // move to next field name or 'end'
		case lexer.END:
			p.cursor = p.cursor.Advance() // move to 'end'
		default:
			p.addError(fmt.Sprintf("expected ';' or 'end' in record expression, got %s", next.Type), ErrUnexpectedToken)
			return nil
		}
	}

	recordExpr.EndPos = p.cursor.Current().End()

	return recordExpr
}

// parseAnonymousRecordField parses a single `name := value` pair of an anonymous
// record expression.
//
// PRE: cursor is the field name (IDENT or STRING)
// POST: cursor is the last token of the field's value expression
func (p *Parser) parseAnonymousRecordField() *ast.FieldInitializer {
	nameToken := p.cursor.Current()

	if nameToken.Type != lexer.IDENT && nameToken.Type != lexer.STRING {
		p.addError(fmt.Sprintf("expected record field name, got %s", nameToken.Type), ErrUnexpectedToken)
		return nil
	}

	fieldName := &ast.Identifier{
		BaseNode: ast.BaseNode{
			Token:  nameToken,
			EndPos: p.endPosFromToken(nameToken),
		},
		Value: nameToken.Literal,
	}

	if p.cursor.Peek(1).Type != lexer.ASSIGN {
		p.addError(fmt.Sprintf("expected ':=' after record field name '%s'", fieldName.Value), ErrUnexpectedToken)
		return nil
	}

	p.cursor = p.cursor.Advance() // move to ':='
	p.cursor = p.cursor.Advance() // move to the value expression

	value := p.parseExpression(LOWEST)
	if isInvalidExpression(value) {
		return nil
	}

	return &ast.FieldInitializer{
		BaseNode: ast.BaseNode{
			Token:  nameToken,
			EndPos: value.End(),
		},
		Name:  fieldName,
		Value: value,
	}
}
