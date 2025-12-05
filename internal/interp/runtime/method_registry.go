package runtime

import (
	"fmt"
	"sync"
)

// MethodID is a unique identifier for a method in the registry.
// It replaces *ast.FunctionDecl references in runtime types.
type MethodID int

// InvalidMethodID represents an uninitialized or invalid method ID.
const InvalidMethodID MethodID = 0

// MethodRegistry stores methods by unique ID, enabling ID-based method calls
// without AST dependencies in the hot path.
//
// Design rationale:
//   - Methods are registered once during class/function declaration (compile-time)
//   - Method calls use IDs to look up MethodMetadata (avoiding AST access at runtime)
//   - Thread-safe for concurrent registration and lookup
//   - IDs are stable within a single interpreter session
//   - Memory-efficient: one copy of each method, referenced by ID
type MethodRegistry struct {
	// methods maps method IDs to their metadata.
	methods map[MethodID]*MethodMetadata
	// nameIndex maps normalized method names to their IDs for lookup by name.
	// Used for debugging and introspection.
	nameIndex map[string][]MethodID

	// nextID is the next method ID to assign (incremented on each registration).
	nextID MethodID

	// mu protects concurrent access to the registry.
	mu sync.RWMutex
}

// NewMethodRegistry creates a new method registry.
func NewMethodRegistry() *MethodRegistry {
	return &MethodRegistry{
		nextID:    1, // Start at 1 (0 is InvalidMethodID)
		methods:   make(map[MethodID]*MethodMetadata),
		nameIndex: make(map[string][]MethodID),
	}
}

// RegisterMethod adds a method to the registry and returns its unique ID.
// The method's ID field will be set to the assigned ID.
//
// This should be called during declaration processing (compile-time),
// not during method execution (runtime).
//
// Returns InvalidMethodID if metadata is nil.
func (r *MethodRegistry) RegisterMethod(metadata *MethodMetadata) MethodID {
	if metadata == nil {
		return InvalidMethodID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Assign ID to the method
	id := r.nextID
	r.nextID++

	// Set the ID field on the metadata
	metadata.ID = id

	// Store in registry
	r.methods[id] = metadata

	// Add to name index for debugging/introspection
	normalizedName := normalizeIdentifier(metadata.Name)
	r.nameIndex[normalizedName] = append(r.nameIndex[normalizedName], id)

	return id
}

// GetMethod retrieves a method by ID.
// Returns nil if the ID is invalid or not found.
//
// This is the primary hot-path operation used during method calls.
// It's thread-safe and optimized for concurrent reads.
func (r *MethodRegistry) GetMethod(id MethodID) *MethodMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.methods[id]
}

// GetMethodsByName retrieves all methods with the given name (case-insensitive).
// Used for introspection and debugging.
//
// Returns an empty slice if no methods with that name exist.
func (r *MethodRegistry) GetMethodsByName(name string) []*MethodMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedName := normalizeIdentifier(name)
	ids := r.nameIndex[normalizedName]

	if len(ids) == 0 {
		return nil
	}

	methods := make([]*MethodMetadata, 0, len(ids))
	for _, id := range ids {
		if method := r.methods[id]; method != nil {
			methods = append(methods, method)
		}
	}

	return methods
}

// HasMethod checks if a method ID exists in the registry.
func (r *MethodRegistry) HasMethod(id MethodID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.methods[id]
	return exists
}

// Count returns the number of methods in the registry.
func (r *MethodRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.methods)
}

// Clear removes all methods from the registry.
// This is primarily useful for testing.
func (r *MethodRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID = 1
	r.methods = make(map[MethodID]*MethodMetadata)
	r.nameIndex = make(map[string][]MethodID)
}

// Stats returns statistics about the registry for debugging.
func (r *MethodRegistry) Stats() MethodRegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return MethodRegistryStats{
		MethodCount:      len(r.methods),
		NextID:           r.nextID,
		UniqueNameCount:  len(r.nameIndex),
		MaxOverloadCount: r.maxOverloadCount(),
	}
}

// maxOverloadCount returns the maximum number of overloads for any method name.
// Must be called with lock held.
func (r *MethodRegistry) maxOverloadCount() int {
	max := 0
	for _, ids := range r.nameIndex {
		if len(ids) > max {
			max = len(ids)
		}
	}
	return max
}

// MethodRegistryStats contains statistics about the method registry.
type MethodRegistryStats struct {
	// MethodCount is the total number of registered methods.
	MethodCount int

	// NextID is the next method ID that will be assigned.
	NextID MethodID

	// UniqueNameCount is the number of unique method names.
	UniqueNameCount int

	// MaxOverloadCount is the maximum number of overloads for any method name.
	MaxOverloadCount int
}

// String returns a human-readable representation of the stats.
func (s MethodRegistryStats) String() string {
	return fmt.Sprintf("MethodRegistry: %d methods, %d unique names, max %d overloads, next ID: %d",
		s.MethodCount, s.UniqueNameCount, s.MaxOverloadCount, s.NextID)
}
