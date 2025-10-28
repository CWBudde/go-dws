package units

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewUnitCache tests creating a new cache
func TestNewUnitCache(t *testing.T) {
	cache := NewUnitCache()

	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}

	if cache.Size() != 0 {
		t.Errorf("Expected empty cache, got size %d", cache.Size())
	}
}

// TestCachePutAndGet tests basic put/get operations
func TestCachePutAndGet(t *testing.T) {
	cache := NewUnitCache()

	// Create a temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.dws")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test unit
	unit := NewUnit("TestUnit", filePath)

	// Put it in the cache
	cache.Put("testunit", unit, filePath)

	// Verify size
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Get it back
	retrieved, found := cache.Get("testunit")
	if !found {
		t.Fatal("Expected to find unit in cache")
	}

	if retrieved != unit {
		t.Error("Expected to retrieve the same unit instance")
	}
}

// TestCacheInvalidation tests cache invalidation
func TestCacheInvalidation(t *testing.T) {
	cache := NewUnitCache()

	// Create a temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.dws")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	unit := NewUnit("TestUnit", filePath)

	cache.Put("testunit", unit, filePath)

	// Verify it's there
	if _, found := cache.Get("testunit"); !found {
		t.Fatal("Expected unit to be in cache")
	}

	// Invalidate it
	cache.Invalidate("testunit")

	// Verify it's gone
	if _, found := cache.Get("testunit"); found {
		t.Error("Expected unit to be invalidated")
	}

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after invalidation, got %d", cache.Size())
	}
}

// TestCacheClear tests clearing the cache
func TestCacheClear(t *testing.T) {
	cache := NewUnitCache()

	// Add multiple units
	for i := 0; i < 5; i++ {
		unit := NewUnit("Unit"+string(rune('A'+i)), "/tmp/test.dws")
		cache.Put("unit"+string(rune('a'+i)), unit, "/tmp/test.dws")
	}

	if cache.Size() != 5 {
		t.Errorf("Expected cache size 5, got %d", cache.Size())
	}

	// Clear the cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}

// TestCacheFileModification tests cache invalidation on file modification
func TestCacheFileModification(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.dws")

	// Write initial content
	initialContent := []byte("unit Test; interface implementation end.")
	if err := os.WriteFile(filePath, initialContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewUnitCache()
	unit := NewUnit("Test", filePath)

	// Cache the unit
	cache.Put("test", unit, filePath)

	// Verify it's in cache
	if _, found := cache.Get("test"); !found {
		t.Fatal("Expected unit to be in cache")
	}

	// Wait a bit to ensure modification time changes
	time.Sleep(10 * time.Millisecond)

	// Modify the file
	modifiedContent := []byte("unit Test; interface function Foo: Integer; implementation function Foo: Integer; begin Result := 42; end; end.")
	if err := os.WriteFile(filePath, modifiedContent, 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Try to get from cache - should return false because file was modified
	if _, found := cache.Get("test"); found {
		t.Error("Expected cache to be invalidated after file modification")
	}
}

// TestCacheFileDeletion tests cache invalidation when file is deleted
func TestCacheFileDeletion(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.dws")

	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewUnitCache()
	unit := NewUnit("Test", filePath)

	// Cache the unit
	cache.Put("test", unit, filePath)

	// Verify it's in cache
	if _, found := cache.Get("test"); !found {
		t.Fatal("Expected unit to be in cache")
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Failed to remove test file: %v", err)
	}

	// Try to get from cache - should return false because file no longer exists
	if _, found := cache.Get("test"); found {
		t.Error("Expected cache to be invalidated after file deletion")
	}
}

// TestCacheStats tests cache statistics
func TestCacheStats(t *testing.T) {
	cache := NewUnitCache()

	// Empty cache stats
	stats := cache.GetStats()
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries in empty cache, got %d", stats.TotalEntries)
	}

	// Add some units
	unit1 := NewUnit("Unit1", "/tmp/unit1.dws")
	cache.Put("unit1", unit1, "/tmp/unit1.dws")

	time.Sleep(5 * time.Millisecond)

	unit2 := NewUnit("Unit2", "/tmp/unit2.dws")
	cache.Put("unit2", unit2, "/tmp/unit2.dws")

	stats = cache.GetStats()
	if stats.TotalEntries != 2 {
		t.Errorf("Expected 2 entries, got %d", stats.TotalEntries)
	}

	// Oldest should be older than newest
	if stats.OldestEntry < stats.NewestEntry {
		t.Error("Expected oldest entry to have longer age than newest entry")
	}
}

// TestCacheConcurrency tests concurrent access to the cache
func TestCacheConcurrency(t *testing.T) {
	cache := NewUnitCache()

	// Create test file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.dws")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run concurrent operations
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			unit := NewUnit("Test", filePath)
			cache.Put("test", unit, filePath)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = cache.Get("test")
		}
		done <- true
	}()

	// Invalidator goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Invalidate("test")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// If we get here without deadlock or panic, the test passes
}

// TestRegistryWithCache tests that the registry uses the cache
func TestRegistryWithCache(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test unit file
	unitContent := `unit TestCache;

interface

function GetValue: Integer;

implementation

function GetValue: Integer;
begin
  Result := 42;
end;

end.`

	unitPath := filepath.Join(tempDir, "TestCache.dws")
	if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
		t.Fatalf("Failed to create test unit file: %v", err)
	}

	registry := NewUnitRegistry([]string{tempDir})

	// Load the unit for the first time
	unit1, err := registry.LoadUnit("TestCache", nil)
	if err != nil {
		t.Fatalf("Failed to load unit: %v", err)
	}

	cacheSize := registry.GetCache().Size()
	if cacheSize != 1 {
		t.Errorf("Expected cache size 1 after first load, got %d", cacheSize)
	}

	// Unregister the unit from the registry (but leave it in cache)
	registry.UnregisterUnit("TestCache")

	// Load again - should come from cache
	unit2, err := registry.LoadUnit("TestCache", nil)
	if err != nil {
		t.Fatalf("Failed to reload unit: %v", err)
	}

	// Should be the same instance from cache
	if unit1 != unit2 {
		t.Error("Expected to get same unit instance from cache")
	}
}

// TestCacheInvalidationViaRegistry tests invalidating cache through registry
func TestCacheInvalidationViaRegistry(t *testing.T) {
	tempDir := t.TempDir()

	unitContent := `unit TestInvalidate;
interface
implementation
end.`

	unitPath := filepath.Join(tempDir, "TestInvalidate.dws")
	if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
		t.Fatalf("Failed to create test unit file: %v", err)
	}

	registry := NewUnitRegistry([]string{tempDir})

	// Load the unit
	_, err := registry.LoadUnit("TestInvalidate", nil)
	if err != nil {
		t.Fatalf("Failed to load unit: %v", err)
	}

	// Check cache size
	if registry.GetCache().Size() != 1 {
		t.Error("Expected 1 entry in cache")
	}

	// Invalidate via registry
	registry.InvalidateCache("TestInvalidate")

	// Cache should be empty
	if registry.GetCache().Size() != 0 {
		t.Error("Expected cache to be empty after invalidation")
	}
}

// TestClearCacheViaRegistry tests clearing cache through registry
func TestClearCacheViaRegistry(t *testing.T) {
	tempDir := t.TempDir()

	// Create two test units
	for _, name := range []string{"Unit1", "Unit2"} {
		content := "unit " + name + "; interface implementation end."
		path := filepath.Join(tempDir, name+".dws")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	registry := NewUnitRegistry([]string{tempDir})

	// Load both units
	_, _ = registry.LoadUnit("Unit1", nil)
	_, _ = registry.LoadUnit("Unit2", nil)

	if registry.GetCache().Size() != 2 {
		t.Errorf("Expected 2 entries in cache, got %d", registry.GetCache().Size())
	}

	// Clear cache
	registry.ClearCache()

	if registry.GetCache().Size() != 0 {
		t.Errorf("Expected empty cache after clear, got %d", registry.GetCache().Size())
	}
}
