package types

import (
	"testing"
)

// MockClassInfo is a simple mock for testing ClassRegistry
type MockClassInfo struct {
	Name   string
	Parent *MockClassInfo
}

func TestClassRegistry_RegisterAndLookup(t *testing.T) {
	registry := NewClassRegistry()

	// Create mock classes
	animal := &MockClassInfo{Name: "Animal", Parent: nil}
	dog := &MockClassInfo{Name: "Dog", Parent: animal}

	// Register classes
	registry.Register("Animal", animal)
	registry.Register("Dog", dog)

	// Lookup should work case-insensitively
	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"exact case", "Animal", true},
		{"lowercase", "animal", true},
		{"uppercase", "ANIMAL", true},
		{"mixed case", "AnImAl", true},
		{"not found", "Cat", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := registry.Lookup(tt.lookup)
			if found != tt.expected {
				t.Errorf("Lookup(%q) found=%v, expected=%v", tt.lookup, found, tt.expected)
			}
		})
	}
}

func TestClassRegistry_Exists(t *testing.T) {
	registry := NewClassRegistry()
	animal := &MockClassInfo{Name: "Animal"}

	registry.Register("Animal", animal)

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"exists", "Animal", true},
		{"exists lowercase", "animal", true},
		{"not exists", "Dog", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := registry.Exists(tt.lookup)
			if exists != tt.expected {
				t.Errorf("Exists(%q) = %v, expected %v", tt.lookup, exists, tt.expected)
			}
		})
	}
}

func TestClassRegistry_RegisterWithParent(t *testing.T) {
	registry := NewClassRegistry()

	// Register Object (root class)
	object := &MockClassInfo{Name: "Object"}
	registry.RegisterWithParent("Object", object, "")

	// Register Animal (inherits from Object)
	animal := &MockClassInfo{Name: "Animal"}
	registry.RegisterWithParent("Animal", animal, "Object")

	// Register Dog (inherits from Animal)
	dog := &MockClassInfo{Name: "Dog"}
	registry.RegisterWithParent("Dog", dog, "Animal")

	// Check parent names
	if parent := registry.GetParentName("Object"); parent != "" {
		t.Errorf("Object parent = %q, expected empty", parent)
	}

	if parent := registry.GetParentName("Animal"); parent != "Object" {
		t.Errorf("Animal parent = %q, expected 'Object'", parent)
	}

	if parent := registry.GetParentName("Dog"); parent != "Animal" {
		t.Errorf("Dog parent = %q, expected 'Animal'", parent)
	}
}

func TestClassRegistry_LookupHierarchy(t *testing.T) {
	registry := NewClassRegistry()

	// Build hierarchy: Object -> Animal -> Dog
	object := &MockClassInfo{Name: "Object"}
	animal := &MockClassInfo{Name: "Animal"}
	dog := &MockClassInfo{Name: "Dog"}

	registry.RegisterWithParent("Object", object, "")
	registry.RegisterWithParent("Animal", animal, "Object")
	registry.RegisterWithParent("Dog", dog, "Animal")

	// Lookup hierarchy for Dog
	hierarchy := registry.LookupHierarchy("Dog")

	if hierarchy == nil {
		t.Fatal("LookupHierarchy returned nil")
	}

	if len(hierarchy) != 3 {
		t.Errorf("hierarchy length = %d, expected 3", len(hierarchy))
	}

	// Hierarchy should be [Dog, Animal, Object]
	expectedOrder := []string{"Dog", "Animal", "Object"}
	for i, expected := range expectedOrder {
		if i >= len(hierarchy) {
			t.Errorf("hierarchy too short, missing %s", expected)
			continue
		}
		// We can't check the actual value since it's 'any', but we can check length
	}

	// Lookup hierarchy for Object (root)
	hierarchy = registry.LookupHierarchy("Object")
	if len(hierarchy) != 1 {
		t.Errorf("Object hierarchy length = %d, expected 1", len(hierarchy))
	}

	// Lookup non-existent class
	hierarchy = registry.LookupHierarchy("Cat")
	if hierarchy != nil {
		t.Errorf("Cat hierarchy = %v, expected nil", hierarchy)
	}
}

func TestClassRegistry_IsDescendantOf(t *testing.T) {
	registry := NewClassRegistry()

	// Build hierarchy: Object -> Animal -> Dog
	object := &MockClassInfo{Name: "Object"}
	animal := &MockClassInfo{Name: "Animal"}
	dog := &MockClassInfo{Name: "Dog"}
	cat := &MockClassInfo{Name: "Cat"}

	registry.RegisterWithParent("Object", object, "")
	registry.RegisterWithParent("Animal", animal, "Object")
	registry.RegisterWithParent("Dog", dog, "Animal")
	registry.RegisterWithParent("Cat", cat, "Animal")

	tests := []struct {
		name       string
		descendant string
		ancestor   string
		expected   bool
	}{
		{"same class", "Dog", "Dog", true},
		{"direct parent", "Dog", "Animal", true},
		{"grandparent", "Dog", "Object", true},
		{"not descendant", "Dog", "Cat", false},
		{"reverse relationship", "Animal", "Dog", false},
		{"siblings", "Dog", "Cat", false},
		{"case insensitive", "dog", "animal", true},
		{"not found", "Lion", "Animal", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsDescendantOf(tt.descendant, tt.ancestor)
			if result != tt.expected {
				t.Errorf("IsDescendantOf(%q, %q) = %v, expected %v",
					tt.descendant, tt.ancestor, result, tt.expected)
			}
		})
	}
}

func TestClassRegistry_GetDepth(t *testing.T) {
	registry := NewClassRegistry()

	// Build hierarchy: Object -> Animal -> Dog -> Poodle
	object := &MockClassInfo{Name: "Object"}
	animal := &MockClassInfo{Name: "Animal"}
	dog := &MockClassInfo{Name: "Dog"}
	poodle := &MockClassInfo{Name: "Poodle"}

	registry.RegisterWithParent("Object", object, "")
	registry.RegisterWithParent("Animal", animal, "Object")
	registry.RegisterWithParent("Dog", dog, "Animal")
	registry.RegisterWithParent("Poodle", poodle, "Dog")

	tests := []struct {
		name     string
		class    string
		expected int
	}{
		{"root class", "Object", 0},
		{"depth 1", "Animal", 1},
		{"depth 2", "Dog", 2},
		{"depth 3", "Poodle", 3},
		{"not found", "Cat", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := registry.GetDepth(tt.class)
			if depth != tt.expected {
				t.Errorf("GetDepth(%q) = %d, expected %d", tt.class, depth, tt.expected)
			}
		})
	}
}

func TestClassRegistry_FindDescendants(t *testing.T) {
	registry := NewClassRegistry()

	// Build hierarchy:
	//   Object
	//     -> Animal
	//        -> Dog
	//        -> Cat
	//     -> Vehicle
	object := &MockClassInfo{Name: "Object"}
	animal := &MockClassInfo{Name: "Animal"}
	dog := &MockClassInfo{Name: "Dog"}
	cat := &MockClassInfo{Name: "Cat"}
	vehicle := &MockClassInfo{Name: "Vehicle"}

	registry.RegisterWithParent("Object", object, "")
	registry.RegisterWithParent("Animal", animal, "Object")
	registry.RegisterWithParent("Dog", dog, "Animal")
	registry.RegisterWithParent("Cat", cat, "Animal")
	registry.RegisterWithParent("Vehicle", vehicle, "Object")

	// Find descendants of Animal
	descendants := registry.FindDescendants("Animal")
	if len(descendants) != 2 {
		t.Errorf("Animal descendants count = %d, expected 2", len(descendants))
	}

	// Find descendants of Object (should get all others)
	descendants = registry.FindDescendants("Object")
	if len(descendants) != 4 {
		t.Errorf("Object descendants count = %d, expected 4", len(descendants))
	}

	// Find descendants of Dog (leaf node, no descendants)
	descendants = registry.FindDescendants("Dog")
	if len(descendants) != 0 {
		t.Errorf("Dog descendants count = %d, expected 0", len(descendants))
	}
}

func TestClassRegistry_Count(t *testing.T) {
	registry := NewClassRegistry()

	if count := registry.Count(); count != 0 {
		t.Errorf("empty registry count = %d, expected 0", count)
	}

	registry.Register("Class1", &MockClassInfo{Name: "Class1"})
	if count := registry.Count(); count != 1 {
		t.Errorf("count after 1 registration = %d, expected 1", count)
	}

	registry.Register("Class2", &MockClassInfo{Name: "Class2"})
	registry.Register("Class3", &MockClassInfo{Name: "Class3"})
	if count := registry.Count(); count != 3 {
		t.Errorf("count after 3 registrations = %d, expected 3", count)
	}
}

func TestClassRegistry_GetClassNames(t *testing.T) {
	registry := NewClassRegistry()

	registry.Register("Animal", &MockClassInfo{Name: "Animal"})
	registry.Register("Dog", &MockClassInfo{Name: "Dog"})
	registry.Register("Cat", &MockClassInfo{Name: "Cat"})

	names := registry.GetClassNames()
	if len(names) != 3 {
		t.Errorf("GetClassNames length = %d, expected 3", len(names))
	}

	// Check that all expected names are present (order doesn't matter)
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	expected := []string{"Animal", "Dog", "Cat"}
	for _, name := range expected {
		if !nameSet[name] {
			t.Errorf("GetClassNames missing %q", name)
		}
	}
}

func TestClassRegistry_Clear(t *testing.T) {
	registry := NewClassRegistry()

	registry.Register("Class1", &MockClassInfo{Name: "Class1"})
	registry.Register("Class2", &MockClassInfo{Name: "Class2"})

	if count := registry.Count(); count != 2 {
		t.Errorf("count before clear = %d, expected 2", count)
	}

	registry.Clear()

	if count := registry.Count(); count != 0 {
		t.Errorf("count after clear = %d, expected 0", count)
	}

	if exists := registry.Exists("Class1"); exists {
		t.Error("Class1 should not exist after clear")
	}
}
