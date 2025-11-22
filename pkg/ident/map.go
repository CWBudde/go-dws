package ident

// Map is a case-insensitive map that stores values with normalized keys
// while optionally preserving the original casing of keys for error messages
// and display purposes.
//
// This type is designed for DWScript's case-insensitive identifier handling,
// where "MyVariable", "myvariable", and "MYVARIABLE" should all refer to
// the same entry.
//
// Thread Safety: Map is NOT safe for concurrent use. If concurrent access
// is needed, the caller must provide synchronization (e.g., sync.RWMutex).
//
// Example:
//
//	m := ident.NewMap[int]()
//	m.Set("MyVariable", 42)
//	val, ok := m.Get("myvariable")  // val=42, ok=true
//	orig := m.GetOriginalKey("MYVARIABLE")  // "MyVariable"
type Map[T any] struct {
	store     map[string]T
	originals map[string]string // normalized -> original key
}

// NewMap creates a new case-insensitive map.
//
// Example:
//
//	symbols := ident.NewMap[*Symbol]()
//	symbols.Set("MyVar", sym)
func NewMap[T any]() *Map[T] {
	return &Map[T]{
		store:     make(map[string]T),
		originals: make(map[string]string),
	}
}

// NewMapWithCapacity creates a new case-insensitive map with pre-allocated capacity.
// Use this when you know approximately how many entries will be stored.
//
// Example:
//
//	// Pre-allocate for ~100 symbols
//	symbols := ident.NewMapWithCapacity[*Symbol](100)
func NewMapWithCapacity[T any](capacity int) *Map[T] {
	return &Map[T]{
		store:     make(map[string]T, capacity),
		originals: make(map[string]string, capacity),
	}
}

// Set stores a value with the given key. The key is normalized for storage,
// but the original casing is preserved for later retrieval via GetOriginalKey.
//
// If a key already exists (case-insensitive match), the value is updated
// and the original key casing is replaced with the new one.
//
// Example:
//
//	m.Set("MyVariable", 42)
//	m.Set("myvariable", 100)  // Updates value, original key becomes "myvariable"
func (m *Map[T]) Set(key string, value T) {
	normalized := Normalize(key)
	m.store[normalized] = value
	m.originals[normalized] = key
}

// SetIfAbsent stores a value only if the key doesn't already exist.
// Returns true if the value was set, false if the key already existed.
//
// This is useful for "define once" semantics where redefinition is an error.
//
// Example:
//
//	if !m.SetIfAbsent("MyVar", 42) {
//	    return fmt.Errorf("variable '%s' already defined", m.GetOriginalKey("MyVar"))
//	}
func (m *Map[T]) SetIfAbsent(key string, value T) bool {
	normalized := Normalize(key)
	if _, exists := m.store[normalized]; exists {
		return false
	}
	m.store[normalized] = value
	m.originals[normalized] = key
	return true
}

// Get retrieves the value for the given key (case-insensitive).
// Returns the value and true if found, or the zero value and false if not found.
//
// Example:
//
//	if val, ok := m.Get("myvariable"); ok {
//	    fmt.Println("Found:", val)
//	}
func (m *Map[T]) Get(key string) (T, bool) {
	val, ok := m.store[Normalize(key)]
	return val, ok
}

// GetOriginalKey returns the original casing of the key as it was stored.
// Returns an empty string if the key doesn't exist.
//
// Use this for error messages and display to show the user the originally
// defined casing rather than normalized or lookup casing.
//
// Example:
//
//	m.Set("MyVariable", 42)
//	orig := m.GetOriginalKey("MYVARIABLE")  // Returns "MyVariable"
func (m *Map[T]) GetOriginalKey(key string) string {
	return m.originals[Normalize(key)]
}

// Has returns true if the key exists in the map (case-insensitive).
//
// Example:
//
//	if m.Has("myvar") {
//	    // Variable exists
//	}
func (m *Map[T]) Has(key string) bool {
	_, ok := m.store[Normalize(key)]
	return ok
}

// Delete removes the entry for the given key (case-insensitive).
// Returns true if an entry was deleted, false if the key didn't exist.
//
// Example:
//
//	if m.Delete("MyVar") {
//	    fmt.Println("Deleted")
//	}
func (m *Map[T]) Delete(key string) bool {
	normalized := Normalize(key)
	if _, exists := m.store[normalized]; !exists {
		return false
	}
	delete(m.store, normalized)
	delete(m.originals, normalized)
	return true
}

// Len returns the number of entries in the map.
//
// Example:
//
//	fmt.Printf("Map has %d entries\n", m.Len())
func (m *Map[T]) Len() int {
	return len(m.store)
}

// Keys returns a slice of all original keys (with their original casing).
// The order of keys is not guaranteed.
//
// Example:
//
//	for _, key := range m.Keys() {
//	    fmt.Println(key)
//	}
func (m *Map[T]) Keys() []string {
	keys := make([]string, 0, len(m.originals))
	for _, original := range m.originals {
		keys = append(keys, original)
	}
	return keys
}

// Range iterates over all entries in the map, calling f for each key-value pair.
// The key passed to f is the original casing as stored.
// The iteration order is not guaranteed.
//
// If f returns false, iteration stops early.
//
// Example:
//
//	m.Range(func(key string, value int) bool {
//	    fmt.Printf("%s = %d\n", key, value)
//	    return true  // Continue iteration
//	})
func (m *Map[T]) Range(f func(key string, value T) bool) {
	for normalized, value := range m.store {
		original := m.originals[normalized]
		if !f(original, value) {
			return
		}
	}
}

// Clear removes all entries from the map.
//
// Example:
//
//	m.Clear()
//	fmt.Println(m.Len())  // 0
func (m *Map[T]) Clear() {
	m.store = make(map[string]T)
	m.originals = make(map[string]string)
}

// Clone creates a shallow copy of the map.
// The values themselves are not cloned (they share the same references).
//
// Example:
//
//	copy := m.Clone()
func (m *Map[T]) Clone() *Map[T] {
	clone := NewMapWithCapacity[T](len(m.store))
	for normalized, value := range m.store {
		clone.store[normalized] = value
		clone.originals[normalized] = m.originals[normalized]
	}
	return clone
}
