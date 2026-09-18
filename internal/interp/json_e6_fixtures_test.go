package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
)

func TestJSONConversionCompatibility_Fixtures(t *testing.T) {
	const category = "JSONConnectorPass"
	for _, name := range []string{
		"global_var", "implicit_to_int2", "associative_array", "assign_static_to_dynamic",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, category, name+".pas")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
				t.Fatalf("%s/%s: %v: %s", category, name, got, detail)
			}
		})
	}
}
