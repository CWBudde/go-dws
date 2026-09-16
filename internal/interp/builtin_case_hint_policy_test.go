package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// Case-mismatch hints are pedantic-only, and upstream runs FunctionsMath at the
// hlStrict default. lcm.pas spells the same builtin "Lcm" and "lcm" and expects
// no hint for either, which no single declared spelling could satisfy.
func TestBuiltinCaseHints_FollowCategoryHintLevel(t *testing.T) {
	const category = "FunctionsMath"
	for _, name := range []string{"lcm", "nans"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%s/%s at the category hint level: %v: %s", category, name, got, detail)
			}
			if got, _ := runFixtureTest(path, false, semantic.HintsLevelPedantic); got == testResultPassed {
				t.Fatalf("%s/%s passed at pedantic hints; the builtin case hint is expected to fire there", category, name)
			}
		})
	}
}
