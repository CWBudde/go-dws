package runtime

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// ClassHierarchy supplies inheritance queries for operator matching.
type ClassHierarchy interface{ IsClassDescendantOf(string, string) bool }

// ClassOperatorEntry describes a class operator binding and its operand signature.
type ClassOperatorEntry struct {
	Class         *ClassInfo
	Operator      string
	BindingName   string
	OperandTypes  []string
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
		if operatorSignatureKey(existing.OperandTypes) == operatorSignatureKey(entry.OperandTypes) {
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

func (r *ClassOperatorRegistry) lookup(operator string, operandTypes []string, typeSystem ClassHierarchy) (*ClassOperatorEntry, bool) {
	if r == nil {
		return nil, false
	}
	key := ident.Normalize(operator)

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
			if !areRuntimeTypesCompatibleForOperator(operandTypes[i], entry.OperandTypes[i], typeSystem) {
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
func areRuntimeTypesCompatibleForOperator(actualType, declaredType string, typeSystem ClassHierarchy) bool {
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

	// Use TypeSystem when available for full inheritance checks
	if typeSystem != nil && typeSystem.IsClassDescendantOf(actualClassName, declaredClassName) {
		return true
	}

	return actualClassName == declaredClassName
}

func operatorSignatureKey(operandTypes []string) string {
	return strings.Join(operandTypes, "|")
}

// NormalizeTypeAnnotation builds the operator lookup key for a declared type.
func NormalizeTypeAnnotation(name string) string {
	trimmed := strings.TrimSpace(name)
	normalized := ident.Normalize(trimmed)

	// Check if it's a primitive type or array
	switch normalized {
	case "integer", "float", "string", "boolean", "variant", "nil":
		return normalized
	default:
		if ident.HasPrefix(trimmed, "array of") {
			return normalized
		}
		return "class:" + normalized
	}
}
