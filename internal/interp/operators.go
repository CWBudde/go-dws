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

// Task 3.5.22e: Conversion registry has been migrated to TypeSystem.Conversions()
// The runtimeConversionEntry and runtimeConversionRegistry types have been removed.
// Use i.typeSystem.Conversions().FindImplicit() and FindConversionPath() instead.

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
