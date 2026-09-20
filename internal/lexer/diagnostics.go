package lexer

// Diagnostics produced during lexical analysis: the severity classification and the
// lexer-side helpers that accumulate diagnostics. The LexerError record itself stays in
// lexer.go alongside the Lexer.

// Severity classifies a lexer diagnostic. DWScript compiler directives such as
// {$HINT} and {$WARNING} emit non-error messages, so the lexer needs more than a single
// error severity.
type Severity int

const (
	// SeverityError is the default severity and blocks compilation.
	SeverityError Severity = iota
	// SeverityWarning is a non-blocking warning ({$WARNING}).
	SeverityWarning
	// SeverityHint is a non-blocking hint ({$HINT}).
	SeverityHint
)

// DirectiveDiagnostics returns the subset of lexer diagnostics produced by compiler
// directives ({$HINT}, {$WARNING}, {$ERROR}, {$FATAL}, and malformed conditional
// directives). Unlike the general advisory error list these are surfaced by the front
// end, so a script using them reports DWScript-compatible messages.
func (l *Lexer) DirectiveDiagnostics() []LexerError {
	return l.directiveDiags
}

// addDirectiveDiagnostic records a compiler-directive diagnostic. rendered may be empty,
// in which case the front end applies its default formatting.
//
// Parser backtracking can re-lex the same directive, so identical diagnostics at the same
// position are recorded only once.
func (l *Lexer) addDirectiveDiagnostic(msg string, pos Position, severity Severity, rendered string) {
	for i := range l.directiveDiags {
		existing := &l.directiveDiags[i]
		if existing.Message == msg && existing.Pos.Line == pos.Line && existing.Pos.Column == pos.Column {
			return
		}
	}
	diag := LexerError{Message: msg, Pos: pos, Severity: severity, Rendered: rendered, Directive: true}
	l.directiveDiags = append(l.directiveDiags, diag)
	l.errors = append(l.errors, diag)
}

// addIncludeError records an include-resolution failure. It is tracked both in the
// general error list and in the dedicated include-error list.
func (l *Lexer) addIncludeError(msg string, pos Position) {
	err := LexerError{Message: msg, Pos: pos}
	l.errors = append(l.errors, err)
	l.includeErrors = append(l.includeErrors, err)
}

// addError adds a new error to the lexer's error list.
// This follows the parser's pattern of accumulating errors instead of stopping at the first error.
func (l *Lexer) addError(msg string, pos Position) {
	l.errors = append(l.errors, LexerError{
		Message: msg,
		Pos:     pos,
	})
}

// DWScript tokenizer messages for malformed string and char constants.
const (
	msgStringEndOfLine          = "End of string constant not found (end of line)"
	msgStringEndOfFile          = "End of string constant not found (end of file)"
	msgTripleApostrophe         = "Incorrect triple apostrophe string"
	msgTripleApostropheIndent   = "Incorrect triple apostrophe string indentation"
	msgInvalidCharConstantQuote = "Invalid char constant %q"
)

// addConstantError records a malformed string or char constant error. It is surfaced
// by the front end through the same channel as directive diagnostics, marked as a
// constant error (see LexerError.Constant). Identical errors at the same position,
// from re-lexing after parser backtracking, are recorded once.
func (l *Lexer) addConstantError(msg string, pos Position) {
	for i := range l.directiveDiags {
		existing := &l.directiveDiags[i]
		if existing.Message == msg && existing.Pos.Line == pos.Line && existing.Pos.Column == pos.Column {
			return
		}
	}
	diag := LexerError{Message: msg, Pos: pos, Severity: SeverityError, Constant: true}
	l.directiveDiags = append(l.directiveDiags, diag)
	l.errors = append(l.errors, diag)
}

// addFatalConstantError records a constant error that DWScript treats as a compiler
// stop: tokenization ends, so nothing after the malformed constant is compiled and the
// parser's truncation artifacts are suppressed as for {$FATAL}.
func (l *Lexer) addFatalConstantError(msg string, pos Position) {
	l.addConstantError(msg, pos)
	l.stopped = true
}

// dwsEndOfFilePos returns the position DWScript reports for an error raised at the end
// of the source. DWScript's tokenizer appends a line terminator to the source text, and
// its script hosts hand the compiler text whose last line is already terminated, so
// the end of file sits at column 1 of the line after the terminated last line.
// Must be called with the lexer positioned at the end of input.
func (l *Lexer) dwsEndOfFilePos() Position {
	pos := l.currentPos()
	if pos.Column > 1 {
		// The last line is unterminated: the host terminates it first.
		pos.Line++
	}
	pos.Line++
	pos.Column = 1
	return pos
}
