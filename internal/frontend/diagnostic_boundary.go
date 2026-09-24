package frontend

import "strings"

// restoreStatementWarningOrder interleaves parser diagnostics with warnings
// emitted before reading a statement or a contract's message. Their anchors are
// known compilation boundaries, unlike (for example) an infinite-loop warning
// whose anchor precedes its body but which is emitted after reading that body.
// Preserve each phase's order, including semantic errors before a warning.
func restoreStatementWarningOrder(diags []Diagnostic) {
	for i := 0; i < len(diags); i++ {
		warning := diags[i]
		if !isStatementBoundaryWarning(warning) {
			continue
		}
		at := parserBoundaryIndex(diags[:i], warning)
		if at < 0 {
			continue
		}
		through := statementWarningGroupEnd(diags, i)
		// A stable partition moves the semantic prefix through this warning
		// ahead of the parser boundary, rather than sorting by display position.
		segment := append([]Diagnostic(nil), diags[at:through+1]...)
		n := at
		for _, diag := range segment {
			if diag.Phase == PhaseSemantic {
				diags[n] = diag
				n++
			}
		}
		for _, diag := range segment {
			if diag.Phase != PhaseSemantic {
				diags[n] = diag
				n++
			}
		}
	}
}

func isStatementBoundaryWarning(diag Diagnostic) bool {
	if diag.Phase != PhaseSemantic || diag.Severity != SeverityWarning || diag.Line == 0 {
		return false
	}
	rendered := diag.Render()
	return strings.HasPrefix(rendered, "Warning: Constant condition [") ||
		strings.HasPrefix(rendered, "Warning: Unreachable code [")
}

// parserBoundaryIndex finds the first parser token not yet reached when the
// warning was emitted.
func parserBoundaryIndex(diags []Diagnostic, warning Diagnostic) int {
	for i, parser := range diags {
		if parser.Phase == PhaseParsing && !parser.lexerDirective && parser.Line > 0 &&
			(parser.Line > warning.Line || (parser.Line == warning.Line && parser.Column >= warning.Column)) {
			return i
		}
	}
	return -1
}

func statementWarningGroupEnd(diags []Diagnostic, at int) int {
	warning := diags[at]
	if !strings.HasPrefix(warning.Render(), "Warning: Constant condition [") {
		return at
	}
	// Message validation follows a constant-condition warning before the
	// parser reads the separator. Its diagnostics share the condition anchor.
	for at+1 < len(diags) {
		next := diags[at+1]
		if next.Phase != PhaseSemantic || next.Line != warning.Line || next.Column != warning.Column {
			break
		}
		at++
	}
	return at
}
