package types

import (
	"strings"
)

// FunctionType represents a function or procedure signature.
// It includes parameter types and an optional return type.
// For procedures (no return value), ReturnType should be VOID.
//
// Metadata arrays (ParamNames, LazyParams, VarParams, ConstParams, DefaultValues) are parallel to Parameters
// and store additional information about each parameter for validation and interpretation.
//
// Variadic support: If IsVariadic is true, the last parameter accepts a variable number of arguments.
// VariadicType holds the element type of the variadic parameter (e.g., Integer for "array of Integer").
// The last entry in Parameters should be an ArrayType whose ElementType matches VariadicType.
//
// Optional parameters: Parameters with default values are optional. DefaultValues stores the default
// expression for each parameter (nil if required). When a function is called with fewer arguments,
// the interpreter evaluates the default expressions to fill in missing arguments.
type FunctionType struct {
	ReturnType    Type
	VariadicType  Type
	Parameters    []Type
	ParamNames    []string
	DefaultValues []interface{}
	LazyParams    []bool
	VarParams     []bool
	ConstParams   []bool
	IsVariadic    bool
}

// String returns a string representation of the function type.
// Examples:
//   - () -> Integer
//   - (Integer, String) -> Boolean
//   - (Float) -> Void
//   - (Integer, ...String) -> Boolean  (variadic)
func (ft *FunctionType) String() string {
	var sb strings.Builder

	sb.WriteString("(")
	for i, param := range ft.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		// Mark variadic parameter with "..." prefix
		if ft.IsVariadic && i == len(ft.Parameters)-1 {
			sb.WriteString("...")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(")")

	sb.WriteString(" -> ")
	sb.WriteString(ft.ReturnType.String())

	return sb.String()
}

// TypeKind returns "FUNCTION" for function types
func (ft *FunctionType) TypeKind() string {
	return "FUNCTION"
}

// Equals checks if two function types are equal.
// Two function types are equal if they have the same number of parameters,
// each parameter type is equal, the return types are equal, and both have
// the same variadic status. For variadic functions, the variadic element types
// must also match.
func (ft *FunctionType) Equals(other Type) bool {
	otherFunc, ok := other.(*FunctionType)
	if !ok {
		return false
	}

	// Check variadic status
	if ft.IsVariadic != otherFunc.IsVariadic {
		return false
	}

	// Check variadic type if both are variadic
	if ft.IsVariadic {
		if ft.VariadicType == nil || otherFunc.VariadicType == nil {
			// One has variadic type set, the other doesn't
			if ft.VariadicType != otherFunc.VariadicType {
				return false
			}
		} else if !ft.VariadicType.Equals(otherFunc.VariadicType) {
			return false
		}
	}

	// Check parameter count
	if len(ft.Parameters) != len(otherFunc.Parameters) {
		return false
	}

	// Check each parameter type
	for i, param := range ft.Parameters {
		if !param.Equals(otherFunc.Parameters[i]) {
			return false
		}
	}

	// Check return type
	return ft.ReturnType.Equals(otherFunc.ReturnType)
}

// IsProcedure returns true if this is a procedure (returns void)
func (ft *FunctionType) IsProcedure() bool {
	return ft.ReturnType.TypeKind() == "VOID"
}

// IsFunction returns true if this is a function (returns a value)
func (ft *FunctionType) IsFunction() bool {
	return !ft.IsProcedure()
}

// NewFunctionType creates a new function type with the given parameters and return type.
// Metadata arrays are initialized to empty (no lazy or var parameters, all parameters required).
func NewFunctionType(params []Type, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters:    params,
		ReturnType:    returnType,
		ParamNames:    make([]string, len(params)),
		DefaultValues: make([]interface{}, len(params)),
		LazyParams:    make([]bool, len(params)),
		VarParams:     make([]bool, len(params)),
		ConstParams:   make([]bool, len(params)),
	}
}

// NewProcedureType creates a new procedure type (returns void).
// Metadata arrays are initialized to empty (no lazy or var parameters, all parameters required).
func NewProcedureType(params []Type) *FunctionType {
	return &FunctionType{
		Parameters:    params,
		ReturnType:    VOID,
		ParamNames:    make([]string, len(params)),
		DefaultValues: make([]interface{}, len(params)),
		LazyParams:    make([]bool, len(params)),
		VarParams:     make([]bool, len(params)),
		ConstParams:   make([]bool, len(params)),
	}
}

// NewFunctionTypeWithMetadata creates a new function type with full parameter metadata.
// This is used during semantic analysis to track lazy, var, const, and optional parameter modifiers.
// All arrays (params, names, defaults, lazy, varParams, constParams) must have the same length.
func NewFunctionTypeWithMetadata(params []Type, names []string, defaults []interface{}, lazy []bool, varParams []bool, constParams []bool, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters:    params,
		ReturnType:    returnType,
		ParamNames:    names,
		DefaultValues: defaults,
		LazyParams:    lazy,
		VarParams:     varParams,
		ConstParams:   constParams,
	}
}

// NewVariadicFunctionType creates a new variadic function type.
// The last parameter in params should be an ArrayType representing the variadic parameter.
// variadicType is the element type of the variadic array (e.g., Integer for "array of Integer").
// Example: NewVariadicFunctionType([]Type{STRING}, INTEGER, BOOLEAN) creates
// (String, ...Integer) -> Boolean
func NewVariadicFunctionType(params []Type, variadicType Type, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters:    params,
		ReturnType:    returnType,
		ParamNames:    make([]string, len(params)),
		DefaultValues: make([]interface{}, len(params)),
		LazyParams:    make([]bool, len(params)),
		VarParams:     make([]bool, len(params)),
		ConstParams:   make([]bool, len(params)),
		IsVariadic:    true,
		VariadicType:  variadicType,
	}
}

// NewVariadicFunctionTypeWithMetadata creates a variadic function type with full parameter metadata.
// This extends NewFunctionTypeWithMetadata to support variadic parameters.
// variadicType specifies the element type of the variadic parameter.
// All arrays (params, names, defaults, lazy, varParams, constParams) must have the same length.
func NewVariadicFunctionTypeWithMetadata(params []Type, names []string, defaults []interface{}, lazy []bool, varParams []bool, constParams []bool, variadicType Type, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters:    params,
		ReturnType:    returnType,
		ParamNames:    names,
		DefaultValues: defaults,
		LazyParams:    lazy,
		VarParams:     varParams,
		ConstParams:   constParams,
		IsVariadic:    true,
		VariadicType:  variadicType,
	}
}
