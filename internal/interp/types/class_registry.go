// Package types provides the type system for the DWScript interpreter.
// This file implements ClassRegistry for managing class types and hierarchies.
//
// Task 3.4.2: Create ClassRegistry abstraction
package types

import (
	"strings"
)

// ClassRegistry manages class type information and supports efficient
// hierarchy queries for inheritance relationships.
//
// The registry provides:
// - Case-insensitive class name lookup
// - Parent class tracking for inheritance
// - Efficient hierarchy traversal
// - Descendant checking (is class A derived from class B?)
type ClassRegistry struct {
	// classes maps lowercase class names to ClassInfo
	classes map[string]*ClassInfoEntry
}

// ClassInfoEntry wraps class information stored in the registry.
// We use 'any' to avoid circular dependencies with the interp package.
type ClassInfoEntry struct {
	// Info is the actual ClassInfo from interp package (stored as any)
	Info any

	// Name is the original case-sensitive class name
	Name string

	// ParentName is the parent class name (case-sensitive)
	// Empty string means no parent (root of hierarchy)
	ParentName string
}

// NewClassRegistry creates a new empty class registry.
func NewClassRegistry() *ClassRegistry {
	return &ClassRegistry{
		classes: make(map[string]*ClassInfoEntry),
	}
}

// Register adds a class to the registry.
// The name is stored case-insensitively (lowercase key).
// If a class with the same name already exists, it is replaced.
func (r *ClassRegistry) Register(name string, classInfo any) {
	if classInfo == nil {
		return
	}

	// Extract parent name if available
	// We need to handle the ClassInfo type without importing interp
	parentName := ""
	// Note: Parent extraction will be handled by the caller since we can't
	// access ClassInfo.Parent without circular dependency

	entry := &ClassInfoEntry{
		Info:       classInfo,
		Name:       name,
		ParentName: parentName,
	}

	r.classes[strings.ToLower(name)] = entry
}

// RegisterWithParent adds a class to the registry with explicit parent name.
// This allows the registry to track inheritance without accessing ClassInfo internals.
func (r *ClassRegistry) RegisterWithParent(name string, classInfo any, parentName string) {
	if classInfo == nil {
		return
	}

	entry := &ClassInfoEntry{
		Info:       classInfo,
		Name:       name,
		ParentName: parentName,
	}

	r.classes[strings.ToLower(name)] = entry
}

// Lookup finds a class by name (case-insensitive).
// Returns the class info and true if found, nil and false otherwise.
func (r *ClassRegistry) Lookup(name string) (any, bool) {
	entry, ok := r.classes[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	return entry.Info, true
}

// Exists checks if a class with the given name exists in the registry.
// The check is case-insensitive.
func (r *ClassRegistry) Exists(name string) bool {
	_, exists := r.classes[strings.ToLower(name)]
	return exists
}

// LookupHierarchy returns all classes in the inheritance hierarchy for the given class.
// The result is ordered from most specific (the class itself) to least specific (root ancestor).
//
// Example: If Dog inherits from Animal, which inherits from Object:
//
//	LookupHierarchy("Dog") returns [Dog, Animal, Object]
//
// Returns nil if the class is not found.
func (r *ClassRegistry) LookupHierarchy(name string) []any {
	entry, ok := r.classes[strings.ToLower(name)]
	if !ok {
		return nil
	}

	hierarchy := []any{entry.Info}

	// Walk up the parent chain
	currentParent := entry.ParentName
	for currentParent != "" {
		parentEntry, ok := r.classes[strings.ToLower(currentParent)]
		if !ok {
			// Parent not found in registry - stop here
			break
		}
		hierarchy = append(hierarchy, parentEntry.Info)
		currentParent = parentEntry.ParentName
	}

	return hierarchy
}

// GetParentName returns the parent class name for the given class.
// Returns empty string if the class has no parent or is not found.
func (r *ClassRegistry) GetParentName(name string) string {
	entry, ok := r.classes[strings.ToLower(name)]
	if !ok {
		return ""
	}
	return entry.ParentName
}

// IsDescendantOf checks if descendantName is a descendant of ancestorName.
// Returns true if descendantName inherits from ancestorName (directly or indirectly).
// Also returns true if descendantName equals ancestorName (a class is its own descendant).
//
// Example:
//
//	IsDescendantOf("Dog", "Animal") returns true if Dog inherits from Animal
//	IsDescendantOf("Dog", "Dog") returns true (class is its own descendant)
func (r *ClassRegistry) IsDescendantOf(descendantName, ancestorName string) bool {
	// Normalize names for comparison
	descendantKey := strings.ToLower(descendantName)
	ancestorKey := strings.ToLower(ancestorName)

	// A class is its own descendant
	if descendantKey == ancestorKey {
		return true
	}

	// Look up the descendant class
	entry, ok := r.classes[descendantKey]
	if !ok {
		return false
	}

	// Walk up the parent chain looking for the ancestor
	currentParent := entry.ParentName
	for currentParent != "" {
		if strings.ToLower(currentParent) == ancestorKey {
			return true
		}

		parentEntry, ok := r.classes[strings.ToLower(currentParent)]
		if !ok {
			// Parent not in registry - can't continue
			break
		}
		currentParent = parentEntry.ParentName
	}

	return false
}

// GetAllClasses returns a map of all registered classes.
// The map uses lowercase keys (case-insensitive).
// The returned map should not be modified directly.
func (r *ClassRegistry) GetAllClasses() map[string]any {
	result := make(map[string]any, len(r.classes))
	for key, entry := range r.classes {
		result[key] = entry.Info
	}
	return result
}

// Count returns the number of classes in the registry.
func (r *ClassRegistry) Count() int {
	return len(r.classes)
}

// GetClassNames returns a slice of all registered class names (original case).
// The names are not sorted.
func (r *ClassRegistry) GetClassNames() []string {
	names := make([]string, 0, len(r.classes))
	for _, entry := range r.classes {
		names = append(names, entry.Name)
	}
	return names
}

// FindDescendants returns all classes that inherit from the given ancestor class.
// The result includes both direct children and all descendants in the hierarchy.
// The ancestor class itself is not included in the result.
//
// Example: If Cat and Dog inherit from Animal:
//
//	FindDescendants("Animal") returns [Cat, Dog]
func (r *ClassRegistry) FindDescendants(ancestorName string) []any {
	ancestorKey := strings.ToLower(ancestorName)
	descendants := []any{}

	for _, entry := range r.classes {
		// Skip the ancestor itself
		if strings.ToLower(entry.Name) == ancestorKey {
			continue
		}

		// Check if this class inherits from the ancestor
		if r.IsDescendantOf(entry.Name, ancestorName) {
			descendants = append(descendants, entry.Info)
		}
	}

	return descendants
}

// GetDepth returns the depth of a class in the inheritance hierarchy.
// Depth 0 means the class has no parent (root class).
// Depth 1 means the class has a parent but the parent has no parent.
// Returns -1 if the class is not found.
//
// Example:
//
//	GetDepth("Object") returns 0 (assuming Object is root)
//	GetDepth("Animal") returns 1 (if Animal inherits from Object)
//	GetDepth("Dog") returns 2 (if Dog inherits from Animal)
func (r *ClassRegistry) GetDepth(name string) int {
	entry, ok := r.classes[strings.ToLower(name)]
	if !ok {
		return -1
	}

	depth := 0
	currentParent := entry.ParentName

	for currentParent != "" {
		parentEntry, ok := r.classes[strings.ToLower(currentParent)]
		if !ok {
			// Parent not found - stop here
			break
		}
		depth++
		currentParent = parentEntry.ParentName
	}

	return depth
}

// Clear removes all classes from the registry.
func (r *ClassRegistry) Clear() {
	r.classes = make(map[string]*ClassInfoEntry)
}
