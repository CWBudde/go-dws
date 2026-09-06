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

func TestBinaryIsStale_MissingBinary(t *testing.T) {
	if _, _, err := binaryIsStale(filepath.Join(t.TempDir(), "missing"), nil); err == nil {
		t.Fatal("expected an error for a missing binary")
	}
}
