package lexer

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// handleIncludeMacro recognizes expression includes before ordinary directive
// processing. Its scanner preserves the consumed-token HotPos used by DWScript
// when a closing percent or brace is missing, including EOF after the final '%'.
func (l *Lexer) handleIncludeMacro(raw string, startPos Position, terminated bool) bool {
	input := raw
	if terminated {
		input += "}"
	}
	scanner := New(input)
	directive := scanner.NextToken()
	if !ident.Equal(directive.Literal, "i") && !ident.Equal(directive.Literal, "include") {
		return false
	}
	opening := scanner.NextToken()
	if opening.Type != PERCENT {
		return false
	}
	if l.isSkippingTokens() {
		l.noteSkippedDirective(startPos)
		l.skipIncludeMacro(scanner, opening, startPos)
		return true
	}
	macroPos := directiveTokenPosition(opening.Pos, startPos)
	name, last := readIncludeMacroName(scanner, opening)
	var replacement *Token
	if name == "" {
		l.addDirectiveDiagnostic("Include item expected", macroPos, SeverityError, "")
		tok := NewToken(STRING, "", directiveNameColumn(startPos))
		replacement = &tok // upstream simulates an empty string after this error
	} else {
		var known bool
		replacement, known = includeMacroValue(name, macroPos.Line, directiveNameColumn(startPos))
		if !known {
			l.addDirectiveDiagnostic(fmt.Sprintf("Include item %q unknown", name), macroPos, SeverityError, "")
			tok := NewToken(STRING, "", directiveNameColumn(startPos))
			replacement = &tok
		}
	}
	// Failed name/percent tests leave the inspected token current. Only EOF
	// falls back to the previously consumed token's HotPos.
	brace := scanner.Peek(0)
	if brace.Type != RBRACE {
		if brace.Type != EOF {
			last = brace
		}
		l.stopIncludeMacro(directiveTokenPosition(last.Pos, startPos))
		return true
	}
	l.directiveToken = replacement
	return true
}

// Inactive switches skip their arguments without interpreting the macro, but
// ReadUntilEndOrElseSwitch still requires the closing brace before continuing.
func (l *Lexer) skipIncludeMacro(scanner *Lexer, last Token, startPos Position) {
	for tok := scanner.NextToken(); tok.Type != EOF; tok = scanner.NextToken() {
		if tok.Type == RBRACE {
			return
		}
		last = tok
	}
	l.stopIncludeMacro(directiveTokenPosition(last.Pos, startPos))
}

func (l *Lexer) stopIncludeMacro(pos Position) {
	l.addDirectiveDiagnostic(`"}" expected`, pos, SeverityError, "")
	l.directiveTruncated = true
	l.stopped = true
}

// readIncludeMacroName consumes successful name/percent tests only. A rejected
// token remains available to the following brace test, just as TestAny/TestDelete
// leave FTok.GetToken unchanged on failure.
func readIncludeMacroName(scanner *Lexer, opening Token) (string, Token) {
	if opening.Type != PERCENT {
		return "", opening
	}
	item := scanner.Peek(0)
	if item.Type == EOF {
		return "", opening
	}
	if item.Type != IDENT && item.Type != FUNCTION {
		return "", opening
	}
	scanner.NextToken()
	closing := scanner.Peek(0)
	if closing.Type == PERCENT {
		scanner.NextToken()
		return item.Literal, closing
	}
	return "", item
}

// directiveTokenPosition maps a directive-body scanner position to the source.
func directiveTokenPosition(pos, startPos Position) Position {
	if pos.Line == 1 {
		pos.Column += startPos.Column + 1
	}
	pos.Line += startPos.Line - 1
	pos.Offset += startPos.Offset + 2
	pos.Hints = startPos.Hints
	return pos
}

// includeMacroValue implements source-line and clock substitutions. Other known
// macros need source identity, enclosing routine or executable version context
// not supplied to this package and retain their existing unsupported behavior.
func includeMacroValue(name string, line int, pos Position) (*Token, bool) {
	var value string
	kind := STRING
	switch ident.Normalize(name) {
	case "line":
		value = strconv.Itoa(line)
	case "linenum":
		value, kind = strconv.Itoa(line), INT
	case "time":
		value = time.Now().Format("15:04:05")
	case "date":
		value = time.Now().Format("2006-01-02")
	case "timestamp":
		value, kind = strconv.FormatInt(time.Now().Unix(), 10), INT
	case "file", "mainfile", "function", "exeversion":
		return nil, true
	default:
		return nil, false
	}
	tok := NewToken(kind, value, pos)
	return &tok, true
}
