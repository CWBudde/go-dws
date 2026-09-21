package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestExplicitHelperFixtures(t *testing.T) {
	for _, tc := range []struct {
		category string
		name     string
		fail     bool
	}{
		{"HelpersPass", "declared_helper", false},
		{"HelpersFail", "helper_explicit", true},
	} {
		t.Run(tc.category+"/"+tc.name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, tc.category, tc.name+".pas")
			if got, detail := runFixtureTest(path, tc.fail, fixtureconfig.HintsLevel(tc.category)); got != testResultPassed {
				t.Fatalf("%s: %v: %s", tc.name, got, detail)
			}
		})
	}
}
