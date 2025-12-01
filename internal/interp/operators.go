package interp

import (
	"fmt"
	"strings"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

type runtimeOperatorEntry struct {
	Class         *ClassInfo
	Operator      string
	BindingName   string
	OperandTypes  []string
	SelfIndex     int
	IsClassMethod bool
}

type runtimeOperatorRegistry struct {
	entries map[string][]*runtimeOperatorEntry
}

func newRuntimeOperatorRegistry() *runtimeOperatorRegistry {
	return &runtimeOperatorRegistry{
		entries: make(map[string][]*runtimeOperatorEntry),
	}
}

func (r *runtimeOperatorRegistry) register(entry *runtimeOperatorEntry) error {
	if entry == nil {
		return fmt.Errorf("runtime operator entry cannot be nil")
	}
	key := strings.ToLower(entry.Operator)
	for _, existing := range r.entries[key] {
		if operatorSignatureKey(existing.OperandTypes) == operatorSignatureKey(entry.OperandTypes) {
			return fmt.Errorf("operator already registered")
		}
	}
	r.entries[key] = append(r.entries[key], entry)
	return nil
}

func (r *runtimeOperatorRegistry) clone() *runtimeOperatorRegistry {
	if r == nil {
		return newRuntimeOperatorRegistry()
	}
	clone := newRuntimeOperatorRegistry()
	for op, list := range r.entries {
		copied := make([]*runtimeOperatorEntry, len(list))
		copy(copied, list)
		clone.entries[op] = copied
	}
	return clone
}

func (r *runtimeOperatorRegistry) lookup(operator string, operandTypes []string) (*runtimeOperatorEntry, bool) {
	if r == nil {
		return nil, false
	}
	key := strings.ToLower(operator)

	// First try exact match for performance
	for _, entry := range r.entries[key] {
		if operatorSignatureKey(entry.OperandTypes) == operatorSignatureKey(operandTypes) {
			return entry, true
		}
	}

	// If no exact match, try assignment-compatible match (for inheritance)
	// This allows subclasses to use operators defined on parent classes
	for _, entry := range r.entries[key] {
		if len(entry.OperandTypes) != len(operandTypes) {
			continue
		}

		allCompatible := true
		for i := range operandTypes {
			if !areRuntimeTypesCompatibleForOperator(operandTypes[i], entry.OperandTypes[i], entry.Class) {
				allCompatible = false
				break
			}
		}

		if allCompatible {
			return entry, true
		}
	}

	return nil, false
}

// areRuntimeTypesCompatibleForOperator checks if actualType can be used where declaredType is expected.
// This supports inheritance: a subclass instance can be used where parent class is expected.
func areRuntimeTypesCompatibleForOperator(actualType, declaredType string, declaredClass *ClassInfo) bool {
	normalizedActual := ident.Normalize(actualType)
	normalizedDeclared := ident.Normalize(declaredType)

	// Exact match (case-insensitive)
	if normalizedActual == normalizedDeclared {
		return true
	}

	// Check array type compatibility
	// array of T is compatible with array of Variant (array of const) for any type T
	if strings.HasPrefix(normalizedActual, "array of ") && normalizedDeclared == "array of variant" {
		return true
	}

	// Check class inheritance: actualType is a subclass of declaredType
	// Both types are in format "class:classname"
	if !strings.HasPrefix(normalizedActual, "class:") || !strings.HasPrefix(normalizedDeclared, "class:") {
		return false
	}

	actualClassName := strings.TrimPrefix(normalizedActual, "class:")
	declaredClassName := strings.TrimPrefix(normalizedDeclared, "class:")

	// TODO: Full inheritance checking is not yet implemented here
	// The function currently only does simple name comparison because we don't have
	// easy access to the actual class hierarchy from the runtime type strings.
	// The full inheritance check is handled in tryCallClassOperator which walks up
	// the parent class chain. The declaredClass parameter is kept for future enhancement
	// when we refactor to pass full ClassInfo objects instead of type strings.
	if declaredClass != nil {
		if ident.Normalize(declaredClass.Name) == declaredClassName {
			return true
		}
	}

	return actualClassName == declaredClassName
}

type runtimeConversionEntry struct {
	From        string
	To          string
	BindingName string
	Implicit    bool
}

type runtimeConversionRegistry struct {
	implicit map[string]*runtimeConversionEntry
	explicit map[string]*runtimeConversionEntry
}

func newRuntimeConversionRegistry() *runtimeConversionRegistry {
	return &runtimeConversionRegistry{
		implicit: make(map[string]*runtimeConversionEntry),
		explicit: make(map[string]*runtimeConversionEntry),
	}
}

func (r *runtimeConversionRegistry) register(entry *runtimeConversionEntry) error {
	if entry == nil {
		return fmt.Errorf("conversion entry cannot be nil")
	}
	key := conversionKey(entry.From, entry.To)
	if entry.Implicit {
		if _, exists := r.implicit[key]; exists {
			return fmt.Errorf("implicit conversion already registered")
		}
		r.implicit[key] = entry
	} else {
		if _, exists := r.explicit[key]; exists {
			return fmt.Errorf("explicit conversion already registered")
		}
		r.explicit[key] = entry
	}
	return nil
}

func (r *runtimeConversionRegistry) findImplicit(from, to string) (*runtimeConversionEntry, bool) {
	if r == nil {
		return nil, false
	}
	entry, ok := r.implicit[conversionKey(from, to)]
	return entry, ok
}

// findConversionPath uses BFS to find the shortest path of implicit conversions from source to target type.
// Returns a slice of intermediate type names representing the conversion path, or nil if no path exists.
// maxDepth limits the number of conversions in the chain (e.g., maxDepth=3 allows A->B->C->D).
func (r *runtimeConversionRegistry) findConversionPath(from, to string, maxDepth int) []string {
	if r == nil || maxDepth <= 0 {
		return nil
	}

	// Normalize type names
	from = ident.Normalize(from)
	to = ident.Normalize(to)

	// Direct conversion check
	if _, ok := r.implicit[conversionKey(from, to)]; ok {
		return []string{from, to}
	}

	// BFS to find shortest conversion path
	type queueItem struct {
		currentType string
		path        []string
	}

	visited := make(map[string]bool)
	queue := []queueItem{{currentType: from, path: []string{from}}}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check if path is too long
		if len(current.path) > maxDepth {
			continue
		}

		// Try all possible conversions from current type
		for _, entry := range r.implicit {
			// Check if this conversion starts from current type
			if ident.Normalize(entry.From) == current.currentType {
				nextType := ident.Normalize(entry.To)

				// Found target!
				if nextType == to {
					return append(current.path, nextType)
				}

				// Add to queue if not visited
				if !visited[nextType] {
					visited[nextType] = true
					newPath := make([]string, len(current.path)+1)
					copy(newPath, current.path)
					newPath[len(current.path)] = nextType
					queue = append(queue, queueItem{
						currentType: nextType,
						path:        newPath,
					})
				}
			}
		}
	}

	// No path found
	return nil
}

func conversionKey(from, to string) string {
	return ident.Normalize(from) + "->" + ident.Normalize(to)
}

func operatorSignatureKey(operandTypes []string) string {
	return strings.Join(operandTypes, "|")
}

// NormalizeTypeAnnotation normalizes a type annotation string for operator lookup.
// Primitive types (integer, float, string, boolean, variant, nil) and array types
// are returned normalized. All other types get a "class:" prefix.
// This is a convenience wrapper around interptypes.NormalizeTypeAnnotation.
func NormalizeTypeAnnotation(name string) string {
	return interptypes.NormalizeTypeAnnotation(name)
}

func valueTypeKey(val Value) string {
	if val == nil {
		return "nil"
	}
	switch v := val.(type) {
	case *ObjectInstance:
		if v.Class != nil {
			return "class:" + ident.Normalize(v.Class.GetName())
		}
		return "class:"
	case *RecordValue:
		if v.RecordType != nil && v.RecordType.Name != "" {
			return "class:" + ident.Normalize(v.RecordType.Name)
		}
		return "record"
	case *ArrayValue:
		// Include array element type for operator overload matching
		if v.ArrayType != nil && v.ArrayType.ElementType != nil {
			elemTypeStr := v.ArrayType.ElementType.String()
			return "array of " + ident.Normalize(elemTypeStr)
		}
		return "array"
	default:
		return ident.Normalize(val.Type())
	}
}
