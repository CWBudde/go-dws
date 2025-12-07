// Package lexer provides lexical analysis for DWScript source code.
// This file contains compiler directive support ({$DEFINE}, {$IFDEF}, {$IF}, etc.).
package lexer

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// conditionalFrame represents a single level in the conditional compilation stack.
// It tracks whether the current conditional block is active and whether we've seen an {$ELSE}.
type conditionalFrame struct {
	cond         bool
	active       bool
	parentActive bool
	elseSeen     bool
	startPos     Position
}

// ifTokenType represents token types for $if expression evaluation.
type ifTokenType int

const (
	ifTokEOF ifTokenType = iota
	ifTokIdent
	ifTokInt
	ifTokString
	ifTokLParen
	ifTokRParen
	ifTokEq
	ifTokNeq
	ifTokLt
	ifTokLte
	ifTokGt
	ifTokGte
	ifTokAnd
	ifTokOr
	ifTokNot
)

// ifToken represents a token in a $if expression.
type ifToken struct {
	val string
	typ ifTokenType
}

// ifValKind represents the kind of value in a $if expression.
type ifValKind int

const (
	ifValBool ifValKind = iota
	ifValInt
	ifValString
)

// ifValue represents a value during $if expression evaluation.
type ifValue struct {
	strVal  string
	intVal  int
	kind    ifValKind
	boolVal bool
}

// asBool converts an ifValue to a boolean for truthiness testing.
func (v ifValue) asBool() bool {
	switch v.kind {
	case ifValBool:
		return v.boolVal
	case ifValInt:
		return v.intVal != 0
	case ifValString:
		return v.strVal != ""
	default:
		return false
	}
}

// isDefined checks if a symbol is defined (either via {$DEFINE} or as a const).
func (l *Lexer) isDefined(name string) bool {
	_, ok := l.defines[ident.Normalize(name)]
	if ok {
		return true
	}
	_, ok = l.constValues[ident.Normalize(name)]
	return ok
}

// define adds a symbol to the defines map.
func (l *Lexer) define(name string) {
	if name == "" {
		return
	}
	l.defines[ident.Normalize(name)] = struct{}{}
}

// undefine removes a symbol from the defines map.
func (l *Lexer) undefine(name string) {
	delete(l.defines, ident.Normalize(name))
}

// isSkippingTokens returns true if we're inside an inactive conditional block.
func (l *Lexer) isSkippingTokens() bool {
	for _, frame := range l.condStack {
		if !frame.active {
			return true
		}
	}
	return false
}

// processDirective handles compiler directives like {$DEFINE}, {$IFDEF}, {$IF}, etc.
func (l *Lexer) processDirective() {
	startPos := l.currentPos()

	content := l.readDirectiveContent(startPos)
	if content == "" {
		return // error already reported
	}

	parts := strings.Fields(content)
	name := strings.ToLower(parts[0])
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}

	parentActive := !l.isSkippingTokens()

	switch name {
	case "define":
		l.handleDefine(arg, parentActive, startPos)
	case "undef":
		l.handleUndef(arg, parentActive, startPos)
	case "ifdef", "ifndef":
		l.handleIfDef(name, arg, parentActive, startPos)
	case "else":
		l.handleElse(startPos)
	case "endif":
		l.handleEndIf(startPos)
	case "if":
		l.handleIf(content, parts[0], parentActive, startPos)
	default:
		l.addError("unknown compiler directive: "+name, startPos)
	}
}

// readDirectiveContent reads the content of a compiler directive.
func (l *Lexer) readDirectiveContent(startPos Position) string {
	// Consume "{$"
	l.readChar() // '{'
	l.readChar() // '$'

	var builder strings.Builder
	for l.ch != 0 && l.ch != '}' {
		builder.WriteRune(l.ch)
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	if l.ch == 0 {
		l.addError("unterminated compiler directive", startPos)
		return ""
	}

	// consume closing '}'
	l.readChar()

	content := strings.TrimSpace(builder.String())
	if content == "" {
		l.addError("empty compiler directive", startPos)
		return ""
	}

	return content
}

// handleDefine handles {$DEFINE} directives.
func (l *Lexer) handleDefine(arg string, parentActive bool, startPos Position) {
	if arg == "" {
		l.addError("name expected after $define", startPos)
		return
	}
	if parentActive {
		l.define(arg)
	}
}

// handleUndef handles {$UNDEF} directives.
func (l *Lexer) handleUndef(arg string, parentActive bool, startPos Position) {
	if arg == "" {
		l.addError("name expected after $undef", startPos)
		return
	}
	if parentActive {
		l.undefine(arg)
	}
}

// handleIfDef handles {$IFDEF} and {$IFNDEF} directives.
func (l *Lexer) handleIfDef(name, arg string, parentActive bool, startPos Position) {
	if arg == "" {
		l.addError("name expected after $"+name, startPos)
		return
	}
	cond := l.isDefined(arg)
	if name == "ifndef" {
		cond = !cond
	}
	frame := conditionalFrame{
		cond:         cond,
		parentActive: parentActive,
		active:       parentActive && cond,
		startPos:     startPos,
	}
	l.condStack = append(l.condStack, frame)
}

// handleElse handles {$ELSE} directives.
func (l *Lexer) handleElse(startPos Position) {
	if len(l.condStack) == 0 {
		l.addError("unbalanced conditional directive", startPos)
		return
	}
	top := &l.condStack[len(l.condStack)-1]
	if top.elseSeen {
		l.addError("unfinished conditional directive", startPos)
		return
	}
	top.elseSeen = true
	if top.parentActive {
		top.active = !top.cond
	} else {
		top.active = false
	}
}

// handleEndIf handles {$ENDIF} directives.
func (l *Lexer) handleEndIf(startPos Position) {
	if len(l.condStack) == 0 {
		l.addError("unbalanced conditional directive", startPos)
	} else {
		l.condStack = l.condStack[:len(l.condStack)-1]
	}
}

// handleIf handles {$IF} directives.
func (l *Lexer) handleIf(content, firstPart string, parentActive bool, startPos Position) {
	cond := l.evalIfExpression(strings.TrimPrefix(content, firstPart))
	frame := conditionalFrame{
		cond:         cond,
		parentActive: parentActive,
		active:       parentActive && cond,
		startPos:     startPos,
	}
	l.condStack = append(l.condStack, frame)
}

// trackConst tracks constant declarations for use in $if expressions.
// It monitors the token stream to identify const declarations and their integer values.
func (l *Lexer) trackConst(tok Token) {
	// skip tracking when inside skipped directive block
	if l.isSkippingTokens() {
		return
	}

	switch tok.Type {
	case CONST:
		l.enterConstBlock()
	case SEMICOLON:
		l.resetConstTracking()
	case VAR, TYPE, FUNCTION, PROCEDURE, CLASS, RECORD, UNIT, IMPLEMENTATION, INTERFACE, BEGIN:
		l.exitConstBlock()
	case IDENT:
		l.handleConstIdent(tok.Literal)
	case COLON:
		// ignore
	case ASSIGN:
		l.handleConstAssign()
	case INT:
		l.handleConstInt(tok.Literal)
	default:
		l.handleConstOther()
	}
}

// enterConstBlock marks the beginning of a const block.
func (l *Lexer) enterConstBlock() {
	l.constBlock = true
	l.constPending = ""
	l.constWait = false
}

// exitConstBlock marks the end of a const block.
func (l *Lexer) exitConstBlock() {
	l.constBlock = false
	l.constPending = ""
	l.constWait = false
}

// resetConstTracking resets the current constant being tracked.
func (l *Lexer) resetConstTracking() {
	l.constPending = ""
	l.constWait = false
}

// handleConstIdent handles identifier tokens in const blocks.
func (l *Lexer) handleConstIdent(literal string) {
	if l.constBlock && l.constPending == "" && !l.constWait {
		l.constPending = literal
	}
}

// handleConstAssign handles assignment operators in const blocks.
func (l *Lexer) handleConstAssign() {
	if l.constBlock && l.constPending != "" {
		l.constWait = true
	}
}

// handleConstInt handles integer literals in const blocks.
func (l *Lexer) handleConstInt(literal string) {
	if l.constBlock && l.constPending != "" && l.constWait {
		if v, err := strconv.Atoi(literal); err == nil {
			l.constValues[ident.Normalize(l.constPending)] = v
		}
		l.resetConstTracking()
	}
}

// handleConstOther handles other tokens in const blocks.
func (l *Lexer) handleConstOther() {
	if l.constBlock {
		l.resetConstTracking()
	}
}

// evalIfExpression evaluates a $if compiler directive expression.
// Supports: defined(NAME), integer constants, comparisons, and/or/not operators.
func (l *Lexer) evalIfExpression(expr string) bool {
	tokens := lexIfExpression(expr)
	pos := Position{}
	cur := 0

	next := func() ifToken {
		if cur >= len(tokens) {
			return ifToken{typ: ifTokEOF}
		}
		tok := tokens[cur]
		cur++
		return tok
	}

	var tok ifToken
	advance := func() { tok = next() }

	var parseExpr func() bool
	var parseAnd func() bool
	var parseUnary func() bool

	var parsePrimary = func() ifValue {
		switch tok.typ {
		case ifTokInt:
			val := tok.val
			advance()
			if v, err := strconv.Atoi(val); err == nil {
				return ifValue{kind: ifValInt, intVal: v}
			}
			return ifValue{kind: ifValBool, boolVal: false}
		case ifTokString:
			val := tok.val
			advance()
			return ifValue{kind: ifValString, strVal: val}
		case ifTokIdent:
			name := tok.val
			advance()
			if tok.typ == ifTokLParen {
				advance()
				arg := tok
				advance()
				if tok.typ != ifTokRParen {
					l.addError("invalid $if expression", pos)
					return ifValue{kind: ifValBool, boolVal: false}
				}
				advance()
				switch strings.ToLower(name) {
				case "defined", "declared":
					if arg.typ == ifTokIdent || arg.typ == ifTokString {
						return ifValue{kind: ifValBool, boolVal: l.isDefined(arg.val)}
					}
					return ifValue{kind: ifValBool, boolVal: false}
				default:
					return ifValue{kind: ifValBool, boolVal: false}
				}
			}
			if v, ok := l.constValues[ident.Normalize(name)]; ok {
				return ifValue{kind: ifValInt, intVal: v}
			}
			return ifValue{kind: ifValBool, boolVal: l.isDefined(name)}
		case ifTokLParen:
			advance()
			val := parseExpr()
			if tok.typ != ifTokRParen {
				l.addError("invalid $if expression", pos)
				return ifValue{kind: ifValBool, boolVal: false}
			}
			advance()
			return ifValue{kind: ifValBool, boolVal: val}
		default:
			l.addError("invalid $if expression", pos)
			return ifValue{kind: ifValBool, boolVal: false}
		}
	}

	parseEquality := func() bool {
		left := parsePrimary()
		for tok.typ == ifTokEq || tok.typ == ifTokNeq || tok.typ == ifTokLt || tok.typ == ifTokLte || tok.typ == ifTokGt || tok.typ == ifTokGte {
			op := tok.typ
			advance()
			right := parsePrimary()
			left = ifValue{kind: ifValBool, boolVal: compareValues(op, left, right)}
		}
		return left.asBool()
	}

	parseUnary = func() bool {
		if tok.typ == ifTokNot {
			advance()
			return !parseUnary()
		}
		return parseEquality()
	}

	parseAnd = func() bool {
		left := parseUnary()
		for tok.typ == ifTokAnd {
			advance()
			right := parseUnary()
			left = left && right
		}
		return left
	}

	parseExpr = func() bool {
		left := parseAnd()
		for tok.typ == ifTokOr {
			advance()
			right := parseAnd()
			left = left || right
		}
		return left
	}

	advance()
	result := parseExpr()
	return result
}

// compareValues compares two ifValues using the specified operator.
func compareValues(op ifTokenType, left, right ifValue) bool {
	// integer comparison
	if left.kind == ifValInt && right.kind == ifValInt {
		return compareInts(op, left.intVal, right.intVal)
	}

	// string comparison
	if left.kind == ifValString && right.kind == ifValString {
		return compareStrings(op, left.strVal, right.strVal)
	}

	// fallback boolean truthiness
	return compareBools(op, left.asBool(), right.asBool())
}

// compareInts compares two integers using the specified operator.
func compareInts(op ifTokenType, left, right int) bool {
	switch op {
	case ifTokEq:
		return left == right
	case ifTokNeq:
		return left != right
	case ifTokLt:
		return left < right
	case ifTokLte:
		return left <= right
	case ifTokGt:
		return left > right
	case ifTokGte:
		return left >= right
	default:
		return false
	}
}

// compareStrings compares two strings using the specified operator.
func compareStrings(op ifTokenType, left, right string) bool {
	switch op {
	case ifTokEq:
		return left == right
	case ifTokNeq:
		return left != right
	case ifTokLt:
		return left < right
	case ifTokLte:
		return left <= right
	case ifTokGt:
		return left > right
	case ifTokGte:
		return left >= right
	default:
		return false
	}
}

// compareBools compares two booleans using the specified operator.
func compareBools(op ifTokenType, left, right bool) bool {
	switch op {
	case ifTokEq:
		return left == right
	case ifTokNeq:
		return left != right
	case ifTokLt:
		return !left && right
	case ifTokLte:
		return (!left && right) || left == right
	case ifTokGt:
		return left && !right
	case ifTokGte:
		return (left && !right) || left == right
	default:
		return false
	}
}

// lexIfExpression tokenizes a $if expression string into tokens.
func lexIfExpression(expr string) []ifToken {
	var tokens []ifToken
	reader := strings.NewReader(expr)

	for {
		ch, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		if unicode.IsSpace(ch) {
			continue
		}

		tok := lexIfToken(ch, reader)
		if tok.typ != ifTokEOF {
			tokens = append(tokens, tok)
		}
	}

	tokens = append(tokens, ifToken{typ: ifTokEOF})
	return tokens
}

// lexIfToken lexes a single token from the $if expression.
func lexIfToken(ch rune, reader *strings.Reader) ifToken {
	switch ch {
	case '(':
		return ifToken{typ: ifTokLParen}
	case ')':
		return ifToken{typ: ifTokRParen}
	case '=':
		return ifToken{typ: ifTokEq}
	case '<':
		return lexIfLessThan(reader)
	case '>':
		return lexIfGreaterThan(reader)
	case '\'', '"':
		return lexIfString(ch, reader)
	default:
		if isDigit(ch) {
			return lexIfNumber(ch, reader)
		}
		if isLetter(ch) {
			return lexIfIdentOrKeyword(ch, reader)
		}
		return ifToken{typ: ifTokEOF} // skip unknown characters
	}
}

// lexIfLessThan lexes '<', '<=', or '<>' operators.
func lexIfLessThan(reader *strings.Reader) ifToken {
	next, _, err := reader.ReadRune()
	if err != nil {
		return ifToken{typ: ifTokLt}
	}
	if next == '>' {
		return ifToken{typ: ifTokNeq}
	}
	if next == '=' {
		return ifToken{typ: ifTokLte}
	}
	_ = reader.UnreadRune()

	return ifToken{typ: ifTokLt}
}

// lexIfGreaterThan lexes '>' or '>=' operators.
func lexIfGreaterThan(reader *strings.Reader) ifToken {
	next, _, err := reader.ReadRune()
	if err != nil {
		return ifToken{typ: ifTokGt}
	}
	if next == '=' {
		return ifToken{typ: ifTokGte}
	}
	_ = reader.UnreadRune()

	return ifToken{typ: ifTokGt}
}

// lexIfString lexes a string literal.
func lexIfString(quote rune, reader *strings.Reader) ifToken {
	var b strings.Builder
	for {
		r, _, err := reader.ReadRune()
		if err != nil || r == quote {
			break
		}
		b.WriteRune(r)
	}
	return ifToken{typ: ifTokString, val: b.String()}
}

// lexIfNumber lexes an integer literal.
func lexIfNumber(first rune, reader *strings.Reader) ifToken {
	var builder strings.Builder
	builder.WriteRune(first)
	for {
		r, _, err := reader.ReadRune()
		if err != nil || !isDigit(r) {
			if err == nil {
				_ = reader.UnreadRune()
			}
			break
		}
		builder.WriteRune(r)
	}
	return ifToken{typ: ifTokInt, val: builder.String()}
}

// lexIfIdentOrKeyword lexes an identifier or keyword.
func lexIfIdentOrKeyword(first rune, reader *strings.Reader) ifToken {
	var builder strings.Builder
	builder.WriteRune(first)
	for {
		r, _, err := reader.ReadRune()
		if err != nil || (!isLetter(r) && !isDigit(r)) {
			if err == nil {
				_ = reader.UnreadRune()
			}
			break
		}
		builder.WriteRune(r)
	}
	word := builder.String()
	switch strings.ToLower(word) {
	case "and":
		return ifToken{typ: ifTokAnd}
	case "or":
		return ifToken{typ: ifTokOr}
	case "not":
		return ifToken{typ: ifTokNot}
	default:
		return ifToken{typ: ifTokIdent, val: word}
	}
}
