package interp

import (
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestE17_Fixtures(t *testing.T) {
	for _, name := range []struct{ category, fixture string }{
		{"LambdaPass", "simple_func"},
		{"OverloadsPass", "overload_ambiguous_delegate"},
		{"OverloadsPass", "overload_class_method"},
	} {
		t.Run(name.fixture, func(t *testing.T) {
			path := filepath.Join(fixturesRoot, name.category, name.fixture+".pas")
			result, detail := runFixtureTest(path, false, semantic.HintsLevelPedantic)
			if result != testResultPassed {
				t.Fatalf("fixture failed: %s", detail)
			}
		})
	}
}
