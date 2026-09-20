package parser

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// expectedTokenSentences holds the tokens DWScript spells out in its own way: a
// delimiter is quoted, and a name, a type or a punctuation mark with a name of its
// own gets that name. Every other token falls through to the rules in
// expectedSentence.
var expectedTokenSentences = map[lexer.TokenType]string{
	lexer.IDENT:     "Name expected",
	lexer.COLON:     `Colon ":" expected`,
	lexer.DOT:       `Dot "." expected`,
	lexer.LPAREN:    `"(" expected`,
	lexer.RPAREN:    `")" expected`,
	lexer.LBRACK:    `"[" expected`,
	lexer.RBRACK:    `"]" expected`,
	lexer.LBRACE:    `"{" expected`,
	lexer.RBRACE:    `"}" expected`,
	lexer.SEMICOLON: `";" expected`,
	lexer.COMMA:     `"," expected`,
	lexer.EQ:        `"=" expected`,
	lexer.ASSIGN:    `":=" expected`,
	lexer.LESS:      `"<" expected`,
	lexer.GREATER:   `">" expected`,
	lexer.DOTDOT:    `".." expected`,
	// The one keyword DWScript quotes.
	lexer.USES: `"USES" expected`,
}

// expectedSentence returns DWScript's wording for a missing token: a delimiter is
// quoted (`")" expected`), a name or a type has its own sentence, and a keyword is
// spelled in upper case (`DO expected`).
func expectedSentence(t lexer.TokenType) string {
	if sentence, ok := expectedTokenSentences[t]; ok {
		return sentence
	}
	if t.IsKeyword() {
		return strings.ToUpper(t.String()) + " expected"
	}
	return t.String() + " expected"
}

// foundToken returns the token DWScript anchors an "X expected" diagnostic at: the
// token found in place of X, which is the one after the cursor. Once the input has
// run out the anchor is the last token of the input, the way DWScript's tokenizer
// keeps its hot position there.
func (p *Parser) foundToken() lexer.Token {
	return p.anchorFor(p.cursor.Peek(1))
}

// anchorFor applies the end-of-input rule to a token found where another was
// expected: EOF is never an anchor, the last real token is.
func (p *Parser) anchorFor(tok lexer.Token) lexer.Token {
	if tok.Type == lexer.EOF {
		return p.cursor.LastToken()
	}
	return tok
}

// addExpected records DWScript's "X expected" error for a missing token, anchored at
// the token found instead (the one after the cursor). The cursor is left alone: the
// caller decides whether to carry on as if the token were present.
func (p *Parser) addExpected(t lexer.TokenType) {
	p.addExpectedAt(p.foundToken(), t)
}

// addExpectedStop records the same diagnostic as a compiler stop, for the places
// DWScript raises with AddCompilerStop.
func (p *Parser) addExpectedStop(t lexer.TokenType) {
	p.addExpectedStopAt(p.foundToken(), t)
}

// addExpectedCurrent records the diagnostic when the cursor already sits on the
// token found in place of the expected one.
func (p *Parser) addExpectedCurrent(t lexer.TokenType) {
	p.addExpectedAt(p.anchorFor(p.cursor.Current()), t)
}

// addExpectedStopCurrent is addExpectedCurrent as a compiler stop.
func (p *Parser) addExpectedStopCurrent(t lexer.TokenType) {
	p.addExpectedStopAt(p.anchorFor(p.cursor.Current()), t)
}

// addExpectedAt records the diagnostic at an explicit anchor, the token found in
// place of the expected one; the end-of-input rule applies to it too.
func (p *Parser) addExpectedAt(found lexer.Token, t lexer.TokenType) {
	anchor := p.anchorFor(found)
	p.recordError(NewParserError(anchor.Pos, anchor.Length(), expectedSentence(t), getErrorCodeForMissingToken(t)))
}

// addExpectedStopAt records the diagnostic at an explicit anchor as a compiler stop.
func (p *Parser) addExpectedStopAt(found lexer.Token, t lexer.TokenType) {
	anchor := p.anchorFor(found)
	p.recordStop(NewParserError(anchor.Pos, anchor.Length(), expectedSentence(t), getErrorCodeForMissingToken(t)))
}

// addTypeExpected records DWScript's "Type expected" for a missing type, anchored
// at the token found instead.
func (p *Parser) addTypeExpectedAt(found lexer.Token) {
	anchor := p.anchorFor(found)
	p.recordError(NewParserError(anchor.Pos, anchor.Length(), "Type expected", ErrExpectedType))
}

// foundTokenDescription renders a token the way DWScript names it in
// `"end" expected but "else" found`: an identifier is described, anything else is
// quoted as written.
func foundTokenDescription(tok lexer.Token) string {
	if tok.Type == lexer.IDENT {
		return "identifier"
	}
	return `"` + tok.Literal + `"`
}

// addExpectedButFound records DWScript's block-closer diagnostic, a compiler stop:
// `"end" expected but "else" found`, with the closers the block accepts listed as
// `"ensure" or "end"`.
func (p *Parser) addExpectedButFound(closers []string, found lexer.Token) {
	quoted := make([]string, len(closers))
	for i, closer := range closers {
		quoted[i] = `"` + closer + `"`
	}
	msg := strings.Join(quoted, " or ") + " expected but " + foundTokenDescription(found) + " found"
	p.recordStop(NewParserError(found.Pos, found.Length(), msg, ErrMissingEnd))
}

// canStartTypeExpression reports whether a token can begin a type expression, for
// the places that must say "Type expected" before reading one.
func (p *Parser) canStartTypeExpression(t lexer.TokenType) bool {
	if p.isIdentifierToken(t) {
		return true
	}
	switch t {
	case lexer.ARRAY, lexer.RECORD, lexer.SET, lexer.CLASS, lexer.INTERFACE,
		lexer.PROCEDURE, lexer.FUNCTION, lexer.CONSTRUCTOR, lexer.LPAREN, lexer.STRING:
		return true
	}
	return false
}

// cannotStartStatement reports the tokens DWScript's block loop refuses outright:
// a block closer that is not the block's own, or a token no statement begins with.
func cannotStartStatement(t lexer.TokenType) bool {
	switch t {
	case lexer.ELSE, lexer.THEN, lexer.DO, lexer.OF, lexer.UNTIL, lexer.EXCEPT, lexer.FINALLY,
		lexer.RPAREN, lexer.RBRACK, lexer.DOT, lexer.INITIALIZATION, lexer.FINALIZATION:
		return true
	}
	return false
}

// closerSet describes the tokens that may end a statement block, for the
// `"end" expected but "else" found` diagnostic.
type closerSet struct {
	names  []string
	tokens []lexer.TokenType
}

var (
	closersEnd             = closerSet{names: []string{"end"}, tokens: []lexer.TokenType{lexer.END}}
	closersEnsureEnd       = closerSet{names: []string{"ensure", "end"}, tokens: []lexer.TokenType{lexer.ENSURE, lexer.END}}
	closersUntil           = closerSet{names: []string{"until"}, tokens: []lexer.TokenType{lexer.UNTIL}}
	closersFinalizationEnd = closerSet{names: []string{"finalization", "end"}, tokens: []lexer.TokenType{lexer.FINALIZATION, lexer.END}}
)

func (c closerSet) accepts(t lexer.TokenType) bool {
	for _, closer := range c.tokens {
		if closer == t {
			return true
		}
	}
	return false
}

// refuseStatementStart reports and stops on a token that cannot begin a statement
// inside a block closed by closers. It returns true when it did.
func (p *Parser) refuseStatementStart(closers closerSet) bool {
	tok := p.cursor.Current()
	if !cannotStartStatement(tok.Type) || closers.accepts(tok.Type) {
		return false
	}
	p.addExpectedButFound(closers.names, tok)
	return true
}

// refuseStatementTail reports and stops when the token after a statement is neither
// ";" nor a closer of the block, the way DWScript's block loop does once a statement
// was read cleanly. lastToken is the token the statement ended on; a statement that
// consumed its own ";" is complete. It returns true when it did.
func (p *Parser) refuseStatementTail(closers closerSet, lastToken lexer.Token, errorsBefore int) bool {
	tok := p.cursor.Current()
	if lastToken.Type == lexer.SEMICOLON || tok.Type == lexer.SEMICOLON || tok.Type == lexer.EOF || closers.accepts(tok.Type) {
		return false
	}
	if tok.Type == lexer.END || tok.Type == lexer.ENSURE {
		// A closer of an enclosing block: the loop ends and the enclosing parser reports it.
		return false
	}
	if len(p.errors) != errorsBefore || p.stopped() {
		// The statement already reported something; upstream's recovery differs.
		return false
	}
	p.addExpectedButFound(closers.names, tok)
	return true
}
