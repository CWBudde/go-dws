package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHintsLevelFor(t *testing.T) {
	if hintsLevelFor("Algorithms") != "normal" || hintsLevelFor("FunctionsString") != "normal" {
		t.Fatal("Algorithms and FunctionsString run at hints=normal, mirroring internal/interp/fixture_test.go hintsLevelOverrides")
	}
	if hintsLevelFor("SimpleScripts") != "pedantic" || hintsLevelFor("FailureScripts") != "pedantic" {
		t.Fatal("every other category runs at hints=pedantic")
	}
}

func TestCollectItems_InScopeDropsHostLibrariesButNotANamedOne(t *testing.T) {
	// A miniature fixture tree: collectItems reads the real directory layout, so the
	// test supplies one rather than depending on the repository's.
	dir := t.TempDir()
	for _, cat := range []string{"SimpleScripts", "DataBaseLib"} {
		catDir := filepath.Join(dir, fixturesBase, cat)
		if err := os.MkdirAll(catDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(catDir, "a.pas"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	all, err := collectItems("", false)
	if err != nil {
		t.Fatalf("collectItems: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("full run collected %d fixtures, want both categories", len(all))
	}

	inScope, err := collectItems("", true)
	if err != nil {
		t.Fatalf("collectItems in-scope: %v", err)
	}
	if len(inScope) != 1 || inScope[0].category != "SimpleScripts" {
		t.Fatalf("in-scope run = %v, want only SimpleScripts", inScope)
	}

	// Naming a category explicitly overrides the filter: --category says what to
	// measure, --in-scope only prunes a whole-tree run.
	named, err := collectItems("DataBaseLib", true)
	if err != nil {
		t.Fatalf("collectItems named: %v", err)
	}
	if len(named) != 1 {
		t.Fatal("an explicitly named out-of-scope category must still be collected")
	}
}

func TestIsErrorCategory(t *testing.T) {
	for _, c := range []string{"FailureScripts", "COMConnectorFailure", "HelpersFail", "InterfacesFail"} {
		if !isErrorCategory(c) {
			t.Errorf("%s must be an error category", c)
		}
	}
	for _, c := range []string{"SimpleScripts", "HelpersPass", "FailureScriptsX"} {
		if isErrorCategory(c) {
			t.Errorf("%s must not be an error category", c)
		}
	}
}

func TestBinaryIsStale(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "cli")
	src := filepath.Join(dir, "a.go")
	if err := os.WriteFile(bin, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(bin, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// Directory creation timestamps can be newer than a new file's timestamp
	// on disk-backed temporary directories. Make the intended ordering explicit.
	sourceTime := time.Now().Add(time.Minute)
	if err := os.Chtimes(src, sourceTime, sourceTime); err != nil {
		t.Fatal(err)
	}
	stale, newest, err := binaryIsStale(bin, []string{src})
	if err != nil {
		t.Fatal(err)
	}
	if !stale || newest != src {
		t.Fatalf("newer source must mark binary stale (stale=%v newest=%q)", stale, newest)
	}
	now := time.Now().Add(time.Hour)
	if err := os.Chtimes(bin, now, now); err != nil {
		t.Fatal(err)
	}
	if stale, _, _ = binaryIsStale(bin, []string{src}); stale {
		t.Fatal("binary newer than every source must not be stale")
	}
}

func TestBinaryIsStale_DeletedSource(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "cli")
	if err := os.WriteFile(bin, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(bin, future, future); err != nil {
		t.Fatal(err)
	}
	// A source git still lists but that is gone from the working tree.
	gone := filepath.Join(dir, "gone.go")
	stale, newest, err := binaryIsStale(bin, []string{gone})
	if err != nil {
		t.Fatal(err)
	}
	if !stale || newest != gone+" (deleted)" {
		t.Fatalf("missing tracked source must mark binary stale (stale=%v newest=%q)", stale, newest)
	}
}

func TestBinaryIsStale_RemovedFromDirectory(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(srcDir, "keep.go")
	removed := filepath.Join(srcDir, "removed.go")
	for _, f := range []string{keep, removed} {
		if err := os.WriteFile(f, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(dir, "cli")
	if err := os.WriteFile(bin, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	for _, f := range []string{keep, removed, srcDir, bin} {
		if err := os.Chtimes(f, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if stale, newest, _ := binaryIsStale(bin, []string{keep}); stale {
		t.Fatalf("nothing changed, must not be stale (newest=%q)", newest)
	}
	// `git rm removed.go`: the file leaves the source list, only its directory changes.
	if err := os.Remove(removed); err != nil {
		t.Fatal(err)
	}
	stale, newest, err := binaryIsStale(bin, []string{keep})
	if err != nil {
		t.Fatal(err)
	}
	if !stale || newest != srcDir+string(filepath.Separator) {
		t.Fatalf("removing a file from a source directory must mark binary stale (stale=%v newest=%q)", stale, newest)
	}
}

func TestBinaryIsStale_MissingBinary(t *testing.T) {
	if _, _, err := binaryIsStale(filepath.Join(t.TempDir(), "missing"), nil); err == nil {
		t.Fatal("expected an error for a missing binary")
	}
}
