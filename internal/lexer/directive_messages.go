package lexer

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// directiveNameColumn returns the column DWScript anchors a directive diagnostic to.
// startPos is the position of the opening '{', and the message is reported at the
// directive name, i.e. just past "{$".
func directiveNameColumn(startPos Position) Position {
	pos := startPos
	pos.Column += 2
	return pos
}

// directiveArgument returns the raw text following the directive name in content.
//
// The caller must not use strings.Fields for this: a quoted message such as
// {$HINT 'first hint'} contains spaces and would be truncated at the first one.
func directiveArgument(content string) string {
	trimmed := strings.TrimSpace(content)
	idx := strings.IndexFunc(trimmed, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(trimmed[idx:])
}

// unquoteDirectiveString strips a matching pair of single or double quotes.
// DWScript accepts both, e.g. {$FATAL 'done'} and {$FATAL "done"}.
// It reports whether arg was a quoted string literal.
func unquoteDirectiveString(arg string) (string, bool) {
	if len(arg) < 2 {
		return "", false
	}
	quote := arg[0]
	if quote != '\'' && quote != '"' {
		return "", false
	}
	if arg[len(arg)-1] != quote {
		return "", false
	}
	return arg[1 : len(arg)-1], true
}

// handleMessageDirective implements the DWScript message directives {$HINT}, {$WARNING},
// {$ERROR} and {$FATAL}.
//
// The message text is emitted verbatim (unquoted) with the standard position suffix,
// anchored at the directive name. {$ERROR} and {$FATAL} share the "Compile Error" prefix
// and differ only in whether compilation continues: {$FATAL} stops tokenizing immediately,
// so nothing after it is compiled, while messages recorded before it are retained.
//
// A directive inside an inactive conditional branch emits nothing.
func (l *Lexer) handleMessageDirective(
	content string,
	parentActive bool,
	startPos, closePos Position,
	severity LexerSeverity,
	prefix string,
	stop bool,
) {
	if !parentActive {
		return
	}

	arg := directiveArgument(content)
	message, ok := unquoteDirectiveString(arg)
	if !ok {
		// A missing or unquoted argument is a syntax error anchored at the closing
		// brace. {$FATAL} still stops, even when its argument is invalid.
		l.addDirectiveDiagnostic("String expected",
			directiveArgPosition(content, startPos, closePos), LexerSeverityError, "")
		if stop {
			l.stopped = true
		}
		return
	}

	namePos := directiveNameColumn(startPos)
	rendered := fmt.Sprintf("%s: %s [line: %d, column: %d]", prefix, message, namePos.Line, namePos.Column)
	l.addDirectiveDiagnostic(message, namePos, severity, rendered)

	if stop {
		l.stopped = true
	}
}

// handleSwitchToggle implements the plural on/off switches {$HINTS} and {$WARNINGS}.
// They take ON or OFF; anything else reports "ON/OFF expected" at the closing brace.
//
// The switches are accepted and parsed for message parity but do not currently change
// which diagnostics are reported.
func (l *Lexer) handleSwitchToggle(content string, parentActive bool, startPos, closePos Position) {
	if !parentActive {
		return
	}

	arg := directiveArgument(content)
	switch ident.Normalize(arg) {
	case "on", "off", "normal", "strict", "pedantic":
		return
	default:
		l.addDirectiveDiagnostic("ON/OFF expected",
			directiveArgPosition(content, startPos, closePos), LexerSeverityError, "")
	}
}

// handleStringSwitch implements switches whose argument must be a quoted string,
// such as {$R 'file'}.
func (l *Lexer) handleStringSwitch(content string, parentActive bool, startPos, closePos Position) {
	if !parentActive {
		return
	}
	if _, ok := unquoteDirectiveString(directiveArgument(content)); !ok {
		l.addDirectiveDiagnostic("String expected",
			directiveArgPosition(content, startPos, closePos), LexerSeverityError, "")
	}
}

// knownSwitches lists the compiler switches the lexer recognizes. An unterminated
// directive names its switch before the missing '}' is reported, and an unrecognized
// name is reported as an unknown switch instead.
var knownSwitches = map[string]struct{}{
	"define": {}, "undef": {}, "ifdef": {}, "ifndef": {}, "else": {}, "endif": {},
	"ifend": {}, "if": {}, "include": {}, "i": {}, "include_once": {}, "filter": {},
	"f": {}, "hint": {}, "hints": {}, "warning": {}, "warnings": {}, "error": {},
	"fatal": {}, "region": {}, "endregion": {}, "r": {}, "resource": {},
}

// directiveArgPosition returns the position a directive's argument diagnostic anchors to:
// the start of the argument when there is one, and the closing brace otherwise.
//
//	{$HINTS BUGGED}  -> the argument   ("ON/OFF expected" at BUGGED)
//	{$HINTS}         -> the brace      ("ON/OFF expected" at })
func directiveArgPosition(content string, startPos, closePos Position) Position {
	trimmed := strings.TrimSpace(content)
	idx := strings.IndexFunc(trimmed, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
	if idx < 0 {
		return closePos
	}
	offset := idx
	for offset < len(trimmed) && (trimmed[offset] == ' ' || trimmed[offset] == '\t') {
		offset++
	}
	if offset >= len(trimmed) {
		return closePos
	}
	pos := directiveNameColumn(startPos)
	pos.Column += offset
	return pos
}

// reportUnterminatedDirective emits the DWScript diagnostics for a directive whose
// closing '}' is missing.
//
// DWScript validates the switch name before it looks for the brace, so an unknown
// switch reports only "Compiler switch "X" unknown". A recognized switch reports
// `"}" expected`, anchored at its argument when it has one and at the switch name
// otherwise. {$INCLUDE} additionally reports its missing file name first.
func (l *Lexer) reportUnterminatedDirective(raw string, startPos Position) {
	// An unterminated directive swallows the rest of the file, so any conditional
	// still open is a consequence of it rather than a separate defect.
	l.directiveTruncated = true

	namePos := directiveNameColumn(startPos)

	name := raw
	argOffset := len(raw)
	if idx := strings.IndexFunc(raw, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	}); idx >= 0 {
		name = raw[:idx]
		argOffset = idx
		for argOffset < len(raw) && (raw[argOffset] == ' ' || raw[argOffset] == '\t') {
			argOffset++
		}
	}

	normalized := ident.Normalize(strings.TrimSpace(name))
	if normalized == "" {
		l.addDirectiveDiagnostic(`"}" expected`, namePos, LexerSeverityError, "")
		return
	}

	if _, ok := knownSwitches[normalized]; !ok {
		l.addDirectiveDiagnostic(
			fmt.Sprintf("Compiler switch %q unknown", strings.ToUpper(strings.TrimSpace(name))),
			namePos, LexerSeverityError, "")
		return
	}

	if normalized == "include" || normalized == "i" || normalized == "include_once" {
		if strings.TrimSpace(raw[argOffset:]) == "" {
			l.addDirectiveDiagnostic("Name of include file expected", namePos, LexerSeverityError, "")
		}
	}

	// The brace is expected where the argument starts, or at the switch name when the
	// directive has no argument.
	bracePos := namePos
	if strings.TrimSpace(raw[argOffset:]) != "" {
		bracePos.Column = namePos.Column + argOffset
	}
	l.addDirectiveDiagnostic(`"}" expected`, bracePos, LexerSeverityError, "")
}

// reportUnbalancedConditionals reports conditional directives still open at end of input.
// The diagnostic is anchored at the most recent directive of the innermost open frame,
// so {$if} ... {$else} <eof> points at the {$else} rather than the {$if}.
func (l *Lexer) reportUnbalancedConditionals() {
	if len(l.condStack) == 0 || l.directiveTruncated {
		l.condStack = nil
		return
	}
	frame := l.condStack[len(l.condStack)-1]
	l.addDirectiveDiagnostic("Unbalanced conditional directive",
		directiveNameColumn(frame.startPos), LexerSeverityError, "")
	l.condStack = nil
}
