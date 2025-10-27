package units

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewUnitRegistry(t *testing.T) {
	t.Run("With search paths", func(t *testing.T) {
		paths := []string{"/path1", "/path2"}
		registry := NewUnitRegistry(paths)

		if registry == nil {
			t.Fatal("expected registry to be created")
		}

		if len(registry.searchPaths) != 2 {
			t.Errorf("expected 2 search paths, got %d", len(registry.searchPaths))
		}

		if registry.units == nil {
			t.Error("expected units map to be initialized")
		}

		if registry.loading == nil {
			t.Error("expected loading map to be initialized")
		}
	})

	t.Run("Without search paths", func(t *testing.T) {
		registry := NewUnitRegistry(nil)

		if len(registry.searchPaths) != 1 || registry.searchPaths[0] != "." {
			t.Errorf("expected default search path [.], got %v", registry.searchPaths)
		}
	})

	t.Run("Empty search paths", func(t *testing.T) {
		registry := NewUnitRegistry([]string{})

		if len(registry.searchPaths) != 0 {
			t.Errorf("expected 0 search paths, got %d", len(registry.searchPaths))
		}
	})
}

func TestRegisterUnit(t *testing.T) {
	registry := NewUnitRegistry([]string{"."})

	t.Run("Register first unit", func(t *testing.T) {
		unit := NewUnit("TestUnit", "/test.dws")
		err := registry.RegisterUnit("TestUnit", unit)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify it was registered
		retrieved, exists := registry.GetUnit("TestUnit")
		if !exists {
			t.Error("expected unit to be registered")
		}

		if retrieved != unit {
			t.Error("retrieved unit is not the same as registered unit")
		}
	})

	t.Run("Register duplicate unit", func(t *testing.T) {
		unit1 := NewUnit("DuplicateUnit", "/test1.dws")
		unit2 := NewUnit("DuplicateUnit", "/test2.dws")

		err := registry.RegisterUnit("DuplicateUnit", unit1)
		if err != nil {
			t.Fatalf("unexpected error on first registration: %v", err)
		}

		err = registry.RegisterUnit("DuplicateUnit", unit2)
		if err == nil {
			t.Error("expected error when registering duplicate unit")
		}

		if !strings.Contains(err.Error(), "already registered") {
			t.Errorf("expected 'already registered' error, got: %v", err)
		}
	})

	t.Run("Case insensitive registration", func(t *testing.T) {
		registry := NewUnitRegistry([]string{"."})
		unit := NewUnit("MyUnit", "/test.dws")

		err := registry.RegisterUnit("MyUnit", unit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Try to register with different case
		unit2 := NewUnit("MYUNIT", "/test2.dws")
		err = registry.RegisterUnit("MYUNIT", unit2)
		if err == nil {
			t.Error("expected error when registering unit with different case")
		}
	})
}

func TestGetUnit(t *testing.T) {
	registry := NewUnitRegistry([]string{"."})
	unit := NewUnit("TestUnit", "/test.dws")
	registry.RegisterUnit("TestUnit", unit)

	tests := []struct {
		name       string
		searchName string
		shouldFind bool
	}{
		{"Exact case", "TestUnit", true},
		{"Lowercase", "testunit", true},
		{"Uppercase", "TESTUNIT", true},
		{"Mixed case", "TeStUnIt", true},
		{"Non-existent", "NonExistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, exists := registry.GetUnit(tt.searchName)

			if exists != tt.shouldFind {
				t.Errorf("GetUnit(%s) exists=%v, expected %v", tt.searchName, exists, tt.shouldFind)
			}

			if tt.shouldFind && retrieved != unit {
				t.Error("retrieved unit is not the expected unit")
			}
		})
	}
}

func TestUnregisterUnit(t *testing.T) {
	registry := NewUnitRegistry([]string{"."})
	unit := NewUnit("TestUnit", "/test.dws")
	registry.RegisterUnit("TestUnit", unit)

	// Verify it's registered
	_, exists := registry.GetUnit("TestUnit")
	if !exists {
		t.Fatal("unit should be registered")
	}

	// Unregister
	registry.UnregisterUnit("TestUnit")

	// Verify it's gone
	_, exists = registry.GetUnit("TestUnit")
	if exists {
		t.Error("unit should be unregistered")
	}
}

func TestClear(t *testing.T) {
	registry := NewUnitRegistry([]string{"."})

	// Register multiple units
	registry.RegisterUnit("Unit1", NewUnit("Unit1", "/unit1.dws"))
	registry.RegisterUnit("Unit2", NewUnit("Unit2", "/unit2.dws"))
	registry.RegisterUnit("Unit3", NewUnit("Unit3", "/unit3.dws"))

	// Clear
	registry.Clear()

	// Verify all are gone
	if _, exists := registry.GetUnit("Unit1"); exists {
		t.Error("Unit1 should be cleared")
	}
	if _, exists := registry.GetUnit("Unit2"); exists {
		t.Error("Unit2 should be cleared")
	}
	if _, exists := registry.GetUnit("Unit3"); exists {
		t.Error("Unit3 should be cleared")
	}
}

func TestListUnits(t *testing.T) {
	registry := NewUnitRegistry([]string{"."})

	t.Run("Empty registry", func(t *testing.T) {
		units := registry.ListUnits()
		if len(units) != 0 {
			t.Errorf("expected 0 units, got %d", len(units))
		}
	})

	t.Run("With units", func(t *testing.T) {
		registry.RegisterUnit("Unit1", NewUnit("Unit1", "/unit1.dws"))
		registry.RegisterUnit("Unit2", NewUnit("Unit2", "/unit2.dws"))

		units := registry.ListUnits()
		if len(units) != 2 {
			t.Errorf("expected 2 units, got %d", len(units))
		}

		// Verify unit names are present (order doesn't matter)
		found := make(map[string]bool)
		for _, name := range units {
			found[name] = true
		}

		if !found["Unit1"] {
			t.Error("expected Unit1 in list")
		}
		if !found["Unit2"] {
			t.Error("expected Unit2 in list")
		}
	})
}

func TestLoadUnit_NotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	registry := NewUnitRegistry([]string{tempDir})

	_, err := registry.LoadUnit("NonExistentUnit", nil)
	if err == nil {
		t.Error("expected error when loading non-existent unit")
	}

	if !strings.Contains(err.Error(), "cannot load unit") {
		t.Errorf("expected 'cannot load unit' error, got: %v", err)
	}
}

func TestLoadUnit_SimpleUnit(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a simple unit file
	// Note: Without proper unit syntax parser (tasks 9.108-9.110), we use a simple program
	unitContent := `var x: Integer;
x := 42;`

	unitPath := filepath.Join(tempDir, "SimpleUnit.dws")
	err := os.WriteFile(unitPath, []byte(unitContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test unit file: %v", err)
	}

	registry := NewUnitRegistry([]string{tempDir})

	// Load the unit
	unit, err := registry.LoadUnit("SimpleUnit", nil)
	if err != nil {
		t.Fatalf("unexpected error loading unit: %v", err)
	}

	if unit == nil {
		t.Fatal("expected unit to be loaded")
	}

	if unit.Name != "SimpleUnit" {
		t.Errorf("expected unit name 'SimpleUnit', got '%s'", unit.Name)
	}

	if unit.FilePath != unitPath {
		t.Errorf("expected file path '%s', got '%s'", unitPath, unit.FilePath)
	}

	// Verify it's cached
	unit2, err := registry.LoadUnit("SimpleUnit", nil)
	if err != nil {
		t.Fatalf("unexpected error loading cached unit: %v", err)
	}

	if unit2 != unit {
		t.Error("expected cached unit to be the same instance")
	}
}

func TestLoadUnit_CircularDependency(t *testing.T) {
	registry := NewUnitRegistry([]string{"."})

	// Simulate circular dependency by manually setting loading state
	registry.loading["unita"] = true

	// Try to load unitA again (simulating circular dependency)
	tempDir := t.TempDir()
	unitPath := filepath.Join(tempDir, "UnitA.dws")
	os.WriteFile(unitPath, []byte("var x: Integer;"), 0644)

	registry.searchPaths = []string{tempDir}

	_, err := registry.LoadUnit("UnitA", nil)
	if err == nil {
		t.Error("expected error for circular dependency")
	}

	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %v", err)
	}
}

func TestLoadUnit_ParseError(t *testing.T) {
	tempDir := t.TempDir()

	// Create a unit file with invalid syntax
	unitContent := `begin var x := ; end.` // Invalid syntax
	unitPath := filepath.Join(tempDir, "InvalidUnit.dws")
	err := os.WriteFile(unitPath, []byte(unitContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test unit file: %v", err)
	}

	registry := NewUnitRegistry([]string{tempDir})

	_, err = registry.LoadUnit("InvalidUnit", nil)
	if err == nil {
		t.Error("expected parse error")
	}

	if !strings.Contains(err.Error(), "parse errors") {
		t.Errorf("expected 'parse errors' message, got: %v", err)
	}
}

func TestLoadUnit_CustomSearchPaths(t *testing.T) {
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	// Create unit in tempDir2
	unitContent := `var x: Integer;`
	unitPath := filepath.Join(tempDir2, "MyUnit.dws")
	err := os.WriteFile(unitPath, []byte(unitContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test unit file: %v", err)
	}

	// Registry has tempDir1 as default, but we pass tempDir2 to LoadUnit
	registry := NewUnitRegistry([]string{tempDir1})

	unit, err := registry.LoadUnit("MyUnit", []string{tempDir2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if unit == nil {
		t.Fatal("expected unit to be loaded")
	}

	if unit.FilePath != unitPath {
		t.Errorf("expected file path '%s', got '%s'", unitPath, unit.FilePath)
	}
}

// TestComputeInitializationOrder tests the topological sort for unit initialization
func TestComputeInitializationOrder(t *testing.T) {
	t.Run("Simple linear dependency", func(t *testing.T) {
		registry := NewUnitRegistry([]string{"."})

		// Create units: A -> B -> C (C depends on B, B depends on A)
		unitA := NewUnit("A", "/path/a.dws")
		unitB := NewUnit("B", "/path/b.dws")
		unitB.Uses = []string{"A"}
		unitC := NewUnit("C", "/path/c.dws")
		unitC.Uses = []string{"B"}

		registry.RegisterUnit("A", unitA)
		registry.RegisterUnit("B", unitB)
		registry.RegisterUnit("C", unitC)

		order, err := registry.ComputeInitializationOrder()
		if err != nil {
			t.Fatalf("ComputeInitializationOrder() failed: %v", err)
		}

		// Expected order: A, B, C (dependencies first)
		expected := []string{"A", "B", "C"}
		if len(order) != len(expected) {
			t.Fatalf("Expected %d units, got %d", len(expected), len(order))
		}

		for i, unitName := range expected {
			if order[i] != unitName {
				t.Errorf("Position %d: expected '%s', got '%s'", i, unitName, order[i])
			}
		}
	})

	t.Run("Diamond dependency", func(t *testing.T) {
		registry := NewUnitRegistry([]string{"."})

		// Create diamond: D depends on B and C, B and C both depend on A
		//     A
		//    / \
		//   B   C
		//    \ /
		//     D
		unitA := NewUnit("A", "/path/a.dws")
		unitB := NewUnit("B", "/path/b.dws")
		unitB.Uses = []string{"A"}
		unitC := NewUnit("C", "/path/c.dws")
		unitC.Uses = []string{"A"}
		unitD := NewUnit("D", "/path/d.dws")
		unitD.Uses = []string{"B", "C"}

		registry.RegisterUnit("A", unitA)
		registry.RegisterUnit("B", unitB)
		registry.RegisterUnit("C", unitC)
		registry.RegisterUnit("D", unitD)

		order, err := registry.ComputeInitializationOrder()
		if err != nil {
			t.Fatalf("ComputeInitializationOrder() failed: %v", err)
		}

		// A must come first, D must come last
		// B and C can be in any order (both depend only on A)
		if len(order) != 4 {
			t.Fatalf("Expected 4 units, got %d", len(order))
		}

		if order[0] != "A" {
			t.Errorf("Expected 'A' first, got '%s'", order[0])
		}

		if order[3] != "D" {
			t.Errorf("Expected 'D' last, got '%s'", order[3])
		}

		// B and C should be in the middle (positions 1 and 2)
		middleUnits := map[string]bool{order[1]: true, order[2]: true}
		if !middleUnits["B"] || !middleUnits["C"] {
			t.Errorf("Expected B and C in middle positions, got %v", order[1:3])
		}
	})

	t.Run("No dependencies", func(t *testing.T) {
		registry := NewUnitRegistry([]string{"."})

		unitA := NewUnit("A", "/path/a.dws")
		unitB := NewUnit("B", "/path/b.dws")
		unitC := NewUnit("C", "/path/c.dws")

		registry.RegisterUnit("A", unitA)
		registry.RegisterUnit("B", unitB)
		registry.RegisterUnit("C", unitC)

		order, err := registry.ComputeInitializationOrder()
		if err != nil {
			t.Fatalf("ComputeInitializationOrder() failed: %v", err)
		}

		// All units can be in any order
		if len(order) != 3 {
			t.Fatalf("Expected 3 units, got %d", len(order))
		}
	})

	t.Run("Circular dependency detection", func(t *testing.T) {
		registry := NewUnitRegistry([]string{"."})

		// Create circular: A -> B -> C -> A
		unitA := NewUnit("A", "/path/a.dws")
		unitA.Uses = []string{"C"}
		unitB := NewUnit("B", "/path/b.dws")
		unitB.Uses = []string{"A"}
		unitC := NewUnit("C", "/path/c.dws")
		unitC.Uses = []string{"B"}

		registry.RegisterUnit("A", unitA)
		registry.RegisterUnit("B", unitB)
		registry.RegisterUnit("C", unitC)

		_, err := registry.ComputeInitializationOrder()
		if err == nil {
			t.Fatal("Expected error for circular dependency, got nil")
		}

		if !strings.Contains(err.Error(), "circular") && !strings.Contains(err.Error(), "cycle") {
			t.Errorf("Error should mention circular/cycle, got: %v", err)
		}
	})
}
