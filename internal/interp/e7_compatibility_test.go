package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestE7Compatibility_Fixtures(t *testing.T) {
	const category = "SimpleScripts"
	for _, name := range []string{"ignore_result", "assert_variant", "assert"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%s/%s: %v: %s", category, name, got, detail)
			}
		})
	}
}
