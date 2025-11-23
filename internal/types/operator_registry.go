package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// ErrOperatorDuplicate is returned when attempting to register a duplicate operator signature.
var ErrOperatorDuplicate = errors.New("operator signature already registered")

// ErrConversionDuplicate is returned when attempting to register a duplicate conversion.
var ErrConversionDuplicate = errors.New("conversion already registered")

// OperatorSignature describes an operator overload, including operand types and result type.
type OperatorSignature struct {
	ResultType   Type
	Operator     string
	Binding      string
	OperandTypes []Type
}

// OperatorRegistry stores operator overloads keyed by operator token.
// Uses ident.Map for case-insensitive operator lookup (e.g., "and" vs "AND").
type OperatorRegistry struct {
	entries *ident.Map[[]*OperatorSignature]
}

// NewOperatorRegistry creates an empty operator registry.
func NewOperatorRegistry() *OperatorRegistry {
	return &OperatorRegistry{
		entries: ident.NewMap[[]*OperatorSignature](),
	}
}

// Register adds an operator signature to the registry.
// Returns ErrOperatorDuplicate if an identical signature already exists.
func (r *OperatorRegistry) Register(signature *OperatorSignature) error {
	if signature == nil {
		return errors.New("nil operator signature")
	}

	key := operatorEntryKey(signature.Operator, signature.OperandTypes)
	if entries, ok := r.entries.Get(signature.Operator); ok {
		for _, existing := range entries {
			if operatorEntryKey(existing.Operator, existing.OperandTypes) == key {
				return ErrOperatorDuplicate
			}
		}
		r.entries.Set(signature.Operator, append(entries, signature))
	} else {
		r.entries.Set(signature.Operator, []*OperatorSignature{signature})
	}
	return nil
}

// Lookup finds an operator signature that matches the given operand types.
// Support inheritance - operands are compatible if they're assignable to declared types.
func (r *OperatorRegistry) Lookup(operator string, operandTypes []Type) (*OperatorSignature, bool) {
	entries, ok := r.entries.Get(operator)
	if !ok {
		return nil, false
	}

	// First, try exact match for performance
	key := operatorEntryKey(operator, operandTypes)
	for _, entry := range entries {
		if operatorEntryKey(entry.Operator, entry.OperandTypes) == key {
			return entry, true
		}
	}

	// If no exact match, try assignment-compatible match (for inheritance)
	// This allows subclasses to use operators defined on parent classes
	for _, entry := range entries {
		if len(entry.OperandTypes) != len(operandTypes) {
			continue
		}

		allCompatible := true
		for i := range operandTypes {
			if !areTypesCompatibleForOperator(operandTypes[i], entry.OperandTypes[i]) {
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

// areTypesCompatibleForOperator checks if actualType can be used where declaredType is expected.
// This supports inheritance: a subclass instance can be used where parent class is expected.
func areTypesCompatibleForOperator(actualType, declaredType Type) bool {
	// Exact match
	if actualType.Equals(declaredType) {
		return true
	}

	// Check class inheritance: actualType is a subclass of declaredType
	actualClass, actualIsClass := actualType.(*ClassType)
	declaredClass, declaredIsClass := declaredType.(*ClassType)

	if actualIsClass && declaredIsClass {
		// Walk up the inheritance chain to see if actualClass is a subclass of declaredClass
		for class := actualClass; class != nil; class = class.Parent {
			if class.Name == declaredClass.Name {
				return true
			}
		}
	}

	// Check array compatibility for array of const
	// An array of any type can be passed where array of Variant is expected
	actualArray, actualIsArray := actualType.(*ArrayType)
	declaredArray, declaredIsArray := declaredType.(*ArrayType)

	if actualIsArray && declaredIsArray {
		// Special case: array of const (dynamic array of Variant) accepts any array type
		// This must be checked BEFORE the dynamic-only restriction, because array of const
		// can accept static arrays like [1, 2] or ['a', 'b', 'c']
		declaredElem := GetUnderlyingType(declaredArray.ElementType)
		if declaredArray.IsDynamic() && declaredElem.TypeKind() == "VARIANT" {
			return true // array of const accepts any array (static or dynamic, any element type)
		}

		// For non-array-of-const cases, both must be dynamic arrays
		if !actualArray.IsDynamic() || !declaredArray.IsDynamic() {
			return false
		}

		// Check if element types are compatible (recursive for nested arrays)
		return areTypesCompatibleForOperator(actualArray.ElementType, declaredArray.ElementType)
	}

	return false
}

// operatorEntryKey creates a stable string key for an operator + operand signature.
func operatorEntryKey(operator string, operandTypes []Type) string {
	parts := make([]string, len(operandTypes))
	for i, operand := range operandTypes {
		parts[i] = typeKey(operand)
	}
	return fmt.Sprintf("%s(%s)", ident.Normalize(operator), strings.Join(parts, ","))
}

// ConversionKind indicates whether a conversion is implicit or explicit.
type ConversionKind int

const (
	// ConversionImplicit registers an implicit conversion (automatically applied).
	ConversionImplicit ConversionKind = iota
	// ConversionExplicit registers an explicit conversion (requires explicit syntax).
	ConversionExplicit
)

// ConversionSignature describes a type conversion operator.
type ConversionSignature struct {
	From    Type
	To      Type
	Binding string
	Kind    ConversionKind
}

// ConversionRegistry stores implicit and explicit conversions.
type ConversionRegistry struct {
	implicit map[string]*ConversionSignature
	explicit map[string]*ConversionSignature
}

// NewConversionRegistry creates an empty conversion registry.
func NewConversionRegistry() *ConversionRegistry {
	return &ConversionRegistry{
		implicit: make(map[string]*ConversionSignature),
		explicit: make(map[string]*ConversionSignature),
	}
}

// Register adds a conversion signature to the registry.
// Returns ErrConversionDuplicate if an identical conversion already exists.
func (r *ConversionRegistry) Register(signature *ConversionSignature) error {
	if signature == nil {
		return errors.New("nil conversion signature")
	}

	key := conversionKey(signature.From, signature.To)
	switch signature.Kind {
	case ConversionImplicit:
		if _, exists := r.implicit[key]; exists {
			return ErrConversionDuplicate
		}
		r.implicit[key] = signature
	case ConversionExplicit:
		if _, exists := r.explicit[key]; exists {
			return ErrConversionDuplicate
		}
		r.explicit[key] = signature
	default:
		return fmt.Errorf("unknown conversion kind: %d", signature.Kind)
	}

	return nil
}

// FindImplicit returns an implicit conversion between types, if any.
func (r *ConversionRegistry) FindImplicit(from, to Type) (*ConversionSignature, bool) {
	if r == nil {
		return nil, false
	}
	sig, ok := r.implicit[conversionKey(from, to)]
	return sig, ok
}

// FindExplicit returns an explicit conversion between types, if any.
func (r *ConversionRegistry) FindExplicit(from, to Type) (*ConversionSignature, bool) {
	if r == nil {
		return nil, false
	}
	sig, ok := r.explicit[conversionKey(from, to)]
	return sig, ok
}

// conversionKey builds a stable key identifying a conversion pair.
func conversionKey(from, to Type) string {
	return typeKey(from) + "->" + typeKey(to)
}

// typeKey generates a canonical string for a Type used in operator/conversion lookups.
// Type names are normalized for case-insensitive matching.
func typeKey(t Type) string {
	switch tt := t.(type) {
	case *ClassType:
		return "class:" + ident.Normalize(tt.Name)
	case *InterfaceType:
		return "interface:" + ident.Normalize(tt.Name)
	case *ArrayType:
		return "array:" + tt.String()
	case *RecordType:
		if tt.Name != "" {
			return "record:" + ident.Normalize(tt.Name)
		}
		return "record:" + tt.String()
	case *FunctionType:
		return "function:" + tt.String()
	default:
		return t.TypeKind() + ":" + t.String()
	}
}
