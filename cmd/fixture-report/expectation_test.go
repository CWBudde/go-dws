package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestEvaluateOne_Expectations exercises scoring through a freshly built CLI,
// including the missing-.txt convention used by upstream's silent fixtures.
func TestEvaluateOne_Expectations(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(t.TempDir(), "dwscript")
	if runtime.GOOS == "windows" {
		cli += ".exe"
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", cli, "./cmd/dwscript")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	t.Run("missing expectation scores actual execution", func(t *testing.T) {
		testMissingFixtureExpectation(t, cli)
	})
	t.Run("excluded categories require an expectation", func(t *testing.T) {
		testExcludedFixtureExpectations(t, cli)
	})
	t.Run("existing expectation overrides silent default", func(t *testing.T) {
		testExistingFixtureExpectation(t, cli)
	})
	t.Run("eligible failure suite compares empty compile diagnostics", func(t *testing.T) {
		testCompileOnlyMissingExpectation(t, cli)
	})
	t.Run("invalid expectation is a failure", func(t *testing.T) {
		testInvalidFixtureExpectation(t, cli)
	})
	t.Run("Memory obj_local is silent at normal hints", func(t *testing.T) {
		testMemoryFixtureExpectation(t, cli, root)
	})
}

const expectationTestTimeout = 10 * time.Second

func writeExpectationFixture(t *testing.T, category, source string) workItem {
	t.Helper()
	dir := t.TempDir()
	item := workItem{
		category: category,
		pasFile:  filepath.Join(dir, "fixture.pas"),
		txtFile:  filepath.Join(dir, "fixture.txt"),
	}
	if err := os.WriteFile(item.pasFile, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return item
}

func assertExpectationVerdict(t *testing.T, got result, pass, fail, noExp bool) {
	t.Helper()
	if got.pass != pass || got.fail != fail || got.noExp != noExp {
		t.Fatalf("verdict = %+v; want pass=%v fail=%v noExp=%v", got, pass, fail, noExp)
	}
}

func testMissingFixtureExpectation(t *testing.T, cli string) {
	t.Helper()
	for _, tc := range []struct {
		name, source, kind string
		pass               bool
	}{
		{name: "silent success", source: "begin end.", pass: true},
		{name: "unexpected output", source: "PrintLn('unexpected');", kind: kindOutput},
		{name: "compile failure", source: "var x: Integer := 'hello';", kind: kindDiagnostics},
		{name: "runtime failure", source: "PrintLn('a'); var x := 1 div 0;", kind: kindMixed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := writeExpectationFixture(t, "SimpleScripts", tc.source)
			got := evaluateOne(cli, item, expectationTestTimeout, true)
			assertExpectationVerdict(t, got, tc.pass, !tc.pass, false)
			if tc.pass {
				if got.class != nil {
					t.Fatal("passing fixture must not receive a failure classification")
				}
				return
			}
			if got.class == nil || got.class.kind != tc.kind || got.class.distance == 0 {
				t.Fatalf("failure classification = %+v; want kind=%q and nonzero distance", got.class, tc.kind)
			}
		})
	}
}

func testExcludedFixtureExpectations(t *testing.T, cli string) {
	t.Helper()
	for _, category := range []string{"BuildScripts", "AutoFormat", "External", "DelegateLib", "FailureScripts"} {
		t.Run(category, func(t *testing.T) {
			item := writeExpectationFixture(t, category, "PrintLn('expected');")
			got := evaluateOne(cli, item, expectationTestTimeout, true)
			assertExpectationVerdict(t, got, false, false, true)
			if got.class != nil {
				t.Fatal("unscored fixture must not receive a failure classification")
			}

			// Even excluded categories are scored when an expectation exists.
			// FailureScripts uses compile-only mode, whose success is silent.
			expected := "expected\n"
			if category == "FailureScripts" {
				expected = ""
			}
			if err := os.WriteFile(item.txtFile, []byte(expected), 0o600); err != nil {
				t.Fatal(err)
			}
			assertExpectationVerdict(t, evaluateOne(cli, item, expectationTestTimeout, true), true, false, false)
		})
	}
}

func testExistingFixtureExpectation(t *testing.T, cli string) {
	t.Helper()
	item := writeExpectationFixture(t, "SimpleScripts", "PrintLn('expected');")
	if err := os.WriteFile(item.txtFile, []byte("expected\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertExpectationVerdict(t, evaluateOne(cli, item, expectationTestTimeout, true), true, false, false)
	if err := os.WriteFile(item.txtFile, []byte("different\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertExpectationVerdict(t, evaluateOne(cli, item, expectationTestTimeout, true), false, true, false)
}

func testCompileOnlyMissingExpectation(t *testing.T, cli string) {
	t.Helper()
	item := writeExpectationFixture(t, "JSFilterScriptsFail", "PrintLn('must not execute');")
	assertExpectationVerdict(t, evaluateOne(cli, item, expectationTestTimeout, true), true, false, false)
}

func testInvalidFixtureExpectation(t *testing.T, cli string) {
	t.Helper()
	for _, category := range []string{"SimpleScripts", "BuildScripts"} {
		for _, malformed := range []bool{false, true} {
			name := category + "/unreadable"
			if malformed {
				name = category + "/malformed UTF-16"
			}
			t.Run(name, func(t *testing.T) {
				item := writeExpectationFixture(t, category, "begin end.")
				if malformed {
					// The decoder replaces this truncated code unit with U+FFFD.
					// Its nonempty expectation must fail comparison, never skip.
					if err := os.WriteFile(item.txtFile, []byte{0xff, 0xfe, 0x00}, 0o600); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Mkdir(item.txtFile, 0o700); err != nil {
					// A directory gives a read failure even under privileged test
					// accounts that could bypass permission bits on a regular file.
					t.Fatal(err)
				}
				assertExpectationVerdict(t, evaluateOne(cli, item, expectationTestTimeout, true), false, true, false)
			})
		}
	}
}

func testMemoryFixtureExpectation(t *testing.T, cli string, root string) {
	t.Helper()
	if got := hintsLevelFor("Memory"); got != "normal" {
		t.Fatalf("Memory hint level = %q; want normal", got)
	}
	base := filepath.Join(root, fixturesBase, "Memory", "obj_local")
	item := workItem{category: "Memory", pasFile: base + ".pas", txtFile: base + ".txt"}
	assertExpectationVerdict(t, evaluateOne(cli, item, expectationTestTimeout, true), true, false, false)
}
