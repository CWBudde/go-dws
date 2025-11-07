package interp

import (
	"fmt"
	"strings"
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
	for _, entry := range r.entries[key] {
		if operatorSignatureKey(entry.OperandTypes) == operatorSignatureKey(operandTypes) {
			return entry, true
		}
	}
	return nil, false
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
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

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
			if strings.ToUpper(entry.From) == current.currentType {
				nextType := strings.ToUpper(entry.To)

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
	return strings.ToUpper(from) + "->" + strings.ToUpper(to)
}

func operatorSignatureKey(operandTypes []string) string {
	return strings.Join(operandTypes, "|")
}

func normalizeTypeAnnotation(name string) string {
	trimmed := strings.TrimSpace(name)
	switch strings.ToLower(trimmed) {
	case "integer":
		return "INTEGER"
	case "float":
		return "FLOAT"
	case "string":
		return "STRING"
	case "boolean":
		return "BOOLEAN"
	case "variant":
		return "VARIANT"
	case "nil":
		return "NIL"
	default:
		if strings.HasPrefix(strings.ToLower(trimmed), "array of") {
			return strings.ToUpper(trimmed)
		}
		return "CLASS:" + trimmed
	}
}

func valueTypeKey(val Value) string {
	if val == nil {
		return "NIL"
	}
	switch v := val.(type) {
	case *ObjectInstance:
		if v.Class != nil {
			return "CLASS:" + v.Class.Name
		}
		return "CLASS:"
	case *RecordValue:
		if v.RecordType != nil && v.RecordType.Name != "" {
			return "CLASS:" + v.RecordType.Name
		}
		return "RECORD"
	default:
		return val.Type()
	}
}
