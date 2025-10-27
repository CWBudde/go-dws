package interp

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/units"
)

// TestSetUnitRegistry tests setting and getting the unit registry
func TestSetUnitRegistry(t *testing.T) {
	interp := New(&bytes.Buffer{})

	// Initially, registry should be nil
	if interp.GetUnitRegistry() != nil {
		t.Error("Expected nil registry initially")
	}

	// Set a registry
	registry := units.NewUnitRegistry([]string{"."})
	interp.SetUnitRegistry(registry)

	// Verify it was set
	if interp.GetUnitRegistry() != registry {
		t.Error("Expected registry to be set")
	}
}

// TestLoadUnit_NoRegistry tests that LoadUnit fails when no registry is set
func TestLoadUnit_NoRegistry(t *testing.T) {
	interp := New(&bytes.Buffer{})

	_, err := interp.LoadUnit("SomeUnit", nil)
	if err == nil {
		t.Fatal("Expected error when loading unit without registry")
	}

	if !strings.Contains(err.Error(), "registry not initialized") {
		t.Errorf("Expected 'registry not initialized' error, got: %v", err)
	}
}

// TestLoadUnit_Success tests successfully loading a unit
func TestLoadUnit_Success(t *testing.T) {
	interp := New(&bytes.Buffer{})

	// Get path to test fixtures
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	// Set up registry
	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load MathUtils unit
	unit, err := interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	if unit == nil {
		t.Fatal("Expected unit to be loaded")
	}

	if unit.Name != "MathUtils" {
		t.Errorf("Expected unit name 'MathUtils', got '%s'", unit.Name)
	}
}

// TestLoadUnit_Tracking tests that loaded units are tracked
func TestLoadUnit_Tracking(t *testing.T) {
	interp := New(&bytes.Buffer{})

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Initially no units loaded
	if len(interp.ListLoadedUnits()) != 0 {
		t.Error("Expected no loaded units initially")
	}

	// Load a unit
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load unit: %v", err)
	}

	// Verify it's tracked
	loaded := interp.ListLoadedUnits()
	if len(loaded) != 1 {
		t.Fatalf("Expected 1 loaded unit, got %d", len(loaded))
	}

	if loaded[0] != "MathUtils" {
		t.Errorf("Expected 'MathUtils' in loaded units, got '%s'", loaded[0])
	}

	// Check IsUnitLoaded
	if !interp.IsUnitLoaded("MathUtils") {
		t.Error("Expected MathUtils to be loaded")
	}

	if interp.IsUnitLoaded("NonExistent") {
		t.Error("Expected NonExistent to not be loaded")
	}
}

// TestLoadUnit_Caching tests that units are cached
func TestLoadUnit_Caching(t *testing.T) {
	interp := New(&bytes.Buffer{})

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load unit twice
	unit1, err1 := interp.LoadUnit("MathUtils", nil)
	if err1 != nil {
		t.Fatalf("Failed first load: %v", err1)
	}

	unit2, err2 := interp.LoadUnit("MathUtils", nil)
	if err2 != nil {
		t.Fatalf("Failed second load: %v", err2)
	}

	// Should return the same instance (cached)
	if unit1 != unit2 {
		t.Error("Expected cached unit instance on second load")
	}

	// Should still only show up once in loaded units
	loaded := interp.ListLoadedUnits()
	if len(loaded) != 1 {
		t.Errorf("Expected 1 loaded unit, got %d", len(loaded))
	}
}

// TestIsUnitLoaded_CaseInsensitive tests case-insensitive unit lookups
func TestIsUnitLoaded_CaseInsensitive(t *testing.T) {
	interp := New(&bytes.Buffer{})

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load with one case
	_, err = interp.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load unit: %v", err)
	}

	// Check with different cases
	if !interp.IsUnitLoaded("mathutils") {
		t.Error("Expected case-insensitive lookup to succeed")
	}

	if !interp.IsUnitLoaded("MATHUTILS") {
		t.Error("Expected case-insensitive lookup to succeed")
	}

	if !interp.IsUnitLoaded("MathUtils") {
		t.Error("Expected case-insensitive lookup to succeed")
	}
}

// TestLoadUnit_NotFound tests loading a non-existent unit
func TestLoadUnit_NotFound(t *testing.T) {
	interp := New(&bytes.Buffer{})

	registry := units.NewUnitRegistry([]string{"."})
	interp.SetUnitRegistry(registry)

	_, err := interp.LoadUnit("NonExistentUnit", nil)
	if err == nil {
		t.Fatal("Expected error when loading non-existent unit")
	}

	if !strings.Contains(err.Error(), "cannot load unit") {
		t.Errorf("Expected 'cannot load unit' error, got: %v", err)
	}
}

// TestListLoadedUnits_Order tests that units are listed in load order
func TestListLoadedUnits_Order(t *testing.T) {
	interp := New(&bytes.Buffer{})

	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := units.NewUnitRegistry([]string{fixturesPath})
	interp.SetUnitRegistry(registry)

	// Load units in specific order
	_, _ = interp.LoadUnit("MathUtils", nil)
	_, _ = interp.LoadUnit("StringUtils", nil)

	loaded := interp.ListLoadedUnits()

	if len(loaded) != 2 {
		t.Fatalf("Expected 2 loaded units, got %d", len(loaded))
	}

	// Verify order
	if loaded[0] != "MathUtils" {
		t.Errorf("Expected first unit to be 'MathUtils', got '%s'", loaded[0])
	}

	if loaded[1] != "StringUtils" {
		t.Errorf("Expected second unit to be 'StringUtils', got '%s'", loaded[1])
	}
}
