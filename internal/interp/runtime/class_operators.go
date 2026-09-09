package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ClassOperatorEntry describes a class operator binding and its operand signature.
type ClassOperatorEntry struct {
	Class         *ClassInfo
	Operator      string
	BindingName   string
	OperandTypes  []types.Type
	SelfIndex     int
	IsClassMethod bool
}

// ClassOperatorRegistry stores the operator overloads available on a class.
type ClassOperatorRegistry struct {
	entries map[string][]*ClassOperatorEntry
}

// NewClassOperatorRegistry creates an empty class operator registry.
func NewClassOperatorRegistry() *ClassOperatorRegistry {
	return &ClassOperatorRegistry{
		entries: make(map[string][]*ClassOperatorEntry),
	}
}

// Register adds an operator binding, rejecting duplicate operand signatures.
func (r *ClassOperatorRegistry) Register(entry *ClassOperatorEntry) error {
	if entry == nil {
		return fmt.Errorf("runtime operator entry cannot be nil")
	}
	key := ident.Normalize(entry.Operator)
	for _, existing := range r.entries[key] {
		if types.OperatorOperandsEqual(existing.OperandTypes, entry.OperandTypes) {
			return fmt.Errorf("operator already registered")
		}
	}
	r.entries[key] = append(r.entries[key], entry)
	return nil
}

// Clone copies the registry while preserving the shared immutable bindings.
func (r *ClassOperatorRegistry) Clone() *ClassOperatorRegistry {
	if r == nil {
		return NewClassOperatorRegistry()
	}
	clone := NewClassOperatorRegistry()
	for op, list := range r.entries {
		copied := make([]*ClassOperatorEntry, len(list))
		copy(copied, list)
		clone.entries[op] = copied
	}
	return clone
}
