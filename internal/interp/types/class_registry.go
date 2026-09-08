// Package types provides the type system for the DWScript interpreter.
// This file implements ClassRegistry for managing class types and hierarchies.
package types

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ident"
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
	classes *ident.Map[runtime.IClassInfo]
	// parentNames records explicitly declared parent names, including forward references.
	parentNames *ident.Map[string]
}

// NewClassRegistry creates a new empty class registry.
func NewClassRegistry() *ClassRegistry {
	return &ClassRegistry{
		classes:     ident.NewMap[runtime.IClassInfo](),
		parentNames: ident.NewMap[string](),
	}
}

// Register adds a class to the registry.
// The name is stored case-insensitively (normalized key).
// If a class with the same name already exists, it is replaced.
func (r *ClassRegistry) Register(name string, classInfo runtime.IClassInfo) {
	if classInfo == nil {
		return
	}
	r.classes.Set(name, classInfo)
	r.parentNames.Delete(name)
}

// RegisterWithParent registers a class and its declared parent name.
func (r *ClassRegistry) RegisterWithParent(name string, classInfo runtime.IClassInfo, parentName string) {
	if classInfo == nil {
		return
	}
	r.classes.Set(name, classInfo)
	r.parentNames.Set(name, parentName)
}

// Lookup finds a class by name (case-insensitive).
// Returns the class info and true if found, nil and false otherwise.
func (r *ClassRegistry) Lookup(name string) (runtime.IClassInfo, bool) {
	entry, ok := r.classes.Get(name)
	if !ok {
		return nil, false
	}
	return entry, true
}

// Exists checks if a class with the given name exists in the registry.
// The check is case-insensitive.
func (r *ClassRegistry) Exists(name string) bool {
	return r.classes.Has(name)
}

// LookupHierarchy returns all classes in the inheritance hierarchy for the given class.
// The result is ordered from most specific (the class itself) to least specific (root ancestor).
//
// Example: If Dog inherits from Animal, which inherits from Object:
//
//	LookupHierarchy("Dog") returns [Dog, Animal, Object]
//
// Returns nil if the class is not found.
func (r *ClassRegistry) LookupHierarchy(name string) []runtime.IClassInfo {
	entry, ok := r.classes.Get(name)
	if !ok {
		return nil
	}

	hierarchy := []runtime.IClassInfo{entry}

	// Walk up the parent chain
	currentParent := r.GetParentName(name)
	for currentParent != "" {
		parentEntry, ok := r.classes.Get(currentParent)
		if !ok {
			// Parent not found in registry - stop here
			break
		}
		hierarchy = append(hierarchy, parentEntry)
		currentParent = r.GetParentName(currentParent)
	}

	return hierarchy
}

// GetParentName returns the parent class name for the given class.
// Returns empty string if the class has no parent or is not found.
func (r *ClassRegistry) GetParentName(name string) string {
	entry, ok := r.classes.Get(name)
	if !ok {
		return ""
	}
	if parent := entry.GetParent(); parent != nil {
		return parent.GetName()
	}
	parentName, _ := r.parentNames.Get(name)
	return parentName
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
	// A class is its own descendant (case-insensitive comparison)
	if ident.Equal(descendantName, ancestorName) {
		return true
	}

	// Look up the descendant class
	_, ok := r.classes.Get(descendantName)
	if !ok {
		return false
	}

	// Walk up the parent chain looking for the ancestor
	currentParent := r.GetParentName(descendantName)
	for currentParent != "" {
		if ident.Equal(currentParent, ancestorName) {
			return true
		}

		_, ok := r.classes.Get(currentParent)
		if !ok {
			// Parent not in registry - can't continue
			break
		}
		currentParent = r.GetParentName(currentParent)
	}

	return false
}

// GetAllClasses returns a map of all registered classes.
// The map uses normalized keys (case-insensitive).
// The returned map should not be modified directly.
func (r *ClassRegistry) GetAllClasses() map[string]runtime.IClassInfo {
	result := make(map[string]runtime.IClassInfo, r.classes.Len())
	r.classes.Range(func(name string, entry runtime.IClassInfo) bool {
		result[ident.Normalize(name)] = entry
		return true
	})
	return result
}

// Count returns the number of classes in the registry.
func (r *ClassRegistry) Count() int {
	return r.classes.Len()
}

// GetClassNames returns a slice of all registered class names (original case).
// The names are not sorted.
func (r *ClassRegistry) GetClassNames() []string {
	return r.classes.Keys()
}

// FindDescendants returns all classes that inherit from the given ancestor class.
// The result includes both direct children and all descendants in the hierarchy.
// The ancestor class itself is not included in the result.
//
// Example: If Cat and Dog inherit from Animal:
//
//	FindDescendants("Animal") returns [Cat, Dog]
func (r *ClassRegistry) FindDescendants(ancestorName string) []runtime.IClassInfo {
	descendants := []runtime.IClassInfo{}

	r.classes.Range(func(name string, entry runtime.IClassInfo) bool {
		// Skip the ancestor itself
		if ident.Equal(name, ancestorName) {
			return true // continue iteration
		}

		// Check if this class inherits from the ancestor
		if r.IsDescendantOf(name, ancestorName) {
			descendants = append(descendants, entry)
		}
		return true // continue iteration
	})

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
	_, ok := r.classes.Get(name)
	if !ok {
		return -1
	}

	depth := 0
	currentParent := r.GetParentName(name)

	for currentParent != "" {
		_, ok := r.classes.Get(currentParent)
		if !ok {
			// Parent not found - stop here
			break
		}
		depth++
		currentParent = r.GetParentName(currentParent)
	}

	return depth
}

// Clear removes all classes from the registry.
func (r *ClassRegistry) Clear() {
	r.classes.Clear()
	r.parentNames.Clear()
}

// Unregister removes a class from the registry by name.
func (r *ClassRegistry) Unregister(name string) {
	r.classes.Delete(name)
	r.parentNames.Delete(name)
}
