package builtins

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ParameterConstraint refines the values accepted by a builtin call parameter.
// The nominal parameter type remains available for function pointer signatures.
type ParameterConstraint struct {
	// ArrayElement accepts static and dynamic arrays with this exact element type.
	ArrayElement types.Type
	// Types overrides the nominal type with a union of accepted types.
	Types []types.Type
	// Exact disables implicit widening and compares primitive type identities.
	Exact bool
	// ResolveAliases unwraps aliases before applying the constraint.
	ResolveAliases bool
	// AllowVariant permits runtime conversion of Variant and JSONVariant values.
	AllowVariant bool
	// AllowJSONVariant permits JSONVariant alongside an exact Variant constraint.
	AllowJSONVariant bool
	// AllowIntegerSubrange accepts a subrange whose base type is Integer.
	AllowIntegerSubrange bool
	// AllowEnum accepts enumeration values after optional alias resolution.
	AllowEnum bool
	// EnumName restricts AllowEnum to the enumeration with this name, so that an
	// unrelated enumeration cannot be passed for its ordinal value.
	EnumName string
	// Numeric accepts the type system's numeric types.
	Numeric bool
	// Any accepts every successfully analyzed expression type.
	Any bool
}

// WithConstraints attaches call constraints to a newly constructed signature.
func (s *FunctionSignature) WithConstraints(constraints ...ParameterConstraint) *FunctionSignature {
	s.Constraints = constraints
	return s
}

// WithVarParams marks the given parameter indexes as by-reference (var)
// parameters and constrains them to an exact type match.
func (s *FunctionSignature) WithVarParams(indexes ...int) *FunctionSignature {
	if len(s.VarParams) < len(s.ParamTypes) {
		s.VarParams = make([]bool, len(s.ParamTypes))
	}
	if len(s.Constraints) < len(s.ParamTypes) {
		constraints := make([]ParameterConstraint, len(s.ParamTypes))
		copy(constraints, s.Constraints)
		s.Constraints = constraints
	}
	for _, index := range indexes {
		if index < 0 || index >= len(s.VarParams) {
			continue
		}
		s.VarParams[index] = true
		s.Constraints[index] = exactParameter
	}
	return s
}

// HasVarParams reports whether any parameter is passed by reference.
func (s *FunctionSignature) HasVarParams() bool {
	for _, isVar := range s.VarParams {
		if isVar {
			return true
		}
	}
	return false
}

// IsVarParam reports whether the parameter at index is passed by reference.
func (s *FunctionSignature) IsVarParam(index int) bool {
	return index >= 0 && index < len(s.VarParams) && s.VarParams[index]
}

// WithArgCounts specifies the allowed disjoint arities of a signature.
func (s *FunctionSignature) WithArgCounts(counts ...int) *FunctionSignature {
	s.AllowedArgCounts = counts
	return s
}

// AcceptsArgCount reports whether the signature accepts count arguments.
func (s *FunctionSignature) AcceptsArgCount(count int) bool {
	if len(s.AllowedArgCounts) != 0 {
		for _, allowed := range s.AllowedArgCounts {
			if count == allowed {
				return true
			}
		}
		return false
	}
	return count >= s.MinArgs && (s.MaxArgs < 0 || count <= s.MaxArgs)
}

// AcceptsArgument reports whether an argument type satisfies its call constraint.
// Unknown argument types are handled by expression analysis, avoiding cascades.
func (s *FunctionSignature) AcceptsArgument(index int, actual types.Type) bool {
	if actual == nil || len(s.ParamTypes) == 0 {
		return true
	}
	if index >= len(s.ParamTypes) {
		if !s.IsVariadic {
			return true
		}
		index = len(s.ParamTypes) - 1
	}
	constraint := ParameterConstraint{}
	if index < len(s.Constraints) {
		constraint = s.Constraints[index]
	}
	return constraint.accepts(s.ParamTypes[index], actual)
}

func (c ParameterConstraint) accepts(expected, actual types.Type) bool {
	if c.Any || actual == nil {
		return true
	}
	if c.acceptsVariant(actual) {
		return true
	}
	if c.ResolveAliases {
		actual = types.GetUnderlyingType(actual)
	}
	if c.acceptsOrdinal(actual) {
		return true
	}
	if c.Numeric {
		return types.IsNumericType(actual)
	}
	if c.ArrayElement != nil {
		array, ok := actual.(*types.ArrayType)
		return ok && array.ElementType == c.ArrayElement
	}
	if len(c.Types) != 0 {
		for _, candidate := range c.Types {
			if candidate == actual {
				return true
			}
		}
		return false
	}
	if c.Exact {
		return expected == actual
	}
	return ordinaryParameterAccepts(expected, actual)
}

var (
	exactParameter        = ParameterConstraint{Exact: true}
	numericParameter      = ParameterConstraint{Types: []types.Type{types.INTEGER, types.FLOAT}}
	variantParameter      = ParameterConstraint{Exact: true, AllowJSONVariant: true}
	integerRangeParameter = ParameterConstraint{Exact: true, AllowIntegerSubrange: true}
)

func (c ParameterConstraint) acceptsVariant(actual types.Type) bool {
	if c.AllowVariant && (types.GetUnderlyingType(actual).TypeKind() == "VARIANT" || types.IsJSONVariant(actual)) {
		return true
	}
	if c.AllowJSONVariant && types.IsJSONVariant(actual) {
		return true
	}
	return false
}

func (c ParameterConstraint) acceptsOrdinal(actual types.Type) bool {
	if c.AllowIntegerSubrange {
		if subrange, ok := actual.(*types.SubrangeType); ok && subrange.BaseType == types.INTEGER {
			return true
		}
	}
	if !c.AllowEnum || actual.TypeKind() != "ENUM" {
		return false
	}
	if c.EnumName == "" {
		return true
	}
	enum, ok := actual.(*types.EnumType)
	return ok && ident.Equal(enum.Name, c.EnumName)
}

func ordinaryParameterAccepts(expected, actual types.Type) bool {
	if expected == nil || expected == types.VARIANT {
		return true
	}
	if expected == types.FLOAT {
		return types.INTEGER.Equals(actual) || types.FLOAT.Equals(actual)
	}
	return expected.Equals(actual)
}
