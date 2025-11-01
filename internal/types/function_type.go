package types

import (
	"strings"
)

// FunctionType represents a function or procedure signature.
// It includes parameter types and an optional return type.
// For procedures (no return value), ReturnType should be VOID.
//
// Metadata arrays (ParamNames, LazyParams, VarParams) are parallel to Parameters
// and store additional information about each parameter for validation and interpretation.
type FunctionType struct {
	ReturnType Type
	Parameters []Type
	ParamNames []string // Parameter names for better error messages
	LazyParams []bool   // true if parameter is lazy (expression capture)
	VarParams  []bool   // true if parameter is var/byref (pass by reference)
}

// String returns a string representation of the function type.
// Examples:
//   - () -> Integer
//   - (Integer, String) -> Boolean
//   - (Float) -> Void
func (ft *FunctionType) String() string {
	var sb strings.Builder

	sb.WriteString("(")
	for i, param := range ft.Parameters {
		if i > 0 {
			sb.WriteString(", ")
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
// each parameter type is equal, and the return types are equal.
func (ft *FunctionType) Equals(other Type) bool {
	otherFunc, ok := other.(*FunctionType)
	if !ok {
		return false
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
// Metadata arrays are initialized to empty (no lazy or var parameters).
func NewFunctionType(params []Type, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters: params,
		ReturnType: returnType,
		ParamNames: make([]string, len(params)),
		LazyParams: make([]bool, len(params)),
		VarParams:  make([]bool, len(params)),
	}
}

// NewProcedureType creates a new procedure type (returns void).
// Metadata arrays are initialized to empty (no lazy or var parameters).
func NewProcedureType(params []Type) *FunctionType {
	return &FunctionType{
		Parameters: params,
		ReturnType: VOID,
		ParamNames: make([]string, len(params)),
		LazyParams: make([]bool, len(params)),
		VarParams:  make([]bool, len(params)),
	}
}

// NewFunctionTypeWithMetadata creates a new function type with full parameter metadata.
// This is used during semantic analysis to track lazy and var parameter modifiers.
// All arrays (params, names, lazy, varParams) must have the same length.
func NewFunctionTypeWithMetadata(params []Type, names []string, lazy []bool, varParams []bool, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters: params,
		ReturnType: returnType,
		ParamNames: names,
		LazyParams: lazy,
		VarParams:  varParams,
	}
}
