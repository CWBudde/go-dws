package lexer

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer represents a lexical scanner for DWScript source code.
type Lexer struct {
	input        string        // The source code being tokenized
	position     int           // Current position in input (points to current char)
	readPosition int           // Current reading position in input (after current char)
	ch           rune          // Current character under examination (0 if EOF)
	line         int           // Current line number (1-indexed)
	column       int           // Current column number (1-indexed)
	errors       []LexerError  // Accumulated lexer errors
	tokenBuffer  []Token       // Buffer for token lookahead (Task 12.3.1)
}

// lexerState represents the complete state of the lexer at a point in time.
// This allows for efficient save/restore operations during lookahead.
type lexerState struct {
	position     int  // Current position in input
	readPosition int  // Current reading position
	ch           rune // Current character
	line         int  // Current line number
	column       int  // Current column number
}

// New creates a new Lexer for the given input string.
// Automatically detects and strips UTF-8 BOM (0xEF 0xBB 0xBF) if present at the start.
// This matches the behavior of the original DWScript which strips BOMs during file reading.
func New(input string) *Lexer {
	// Strip UTF-8 BOM if present (matches DWScript behavior)
	// UTF-8 BOM: EF BB BF (239 187 191)
	if len(input) >= 3 &&
		input[0] == 0xEF &&
		input[1] == 0xBB &&
		input[2] == 0xBF {
		input = input[3:]
	}

	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar() // Initialize first character
	return l
}

// readChar advances the lexer to the next character in the input.
// Properly handles UTF-8 multi-byte sequences.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
		l.position = l.readPosition
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
	}
	l.column++
}

// peekChar returns the next character without advancing the position.
// Properly handles UTF-8 multi-byte sequences.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// peekCharN returns the character n positions ahead without advancing.
// Properly handles UTF-8 multi-byte sequences by counting runes, not bytes.
func (l *Lexer) peekCharN(n int) rune {
	pos := l.readPosition
	// Skip n-1 runes to get to the nth position
	for i := 0; i < n-1 && pos < len(l.input); i++ {
		_, size := utf8.DecodeRuneInString(l.input[pos:])
		pos += size
	}
	if pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[pos:])
	return r
}

// matchAndConsume checks if the next character matches the expected rune.
// If it matches, advances the lexer position and returns true.
// If it doesn't match, leaves the lexer position unchanged and returns false.
func (l *Lexer) matchAndConsume(expected rune) bool {
	if l.peekChar() == expected {
		l.readChar()
		return true
	}
	return false
}

// currentPos returns the current Position for token creation.
func (l *Lexer) currentPos() Position {
	return Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.position,
	}
}

// Input returns the source code being tokenized.
//
// Deprecated: Use Peek(n) for token lookahead instead of creating temporary lexers.
// This method is kept for backward compatibility but may be removed in a future version.
// See Task 12.3 for the new Peek() API.
func (l *Lexer) Input() string {
	return l.input
}

// Errors returns all accumulated lexer errors.
// This allows the parser to check for lexical errors after tokenization.
func (l *Lexer) Errors() []LexerError {
	return l.errors
}

// addError adds a new error to the lexer's error list.
// This follows the parser's pattern of accumulating errors instead of stopping at the first error.
func (l *Lexer) addError(msg string, pos Position) {
	l.errors = append(l.errors, LexerError{
		Message: msg,
		Pos:     pos,
	})
}

// saveState captures the current lexer state for later restoration.
// This is useful for lookahead operations that need to be undone.
func (l *Lexer) saveState() lexerState {
	return lexerState{
		position:     l.position,
		readPosition: l.readPosition,
		ch:           l.ch,
		line:         l.line,
		column:       l.column,
	}
}

// restoreState restores the lexer to a previously saved state.
// This is used after lookahead operations to return to the original position.
func (l *Lexer) restoreState(s lexerState) {
	l.position = s.position
	l.readPosition = s.readPosition
	l.ch = s.ch
	l.line = s.line
	l.column = s.column
}

// Peek returns the token n positions ahead without consuming it.
// Peek(0) returns the next token (same as NextToken would return).
// Peek(1) returns the token after that, etc.
// Tokens are buffered lazily as needed.
// This eliminates the need for creating temporary lexers for lookahead.
func (l *Lexer) Peek(n int) Token {
	// Ensure we have buffered enough tokens
	for len(l.tokenBuffer) <= n {
		// Generate and buffer the next token
		tok := l.nextTokenInternal()
		l.tokenBuffer = append(l.tokenBuffer, tok)
	}

	return l.tokenBuffer[n]
}

// Operator handler functions (Task 12.4.1 - Arithmetic operators)

// handlePlus handles the '+' operator and its variants (++, +=).
func (l *Lexer) handlePlus(pos Position) Token {
	if l.matchAndConsume('+') {
		tok := NewToken(INC, "++", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('=') {
		tok := NewToken(PLUS_ASSIGN, "+=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(PLUS, "+", pos)
	l.readChar()
	return tok
}

// handleMinus handles the '-' operator and its variants (--, -=).
func (l *Lexer) handleMinus(pos Position) Token {
	if l.matchAndConsume('-') {
		tok := NewToken(DEC, "--", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('=') {
		tok := NewToken(MINUS_ASSIGN, "-=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(MINUS, "-", pos)
	l.readChar()
	return tok
}

// handleAsterisk handles the '*' operator and its variants (**, *=).
func (l *Lexer) handleAsterisk(pos Position) Token {
	if l.matchAndConsume('*') {
		tok := NewToken(POWER, "**", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('=') {
		tok := NewToken(TIMES_ASSIGN, "*=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(ASTERISK, "*", pos)
	l.readChar()
	return tok
}

// handleSlash handles the '/' operator and its variants (//, /*, /=).
// Note: Comments are handled in nextTokenInternal before calling this.
func (l *Lexer) handleSlash(pos Position) Token {
	if l.matchAndConsume('=') {
		tok := NewToken(DIVIDE_ASSIGN, "/=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(SLASH, "/", pos)
	l.readChar()
	return tok
}

// handlePercent handles the '%' operator and its variants (%=).
// Also handles binary literals starting with %.
func (l *Lexer) handlePercent(pos Position) Token {
	// Could be binary literal or modulo
	if isDigit(l.peekChar()) && (l.peekChar() == '0' || l.peekChar() == '1') {
		// Binary literal
		tokenType, literal := l.readNumber()
		return NewToken(tokenType, literal, pos)
	} else if l.matchAndConsume('=') {
		tok := NewToken(PERCENT_ASSIGN, "%=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(PERCENT, "%", pos)
	l.readChar()
	return tok
}

// Operator handler functions (Task 12.4.2 - Comparison and logical operators)

// handleEquals handles the '=' operator and its variants (==, ===, =>).
func (l *Lexer) handleEquals(pos Position) Token {
	if l.matchAndConsume('=') {
		if l.matchAndConsume('=') {
			tok := NewToken(EQ_EQ_EQ, "===", pos)
			l.readChar()
			return tok
		}
		tok := NewToken(EQ_EQ, "==", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('>') {
		tok := NewToken(FAT_ARROW, "=>", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(EQ, "=", pos)
	l.readChar()
	return tok
}

// handleLess handles the '<' operator and its variants (<>, <=, <<).
func (l *Lexer) handleLess(pos Position) Token {
	if l.matchAndConsume('>') {
		tok := NewToken(NOT_EQ, "<>", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('=') {
		tok := NewToken(LESS_EQ, "<=", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('<') {
		tok := NewToken(LESS_LESS, "<<", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(LESS, "<", pos)
	l.readChar()
	return tok
}

// handleGreater handles the '>' operator and its variants (>=, >>).
func (l *Lexer) handleGreater(pos Position) Token {
	if l.matchAndConsume('=') {
		tok := NewToken(GREATER_EQ, ">=", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('>') {
		tok := NewToken(GREATER_GREATER, ">>", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(GREATER, ">", pos)
	l.readChar()
	return tok
}

// handleExclamation handles the '!' operator and its variant (!=).
func (l *Lexer) handleExclamation(pos Position) Token {
	if l.matchAndConsume('=') {
		tok := NewToken(EXCL_EQ, "!=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(EXCLAMATION, "!", pos)
	l.readChar()
	return tok
}

// handleQuestion handles the '?' operator and its variants (??, ?.).
func (l *Lexer) handleQuestion(pos Position) Token {
	if l.matchAndConsume('?') {
		tok := NewToken(QUESTION_QUESTION, "??", pos)
		l.readChar()
		return tok
	} else if l.matchAndConsume('.') {
		tok := NewToken(QUESTION_DOT, "?.", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(QUESTION, "?", pos)
	l.readChar()
	return tok
}

// handleAmpersand handles the '&' operator and its variant (&&).
func (l *Lexer) handleAmpersand(pos Position) Token {
	if l.matchAndConsume('&') {
		tok := NewToken(AMP_AMP, "&&", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(AMP, "&", pos)
	l.readChar()
	return tok
}

// handlePipe handles the '|' operator and its variant (||).
func (l *Lexer) handlePipe(pos Position) Token {
	if l.matchAndConsume('|') {
		tok := NewToken(PIPE_PIPE, "||", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(PIPE, "|", pos)
	l.readChar()
	return tok
}

// handleCaret handles the '^' operator and its variant (^=).
func (l *Lexer) handleCaret(pos Position) Token {
	if l.matchAndConsume('=') {
		tok := NewToken(CARET_ASSIGN, "^=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(CARET, "^", pos)
	l.readChar()
	return tok
}

// handleAt handles the '@' operator and its variant (@=).
func (l *Lexer) handleAt(pos Position) Token {
	if l.matchAndConsume('=') {
		tok := NewToken(AT_ASSIGN, "@=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(AT, "@", pos)
	l.readChar()
	return tok
}

// handleTilde handles the '~' operator and its variant (~=).
func (l *Lexer) handleTilde(pos Position) Token {
	if l.matchAndConsume('=') {
		tok := NewToken(TILDE_ASSIGN, "~=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(TILDE, "~", pos)
	l.readChar()
	return tok
}

// skipWhitespace skips over whitespace characters (space, tab, newline, carriage return).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

// skipLineComment skips a line comment starting with //.
func (l *Lexer) skipLineComment() {
	// Skip until end of line or EOF
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips a block comment enclosed in { } or (* *).
// style: '{' for {} comments, '(' for (* *) comments
// Adds an error if the comment is not properly terminated.
func (l *Lexer) skipBlockComment(style rune) {
	startPos := l.currentPos()

	if style == '{' {
		l.readChar() // skip {
		for l.ch != 0 {
			if l.ch == '}' {
				l.readChar() // skip }
				return
			}
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
		// Unterminated comment
		l.addError("unterminated block comment", startPos)
		return
	}

	// style == '(' for (* *)
	l.readChar() // skip (
	l.readChar() // skip *

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == ')' {
			l.readChar() // skip *
			l.readChar() // skip )
			return
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	// Unterminated comment
	l.addError("unterminated block comment", startPos)
}

// skipCStyleComment skips a C-style multi-line comment /* */.
// Adds an error if the comment is not properly terminated.
func (l *Lexer) skipCStyleComment() {
	startPos := l.currentPos()

	l.readChar() // skip /
	l.readChar() // skip *

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip *
			l.readChar() // skip /
			return
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	// Unterminated comment
	l.addError("unterminated C-style comment", startPos)
}

// readIdentifier reads an identifier or keyword from the input.
// Identifiers start with a letter or underscore and continue with letters, digits, or underscores.
func (l *Lexer) readIdentifier() string {
	position := l.position

	// First character: letter or underscore
	if isLetter(l.ch) {
		l.readChar()
	}

	// Subsequent characters: letter, digit, or underscore
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

// readNumber reads a numeric literal (integer or float).
// Supports: decimal (123), hexadecimal ($FF or 0x), binary (%1010), and floats (123.45, 1.5e10)
func (l *Lexer) readNumber() (TokenType, string) {
	startPos := l.position

	// Check for hex with $ prefix
	if l.ch == '$' {
		l.readChar()
		for isHexDigit(l.ch) {
			l.readChar()
		}
		return INT, l.input[startPos:l.position]
	}

	// Check for binary with % prefix
	if l.ch == '%' {
		l.readChar()
		for l.ch == '0' || l.ch == '1' {
			l.readChar()
		}
		return INT, l.input[startPos:l.position]
	}

	// Check for hex with 0x prefix
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar() // skip 0
		l.readChar() // skip x
		for isHexDigit(l.ch) {
			l.readChar()
		}
		return INT, l.input[startPos:l.position]
	}

	// Read decimal digits
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for float (decimal point or exponent)
	isFloat := false

	// Decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // skip .
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Exponent (e or E)
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar() // skip e/E

		// Optional sign
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}

		// Exponent digits
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	tokenType := INT
	if isFloat {
		tokenType = FLOAT
	}

	return tokenType, l.input[startPos:l.position]
}

// readString reads a string literal enclosed in single or double quotes.
// DWScript uses single quotes by default, with " as escape for a single quote.
// If the string is unterminated, adds an error and returns the partial string.
func (l *Lexer) readString(quote rune) string {
	startPos := l.position
	startLine := l.line
	startColumn := l.column
	l.readChar() // skip opening quote

	var builder strings.Builder

	for l.ch != 0 {
		if l.ch == quote {
			// Check for escaped quote (doubled quote)
			if l.peekChar() == quote {
				builder.WriteRune(quote)
				l.readChar() // skip first quote
				l.readChar() // skip second quote
				continue
			}
			// End of string
			l.readChar() // skip closing quote
			return builder.String()
		}

		if l.ch == '\n' {
			l.line++
			l.column = 0
		}

		builder.WriteRune(l.ch)
		l.readChar()
	}

	// Unterminated string - add error and return partial string
	l.addError("unterminated string literal", Position{
		Line:   startLine,
		Column: startColumn,
		Offset: startPos,
	})
	return builder.String()
}

// readCharLiteral reads a character literal starting with #.
// Formats: #65 (decimal), #$41 (hex)
func (l *Lexer) readCharLiteral() string {
	startPos := l.position
	l.readChar() // skip #

	// Hex char literal: #$41
	if l.ch == '$' {
		l.readChar() // skip $
		for isHexDigit(l.ch) {
			l.readChar()
		}
		return l.input[startPos:l.position]
	}

	// Decimal char literal: #65
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[startPos:l.position]
}

// isCharLiteralStandalone checks if the '#' at the current position starts a standalone
// character literal (not part of a string concatenation sequence).
// Returns true if standalone, false if followed immediately by another string/char literal.
func (l *Lexer) isCharLiteralStandalone() bool {
	// Save current state for lookahead
	state := l.saveState()

	// Read the character literal to see what follows
	_ = l.readCharLiteral()

	// Check if immediately followed by another string/char literal (no whitespace)
	isStandalone := l.ch != '#' && l.ch != '\'' && l.ch != '"'

	// Restore state after lookahead
	l.restoreState(state)

	return isStandalone
}

// charLiteralToRune converts a character literal string (like "#13" or "#$0D") to a rune.
// Returns the rune and true on success, or 0 and false if the literal is invalid.
func charLiteralToRune(literal string) (rune, bool) {
	if len(literal) < 2 || literal[0] != '#' {
		return 0, false
	}

	var numStr string
	var base int

	// Check for hex prefix (#$XX) or decimal (#XX)
	if len(literal) > 2 && literal[1] == '$' {
		base = 16
		numStr = literal[2:] // Skip '#$'
	} else {
		base = 10
		numStr = literal[1:] // Skip '#'
	}

	// Parse the numeric value using strconv
	value, err := strconv.ParseInt(numStr, base, 32)
	if err != nil {
		return 0, false
	}

	return rune(value), true
}

// readStringOrCharSequence reads a sequence of adjacent string and character literals
// and concatenates them into a single string value.
// This handles DWScript's implicit concatenation: 'hello'#13#10'world' → "hello\r\nworld"
// Note: Only truly adjacent literals (no whitespace) are concatenated.
// Errors are accumulated in the lexer's error list rather than returned.
func (l *Lexer) readStringOrCharSequence() string {
	var builder strings.Builder

	for {
		// Save position for potential error reporting
		pos := l.currentPos()

		switch l.ch {
		case '\'', '"':
			// Read string literal
			quote := l.ch
			literal := l.readString(quote)
			builder.WriteString(literal)

		case '#':
			// Read character literal
			literal := l.readCharLiteral()
			r, ok := charLiteralToRune(literal)
			if !ok {
				l.addError("invalid character literal: "+literal, pos)
				// Continue processing to consume the invalid literal
			} else {
				builder.WriteRune(r)
			}

		default:
			// No more string/char literals to concatenate
			return builder.String()
		}

		// Check if next token is immediately adjacent (no whitespace allowed)
		// Only concatenate truly adjacent string/char literals
		if l.ch != '\'' && l.ch != '"' && l.ch != '#' {
			// No more adjacent literals
			return builder.String()
		}
	}
}

// NextToken returns the next token from the input.
// If tokens have been buffered by Peek(), it returns from the buffer first.
func (l *Lexer) NextToken() Token {
	// Check if we have buffered tokens from Peek()
	if len(l.tokenBuffer) > 0 {
		// Return the first buffered token and remove it
		tok := l.tokenBuffer[0]
		l.tokenBuffer = l.tokenBuffer[1:]
		return tok
	}

	// No buffered tokens, generate the next one
	return l.nextTokenInternal()
}

// nextTokenInternal generates the next token from the input.
// This is the internal tokenization logic, called by both NextToken() and Peek().
func (l *Lexer) nextTokenInternal() Token {
	l.skipWhitespace()

	var tok Token
	pos := l.currentPos()

	switch l.ch {
	case 0:
		tok = NewToken(EOF, "", pos)

	// Comments
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.nextTokenInternal() // Skip comment and get next token
		}
		if l.peekChar() == '*' {
			// C-style multi-line comment /* */
			l.skipCStyleComment()
			return l.nextTokenInternal()
		}
		return l.handleSlash(pos)

	case '{':
		// Block comment or compiler directive - both skip to }
		l.skipBlockComment('{')
		return l.nextTokenInternal()

	case '(':
		if l.peekChar() == '*' {
			// Block comment (* *)
			l.skipBlockComment('(')
			return l.nextTokenInternal()
		}
		tok = NewToken(LPAREN, "(", pos)
		l.readChar()

	case ')':
		tok = NewToken(RPAREN, ")", pos)
		l.readChar()

	case '[':
		tok = NewToken(LBRACK, "[", pos)
		l.readChar()

	case ']':
		tok = NewToken(RBRACK, "]", pos)
		l.readChar()

	case '}':
		tok = NewToken(RBRACE, "}", pos)
		l.readChar()

	case ';':
		tok = NewToken(SEMICOLON, ";", pos)
		l.readChar()

	case ',':
		tok = NewToken(COMMA, ",", pos)
		l.readChar()

	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			tok = NewToken(DOTDOT, "..", pos)
			l.readChar()
		} else {
			tok = NewToken(DOT, ".", pos)
			l.readChar()
		}

	case ':':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(ASSIGN, ":=", pos)
			l.readChar()
		} else {
			tok = NewToken(COLON, ":", pos)
			l.readChar()
		}

	case '+':
		return l.handlePlus(pos)

	case '-':
		return l.handleMinus(pos)

	case '*':
		return l.handleAsterisk(pos)

	case '%':
		return l.handlePercent(pos)

	case '=':
		return l.handleEquals(pos)

	case '<':
		return l.handleLess(pos)

	case '>':
		return l.handleGreater(pos)

	case '!':
		return l.handleExclamation(pos)

	case '?':
		return l.handleQuestion(pos)

	case '&':
		return l.handleAmpersand(pos)

	case '|':
		return l.handlePipe(pos)

	case '^':
		return l.handleCaret(pos)

	case '@':
		return l.handleAt(pos)

	case '~':
		return l.handleTilde(pos)

	case '\\':
		tok = NewToken(BACKSLASH, "\\", pos)
		l.readChar()

	case '$':
		// Check if followed by hex digit for hex literal
		if isHexDigit(l.peekChar()) {
			tokenType, literal := l.readNumber()
			tok = NewToken(tokenType, literal, pos)
		} else {
			tok = NewToken(DOLLAR, "$", pos)
			l.readChar()
		}

	case '#':
		// Character literal - check if standalone or part of concatenation
		if l.isCharLiteralStandalone() {
			// Standalone character literal: emit CHAR token
			literal := l.readCharLiteral()
			tok = NewToken(CHAR, literal, pos)
		} else {
			// Part of string concatenation: 'hello'#13#10 → "hello\r\n"
			literal := l.readStringOrCharSequence()
			tok = NewToken(STRING, literal, pos)
		}

	case '\'', '"':
		// String or character literal (with automatic concatenation)
		// DWScript concatenates adjacent string/char literals: 'hello'#13#10 → "hello\r\n"
		literal := l.readStringOrCharSequence()
		tok = NewToken(STRING, literal, pos)

	default:
		if isLetter(l.ch) {
			// Identifier or keyword
			literal := l.readIdentifier()
			tokenType := LookupIdent(literal)
			tok = NewToken(tokenType, literal, pos)
			return tok
		} else if isDigit(l.ch) {
			// Number literal
			tokenType, literal := l.readNumber()
			tok = NewToken(tokenType, literal, pos)
			return tok
		} else {
			// Illegal character - add error and emit ILLEGAL token
			l.addError("illegal character: "+string(l.ch), pos)
			tok = NewToken(ILLEGAL, string(l.ch), pos)
			l.readChar()
		}
	}

	return tok
}

// Helper functions

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return ('0' <= ch && ch <= '9') ||
		('a' <= ch && ch <= 'f') ||
		('A' <= ch && ch <= 'F')
}

// LexerError represents an error encountered during lexical analysis.
type LexerError struct {
	Message string
	Pos     Position
}

func (e *LexerError) Error() string {
	return e.Message
}
