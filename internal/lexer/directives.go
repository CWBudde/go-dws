// Package lexer provides lexical analysis for DWScript source code.
// This file contains compiler directive support ({$DEFINE}, {$IFDEF}, {$IF}, etc.).
package lexer

import (
	"fmt"
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
	// fromIf marks a frame opened by {$IF} rather than {$IFDEF}/{$IFNDEF}. DWScript
	// anchors an unbalanced-conditional report on the {$ELSE} of an {$IF}, but on the
	// opening directive of an {$IFDEF} (FailureScripts/conditionals_else3 vs
	// FailureScripts/invalid_switch).
	fromIf   bool
	startPos Position
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
	// off is the token's 0-based offset within the expression text, used to anchor
	// argument diagnostics on the right column.
	off int
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

// isDefined reports whether name is a conditional-compilation symbol, i.e. whether it was
// introduced by {$DEFINE} (or is one of the built-ins) and not since removed by {$UNDEF}.
//
// This is deliberately narrower than isDeclared: DWScript's Defined() answers only the
// preprocessor question, while Declared() answers the symbol-table one.
func (l *Lexer) isDefined(name string) bool {
	_, ok := l.defines[ident.Normalize(name)]
	return ok
}

// isDeclared reports whether name has been declared in the source up to this point.
//
// It consults the heuristic declaration tracker (see declarations.go) plus the integer
// constants scraped by trackConst, so that both "type"/"var"/"const"/routine names and
// dotted member names such as TObject.Create resolve. Because both are fed from the
// forward-only token stream, the answer is the point-of-use one: a name declared later in
// the file is not visible here.
func (l *Lexer) isDeclared(name string) bool {
	if l.decls.isDeclared(name) {
		return true
	}
	_, ok := l.constValues[ident.Normalize(name)]
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
//
//nolint:gocyclo // A flat switch over the compiler switches is clearer than dispatch indirection.
func (l *Lexer) processDirective() {
	startPos := l.currentPos()

	content, closePos := l.readDirectiveContent(startPos)
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
	case "endif", "ifend":
		l.handleEndIf(startPos)
	case "if":
		l.handleIf(content, parts[0], parentActive, startPos)
	case "include", "i", "include_once":
		l.handleInclude(name, content, parentActive, startPos)
	case "hint":
		l.handleMessageDirective(content, parentActive, startPos, closePos, LexerSeverityHint, "Hint", false)
	case "warning":
		l.handleMessageDirective(content, parentActive, startPos, closePos, LexerSeverityWarning, "Warning", false)
	case "error":
		l.handleMessageDirective(content, parentActive, startPos, closePos, LexerSeverityError, "Compile Error", false)
	case "fatal":
		l.handleMessageDirective(content, parentActive, startPos, closePos, LexerSeverityError, "Compile Error", true)
	case "hints", "warnings":
		l.handleSwitchToggle(name, content, parentActive, startPos, closePos)
	case "r", "resource":
		l.handleStringSwitch(content, parentActive, startPos, closePos)
	case "region", "endregion", "filter", "f":
		// Recognized and ignored: {$REGION} only structures source for editors, and
		// {$FILTER} is an include variant whose filtering is not implemented.
	default:
		// Gated on parentActive: an unknown switch inside a dead {$IFDEF} branch is
		// not reported (FailureScripts/switch_invalid3).
		if parentActive {
			l.addDirectiveDiagnostic(
				fmt.Sprintf("Compiler switch %q unknown", strings.ToUpper(name)),
				directiveNameColumn(startPos), LexerSeverityError, "")
		}
	}
}

// readDirectiveContent reads the content of a compiler directive.
func (l *Lexer) readDirectiveContent(startPos Position) (string, Position) {
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
		l.reportUnterminatedDirective(builder.String(), startPos)
		return "", l.currentPos()
	}

	// The closing brace anchors "String expected" / "ON/OFF expected" diagnostics,
	// so capture it before it is consumed.
	closePos := l.currentPos()

	// consume closing '}'
	l.readChar()

	content := strings.TrimSpace(builder.String())
	if content == "" {
		l.addError("empty compiler directive", startPos)
		return "", closePos
	}

	return content, closePos
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
		l.addDirectiveDiagnostic("Unbalanced conditional directive",
			directiveNameColumn(startPos), LexerSeverityError, "")
		return
	}
	top := &l.condStack[len(l.condStack)-1]
	if top.elseSeen {
		l.addDirectiveDiagnostic("Unfinished conditional directive",
			directiveNameColumn(startPos), LexerSeverityError, "")
		return
	}
	top.elseSeen = true
	if top.fromIf {
		top.startPos = startPos
	}
	if top.parentActive {
		top.active = !top.cond
	} else {
		top.active = false
	}
}

// handleEndIf handles {$ENDIF} directives.
func (l *Lexer) handleEndIf(startPos Position) {
	if len(l.condStack) == 0 {
		l.addDirectiveDiagnostic("Unbalanced conditional directive",
			directiveNameColumn(startPos), LexerSeverityError, "")
	} else {
		l.condStack = l.condStack[:len(l.condStack)-1]
	}
}

// handleIf handles {$IF} directives.
//
// The expression's base position is the column of the character just past the directive
// name, so that diagnostics raised while evaluating it (see evalIfExpression) land on the
// offending argument rather than on the directive.
func (l *Lexer) handleIf(content, firstPart string, parentActive bool, startPos Position) {
	base := directiveNameColumn(startPos)
	base.Column += len(firstPart)
	cond := l.evalIfExpression(strings.TrimPrefix(content, firstPart), base, parentActive)
	frame := conditionalFrame{
		cond:         cond,
		parentActive: parentActive,
		active:       parentActive && cond,
		startPos:     startPos,
		fromIf:       true,
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
	case ASSIGN, EQ:
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
// Supports: Defined(NAME), Declared(NAME), integer constants, comparisons and the
// and/or/not operators.
//
// base is the source position of offset 0 of expr, so that argument diagnostics can be
// reported at their real column. active is false when the directive sits inside an
// inactive conditional branch, in which case no diagnostic is emitted at all.
//
//nolint:gocyclo // Lexer complexity is acceptable for expression parsing
func (l *Lexer) evalIfExpression(expr string, base Position, active bool) bool {
	tokens := lexIfExpression(expr)
	pos := base
	cur := 0

	// posOf maps a $if token back to its position in the directive.
	posOf := func(t ifToken) Position {
		p := base
		p.Column += t.off
		return p
	}
	report := func(msg string, t ifToken) {
		if !active {
			return
		}
		l.addDirectiveDiagnostic(msg, posOf(t), LexerSeverityError, "")
	}

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
				query := ident.Normalize(name)
				isQuery := query == "defined" || query == "declared"
				if tok.typ != ifTokRParen {
					// The argument is not a lone name or literal, e.g.
					// Declared(IntToStr(i)). DWScript requires a constant expression.
					if isQuery {
						report("Constant expression expected", arg)
						for tok.typ != ifTokRParen && tok.typ != ifTokEOF {
							advance()
						}
						if tok.typ == ifTokRParen {
							advance()
						}
						return ifValue{kind: ifValBool, boolVal: false}
					}
					l.addError("invalid $if expression", pos)
					return ifValue{kind: ifValBool, boolVal: false}
				}
				advance()
				if !isQuery {
					return ifValue{kind: ifValBool, boolVal: false}
				}
				if arg.typ != ifTokIdent && arg.typ != ifTokString {
					// Defined(123) and friends: the argument must name a symbol.
					report("String expected", arg)
					return ifValue{kind: ifValBool, boolVal: false}
				}
				if query == "declared" {
					return ifValue{kind: ifValBool, boolVal: l.isDeclared(arg.val)}
				}
				return ifValue{kind: ifValBool, boolVal: l.isDefined(arg.val)}
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
//
// Each token records its offset within expr. The offset is counted in bytes, which equals
// the rune column DWScript reports because compiler directives are ASCII.
func lexIfExpression(expr string) []ifToken {
	var tokens []ifToken
	reader := strings.NewReader(expr)

	for {
		off := len(expr) - reader.Len()
		ch, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		if unicode.IsSpace(ch) {
			continue
		}

		tok := lexIfToken(ch, reader)
		if tok.typ != ifTokEOF {
			tok.off = off
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
	// Safe to ignore error: we just successfully read a rune
	_ = reader.UnreadRune() //nolint:errcheck

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
	// Safe to ignore error: we just successfully read a rune
	_ = reader.UnreadRune() //nolint:errcheck

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
				// Safe to ignore error: we just successfully read a rune
				_ = reader.UnreadRune() //nolint:errcheck
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
				// Safe to ignore error: we just successfully read a rune
				_ = reader.UnreadRune() //nolint:errcheck
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
