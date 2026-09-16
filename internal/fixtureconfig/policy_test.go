package fixtureconfig

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

// upstreamPedanticCategories re-derives the pedantic set from the bundled
// UScriptTests.pas, the only output-comparison runner that raises
// Config.HintsLevel. A new upstream drop that moves a directory between runners
// then fails here instead of silently changing which hints fixtures expect.
func TestHintsLevel_MatchesBundledUScriptTests(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "UScriptTests.pas"))
	if err != nil {
		t.Skipf("bundled upstream runner unavailable: %v", err)
	}
	if !regexp.MustCompile(`(?i)HintsLevel\s*:=\s*hlPedantic`).Match(source) {
		t.Fatal("UScriptTests.pas no longer sets HintsLevel := hlPedantic")
	}
	collected := map[string]bool{}
	for _, m := range regexp.MustCompile(`CollectFiles\(basePath\+'([A-Za-z0-9_]+)'`).FindAllSubmatch(source, -1) {
		collected[string(m[1])] = true
	}
	if len(collected) == 0 {
		t.Fatal("no categories parsed out of UScriptTests.pas")
	}
	for category := range collected {
		if got := HintsLevel(category); got != semantic.HintsLevelPedantic {
			t.Errorf("HintsLevel(%q) = %v; UScriptTests collects it, want pedantic", category, got)
		}
	}
	for category := range pedanticCategories {
		if !collected[category] {
			t.Errorf("HintsLevel(%q) is pedantic but UScriptTests does not collect it", category)
		}
	}
}

// Case-mismatch hints are pedantic-only, so a category run by any other upstream
// runner must stay at the hlStrict default. FunctionsMath/lcm.pas is the proof:
// it spells the same builtin "Lcm" and "lcm" and expects no hint for either.
func TestHintsLevel_OtherRunnersStayStrict(t *testing.T) {
	for _, category := range []string{"Algorithms", "ClassesLib", "FunctionsMath", "FunctionsString", "Memory"} {
		if got := HintsLevel(category); got != semantic.HintsLevelStrict {
			t.Errorf("HintsLevel(%q) = %v; want strict", category, got)
		}
	}
}
