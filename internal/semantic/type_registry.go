package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Type Descriptor
// ============================================================================

// TypeDescriptor contains metadata about a registered type.
// It includes the type itself, its position in source code, and visibility information.
type TypeDescriptor struct {
	// Name is the canonical name of the type (case-preserved)
	Name string

	// Type is the actual type instance
	Type types.Type

	// Position is the source location where the type was defined
	Position token.Position

	// Visibility controls type scope (0=private, 1=unit, 2=public)
	// Used for unit system and access control
	Visibility int
}

// ============================================================================
// Type Registry
// ============================================================================

// TypeRegistry manages all type registrations and lookups in a DWScript program.
// It replaces the 7 scattered type maps (classes, interfaces, enums, records, sets, arrays, typeAliases)
// with a unified registry that provides:
//   - Centralized type registration and lookup
//   - Case-insensitive type resolution
//   - Position tracking for LSP support
//   - Type iteration and querying by kind
//   - Duplicate type detection
//
// The registry is used during semantic analysis to track all user-defined types
// and built-in types in the program.
type TypeRegistry struct {
	// types is a case-insensitive map of type names to their descriptors.
	// Uses ident.Map for automatic case normalization and original casing preservation.
	types *ident.Map[*TypeDescriptor]

	// typesByKind provides fast lookup of types by their TypeKind()
	// This is populated lazily on first TypesByKind() call
	kindIndex map[string][]string
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types:     ident.NewMap[*TypeDescriptor](),
		kindIndex: make(map[string][]string),
	}
}

// ============================================================================
// Core Registration and Lookup
// ============================================================================

// Register adds a type to the registry with position tracking.
// Returns an error if a type with the same name (case-insensitive) already exists.
//
// Parameters:
//   - name: The type name (case will be preserved in TypeDescriptor.Name)
//   - typ: The type to register
//   - pos: Source position where the type is defined
//   - visibility: Visibility level (0=private, 1=unit, 2=public)
func (r *TypeRegistry) Register(name string, typ types.Type, pos token.Position, visibility int) error {
	if name == "" {
		return fmt.Errorf("cannot register type with empty name")
	}
	if typ == nil {
		return fmt.Errorf("cannot register nil type")
	}

	// Check for duplicates
	if existing, exists := r.types.Get(name); exists {
		return fmt.Errorf("type '%s' already defined at %s", existing.Name, existing.Position)
	}

	// Register the type
	descriptor := &TypeDescriptor{
		Name:       name,
		Type:       typ,
		Position:   pos,
		Visibility: visibility,
	}
	r.types.Set(name, descriptor)

	// Invalidate kind index since we added a new type
	r.kindIndex = make(map[string][]string)

	return nil
}

// Resolve looks up a type by name (case-insensitive).
// Returns the type and true if found, nil and false otherwise.
func (r *TypeRegistry) Resolve(name string) (types.Type, bool) {
	descriptor, exists := r.types.Get(name)
	if !exists {
		return nil, false
	}
	return descriptor.Type, true
}

// ResolveDescriptor looks up a type descriptor by name (case-insensitive).
// Returns the full descriptor and true if found, nil and false otherwise.
// This is useful when you need position information or visibility.
func (r *TypeRegistry) ResolveDescriptor(name string) (*TypeDescriptor, bool) {
	return r.types.Get(name)
}

// MustResolve looks up a type and panics if not found.
// This should only be used for built-in types that are guaranteed to exist.
func (r *TypeRegistry) MustResolve(name string) types.Type {
	typ, ok := r.Resolve(name)
	if !ok {
		panic(fmt.Sprintf("type '%s' not found in registry", name))
	}
	return typ
}

// ResolveUnderlying looks up a type by name and resolves through any type aliases
// to get the ultimate underlying type. This is useful when you need the actual
// concrete type behind potentially multiple layers of aliases.
//
// For example:
//   - type MyInt = Integer;      // ResolveUnderlying("MyInt") -> IntegerType
//   - type A = Integer; type B = A;  // ResolveUnderlying("B") -> IntegerType
//
// Returns the underlying type and true if found, nil and false otherwise.
func (r *TypeRegistry) ResolveUnderlying(name string) (types.Type, bool) {
	typ, ok := r.Resolve(name)
	if !ok {
		return nil, false
	}
	// Use the existing GetUnderlyingType to follow the alias chain
	return types.GetUnderlyingType(typ), true
}

// ============================================================================
// Query and Iteration
// ============================================================================

// AllTypes returns a map of all registered types.
// The keys are the canonical (case-preserved) names.
// This creates a new map to avoid external modifications.
func (r *TypeRegistry) AllTypes() map[string]types.Type {
	result := make(map[string]types.Type, r.types.Len())
	r.types.Range(func(name string, descriptor *TypeDescriptor) bool {
		result[name] = descriptor.Type
		return true
	})
	return result
}

// AllDescriptors returns a map of all type descriptors.
// The keys are the canonical (case-preserved) names.
// This creates a new map to avoid external modifications.
func (r *TypeRegistry) AllDescriptors() map[string]*TypeDescriptor {
	result := make(map[string]*TypeDescriptor, r.types.Len())
	r.types.Range(func(name string, descriptor *TypeDescriptor) bool {
		result[name] = descriptor
		return true
	})
	return result
}

// TypesByKind returns all types of a specific kind.
// The kind should match the TypeKind() return value (e.g., "CLASS", "INTERFACE", "ENUM").
// Returns a slice of type names (canonical, case-preserved).
func (r *TypeRegistry) TypesByKind(kind string) []string {
	// Build kind index if not already built
	if len(r.kindIndex) == 0 && r.types.Len() > 0 {
		r.buildKindIndex()
	}

	// Return types for this kind (may be empty slice)
	return r.kindIndex[kind]
}

// buildKindIndex creates the kindIndex for fast TypesByKind queries.
// This is called lazily on first TypesByKind() call.
func (r *TypeRegistry) buildKindIndex() {
	r.kindIndex = make(map[string][]string)
	r.types.Range(func(name string, descriptor *TypeDescriptor) bool {
		kind := descriptor.Type.TypeKind()
		r.kindIndex[kind] = append(r.kindIndex[kind], name)
		return true
	})
}

// Count returns the total number of registered types
func (r *TypeRegistry) Count() int {
	return r.types.Len()
}

// ============================================================================
// LSP Support - Position-based Queries
// ============================================================================

// FindTypeByPosition finds a type defined at the given position.
// Returns the type descriptor and true if found, nil and false otherwise.
// This is useful for LSP "type at cursor" features.
func (r *TypeRegistry) FindTypeByPosition(pos token.Position) (*TypeDescriptor, bool) {
	var found *TypeDescriptor
	r.types.Range(func(_ string, descriptor *TypeDescriptor) bool {
		// Exact position match (line and column)
		if descriptor.Position.Line == pos.Line && descriptor.Position.Column == pos.Column {
			found = descriptor
			return false // Stop iteration
		}
		return true // Continue iteration
	})
	if found != nil {
		return found, true
	}
	return nil, false
}

// TypesInRange returns all types defined within a line range.
// This is useful for LSP features like "show all types in current scope".
func (r *TypeRegistry) TypesInRange(startLine, endLine int) []*TypeDescriptor {
	var result []*TypeDescriptor
	r.types.Range(func(_ string, descriptor *TypeDescriptor) bool {
		if descriptor.Position.Line >= startLine && descriptor.Position.Line <= endLine {
			result = append(result, descriptor)
		}
		return true
	})
	return result
}

// ============================================================================
// Dependency Analysis
// ============================================================================

// GetTypeDependencies returns all types that the given type depends on.
// For example, a record type depends on the types of its fields.
// Returns a slice of type names (canonical, case-preserved).
func (r *TypeRegistry) GetTypeDependencies(typeName string) []string {
	descriptor, ok := r.ResolveDescriptor(typeName)
	if !ok {
		return nil
	}

	var dependencies []string

	// Extract dependencies based on type kind
	switch t := descriptor.Type.(type) {
	case *types.RecordType:
		// Record depends on field types (Fields is map[string]Type)
		for _, fieldType := range t.Fields {
			dependencies = append(dependencies, fieldType.String())
		}
	case *types.ClassType:
		// Class depends on field types and parent class
		if t.Parent != nil {
			dependencies = append(dependencies, t.Parent.String())
		}
		// Fields is map[string]Type
		for _, fieldType := range t.Fields {
			dependencies = append(dependencies, fieldType.String())
		}
	case *types.ArrayType:
		// Array depends on element type
		dependencies = append(dependencies, t.ElementType.String())
	case *types.SetType:
		// Set depends on element type
		dependencies = append(dependencies, t.ElementType.String())
	case *types.TypeAlias:
		// Type alias depends on aliased type
		dependencies = append(dependencies, t.AliasedType.String())
	case *types.FunctionPointerType:
		// Function pointer depends on parameter and return types
		for _, paramType := range t.Parameters {
			dependencies = append(dependencies, paramType.String())
		}
		dependencies = append(dependencies, t.ReturnType.String())
	case *types.InterfaceType:
		// Interface depends on parent interface (if any)
		if t.Parent != nil {
			dependencies = append(dependencies, t.Parent.String())
		}
	}

	return dependencies
}

// ============================================================================
// Utility Methods
// ============================================================================

// Clear removes all types from the registry.
// This is mainly useful for testing.
func (r *TypeRegistry) Clear() {
	r.types.Clear()
	r.kindIndex = make(map[string][]string)
}

// Unregister removes a type from the registry.
// Returns true if the type was found and removed, false otherwise.
func (r *TypeRegistry) Unregister(name string) bool {
	if r.types.Delete(name) {
		// Invalidate kind index
		r.kindIndex = make(map[string][]string)
		return true
	}
	return false
}

// Has checks if a type with the given name is registered (case-insensitive).
// This is useful for checking existence without retrieving the type.
func (r *TypeRegistry) Has(name string) bool {
	return r.types.Has(name)
}

// RegisterBuiltIn is a convenience method for registering built-in types
// that don't have a source position. It uses position 0:0 and public visibility (2).
func (r *TypeRegistry) RegisterBuiltIn(name string, typ types.Type) error {
	return r.Register(name, typ, token.Position{Line: 0, Column: 0, Offset: 0}, 2)
}

// MustRegisterBuiltIn registers a built-in type and panics if registration fails.
// This should only be used during initialization when built-in types are guaranteed to be unique.
func (r *TypeRegistry) MustRegisterBuiltIn(name string, typ types.Type) {
	if err := r.RegisterBuiltIn(name, typ); err != nil {
		panic(fmt.Sprintf("failed to register built-in type '%s': %v", name, err))
	}
}
