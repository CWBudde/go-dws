package units

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestNewUnit(t *testing.T) {
	tests := []struct {
		name     string
		unitName string
		filePath string
	}{
		{"Simple unit", "MyUnit", "/path/to/MyUnit.dws"},
		{"Lowercase unit", "myunit", "/path/to/myunit.dws"},
		{"Uppercase unit", "MYUNIT", "/path/to/MYUNIT.dws"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := NewUnit(tt.unitName, tt.filePath)

			if unit.Name != tt.unitName {
				t.Errorf("expected Name=%s, got %s", tt.unitName, unit.Name)
			}

			if unit.FilePath != tt.filePath {
				t.Errorf("expected FilePath=%s, got %s", tt.filePath, unit.FilePath)
			}

			if unit.Symbols == nil {
				t.Error("expected Symbols to be initialized, got nil")
			}

			if unit.Uses == nil {
				t.Error("expected Uses to be initialized, got nil")
			}

			if len(unit.Uses) != 0 {
				t.Errorf("expected empty Uses list, got %d items", len(unit.Uses))
			}
		})
	}
}

func TestUnitNormalizedName(t *testing.T) {
	tests := []struct {
		name         string
		unitName     string
		expectedNorm string
	}{
		{"Lowercase", "myunit", "myunit"},
		{"Uppercase", "MYUNIT", "myunit"},
		{"MixedCase", "MyUnit", "myunit"},
		{"CamelCase", "MySpecialUnit", "myspecialunit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := NewUnit(tt.unitName, "/test.dws")
			normalized := unit.NormalizedName()

			if normalized != tt.expectedNorm {
				t.Errorf("expected normalized name=%s, got %s", tt.expectedNorm, normalized)
			}
		})
	}
}

func TestUnitHasDependency(t *testing.T) {
	unit := NewUnit("TestUnit", "/test.dws")
	unit.Uses = []string{"System", "Math", "Graphics"}

	tests := []struct {
		name       string
		searchFor  string
		shouldFind bool
	}{
		{"Exact match", "System", true},
		{"Different case", "SYSTEM", true},
		{"Different case 2", "system", true},
		{"Another dependency", "Math", true},
		{"Not a dependency", "Database", false},
		{"Empty string", "", false},
		{"Partial match", "Sys", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unit.HasDependency(tt.searchFor)

			if result != tt.shouldFind {
				t.Errorf("HasDependency(%s) = %v, expected %v", tt.searchFor, result, tt.shouldFind)
			}
		})
	}
}

func TestUnitHasDependency_EmptyList(t *testing.T) {
	unit := NewUnit("TestUnit", "/test.dws")
	// Uses is empty by default

	if unit.HasDependency("System") {
		t.Error("expected HasDependency to return false for empty Uses list")
	}
}

func TestUnitString(t *testing.T) {
	tests := []struct {
		name          string
		setupUnit     func() *Unit
		expectedParts []string
	}{
		{
			name: "Minimal unit",
			setupUnit: func() *Unit {
				return NewUnit("MinimalUnit", "/test.dws")
			},
			expectedParts: []string{"unit MinimalUnit;", "end."},
		},
		{
			name: "Unit with dependencies",
			setupUnit: func() *Unit {
				unit := NewUnit("MyUnit", "/test.dws")
				unit.Uses = []string{"System", "Math"}
				return unit
			},
			expectedParts: []string{"unit MyUnit;", "uses System, Math;", "end."},
		},
		{
			name: "Unit with single dependency",
			setupUnit: func() *Unit {
				unit := NewUnit("SimpleUnit", "/test.dws")
				unit.Uses = []string{"System"}
				return unit
			},
			expectedParts: []string{"unit SimpleUnit;", "uses System;", "end."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := tt.setupUnit()
			result := unit.String()

			for _, part := range tt.expectedParts {
				if !strings.Contains(result, part) {
					t.Errorf("expected String() to contain %q, but it didn't. Got:\n%s", part, result)
				}
			}
		})
	}
}

func TestUnitSymbols(t *testing.T) {
	unit := NewUnit("TestUnit", "/test.dws")

	// Verify symbol table is initialized
	if unit.Symbols == nil {
		t.Fatal("expected Symbols to be initialized")
	}

	// Verify it's a valid symbol table by trying to use it
	// (This tests the integration with semantic.SymbolTable)
	st := unit.Symbols

	// Should be able to resolve symbols (even if empty)
	_, found := st.Resolve("NonExistent")
	if found {
		t.Error("expected Resolve to return false for non-existent symbol")
	}
}

func TestUnitFields(t *testing.T) {
	unit := NewUnit("TestUnit", "/test/path.dws")

	// Test that all fields can be set and retrieved
	t.Run("InterfaceSection", func(t *testing.T) {
		if unit.InterfaceSection != nil {
			t.Error("expected InterfaceSection to be nil initially")
		}
		// Could set it here if we had a BlockStatement to test with
	})

	t.Run("ImplementationSection", func(t *testing.T) {
		if unit.ImplementationSection != nil {
			t.Error("expected ImplementationSection to be nil initially")
		}
	})

	t.Run("InitializationSection", func(t *testing.T) {
		if unit.InitializationSection != nil {
			t.Error("expected InitializationSection to be nil initially")
		}
	})

	t.Run("FinalizationSection", func(t *testing.T) {
		if unit.FinalizationSection != nil {
			t.Error("expected FinalizationSection to be nil initially")
		}
	})
}

func TestUnitSymbolTable_Integration(t *testing.T) {
	// Test that the unit's symbol table works correctly with semantic package
	unit := NewUnit("TestUnit", "/test.dws")

	// This is a simple integration test to verify the symbol table works
	// More comprehensive symbol table tests are in the semantic package
	if unit.Symbols == nil {
		t.Fatal("expected Symbols to be initialized")
	}

	// Verify it's the correct type
	if _, ok := interface{}(unit.Symbols).(*semantic.SymbolTable); !ok {
		t.Error("expected Symbols to be a *semantic.SymbolTable")
	}
}
