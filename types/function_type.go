package types

import (
	"strings"
)

// FunctionType represents a function or procedure signature.
// It includes parameter types and an optional return type.
// For procedures (no return value), ReturnType should be VOID.
type FunctionType struct {
	Parameters []Type // Parameter types in order
	ReturnType Type   // Return type (VOID for procedures)
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

// NewFunctionType creates a new function type with the given parameters and return type
func NewFunctionType(params []Type, returnType Type) *FunctionType {
	return &FunctionType{
		Parameters: params,
		ReturnType: returnType,
	}
}

// NewProcedureType creates a new procedure type (returns void)
func NewProcedureType(params []Type) *FunctionType {
	return &FunctionType{
		Parameters: params,
		ReturnType: VOID,
	}
}
