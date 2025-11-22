package builtins

import (
	"sort"
	"sync"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// Category represents a category of built-in functions.
type Category string

const (
	// CategoryMath includes mathematical functions (Abs, Sin, Cos, Sqrt, etc.)
	CategoryMath Category = "math"

	// CategoryString includes string manipulation functions (UpperCase, Trim, etc.)
	CategoryString Category = "string"

	// CategoryDateTime includes date/time functions (Now, FormatDateTime, etc.)
	CategoryDateTime Category = "datetime"

	// CategoryConversion includes type conversion functions (IntToStr, StrToInt, etc.)
	CategoryConversion Category = "conversion"

	// CategoryEncoding includes encoding/escaping functions (StrToHtml, StrToJSON, etc.)
	CategoryEncoding Category = "encoding"

	// CategoryJSON includes JSON parsing and manipulation functions
	CategoryJSON Category = "json"

	// CategoryType includes type introspection functions (TypeOf, TypeOfClass)
	CategoryType Category = "type"

	// CategoryArray includes array operations (Length, Copy, Reverse, etc.)
	CategoryArray Category = "array"

	// CategoryCollections includes higher-order collection functions (Map, Filter, Reduce, etc.)
	CategoryCollections Category = "collections"

	// CategoryIO includes input/output functions (Print, PrintLn)
	CategoryIO Category = "io"

	// CategoryVariant includes Variant introspection and conversion functions
	CategoryVariant Category = "variant"

	// CategorySystem includes system and miscellaneous functions
	CategorySystem Category = "system"
)

// FunctionInfo holds metadata about a built-in function.
type FunctionInfo struct {
	// Name is the canonical name of the function (case-sensitive storage)
	Name string

	// Function is the implementation
	Function BuiltinFunc

	// Category is the functional category
	Category Category

	// Description is a brief description of what the function does
	Description string
}

// Registry manages all built-in functions.
// It provides case-insensitive lookup and categorization of built-in functions.
type Registry struct {
	functions  *ident.Map[*FunctionInfo]
	categories map[Category][]string
	mu         sync.RWMutex
}

// NewRegistry creates a new built-in function registry.
func NewRegistry() *Registry {
	return &Registry{
		functions:  ident.NewMap[*FunctionInfo](),
		categories: make(map[Category][]string),
	}
}

// Register adds a built-in function to the registry.
// The name lookup will be case-insensitive (DWScript is case-insensitive).
// If a function with the same name is already registered, it will be replaced,
// but the category list will not contain duplicate entries.
func (r *Registry) Register(name string, fn BuiltinFunc, category Category, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := &FunctionInfo{
		Name:        name,
		Function:    fn,
		Category:    category,
		Description: description,
	}

	// Check if already registered to prevent duplicate category entries
	if r.functions.Has(name) {
		// Update the function but don't add to category list again
		r.functions.Set(name, info)
		return
	}

	// Store with case-insensitive lookup (ident.Map handles normalization)
	r.functions.Set(name, info)

	// Add to category list (only for new registrations)
	r.categories[category] = append(r.categories[category], name)
}

// RegisterBatch registers multiple functions at once.
// Each entry in the batch is a tuple of (name, function, category, description).
func (r *Registry) RegisterBatch(entries []struct {
	Name        string
	Function    BuiltinFunc
	Category    Category
	Description string
}) {
	for _, entry := range entries {
		r.Register(entry.Name, entry.Function, entry.Category, entry.Description)
	}
}

// Lookup finds a built-in function by name (case-insensitive).
// Returns the function and true if found, nil and false otherwise.
func (r *Registry) Lookup(name string) (BuiltinFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if info, ok := r.functions.Get(name); ok {
		return info.Function, true
	}
	return nil, false
}

// Get retrieves the full FunctionInfo for a function by name (case-insensitive).
// Returns the info and true if found, nil and false otherwise.
func (r *Registry) Get(name string) (*FunctionInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.functions.Get(name)
}

// GetByCategory returns all functions in a given category.
// The returned slice is sorted alphabetically by function name.
func (r *Registry) GetByCategory(category Category) []*FunctionInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.categories[category]
	result := make([]*FunctionInfo, 0, len(names))

	for _, name := range names {
		if info, ok := r.functions.Get(name); ok {
			result = append(result, info)
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// AllCategories returns a list of all categories that have registered functions.
func (r *Registry) AllCategories() []Category {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]Category, 0, len(r.categories))
	for category := range r.categories {
		categories = append(categories, category)
	}

	// Sort for consistent ordering
	sort.Slice(categories, func(i, j int) bool {
		return categories[i] < categories[j]
	})

	return categories
}

// AllFunctions returns all registered functions, sorted alphabetically by name.
func (r *Registry) AllFunctions() []*FunctionInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*FunctionInfo, 0, r.functions.Len())
	r.functions.Range(func(_ string, info *FunctionInfo) bool {
		result = append(result, info)
		return true
	})

	// Sort by name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// Count returns the total number of registered functions.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.functions.Len()
}

// CategoryCount returns the number of functions in a specific category.
func (r *Registry) CategoryCount(category Category) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.categories[category])
}

// Has checks if a function is registered (case-insensitive).
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.functions.Has(name)
}

// Clear removes all registered functions.
// This is primarily useful for testing.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.functions.Clear()
	r.categories = make(map[Category][]string)
}
