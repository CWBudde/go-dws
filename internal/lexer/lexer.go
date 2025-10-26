package lexer

import (
	"strings"
	"unicode"
)

// Lexer represents a lexical scanner for DWScript source code.
type Lexer struct {
	input        string // The source code being tokenized
	position     int    // Current position in input (points to current char)
	readPosition int    // Current reading position in input (after current char)
	ch           byte   // Current character under examination (0 if EOF)
	line         int    // Current line number (1-indexed)
	column       int    // Current column number (1-indexed)
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar() // Initialize first character
	return l
}

// readChar advances the lexer to the next character in the input.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL for EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar returns the next character without advancing the position.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekCharN returns the character n positions ahead without advancing.
func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// currentPos returns the current Position for token creation.
func (l *Lexer) currentPos() Position {
	return Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.position,
	}
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
// Returns true if comment was properly terminated, false otherwise
func (l *Lexer) skipBlockComment(style byte) bool {
	if style == '{' {
		l.readChar() // skip {
		for l.ch != 0 {
			if l.ch == '}' {
				l.readChar() // skip }
				return true
			}
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
		// Unterminated comment
		return false
	}

	// style == '(' for (* *)
	l.readChar() // skip (
	l.readChar() // skip *

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == ')' {
			l.readChar() // skip *
			l.readChar() // skip )
			return true
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	// Unterminated comment
	return false
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
// DWScript uses single quotes by default, with ” as escape for a single quote.
func (l *Lexer) readString(quote byte) (string, error) {
	startPos := l.position
	l.readChar() // skip opening quote

	var builder strings.Builder

	for l.ch != 0 {
		if l.ch == quote {
			// Check for escaped quote (doubled quote)
			if l.peekChar() == quote {
				builder.WriteByte(quote)
				l.readChar() // skip first quote
				l.readChar() // skip second quote
				continue
			}
			// End of string
			l.readChar() // skip closing quote
			return builder.String(), nil
		}

		if l.ch == '\n' {
			l.line++
			l.column = 0
		}

		builder.WriteByte(l.ch)
		l.readChar()
	}

	// Unterminated string
	return l.input[startPos:l.position], &LexerError{
		Message: "unterminated string literal",
		Pos:     Position{Line: l.line, Column: l.column, Offset: startPos},
	}
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

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
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
			return l.NextToken() // Skip comment and get next token
		}
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(DIVIDE_ASSIGN, "/=", pos)
			l.readChar()
		} else {
			tok = NewToken(SLASH, "/", pos)
			l.readChar()
		}

	case '{':
		// Block comment or compiler directive - both skip to }
		if !l.skipBlockComment('{') {
			tok = NewToken(ILLEGAL, "unterminated block comment", pos)
			return tok
		}
		return l.NextToken()

	case '(':
		if l.peekChar() == '*' {
			// Block comment (* *)
			if !l.skipBlockComment('(') {
				tok = NewToken(ILLEGAL, "unterminated block comment", pos)
				return tok
			}
			return l.NextToken()
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
		if l.peekChar() == '+' {
			l.readChar()
			tok = NewToken(INC, "++", pos)
			l.readChar()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(PLUS_ASSIGN, "+=", pos)
			l.readChar()
		} else {
			tok = NewToken(PLUS, "+", pos)
			l.readChar()
		}

	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = NewToken(DEC, "--", pos)
			l.readChar()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(MINUS_ASSIGN, "-=", pos)
			l.readChar()
		} else {
			tok = NewToken(MINUS, "-", pos)
			l.readChar()
		}

	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			tok = NewToken(POWER, "**", pos)
			l.readChar()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TIMES_ASSIGN, "*=", pos)
			l.readChar()
		} else {
			tok = NewToken(ASTERISK, "*", pos)
			l.readChar()
		}

	case '%':
		// Could be binary literal or modulo
		if isDigit(l.peekChar()) && (l.peekChar() == '0' || l.peekChar() == '1') {
			// Binary literal
			tokenType, literal := l.readNumber()
			tok = NewToken(tokenType, literal, pos)
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(PERCENT_ASSIGN, "%=", pos)
			l.readChar()
		} else {
			tok = NewToken(PERCENT, "%", pos)
			l.readChar()
		}

	case '=':
		if l.peekChar() == '=' {
			if l.peekCharN(2) == '=' {
				l.readChar()
				l.readChar()
				tok = NewToken(EQ_EQ_EQ, "===", pos)
				l.readChar()
			} else {
				l.readChar()
				tok = NewToken(EQ_EQ, "==", pos)
				l.readChar()
			}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = NewToken(FAT_ARROW, "=>", pos)
			l.readChar()
		} else {
			tok = NewToken(EQ, "=", pos)
			l.readChar()
		}

	case '<':
		if l.peekChar() == '>' {
			l.readChar()
			tok = NewToken(NOT_EQ, "<>", pos)
			l.readChar()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(LESS_EQ, "<=", pos)
			l.readChar()
		} else if l.peekChar() == '<' {
			l.readChar()
			tok = NewToken(LESS_LESS, "<<", pos)
			l.readChar()
		} else {
			tok = NewToken(LESS, "<", pos)
			l.readChar()
		}

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(GREATER_EQ, ">=", pos)
			l.readChar()
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = NewToken(GREATER_GREATER, ">>", pos)
			l.readChar()
		} else {
			tok = NewToken(GREATER, ">", pos)
			l.readChar()
		}

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(EXCL_EQ, "!=", pos)
			l.readChar()
		} else {
			tok = NewToken(EXCLAMATION, "!", pos)
			l.readChar()
		}

	case '?':
		if l.peekChar() == '?' {
			l.readChar()
			tok = NewToken(QUESTION_QUESTION, "??", pos)
			l.readChar()
		} else if l.peekChar() == '.' {
			l.readChar()
			tok = NewToken(QUESTION_DOT, "?.", pos)
			l.readChar()
		} else {
			tok = NewToken(QUESTION, "?", pos)
			l.readChar()
		}

	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = NewToken(AMP_AMP, "&&", pos)
			l.readChar()
		} else {
			tok = NewToken(AMP, "&", pos)
			l.readChar()
		}

	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = NewToken(PIPE_PIPE, "||", pos)
			l.readChar()
		} else {
			tok = NewToken(PIPE, "|", pos)
			l.readChar()
		}

	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(CARET_ASSIGN, "^=", pos)
			l.readChar()
		} else {
			tok = NewToken(CARET, "^", pos)
			l.readChar()
		}

	case '@':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(AT_ASSIGN, "@=", pos)
			l.readChar()
		} else {
			tok = NewToken(AT, "@", pos)
			l.readChar()
		}

	case '~':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TILDE_ASSIGN, "~=", pos)
			l.readChar()
		} else {
			tok = NewToken(TILDE, "~", pos)
			l.readChar()
		}

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
		// Character literal
		literal := l.readCharLiteral()
		tok = NewToken(CHAR, literal, pos)

	case '\'', '"':
		// String literal
		quote := l.ch
		literal, err := l.readString(quote)
		if err != nil {
			tok = NewToken(ILLEGAL, literal, pos)
		} else {
			tok = NewToken(STRING, literal, pos)
		}

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
			// Illegal character
			tok = NewToken(ILLEGAL, string(l.ch), pos)
			l.readChar()
		}
	}

	return tok
}

// Helper functions

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
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
