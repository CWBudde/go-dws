package frontend

// refineTypePunctuationDiagnostics replaces expression recovery at a token
// where semantic type resolution instead requires a record initializer's '('.
// Apply this before choosing the earliest stop, so the semantic stop owns its
// own emission order. Ordinary missing expressions retain their parser stop.
func refineTypePunctuationDiagnostics(diags []Diagnostic) []Diagnostic {
	positions := make(map[[2]int]bool)
	for _, diag := range diags {
		if diag.Phase == PhaseSemantic && diag.Stop && diag.Message == `"(" expected` {
			positions[[2]int{diag.Line, diag.Column}] = true
		}
	}
	kept := diags[:0]
	for _, diag := range diags {
		if diag.Phase == PhaseParsing && diag.Stop && diag.Message == "Expression expected" &&
			positions[[2]int{diag.Line, diag.Column}] {
			continue
		}
		kept = append(kept, diag)
	}
	return kept
}
