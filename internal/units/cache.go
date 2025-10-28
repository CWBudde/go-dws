package units

import (
	"os"
	"sync"
	"time"
)

// UnitCache caches parsed and analyzed units to speed up repeated runs.
// The cache is invalidated when the source file is modified.
//
// Task 9.140: Unit compilation cache
type UnitCache struct {
	// entries maps unit names (normalized) to cache entries
	entries map[string]*CacheEntry

	// mutex protects concurrent access to the cache
	mutex sync.RWMutex
}

// CacheEntry represents a cached unit with its metadata
type CacheEntry struct {
	// Unit is the cached unit instance
	Unit *Unit

	// FilePath is the source file path
	FilePath string

	// ModTime is the modification time when the unit was cached
	ModTime time.Time

	// LoadTime is when the unit was loaded into cache
	LoadTime time.Time
}

// NewUnitCache creates a new empty unit cache
func NewUnitCache() *UnitCache {
	return &UnitCache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get retrieves a unit from the cache if it exists and is still valid.
// Returns the unit and true if found and valid, nil and false otherwise.
//
// A cached unit is considered invalid if:
//   - The source file has been modified since caching
//   - The source file no longer exists
func (c *UnitCache) Get(name string) (*Unit, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[name]
	if !exists {
		return nil, false
	}

	// Check if the source file still exists
	fileInfo, err := os.Stat(entry.FilePath)
	if err != nil {
		// File no longer exists or is inaccessible - invalidate cache entry
		return nil, false
	}

	// Check if the file has been modified since we cached it
	if !fileInfo.ModTime().Equal(entry.ModTime) {
		// File has been modified - cache is stale
		return nil, false
	}

	// Cache entry is valid
	return entry.Unit, true
}

// Put adds a unit to the cache with its file modification time.
//
// If the file cannot be stat'd, the unit is still cached but will be
// invalidated on the next Get() call.
func (c *UnitCache) Put(name string, unit *Unit, filePath string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get the file modification time
	modTime := time.Now() // Default to current time if stat fails
	if fileInfo, err := os.Stat(filePath); err == nil {
		modTime = fileInfo.ModTime()
	}

	c.entries[name] = &CacheEntry{
		Unit:     unit,
		FilePath: filePath,
		ModTime:  modTime,
		LoadTime: time.Now(),
	}
}

// Invalidate removes a unit from the cache by name
func (c *UnitCache) Invalidate(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, name)
}

// Clear removes all entries from the cache
func (c *UnitCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Size returns the number of entries in the cache
func (c *UnitCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.entries)
}

// Stats returns cache statistics
type CacheStats struct {
	// TotalEntries is the number of units in the cache
	TotalEntries int

	// OldestEntry is the age of the oldest cached entry
	OldestEntry time.Duration

	// NewestEntry is the age of the newest cached entry
	NewestEntry time.Duration
}

// GetStats returns statistics about the cache
func (c *UnitCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := CacheStats{
		TotalEntries: len(c.entries),
	}

	if len(c.entries) == 0 {
		return stats
	}

	now := time.Now()
	var oldest, newest time.Time

	for _, entry := range c.entries {
		if oldest.IsZero() || entry.LoadTime.Before(oldest) {
			oldest = entry.LoadTime
		}
		if newest.IsZero() || entry.LoadTime.After(newest) {
			newest = entry.LoadTime
		}
	}

	if !oldest.IsZero() {
		stats.OldestEntry = now.Sub(oldest)
	}
	if !newest.IsZero() {
		stats.NewestEntry = now.Sub(newest)
	}

	return stats
}
