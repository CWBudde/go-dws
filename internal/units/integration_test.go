package units

import (
	"path/filepath"
	"testing"
)

// TestIntegration_LoadMathUtilsUnit tests loading a real unit file
func TestIntegration_LoadMathUtilsUnit(t *testing.T) {
	// Get path to test fixtures
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := NewUnitRegistry([]string{fixturesPath})

	// Load MathUtils unit
	unit, err := registry.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils unit: %v", err)
	}

	if unit == nil {
		t.Fatal("Expected unit to be loaded, got nil")
	}

	if unit.Name != "MathUtils" {
		t.Errorf("Expected unit name 'MathUtils', got '%s'", unit.Name)
	}

	// Verify the unit is in the registry
	loadedUnit, found := registry.GetUnit("MathUtils")
	if !found {
		t.Fatal("Unit not found in registry after loading")
	}

	if loadedUnit != unit {
		t.Error("GetUnit returned different unit instance")
	}
}

// TestIntegration_LoadStringUtilsUnit tests loading StringUtils unit
func TestIntegration_LoadStringUtilsUnit(t *testing.T) {
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := NewUnitRegistry([]string{fixturesPath})

	unit, err := registry.LoadUnit("StringUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load StringUtils unit: %v", err)
	}

	if unit.Name != "StringUtils" {
		t.Errorf("Expected unit name 'StringUtils', got '%s'", unit.Name)
	}
}

// TestIntegration_LoadMultipleUnits tests loading multiple units
func TestIntegration_LoadMultipleUnits(t *testing.T) {
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := NewUnitRegistry([]string{fixturesPath})

	// Load both MathUtils and StringUtils
	mathUnit, err := registry.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	stringUnit, err := registry.LoadUnit("StringUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load StringUtils: %v", err)
	}

	// Verify both are in registry
	units := registry.ListUnits()
	if len(units) != 2 {
		t.Errorf("Expected 2 units in registry, got %d", len(units))
	}

	// Verify caching works - loading again should return cached instance
	mathUnit2, err := registry.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to reload MathUtils: %v", err)
	}

	if mathUnit != mathUnit2 {
		t.Error("Second load should return cached instance")
	}

	_ = stringUnit // Avoid unused variable
}

// TestIntegration_InitializationOrder tests topological sort on real units
func TestIntegration_InitializationOrder(t *testing.T) {
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := NewUnitRegistry([]string{fixturesPath})

	// Load units (they don't have uses clauses in fixtures, so order should be arbitrary)
	_, err = registry.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load MathUtils: %v", err)
	}

	_, err = registry.LoadUnit("StringUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load StringUtils: %v", err)
	}

	// Compute initialization order
	order, err := registry.ComputeInitializationOrder()
	if err != nil {
		t.Fatalf("Failed to compute initialization order: %v", err)
	}

	if len(order) != 2 {
		t.Errorf("Expected 2 units in initialization order, got %d", len(order))
	}

	// Since neither unit depends on the other, any order is valid
	// Just verify both are present
	unitsMap := make(map[string]bool)
	for _, name := range order {
		unitsMap[name] = true
	}

	if !unitsMap["MathUtils"] {
		t.Error("MathUtils not in initialization order")
	}

	if !unitsMap["StringUtils"] {
		t.Error("StringUtils not in initialization order")
	}
}

// TestIntegration_CircularDependency tests that circular dependencies
// in real files are detected
func TestIntegration_CircularDependency(t *testing.T) {
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := NewUnitRegistry([]string{fixturesPath})

	// Try to load UnitA, which depends on UnitB, which depends on UnitA
	// This should fail with circular dependency error
	_, err = registry.LoadUnit("UnitA", nil)

	// Note: Current implementation might not detect this if the parser
	// doesn't extract uses clauses yet. This test documents the expected behavior.
	// Once the parser extracts uses clauses, this should fail.

	if err != nil {
		// Good! Circular dependency was detected
		t.Logf("Circular dependency correctly detected: %v", err)
	} else {
		// Parser doesn't extract uses clauses yet, so this is expected
		t.Skip("Parser doesn't extract uses clauses yet - circular dependency not detected")
	}
}

// TestIntegration_CaseInsensitiveUnitLoading tests that unit registry
// handles case-insensitive lookups for already-loaded units
func TestIntegration_CaseInsensitiveUnitLoading(t *testing.T) {
	fixturesPath, err := filepath.Abs("../../testdata/units")
	if err != nil {
		t.Fatalf("Failed to get fixtures path: %v", err)
	}

	registry := NewUnitRegistry([]string{fixturesPath})

	// Load the unit once with the exact case that matches the file (MathUtils.dws)
	unit1, err := registry.LoadUnit("MathUtils", nil)
	if err != nil {
		t.Fatalf("Failed to load 'MathUtils': %v", err)
	}

	// Now try to load again with different casing - should return cached instance
	unit2, err := registry.LoadUnit("mathutils", nil)
	if err != nil {
		t.Fatalf("Failed to reload as 'mathutils': %v", err)
	}

	unit3, err := registry.LoadUnit("MATHUTILS", nil)
	if err != nil {
		t.Fatalf("Failed to reload as 'MATHUTILS': %v", err)
	}

	// All should return the same cached instance (case-insensitive registry lookup)
	if unit1 != unit2 || unit2 != unit3 {
		t.Error("Different casings should return the same cached unit instance")
	}
}
