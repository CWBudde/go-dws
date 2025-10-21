package interp

import (
	"fmt"
	"strings"
)

type runtimeOperatorEntry struct {
	Operator      string
	OperandTypes  []string
	BindingName   string
	Class         *ClassInfo
	IsClassMethod bool
	SelfIndex     int
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
	switch v := val.(type) {
	case *ObjectInstance:
		if v.Class != nil {
			return "CLASS:" + v.Class.Name
		}
		return "CLASS:"
	default:
		return val.Type()
	}
}
