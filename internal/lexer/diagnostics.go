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
	diag := LexerError{Message: msg, Pos: pos, Severity: severity, Rendered: rendered}
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
