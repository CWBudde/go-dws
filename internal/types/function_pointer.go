package types

import (
	"strings"
)

// FunctionPointerType represents a function or procedure pointer type in DWScript.
// Function pointers can be stored in variables, passed as parameters, and called indirectly.
//
// Examples:
//   - type TComparator = function(a, b: Integer): Integer;
//   - type TCallback = procedure(msg: String);
type FunctionPointerType struct {
	ReturnType Type
	Parameters []Type
}

// NewFunctionPointerType creates a new function pointer type with the given parameters and return type.
func NewFunctionPointerType(params []Type, returnType Type) *FunctionPointerType {
	return &FunctionPointerType{
		Parameters: params,
		ReturnType: returnType,
	}
}

// NewProcedurePointerType creates a new procedure pointer type (no return value).
func NewProcedurePointerType(params []Type) *FunctionPointerType {
	return &FunctionPointerType{
		Parameters: params,
		ReturnType: nil,
	}
}

// TypeKind returns the type kind identifier for function pointers.
func (f *FunctionPointerType) TypeKind() string {
	return "FUNCTION_POINTER"
}

// String returns a string representation of the function pointer type.
// Examples:
//   - "function(Integer, String): Boolean"
//   - "procedure(Integer)"
//   - "function(): Integer"
//   - "procedure()"
func (f *FunctionPointerType) String() string {
	var sb strings.Builder

	// Write function or procedure keyword
	if f.IsProcedure() {
		sb.WriteString("procedure")
	} else {
		sb.WriteString("function")
	}

	// Write parameters
	sb.WriteRune('(')
	for i, param := range f.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteRune(')')

	// Write return type for functions
	if f.IsFunction() {
		sb.WriteString(": ")
		sb.WriteString(f.ReturnType.String())
	}

	return sb.String()
}

// Equals checks if two function pointer types are identical.
// Two function pointer types are equal if they have the same parameter types and return type.
func (f *FunctionPointerType) Equals(other Type) bool {
	// Resolve type aliases
	other = GetUnderlyingType(other)

	// Check if the other type is a function pointer
	otherFunc, ok := other.(*FunctionPointerType)
	if !ok {
		return false
	}

	// Check parameter count (nil and empty slice are equivalent)
	if len(f.Parameters) != len(otherFunc.Parameters) {
		return false
	}

	// Check each parameter type
	for i := range f.Parameters {
		if !f.Parameters[i].Equals(otherFunc.Parameters[i]) {
			return false
		}
	}

	// Check return type (nil-safe comparison)
	if f.ReturnType == nil && otherFunc.ReturnType == nil {
		return true
	}
	if f.ReturnType == nil || otherFunc.ReturnType == nil {
		return false
	}
	return f.ReturnType.Equals(otherFunc.ReturnType)
}

// IsProcedure returns true if this is a procedure pointer (no return type).
func (f *FunctionPointerType) IsProcedure() bool {
	return f.ReturnType == nil
}

// IsFunction returns true if this is a function pointer (has a return type).
func (f *FunctionPointerType) IsFunction() bool {
	return f.ReturnType != nil
}

// IsCompatibleWith checks if this function pointer type is compatible with another type.
// Function pointers are compatible if their signatures match exactly.
// Function pointers CANNOT be assigned to method pointers (methods need object context).
func (f *FunctionPointerType) IsCompatibleWith(other Type) bool {
	// Resolve type aliases
	other = GetUnderlyingType(other)

	// Check if the other type is a function pointer
	if otherFunc, ok := other.(*FunctionPointerType); ok {
		return f.Equals(otherFunc)
	}

	// Function pointers cannot be assigned to method pointers
	// (method pointers need an object instance, regular functions don't have one)
	return false
}

// MethodPointerType represents a method pointer type (procedure/function of object).
// Method pointers bind both a method and an object instance.
//
// Examples:
//   - type TNotifyEvent = procedure(Sender: TObject) of object;
//   - type TCompareFunc = function(a, b: Integer): Integer of object;
type MethodPointerType struct {
	FunctionPointerType      // Embedded function pointer type
	OfObject            bool // Always true for method pointers
}

// NewMethodPointerType creates a new method pointer type with the given parameters and return type.
func NewMethodPointerType(params []Type, returnType Type) *MethodPointerType {
	return &MethodPointerType{
		FunctionPointerType: FunctionPointerType{
			Parameters: params,
			ReturnType: returnType,
		},
		OfObject: true,
	}
}

// TypeKind returns the type kind identifier for method pointers.
func (m *MethodPointerType) TypeKind() string {
	return "METHOD_POINTER"
}

// String returns a string representation of the method pointer type.
// Examples:
//   - "function(Integer, String): Boolean of object"
//   - "procedure(Integer) of object"
func (m *MethodPointerType) String() string {
	return m.FunctionPointerType.String() + " of object"
}

// Equals checks if two method pointer types are identical.
// Two method pointer types are equal if they have the same parameter types and return type.
// Method pointers are NOT equal to function pointers, even if their signatures match.
func (m *MethodPointerType) Equals(other Type) bool {
	// Resolve type aliases
	other = GetUnderlyingType(other)

	// Check if the other type is a method pointer
	otherMethod, ok := other.(*MethodPointerType)
	if !ok {
		return false
	}

	// Compare the underlying function signatures
	return m.FunctionPointerType.Equals(&otherMethod.FunctionPointerType)
}

// IsCompatibleWith checks if this method pointer type is compatible with another type.
// Method pointers are compatible with:
//   - Other method pointers with matching signatures
//   - Function pointers with matching signatures (method can be used as function)
func (m *MethodPointerType) IsCompatibleWith(other Type) bool {
	// Resolve type aliases
	other = GetUnderlyingType(other)

	// Check if the other type is a method pointer
	if otherMethod, ok := other.(*MethodPointerType); ok {
		return m.Equals(otherMethod)
	}

	// Check if the other type is a function pointer
	// Method pointers can be assigned to function pointers if signatures match
	if otherFunc, ok := other.(*FunctionPointerType); ok {
		return m.FunctionPointerType.Equals(otherFunc)
	}

	return false
}
