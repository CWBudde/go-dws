package lexer

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer represents a lexical scanner for DWScript source code.
//
// # Unicode and Column Positions
//
// The lexer handles UTF-8 encoded source code correctly. Column positions are
// reported as rune counts, not byte offsets or display widths.
//
// Important characteristics:
//   - "column" represents the count of Unicode code points (runes) from the start of the line
//   - Multi-byte UTF-8 sequences (like emoji 🚀, Greek Δ, or Chinese 中) each count as 1 column
//   - This differs from display width: emoji may render as 2 cells, but count as 1 column
//   - Combining characters are counted as separate runes
//
// This design choice prioritizes:
//   - Simplicity: rune counting is straightforward and well-defined
//   - Performance: no need to calculate complex display widths
//   - Consistency: same behavior across all Unicode characters
//
// Trade-off: Error messages may not align perfectly with visual column positions
// in terminals when emoji or wide characters are present, but positions are
// consistent and reproducible.
//
// Examples:
//   - "var x" → 'x' is at column 5 (5 runes: v, a, r, space, x)
//   - "var Δ" → 'Δ' is at column 5 (5 runes, Δ is a single multi-byte rune)
//   - "// 🚀" → '🚀' is at column 4 (4 runes: /, /, space, 🚀)
type Lexer struct {
	input            string
	errors           []LexerError
	tokenBuffer      []Token
	position         int
	readPosition     int
	line             int
	column           int
	ch               rune
	preserveComments bool
	tracing          bool
}

// LexerState represents the complete state of the Lexer at a specific point in time.
// It can be saved and restored to enable backtracking during parsing.
// This allows for efficient save/restore operations during lookahead.
type LexerState struct {
	tokenBuffer  []Token
	position     int
	readPosition int
	line         int
	column       int
	ch           rune
}

// LexerOption is a function that configures a Lexer.
// Options are applied during lexer creation via New().
type LexerOption func(*Lexer)

// WithPreserveComments enables or disables comment preservation.
// When enabled, the lexer will return COMMENT tokens instead of skipping comments.
// This is useful for formatters and documentation tools that need to preserve comments.
func WithPreserveComments(preserve bool) LexerOption {
	return func(l *Lexer) {
		l.preserveComments = preserve
	}
}

// WithTracing enables or disables debug tracing output.
// When enabled, the lexer may output debug information about its operation.
// This is useful for debugging lexer behavior during development.
func WithTracing(trace bool) LexerOption {
	return func(l *Lexer) {
		l.tracing = trace
	}
}

// New creates a new Lexer for the given input string with optional configuration.
// Automatically detects and strips UTF-8 BOM (0xEF 0xBB 0xBF) if present at the start.
// This matches the behavior of the original DWScript which strips BOMs during file reading.
//
// Options can be provided to configure the lexer:
//   - WithPreserveComments(true): Return COMMENT tokens instead of skipping them
//   - WithTracing(true): Enable debug tracing output
//
// Example:
//
//	l := New(input, WithPreserveComments(true), WithTracing(false))
func New(input string, opts ...LexerOption) *Lexer {
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

	// Apply options
	for _, opt := range opts {
		opt(l)
	}

	l.readChar() // Initialize first character
	return l
}

// readChar advances the lexer to the next character in the input.
// Properly handles UTF-8 multi-byte sequences and detects invalid UTF-8.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
		l.position = l.readPosition
		l.column++
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		// Update state first so currentPos() returns the correct position
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
		l.column++
		// Check for invalid UTF-8 encoding after updating position
		if r == utf8.RuneError && size == 1 {
			l.addError("invalid UTF-8 encoding", l.currentPos())
		}
	}
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
// Optimized to decode the UTF-8 rune only once instead of twice.
func (l *Lexer) matchAndConsume(expected rune) bool {
	var r rune
	var size int

	if l.readPosition >= len(l.input) {
		r = 0 // EOF
		size = 0
	} else {
		r, size = utf8.DecodeRuneInString(l.input[l.readPosition:])
	}

	if r == expected {
		// Update state first so currentPos() returns the correct position
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
		l.column++
		// Check for invalid UTF-8 encoding after updating position (only for non-EOF)
		if r == utf8.RuneError && size == 1 {
			l.addError("invalid UTF-8 encoding", l.currentPos())
		}
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

// SaveState captures the current lexer state for later restoration.
// This is useful for lookahead operations and parser backtracking.
// Deep copies the tokenBuffer to prevent corruption during speculative parsing.
func (l *Lexer) SaveState() LexerState {
	// Deep copy the token buffer to preserve lookahead state
	bufferCopy := make([]Token, len(l.tokenBuffer))
	copy(bufferCopy, l.tokenBuffer)

	return LexerState{
		position:     l.position,
		readPosition: l.readPosition,
		ch:           l.ch,
		line:         l.line,
		column:       l.column,
		tokenBuffer:  bufferCopy,
	}
}

// RestoreState restores the lexer to a previously saved state.
// This is used after lookahead operations or parser backtracking to return to the original position.
// Restores the tokenBuffer to prevent token duplication or skipping.
func (l *Lexer) RestoreState(s LexerState) {
	l.position = s.position
	l.readPosition = s.readPosition
	l.ch = s.ch
	l.line = s.line
	l.column = s.column
	l.tokenBuffer = s.tokenBuffer
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

// SetPreserveComments enables or disables comment preservation.
// When enabled, the lexer will return COMMENT tokens instead of skipping comments.
// This is useful for formatters and documentation tools that need to preserve comments.
func (l *Lexer) SetPreserveComments(preserve bool) {
	l.preserveComments = preserve
}

// PreserveComments returns the current comment preservation setting.
func (l *Lexer) PreserveComments() bool {
	return l.preserveComments
}

// readLineComment reads a line comment starting with //.
// Returns the comment text including the // prefix.
func (l *Lexer) readLineComment() string {
	startPos := l.position
	// Read until end of line or EOF
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[startPos:l.position]
}

// readBlockComment reads a block comment enclosed in { } or (* *).
// style: '{' for {} comments, '(' for (* *) comments
// Returns the comment text including delimiters, and true if properly terminated.
func (l *Lexer) readBlockComment(style rune) (string, bool) {
	startPos := l.position

	if style == '{' {
		l.readChar() // skip {
		for l.ch != 0 {
			if l.ch == '}' {
				l.readChar() // skip }
				return l.input[startPos:l.position], true
			}
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
		// Unterminated comment
		return l.input[startPos:l.position], false
	}

	// style == '(' for (* *)
	l.readChar() // skip (
	l.readChar() // skip *

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == ')' {
			l.readChar() // skip *
			l.readChar() // skip )
			return l.input[startPos:l.position], true
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	// Unterminated comment
	return l.input[startPos:l.position], false
}

// readCStyleComment reads a C-style comment /* */.
// Returns the comment text including delimiters, and true if properly terminated.
func (l *Lexer) readCStyleComment() (string, bool) {
	startPos := l.position
	l.readChar() // skip /
	l.readChar() // skip *

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip *
			l.readChar() // skip /
			return l.input[startPos:l.position], true
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	// Unterminated comment
	return l.input[startPos:l.position], false
}

// Operator handler functions (Task 12.4.1 - Arithmetic operators)

// tokenHandler is a function that handles a specific character/operator.
// It takes the current position and returns the corresponding token.
type tokenHandler func(*Lexer, Position) Token

// tokenHandlers maps characters to their handler functions.
// This dispatch table eliminates the need for a large switch statement.
var tokenHandlers = map[rune]tokenHandler{
	'+': (*Lexer).handlePlus,
	'-': (*Lexer).handleMinus,
	'*': (*Lexer).handleAsterisk,
	'=': (*Lexer).handleEquals,
	'<': (*Lexer).handleLess,
	'>': (*Lexer).handleGreater,
	'!': (*Lexer).handleExclamation,
	'?': (*Lexer).handleQuestion,
	'&': (*Lexer).handleAmpersand,
	'|': (*Lexer).handlePipe,
	'^': (*Lexer).handleCaret,
	'@': (*Lexer).handleAt,
	'~': (*Lexer).handleTilde,
}

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
	var tok Token
	switch l.peekChar() {
	case '>':
		l.readChar()
		tok = NewToken(NOT_EQ, "<>", pos)
		l.readChar()
	case '=':
		l.readChar()
		tok = NewToken(LESS_EQ, "<=", pos)
		l.readChar()
	case '<':
		l.readChar()
		tok = NewToken(LESS_LESS, "<<", pos)
		l.readChar()
	default:
		tok = NewToken(LESS, "<", pos)
		l.readChar()
	}
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
// Implemented by calling readLineComment and discarding the result.
func (l *Lexer) skipLineComment() {
	_ = l.readLineComment()
}

// skipBlockComment skips a block comment enclosed in { } or (* *).
// style: '{' for {} comments, '(' for (* *) comments
// Adds an error if the comment is not properly terminated.
// Implemented by calling readBlockComment and discarding the result.
func (l *Lexer) skipBlockComment(style rune) {
	startPos := l.currentPos()
	_, terminated := l.readBlockComment(style)
	if !terminated {
		l.addError("unterminated block comment", startPos)
	}
}

// skipCStyleComment skips a C-style multi-line comment /* */.
// Adds an error if the comment is not properly terminated.
// Implemented by calling readCStyleComment and discarding the result.
func (l *Lexer) skipCStyleComment() {
	startPos := l.currentPos()
	_, terminated := l.readCStyleComment()
	if !terminated {
		l.addError("unterminated C-style comment", startPos)
	}
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
// Supports: decimal (123), hexadecimal ($FF or 0x), binary (%1010 or 0b), and floats (123.45, 1.5e10)
func (l *Lexer) readNumber() (TokenType, string) {
	startPos := l.position

	// Check for hex with $ prefix
	if l.ch == '$' {
		return l.readHexNumber(startPos)
	}

	// Check for binary with % prefix
	if l.ch == '%' {
		return l.readBinaryNumber(startPos)
	}

	// Check for hex with 0x prefix or binary with 0b prefix
	if l.ch == '0' {
		nextCh := l.peekChar()
		if nextCh == 'x' || nextCh == 'X' {
			return l.readHexNumber0x(startPos)
		}
		if nextCh == 'b' || nextCh == 'B' {
			return l.readBinaryNumber0b(startPos)
		}
	}

	// Read decimal number (may be float)
	return l.readDecimalNumber(startPos)
}

// readHexNumber reads a hexadecimal number starting with $ (e.g., $FF).
func (l *Lexer) readHexNumber(startPos int) (TokenType, string) {
	l.readChar() // skip $
	for isHexDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return INT, l.input[startPos:l.position]
}

// readBinaryNumber reads a binary number starting with % (e.g., %1010).
func (l *Lexer) readBinaryNumber(startPos int) (TokenType, string) {
	pos := l.currentPos() // Save position for error reporting
	l.readChar()          // skip %

	digitStart := l.position
	for l.ch == '0' || l.ch == '1' || l.ch == '_' {
		l.readChar()
	}

	// Validate that at least one binary digit was present
	if l.position == digitStart {
		l.addError("binary literal requires at least one digit after '%'", pos)
	}

	return INT, l.input[startPos:l.position]
}

// readBinaryNumber0b reads a binary number starting with 0b (e.g., 0b1010).
func (l *Lexer) readBinaryNumber0b(startPos int) (TokenType, string) {
	pos := l.currentPos() // Save position for error reporting
	l.readChar()          // skip 0
	l.readChar()          // skip b

	digitStart := l.position
	for l.ch == '0' || l.ch == '1' || l.ch == '_' {
		l.readChar()
	}

	// Validate that at least one binary digit was present
	if l.position == digitStart {
		l.addError("binary literal requires at least one digit after '0b'", pos)
	}

	return INT, l.input[startPos:l.position]
}

// readHexNumber0x reads a hexadecimal number starting with 0x (e.g., 0xFF).
func (l *Lexer) readHexNumber0x(startPos int) (TokenType, string) {
	pos := l.currentPos() // Save position for error reporting
	l.readChar()          // skip 0
	l.readChar()          // skip x

	digitStart := l.position
	for isHexDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Validate that at least one hex digit was present
	if l.position == digitStart {
		l.addError("hexadecimal literal requires at least one digit after '0x'", pos)
	}

	return INT, l.input[startPos:l.position]
}

// readDecimalNumber reads a decimal number, potentially with float components.
func (l *Lexer) readDecimalNumber(startPos int) (TokenType, string) {
	// Read decimal digits
	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Check for float (decimal point or exponent)
	isFloat := false

	// Decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // skip .
		for isDigit(l.ch) || l.ch == '_' {
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
		for isDigit(l.ch) || l.ch == '_' {
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
	state := l.SaveState()

	// Read the character literal to see what follows
	_ = l.readCharLiteral()

	// Check if immediately followed by another string/char literal (no whitespace)
	isStandalone := l.ch != '#' && l.ch != '\'' && l.ch != '"'

	// Restore state after lookahead
	l.RestoreState(state)

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

// handleSimpleToken creates a simple token and advances the lexer.
func (l *Lexer) handleSimpleToken(tokenType TokenType, literal string, pos Position) Token {
	tok := NewToken(tokenType, literal, pos)
	l.readChar()
	return tok
}

// handleString handles string or character literals with automatic concatenation.
func (l *Lexer) handleString(pos Position) Token {
	// DWScript concatenates adjacent string/char literals: 'hello'#13#10 → "hello\r\n"
	literal := l.readStringOrCharSequence()
	return NewToken(STRING, literal, pos)
}

// handleDefault handles characters not matched by specific cases.
// It checks the dispatch table, then handles identifiers, numbers, or illegal characters.
func (l *Lexer) handleDefault(pos Position) Token {
	// Check dispatch table for operator handlers
	if handler, ok := tokenHandlers[l.ch]; ok {
		return handler(l, pos)
	}

	switch {
	case isLetter(l.ch):
		// Identifier or keyword
		literal := l.readIdentifier()
		tokenType := LookupIdent(literal)
		return NewToken(tokenType, literal, pos)
	case isDigit(l.ch):
		// Number literal
		tokenType, literal := l.readNumber()
		return NewToken(tokenType, literal, pos)
	default:
		// Illegal character - add error and emit ILLEGAL token
		// Note: Don't report RuneError here, as invalid UTF-8 was already reported by readChar()
		if l.ch != utf8.RuneError {
			l.addError("illegal character: "+string(l.ch), pos)
		}
		tok := NewToken(ILLEGAL, string(l.ch), pos)
		l.readChar()
		return tok
	}
}

// handleDot handles the '.' character which could be a single dot or range (..).
func (l *Lexer) handleDot(pos Position) Token {
	if l.peekChar() == '.' {
		l.readChar()
		tok := NewToken(DOTDOT, "..", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(DOT, ".", pos)
	l.readChar()
	return tok
}

// handleColon handles the ':' character which could be a colon or assignment (:=).
func (l *Lexer) handleColon(pos Position) Token {
	if l.peekChar() == '=' {
		l.readChar()
		tok := NewToken(ASSIGN, ":=", pos)
		l.readChar()
		return tok
	}
	tok := NewToken(COLON, ":", pos)
	l.readChar()
	return tok
}

// handleDollar handles the '$' character which could be a hex literal or address-of operator.
func (l *Lexer) handleDollar(pos Position) Token {
	// Check if followed by hex digit for hex literal
	if isHexDigit(l.peekChar()) {
		tokenType, literal := l.readNumber()
		return NewToken(tokenType, literal, pos)
	}
	tok := NewToken(DOLLAR, "$", pos)
	l.readChar()
	return tok
}

// handleHash handles the '#' character for character literals or string concatenation.
func (l *Lexer) handleHash(pos Position) Token {
	// Character literal - check if standalone or part of concatenation
	if l.isCharLiteralStandalone() {
		// Standalone character literal: emit CHAR token
		literal := l.readCharLiteral()
		return NewToken(CHAR, literal, pos)
	}
	// Part of string concatenation: 'hello'#13#10 → "hello\r\n"
	literal := l.readStringOrCharSequence()
	return NewToken(STRING, literal, pos)
}

// handleSlashToken handles the '/' character which could be division, comment, or compound assignment.
func (l *Lexer) handleSlashToken(pos Position) Token {
	switch l.peekChar() {
	case '/':
		if l.preserveComments {
			text := l.readLineComment()
			return NewToken(COMMENT, text, pos)
		}
		l.skipLineComment()
		return l.nextTokenInternal() // Skip comment and get next token
	case '*':
		// C-style multi-line comment /* */
		if l.preserveComments {
			text, ok := l.readCStyleComment()
			if !ok {
				return NewToken(ILLEGAL, "unterminated C-style comment", pos)
			}
			return NewToken(COMMENT, text, pos)
		}
		l.skipCStyleComment()
		return l.nextTokenInternal()
	default:
		return l.handleSlash(pos)
	}
}

// handleCurlyBrace handles '{' which could be a block comment or would be handled by the dispatch table.
func (l *Lexer) handleCurlyBrace(pos Position) Token {
	// Block comment or compiler directive - both skip to }
	if l.preserveComments {
		text, ok := l.readBlockComment('{')
		if !ok {
			return NewToken(ILLEGAL, "unterminated block comment", pos)
		}
		return NewToken(COMMENT, text, pos)
	}
	l.skipBlockComment('{')
	return l.nextTokenInternal()
}

// handleLeftParen handles '(' which could be a block comment (* *) or just a left parenthesis.
func (l *Lexer) handleLeftParen(pos Position) Token {
	if l.peekChar() == '*' {
		// Block comment (* *)
		if l.preserveComments {
			text, ok := l.readBlockComment('(')
			if !ok {
				return NewToken(ILLEGAL, "unterminated block comment", pos)
			}
			return NewToken(COMMENT, text, pos)
		}
		l.skipBlockComment('(')
		return l.nextTokenInternal()
	}
	tok := NewToken(LPAREN, "(", pos)
	l.readChar()
	return tok
}

// nextTokenInternal generates the next token from the input.
// This is the internal tokenization logic, called by both NextToken() and Peek().
func (l *Lexer) nextTokenInternal() Token {
	l.skipWhitespace()
	pos := l.currentPos()

	switch l.ch {
	case 0:
		return NewToken(EOF, "", pos)
	case '/':
		return l.handleSlashToken(pos)
	case '{':
		return l.handleCurlyBrace(pos)

	case '(':
		return l.handleLeftParen(pos)
	case ')':
		return l.handleSimpleToken(RPAREN, ")", pos)
	case '[':
		return l.handleSimpleToken(LBRACK, "[", pos)
	case ']':
		return l.handleSimpleToken(RBRACK, "]", pos)
	case '}':
		return l.handleSimpleToken(RBRACE, "}", pos)
	case ';':
		return l.handleSimpleToken(SEMICOLON, ";", pos)
	case ',':
		return l.handleSimpleToken(COMMA, ",", pos)
	case '.':
		return l.handleDot(pos)
	case ':':
		return l.handleColon(pos)
	case '%':
		return l.handlePercent(pos)
	case '\\':
		return l.handleSimpleToken(BACKSLASH, "\\", pos)
	case '$':
		return l.handleDollar(pos)

	case '#':
		return l.handleHash(pos)

	case '\'', '"':
		return l.handleString(pos)

	default:
		return l.handleDefault(pos)
	}
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
