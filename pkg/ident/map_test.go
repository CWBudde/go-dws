package ident

import (
	"sort"
	"testing"
)

func TestNewMap(t *testing.T) {
	m := NewMap[int]()
	if m == nil {
		t.Fatal("NewMap returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("NewMap().Len() = %d, want 0", m.Len())
	}
}

func TestNewMapWithCapacity(t *testing.T) {
	m := NewMapWithCapacity[string](100)
	if m == nil {
		t.Fatal("NewMapWithCapacity returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("NewMapWithCapacity().Len() = %d, want 0", m.Len())
	}
}

func TestMapSetAndGet(t *testing.T) {
	m := NewMap[int]()

	// Set with original casing
	m.Set("MyVariable", 42)

	// Get with same casing
	if val, ok := m.Get("MyVariable"); !ok || val != 42 {
		t.Errorf("Get(MyVariable) = %d, %v, want 42, true", val, ok)
	}

	// Get with lowercase
	if val, ok := m.Get("myvariable"); !ok || val != 42 {
		t.Errorf("Get(myvariable) = %d, %v, want 42, true", val, ok)
	}

	// Get with uppercase
	if val, ok := m.Get("MYVARIABLE"); !ok || val != 42 {
		t.Errorf("Get(MYVARIABLE) = %d, %v, want 42, true", val, ok)
	}

	// Get non-existent key
	if val, ok := m.Get("nonexistent"); ok || val != 0 {
		t.Errorf("Get(nonexistent) = %d, %v, want 0, false", val, ok)
	}
}

func TestMapSetOverwrite(t *testing.T) {
	m := NewMap[int]()

	m.Set("MyVar", 10)
	m.Set("myvar", 20) // Should overwrite

	if val, ok := m.Get("MyVar"); !ok || val != 20 {
		t.Errorf("Get(MyVar) after overwrite = %d, %v, want 20, true", val, ok)
	}

	// Original key should now be "myvar"
	if orig := m.GetOriginalKey("MyVar"); orig != "myvar" {
		t.Errorf("GetOriginalKey(MyVar) = %q, want %q", orig, "myvar")
	}
}

func TestMapSetIfAbsent(t *testing.T) {
	m := NewMap[int]()

	// First set should succeed
	if !m.SetIfAbsent("MyVar", 42) {
		t.Error("SetIfAbsent should return true for new key")
	}

	// Second set with same key (different case) should fail
	if m.SetIfAbsent("myvar", 100) {
		t.Error("SetIfAbsent should return false for existing key")
	}

	// Value should not have changed
	if val, _ := m.Get("MyVar"); val != 42 {
		t.Errorf("Value changed after SetIfAbsent returned false: got %d, want 42", val)
	}

	// Original key should still be "MyVar"
	if orig := m.GetOriginalKey("myvar"); orig != "MyVar" {
		t.Errorf("Original key changed: got %q, want %q", orig, "MyVar")
	}
}

func TestMapGetOriginalKey(t *testing.T) {
	m := NewMap[int]()

	m.Set("MyVariable", 42)
	m.Set("COUNTER", 10)

	tests := []struct {
		lookup   string
		expected string
	}{
		{"MyVariable", "MyVariable"},
		{"myvariable", "MyVariable"},
		{"MYVARIABLE", "MyVariable"},
		{"counter", "COUNTER"},
		{"Counter", "COUNTER"},
		{"COUNTER", "COUNTER"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.lookup, func(t *testing.T) {
			if orig := m.GetOriginalKey(tt.lookup); orig != tt.expected {
				t.Errorf("GetOriginalKey(%q) = %q, want %q", tt.lookup, orig, tt.expected)
			}
		})
	}
}

func TestMapHas(t *testing.T) {
	m := NewMap[int]()
	m.Set("MyVar", 42)

	tests := []struct {
		key      string
		expected bool
	}{
		{"MyVar", true},
		{"myvar", true},
		{"MYVAR", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := m.Has(tt.key); got != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestMapDelete(t *testing.T) {
	m := NewMap[int]()
	m.Set("MyVar", 42)
	m.Set("Counter", 10)

	// Delete existing key with different casing
	if !m.Delete("myvar") {
		t.Error("Delete(myvar) should return true")
	}

	// Verify deleted
	if m.Has("MyVar") {
		t.Error("MyVar should not exist after delete")
	}
	if m.GetOriginalKey("MyVar") != "" {
		t.Error("GetOriginalKey should return empty after delete")
	}

	// Counter should still exist
	if !m.Has("Counter") {
		t.Error("Counter should still exist")
	}

	// Delete non-existent key
	if m.Delete("nonexistent") {
		t.Error("Delete(nonexistent) should return false")
	}
}

func TestMapLen(t *testing.T) {
	m := NewMap[int]()

	if m.Len() != 0 {
		t.Errorf("Empty map Len() = %d, want 0", m.Len())
	}

	m.Set("A", 1)
	if m.Len() != 1 {
		t.Errorf("After 1 Set, Len() = %d, want 1", m.Len())
	}

	m.Set("B", 2)
	if m.Len() != 2 {
		t.Errorf("After 2 Sets, Len() = %d, want 2", m.Len())
	}

	// Setting same key with different case shouldn't increase len
	m.Set("a", 10)
	if m.Len() != 2 {
		t.Errorf("After overwrite, Len() = %d, want 2", m.Len())
	}

	m.Delete("A")
	if m.Len() != 1 {
		t.Errorf("After delete, Len() = %d, want 1", m.Len())
	}
}

func TestMapKeys(t *testing.T) {
	m := NewMap[int]()
	m.Set("MyVar", 1)
	m.Set("Counter", 2)
	m.Set("VALUE", 3)

	keys := m.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() len = %d, want 3", len(keys))
	}

	// Sort for deterministic comparison
	sort.Strings(keys)
	expected := []string{"Counter", "MyVar", "VALUE"}
	sort.Strings(expected)

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, key, expected[i])
		}
	}
}

func TestMapRange(t *testing.T) {
	m := NewMap[int]()
	m.Set("A", 1)
	m.Set("B", 2)
	m.Set("C", 3)

	// Collect all entries
	entries := make(map[string]int)
	m.Range(func(key string, value int) bool {
		entries[key] = value
		return true
	})

	if len(entries) != 3 {
		t.Errorf("Range visited %d entries, want 3", len(entries))
	}

	// Verify entries have original keys
	if entries["A"] != 1 || entries["B"] != 2 || entries["C"] != 3 {
		t.Errorf("Range entries incorrect: %v", entries)
	}
}

func TestMapRangeEarlyStop(t *testing.T) {
	m := NewMap[int]()
	m.Set("A", 1)
	m.Set("B", 2)
	m.Set("C", 3)

	count := 0
	m.Range(func(key string, value int) bool {
		count++
		return count < 2 // Stop after 2 iterations
	})

	if count != 2 {
		t.Errorf("Range with early stop visited %d entries, want 2", count)
	}
}

func TestMapClear(t *testing.T) {
	m := NewMap[int]()
	m.Set("A", 1)
	m.Set("B", 2)

	m.Clear()

	if m.Len() != 0 {
		t.Errorf("After Clear(), Len() = %d, want 0", m.Len())
	}

	if m.Has("A") {
		t.Error("After Clear(), Has(A) should be false")
	}

	// Should still be usable after clear
	m.Set("C", 3)
	if val, ok := m.Get("C"); !ok || val != 3 {
		t.Errorf("After Clear() and Set(), Get(C) = %d, %v, want 3, true", val, ok)
	}
}

func TestMapClone(t *testing.T) {
	m := NewMap[int]()
	m.Set("A", 1)
	m.Set("B", 2)

	clone := m.Clone()

	// Clone should have same content
	if clone.Len() != 2 {
		t.Errorf("Clone Len() = %d, want 2", clone.Len())
	}
	if val, _ := clone.Get("A"); val != 1 {
		t.Errorf("Clone Get(A) = %d, want 1", val)
	}
	if orig := clone.GetOriginalKey("a"); orig != "A" {
		t.Errorf("Clone GetOriginalKey(a) = %q, want %q", orig, "A")
	}

	// Modifying clone shouldn't affect original
	clone.Set("A", 100)
	clone.Delete("B")

	if val, _ := m.Get("A"); val != 1 {
		t.Errorf("Original affected by clone modification: Get(A) = %d, want 1", val)
	}
	if !m.Has("B") {
		t.Error("Original affected by clone deletion: B should still exist")
	}
}

func TestMapWithPointerValues(t *testing.T) {
	type Symbol struct {
		Name  string
		Value int
	}

	m := NewMap[*Symbol]()
	sym := &Symbol{Name: "MyVar", Value: 42}
	m.Set("MyVar", sym)

	// Get should return same pointer
	if got, ok := m.Get("myvar"); !ok || got != sym {
		t.Errorf("Get(myvar) returned different pointer or not found")
	}

	// Clone shares references (shallow copy)
	clone := m.Clone()
	if got, _ := clone.Get("MyVar"); got != sym {
		t.Error("Clone should share same pointer references")
	}
}

func TestMapEmptyKey(t *testing.T) {
	m := NewMap[int]()

	// Empty string as key should work
	m.Set("", 42)
	if val, ok := m.Get(""); !ok || val != 42 {
		t.Errorf("Get('') = %d, %v, want 42, true", val, ok)
	}
	if orig := m.GetOriginalKey(""); orig != "" {
		t.Errorf("GetOriginalKey('') = %q, want empty string", orig)
	}
}

// Benchmarks

func BenchmarkMapSet(b *testing.B) {
	m := NewMap[int]()
	keys := []string{"MyVariable", "Counter", "RESULT", "tempValue"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(keys[i%len(keys)], i)
	}
}

func BenchmarkMapGet(b *testing.B) {
	m := NewMap[int]()
	m.Set("MyVariable", 42)
	m.Set("Counter", 10)
	m.Set("Result", 100)

	lookups := []string{"myvariable", "COUNTER", "result", "MyVariable"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get(lookups[i%len(lookups)])
	}
}

func BenchmarkMapGetOriginalKey(b *testing.B) {
	m := NewMap[int]()
	m.Set("MyVariable", 42)

	lookups := []string{"myvariable", "MYVARIABLE", "MyVariable"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.GetOriginalKey(lookups[i%len(lookups)])
	}
}
