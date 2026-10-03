package frontend

import (
	"sort"
	"strings"
)

// restoreDeclarationDiagnosticOrder places declaration visibility diagnostics
// and parser stops at their source boundaries. Other diagnostics retain their
// relative order, including child-first errors and deferred checks. The caller
// must pass diagnostics from one source before merging imported units.
func restoreDeclarationDiagnosticOrder(diags []Diagnostic) {
	boundaries := make([]Diagnostic, 0)
	remaining := make([]Diagnostic, 0, len(diags))
	for _, diag := range diags {
		if isDeclarationVisibilityDiagnostic(diag) || (diag.Phase == PhaseParsing && diag.Stop) {
			boundaries = append(boundaries, diag)
		} else {
			remaining = append(remaining, diag)
		}
	}
	sort.SliceStable(boundaries, func(i, j int) bool {
		return diagnosticPositionBefore(boundaries[i], boundaries[j])
	})
	for _, boundary := range boundaries {
		at := len(remaining)
		for i, diag := range remaining {
			if diagnosticDeferredBucket(diag) != 0 ||
				(diag.Line > 0 && diagnosticPositionBefore(boundary, diag)) {
				at = i
				break
			}
		}
		remaining = append(remaining, Diagnostic{})
		copy(remaining[at+1:], remaining[at:])
		remaining[at] = boundary
	}
	copy(diags, remaining)
}

func diagnosticPositionBefore(left, right Diagnostic) bool {
	return left.Line < right.Line || (left.Line == right.Line && left.Column < right.Column)
}

func isDeclarationVisibilityDiagnostic(diag Diagnostic) bool {
	if diag.Phase != PhaseSemantic {
		return false
	}
	return strings.HasPrefix(diag.Render(), "Hint: Redundant specifier, visibility is already ") ||
		diag.Message == `Helpers do not supported "protected" visibility specifier` ||
		diag.Message == `Records do not supported "protected" visibility specifier`
}
