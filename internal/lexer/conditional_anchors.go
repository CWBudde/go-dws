package lexer

// inactiveConditional returns the frame whose inactive branch started the skip.
// Nested switches are scanned by that frame, rather than opening another scan.
func (l *Lexer) inactiveConditional() *conditionalFrame {
	for i := range l.condStack {
		if !l.condStack[i].active {
			return &l.condStack[i]
		}
	}
	return nil
}

func (l *Lexer) noteSkippedDirective(pos Position) {
	if frame := l.inactiveConditional(); frame != nil {
		frame.skipPos = directiveNameColumn(pos)
		frame.skipHasToken = false
	}
}

// conditionalArgumentEnd mirrors HotPos after ReadExpr consumes the condition.
func conditionalArgumentEnd(content string, startPos Position) Position {
	scanner := New(content)
	last := scanner.NextToken()
	for tok := scanner.NextToken(); tok.Type != EOF; tok = scanner.NextToken() {
		last = tok
	}
	pos := last.Pos
	pos.Line += startPos.Line - 1
	if last.Pos.Line == 1 {
		pos.Column += startPos.Column + 1
	}
	pos.Offset += startPos.Offset + 2
	pos.Hints = startPos.Hints
	return pos
}

// skipInactiveComment excludes comments from the skipped-token EOF distinction.
func (l *Lexer) skipInactiveComment() bool {
	switch {
	case l.ch == '{':
		l.skipBlockComment('{')
	case l.ch == '(' && l.peekChar() == '*':
		l.skipBlockComment('(')
	case l.ch == '/' && l.peekChar() == '/':
		l.skipLineComment()
	case l.ch == '/' && l.peekChar() == '*':
		l.skipCStyleComment()
	default:
		return false
	}
	return true
}
