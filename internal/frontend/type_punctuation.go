package frontend

import "github.com/cwbudde/go-dws/pkg/ast"

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

// refineDeferredPropertyCallDiagnostics removes only the provisional diagnostic
// belonging to a call resolved as a recoverable property read. Ordinary and
// parser-only calls keep their original compiler-stop diagnostic.
func refineDeferredPropertyCallDiagnostics(diags []Diagnostic, info *ast.SemanticInfo) []Diagnostic {
	kept := diags[:0]
	for _, diag := range diags {
		if diag.deferredCall != nil && info != nil && info.PropertyRead(diag.deferredCall) != nil {
			continue
		}
		if diag.deferredIndex != nil && info != nil && (info.IndexedPropertyRead(diag.deferredIndex) != nil || info.IsResolvedIndexedProperty(diag.deferredIndex)) {
			continue
		}
		kept = append(kept, diag)
	}
	return kept
}
