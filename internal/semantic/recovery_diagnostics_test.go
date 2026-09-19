package semantic

import (
	"strings"
	"testing"
)

// TestIfExpression_NestedParserRecoveryAddsNoFollowUp checks that a branch holding the
// parser's recovery placeholder below its root (here inside a binary expression) gets no
// "invalid ... expression" follow-up: the parser already reported the compiler stop.
func TestIfExpression_NestedParserRecoveryAddsNoFollowUp(t *testing.T) {
	sources := []string{
		"var x := if true then (1 + ) else 0;",
		"var x := if true then 1 else (1 + );",
	}
	for _, source := range sources {
		analyzer := parseAndAnalyze(t, source)
		for _, err := range analyzer.Errors() {
			if strings.Contains(err, "if-then-else") {
				t.Errorf("%q: unexpected follow-up %q", source, err)
			}
		}
	}
}
