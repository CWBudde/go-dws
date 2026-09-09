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

	if entries, ok := r.entries.Get(signature.Operator); ok {
		for _, existing := range entries {
			if OperatorOperandsEqual(existing.OperandTypes, signature.OperandTypes) {
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

	return SelectOperatorOverload(operandTypes, entries, func(entry *OperatorSignature) []Type { return entry.OperandTypes })
}

// SelectOperatorOverload chooses an exact signature before compatible signatures.
// Class ancestor distances are ordered left to right, matching the runtime's
// historical operand-chain traversal. Equal matches retain declaration order.
func SelectOperatorOverload[T any](operands []Type, entries []T, signature func(T) []Type) (T, bool) {
	for _, entry := range entries {
		if OperatorOperandsEqual(operands, signature(entry)) {
			return entry, true
		}
	}
	var best T
	var bestDistances []int
	found := false
	for _, entry := range entries {
		distances, compatible := operatorOperandDistances(operands, signature(entry))
		if compatible && (!found || operatorDistancesLess(distances, bestDistances)) {
			best = entry
			bestDistances = distances
			found = true
		}
	}
	return best, found
}

func operatorOperandDistances(actual, declared []Type) ([]int, bool) {
	if len(actual) != len(declared) {
		return nil, false
	}
	// The final position ranks exact non-class operands ahead of conversions
	// at the same ancestor combination, as the old exact-then-compatible lookup did.
	distances := make([]int, len(actual)+1)
	for index, operand := range actual {
		if !OperatorTypesCompatible(operand, declared[index]) {
			return nil, false
		}
		class, isClass := GetUnderlyingType(operand).(*ClassType)
		target, targetIsClass := GetUnderlyingType(declared[index]).(*ClassType)
		if !isClass || !targetIsClass {
			if !OperatorTypesEqual(operand, declared[index]) {
				distances[len(actual)] = 1
			}
			continue
		}
		for current := class; current != nil; current = current.Parent {
			if ident.Equal(current.Name, target.Name) {
				break
			}
			distances[index]++
		}
	}
	return distances, true
}

func operatorDistancesLess(left, right []int) bool {
	for index, distance := range left {
		if distance != right[index] {
			return distance < right[index]
		}
	}
	return false
}

// OperatorTypesCompatible checks if actualType can be used where declaredType is expected.
// This supports inheritance: a subclass instance can be used where parent class is expected.
func OperatorTypesCompatible(actualType, declaredType Type) bool {
	// Exact match
	if actualType == nil || declaredType == nil {
		return false
	}
	actualType, declaredType = GetUnderlyingType(actualType), GetUnderlyingType(declaredType)
	if OperatorTypesEqual(actualType, declaredType) {
		return true
	}

	// Check class inheritance: actualType is a subclass of declaredType
	actualClass, actualIsClass := actualType.(*ClassType)
	declaredClass, declaredIsClass := declaredType.(*ClassType)

	if actualIsClass && declaredIsClass {
		// Walk up the inheritance chain to see if actualClass is a subclass of declaredClass
		for class := actualClass; class != nil; class = class.Parent {
			if ident.Equal(class.Name, declaredClass.Name) {
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
		_, variantElement := declaredElem.(*VariantType)
		if declaredArray.IsDynamic() && variantElement {
			return true // array of const accepts any array (static or dynamic, any element type)
		}

		// For non-array-of-const cases, both must be dynamic arrays
		if !actualArray.IsDynamic() || !declaredArray.IsDynamic() {
			return false
		}

		// Check if element types are compatible (recursive for nested arrays)
		return OperatorTypesCompatible(actualArray.ElementType, declaredArray.ElementType)
	}

	return false
}

// OperatorTypesEqual compares resolved operand identities without formatting types.
// Named identifiers are case insensitive; aliases use their underlying type.
func OperatorTypesEqual(left, right Type) bool {
	if left == nil || right == nil {
		return false
	}
	left, right = GetUnderlyingType(left), GetUnderlyingType(right)
	switch l := left.(type) {
	case *ClassType:
		r, ok := right.(*ClassType)
		return ok && ident.Equal(l.Name, r.Name)
	case *InterfaceType:
		r, ok := right.(*InterfaceType)
		return ok && ident.Equal(l.Name, r.Name)
	case *EnumType:
		r, ok := right.(*EnumType)
		return ok && ident.Equal(l.Name, r.Name)
	case *ArrayType:
		r, ok := right.(*ArrayType)
		return ok && operatorArrayTypesEqual(l, r)
	case *RecordType:
		r, ok := right.(*RecordType)
		if !ok {
			return false
		}
		if l.Name != "" || r.Name != "" {
			return ident.Equal(l.Name, r.Name)
		}
		return l.Equals(r)
	default:
		return left.Equals(right)
	}
}

func operatorArrayTypesEqual(left, right *ArrayType) bool {
	if !OperatorTypesEqual(left.ElementType, right.ElementType) {
		return false
	}
	if (left.IndexType == nil) != (right.IndexType == nil) {
		return false
	}
	if left.IndexType != nil && !OperatorTypesEqual(left.IndexType, right.IndexType) {
		return false
	}
	return equalArrayBound(left.LowBound, right.LowBound) && equalArrayBound(left.HighBound, right.HighBound)
}

func equalArrayBound(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

// OperatorOperandsEqual compares complete resolved operand signatures.
func OperatorOperandsEqual(left, right []Type) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !OperatorTypesEqual(left[i], right[i]) {
			return false
		}
	}
	return true
}

// FormatTypeList formats resolved types for diagnostics only.
func FormatTypeList(operands []Type) string {
	parts := make([]string, len(operands))
	for i, operand := range operands {
		if operand == nil {
			parts[i] = "<unresolved>"
		} else {
			parts[i] = operand.String()
		}
	}
	return strings.Join(parts, ", ")
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

// ConversionRegistry stores typed implicit and explicit conversion signatures.
type ConversionRegistry struct{ implicit, explicit []*ConversionSignature }

// NewConversionRegistry creates an empty conversion registry.
func NewConversionRegistry() *ConversionRegistry { return &ConversionRegistry{} }

// Register rejects duplicate resolved conversion signatures.
func (r *ConversionRegistry) Register(signature *ConversionSignature) error {
	if signature == nil || signature.From == nil || signature.To == nil {
		return errors.New("conversion requires resolved source and target types")
	}
	var entries *[]*ConversionSignature
	switch signature.Kind {
	case ConversionImplicit:
		entries = &r.implicit
	case ConversionExplicit:
		entries = &r.explicit
	default:
		return fmt.Errorf("unknown conversion kind: %d", signature.Kind)
	}
	for _, entry := range *entries {
		if OperatorTypesEqual(entry.From, signature.From) && OperatorTypesEqual(entry.To, signature.To) {
			return ErrConversionDuplicate
		}
	}
	*entries = append(*entries, signature)
	return nil
}

// FindImplicit returns an implicit conversion between resolved types.
func (r *ConversionRegistry) FindImplicit(from, to Type) (*ConversionSignature, bool) {
	if r == nil {
		return nil, false
	}
	return findConversion(r.implicit, from, to)
}

// FindExplicit returns an explicit conversion between resolved types.
func (r *ConversionRegistry) FindExplicit(from, to Type) (*ConversionSignature, bool) {
	if r == nil {
		return nil, false
	}
	return findConversion(r.explicit, from, to)
}

func findConversion(entries []*ConversionSignature, from, to Type) (*ConversionSignature, bool) {
	for _, entry := range entries {
		if OperatorTypesEqual(entry.From, from) && OperatorTypesEqual(entry.To, to) {
			return entry, true
		}
	}
	return nil, false
}
